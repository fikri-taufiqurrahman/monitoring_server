package gpu

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const drmBasePath = "/sys/class/drm"

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) GetGPUInfo() (GPUInfo, error) {
	cardPath, err := findAMDCard()
	if err != nil {
		return GPUInfo{}, fmt.Errorf("finding AMD card: %w", err)
	}

	busyPercent, err := readFloat(filepath.Join(cardPath, "device/gpu_busy_percent"))
	if err != nil {
		return GPUInfo{}, fmt.Errorf("reading gpu_busy_percent: %w", err)
	}

	vramUsed, err := readInt(filepath.Join(cardPath, "device/mem_info_vram_used"))
	if err != nil {
		return GPUInfo{}, fmt.Errorf("reading mem_info_vram_used: %w", err)
	}

	vramTotal, err := readInt(filepath.Join(cardPath, "device/mem_info_vram_total"))
	if err != nil {
		return GPUInfo{}, fmt.Errorf("reading mem_info_vram_total: %w", err)
	}

	tempMilliCelsius, err := readTempInput(filepath.Join(cardPath, "device/hwmon"))
	if err != nil {
		return GPUInfo{}, fmt.Errorf("reading temperature: %w", err)
	}

	return GPUInfo{
		BusyPercent: busyPercent,
		VRAMUsed:    vramUsed,
		VRAMTotal:   vramTotal,
		TempCelsius: float64(tempMilliCelsius) / 1000.0,
	}, nil
}

// findAMDCard returns the sysfs path of the first AMD DRM card found.
func findAMDCard() (string, error) {
	entries, err := os.ReadDir(drmBasePath)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !strings.HasPrefix(e.Name(), "card") {
			continue
		}
		cardPath := filepath.Join(drmBasePath, e.Name())
		vendorPath := filepath.Join(cardPath, "device/vendor")
		data, err := os.ReadFile(vendorPath)
		if err != nil {
			continue
		}
		// 0x1002 = AMD
		if strings.TrimSpace(string(data)) == "0x1002" {
			return cardPath, nil
		}
	}
	return "", fmt.Errorf("no AMD GPU found under %s", drmBasePath)
}

func readFloat(path string) (float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
}

func readInt(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}

// readTempInput reads temp1_input from the hwmon subdirectory.
func readTempInput(hwmonBase string) (int64, error) {
	matches, err := filepath.Glob(filepath.Join(hwmonBase, "hwmon*", "temp1_input"))
	if err != nil || len(matches) == 0 {
		return 0, fmt.Errorf("temp1_input not found under %s", hwmonBase)
	}
	return readInt(matches[0])
}
