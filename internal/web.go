/*
Copyright © 2026 Patrick Hermann patrick.hermann@sva.de
*/

package internal

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

// StartWebServer starts the HTTP server for HTMX frontend and REST API
func StartWebServer(httpPort string, monitor *Monitor, clusterStore *ClusterStore, loadFrom, configLoc, configNm string) {
	mux := http.NewServeMux()

	// HTMX frontend routes
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		handleDashboard(w, r, monitor)
	})
	mux.HandleFunc("GET /service/{name}", func(w http.ResponseWriter, r *http.Request) {
		handleServiceDetail(w, r, monitor)
	})
	mux.HandleFunc("GET /clusters", func(w http.ResponseWriter, r *http.Request) {
		handleClusters(w, r, clusterStore)
	})
	mux.HandleFunc("GET /cluster/{name}", func(w http.ResponseWriter, r *http.Request) {
		handleClusterDetail(w, r, clusterStore)
	})
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) {
		handleAdmin(w, r, monitor)
	})

	// HTMX partial routes
	mux.HandleFunc("POST /htmx/add-service", func(w http.ResponseWriter, r *http.Request) {
		handleHTMXAddService(w, r, monitor, loadFrom, configLoc, configNm)
	})
	mux.HandleFunc("POST /htmx/edit-service", func(w http.ResponseWriter, r *http.Request) {
		handleHTMXEditService(w, r, monitor, loadFrom, configLoc, configNm)
	})
	mux.HandleFunc("POST /htmx/delete-service", func(w http.ResponseWriter, r *http.Request) {
		handleHTMXDeleteService(w, r, monitor, loadFrom, configLoc, configNm)
	})
	mux.HandleFunc("GET /htmx/status/{name}", func(w http.ResponseWriter, r *http.Request) {
		handleHTMXStatus(w, r, monitor)
	})
	mux.HandleFunc("GET /htmx/history/{name}", func(w http.ResponseWriter, r *http.Request) {
		handleHTMXHistory(w, r, monitor)
	})
	mux.HandleFunc("GET /htmx/cards", func(w http.ResponseWriter, r *http.Request) {
		handleHTMXCards(w, r, monitor)
	})

	// REST API routes
	mux.HandleFunc("GET /api/v1/services", func(w http.ResponseWriter, r *http.Request) {
		handleAPIListServices(w, r, monitor)
	})
	mux.HandleFunc("GET /api/v1/services/{name}", func(w http.ResponseWriter, r *http.Request) {
		handleAPIGetService(w, r, monitor)
	})
	mux.HandleFunc("POST /api/v1/services", func(w http.ResponseWriter, r *http.Request) {
		handleAPICreateService(w, r, monitor, loadFrom, configLoc, configNm)
	})
	mux.HandleFunc("DELETE /api/v1/services/{name}", func(w http.ResponseWriter, r *http.Request) {
		handleAPIDeleteService(w, r, monitor, loadFrom, configLoc, configNm)
	})
	mux.HandleFunc("GET /api/v1/clusters", func(w http.ResponseWriter, r *http.Request) {
		handleAPIClusters(w, r, clusterStore)
	})
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	log.Printf("HTTP/HTMX SERVER LISTENING AT :%s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		log.Fatalf("FAILED TO START HTTP SERVER: %v", err)
	}
}

// --- Dashboard ---

type dashboardData struct {
	Services   []serviceCardData
	Categories []string
	Version    string
	Commit     string
	Date       string
}

type serviceCardData struct {
	Name         string
	Description  string
	Category     string
	URL          string
	LogoURL      string
	Icon         string
	Status       string
	ResponseTime int64
	TLSDaysLeft  int
	History      []string // last 20 status colors
}

