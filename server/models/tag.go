package models

// Tag is a normalized tag plus usage count.
type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
