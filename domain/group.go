package domain

// Group - Contains the detail for each group
type Group struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Location   string                 `json:"location"`
	Tags       map[string]interface{} `json:"tags"`
	Properties map[string]interface{} `json:"properties"`
}
