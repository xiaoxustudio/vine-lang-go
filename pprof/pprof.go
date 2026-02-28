package pprof

import (
	"fmt"
	"os"
	"runtime/pprof"
)

var (
	cpuProfileFile *os.File
	memProfileFile *os.File
)

// StartCPUProfile 启动CPU性能分析
func StartCPUProfile(filename string) error {
	if filename == "" {
		return nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create CPU profile: %v", err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return fmt.Errorf("could not start CPU profile: %v", err)
	}

	cpuProfileFile = f
	return nil
}

// StopCPUProfile 停止CPU性能分析
func StopCPUProfile() {
	if cpuProfileFile != nil {
		pprof.StopCPUProfile()
		cpuProfileFile.Close()
		cpuProfileFile = nil
	}
}

// WriteHeapProfile 写入内存性能分析
func WriteHeapProfile(filename string) error {
	if filename == "" {
		return nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create memory profile: %v", err)
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("could not write memory profile: %v", err)
	}

	return nil
}
