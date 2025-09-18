package security

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/palaseus/adrenochain/pkg/logger"
)

// SecurityDashboard provides a web-based security monitoring dashboard
type SecurityDashboard struct {
	monitor *SecurityMonitor
	logger  *logger.Logger
	router  *mux.Router
}

// DashboardData represents the data structure for the dashboard
type DashboardData struct {
	Metrics         *SecurityMetrics `json:"metrics"`
	RecentAnomalies []*Anomaly       `json:"recent_anomalies"`
	RecentEvents    []*SecurityEvent `json:"recent_events"`
	Status          string           `json:"status"`
	LastUpdated     time.Time        `json:"last_updated"`
	AlertThresholds *AlertThresholds `json:"alert_thresholds"`
}

// NewSecurityDashboard creates a new security dashboard
func NewSecurityDashboard(monitor *SecurityMonitor, logger *logger.Logger) *SecurityDashboard {
	dashboard := &SecurityDashboard{
		monitor: monitor,
		logger:  logger,
		router:  mux.NewRouter(),
	}

	dashboard.setupRoutes()
	return dashboard
}

// setupRoutes configures the dashboard API routes
func (sd *SecurityDashboard) setupRoutes() {
	// Dashboard main page
	sd.router.HandleFunc("/", sd.dashboardHandler).Methods("GET")

	// API endpoints
	sd.router.HandleFunc("/api/metrics", sd.metricsHandler).Methods("GET")
	sd.router.HandleFunc("/api/anomalies", sd.anomaliesHandler).Methods("GET")
	sd.router.HandleFunc("/api/events", sd.eventsHandler).Methods("GET")
	sd.router.HandleFunc("/api/status", sd.statusHandler).Methods("GET")

	// Static files (CSS, JS)
	sd.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
}

