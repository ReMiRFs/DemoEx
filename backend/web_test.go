package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProtectedRoutesRejectMissingSession(t *testing.T) {
	a := &app{allowedOrigins: map[string]bool{}}
	for _, endpoint := range []struct{ method, path string }{
		{"GET", "/api/auth/me"}, {"GET", "/api/requests"},
		{"POST", "/api/requests"}, {"PUT", "/api/requests/1"},
		{"DELETE", "/api/requests/1"}, {"POST", "/api/requests/1/review"},
		{"GET", "/api/admin/requests"}, {"PATCH", "/api/admin/requests/1/status"},
	} {
		t.Run(endpoint.method+endpoint.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			a.routes().ServeHTTP(response, httptest.NewRequest(endpoint.method, endpoint.path, nil))
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("получен %d: %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestOriginProtection(t *testing.T) {
	a := &app{allowedOrigins: map[string]bool{"http://localhost:4200": true}}
	cases := []struct {
		origin, site string
		status       int
	}{
		{"http://example.com", "same-origin", http.StatusNoContent},
		{"http://localhost:4200", "same-site", http.StatusNoContent},
		{"http://attacker.test", "cross-site", http.StatusForbidden},
		{"http://attacker.test", "", http.StatusForbidden},
		{"null", "", http.StatusForbidden},
		{"", "", http.StatusNoContent}, // CLI без Origin.
	}
	for _, tc := range cases {
		t.Run(tc.origin+tc.site, func(t *testing.T) {
			req := httptest.NewRequest("POST", "http://example.com/api/auth/logout", nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("Sec-Fetch-Site", tc.site)
			res := httptest.NewRecorder()
			a.protect(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })).ServeHTTP(res, req)
			if res.Code != tc.status {
				t.Fatalf("ожидался %d, получен %d", tc.status, res.Code)
			}
		})
	}
}

func TestSPAAndAPIBoundaries(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "index.html"), []byte("<h1>Angular</h1>"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.js"), []byte("console.log('loaded')"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("secret"), 0600); err != nil {
		t.Fatal(err)
	}
	a := &app{staticDir: root}
	for _, tc := range []struct {
		path     string
		status   int
		contains string
	}{
		{"/requests", 200, "Angular"},
		{"/main.js", 200, "loaded"},
		{"/missing.js", 404, ""},
		{"/.env", 404, ""},
		{"/api/missing", 404, "Маршрут API"},
		{"/api", 404, "Маршрут API"},
	} {
		t.Run(tc.path, func(t *testing.T) {
			res := httptest.NewRecorder()
			a.routes().ServeHTTP(res, httptest.NewRequest("GET", tc.path, nil))
			if res.Code != tc.status || !strings.Contains(res.Body.String(), tc.contains) {
				t.Fatalf("получен %d: %s", res.Code, res.Body.String())
			}
		})
	}
}

func TestReadJSONRejectsUnknownFieldsAndTrailingData(t *testing.T) {
	for _, body := range []string{
		`{"room":"Зал","userId":42}`,
		`{"room":"Зал"} {"room":"Другой"}`,
		`{"room":`,
	} {
		res := httptest.NewRecorder()
		var in requestInput
		if readJSON(res, httptest.NewRequest("POST", "/", strings.NewReader(body)), &in) || res.Code != 400 {
			t.Fatalf("не отклонён JSON: %s", body)
		}
	}
}
