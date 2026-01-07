package tasmota

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Status represents the full Tasmota status response from Status 0.
type Status struct {
	Status    DeviceStatus `json:"Status"`
	StatusSNS SensorStatus `json:"StatusSNS"`
	StatusSTS StateStatus  `json:"StatusSTS"`
}

// DeviceStatus contains basic device information.
type DeviceStatus struct {
	DeviceName   string   `json:"DeviceName"`
	FriendlyName []string `json:"FriendlyName"`
	Power        string   `json:"Power"` // "0" or "1"
}

// SensorStatus contains sensor readings including energy data.
type SensorStatus struct {
	Time   string       `json:"Time"`
	Energy EnergyStatus `json:"ENERGY"`
}

// EnergyStatus contains power and energy metrics.
type EnergyStatus struct {
	Total         float64 `json:"Total"`
	Yesterday     float64 `json:"Yesterday"`
	Today         float64 `json:"Today"`
	Power         float64 `json:"Power"`
	ApparentPower float64 `json:"ApparentPower"`
	ReactivePower float64 `json:"ReactivePower"`
	Factor        float64 `json:"Factor"`
	Voltage       float64 `json:"Voltage"`
	Current       float64 `json:"Current"`
}

// StateStatus contains device state information.
type StateStatus struct {
	Uptime    string     `json:"Uptime"`
	UptimeSec int64      `json:"UptimeSec"`
	Power     string     `json:"POWER"` // "ON" or "OFF"
	Wifi      WifiStatus `json:"Wifi"`
}

// WifiStatus contains WiFi connection information.
type WifiStatus struct {
	RSSI   int `json:"RSSI"`
	Signal int `json:"Signal"`
}

// Client is an HTTP client for Tasmota devices.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new Tasmota client with the specified timeout.
func NewClient(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchStatus retrieves the full status from a Tasmota device.
func (c *Client) FetchStatus(ctx context.Context, address string) (*Status, error) {
	url := fmt.Sprintf("http://%s/cm?cmnd=Status%%200", address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching status: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var status Status
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &status, nil
}

// DeviceName returns the device name, preferring FriendlyName if available.
func (s *Status) DeviceName() string {
	if len(s.Status.FriendlyName) > 0 && s.Status.FriendlyName[0] != "" {
		return s.Status.FriendlyName[0]
	}
	return s.Status.DeviceName
}

// RelayState returns 1 if the relay is on, 0 if off.
func (s *Status) RelayState() float64 {
	if s.StatusSTS.Power == "ON" || s.Status.Power == "1" {
		return 1
	}
	return 0
}
