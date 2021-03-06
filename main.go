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
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const schedulerName = "k8s-resource-scheduler"

var Queue chan *Pod

func main() {
	log.Println(fmt.Sprintf("Starting %s scheduler...", schedulerName))

	doneChan := make(chan struct{})
	Queue = make(chan *Pod, 100)

	var wg sync.WaitGroup

	wg.Add(1)
	go monitorUnscheduledPods(doneChan, &wg)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go scheduleQueue(doneChan, Queue, &wg)
	}

	wg.Add(1)
	go reconcileUnscheduledPods(30, doneChan, &wg)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-signalChan:
			log.Printf("Shutdown signal received, exiting...")
			close(doneChan)
			wg.Wait()
			os.Exit(0)
		}
	}
}
