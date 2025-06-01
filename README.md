# mget_temp Prometheus Exporter

A Prometheus exporter for NVIDIA/Mellanox network adapter temperature monitoring using the `mget_temp` utility from NVIDIA Firmware Tools (MFT).

## What This Does

This exporter polls comprehensive temperature data from NVIDIA/Mellanox network adapters using a single command per device and exposes the metrics in Prometheus format. It provides:

- **Device temperatures** (`mget_temp`): Main temperature readings from network adapters (extracted from iopx thermal diode)
- **Thermal diode temperatures** (`mget_thermal_diode_temp_celsius`): Individual thermal sensor temperature readings in Celsius
- **Thermal diode temperature thresholds** (`mget_thermal_diode_temp_threshold_celsius`): Temperature thresholds for thermal sensors in Celsius
- **Thermal diode voltages** (`mget_thermal_diode_voltage_volts`): Individual thermal sensor voltage readings in Volts
- **Thermal diode voltage thresholds** (`mget_thermal_diode_voltage_threshold_volts`): Voltage thresholds for thermal sensors in Volts

The exporter uses a single `mget_temp -d device -v` command per device to gather both detailed thermal diode data and main temperature (from the iopx diode). It automatically distinguishes between temperature (T) and voltage (V) measurements and creates appropriate metrics for each type.

The exporter reads device configurations from `devices.cfg`, polls temperature data every 10 seconds, and exposes only custom metrics (no Go runtime or process metrics) at the `/metrics` endpoint on port 6656 by default (configurable).

## Requirements

- Go 1.22.2
- NVIDIA Firmware Tools (MFT) must be installed and available in your PATH
- NVIDIA/Mellanox network adapters
- **Superuser/Administrator privileges** - The exporter must be run with elevated privileges as the `mget_temp` utility requires superuser access

### Installing NVIDIA Firmware Tools

Download and install NVIDIA Firmware Tools from:
https://network.nvidia.com/products/adapter-software/firmware-tools/

Ensure the `mget_temp` utility is available in your system PATH.

## Configuration

Edit the `devices.cfg` file to specify your network adapter devices (one per line):

**Linux format (full path):**
```
/dev/mst/mt4127_pciconf0
/dev/mst/mt4127_pciconf1
```

**Windows format (short name):**
```
mt4127_pciconf0
mt4127_pciconf1
```

Use the appropriate format for your operating system. On Linux, devices are typically accessed via `/dev/mst/` paths, while on Windows, you can use the short device names directly.

## Compilation

### Linux

To build the exporter:

```bash
go build -o mget_exporter main.go
```

Or to build and install:

```bash
go install
```

### Windows

To build the exporter:

```cmd
go build -o mget_exporter.exe main.go
```

Or to build and install:

```cmd
go install
```

## Usage

⚠️ **IMPORTANT: The exporter must be run with superuser/administrator privileges!**

### Command Line Options

- `-port string`: Port to listen on (default "6656")

### Linux

1. Configure your devices in `devices.cfg`
2. Run the exporter with sudo:
   ```bash
   sudo ./mget_exporter
   ```
   
   Or with a custom port:
   ```bash
   sudo ./mget_exporter -port 8080
   ```
3. Access metrics at `http://localhost:6656/metrics` (or your custom port)

### Windows

1. Configure your devices in `devices.cfg`
2. Open Command Prompt or PowerShell **as Administrator**
3. Run the exporter:
   ```cmd
   mget_exporter.exe
   ```
   
   Or with a custom port:
   ```cmd
   mget_exporter.exe -port 8080
   ```
4. Access metrics at `http://localhost:6656/metrics` (or your custom port)

**Note:** On Windows, you must run the entire terminal session as Administrator before executing the exporter.

## Metrics

The exporter provides the following Prometheus metrics (only custom metrics, no Go runtime or process metrics):

- `mget_temp{device="device_name"}`: Main temperature reading from the network adapter (extracted from iopx thermal diode)
- `mget_thermal_diode_temp_celsius{device="device_name", diode="diode_name"}`: Temperature from individual thermal diodes in Celsius
- `mget_thermal_diode_temp_threshold_celsius{device="device_name", diode="diode_name"}`: Temperature thresholds for thermal diodes in Celsius
- `mget_thermal_diode_voltage_volts{device="device_name", diode="diode_name"}`: Voltage from individual thermal diodes in Volts
- `mget_thermal_diode_voltage_threshold_volts{device="device_name", diode="diode_name"}`: Voltage thresholds for thermal diodes in Volts

## How It Works

The exporter efficiently gathers all temperature data using a single command per device:

1. **Single Command Execution**: Runs `mget_temp -d device -v` for each configured device
2. **Data Parsing**: Parses the tabular output to extract thermal diode information
3. **Main Temperature Extraction**: Automatically extracts the main device temperature from the "iopx" thermal diode

This approach is more efficient than running separate commands and ensures consistency between the main temperature metric and detailed thermal diode data.

## Copyright and Legal Notice

This software uses and integrates with proprietary NVIDIA tools and technologies:

- **NVIDIA** and **Mellanox** are trademarks of NVIDIA Corporation
- **mget_temp** utility is part of NVIDIA Firmware Tools (MFT) and is copyrighted by NVIDIA Corporation
- **NVIDIA Firmware Tools** are copyrighted by NVIDIA Corporation

This exporter is a third-party tool that interfaces with NVIDIA's MFT utilities and is not officially endorsed by NVIDIA Corporation.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

**Important:** While this exporter code is open source under the MIT License, it depends on proprietary NVIDIA Firmware Tools (MFT):

- This exporter code itself is free to use and modify under the MIT License
- **NVIDIA Firmware Tools (MFT) are proprietary software** - ensure compliance with NVIDIA's licensing terms
- The `mget_temp` utility and other MFT components are subject to NVIDIA's license agreements
- Users must obtain and install NVIDIA Firmware Tools separately from official NVIDIA sources

By using this exporter, you acknowledge that you will comply with all applicable NVIDIA license terms for the MFT tools.

