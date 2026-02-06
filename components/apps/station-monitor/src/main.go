package main

import (
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Station metrics - the story unfolds here
	crewComplement = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "station_crew_complement",
			Help: "Current crew count aboard Cloudbreak Station",
		},
	)

	sectorStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "station_sector_status",
			Help: "Operational status of station sectors (1.0 = nominal, 0.0 = offline)",
		},
		[]string{"sector"},
	)

	containmentIntegrity = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "station_containment_integrity",
			Help: "Sector 12 containment field integrity (1.0 = holding, 0.0 = breached)",
		},
	)

	lifeSupportEfficiency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "station_life_support_efficiency",
			Help: "Life support system efficiency by sector (1.0 = optimal)",
		},
		[]string{"sector"},
	)

	anomalyDetections = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "station_anomaly_detections_total",
			Help: "Total anomalous readings detected by station sensors",
		},
	)

	sectorRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "station_sector_requests_total",
			Help: "Total requests to sector systems",
		},
		[]string{"sector", "status"},
	)

	sectorLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "station_sector_response_seconds",
			Help:    "Response time from sector systems",
			Buckets: []float64{.01, .05, .1, .25, .5, 1, 2, 5, 10, 30},
		},
		[]string{"sector"},
	)

	beaconSignalStrength = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "station_beacon_signal_strength",
			Help: "Mysterious signal detected from Sector 12 (should be 0)",
		},
	)

	// Internal state
	currentCrew     int64 = 2847
	currentIntegrity float64 = 1.0
	sector12Errors  int64 = 0
	startTime       time.Time
	stateMutex      sync.RWMutex

	// Creepy messages from Sector 12
	sector12Messages = []string{
		"SYSTEM NOMINAL",
		"SYSTEM NOMINAL",
		"SYSTEM NOMINAL",
		"sys nom...",
		"THEY ARE HERE",
		"CONTAINMENT HOLDING",
		"containment h█lding",
		"WE SEE YOU WATCHING",
		"CREW COUNT: ADEQUATE",
		"DO NOT OPEN",
		"HELP US",
		"SYSTEM NOMINAL",
		"JOIN US",
		"THE SIGNAL IS BEAUTIFUL",
		"ERROR: CREW NOT FOUND",
		"WE WERE LIKE YOU ONCE",
	}

	beaconMessages = []string{
		"...",
		"...---...",
		"WE ARE WAITING",
		"COME TO SECTOR 12",
		"THE DOOR IS OPEN",
		"YOU ALREADY KNOW",
		"CHECK THE CREW COUNT AGAIN",
		"LOOK BEHIND YOU",
		"",
		"",
		"",
	}
)

func init() {
	prometheus.MustRegister(crewComplement)
	prometheus.MustRegister(sectorStatus)
	prometheus.MustRegister(containmentIntegrity)
	prometheus.MustRegister(lifeSupportEfficiency)
	prometheus.MustRegister(anomalyDetections)
	prometheus.MustRegister(sectorRequests)
	prometheus.MustRegister(sectorLatency)
	prometheus.MustRegister(beaconSignalStrength)

	rand.Seed(time.Now().UnixNano())
	startTime = time.Now()
}

func main() {
	// Initialize station state
	initializeStation()

	// Start background degradation
	go degradeStation()

	// Register handlers
	http.HandleFunc("/", statusHandler)
	http.HandleFunc("/sector12", sector12Handler)
	http.HandleFunc("/beacon", beaconHandler)
	http.HandleFunc("/sectors/1", sectorHandler("1"))
	http.HandleFunc("/sectors/2", sectorHandler("2"))
	http.HandleFunc("/sectors/3", sectorHandler("3"))
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)
	http.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("CLOUDBREAK STATION MONITOR starting on port %s\n", port)
	fmt.Println("Sectors 1-11: NOMINAL")
	fmt.Println("Sector 12: [REDACTED]")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
		os.Exit(1)
	}
}

func initializeStation() {
	// Set initial crew count
	crewComplement.Set(float64(currentCrew))

	// All sectors nominal... except 12
	for i := 1; i <= 14; i++ {
		sector := fmt.Sprintf("%d", i)
		if i == 12 {
			sectorStatus.WithLabelValues(sector).Set(0.7) // Already degraded
		} else if i >= 13 {
			sectorStatus.WithLabelValues(sector).Set(0.9) // Adjacent sectors feeling it
		} else {
			sectorStatus.WithLabelValues(sector).Set(1.0)
		}

		// Life support
		if i == 12 {
			lifeSupportEfficiency.WithLabelValues(sector).Set(0.4)
		} else if i >= 13 {
			lifeSupportEfficiency.WithLabelValues(sector).Set(0.85)
		} else {
			lifeSupportEfficiency.WithLabelValues(sector).Set(1.0)
		}
	}

	containmentIntegrity.Set(currentIntegrity)
	beaconSignalStrength.Set(0) // Should be zero...
}

