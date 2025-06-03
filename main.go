package main

import (
	"bufio"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	version   = "0.2"
	buildDate = "2025-06-03"
	goVersion = runtime.Version()
)

// Custom registry to control metric ordering
var customRegistry = prometheus.NewRegistry()

// Map to hold a Gauge for each device
var mgetTemps = map[string]prometheus.Gauge{}

// Maps to hold thermal diode metrics
var thermalDiodeTemps = map[string]prometheus.Gauge{}
var thermalDiodeThresholds = map[string]prometheus.Gauge{}
var thermalDiodeVoltages = map[string]prometheus.Gauge{}
var thermalDiodeVoltageThresholds = map[string]prometheus.Gauge{}

// Reads devices from devices.cfg (one per line)
func readDevicesConfig(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var devices []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and lines that start with #
		if line != "" && !strings.HasPrefix(line, "#") {
			devices = append(devices, line)
		}
	}
	return devices, scanner.Err()
}

// createTemperatureMetrics creates and registers temperature metrics for a thermal diode
func createTemperatureMetrics(device, diodeName, diodeKey string) {
	if _, exists := thermalDiodeTemps[diodeKey]; !exists {
		tempGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "mget_thermal_diode_temp_celsius",
			Help:        "Temperature from thermal diode in Celsius",
			ConstLabels: prometheus.Labels{"device": device, "diode": diodeName},
		})
		customRegistry.MustRegister(tempGauge)
		thermalDiodeTemps[diodeKey] = tempGauge

		threshGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "mget_thermal_diode_temp_threshold_celsius",
			Help:        "Temperature threshold for thermal diode in Celsius",
			ConstLabels: prometheus.Labels{"device": device, "diode": diodeName},
		})
		customRegistry.MustRegister(threshGauge)
		thermalDiodeThresholds[diodeKey] = threshGauge
	}
}

// createVoltageMetrics creates and registers voltage metrics for a thermal diode
func createVoltageMetrics(device, diodeName, diodeKey string) {
	if _, exists := thermalDiodeVoltages[diodeKey]; !exists {
		voltageGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "mget_thermal_diode_voltage_volts",
			Help:        "Voltage from thermal diode in Volts",
			ConstLabels: prometheus.Labels{"device": device, "diode": diodeName},
		})
		customRegistry.MustRegister(voltageGauge)
		thermalDiodeVoltages[diodeKey] = voltageGauge

		voltageThreshGauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "mget_thermal_diode_voltage_threshold_volts",
			Help:        "Voltage threshold for thermal diode in Volts",
			ConstLabels: prometheus.Labels{"device": device, "diode": diodeName},
		})
		customRegistry.MustRegister(voltageThreshGauge)
		thermalDiodeVoltageThresholds[diodeKey] = voltageThreshGauge
	}
}

// updateTemperatureMetrics updates temperature metrics with new values
func updateTemperatureMetrics(diodeKey string, temp, threshold float64) {
	thermalDiodeTemps[diodeKey].Set(temp)
	thermalDiodeThresholds[diodeKey].Set(threshold)
}

// updateVoltageMetrics updates voltage metrics with new values
func updateVoltageMetrics(diodeKey string, voltage, threshold float64) {
	thermalDiodeVoltages[diodeKey].Set(voltage)
	thermalDiodeVoltageThresholds[diodeKey].Set(threshold)
}

// parseThermalDiodeData parses a single line of thermal diode data and updates metrics
func parseThermalDiodeData(device, line string, dataRegex *regexp.Regexp) {
	matches := dataRegex.FindStringSubmatch(line)
	if len(matches) != 5 {
		return // Skip lines that don't match the expected format
	}

	diodeName := matches[1]
	measurementType := matches[2] // T for temperature, V for voltage
	valueStr := matches[3]
	threshStr := matches[4]

	// Parse the numeric values
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		slog.Error("Error parsing value for thermal diode", "device", device, "diode", diodeName, "error", err)
		return
	}

	threshold, err := strconv.ParseFloat(threshStr, 64)
	if err != nil {
		slog.Error("Error parsing threshold for thermal diode", "device", device, "diode", diodeName, "error", err)
		return
	}

	// Create unique key combining device and diode name
	diodeKey := device + "_" + diodeName

	// Handle temperature vs voltage measurements
	switch measurementType {
	case "T":
		createTemperatureMetrics(device, diodeName, diodeKey)
		updateTemperatureMetrics(diodeKey, value, threshold)

		// If this is the iopx diode, also set the main temperature gauge
		if diodeName == "iopx" {
			if gauge, exists := mgetTemps[device]; exists {
				gauge.Set(value)
			}
		}
	case "V":
		createVoltageMetrics(device, diodeName, diodeKey)
		updateVoltageMetrics(diodeKey, value, threshold)
	default:
		slog.Error("Unknown measurement type", "device", device, "diode", diodeName, "measurement_type", measurementType)
	}
}

// pollSingleDevice polls both thermal diode data and extracts main temperature for a single device
func pollSingleDevice(device string) {
	cmd := exec.Command("mget_temp", "-d", device, "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("Error running mget_temp -d <device> -v", "device", device, "error", err, "output", string(output), "output_length", len(output))

		// Test if mget_temp exists by trying to run it without args
		testCmd := exec.Command("mget_temp")
		_, _ = testCmd.CombinedOutput()
		return
	}

	// Parse the tabular output
	lines := strings.Split(string(output), "\n")
	// Regex to match the data lines (skip header and empty lines)
	dataRegex := regexp.MustCompile(`^\s*\d+\s+(\S+)\s+([TV])\s+(\d+(?:\.\d+)?)\s+(\d+(?:\.\d+)?)`)

	for _, line := range lines {
		parseThermalDiodeData(device, line, dataRegex)
	}
}

// Polls all device data using mget_temp -d <device> -v (replaces both previous polling functions)
func pollDevices(devices []string) {
	for {
		for _, device := range devices {
			go pollSingleDevice(device)
		}
		time.Sleep(10 * time.Second)
	}
}

func main() {
	// Configure slog with text handler that includes timestamps and source
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))

	// Parse command line flags
	port := flag.String("port", "6656", "Port to listen on")
	flag.Parse()

	devices, err := readDevicesConfig("devices.cfg")
	if err != nil {
		slog.Error("Failed to load devices.cfg", "error", err)
		os.Exit(1)
	}

	// Start polling all device data
	go pollDevices(devices)

	// Register a Gauge for each device last (these will appear at the end)
	for _, dev := range devices {
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "mget_temp",
			Help:        "Main temperature reading from the network adapter (extracted from iopx thermal diode)",
			ConstLabels: prometheus.Labels{"device": dev},
		})
		customRegistry.MustRegister(gauge)
		mgetTemps[dev] = gauge
	}

	// Handler for root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("mget_exporter<br><br>Metrics are at <a href=\"/metrics\">/metrics</a>\n"))
	})

	// Use custom registry with the HTTP handler
	http.Handle("/metrics", promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{}))

	slog.Info("Starting mget_exporter",
		"version", version,
		"goversion", goVersion,
		"builddate", buildDate,
		"port", *port,
		"devices", len(devices),
		"maxprocs", runtime.GOMAXPROCS(0))

	slog.Info("Listening on", "address", ":"+*port)

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		slog.Error("Failed to start HTTP server", "error", err)
		os.Exit(1)
	}
}
