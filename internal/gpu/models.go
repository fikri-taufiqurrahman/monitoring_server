package gpu

type GPUInfo struct {
	BusyPercent float64 `json:"busy_percent"`
	VRAMUsed    int64   `json:"vram_used"`
	VRAMTotal   int64   `json:"vram_total"`
	TempCelsius float64 `json:"temp_celsius"`
}
