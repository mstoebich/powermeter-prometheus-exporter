package main

import (
	"encoding/binary"
	"log"
	"net/http"
	"os"
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
		Name: "powermeter_frequency_hz",
		Help: "frequency in Hz",
	})
	voltageL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "powermeter_voltage_l1_v",
		Help: "voltage L1 in Volt",
	})
	currentL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "powermeter_current_l1_a",
		Help: "current L1 in Ampere",
	})
	activePowerL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "powermeter_power_active_w",
		Help: "active power in W",
	})
	reactivePowerL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "powermeter_power_reactive_var",
		Help: "reactive power in VAr",
	})
	apparentPowerL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "powermeter_power_apparent_va",
		Help: "apparent power in VA",
	})
	powerFactorL1 = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "powermeter_power_factor",
		Help: "power factor",
	})
	totalEnergy = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "powermeter_energy_total_kwh",
		Help: "total energy in kWh",
	})
)

func init() {
	prometheus.MustRegister(frequency, voltageL1, currentL1, activePowerL1, reactivePowerL1, apparentPowerL1, powerFactorL1, totalEnergy)
}

func main() {
	powermeter_conn, ok := os.LookupEnv("POWERMETER_CONN")
	if !ok {
		log.Fatal("POWERMETER_CONN environment variable not set")
	}

	http.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler := modbus.NewTCPClientHandler(powermeter_conn)
		handler.Timeout = 5 * time.Second
		handler.SlaveID = 1
		if err := handler.Connect(); err != nil {
			log.Printf("Modbus connection error: %v", err)
			http.Error(w, "Modbus connection failed", http.StatusInternalServerError)
			return
		}
		defer handler.Close()

		client := modbus.NewClient(handler)
		collectMetrics(client)
		promhttp.Handler().ServeHTTP(w, r)
	}))

	log.Println("Exporter läuft auf :9100/metrics")
	log.Fatal(http.ListenAndServe(":9100", nil))
}

func collectMetrics(client modbus.Client) {
	data, err := ReadPowerMetrics(client)
	if err != nil {
		log.Println("Modbus read error:", err)
		return
	}

	// log.Printf("Voltage L1: %.2f V\n", data.VoltageL1)
	// log.Printf("Current L1: %.2f A\n", data.CurrentL1)
	// log.Printf("Active Power L1: %.2f W\n", data.ActivePowerL1)
	// log.Printf("Reactive Power L1: %.2f Var\n", data.ReactivePowerL1)
	// log.Printf("Apparent Power L1: %.2f VA\n", data.ApparentPowerL1)
	// log.Printf("Power Factor L1: %.2f\n", data.PowerFactorL1)
	// log.Printf("Frequency: %.2f Hz\n", data.Frequency)
	// log.Printf("Energy Total: %.2f kWh\n", data.EnergyTotal)

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
		EnergyTotal:     float64(energyTotalRaw) * 0.01,
	}, nil
}
