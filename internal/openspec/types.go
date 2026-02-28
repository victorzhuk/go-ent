package openspec

// ListItem represents a change or spec in list output.
type ListItem struct {
	Name           string `json:"name"`
	CompletedTasks int    `json:"completedTasks,omitempty"`
	TotalTasks     int    `json:"totalTasks,omitempty"`
	LastModified   string `json:"lastModified,omitempty"`
	Status         string `json:"status,omitempty"`
	Description    string `json:"description,omitempty"`
}

// ListResponse wraps the list output from openspec CLI.
type ListResponse struct {
	Changes []ListItem `json:"changes,omitempty"`
	Specs   []ListItem `json:"specs,omitempty"`
}
