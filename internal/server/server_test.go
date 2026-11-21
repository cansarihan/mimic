package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cansarihan/mimic/internal/scan"
)

const sampleDir = "../../deploy/sample"

func testHandler(token string) http.Handler {
	return New(Options{Dir: sampleDir, Token: token, UI: false},
		BuildInfo{Version: "test", Commit: "abc", GoVersion: "go1", Platform: "test/arch"}).Handler()
}

func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(testHandler(""))
	defer srv.Close()
	res, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	_ = res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("healthz = %d", res.StatusCode)
	}
}

func TestTokenGuard(t *testing.T) {
	srv := httptest.NewServer(testHandler("secret"))
	defer srv.Close()
	res, _ := http.Get(srv.URL + "/api/v1/status")
	_ = res.Body.Close()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token = %d, want 401", res.StatusCode)
	}
	res2, _ := http.Get(srv.URL + "/api/v1/status?token=secret")
	_ = res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("with token = %d, want 200", res2.StatusCode)
	}
}

func TestFindingsEndpoint(t *testing.T) {
	srv := httptest.NewServer(testHandler(""))
	defer srv.Close()
	res, err := http.Get(srv.URL + "/api/v1/findings")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	var report scan.RuleReport
	if err := json.NewDecoder(res.Body).Decode(&report); err != nil {
		t.Fatal(err)
	}
	if len(report.Findings) < 3 {
		t.Errorf("expected several findings in the sample, got %d", len(report.Findings))
	}
}

func TestLikeEndpoint(t *testing.T) {
	srv := httptest.NewServer(testHandler(""))
	defer srv.Close()
	res, err := http.Get(srv.URL + "/api/v1/like?ref=validateUser&min=0.5")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	var report scan.LikeReport
	if err := json.NewDecoder(res.Body).Decode(&report); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range report.Matches {
		if m.Func == "validateAdmin" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected validateAdmin as a sibling of validateUser, got %+v", report.Matches)
	}
}

func TestLikeRequiresRef(t *testing.T) {
	srv := httptest.NewServer(testHandler(""))
	defer srv.Close()
	res, _ := http.Get(srv.URL + "/api/v1/like")
	_ = res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing ref = %d, want 400", res.StatusCode)
	}
}

func TestClustersEndpoint(t *testing.T) {
	srv := httptest.NewServer(testHandler(""))
	defer srv.Close()
	res, err := http.Get(srv.URL + "/api/v1/clusters?min=0.9")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	var report scan.ClusterReport
	if err := json.NewDecoder(res.Body).Decode(&report); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, c := range report.Clusters {
		for _, m := range c.Members {
			if m.Func == "validateAdmin" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected the validate* pair in a cluster, got %+v", report.Clusters)
	}
}
