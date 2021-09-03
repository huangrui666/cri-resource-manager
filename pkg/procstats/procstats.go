// Copyright 2019 Intel Corporation. All Rights Reserved.
//
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

package procstats

import (
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
)

// CPUTimeStat is used to calculate the CPU usage.
type CPUTimeStat struct {
	sync.RWMutex
	PrevIdleTime       []uint64
	PrevTotalTime      []uint64
	CurIdleTime        []uint64
	CurTotalTime       []uint64
	DeltaIdleTime      []uint64
	DeltaTotalTime     []uint64
	CPUUsage           []float64
	IsGetCPUUsageBegin bool
	CPUNumber          int
}

var (
	// procRoot is the mount point for the proc filesystem
	procRoot = "/proc"
	procStat = procRoot + "/stat"
)

// GetCPUTimeStat calculates CPU usage by using the CPU time statistics from /proc/stat
func (t *CPUTimeStat) GetCPUTimeStat() error {
	lines, err := readLines(procStat)
	if err != nil {
		return err
	}
	// /proc/stat looks like this:
	// cpuid：user，nice, system, idle, iowait, irq, softirq
	// cpu  130216 19944 162525 1491240 3784 24749 17773 0 0 0
	// cpu0 40321 11452 49784 403099 2615 6076 6748 0 0 0
	// cpu1 26585 2425 36639 151166 404 2533 3541 0 0 0
	// ...
	t.Lock()
	defer t.Unlock()
	for i := 0; i < t.CPUNumber; i++ {
		index := i + 1
		split := strings.Split(lines[index], " ")
		t.CurIdleTime[i], _ = strconv.ParseUint(split[4], 10, 64)
		totalTime := uint64(0)
		for _, s := range split {
			u, _ := strconv.ParseUint(s, 10, 64)
			totalTime += u
		}
		t.CurTotalTime[i] = totalTime
		t.CPUUsage[i] = 0.0
		if t.IsGetCPUUsageBegin {
			t.DeltaIdleTime[i] = t.CurIdleTime[i] - t.PrevIdleTime[i]
			t.DeltaTotalTime[i] = t.CurTotalTime[i] - t.PrevTotalTime[i]
			t.CPUUsage[i] = (1.0 - float64(t.DeltaIdleTime[i])/float64(t.DeltaTotalTime[i])) * 100.0
		}
		t.PrevIdleTime[i] = t.CurIdleTime[i]
		t.PrevTotalTime[i] = t.CurTotalTime[i]
	}
	t.IsGetCPUUsageBegin = true
	return nil
}

func readLines(filePath string) ([]string, error) {
	f, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	data := string(f)
	rawLines := strings.Split(data, "\n")
	lines := make([]string, 0)
	for _, rawLine := range rawLines {
		if len(strings.TrimSpace(rawLine)) > 0 {
			lines = append(lines, rawLine)
		}
	}
	return lines, nil
}
