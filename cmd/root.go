/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cobra"
)

type cpuUsageEntry struct {
	Timestamp time.Time `json:"timestamp"`
	UserCPU   float64   `json:"user_cpu"`
	SystemCPU float64   `json:"system_cpu"`
}

type processTracker struct {
	Process    *process.Process
	CPUUsage   []float64
	CPUDetails []cpuUsageEntry
	PrevUser   float64
	PrevSystem float64
	PrevTime   time.Time
	FirstEntry bool
}

var (
	trackers = make(map[string]*processTracker)
	mutex    sync.Mutex
	port     string
)

func trackCPU() {
	for {
		mutex.Lock()
		for _, tracker := range trackers {
			userCPU, systemCPU := getProcessCPU(tracker.Process)
			curTime := time.Now()

			if !tracker.PrevTime.IsZero() {
				elapsed := curTime.Sub(tracker.PrevTime).Seconds()
				if elapsed > 0 && !tracker.FirstEntry {
					deltaUser := (userCPU - tracker.PrevUser) / elapsed * 100
					deltaSystem := (systemCPU - tracker.PrevSystem) / elapsed * 100
					tracker.CPUDetails = append(tracker.CPUDetails, cpuUsageEntry{
						Timestamp: curTime,
						UserCPU:   deltaUser,
						SystemCPU: deltaSystem,
					})
					tracker.CPUUsage = append(tracker.CPUUsage, deltaUser+deltaSystem)
				}
				tracker.FirstEntry = false
			}
			tracker.PrevUser = userCPU
			tracker.PrevSystem = systemCPU
			tracker.PrevTime = curTime
		}
		mutex.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func getProcessCPU(p *process.Process) (float64, float64) {
	times, err := p.Times()
	if err != nil {
		return 0, 0
	}
	return times.User, times.System
}

func findProcessByName(pattern string) (*process.Process, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	for _, p := range procs {
		procName, _ := p.Name()
		if re.MatchString(procName) {
			return p, nil
		}
	}
	return nil, fmt.Errorf("process matching pattern '%s' not found", pattern)
}

func calculateStatistics(data []float64) (median, p95, p99 float64) {
	if len(data) == 0 {
		return 0, 0, 0
	}

	sort.Float64s(data)
	median = data[len(data)/2]
	p95 = data[int(math.Ceil(float64(len(data))*0.95))-1]
	p99 = data[int(math.Ceil(float64(len(data))*0.99))-1]

	return median, p95, p99
}

func startTracking(w http.ResponseWriter, r *http.Request) {
	processName := r.URL.Path[len("/start/pgrep/"):]
	proc, err := findProcessByName(processName)
	if err != nil {
		http.Error(w, "Process not found", http.StatusNotFound)
		return
	}
	mutex.Lock()
	trackers[processName] = &processTracker{
		Process:    proc,
		CPUUsage:   []float64{},
		CPUDetails: []cpuUsageEntry{},
		PrevTime:   time.Now(),
		FirstEntry: true,
	}
	mutex.Unlock()
	json.NewEncoder(w).Encode(map[string]string{"message": "Started tracking " + processName})
}

func stopTrackingHandler(w http.ResponseWriter, r *http.Request) {
	processName := r.URL.Path[len("/stop/pgrep/"):]
	if processName == "" {
		response := stopAllTracking()
		json.NewEncoder(w).Encode(response)
	} else {
		response := stopTracking(processName)
		json.NewEncoder(w).Encode(response)
	}
}

func stopTracking(processName string) map[string]interface{} {
	mutex.Lock()
	defer mutex.Unlock()
	tracker, exists := trackers[processName]
	if !exists {
		return map[string]interface{}{"error": "Tracking is not active for " + processName}
	}
	median, p95, p99 := calculateStatistics(tracker.CPUUsage)
	delete(trackers, processName)
	return map[string]interface{}{
		"message":     "Stopped process " + processName,
		"median_cpu":  median,
		"p95_cpu":     p95,
		"p99_cpu":     p99,
		"cpu_details": tracker.CPUDetails,
	}
}

func stopAllTracking() map[string]interface{} {
	mutex.Lock()
	defer mutex.Unlock()
	results := make(map[string]interface{})
	for name, tracker := range trackers {
		median, p95, p99 := calculateStatistics(tracker.CPUUsage)
		_ = tracker.Process.Kill()
		results[name] = map[string]interface{}{
			"message":     "Stopped process " + name,
			"median_cpu":  median,
			"p95_cpu":     p95,
			"p99_cpu":     p99,
			"cpu_details": tracker.CPUDetails,
		}
		delete(trackers, name)
	}
	return results
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

var rootCmd = &cobra.Command{
	Use:   "cpu-tracker",
	Short: "A CPU usage tracker for processes",
	Run: func(cmd *cobra.Command, args []string) {
		go trackCPU()
		http.HandleFunc("/start/pgrep/", startTracking)
		http.HandleFunc("/stop/pgrep/", stopTrackingHandler)
		fmt.Printf("Server running on :%s\n", port)
		http.ListenAndServe(":"+port, nil)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&port, "port", "p", "5000", "Port to run the server on")
}
