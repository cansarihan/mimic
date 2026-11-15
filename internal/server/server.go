// Package server wraps the analyses in an HTTP API and serves the embedded
// multi-page panel. It reads the Go source under one fixed directory chosen at
// startup, so a request can never point it at an arbitrary path on the host.
package server

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cansarihan/mimic/internal/rules"
	"github.com/cansarihan/mimic/internal/scan"
	"github.com/cansarihan/mimic/internal/srcwalk"
)

// BuildInfo is the version metadata reported on the status endpoint.
type BuildInfo struct {
	Version   string
	Commit    string
	GoVersion string
	Platform  string
}

// Options configures a server instance.
type Options struct {
	Dir   string
	Token string
	UI    bool
}

// Server answers analysis requests over one source directory.
type Server struct {
	options Options
	info    BuildInfo
	started time.Time
}

// New builds a server over one source directory.
func New(options Options, info BuildInfo) *Server {
	return &Server{options: options, info: info, started: time.Now()}
}

// Handler returns the HTTP handler for the API and, when enabled, the panel.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("GET /api/v1/status", s.guard(s.handleStatus))
	mux.HandleFunc("GET /api/v1/rules", s.guard(s.handleRules))
	mux.HandleFunc("GET /api/v1/findings", s.guard(s.handleFindings))
	mux.HandleFunc("GET /api/v1/functions", s.guard(s.handleFunctions))
	mux.HandleFunc("GET /api/v1/like", s.guard(s.handleLike))
	mux.HandleFunc("GET /api/v1/clusters", s.guard(s.handleClusters))
	if s.options.UI {
		mux.Handle("GET /", uiHandler())
	}
	return mux
}

func (s *Server) guard(next http.HandlerFunc) http.HandlerFunc {
	token := s.options.Token
	return func(w http.ResponseWriter, r *http.Request) {
		if token != "" {
			provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if provided == "" {
				provided = r.URL.Query().Get("token")
			}
			if subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
				return
			}
		}
		next(w, r)
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	report, err := scan.Rules(s.options.Dir, rules.Registry())
	funcs := report.Functions
	findings := len(report.Findings)
	if err != nil {
		funcs, findings = 0, 0
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version":    s.info.Version,
		"commit":     s.info.Commit,
		"go_version": s.info.GoVersion,
		"platform":   s.info.Platform,
		"dir":        s.options.Dir,
		"uptime":     time.Since(s.started).Round(time.Second).String(),
		"functions":  funcs,
		"findings":   findings,
		"rules":      len(rules.Registry()),
	})
}

func (s *Server) handleRules(w http.ResponseWriter, _ *http.Request) {
	type rule struct {
		Name     string `json:"name"`
		Describe string `json:"describe"`
	}
	out := make([]rule, 0)
	for _, m := range rules.Registry() {
		out = append(out, rule{Name: m.Name(), Describe: m.Describe()})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleFindings(w http.ResponseWriter, _ *http.Request) {
	report, err := scan.Rules(s.options.Dir, rules.Registry())
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleFunctions(w http.ResponseWriter, _ *http.Request) {
	funcs, err := srcwalk.Walk(s.options.Dir)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	type fn struct {
		Name string `json:"name"`
		File string `json:"file"`
		Line int    `json:"line"`
	}
	out := make([]fn, 0, len(funcs))
	for _, f := range funcs {
		out = append(out, fn{Name: f.Name, File: f.File, Line: f.Line})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleLike(w http.ResponseWriter, r *http.Request) {
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ref is required"})
		return
	}
	min := 0.6
	if v := r.URL.Query().Get("min"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			min = parsed
		}
	}
	report, err := scan.Like(s.options.Dir, ref, min)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleClusters(w http.ResponseWriter, r *http.Request) {
	min := 0.75
	if v := r.URL.Query().Get("min"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			min = parsed
		}
	}
	report, err := scan.Clusters(s.options.Dir, min)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
