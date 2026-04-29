package models

// Video is the API representation of one stored video record.
type Video struct {
	ID              int64    `json:"id"`
	Title           string   `json:"title"`
	Description     *string  `json:"description"`
	Filename        string   `json:"filename"`
	OriginalName    string   `json:"original_name"`
	FileSize        int64    `json:"file_size"`
	Duration        *int64   `json:"duration"`
	Thumbnail       *string  `json:"thumbnail"`
	Platform        string   `json:"platform"`
	Status          string   `json:"status"`
	ScheduledAt     *string  `json:"scheduled_at"`
	PostedAt        *string  `json:"posted_at"`
	N8NWorkflowID   *string  `json:"n8n_workflow_id"`
	N8NExecutionLog string   `json:"n8n_execution_log"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	Tags            []string `json:"tags"`
}
