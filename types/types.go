package types

// Project represents a project.
type Project struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
}

// Field represents a field in a project.
type Field struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	Options []FieldOption `json:"options"`
}

// FieldOption represents a selectable option for a field.
type FieldOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Item represents an item (issue) in a project.
type Item struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Status     interface{} `json:"status"`
	Kind       interface{} `json:"kind"`
	Workstream interface{} `json:"workstream"`
	StartDate  string      `json:"start Date"`
	TargetDate string      `json:"target Date"`
	Content    struct {
		Type  string `json:"type"`
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"content"`
}
