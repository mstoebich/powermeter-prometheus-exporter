package main

import (
	"encoding/binary"
	"net/http"

	// "fmt"
	"log"
	"time"

	"github.com/grid-x/modbus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PowerMetrics struct {
	Frequency       float64 // Hz
	VoltageL1       float64 // V
	CurrentL1       float64 // A
	ActivePowerL1   float64 // kW
	ReactivePowerL1 float64 // kvar
	ApparentPowerL1 float64 // kVA
	PowerFactorL1   float64 // -
	EnergyTotal     float64 // kWh
}

const (
	frequencyRegister       uint16 = 0x0130 // Register-Adresse für Frequenz
	voltageL1Register       uint16 = 0x0131 // Register-Adresse für Spannung Phase L1
	currentL1Register       uint16 = 0x0139 // Register-Adresse für Strom Phase L1
	activePowerL1Register   uint16 = 0x0140 // Register-Adresse für Wirkleistung Phase L1
	reactivePowerL1Register uint16 = 0x0148 // Register-Adresse für Blindleistung Phase L1
	apparentPowerL1Register uint16 = 0x0150 // Register-Adresse für Scheinleistung Phase L1
	powerFactorL1Register   uint16 = 0x0158 // Register-Adresse für Leistungsfaktor Phase L1
	energyTotalRegister     uint16 = 0xA000 // Register-Adresse für Gesamternergie

)

var (
	frequency = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "stromzaehler_frequency_hz",
		Help: "Netzfrequenz in Hz",
	})
	voltageL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "stromzaehler_voltage_l1_v",
		Help: "Spannung L1 in Volt",
	})
	currentL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "stromzaehler_current_l1_a",
		Help: "Strom L1 in Ampere",
	})
	activePowerL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "stromzaehler_power_active_kw",
		Help: "Wirkleistung in kW",
	})
	reactivePowerL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "stromzaehler_power_reactive_kvar",
		Help: "Blindleistung in kvar",
	})
	apparentPowerL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "stromzaehler_power_apparent_kva",
		Help: "Scheinleistung in kVA",
	})
	powerFactorL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "stromzaehler_power_factor",
		Help: "Leistungsfaktor",
	})
	totalEnergy = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "stromzaehler_energy_total_kwh",
		Help: "Gesamtenergieverbrauch in kWh",
	})
)

func init() {
	prometheus.MustRegister(frequency, voltageL1, currentL1, activePowerL1, reactivePowerL1, apparentPowerL1, powerFactorL1, totalEnergy)
}

func main() {
	handler := modbus.NewTCPClientHandler("192.168.1.200:502") // IP und Port deines Gateways
	handler.Timeout = 5 * time.Second
	handler.SlaveID = 1 // Modbus-Adresse des Stromzählers (z. B. 1)
	err := handler.Connect()
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer handler.Close()

	client := modbus.NewClient(handler)

	go func() {
		for {
			collectMetrics(client)
			time.Sleep(10 * time.Second)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Exporter läuft auf :9100/metrics")
	log.Fatal(http.ListenAndServe(":9100", nil))
	// powerMetrics, err := ReadPowerMetrics(client)

	// if err != nil {
	// 	log.Fatalf("Failed to read power metrics: %v", err)
	// } else {
	// 	fmt.Printf("Voltage L1: %.2f V\n", powerMetrics.VoltageL1)
	// 	fmt.Printf("Current L1: %.2f A\n", powerMetrics.CurrentL1)
	// 	fmt.Printf("Active Power L1: %.2f kW\n", powerMetrics.ActivePowerL1)
	// 	fmt.Printf("Reactive Power L1: %.2f Var\n", powerMetrics.ReactivePowerL1)
	// 	fmt.Printf("Apparent Power L1: %.2f VA\n", powerMetrics.ApparentPowerL1)
	// 	fmt.Printf("Power Factor L1: %.2f\n", powerMetrics.PowerFactorL1)
	// 	fmt.Printf("Frequency: %.2f Hz\n", powerMetrics.Frequency)
	// 	fmt.Printf("Energy Total: %.2f kWh\n", powerMetrics.EnergyTotal)
	// }

}

func collectMetrics(client modbus.Client) {
	data, err := ReadPowerMetrics(client)
	if err != nil {
		log.Println("Modbus read error:", err)
		return
	}

	frequency.Set(data.Frequency)
	voltageL1.Set(data.VoltageL1)
	currentL1.Set(data.CurrentL1)
	activePowerL1.Set(data.ActivePowerL1)
	reactivePowerL1.Set(data.ReactivePowerL1)
	apparentPowerL1.Set(data.ApparentPowerL1)
	powerFactorL1.Set(data.PowerFactorL1)
	totalEnergy.Set(data.EnergyTotal)
}

func ReadPowerMetrics(client modbus.Client) (*PowerMetrics, error) {
	readU16 := func(addr uint16) (uint16, error) {
		data, err := client.ReadHoldingRegisters(addr, 1)
		if err != nil {
			return 0, err
		}
		return binary.BigEndian.Uint16(data), nil
	}

	readU32 := func(addr uint16) (uint32, error) {
		data, err := client.ReadHoldingRegisters(addr, 2)
		if err != nil {
			return 0, err
		}
		return binary.BigEndian.Uint32(data), nil
	}

	frequencyRaw, err := readU16(frequencyRegister)
	if err != nil {
		return nil, err
	}

	voltageRaw, err := readU16(voltageL1Register)
	if err != nil {
		return nil, err
	}

	currentRaw, err := readU32(currentL1Register)
	if err != nil {
		return nil, err
	}

	activePowerL1Raw, err := readU32(activePowerL1Register)
	if err != nil {
		return nil, err
	}

	reactivePowerL1Raw, err := readU32(reactivePowerL1Register)
	if err != nil {
		return nil, err
	}

	apparentPowerL1Raw, err := readU32(apparentPowerL1Register)
	if err != nil {
		return nil, err
	}

	powerFactorL1Raw, err := readU16(powerFactorL1Register)
	if err != nil {
		return nil, err
	}

	energyTotalRaw, err := readU32(energyTotalRegister)
	if err != nil {
		return nil, err
	}

	return &PowerMetrics{
		Frequency:       float64(frequencyRaw) * 0.01,
		VoltageL1:       float64(voltageRaw) * 0.01,
		CurrentL1:       float64(currentRaw) * 0.001,
		ActivePowerL1:   float64(activePowerL1Raw),
		ReactivePowerL1: float64(reactivePowerL1Raw),
		ApparentPowerL1: float64(apparentPowerL1Raw),
		PowerFactorL1:   float64(powerFactorL1Raw) * 0.001,
		EnergyTotal:     float64(energyTotalRaw) * 0.001,
	}, nil
}
