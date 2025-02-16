package types

// Data types for JSON responses.
type Project struct {
	Number int    `json:"number"`
	ID     string `json:"id"`
}

type ProjectList struct {
	Projects []Project `json:"projects"`
}

type Option struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Field struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Options []Option `json:"options"`
}

// Renamed FieldResponse to Fields.
type Fields struct {
	Fields []Field `json:"fields"`
}

type Content struct {
	URL   string `json:"url"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

type Item struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Content    Content `json:"content"`
	Status     string  `json:"status"`
	Kind       string  `json:"kind"`
	Workstream string  `json:"workstream"`
	StartDate  string  `json:"start Date"`
	TargetDate string  `json:"target Date"`
}

type ItemList struct {
	Items []Item `json:"items"`
}
