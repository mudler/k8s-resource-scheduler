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
	"strings"
)

func cpuUnits(s string) (int, error) {
	strip := strings.ReplaceAll(s, "n", "")
	usage, err := strconv.Atoi(strip)
	if err != nil {
		return 0, err
	}
	return usage, nil
}

func memoryUnits(s string) (int, error) {
	strip := strings.ReplaceAll(s, "Ki", "")
	usage, err := strconv.Atoi(strip)
	if err != nil {
		return 0, err
	}
	return usage, nil
}

func getPropertyBool(property string, m Metadata) bool {
	prop := getProperty(property, m)
	if prop == "" {
		prop = "false"
	}

	prop = strings.ToLower(prop)

	if prop == "true" {
		return true
	}

	return false
}

func getProperty(property string, m Metadata) string {
	prop, ok := m.Annotations[fmt.Sprintf("%s/%s", schedulerName, property)]
	if !ok {
		return ""
	}

	return prop
}

func bestNode(pod *Pod, nodes []Node) (*Node, error) {
	var bestNode *Node

	podCPUBound := getPropertyBool("cpu-bound", pod.Metadata)
	if podCPUBound {
		log.Println("Pod", pod.Metadata.Name, "is cpu bound")
	}

	podMemoryBound := getPropertyBool("memory-bound", pod.Metadata)
	if podMemoryBound {
		log.Println("Pod", pod.Metadata.Name, "is memory bound")
	}

NODES:
	for i := range nodes {
		currentNode := &nodes[i]

		for l, v := range pod.Spec.NodeSelector {
			if currentNode.Metadata.Labels[l] != v {
				log.Println("Node", currentNode.Metadata.Name, "does not match selector", currentNode.Metadata.Labels)
				continue NODES
			}
		}

		nodeCPUBound := getPropertyBool("cpu-bound", currentNode.Metadata)
		if nodeCPUBound {
			log.Println("Node", currentNode.Metadata.Name, "is cpu bound")
		}
		nodeMemoryBound := getPropertyBool("memory-bound", currentNode.Metadata)
		if nodeMemoryBound {
			log.Println("Node", currentNode.Metadata.Name, "is memory bound")
		}

		if bestNode == nil {
			bestNode = currentNode
			continue
		}
		log.Println(fmt.Sprintf("Current best CPU %s Memory %s (%s)", bestNode.NodeMetrics.Usage.Cpu, bestNode.NodeMetrics.Usage.Memory, bestNode.Metadata.Name))

		bestNodeCPUUsage, err := cpuUnits(bestNode.NodeMetrics.Usage.Cpu)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		bestNodeMemoryUsage, err := memoryUnits(bestNode.NodeMetrics.Usage.Memory)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		currentNodeCPUUsage, err := cpuUnits(currentNode.NodeMetrics.Usage.Cpu)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		currentNodeMemoryUsage, err := memoryUnits(currentNode.NodeMetrics.Usage.Memory)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		if (((podCPUBound || nodeCPUBound) && (!podMemoryBound || !nodeMemoryBound)) && (currentNodeCPUUsage < bestNodeCPUUsage)) || // CPU bound
			(((podMemoryBound || nodeMemoryBound) && (!podCPUBound || !nodeCPUBound)) && (currentNodeMemoryUsage < bestNodeMemoryUsage)) || // Memory bound
			(currentNodeMemoryUsage < bestNodeMemoryUsage && currentNodeCPUUsage < bestNodeCPUUsage) { // Otherwise
			log.Println(
				fmt.Sprintf("switch: %s (cpu %s mem %s) has a lower resource usage than %s (cpu %s mem %s)",
					currentNode.Metadata.Name, currentNode.NodeMetrics.Usage.Cpu, currentNode.NodeMetrics.Usage.Memory,
					bestNode.Metadata.Name, bestNode.NodeMetrics.Usage.Cpu, bestNode.NodeMetrics.Usage.Memory))
			bestNode = currentNode
		} else {
			log.Println(
				fmt.Sprintf("%s (cpu %s mem %s) has a lower resource usage than %s (cpu %s mem %s)",
					bestNode.Metadata.Name, bestNode.NodeMetrics.Usage.Cpu, bestNode.NodeMetrics.Usage.Memory,
					currentNode.Metadata.Name, currentNode.NodeMetrics.Usage.Cpu, currentNode.NodeMetrics.Usage.Memory))
		}
	}

	return bestNode, nil
}
