package main

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/soccertools/soccertools/internal/model"
	"github.com/soccertools/soccertools/internal/store"
)

// 契约测试（Contract Test）：以 specs/api/openapi.yaml 与 specs/behavior/example-match-list.feature 为规约。
//
// 与 behavior 场景对应：
//   - 健康检查           → TestHealth_Contract, TestHealth_MethodNotAllowed
//   - 查询最近 7/14 天   → TestReplays_Contract, TestReplays_DaysParameter
//   - 参数 days 400      → TestReplays_400_Contract
//   - 响应不得包含比分   → TestReplays_NoScore
//   - 立即刷新录像数据   → TestReplaysRefresh_Contract, TestReplaysRefresh_MethodNotAllowed
//   - Web 页面可查询与刷新 → TestReplaysPage_Contract

func TestHealth_Contract(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	methodGET(health)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /health status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("status = %q, want ok", body.Status)
	}
}

func TestHealth_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()
	methodGET(health)(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /health status = %d, want 405", rr.Code)
	}
}

func TestReplays_Contract(t *testing.T) {
	s := store.New()
	s.Add([]model.ReplayItem{
		{Title: "巴塞罗那vs莱万特", URL: "https://example.com/1", Date: "2026-02-23"},
	})

	req := httptest.NewRequest(http.MethodGet, "/replays", nil)
	rr := httptest.NewRecorder()
	methodGET(replays(s))(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("GET /replays status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body struct {
		Items []model.ReplayItem `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Items == nil {
		t.Error("items must be present (array)")
	}
	dateRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`) // format: date (YYYY-MM-DD)
	for i, it := range body.Items {
		if it.Title == "" || it.URL == "" || it.Date == "" {
			t.Errorf("items[%d] missing required field: title=%q url=%q date=%q", i, it.Title, it.URL, it.Date)
		}
		if !dateRe.MatchString(it.Date) {
			t.Errorf("items[%d].date = %q, want YYYY-MM-DD", i, it.Date)
		}
	}
}

func TestReplays_MethodNotAllowed(t *testing.T) {
	s := store.New()
	req := httptest.NewRequest(http.MethodPost, "/replays", nil)
	rr := httptest.NewRecorder()
	methodGET(replays(s))(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /replays status = %d, want 405", rr.Code)
	}
}

// TestReplays_NoScore 规约：不透露比赛比分（specs/domain/replay.md）。响应仅允许 title/url/date。
func TestReplays_NoScore(t *testing.T) {
	s := store.New()
	s.Add([]model.ReplayItem{
		{Title: "巴塞罗那vs莱万特", URL: "https://example.com/1", Date: "2026-02-23"},
	})
	req := httptest.NewRequest(http.MethodGet, "/replays", nil)
	rr := httptest.NewRecorder()
	methodGET(replays(s))(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	var body struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	disallowed := []string{"score", "比分", "result", "goals", "进球"}
	for i, item := range body.Items {
		for _, key := range disallowed {
			if _, ok := item[key]; ok {
				t.Errorf("items[%d] 不得包含字段 %q（规约：不透露比分）", i, key)
			}
		}
		// 仅允许 title, url, date
		for k := range item {
			if k != "title" && k != "url" && k != "date" {
				t.Errorf("items[%d] 仅允许 title/url/date，发现 %q", i, k)
			}
		}
	}
}

func TestReplays_DaysParameter(t *testing.T) {
	s := store.New()

	tests := []struct {
		name   string
		query  string
		wantOK bool
	}{
		{"default", "", true},
		{"days=14", "?days=14", true},
		{"days=1", "?days=1", true},
		{"days=30", "?days=30", true},
		{"days=0 invalid", "?days=0", false},
		{"days=31 invalid", "?days=31", false},
		{"days=abc invalid", "?days=abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/replays"+tt.query, nil)
			rr := httptest.NewRecorder()
			methodGET(replays(s))(rr, req)
			if tt.wantOK && rr.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", rr.Code)
			}
			if !tt.wantOK && rr.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", rr.Code)
			}
		})
	}
}

// TestReplaysRefresh_Contract 对应 behavior：立即刷新录像数据
func TestReplaysRefresh_Contract(t *testing.T) {
	s := store.New()
	req := httptest.NewRequest(http.MethodPost, "/replays/refresh", nil)
	rr := httptest.NewRecorder()
	methodPOST(replaysRefresh(s))(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("POST /replays/refresh status = %d, want 200", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body struct {
		OK      bool   `json:"ok"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.OK {
		t.Error("response should have ok: true (规约：响应 JSON 包含 ok 且值为 true)")
	}
	// OpenAPI 允许返回 message，行为上刷新后应有提示
	if body.Message == "" {
		t.Log("response message is empty (optional in spec)")
	}
}

func TestReplaysRefresh_MethodNotAllowed(t *testing.T) {
	s := store.New()
	req := httptest.NewRequest(http.MethodGet, "/replays/refresh", nil)
	rr := httptest.NewRecorder()
	methodPOST(replaysRefresh(s))(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /replays/refresh status = %d, want 405", rr.Code)
	}
}

// TestReplaysPage_Contract 对应 behavior：Web 页面可查询与刷新
func TestReplaysPage_Contract(t *testing.T) {
	staticRoot, _ := fs.Sub(staticFS, "static")
	handler := http.FileServer(http.FS(staticRoot))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("GET / status = %d, want 200", rr.Code)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := rr.Body.String()
	// 规约：页面包含「巴萨」与「录像」文案
	if !strings.Contains(body, "巴萨") || !strings.Contains(body, "录像") {
		t.Error("page should contain 巴萨 and 录像")
	}
	// 规约：页面支持选择最近 7/14/30 天并查询
	if !strings.Contains(body, "7 天") || !strings.Contains(body, "14 天") || !strings.Contains(body, "30 天") {
		t.Error("page should offer 7/14/30 days options")
	}
	// 规约：页面支持点击「刷新数据」触发 POST /replays/refresh
	if !strings.Contains(body, "刷新数据") {
		t.Error("page should have 刷新数据 button")
	}
	// 确保有调用 /replays 与 /replays/refresh 的端到端能力（页面内脚本）
	if !strings.Contains(body, "/replays") {
		t.Error("page should reference /replays API")
	}
}

// TestReplays_400_Contract 规约：400 为参数错误，响应为 application/json（便于客户端解析）
func TestReplays_400_Contract(t *testing.T) {
	s := store.New()
	req := httptest.NewRequest(http.MethodGet, "/replays?days=0", nil)
	rr := httptest.NewRecorder()
	methodGET(replays(s))(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode 400 body: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Error("400 response should include error field")
	}
}
