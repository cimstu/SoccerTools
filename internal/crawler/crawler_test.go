package crawler

import (
	"fmt"
	"os"
	"testing"
)

func TestParseFileBarcelona(t *testing.T) {
	path := "test.html" // 相对 internal/crawler 目录，可选本地爬取样例
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skip("test.html not found (optional fixture for local parse test)")
	}
	items, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	fmt.Printf("从 test.html 解析到 %d 场巴萨相关比赛：\n", len(items))
	for i, it := range items {
		fmt.Printf("%d. [%s] %s\n   %s\n", i+1, it.Date, it.Title, it.URL)
	}
}
