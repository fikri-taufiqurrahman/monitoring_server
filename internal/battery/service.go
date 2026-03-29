package battery

import (
	"fmt"
	"time"

	"github.com/distatus/battery"
)

type Service struct {
	lastCheck time.Time
}

func NewService() *Service {
	return &Service{}
}

// GetBatteryInfo mengambil info baterai lengkap
func (s *Service) GetBatteryInfo() (BatteryInfo, error) {
	batteries, err := battery.GetAll()
	if err != nil {
		return BatteryInfo{}, fmt.Errorf("failed to get battery info: %v", err)
	}

	if len(batteries) == 0 {
		return BatteryInfo{}, fmt.Errorf("no battery found")
	}

	// Ambil baterai pertama
	bat := batteries[0]

	// Hitung persentase (hindari division by zero)
	var percentage float64
	if bat.Full > 0 {
		percentage = (bat.Current / bat.Full) * 100
	}

	info := BatteryInfo{
		Percentage:   percentage,
		Status:       bat.State.String(), // Gunakan method String() dari State
		EnergyNow:    int64(bat.Current),
		EnergyFull:   int64(bat.Full),
		EnergyDesign: int64(bat.Design),
		Voltage:      int64(bat.Voltage * 1000), // Konversi dari Volt ke mV
	}

	// Hitung health baterai
	if info.EnergyDesign > 0 {
		info.Health = (float64(info.EnergyFull) / float64(info.EnergyDesign)) * 100
	}

	s.lastCheck = time.Now()
	return info, nil
}

// GetBatteryPercentage mengambil persentase baterai saja
func (s *Service) GetBatteryPercentage() (float64, error) {
	info, err := s.GetBatteryInfo()
	if err != nil {
		return 0, err
	}
	return info.Percentage, nil
}

// GetBatteryStatus mengambil status baterai
func (s *Service) GetBatteryStatus() (string, error) {
	info, err := s.GetBatteryInfo()
	if err != nil {
		return "", err
	}
	return info.Status, nil
}

// IsCharging mengecek apakah baterai sedang dicharge
func (s *Service) IsCharging() (bool, error) {
	info, err := s.GetBatteryInfo()
	if err != nil {
		return false, err
	}
	return info.Status == "Charging", nil
}

// GetBatteryHealth mendapatkan health baterai dalam persen
func (s *Service) GetBatteryHealth() (float64, error) {
	info, err := s.GetBatteryInfo()
	if err != nil {
		return 0, err
	}
	return info.Health, nil
}

// GetLastCheck mengembalikan waktu terakhir pengecekan
func (s *Service) GetLastCheck() time.Time {
	return s.lastCheck
}
