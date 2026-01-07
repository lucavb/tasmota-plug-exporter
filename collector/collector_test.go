package collector

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lucavb/tasmota-plug-exporter/tasmota"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

const testResponse = `{
  "Status": {
    "DeviceName": "Tasmota",
    "FriendlyName": ["Test Plug"],
    "Power": "1"
  },
  "StatusSNS": {
    "Time": "2026-01-07T17:25:17",
    "ENERGY": {
      "Total": 100.5,
      "Yesterday": 2.0,
      "Today": 1.5,
      "Power": 50,
      "ApparentPower": 55,
      "ReactivePower": 15,
      "Factor": 0.91,
      "Voltage": 235,
      "Current": 0.213
    }
  },
  "StatusSTS": {
    "Uptime": "0T01:00:00",
    "UptimeSec": 3600,
    "POWER": "ON",
    "Wifi": {
      "RSSI": 80,
      "Signal": -60
    }
  }
}`

func TestCollector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testResponse))
	}))
	defer server.Close()

	client := tasmota.NewClient(5 * time.Second)
	coll := New(client, []string{server.Listener.Addr().String()}, 5*time.Second)

	// Test that metrics are collected without error
	count := testutil.CollectAndCount(coll)
	if count != 14 {
		t.Errorf("expected 14 metrics, got %d", count)
	}
}

func TestCollectorTargetDown(t *testing.T) {
	client := tasmota.NewClient(1 * time.Second)
	coll := New(client, []string{"127.0.0.1:1"}, 1*time.Second)

	expected := `
# HELP tasmota_up Whether the Tasmota device is reachable
# TYPE tasmota_up gauge
tasmota_up{address="127.0.0.1:1"} 0
`

	if err := testutil.CollectAndCompare(coll, strings.NewReader(expected), "tasmota_up"); err != nil {
		t.Errorf("unexpected metrics: %v", err)
	}
}

func TestCollectorDescribe(t *testing.T) {
	client := tasmota.NewClient(5 * time.Second)
	coll := New(client, []string{"127.0.0.1:9999"}, 5*time.Second)

	ch := make(chan *prometheus.Desc, 20)
	coll.Describe(ch)
	close(ch)

	count := 0
	for range ch {
		count++
	}

	if count != 14 {
		t.Errorf("expected 14 metric descriptions, got %d", count)
	}
}
