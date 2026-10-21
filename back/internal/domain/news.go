package domain

import "time"

type News struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"publishedAt"`
	Summary     string    `json:"summary"`
	URL         string    `json:"url"`
	Category    string    `json:"category"`
}
