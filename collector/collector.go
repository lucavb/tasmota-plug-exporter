package collector

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/lucavb/tasmota-plug-exporter/tasmota"
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "tasmota"

// Collector implements prometheus.Collector for Tasmota devices.
type Collector struct {
	client  *tasmota.Client
	targets []string
	timeout time.Duration

	// Metrics
	up             *prometheus.Desc
	powerWatts     *prometheus.Desc
	voltageVolts   *prometheus.Desc
	currentAmps    *prometheus.Desc
	energyTotalKwh *prometheus.Desc
	energyTodayKwh *prometheus.Desc
	energyYestKwh  *prometheus.Desc
	powerFactor    *prometheus.Desc
	apparentPower  *prometheus.Desc
	reactivePower  *prometheus.Desc
	relayState     *prometheus.Desc
	uptimeSeconds  *prometheus.Desc
	wifiRSSI       *prometheus.Desc
	wifiSignalDBm  *prometheus.Desc
}

// New creates a new Collector.
func New(client *tasmota.Client, targets []string, timeout time.Duration) *Collector {
	labels := []string{"address", "device"}

	return &Collector{
		client:  client,
		targets: targets,
		timeout: timeout,

		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Whether the Tasmota device is reachable",
			[]string{"address"}, nil,
		),
		powerWatts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "power_watts"),
			"Current power consumption in watts",
			labels, nil,
		),
		voltageVolts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "voltage_volts"),
			"Current voltage in volts",
			labels, nil,
		),
		currentAmps: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "current_amps"),
			"Current in amperes",
			labels, nil,
		),
		energyTotalKwh: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_total_kwh"),
			"Total energy consumed in kWh",
			labels, nil,
		),
		energyTodayKwh: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_today_kwh"),
			"Energy consumed today in kWh",
			labels, nil,
		),
		energyYestKwh: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "energy_yesterday_kwh"),
			"Energy consumed yesterday in kWh",
			labels, nil,
		),
		powerFactor: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "power_factor"),
			"Power factor (0-1)",
			labels, nil,
		),
		apparentPower: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "apparent_power_va"),
			"Apparent power in VA",
			labels, nil,
		),
		reactivePower: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "reactive_power_var"),
			"Reactive power in VAR",
			labels, nil,
		),
		relayState: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "relay_state"),
			"Relay state (1=on, 0=off)",
			labels, nil,
		),
		uptimeSeconds: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "uptime_seconds"),
			"Device uptime in seconds",
			labels, nil,
		),
		wifiRSSI: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "wifi_rssi_percent"),
			"WiFi RSSI as percentage",
			labels, nil,
		),
		wifiSignalDBm: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "wifi_signal_dbm"),
			"WiFi signal strength in dBm",
			labels, nil,
		),
	}
}

// Describe implements prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.powerWatts
	ch <- c.voltageVolts
	ch <- c.currentAmps
	ch <- c.energyTotalKwh
	ch <- c.energyTodayKwh
	ch <- c.energyYestKwh
	ch <- c.powerFactor
	ch <- c.apparentPower
	ch <- c.reactivePower
	ch <- c.relayState
	ch <- c.uptimeSeconds
	ch <- c.wifiRSSI
	ch <- c.wifiSignalDBm
}

// Collect implements prometheus.Collector.
func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup

	for _, target := range c.targets {
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			c.collectTarget(ch, addr)
		}(target)
	}

	wg.Wait()
}

func (c *Collector) collectTarget(ch chan<- prometheus.Metric, address string) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	status, err := c.client.FetchStatus(ctx, address)
	if err != nil {
		log.Printf("Error fetching status from %s: %v", address, err)
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0, address)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1, address)

	device := status.DeviceName()
	energy := status.StatusSNS.Energy

	ch <- prometheus.MustNewConstMetric(c.powerWatts, prometheus.GaugeValue, energy.Power, address, device)
	ch <- prometheus.MustNewConstMetric(c.voltageVolts, prometheus.GaugeValue, energy.Voltage, address, device)
	ch <- prometheus.MustNewConstMetric(c.currentAmps, prometheus.GaugeValue, energy.Current, address, device)
	ch <- prometheus.MustNewConstMetric(c.energyTotalKwh, prometheus.CounterValue, energy.Total, address, device)
	ch <- prometheus.MustNewConstMetric(c.energyTodayKwh, prometheus.GaugeValue, energy.Today, address, device)
	ch <- prometheus.MustNewConstMetric(c.energyYestKwh, prometheus.GaugeValue, energy.Yesterday, address, device)
	ch <- prometheus.MustNewConstMetric(c.powerFactor, prometheus.GaugeValue, energy.Factor, address, device)
	ch <- prometheus.MustNewConstMetric(c.apparentPower, prometheus.GaugeValue, energy.ApparentPower, address, device)
	ch <- prometheus.MustNewConstMetric(c.reactivePower, prometheus.GaugeValue, energy.ReactivePower, address, device)
	ch <- prometheus.MustNewConstMetric(c.relayState, prometheus.GaugeValue, status.RelayState(), address, device)
	ch <- prometheus.MustNewConstMetric(c.uptimeSeconds, prometheus.GaugeValue, float64(status.StatusSTS.UptimeSec), address, device)
	ch <- prometheus.MustNewConstMetric(c.wifiRSSI, prometheus.GaugeValue, float64(status.StatusSTS.Wifi.RSSI), address, device)
	ch <- prometheus.MustNewConstMetric(c.wifiSignalDBm, prometheus.GaugeValue, float64(status.StatusSTS.Wifi.Signal), address, device)
}
