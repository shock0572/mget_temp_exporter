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

The exporter automatically discovers MST devices using `mst status` command, polls temperature data every 10 seconds, and exposes only custom metrics (no Go runtime or process metrics) at the `/metrics` endpoint on port 6656 by default (configurable).

## Requirements

- Go 1.22.2
- NVIDIA Firmware Tools (MFT) must be installed and available in your PATH
- NVIDIA/Mellanox network adapters
- **Superuser/Administrator privileges** - The exporter must be run with elevated privileges as the `mget_temp` utility requires superuser access

### Installing NVIDIA Firmware Tools

Download and install NVIDIA Firmware Tools from:
https://network.nvidia.com/products/adapter-software/firmware-tools/

Ensure both the `mget_temp` and `mst` utilities are available in your system PATH.

## Device Discovery

The exporter automatically discovers MST devices using the `mst status` command. No manual configuration is required - it will find and monitor all available MST devices.

⚠️ **Important for Linux:** You must run `mst start` first before device discovery will work:
```bash
sudo mst start
```

**Manual Device Specification (Optional):**

If you need to specify devices manually, use the `-devices` flag:
```bash
# Linux
sudo ./mget_exporter -devices "/dev/mst/mt4127_pciconf0,/dev/mst/mt4128_pciconf0"

# Windows
mget_exporter.exe -devices "mt4115_pciconf0,mt4116_pciconf0"
```

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
- `-devices string`: Comma-separated list of device IDs (optional - if not specified, will auto-discover using `mst status` - ⚠️ On Linux you will have to run `mst start` first )

### Linux

1. **First, start the MST service:**

   ```bash
   sudo mst start
   ```

2. Run the exporter with sudo (automatic device discovery):

   ```bash
   sudo ./mget_exporter
   ```
   
   Or with a custom port:
   ```bash
   sudo ./mget_exporter -port 8080
   ```

   Or with manual device specification:
   ```bash
   sudo ./mget_exporter -devices "/dev/mst/mt4127_pciconf0,/dev/mst/mt4128_pciconf0"
   ```

3. Access metrics at `http://localhost:6656/metrics` (or your custom port)

### Windows

1. Open Command Prompt or PowerShell **as Administrator**
2. Run the exporter (automatic device discovery):
   ```cmd
   mget_exporter.exe
   ```
   
   Or with a custom port:
   ```cmd
   mget_exporter.exe -port 8080
   ```

   Or with manual device specification:
   ```cmd
   mget_exporter.exe -devices "mt4115_pciconf0,mt4116_pciconf0"
   ```

3. Access metrics at `http://localhost:6656/metrics` (or your custom port)

**Note:** On Windows, you must run the entire terminal session as Administrator before executing the exporter.

## Metrics

The exporter provides the following Prometheus metrics (only custom metrics, no Go runtime or process metrics):

- `mget_temp{device="device_name"}`: Main temperature reading from the network adapter (direct reading from mget_temp -d DEVICE)
- `mget_thermal_diode_temp_celsius{device="device_name", diode="diode_name", threshold="max_temp"}`: Temperature from individual thermal diodes in Celsius. The threshold label contains the maximum allowed temperature as an unsigned integer.
- `mget_thermal_diode_voltage_volts{device="device_name", diode="diode_name", threshold="max_voltage"}`: Voltage from individual thermal diodes in Volts. The threshold label contains the maximum allowed voltage as an unsigned integer.

## How It Works

The exporter efficiently gathers temperature data using the following approach:

1. **Main Temperature**: Runs `mget_temp -d device` to get the main device temperature for each discovered device
2. **Thermal Diode Data**: Runs `mget_temp -d device -v` to get detailed thermal diode information for each device
3. **Data Parsing**: Parses the tabular output to extract thermal diode information
4. **Parallel Processing**: Each device is monitored in parallel for optimal performance

This approach ensures automatic discovery of all available devices while maintaining the most accurate temperature readings.

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