// dashboardHandler serves the main dashboard page
func (sd *SecurityDashboard) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Adrenochain Security Dashboard</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            margin-bottom: 20px;
            text-align: center;
        }
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .metric-card {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            border-left: 4px solid #667eea;
        }
        .metric-value {
            font-size: 2em;
            font-weight: bold;
            color: #333;
        }
        .metric-label {
            color: #666;
            margin-top: 5px;
        }
        .anomalies-section, .events-section {
            background: white;
            padding: 20px;
            border-radius: 10px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        .section-title {
            font-size: 1.5em;
            font-weight: bold;
            margin-bottom: 15px;
            color: #333;
        }
        .anomaly-item, .event-item {
            padding: 10px;
            border-left: 3px solid #ff6b6b;
            margin-bottom: 10px;
            background: #f8f9fa;
            border-radius: 5px;
        }
        .event-item {
            border-left-color: #4ecdc4;
        }
        .severity-critical { border-left-color: #ff4757; }
        .severity-high { border-left-color: #ff6b6b; }
        .severity-medium { border-left-color: #ffa502; }
        .severity-low { border-left-color: #2ed573; }
        .timestamp {
            font-size: 0.9em;
            color: #666;
        }
        .refresh-btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 10px 20px;
            border-radius: 5px;
            cursor: pointer;
            margin-bottom: 20px;
        }
        .refresh-btn:hover {
            background: #5a6fd8;
        }
        .status-indicator {
            display: inline-block;
            width: 12px;
            height: 12px;
            border-radius: 50%;
            margin-right: 8px;
        }
        .status-ok { background-color: #2ed573; }
        .status-warning { background-color: #ffa502; }
        .status-error { background-color: #ff4757; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🛡️ Adrenochain Security Dashboard</h1>
            <p>Real-time security monitoring and threat detection</p>
        </div>
        
        <button class="refresh-btn" onclick="refreshDashboard()">🔄 Refresh</button>
        
        <div class="metrics-grid" id="metrics-grid">
            <!-- Metrics will be loaded here -->
        </div>
        
        <div class="anomalies-section">
            <h2 class="section-title">🚨 Recent Anomalies</h2>
            <div id="anomalies-list">
                <!-- Anomalies will be loaded here -->
            </div>
        </div>
        
        <div class="events-section">
            <h2 class="section-title">📊 Recent Security Events</h2>
            <div id="events-list">
                <!-- Events will be loaded here -->
            </div>
        </div>
    </div>

    <script>
        function refreshDashboard() {
            loadMetrics();
            loadAnomalies();
            loadEvents();
        }

        function loadMetrics() {
            fetch('/api/metrics')
                .then(response => response.json())
                .then(data => {
                    const grid = document.getElementById('metrics-grid');
                    grid.innerHTML = 
                        '<div class="metric-card">' +
                            '<div class="metric-value">' + data.total_events + '</div>' +
                            '<div class="metric-label">Total Events</div>' +
                        '</div>' +
                        '<div class="metric-card">' +
                            '<div class="metric-value">' + data.anomalies_detected + '</div>' +
                            '<div class="metric-label">Anomalies Detected</div>' +
                        '</div>' +
                        '<div class="metric-card">' +
                            '<div class="metric-value">' + data.alerts_triggered + '</div>' +
                            '<div class="metric-label">Alerts Triggered</div>' +
                        '</div>' +
                        '<div class="metric-card">' +
                            '<div class="metric-value">' + (data.events_by_severity.critical || 0) + '</div>' +
                            '<div class="metric-label">Critical Events</div>' +
                        '</div>';
                })
                .catch(error => console.error('Error loading metrics:', error));
        }

        function loadAnomalies() {
            fetch('/api/anomalies')
                .then(response => response.json())
                .then(data => {
                    const list = document.getElementById('anomalies-list');
                    if (data.length === 0) {
                        list.innerHTML = '<p>No recent anomalies detected.</p>';
                        return;
                    }
                    
                    list.innerHTML = data.map(anomaly => 
                        '<div class="anomaly-item severity-' + anomaly.severity + '">' +
                            '<strong>' + anomaly.type.replace('_', ' ').toUpperCase() + '</strong>' +
                            '<p>' + anomaly.description + '</p>' +
                            '<div class="timestamp">' +
                                new Date(anomaly.timestamp).toLocaleString() + 
                                ' (Confidence: ' + (anomaly.confidence * 100).toFixed(1) + '%)' +
                            '</div>' +
                        '</div>'
                    ).join('');
                })
                .catch(error => console.error('Error loading anomalies:', error));
        }

        function loadEvents() {
            fetch('/api/events')
                .then(response => response.json())
                .then(data => {
                    const list = document.getElementById('events-list');
                    if (data.length === 0) {
                        list.innerHTML = '<p>No recent events.</p>';
                        return;
                    }
                    
                    list.innerHTML = data.slice(0, 10).map(event => 
                        '<div class="event-item severity-' + event.severity + '">' +
                            '<strong>' + event.type.replace('_', ' ').toUpperCase() + '</strong>' +
                            '<p>' + event.description + '</p>' +
                            '<div class="timestamp">' +
                                new Date(event.timestamp).toLocaleString() + 
                                ' from ' + event.source +
                            '</div>' +
                        '</div>'
                    ).join('');
                })
                .catch(error => console.error('Error loading events:', error));
        }

        // Load data on page load
        refreshDashboard();
        
        // Auto-refresh every 30 seconds
        setInterval(refreshDashboard, 30000);
    </script>
</body>
</html>
	`

	fmt.Fprint(w, html)
}

// metricsHandler returns security metrics as JSON
func (sd *SecurityDashboard) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := sd.monitor.GetMetrics()

	// Convert to JSON-friendly format
	response := map[string]interface{}{
		"total_events":       metrics.TotalEvents,
		"anomalies_detected": metrics.AnomaliesDetected,
		"alerts_triggered":   metrics.AlertsTriggered,
		"events_by_type":     metrics.EventsByType,
		"events_by_severity": metrics.EventsBySeverity,
		"last_updated":       metrics.LastUpdated,
	}

	json.NewEncoder(w).Encode(response)
}

// anomaliesHandler returns recent anomalies as JSON
func (sd *SecurityDashboard) anomaliesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	anomalies := sd.monitor.GetAnomalies(20) // Get last 20 anomalies
	json.NewEncoder(w).Encode(anomalies)
}

// eventsHandler returns recent security events as JSON
func (sd *SecurityDashboard) eventsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get recent events (this would need to be implemented in SecurityMonitor)
	events := []*SecurityEvent{} // Placeholder
	json.NewEncoder(w).Encode(events)
}

// statusHandler returns the current security status
func (sd *SecurityDashboard) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := sd.monitor.GetMetrics()
	anomalies := sd.monitor.GetAnomalies(5)

	// Determine overall status
	status := "ok"
	if metrics.EventsBySeverity[MonitorSeverityCritical] > 0 {
		status = "critical"
	} else if metrics.EventsBySeverity[MonitorSeverityHigh] > 5 {
		status = "warning"
	} else if len(anomalies) > 0 {
		status = "warning"
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now(),
		"metrics":   metrics,
		"anomalies": len(anomalies),
	}

	json.NewEncoder(w).Encode(response)
}

// GetRouter returns the HTTP router for the dashboard
func (sd *SecurityDashboard) GetRouter() *mux.Router {
	return sd.router
}

// StartDashboard starts the security dashboard server
func (sd *SecurityDashboard) StartDashboard(port int) error {
	sd.logger.Info("Starting security dashboard on port %d", port)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: sd.router,
	}

	return server.ListenAndServe()
}
