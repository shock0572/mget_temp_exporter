# mget_temp Prometheus Exporter

A Prometheus exporter for NVIDIA/Mellanox network adapter temperature monitoring using the `mget_temp` utility from NVIDIA Firmware Tools (MFT).

## What This Does

This exporter polls comprehensive temperature data from NVIDIA/Mellanox network adapters using two commands per device and exposes the metrics in Prometheus format. It provides:

- **Device temperatures** (`mget_temp`): Main temperature readings from network adapters (direct reading from mget_temp -d DEVICE)
- **Thermal diode temperatures** (`mget_thermal_diode_temp_celsius`): Individual thermal sensor temperature readings in Celsius, including maximum allowed temperature as a label
- **Thermal diode voltages** (`mget_thermal_diode_voltage_volts`): Individual thermal sensor voltage readings in Volts, including maximum allowed voltage as a label

The exporter uses two commands per device:
- `mget_temp -d device` for the main temperature reading
- `mget_temp -d device -v` for detailed thermal diode data

It automatically distinguishes between temperature (T) and voltage (V) measurements and creates appropriate metrics for each type, with thresholds included as labels.

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

- `mget_temp{device="device_name"}`: Main temperature reading from the network adapter (direct reading from mget_temp -d DEVICE)
- `mget_thermal_diode_temp_celsius{device="device_name", diode="diode_name", threshold="max_temp"}`: Temperature from individual thermal diodes in Celsius. The threshold label contains the maximum allowed temperature as an unsigned integer.
- `mget_thermal_diode_voltage_volts{device="device_name", diode="diode_name", threshold="max_voltage"}`: Voltage from individual thermal diodes in Volts. The threshold label contains the maximum allowed voltage as an unsigned integer.

## How It Works

The exporter efficiently gathers temperature data using two commands per device:

1. **Main Temperature**: Runs `mget_temp -d device` to get the main device temperature
2. **Thermal Diode Data**: Runs `mget_temp -d device -v` to get detailed thermal diode information
3. **Data Parsing**: Parses the tabular output to extract thermal diode information

This approach ensures we get the most accurate main temperature reading directly from the device while still maintaining detailed thermal diode monitoring.

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

