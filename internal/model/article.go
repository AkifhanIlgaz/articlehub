package model

type Article struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Content     string   `json:"content"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	ReadingTime int      `json:"reading_time"`
	ViewCount   int      `json:"view_count"`
	LikeCount   int      `json:"like_count"`
	PublishedAt string   `json:"published_at"`
}