func handleDashboard(w http.ResponseWriter, r *http.Request, monitor *Monitor) {
	states := monitor.GetStates()
	categoryFilter := r.URL.Query().Get("category")

	cards := buildCards(states, categoryFilter)
	categories := getCategories(states)

	data := dashboardData{
		Services:   cards,
		Categories: categories,
		Version:    version,
		Commit:     commit,
		Date:       date,
	}

	tmpl := template.Must(template.New("dashboard").Funcs(templateFuncs()).Parse(dashboardTemplate))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func buildCards(states map[string]*ServiceState, categoryFilter string) []serviceCardData {
	var cards []serviceCardData
	for _, state := range states {
		svc := state.Service
		if categoryFilter != "" && svc.Category != categoryFilter {
			continue
		}

		card := serviceCardData{
			Name:        svc.Name,
			Description: svc.Description,
			Category:    svc.Category,
			URL:         svc.URL,
			LogoURL:     svc.LogoURL,
			Icon:        svc.Icon,
			Status:      state.GetCurrentStatus(),
		}

		if last, ok := state.GetLastResult(); ok {
			card.ResponseTime = last.ResponseTime
			card.TLSDaysLeft = last.TLSDaysLeft
		}

		// Build history colors (last 20)
		results := state.GetResults()
		start := 0
		if len(results) > 20 {
			start = len(results) - 20
		}
		for _, r := range results[start:] {
			switch r.Status {
			case "UP":
				card.History = append(card.History, "#44dd88")
			case "DEGRADED":
				card.History = append(card.History, "#f0a030")
			case "DOWN":
				card.History = append(card.History, "#f05050")
			default:
				card.History = append(card.History, "#333")
			}
		}

		cards = append(cards, card)
	}

	sort.Slice(cards, func(i, j int) bool {
		if cards[i].Category != cards[j].Category {
			return cards[i].Category < cards[j].Category
		}
		return cards[i].Name < cards[j].Name
	})
	return cards
}

func getCategories(states map[string]*ServiceState) []string {
	catMap := map[string]bool{}
	for _, s := range states {
		if s.Service.Category != "" {
			catMap[s.Service.Category] = true
		}
	}
	cats := make([]string, 0, len(catMap))
	for c := range catMap {
		cats = append(cats, c)
	}
	sort.Strings(cats)
	return cats
}

// --- Service Detail ---

func handleServiceDetail(w http.ResponseWriter, r *http.Request, monitor *Monitor) {
	name := r.PathValue("name")
	state, ok := monitor.GetState(name)
	if !ok {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	data := struct {
		Service Service
		Status  string
		Results []CheckResult
		Version string
	}{
		Service: state.Service,
		Status:  state.GetCurrentStatus(),
		Results: state.GetResults(),
		Version: version,
	}

	tmpl := template.Must(template.New("detail").Funcs(templateFuncs()).Parse(serviceDetailTemplate))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// --- Clusters ---

func handleClusters(w http.ResponseWriter, r *http.Request, cs *ClusterStore) {
	clusters := cs.GetAllClusters()
	data := struct {
		Clusters []ClusterInfo
		Version  string
	}{clusters, version}

	tmpl := template.Must(template.New("clusters").Funcs(templateFuncs()).Parse(clustersTemplate))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleClusterDetail(w http.ResponseWriter, r *http.Request, cs *ClusterStore) {
	name := r.PathValue("name")
	inv, ok := cs.GetInventory(name)
	if !ok {
		http.Error(w, "Cluster not found", http.StatusNotFound)
		return
	}

	data := struct {
		Inventory *ClusterInventory
		Version   string
	}{inv, version}

	tmpl := template.Must(template.New("cluster-detail").Funcs(templateFuncs()).Parse(clusterDetailTemplate))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// --- Admin ---

func handleAdmin(w http.ResponseWriter, r *http.Request, monitor *Monitor) {
	services := monitor.GetServices()
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})

	data := struct {
		Services []Service
		Version  string
	}{services, version}

	tmpl := template.Must(template.New("admin").Funcs(templateFuncs()).Parse(adminTemplate))
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// --- HTMX Handlers ---

func handleHTMXAddService(w http.ResponseWriter, r *http.Request, monitor *Monitor, loadFrom, configLoc, configNm string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	svc := Service{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Category:    r.FormValue("category"),
		URL:         r.FormValue("url"),
		LogoURL:     r.FormValue("logo_url"),
		Icon:        r.FormValue("icon"),
		Tags:        parseTags(r.FormValue("tags")),
		HealthCheck: HealthCheckConfig{
			Enabled:        r.FormValue("health_enabled") == "on",
			Interval:       30,
			ExpectedStatus: 200,
			TLSCheck:       r.FormValue("tls_check") == "on",
		},
	}

	if svc.Name == "" || svc.URL == "" {
		http.Error(w, "Name and URL are required", http.StatusBadRequest)
		return
	}

	monitor.AddService(svc)
	persistServices(monitor, loadFrom, configLoc, configNm)

	w.Header().Set("HX-Redirect", "/admin")
	w.WriteHeader(http.StatusOK)
}

func handleHTMXEditService(w http.ResponseWriter, r *http.Request, monitor *Monitor, loadFrom, configLoc, configNm string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	svc := Service{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Category:    r.FormValue("category"),
		URL:         r.FormValue("url"),
		LogoURL:     r.FormValue("logo_url"),
		Icon:        r.FormValue("icon"),
		Tags:        parseTags(r.FormValue("tags")),
		HealthCheck: HealthCheckConfig{
			Enabled:        r.FormValue("health_enabled") == "on",
			Interval:       30,
			ExpectedStatus: 200,
			TLSCheck:       r.FormValue("tls_check") == "on",
		},
	}

	monitor.UpdateService(svc)
	persistServices(monitor, loadFrom, configLoc, configNm)

	w.Header().Set("HX-Redirect", "/admin")
	w.WriteHeader(http.StatusOK)
}

// parseTags splits a comma-separated string into trimmed, non-empty tags.
func parseTags(raw string) []string {
	var tags []string
	for _, t := range strings.Split(raw, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func handleHTMXDeleteService(w http.ResponseWriter, r *http.Request, monitor *Monitor, loadFrom, configLoc, configNm string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	monitor.RemoveService(name)
	persistServices(monitor, loadFrom, configLoc, configNm)

	w.Header().Set("HX-Redirect", "/admin")
	w.WriteHeader(http.StatusOK)
}

func handleHTMXStatus(w http.ResponseWriter, r *http.Request, monitor *Monitor) {
	name := r.PathValue("name")
	state, ok := monitor.GetState(name)
	if !ok {
		fmt.Fprint(w, `<span class="status-dot status-unknown"></span>`)
		return
	}

	status := state.GetCurrentStatus()
	var cssClass string
	switch status {
	case "UP":
		cssClass = "status-up"
	case "DEGRADED":
		cssClass = "status-degraded"
	case "DOWN":
		cssClass = "status-down"
	default:
		cssClass = "status-unknown"
	}

	last, hasLast := state.GetLastResult()
	responseTime := ""
	if hasLast {
		responseTime = fmt.Sprintf("%dms", last.ResponseTime)
	}

	fmt.Fprintf(w, `<span class="status-dot %s" title="%s"></span><span class="response-time">%s</span>`, cssClass, status, responseTime)
}

func handleHTMXHistory(w http.ResponseWriter, r *http.Request, monitor *Monitor) {
	name := r.PathValue("name")
	state, ok := monitor.GetState(name)
	if !ok {
		return
	}

	results := state.GetResults()
	start := 0
	if len(results) > 20 {
		start = len(results) - 20
	}

	fmt.Fprint(w, `<div class="mini-timeline">`)
	for _, res := range results[start:] {
		var color string
		switch res.Status {
		case "UP":
			color = "#44dd88"
		case "DEGRADED":
			color = "#f0a030"
		case "DOWN":
			color = "#f05050"
		default:
			color = "#333"
		}
		fmt.Fprintf(w, `<div class="timeline-bar" style="background:%s" title="%s %dms"></div>`, color, res.Status, res.ResponseTime)
	}
	fmt.Fprint(w, `</div>`)
}

func handleHTMXCards(w http.ResponseWriter, r *http.Request, monitor *Monitor) {
	states := monitor.GetStates()
	categoryFilter := r.URL.Query().Get("category")
	cards := buildCards(states, categoryFilter)

	tmpl := template.Must(template.New("cards").Funcs(templateFuncs()).Parse(cardsPartial))
	tmpl.Execute(w, cards)
}

// --- REST API Handlers ---

func handleAPIListServices(w http.ResponseWriter, r *http.Request, monitor *Monitor) {
	states := monitor.GetStates()
	type serviceStatus struct {
		Service      Service `json:"service"`
		Status       string  `json:"status"`
		ResponseTime int64   `json:"responseTime,omitempty"`
	}

	var result []serviceStatus
	for _, s := range states {
		ss := serviceStatus{
			Service: s.Service,
			Status:  s.GetCurrentStatus(),
		}
		if last, ok := s.GetLastResult(); ok {
			ss.ResponseTime = last.ResponseTime
		}
		result = append(result, ss)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleAPIGetService(w http.ResponseWriter, r *http.Request, monitor *Monitor) {
	name := r.PathValue("name")
	state, ok := monitor.GetState(name)
	if !ok {
		http.Error(w, `{"error":"service not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(struct {
		Service Service       `json:"service"`
		Status  string        `json:"status"`
		Results []CheckResult `json:"results"`
	}{
		Service: state.Service,
		Status:  state.GetCurrentStatus(),
		Results: state.GetResults(),
	})
}

func handleAPICreateService(w http.ResponseWriter, r *http.Request, monitor *Monitor, loadFrom, configLoc, configNm string) {
	var svc Service
	if err := json.NewDecoder(r.Body).Decode(&svc); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if svc.Name == "" || svc.URL == "" {
		http.Error(w, `{"error":"name and url are required"}`, http.StatusBadRequest)
		return
	}

	monitor.AddService(svc)
	persistServices(monitor, loadFrom, configLoc, configNm)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "service created"})
}

func handleAPIDeleteService(w http.ResponseWriter, r *http.Request, monitor *Monitor, loadFrom, configLoc, configNm string) {
	name := r.PathValue("name")
	_, ok := monitor.GetState(name)
	if !ok {
		http.Error(w, `{"error":"service not found"}`, http.StatusNotFound)
		return
	}

	monitor.RemoveService(name)
	persistServices(monitor, loadFrom, configLoc, configNm)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "service deleted"})
}

func handleAPIClusters(w http.ResponseWriter, r *http.Request, cs *ClusterStore) {
	clusters := cs.GetAllClusters()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clusters)
}

// persistServices saves the current service list to the configured backend
func persistServices(monitor *Monitor, loadFrom, configLoc, configNm string) {
	services := monitor.GetServices()
	SaveServices(services, loadFrom, configLoc, configNm)
}

// --- Template Functions ---

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,
		"lower":     strings.ToLower,
		"statusColor": func(status string) string {
			switch status {
			case "UP":
				return "#44dd88"
			case "DEGRADED":
				return "#f0a030"
			case "DOWN":
				return "#f05050"
			default:
				return "#666"
			}
		},
		"statusBg": func(status string) string {
			switch status {
			case "UP":
				return "#0d2e1a"
			case "DEGRADED":
				return "#2e1f08"
			case "DOWN":
				return "#2e0d0d"
			default:
				return "#1a1a24"
			}
		},
		"joinTags": func(tags []string) string {
			return strings.Join(tags, ", ")
		},
	}
}

// --- HTML Templates ---
// Uses Pico CSS as base framework (like homerun2-core-catcher)
// with cinema-themed warm color palette overrides

const pageHead = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>we run things{{.TitleSuffix}}</title>
    <link rel="icon" type="image/png" href="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/sthings-argo.png">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <link href="https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap" rel="stylesheet">
`

const commonStyles = `
    <style>
        :root { --pico-font-size: 14px; }
        body { background: #100c18; }
        .header-bar { background: rgba(22,16,36,0.8); color: #f8fafc; padding: 0.6rem 2rem; margin-bottom: 0; display: flex; justify-content: space-between; align-items: center; position: sticky; top: 0; z-index: 100; overflow: hidden; border-bottom: 1px solid rgba(255,255,255,0.06); backdrop-filter: blur(14px); -webkit-backdrop-filter: blur(14px); box-shadow: 0 1px 12px rgba(0,0,0,0.4), 0 0 40px rgba(120,80,255,0.06); transition: padding 0.3s ease, box-shadow 0.3s ease; }
        .header-bar::before { content: ''; position: absolute; top: 0; left: 0; right: 0; bottom: 0; background: linear-gradient(135deg, rgba(26,18,48,0.92) 0%, rgba(36,24,64,0.88) 50%, rgba(42,20,48,0.92) 100%); pointer-events: none; }
        .header-bar::after { content: ''; position: absolute; bottom: 0; left: 0; right: 0; height: 1px; background: linear-gradient(90deg, transparent 5%, rgba(240,160,48,0.25) 30%, rgba(200,56,56,0.15) 70%, transparent 95%); pointer-events: none; }
        .header-bar > * { position: relative; z-index: 1; }
        .header-bar .header-left { display: flex; align-items: center; gap: 1.25rem; flex: 1; }
        .header-bar .header-left img { height: 110px; border-radius: 8px; filter: drop-shadow(1px 1px 3px rgba(0,0,0,0.4)); transition: transform 0.3s ease, height 0.3s ease; transform: translateY(1px); }
        .header-bar .header-left img:hover { transform: translateY(1px) scale(1.08) rotate(-2deg); }
        .header-bar .header-brand { display: flex; flex-direction: column; align-items: center; gap: 0.1rem; flex: 2; }
        .header-bar h1 { margin: 0; font-family: 'Press Start 2P', monospace; font-size: 2.1rem; font-weight: 400; color: #f8fafc; letter-spacing: 0.06em; text-transform: uppercase; line-height: 1.2; transition: font-size 0.3s ease; text-shadow: 0 0 10px rgba(240,160,48,0.4), 0 0 20px rgba(240,160,48,0.15); }
        .header-bar .subtitle { font-size: 0.9rem; color: rgba(248,250,252,0.6); letter-spacing: 0.08em; text-transform: uppercase; font-weight: 500; padding-bottom: 0.15rem; transition: font-size 0.3s ease; }
        .header-bar .nav-links { display: flex; gap: 0.2rem; align-items: flex-end; padding-bottom: 0; flex: 1; justify-content: flex-end; align-self: flex-end; }
        .header-bar .nav-links a { color: rgba(248,250,252,0.5); font-size: 0.8rem; font-weight: 500; text-decoration: none; padding: 0.35rem 0.85rem; border-radius: 6px; transition: all 0.2s ease; position: relative; }
        .header-bar .nav-links a:hover { color: #f8fafc; background: rgba(255,255,255,0.07); }
        .header-bar .nav-links a.active { color: #fff; background: rgba(240,160,48,0.12); font-weight: 600; text-shadow: 0 0 8px rgba(240,160,48,0.35); }
        .header-bar .nav-links a.active::after { content: ''; position: absolute; bottom: -0.45rem; left: 20%; right: 20%; height: 2px; background: #f0a030; border-radius: 1px; box-shadow: 0 0 6px rgba(240,160,48,0.4); }
        table { font-size: 0.85rem; }
        thead th { background: #1e1a28; color: #c8a040; text-transform: uppercase; }
        .clickable { cursor: pointer; }
        .clickable:hover { background: rgba(240,160,48,0.1); }
        .detail-panel { background: #1e1a28; border: 1px solid #2a2434; padding: 1.5rem; border-radius: 0.5rem; margin-bottom: 1rem; }
        .detail-panel dt { color: #c8a040; font-weight: bold; }
        .detail-panel dd { margin-bottom: 0.5rem; margin-left: 0; color: #e0dce8; }
        .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 1.25rem; margin-bottom: 1rem; margin-top: 0.5rem; }
        .card { background: #18141f; border-radius: 12px; padding: 1.5rem; border: 1px solid #241e2e; transition: all 0.2s; position: relative; overflow: hidden; }
        .card::before { content: ''; position: absolute; top: 0; left: 0; right: 0; height: 2px; background: linear-gradient(90deg, #f0a030, #f05050, #4a30c8); opacity: 0; transition: opacity 0.2s; }
        .card:hover { border-color: #3a3048; transform: translateY(-2px); box-shadow: 0 8px 24px rgba(0,0,0,0.3); }
        .card:hover::before { opacity: 1; }
        .card a { color: inherit; text-decoration: none; }
        .badge { display: inline-block; padding: 0.15rem 0.5rem; border-radius: 4px; font-size: 0.7rem; font-weight: 600; }
        .status-dot { display: inline-block; width: 8px; height: 8px; border-radius: 50%; animation: pulse 2s infinite; }
        .status-up { background: #3cc878; box-shadow: 0 0 6px rgba(60,200,120,0.4); }
        .status-degraded { background: #d4943a; box-shadow: 0 0 6px rgba(212,148,58,0.4); }
        .status-down { background: #d04848; box-shadow: 0 0 6px rgba(208,72,72,0.4); }
        .status-unknown { background: #3a3048; animation: none; }
        @keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
        .response-time { font-size: 0.65rem; color: #5a5068; margin-left: 0.5rem; font-family: monospace; }
        .mini-timeline { display: flex; gap: 2px; align-items: flex-end; height: 24px; }
        .timeline-bar { width: 5px; height: 100%; border-radius: 2px; }
        .category-filter { display: flex; gap: 0.4rem; flex-wrap: wrap; margin-bottom: 1.5rem; }
        .category-btn { padding: 0.35rem 0.8rem; border-radius: 6px; border: 1px solid #2a2434; background: transparent; color: #6a6080; cursor: pointer; font-size: 0.75rem; font-weight: 600; transition: all 0.2s; }
        .category-btn:hover { color: #e0dce8; border-color: #3a3048; background: rgba(240,160,48,0.05); }
        .category-btn.active { background: rgba(240,160,48,0.12); color: #f0a030; border-color: rgba(240,160,48,0.3); }
        .tls-warning { display: inline-block; background: rgba(240,160,48,0.1); color: #f0a030; padding: 0.1rem 0.4rem; border-radius: 4px; font-size: 0.6rem; font-weight: 600; }
        .build-footer { background: #1e1a28; color: #475569; padding: 0.6rem 1.5rem; margin-top: 1rem; display: flex; gap: 1.5rem; font-size: 0.75rem; border-top: 1px solid #2a2434; }
        .build-footer .label { color: rgba(248,250,252,0.85); }
        .build-footer .value { color: #8a7aa0; }
        .btn-action { padding: 0.35rem 0.75rem; border-radius: 6px; border: none; cursor: pointer; font-size: 0.75rem; font-weight: 600; transition: all 0.2s; }
        .btn-add { background: #f0a030; color: #1a1014; }
        .btn-add:hover { background: #f0b848; }
        .btn-del { background: #f05050; color: white; }
        .btn-del:hover { background: #f06868; }
        .btn-edit { background: #4a6cf0; color: white; }
        .btn-edit:hover { background: #5a7cf8; }
        .btn-cancel { background: #3a3048; color: #a0a0b0; }
        .btn-cancel:hover { background: #4a4060; }
        .tag { display: inline-block; padding: 0.1rem 0.45rem; border-radius: 4px; font-size: 0.65rem; font-weight: 600; background: rgba(74,108,240,0.1); color: #6a8af0; margin-right: 0.25rem; }
        .edit-row td { background: #1a1424 !important; }
        .edit-row .edit-form { padding: 0.75rem 0; }
        .header-bar.scrolled { padding: 0.3rem 2rem; box-shadow: 0 1px 8px rgba(0,0,0,0.5), 0 0 30px rgba(120,80,255,0.04); }
        .header-bar.scrolled .header-left img { height: 60px; }
        .header-bar.scrolled h1 { font-size: 0.75rem; }
        .header-bar.scrolled .subtitle { font-size: 0.5rem; }
        .container-fluid { padding-top: 0.75rem; }
    </style>
    <script>
        window.addEventListener('scroll', function() {
            var h = document.querySelector('.header-bar');
            if (h) h.classList.toggle('scrolled', window.scrollY > 60);
        });
    </script>
`

const headerBar = `
    <div class="header-bar">
        <div class="header-left">
            <a href="/"><img src="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/run-sthings2.png" alt="we run things"></a>
        </div>
        <div class="header-brand">
            <h1>WE RUN THINGS</h1>
            <div class="subtitle">Service Portal &amp; Health Monitor</div>
        </div>
        <div class="nav-links">
            <a href="/" {{.DashboardActive}}>Dashboard</a>
            <a href="/clusters" {{.ClustersActive}}>Clusters</a>
            <a href="/admin" {{.AdminActive}}>Admin</a>
        </div>
    </div>
`

const buildFooter = `
    <div class="build-footer">
        <div style="display:flex;gap:1.5rem">
            <div><span class="label">version</span> <span class="value">{{.Version}}</span></div>
            <div><span class="label">commit</span> <span class="value">{{if gt (len .Commit) 7}}{{slice .Commit 0 7}}{{else}}{{.Commit}}{{end}}</span></div>
            <div><span class="label">built</span> <span class="value">{{.Date}}</span></div>
        </div>
        <div style="margin-left:auto;display:flex;align-items:center;gap:0.5rem"><span class="label">a</span> <a href="https://github.com/stuttgart-things" target="_blank" style="color:#c8a040;text-decoration:none">stuttgart-things</a> <span class="label">project</span> <img src="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/sthings-logo.png" alt="sthings" style="height:24px"></div>
    </div>
`

const dashboardTemplate = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>we run things</title>
    <link rel="icon" type="image/png" href="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/sthings-argo.png">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <link href="https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap" rel="stylesheet">
` + commonStyles + `
</head>
<body>
    <div class="header-bar">
        <div class="header-left">
            <a href="/"><img src="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/run-sthings2.png" alt="we run things"></a>
        </div>
        <div class="header-brand">
            <h1>WE RUN THINGS</h1>
            <div class="subtitle">Service Portal &amp; Health Monitor</div>
        </div>
        <div class="nav-links">
            <a href="/" class="active">Dashboard</a>
            <a href="/clusters">Clusters</a>
            <a href="/admin">Admin</a>
        </div>
    </div>
    <main class="container-fluid">
        {{if .Categories}}
        <div class="category-filter" style="margin-top:1rem;">
            <button class="category-btn active" hx-get="/htmx/cards" hx-target="#cards-grid" hx-swap="innerHTML">All</button>
            {{range .Categories}}
            <button class="category-btn" hx-get="/htmx/cards?category={{.}}" hx-target="#cards-grid" hx-swap="innerHTML" onclick="document.querySelectorAll('.category-btn').forEach(b=>b.classList.remove('active'));this.classList.add('active')">{{.}}</button>
            {{end}}
        </div>
        {{end}}
        <div id="cards-grid" class="grid">
        ` + cardsPartial + `
        </div>
    </main>
    <div class="build-footer">
        <div style="display:flex;gap:1.5rem">
            <div><span class="label">version</span> <span class="value">{{.Version}}</span></div>
            <div><span class="label">commit</span> <span class="value">{{if gt (len .Commit) 7}}{{slice .Commit 0 7}}{{else}}{{.Commit}}{{end}}</span></div>
            <div><span class="label">built</span> <span class="value">{{.Date}}</span></div>
        </div>
        <div style="margin-left:auto;display:flex;align-items:center;gap:0.5rem"><span class="label">a</span> <a href="https://github.com/stuttgart-things" target="_blank" style="color:#c8a040;text-decoration:none">stuttgart-things</a> <span class="label">project</span> <img src="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/sthings-logo.png" alt="sthings" style="height:24px"></div>
    </div>
</body>
</html>`

const cardsPartial = `{{range .Services}}
<div class="card">
    <div style="display:flex;align-items:center;gap:1rem;margin-bottom:0.75rem;">
        {{if .LogoURL}}<img src="{{.LogoURL}}" alt="{{.Name}}" style="width:44px;height:44px;border-radius:8px;background:#1e1a28;padding:4px;">{{else if .Icon}}<span style="font-size:2rem;">{{.Icon}}</span>{{end}}
        <div style="flex:1;min-width:0;">
            <div style="display:flex;align-items:center;gap:0.5rem;">
                <a href="{{.URL}}" target="_blank" style="font-size:1rem;font-weight:700;color:#f8fafc;">{{.Name}}</a>
                <span hx-get="/htmx/status/{{.Name}}" hx-trigger="every 30s" hx-swap="innerHTML" style="display:flex;align-items:center;gap:0.25rem;">
                    <span class="status-dot status-{{lower .Status}}"></span>
                    <span class="response-time">{{if .ResponseTime}}{{.ResponseTime}}ms{{end}}</span>
                </span>
            </div>
            <div style="font-size:0.78rem;color:#6a6080;margin-top:0.15rem;white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">{{.Description}}</div>
        </div>
    </div>
    <div style="display:flex;align-items:center;justify-content:space-between;">
        <div style="display:flex;align-items:center;gap:0.5rem;">
            <span class="badge" style="background:rgba(240,160,48,0.08);color:#8a7a60;">{{.Category}}</span>
            {{if and .TLSDaysLeft (gt .TLSDaysLeft 0)}}{{if le .TLSDaysLeft 30}}<span class="tls-warning">TLS {{.TLSDaysLeft}}d</span>{{end}}{{end}}
        </div>
        <div hx-get="/htmx/history/{{.Name}}" hx-trigger="every 30s" hx-swap="innerHTML" class="mini-timeline">
            {{range .History}}<div class="timeline-bar" style="background:{{.}}"></div>{{end}}
        </div>
    </div>
    <div style="margin-top:0.5rem;text-align:right;">
        <a href="/service/{{.Name}}" style="font-size:0.7rem;color:#c8a040;">details &rarr;</a>
    </div>
</div>
{{end}}`

const serviceDetailTemplate = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>we run things - {{.Service.Name}}</title>
    <link rel="icon" type="image/png" href="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/sthings-argo.png">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <link href="https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap" rel="stylesheet">
` + commonStyles + `
</head>
<body>
    <div class="header-bar">
        <div class="header-left">
            <a href="/"><img src="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/run-sthings2.png" alt="we run things"></a>
        </div>
        <div class="header-brand">
            <h1>WE RUN THINGS</h1>
            <div class="subtitle">Service Portal &amp; Health Monitor</div>
        </div>
        <div class="nav-links">
            <a href="/">Dashboard</a>
            <a href="/clusters">Clusters</a>
            <a href="/admin">Admin</a>
        </div>
    </div>
    <main class="container-fluid">
        <div style="display:flex;align-items:center;gap:1rem;margin-bottom:1.5rem;">
            <a href="/" style="font-size:1.3rem;color:#c8a040;text-decoration:none;">&larr;</a>
            <h2 style="margin:0;">{{.Service.Name}}</h2>
            <span class="badge" style="background:{{statusBg .Status}};color:{{statusColor .Status}};">{{.Status}}</span>
        </div>
        <div class="detail-panel" style="margin-bottom:1.5rem;">
            <dl>
                <dt>Description</dt><dd>{{.Service.Description}}</dd>
                <dt>URL</dt><dd><a href="{{.Service.URL}}" target="_blank" style="color:#c8a040;">{{.Service.URL}}</a></dd>
                <dt>Category</dt><dd><span class="badge" style="background:rgba(240,160,48,0.08);color:#8a7a60;">{{.Service.Category}}</span></dd>
            </dl>
        </div>
        <h3>Health Check History</h3>
        <table>
            <thead>
                <tr><th>Time</th><th>Status</th><th>Code</th><th>Response Time</th><th>TLS Days Left</th><th>Error</th></tr>
            </thead>
            <tbody>
                {{range .Results}}
                <tr>
                    <td style="font-family:monospace;font-size:0.8rem;">{{.Timestamp.Format "2006-01-02 15:04:05"}}</td>
                    <td><span class="badge" style="background:{{statusBg .Status}};color:{{statusColor .Status}};">{{.Status}}</span></td>
                    <td>{{.StatusCode}}</td>
                    <td style="font-family:monospace;">{{.ResponseTime}}ms</td>
                    <td>{{if .TLSDaysLeft}}{{.TLSDaysLeft}}d{{else}}-{{end}}</td>
                    <td style="font-size:0.8rem;color:#f05050;">{{.Error}}</td>
                </tr>
                {{end}}
            </tbody>
        </table>
    </main>
</body>
</html>`

const clustersTemplate = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>we run things - Clusters</title>
    <link rel="icon" type="image/png" href="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/sthings-argo.png">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <link href="https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap" rel="stylesheet">
` + commonStyles + `
</head>
<body>
    <div class="header-bar">
        <div class="header-left">
            <a href="/"><img src="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/run-sthings2.png" alt="we run things"></a>
        </div>
        <div class="header-brand">
            <h1>WE RUN THINGS</h1>
            <div class="subtitle">Service Portal &amp; Health Monitor</div>
        </div>
        <div class="nav-links">
            <a href="/">Dashboard</a>
            <a href="/clusters" class="active">Clusters</a>
            <a href="/admin">Admin</a>
        </div>
    </div>
    <main class="container-fluid">
        <h2>Connected Clusters</h2>
        {{if .Clusters}}
        <div class="grid">
            {{range .Clusters}}
            <div class="card">
                <a href="/cluster/{{.ClusterName}}" style="display:block;">
                    <div style="font-size:1.1rem;font-weight:700;color:#f8fafc;margin-bottom:0.5rem;">{{.ClusterName}}</div>
                    {{if .Endpoint}}<div style="font-size:0.78rem;color:#6a6080;margin-bottom:0.5rem;font-family:monospace;">{{.Endpoint}}</div>{{end}}
                    <div style="font-size:0.7rem;color:#4a4060;">Last seen: {{.LastSeen.Format "2006-01-02 15:04:05"}}</div>
                </a>
            </div>
            {{end}}
        </div>
        {{else}}
        <div class="detail-panel" style="text-align:center;padding:3rem;">
            <p style="color:#6a6080;font-size:0.95rem;">No clusters connected yet.</p>
            <p style="color:#4a4060;font-size:0.8rem;margin-top:0.5rem;">Deploy collectors to your clusters to see inventory data here.</p>
        </div>
        {{end}}
    </main>
</body>
</html>`

const clusterDetailTemplate = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>we run things - {{.Inventory.ClusterName}}</title>
    <link rel="icon" type="image/png" href="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/sthings-argo.png">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <link href="https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap" rel="stylesheet">
` + commonStyles + `
</head>
<body>
    <div class="header-bar">
        <div class="header-left">
            <a href="/"><img src="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/run-sthings2.png" alt="we run things"></a>
        </div>
        <div class="header-brand">
            <h1>WE RUN THINGS</h1>
            <div class="subtitle">Service Portal &amp; Health Monitor</div>
        </div>
        <div class="nav-links">
            <a href="/">Dashboard</a>
            <a href="/clusters" class="active">Clusters</a>
            <a href="/admin">Admin</a>
        </div>
    </div>
    <main class="container-fluid">
        <div style="display:flex;align-items:center;gap:1rem;margin-bottom:1.5rem;">
            <a href="/clusters" style="font-size:1.3rem;color:#c8a040;text-decoration:none;">&larr;</a>
            <h2 style="margin:0;">{{.Inventory.ClusterName}}</h2>
            <span style="font-size:0.75rem;color:#4a4060;font-family:monospace;">Updated: {{.Inventory.LastUpdated.Format "2006-01-02 15:04:05"}}</span>
        </div>

        {{if .Inventory.Deployments}}
        <h3>Deployments ({{len .Inventory.Deployments}})</h3>
        <table>
            <thead><tr><th>Name</th><th>Namespace</th><th>Replicas</th><th>Ready</th><th>Images</th><th>Age</th></tr></thead>
            <tbody>
            {{range .Inventory.Deployments}}
            <tr>
                <td style="font-weight:600;">{{.Name}}</td>
                <td><span class="badge" style="background:rgba(240,160,48,0.08);color:#8a7a60;">{{.Namespace}}</span></td>
                <td>{{.Replicas}}</td><td>{{.Ready}}</td>
                <td style="font-size:0.75rem;font-family:monospace;color:#6a6080;">{{range .Images}}{{.}}<br>{{end}}</td>
                <td style="font-size:0.8rem;color:#4a4060;">{{.CreationTime.Format "2006-01-02"}}</td>
            </tr>
            {{end}}
            </tbody>
        </table>
        {{end}}

        {{if .Inventory.StatefulSets}}
        <h3>StatefulSets ({{len .Inventory.StatefulSets}})</h3>
        <table>
            <thead><tr><th>Name</th><th>Namespace</th><th>Replicas</th><th>Ready</th><th>Images</th><th>Age</th></tr></thead>
            <tbody>
            {{range .Inventory.StatefulSets}}
            <tr>
                <td style="font-weight:600;">{{.Name}}</td>
                <td><span class="badge" style="background:rgba(240,160,48,0.08);color:#8a7a60;">{{.Namespace}}</span></td>
                <td>{{.Replicas}}</td><td>{{.Ready}}</td>
                <td style="font-size:0.75rem;font-family:monospace;color:#6a6080;">{{range .Images}}{{.}}<br>{{end}}</td>
                <td style="font-size:0.8rem;color:#4a4060;">{{.CreationTime.Format "2006-01-02"}}</td>
            </tr>
            {{end}}
            </tbody>
        </table>
        {{end}}

        {{if .Inventory.DaemonSets}}
        <h3>DaemonSets ({{len .Inventory.DaemonSets}})</h3>
        <table>
            <thead><tr><th>Name</th><th>Namespace</th><th>Replicas</th><th>Ready</th><th>Images</th><th>Age</th></tr></thead>
            <tbody>
            {{range .Inventory.DaemonSets}}
            <tr>
                <td style="font-weight:600;">{{.Name}}</td>
                <td><span class="badge" style="background:rgba(240,160,48,0.08);color:#8a7a60;">{{.Namespace}}</span></td>
                <td>{{.Replicas}}</td><td>{{.Ready}}</td>
                <td style="font-size:0.75rem;font-family:monospace;color:#6a6080;">{{range .Images}}{{.}}<br>{{end}}</td>
                <td style="font-size:0.8rem;color:#4a4060;">{{.CreationTime.Format "2006-01-02"}}</td>
            </tr>
            {{end}}
            </tbody>
        </table>
        {{end}}

        {{if .Inventory.Services}}
        <h3>Services ({{len .Inventory.Services}})</h3>
        <table>
            <thead><tr><th>Name</th><th>Namespace</th><th>Kind</th></tr></thead>
            <tbody>
            {{range .Inventory.Services}}<tr><td style="font-weight:600;">{{.Name}}</td><td><span class="badge" style="background:rgba(240,160,48,0.08);color:#8a7a60;">{{.Namespace}}</span></td><td>{{.Kind}}</td></tr>{{end}}
            </tbody>
        </table>
        {{end}}

        {{if .Inventory.Ingresses}}
        <h3>Ingresses ({{len .Inventory.Ingresses}})</h3>
        <table>
            <thead><tr><th>Name</th><th>Namespace</th><th>Kind</th></tr></thead>
            <tbody>
            {{range .Inventory.Ingresses}}<tr><td style="font-weight:600;">{{.Name}}</td><td><span class="badge" style="background:rgba(240,160,48,0.08);color:#8a7a60;">{{.Namespace}}</span></td><td>{{.Kind}}</td></tr>{{end}}
            </tbody>
        </table>
        {{end}}
    </main>
</body>
</html>`

const adminTemplate = `<!DOCTYPE html>
<html lang="en" data-theme="dark">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>we run things - Admin</title>
    <link rel="icon" type="image/png" href="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/sthings-argo.png">
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <link href="https://fonts.googleapis.com/css2?family=Press+Start+2P&display=swap" rel="stylesheet">
` + commonStyles + `
</head>
<body>
    <div class="header-bar">
        <div class="header-left">
            <a href="/"><img src="https://raw.githubusercontent.com/stuttgart-things/docs/main/hugo/run-sthings2.png" alt="we run things"></a>
        </div>
        <div class="header-brand">
            <h1>WE RUN THINGS</h1>
            <div class="subtitle">Service Portal &amp; Health Monitor</div>
        </div>
        <div class="nav-links">
            <a href="/">Dashboard</a>
            <a href="/clusters">Clusters</a>
            <a href="/admin" class="active">Admin</a>
        </div>
    </div>
    <main class="container-fluid">
        <h2>Service Administration</h2>

        <div class="detail-panel" style="margin-bottom:1.5rem;">
            <h3 style="margin-top:0;">Add New Service</h3>
            <form hx-post="/htmx/add-service" hx-swap="none">
                <div class="grid" style="grid-template-columns:1fr 1fr;">
                    <label>Name * <input type="text" name="name" required></label>
                    <label>URL * <input type="url" name="url" required></label>
                    <label>Description <input type="text" name="description"></label>
                    <label>Category <input type="text" name="category"></label>
                    <label>Logo URL <input type="url" name="logo_url"></label>
                    <label>Icon (emoji) <input type="text" name="icon"></label>
                    <label style="grid-column:1/-1;">Tags <input type="text" name="tags" placeholder="comma-separated, e.g. gitops, kubernetes, monitoring"></label>
                </div>
                <div style="display:flex;align-items:center;gap:1.5rem;margin-top:0.5rem;">
                    <label style="display:flex;align-items:center;gap:0.35rem;font-size:0.85rem;cursor:pointer;">
                        <input type="checkbox" name="health_enabled" checked> Health Check
                    </label>
                    <label style="display:flex;align-items:center;gap:0.35rem;font-size:0.85rem;cursor:pointer;">
                        <input type="checkbox" name="tls_check"> TLS Check
                    </label>
                    <button type="submit" class="btn-action btn-add">Add Service</button>
                </div>
            </form>
        </div>

        <h3>Existing Services ({{len .Services}})</h3>
        <table>
            <thead>
                <tr><th>Name</th><th>Category</th><th>URL</th><th>Tags</th><th>Health</th><th>Actions</th></tr>
            </thead>
            <tbody>
                {{range .Services}}
                <tr id="row-{{.Name}}">
                    <td style="font-weight:600;">{{.Name}}</td>
                    <td><span class="badge" style="background:rgba(240,160,48,0.08);color:#8a7a60;">{{.Category}}</span></td>
                    <td style="font-size:0.8rem;"><a href="{{.URL}}" target="_blank" style="color:#c8a040;">{{.URL}}</a></td>
                    <td>{{range .Tags}}<span class="tag">{{.}}</span>{{end}}</td>
                    <td>{{if .HealthCheck.Enabled}}<span style="color:#44dd88;font-weight:bold;">On</span>{{else}}<span style="color:#4a4060;">Off</span>{{end}}</td>
                    <td style="white-space:nowrap;">
                        <button class="btn-action btn-edit" onclick="toggleEdit('{{.Name}}')">Edit</button>
                        <form hx-post="/htmx/delete-service" hx-swap="none" hx-confirm="Delete {{.Name}}?" style="display:inline;">
                            <input type="hidden" name="name" value="{{.Name}}">
                            <button type="submit" class="btn-action btn-del">Delete</button>
                        </form>
                    </td>
                </tr>
                <tr id="edit-{{.Name}}" class="edit-row" style="display:none;">
                    <td colspan="6">
                        <form hx-post="/htmx/edit-service" hx-swap="none" class="edit-form">
                            <input type="hidden" name="name" value="{{.Name}}">
                            <div class="grid" style="grid-template-columns:1fr 1fr 1fr;">
                                <label>URL * <input type="url" name="url" value="{{.URL}}" required></label>
                                <label>Description <input type="text" name="description" value="{{.Description}}"></label>
                                <label>Category <input type="text" name="category" value="{{.Category}}"></label>
                                <label>Logo URL <input type="url" name="logo_url" value="{{.LogoURL}}"></label>
                                <label>Icon (emoji) <input type="text" name="icon" value="{{.Icon}}"></label>
                                <label>Tags <input type="text" name="tags" value="{{joinTags .Tags}}"></label>
                            </div>
                            <div style="display:flex;align-items:center;gap:1.5rem;margin-top:0.5rem;">
                                <label style="display:flex;align-items:center;gap:0.35rem;font-size:0.85rem;cursor:pointer;">
                                    <input type="checkbox" name="health_enabled" {{if .HealthCheck.Enabled}}checked{{end}}> Health Check
                                </label>
                                <label style="display:flex;align-items:center;gap:0.35rem;font-size:0.85rem;cursor:pointer;">
                                    <input type="checkbox" name="tls_check" {{if .HealthCheck.TLSCheck}}checked{{end}}> TLS Check
                                </label>
                                <button type="submit" class="btn-action btn-add">Save</button>
                                <button type="button" class="btn-action btn-cancel" onclick="toggleEdit('{{.Name}}')">Cancel</button>
                            </div>
                        </form>
                    </td>
                </tr>
                {{end}}
            </tbody>
        </table>
        <script>
        function toggleEdit(name) {
            var editRow = document.getElementById('edit-' + name);
            if (editRow.style.display === 'none') {
                document.querySelectorAll('.edit-row').forEach(function(r) { r.style.display = 'none'; });
                editRow.style.display = 'table-row';
            } else {
                editRow.style.display = 'none';
            }
        }
        </script>
    </main>
</body>
</html>`
