package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lucavb/tasmota-plug-exporter/collector"
	"github.com/lucavb/tasmota-plug-exporter/tasmota"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	listenAddr := getEnv("LISTEN_ADDRESS", ":9184")
	targets := getTargets()
	timeout := getTimeout()

	if len(targets) == 0 {
		log.Fatal("TASMOTA_TARGETS environment variable is required")
	}

	log.Printf("Starting Tasmota exporter on %s", listenAddr)
	log.Printf("Monitoring targets: %v", targets)

	client := tasmota.NewClient(timeout)
	coll := collector.New(client, targets, timeout)

	prometheus.MustRegister(coll)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK\n"))
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
<head><title>Tasmota Exporter</title></head>
<body>
<h1>Tasmota Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
<p><a href="/health">Health</a></p>
</body>
</html>`))
	})

	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getTargets() []string {
	raw := os.Getenv("TASMOTA_TARGETS")
	if raw == "" {
		return nil
	}

	var targets []string
	for _, t := range strings.Split(raw, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			targets = append(targets, t)
		}
	}
	return targets
}

func getTimeout() time.Duration {
	raw := getEnv("SCRAPE_TIMEOUT", "5s")
	d, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("Invalid SCRAPE_TIMEOUT %q, using default 5s", raw)
		return 5 * time.Second
	}
	return d
}
