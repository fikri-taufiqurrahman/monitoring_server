package battery

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const sysfs = "/sys/class/power_supply"

type Service struct {
	lastCheck time.Time
}

func NewService() *Service {
	return &Service{}
}

// --- FUNGSI INTERNAL ---

func readFloat(path, filename string) (float64, error) {
	data, err := os.ReadFile(filepath.Join(path, filename))
	if err != nil {
		return 0, err
	}
	str := strings.TrimSpace(string(data))
	num, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, err
	}
	return num / 1000, nil
}

func isBattery(path string) bool {
	data, err := os.ReadFile(filepath.Join(path, "type"))
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "Battery"
}

func getBatteryPaths() ([]string, error) {
	files, err := os.ReadDir(sysfs)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, file := range files {
		path := filepath.Join(sysfs, file.Name())
		if isBattery(path) {
			paths = append(paths, path)
		}
	}
	return paths, nil
}

// --- PUBLIC METHODS (Yang dibutuhkan handler.go) ---
func (s *Service) GetBatteryInfo() (BatteryInfo, error) {
	paths, err := getBatteryPaths()
	if err != nil || len(paths) == 0 {
		return BatteryInfo{}, fmt.Errorf("no battery found")
	}
	path := paths[0]

	var current, full, design, voltage float64 // voltage sudah dideklarasikan di sini
	var statusStr string

	// 1. Gunakan 'voltage' langsung, pakai '=' karena sudah di-declare di atas
	voltage, _ = readFloat(path, "voltage_now")
	voltageV := voltage / 1000

	energyNow, err := readFloat(path, "energy_now")
	if err == nil {
		current = energyNow
		full, _ = readFloat(path, "energy_full")
		design, _ = readFloat(path, "energy_full_design")
	} else {
		chargeNow, _ := readFloat(path, "charge_now")
		current = chargeNow * voltageV
		chargeFull, _ := readFloat(path, "charge_full")
		full = chargeFull * voltageV
		chargeDesign, _ := readFloat(path, "charge_full_design")
		design = chargeDesign * voltageV
	}

	statusData, err := os.ReadFile(filepath.Join(path, "status"))
	if err == nil {
		statusStr = strings.TrimSpace(string(statusData))
		if statusStr == "Not charging" {
			statusStr = "Idle"
		}
	} else {
		statusStr = "Unknown"
	}

	var percentage, health float64
	if full > 0 {
		percentage = (current / full) * 100
	}
	if design > 0 {
		health = (full / design) * 100
	}

	s.lastCheck = time.Now()

	return BatteryInfo{
		Percentage:   percentage,
		Status:       statusStr,
		EnergyNow:    int64(current),
		EnergyFull:   int64(full),
		EnergyDesign: int64(design),
		Voltage:      int64(voltage), // 2. Pakai 'voltage' di sini
		Health:       health,
	}, nil
}
func (s *Service) GetBatteryPercentage() (float64, error) {
	info, err := s.GetBatteryInfo()
	if err != nil {
		return 0, err
	}
	return info.Percentage, nil
}

func (s *Service) GetBatteryStatus() (string, error) {
	info, err := s.GetBatteryInfo()
	if err != nil {
		return "", err
	}
	return info.Status, nil
}

func (s *Service) GetBatteryHealth() (float64, error) {
	info, err := s.GetBatteryInfo()
	if err != nil {
		return 0, err
	}
	return info.Health, nil
}

func (s *Service) IsCharging() (bool, error) {
	info, err := s.GetBatteryInfo()
	if err != nil {
		return false, err
	}
	return info.Status == "Charging", nil
}

func (s *Service) GetLastCheck() time.Time {
	return s.lastCheck
}
