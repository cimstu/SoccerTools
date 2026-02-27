package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/soccertools/soccertools/internal/model"
	"golang.org/x/net/html"
)

const zhibo8URL = "https://www.zhibo8.com/zuqiu/luxiang.htm"
const zhibo8Base = "https://www.zhibo8.com"

// keywords 用于过滤巴萨相关
var barcelonaKeywords = []string{"巴塞罗那", "巴萨"}

// ParseFile 从本地 HTML 文件解析，只保留含巴塞罗那/巴萨的条目（全量，不按天数截断）
func ParseFile(path string) ([]model.ReplayItem, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	return parseHTML(f)
}

// FetchAndParse 请求 zhibo8 录像页并解析出巴塞罗那相关录像（全量，由 API 的 days 参数控制返回范围）
func FetchAndParse() ([]model.ReplayItem, error) {
	resp, err := http.Get(zhibo8URL)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return parseHTML(resp.Body)
}

// filterLastNDays 只保留最近 n 天内的条目（含今天，共 n 天）
func filterLastNDays(items []model.ReplayItem, n int) []model.ReplayItem {
	if n <= 0 {
		return items
	}
	cutoff := time.Now().AddDate(0, 0, -n+1) // 今天算第 1 天
	cutoff = time.Date(cutoff.Year(), cutoff.Month(), cutoff.Day(), 0, 0, 0, 0, cutoff.Location())
	var out []model.ReplayItem
	for _, it := range items {
		t, err := time.Parse("2006-01-02", it.Date)
		if err != nil {
			continue
		}
		if !t.Before(cutoff) {
			out = append(out, it)
		}
	}
	return out
}

// urlDateRe 从链接路径解析日期，如 /zuqiu/2026/0226-match... 或 /zuqiu/2020/1230-match...
var urlDateRe = regexp.MustCompile(`/zuqiu/(\d{4})/(\d{4})`)

// dateFromURL 从 href 解析 YYYY-MM-DD，失败返回空字符串
func dateFromURL(href string) string {
	ms := urlDateRe.FindStringSubmatch(href)
	if len(ms) < 3 {
		return ""
	}
	year, mmdd := ms[1], ms[2]
	if len(mmdd) != 4 {
		return ""
	}
	return year + "-" + mmdd[:2] + "-" + mmdd[2:]
}

// normalizeTitle 去掉前导 "| " 和首尾空格
func normalizeTitle(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "|")
	return strings.TrimSpace(s)
}

// parseHTML 从 HTML 中解析日期块和链接，只保留含巴塞罗那/巴萨的条目
// 页面结构：每天一个 .box，.titlebar h2 为日期（如 "2月26日 星期四"），.content 内多个 <b>，每条为 "队名vs队名 <a>全场录像</a>"
func parseHTML(r io.Reader) ([]model.ReplayItem, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	dateLineRe := regexp.MustCompile(`(\d{1,2})月(\d{1,2})日`)
	var items []model.ReplayItem

	doc.Find(".box").Each(func(_ int, box *goquery.Selection) {
		// 从 .titlebar h2 取日期，如 "2月26日 星期四"（仅作 fallback，优先用 URL 里的日期）
		dateText := strings.TrimSpace(box.Find(".titlebar h2").First().Text())
		ms := dateLineRe.FindStringSubmatch(dateText)
		if len(ms) < 3 {
			return
		}
		month, day := ms[1], ms[2]
		year := time.Now().Year()
		boxDate := fmt.Sprintf("%d-%s-%s", year, zeroPad(month), zeroPad(day))

		// .content 内每条比赛：有的在 <b> 里（标题 <a>全场录像</a>），有的直接在 span 里（同上）
		// 1) <b> 包裹的
		box.Find(".content b").Each(func(_ int, b *goquery.Selection) {
			a := b.Find("a[href]").First()
			href, ok := a.Attr("href")
			if !ok || href == "" {
				return
			}
			title := normalizeTitle(b.Text())
			title = strings.TrimSuffix(title, "全场录像")
			title = normalizeTitle(title)
			if title == "" || !containsBarcelona(title) {
				return
			}
			itemDate := dateFromURL(href)
			if itemDate == "" {
				itemDate = boxDate
			}
			items = append(items, model.ReplayItem{
				Title: title,
				URL:   resolveURL(href),
				Date:  itemDate,
			})
		})
		// 2) span 内直接 "标题 <a>全场录像</a>"，无 <b>
		box.Find(".content a[href]").Each(func(_ int, a *goquery.Selection) {
			if strings.TrimSpace(a.Text()) != "全场录像" {
				return
			}
			if a.Parent().Is("b") {
				return // 已在上面 b 逻辑中处理
			}
			href, _ := a.Attr("href")
			if href == "" {
				return
			}
			title := normalizeTitle(titleFromPrevSibling(a))
			title = strings.TrimSuffix(title, "全场录像")
			title = normalizeTitle(title)
			if title == "" || !containsBarcelona(title) {
				return
			}
			itemDate := dateFromURL(href)
			if itemDate == "" {
				itemDate = boxDate
			}
			items = append(items, model.ReplayItem{
				Title: title,
				URL:   resolveURL(href),
				Date:  itemDate,
			})
		})
	})
	return items, nil
}

// titleFromPrevSibling 从 <a> 的前一个兄弟文本节点取标题（span 内 "标题 <a>全场录像</a>"）
func titleFromPrevSibling(a *goquery.Selection) string {
	if a.Length() == 0 {
		return ""
	}
	node := a.Get(0)
	if node.PrevSibling == nil || node.PrevSibling.Type != html.TextNode {
		return ""
	}
	return node.PrevSibling.Data
}

func resolveURL(href string) string {
	if href == "" {
		return href
	}
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}
	base, _ := url.Parse(zhibo8Base)
	ref, err := url.Parse(href)
	if err != nil {
		return href
	}
	return base.ResolveReference(ref).String()
}

func zeroPad(s string) string {
	if len(s) == 1 {
		return "0" + s
	}
	return s
}

func containsBarcelona(s string) bool {
	for _, k := range barcelonaKeywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}
