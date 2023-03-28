package lutil

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// 获取程序的cpu使用率、内存使用率、虚存占用大小、实际内存占用大小
func GetAppInfo() (cpuPercent, memPercent, vszSize, rssSize int) {
	if runtime.GOOS == "windows" {
		return
	}
	cmd := exec.Command("ps", "aux")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return
	}
	mypid := os.Getpid()

	for {
		line, err := out.ReadString('\n')
		if err != nil {
			break
		}
		tokens := strings.Split(line, " ")

		ft := make([]string, 0)
		for _, t := range tokens {
			if t != "" && t != "\t" {
				ft = append(ft, t)
			}
		}

		pid, err := strconv.Atoi(ft[1])
		if pid == mypid {
			if val, err := strconv.ParseFloat(ft[2], 32); err == nil {
				cpuPercent = int(val * 100)
			}
			if val, err := strconv.ParseFloat(ft[3], 32); err == nil {
				memPercent = int(val * 100)
			}
			if val, err := strconv.ParseFloat(ft[4], 32); err == nil {
				vszSize = int(val / 1024 * 100)
			}
			if val, err := strconv.ParseFloat(ft[5], 32); err == nil {
				rssSize = int(val / 1024 * 100)
			}
			break
		}
	}
	return
}
