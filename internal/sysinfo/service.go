package sysinfo

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/process"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetSystemInfo() (SystemInfo, error) {
	hostInfo, _ := host.Info()
	loadAvg, _ := load.Avg()
	processes, _ := process.Processes()
	temperature := s.getTemperature()

	loadInfo := LoadInfo{}
	if loadAvg != nil {
		loadInfo = LoadInfo{
			Load1:  loadAvg.Load1,
			Load5:  loadAvg.Load5,
			Load15: loadAvg.Load15,
		}
	}

	return SystemInfo{
		Hostname:    hostInfo.Hostname,
		OS:          hostInfo.OS,
		Platform:    hostInfo.Platform,
		PlatformVer: hostInfo.PlatformVersion,
		Uptime:      hostInfo.Uptime,
		UptimeHuman: formatUptime(hostInfo.Uptime),
		Temperature: temperature,
		Processes:   len(processes),
		LastUpdate:  time.Now(),
		LoadAvg:     loadInfo,
	}, nil
}

func (s *Service) getTemperature() float64 {
	tempSensors, err := host.SensorsTemperatures()
	if err != nil || len(tempSensors) == 0 {
		return 0
	}

	var cpuTemp float64
	for _, sensor := range tempSensors {
		// Cari sensor AMD (Tctl/Tdie) atau Package Temp
		// Biasanya SensorKey-nya mengandung "Tctl" atau "composite"
		if sensor.SensorKey == "Tctl" || sensor.SensorKey == "package_id_0" {
			return sensor.Temperature
		}

		// Simpan suhu tertinggi sebagai cadangan (fallback)
		if sensor.Temperature > cpuTemp {
			cpuTemp = sensor.Temperature
		}
	}

	return cpuTemp // Return yang tertinggi kalau Tctl gak ketemu
}

func formatUptime(seconds uint64) string {
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60

	if days > 0 {
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	} else if hours > 0 {
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if minutes == 1 {
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", minutes)
}
