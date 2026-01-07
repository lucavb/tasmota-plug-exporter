package tasmota

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testResponse = `{
  "Status": {
    "DeviceName": "Tasmota",
    "FriendlyName": ["Living Room Plug"],
    "Power": "1"
  },
  "StatusSNS": {
    "Time": "2026-01-07T17:25:17",
    "ENERGY": {
      "Total": 123.456,
      "Yesterday": 1.5,
      "Today": 0.8,
      "Power": 42,
      "ApparentPower": 45,
      "ReactivePower": 10,
      "Factor": 0.93,
      "Voltage": 230,
      "Current": 0.183
    }
  },
  "StatusSTS": {
    "Uptime": "1T02:30:00",
    "UptimeSec": 95400,
    "POWER": "ON",
    "Wifi": {
      "RSSI": 72,
      "Signal": -64
    }
  }
}`

func TestFetchStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cm" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("cmnd") != "Status 0" {
			t.Errorf("unexpected cmnd: %s", r.URL.Query().Get("cmnd"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testResponse))
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	status, err := client.FetchStatus(context.Background(), server.Listener.Addr().String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.DeviceName() != "Living Room Plug" {
		t.Errorf("expected device name 'Living Room Plug', got %q", status.DeviceName())
	}

	if status.StatusSNS.Energy.Power != 42 {
		t.Errorf("expected power 42, got %f", status.StatusSNS.Energy.Power)
	}

	if status.StatusSNS.Energy.Voltage != 230 {
		t.Errorf("expected voltage 230, got %f", status.StatusSNS.Energy.Voltage)
	}

	if status.StatusSTS.UptimeSec != 95400 {
		t.Errorf("expected uptime 95400, got %d", status.StatusSTS.UptimeSec)
	}

	if status.RelayState() != 1 {
		t.Errorf("expected relay state 1, got %f", status.RelayState())
	}
}

func TestFetchStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(5 * time.Second)
	_, err := client.FetchStatus(context.Background(), server.Listener.Addr().String())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeviceNameFallback(t *testing.T) {
	status := &Status{
		Status: DeviceStatus{
			DeviceName:   "Tasmota",
			FriendlyName: []string{""},
		},
	}

	if status.DeviceName() != "Tasmota" {
		t.Errorf("expected fallback to DeviceName, got %q", status.DeviceName())
	}
}

func TestRelayStateOff(t *testing.T) {
	status := &Status{
		Status:    DeviceStatus{Power: "0"},
		StatusSTS: StateStatus{Power: "OFF"},
	}

	if status.RelayState() != 0 {
		t.Errorf("expected relay state 0, got %f", status.RelayState())
	}
}
