package model

// ReplayItem 单条比赛录像
type ReplayItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Date  string `json:"date"` // YYYY-MM-DD
}
