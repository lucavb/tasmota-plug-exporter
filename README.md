# Tasmota Plug Exporter

Prometheus exporter for Tasmota smart plugs with energy monitoring.

## Metrics

| Metric | Description |
|--------|-------------|
| `tasmota_up` | Target reachable (1=up, 0=down) |
| `tasmota_power_watts` | Current power consumption |
| `tasmota_voltage_volts` | Voltage |
| `tasmota_current_amps` | Current |
| `tasmota_energy_total_kwh` | Total energy consumed |
| `tasmota_energy_today_kwh` | Energy consumed today |
| `tasmota_energy_yesterday_kwh` | Energy consumed yesterday |
| `tasmota_power_factor` | Power factor |
| `tasmota_apparent_power_va` | Apparent power |
| `tasmota_reactive_power_var` | Reactive power |
| `tasmota_relay_state` | Relay state (1=on, 0=off) |
| `tasmota_uptime_seconds` | Device uptime |
| `tasmota_wifi_rssi_percent` | WiFi RSSI percentage |
| `tasmota_wifi_signal_dbm` | WiFi signal in dBm |

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `TASMOTA_TARGETS` | (required) | Comma-separated list of Tasmota device addresses |
| `LISTEN_ADDRESS` | `:9184` | HTTP listen address |
| `SCRAPE_TIMEOUT` | `5s` | Timeout for Tasmota API requests |

## Usage

### Docker Compose

```yaml
services:
  tasmota-exporter:
    image: ghcr.io/lucavb/tasmota-plug-exporter:latest
    environment:
      - TASMOTA_TARGETS=172.16.0.81,172.16.0.82
    ports:
      - "9184:9184"
    restart: unless-stopped
```

The image includes a built-in healthcheck on `/health`.

### Prometheus Configuration

```yaml
scrape_configs:
  - job_name: tasmota
    static_configs:
      - targets: ["tasmota-exporter:9184"]
```

### Binary

```bash
TASMOTA_TARGETS=172.16.0.81 ./tasmota-plug-exporter
```

## Build

```bash
go build -o tasmota-plug-exporter .
```

## License

MIT
