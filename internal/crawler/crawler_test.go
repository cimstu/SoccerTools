package crawler

import (
	"fmt"
	"testing"
)

func TestParseFileBarcelona(t *testing.T) {
	path := "test.html" // 相对 internal/crawler 目录
	items, err := ParseFile(path)
	if err != nil {
		t.Fatalf("ParseFile: %v", err)
	}
	fmt.Printf("从 test.html 解析到 %d 场巴萨相关比赛：\n", len(items))
	for i, it := range items {
		fmt.Printf("%d. [%s] %s\n   %s\n", i+1, it.Date, it.Title, it.URL)
	}
}
