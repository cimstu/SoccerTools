package store

import (
	"sort"
	"sync"
	"time"

	"github.com/soccertools/soccertools/internal/model"
)

// Store 内存存储最近录像
type Store struct {
	mu    sync.RWMutex
	items []model.ReplayItem
}

func New() *Store {
	return &Store{}
}

func (s *Store) Add(items []model.ReplayItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 去重：同 URL 只保留最新
	seen := make(map[string]int)
	for i, it := range s.items {
		seen[it.URL] = i
	}
	for _, it := range items {
		if idx, ok := seen[it.URL]; ok {
			s.items[idx] = it
		} else {
			s.items = append(s.items, it)
		}
	}
	sort.Slice(s.items, func(i, j int) bool {
		return s.items[i].Date > s.items[j].Date
	})
}

// GetLastNDays 返回最近 n 天的录像（含今天，共 n 天）
func (s *Store) GetLastNDays(days int) []model.ReplayItem {
	if days <= 0 {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	cutoff := time.Now().AddDate(0, 0, -days+1) // 今天算第 1 天
	cutoffStr := time.Date(cutoff.Year(), cutoff.Month(), cutoff.Day(), 0, 0, 0, 0, cutoff.Location()).Format("2006-01-02")
	var out []model.ReplayItem
	for _, it := range s.items {
		if it.Date >= cutoffStr {
			out = append(out, it)
		}
	}
	return out
}