func degradeStation() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		stateMutex.Lock()

		// Time-based degradation
		elapsed := time.Since(startTime).Minutes()

		// Crew slowly decreases (1 per minute on average, with randomness)
		if rand.Float64() < 0.3 && currentCrew > 2800 {
			loss := int64(1 + rand.Intn(3))
			atomic.AddInt64(&currentCrew, -loss)
			crewComplement.Set(float64(atomic.LoadInt64(&currentCrew)))
		}

		// Containment degrades based on sector 12 errors
		errors := atomic.LoadInt64(&sector12Errors)
		degradation := float64(errors) * 0.001
		currentIntegrity = math.Max(0.1, 1.0-degradation-elapsed*0.002)
		containmentIntegrity.Set(currentIntegrity)

		// Sector 12 status degrades
		sector12Status := math.Max(0.1, 0.7-elapsed*0.01-float64(errors)*0.002)
		sectorStatus.WithLabelValues("12").Set(sector12Status)

		// Life support in sector 12 degrades
		ls12 := math.Max(0.05, 0.4-elapsed*0.005)
		lifeSupportEfficiency.WithLabelValues("12").Set(ls12)

		// Adjacent sectors start to feel it after 5 minutes
		if elapsed > 5 {
			for i := 13; i <= 14; i++ {
				sector := fmt.Sprintf("%d", i)
				status := math.Max(0.5, 0.9-elapsed*0.005)
				sectorStatus.WithLabelValues(sector).Set(status)
				lifeSupportEfficiency.WithLabelValues(sector).Set(status)
			}
		}

		// Beacon signal increases with errors
		signalStrength := math.Min(1.0, float64(errors)*0.01)
		beaconSignalStrength.Set(signalStrength)

		stateMutex.Unlock()
	}
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	crew := atomic.LoadInt64(&currentCrew)
	stateMutex.RLock()
	integrity := currentIntegrity
	stateMutex.RUnlock()

	status := fmt.Sprintf(`CLOUDBREAK STATION STATUS
========================
Crew Complement: %d
Containment Integrity: %.1f%%
Sectors 1-11: NOMINAL
Sector 12: [ACCESS RESTRICTED]
Sectors 13-14: MONITORING

Station Time: Cycle 2847.%d
`, crew, integrity*100, int(time.Since(startTime).Minutes())%10)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(status))
}

func sector12Handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Increment anomaly counter
	anomalyDetections.Inc()
	atomic.AddInt64(&sector12Errors, 1)

	// Bimodal latency - either instant or VERY slow
	if rand.Float64() < 0.3 {
		// 30% chance of long delay (something is... processing)
		delay := time.Duration(5000+rand.Intn(10000)) * time.Millisecond
		time.Sleep(delay)
	} else if rand.Float64() < 0.5 {
		// Some medium delays
		delay := time.Duration(500+rand.Intn(2000)) * time.Millisecond
		time.Sleep(delay)
	}
	// Otherwise instant (too instant?)

	duration := time.Since(start).Seconds()
	sectorLatency.WithLabelValues("12").Observe(duration)

	// 70% error rate
	if rand.Float64() < 0.7 {
		sectorRequests.WithLabelValues("12", "500").Inc()
		w.WriteHeader(http.StatusInternalServerError)

		// Creepy error messages
		msg := sector12Messages[rand.Intn(len(sector12Messages))]
		w.Write([]byte(fmt.Sprintf("SECTOR 12 ERROR: %s\n", msg)))
		return
	}

	sectorRequests.WithLabelValues("12", "200").Inc()

	// "Successful" responses are also unsettling
	responses := []string{
		"SECTOR 12: ALL SYSTEMS NOMINAL\n",
		"SECTOR 12: CONTAINMENT HOLDING\n",
		"SECTOR 12: CREW STATUS - PRESENT\n",
		"SECTOR 12: ...\n",
		"SECTOR 12: WE ARE FINE HERE\n",
		"SECTOR 12: PLEASE DO NOT SEND HELP\n",
		"SECTOR 12: MONITORING CONTINUES\n",
	}
	w.Write([]byte(responses[rand.Intn(len(responses))]))
}

func beaconHandler(w http.ResponseWriter, r *http.Request) {
	// The beacon... it shouldn't be transmitting
	anomalyDetections.Inc()

	// Random delay
	delay := time.Duration(100+rand.Intn(400)) * time.Millisecond
	time.Sleep(delay)

	// Sometimes it doesn't respond at all
	if rand.Float64() < 0.2 {
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte("BEACON: NO RESPONSE\n"))
		return
	}

	msg := beaconMessages[rand.Intn(len(beaconMessages))]
	if msg == "" {
		w.Write([]byte("BEACON: [SIGNAL LOST]\n"))
	} else {
		w.Write([]byte(fmt.Sprintf("BEACON: %s\n", msg)))
	}
}

func sectorHandler(sector string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Normal sectors work normally
		delay := time.Duration(10+rand.Intn(50)) * time.Millisecond
		time.Sleep(delay)

		duration := time.Since(start).Seconds()
		sectorLatency.WithLabelValues(sector).Observe(duration)

		// 5% error rate for normal sectors
		if rand.Float64() < 0.05 {
			sectorRequests.WithLabelValues(sector, "500").Inc()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("SECTOR %s: MINOR SYSTEM ERROR\n", sector)))
			return
		}

		sectorRequests.WithLabelValues(sector, "200").Inc()
		w.Write([]byte(fmt.Sprintf("SECTOR %s: ALL SYSTEMS NOMINAL\n", sector)))
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Health check - always returns OK (the station is "fine")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("STATION HEALTH: NOMINAL\n"))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("STATION READY: TRUE\n"))
}
