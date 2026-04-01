package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/stockyard-dev/stockyard-rustler/internal/store"
)

type Server struct {
	db     *store.DB
	mux    *http.ServeMux
	port   int
	limits Limits
}

func New(db *store.DB, port int, limits Limits) *Server {
	s := &Server{db: db, mux: http.NewServeMux(), port: port, limits: limits}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("POST /api/scans", s.handleCreateScan)
	s.mux.HandleFunc("GET /api/scans", s.handleListScans)
	s.mux.HandleFunc("GET /api/scans/{id}", s.handleGetScan)
	s.mux.HandleFunc("DELETE /api/scans/{id}", s.handleDeleteScan)
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /ui", s.handleUI)
	s.mux.HandleFunc("GET /api/version", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{"product": "stockyard-rustler", "version": "0.1.0"})
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("[rustler] listening on %s", addr)
	return http.ListenAndServe(addr, s.mux)
}

type LinkResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error,omitempty"`
	Broken     bool   `json:"broken"`
	SSLExpiry  string `json:"ssl_expiry,omitempty"`
	SSLIssue   string `json:"ssl_issue,omitempty"`
	FoundOn    string `json:"found_on"`
}

var hrefRe = regexp.MustCompile(`(?i)href=["']([^"']+)["']`)

func (s *Server) runScan(scan *store.Scan) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	baseURL, err := url.Parse(scan.URL)
	if err != nil {
		s.db.FailScan(scan.ID, err.Error())
		return
	}

	// Fetch the page
	resp, err := client.Get(scan.URL)
	if err != nil {
		s.db.FailScan(scan.ID, err.Error())
		return
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	resp.Body.Close()

	// Extract links
	matches := hrefRe.FindAllStringSubmatch(string(body), -1)
	seen := make(map[string]bool)
	var results []LinkResult
	var broken, sslIssues int

	for _, m := range matches {
		href := m[1]
		if strings.HasPrefix(href, "#") || strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "javascript:") {
			continue
		}

		linkURL, err := url.Parse(href)
		if err != nil {
			continue
		}
		resolved := baseURL.ResolveReference(linkURL).String()
		if seen[resolved] {
			continue
		}
		seen[resolved] = true

		result := LinkResult{URL: resolved, FoundOn: scan.URL}

		// Check the link
		linkResp, err := client.Get(resolved)
		if err != nil {
			result.Error = err.Error()
			result.Broken = true
			broken++
		} else {
			result.StatusCode = linkResp.StatusCode
			io.Copy(io.Discard, linkResp.Body)
			linkResp.Body.Close()
			if linkResp.StatusCode >= 400 {
				result.Broken = true
				broken++
			}
		}

		// SSL check for HTTPS links
		if s.limits.SSLCheck && strings.HasPrefix(resolved, "https://") {
			parsedLink, _ := url.Parse(resolved)
			if parsedLink != nil {
				conn, err := tls.DialWithDialer(
					&net.Dialer{Timeout: 5 * time.Second},
					"tcp", parsedLink.Host+":443",
					&tls.Config{InsecureSkipVerify: true},
				)
				if err == nil {
					certs := conn.ConnectionState().PeerCertificates
					if len(certs) > 0 {
						expiry := certs[0].NotAfter
						result.SSLExpiry = expiry.Format("2006-01-02")
						if time.Until(expiry) < 30*24*time.Hour {
							result.SSLIssue = "expires within 30 days"
							sslIssues++
						}
						if time.Now().After(expiry) {
							result.SSLIssue = "expired"
							sslIssues++
						}
					}
					conn.Close()
				}
			}
		}

		results = append(results, result)
	}

	resultsJSON, _ := json.Marshal(results)
	s.db.CompleteScan(scan.ID, len(results), broken, sslIssues, string(resultsJSON))
	log.Printf("[scan] %s complete: %d links, %d broken, %d ssl issues", scan.ID, len(results), broken, sslIssues)
}

func (s *Server) handleCreateScan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		writeJSON(w, 400, map[string]string{"error": "url is required"})
		return
	}
	if !strings.HasPrefix(req.URL, "http") {
		req.URL = "https://" + req.URL
	}
	scan, err := s.db.CreateScan(req.URL)
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	go s.runScan(scan)
	writeJSON(w, 201, map[string]any{"scan": scan})
}

func (s *Server) handleListScans(w http.ResponseWriter, r *http.Request) {
	scans, _ := s.db.ListScans()
	if scans == nil { scans = []store.Scan{} }
	writeJSON(w, 200, map[string]any{"scans": scans, "count": len(scans)})
}

func (s *Server) handleGetScan(w http.ResponseWriter, r *http.Request) {
	scan, err := s.db.GetScan(r.PathValue("id"))
	if err != nil {
		writeJSON(w, 404, map[string]string{"error": "scan not found"})
		return
	}
	writeJSON(w, 200, map[string]any{"scan": scan})
}

func (s *Server) handleDeleteScan(w http.ResponseWriter, r *http.Request) {
	s.db.DeleteScan(r.PathValue("id"))
	writeJSON(w, 200, map[string]string{"status": "deleted"})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) { writeJSON(w, 200, s.db.Stats()) }
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
