// Copyright 2016 Google Inc. All Rights Reserved.
// Copyright 2020 Ettore Di Giacinto
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

var processorLock = &sync.Mutex{}
var lastAllocation *time.Time

func reconcileUnscheduledPods(interval int, done chan struct{}, wg *sync.WaitGroup) {
	for {
		select {
		case <-time.After(time.Duration(interval) * time.Second):
			err := schedulePods()
			if err != nil {
				log.Println(err)
			}
		case <-done:
			wg.Done()
			log.Println("Stopped reconciliation loop.")
			return
		}
	}
}

func monitorUnscheduledPods(done chan struct{}, wg *sync.WaitGroup) {
	pods, errc := watchUnscheduledPods()

	for {
		select {
		case err := <-errc:
			log.Println(err)
		case pod := <-pods:
			processorLock.Lock()
			time.Sleep(2 * time.Second)
			err := schedulePod(&pod)
			if err != nil {
				log.Println(err)
			}
			processorLock.Unlock()
		case <-done:
			wg.Done()
			log.Println("Stopped scheduler.")
			return
		}
	}
}

func getPropertyInt(property string, m Metadata) int {
	prop := getProperty(property, m)
	if prop == "" {
		return 0
	}

	i, err := strconv.Atoi(prop)
	if err != nil {
		return 0
	}
	return i
}

func scheduleQueue(done chan struct{}, queue chan *Pod, wg *sync.WaitGroup) {
	for {
		select {
		case pod := <-queue:
			processorLock.Lock()
			time.Sleep(2 * time.Second)
			err := schedulePod(pod)
			if err != nil {
				log.Println(err)
			}
			processorLock.Unlock()
		case <-done:
			wg.Done()
			log.Println("Stopped scheduler.")
			return
		}
	}
}

func schedulePod(pod *Pod) error {
	now := time.Now()

	burstProtect := getPropertyInt("burst-protect", pod.Metadata)
	if lastAllocation != nil && burstProtect != 0 {
		diff := now.Sub(*lastAllocation)
		if diff.Seconds() < float64(burstProtect) {
			Queue <- pod
			return fmt.Errorf("burst detected for pod '%s' - waiting (diff: %f burst: %f)", pod.Metadata.Name, diff.Seconds(), float64(burstProtect))
		}
	}

	nodes, err := fit(pod)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return fmt.Errorf("Unable to schedule pod (%s) failed to fit in any node", pod.Metadata.Name)
	}
	node, err := bestNode(pod, nodes)
	if err != nil {
		return err
	}
	err = bind(pod, node)
	if err != nil {
		return err
	}

	lastAllocation = &now
	return nil
}

func schedulePods() error {
	processorLock.Lock()
	defer processorLock.Unlock()
	pods, err := getUnscheduledPods()
	if err != nil {
		return err
	}
	for _, pod := range pods {
		err := schedulePod(pod)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
