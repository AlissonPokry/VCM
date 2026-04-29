package models

// N8NExecutionEntry is one JSON entry stored in videos.n8n_execution_log.
type N8NExecutionEntry struct {
	ExecutionID *string `json:"execution_id"`
	Timestamp   string  `json:"timestamp"`
	Result      string  `json:"result"`
	Platform    *string `json:"platform"`
	Error       *string `json:"error"`
}
