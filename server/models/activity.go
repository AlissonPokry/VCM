package models

// ActivityEntry describes an audit-log event exposed through the API.
type ActivityEntry struct {
	ID         int64   `json:"id"`
	VideoID    *int64  `json:"video_id"`
	Action     string  `json:"action"`
	Detail     *string `json:"detail"`
	Source     string  `json:"source"`
	CreatedAt  string  `json:"created_at"`
	VideoTitle *string `json:"video_title"`
}
