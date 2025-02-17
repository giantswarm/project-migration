package types

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestProjectListJSON(t *testing.T) {
	// Create a ProjectList and marshal/unmarshal
	orig := ProjectList{
		Projects: []Project{
			{Number: 301, ID: "proj-301"},
			{Number: 302, ID: "proj-302"},
		},
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshalling ProjectList failed: %v", err)
	}
	var decoded ProjectList
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshalling ProjectList failed: %v", err)
	}
	if !reflect.DeepEqual(orig, decoded) {
		t.Errorf("Expected %+v, got %+v", orig, decoded)
	}
}

func TestFieldsJSON(t *testing.T) {
	// Create a Fields instance and marshal/unmarshal it.
	orig := Fields{
		Fields: []Field{
			{
				ID:   "1",
				Name: "Status",
				Options: []Option{
					{ID: "opt1", Name: "Open"},
					{ID: "opt2", Name: "Closed"},
				},
			},
			{
				ID:   "2",
				Name: "Priority",
				Options: []Option{
					{ID: "opt3", Name: "High"},
					{ID: "opt4", Name: "Low"},
				},
			},
		},
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshalling Fields failed: %v", err)
	}
	var decoded Fields
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshalling Fields failed: %v", err)
	}
	if !reflect.DeepEqual(orig, decoded) {
		t.Errorf("Expected %+v, got %+v", orig, decoded)
	}
}

func TestItemListJSON(t *testing.T) {
	// Create a sample ItemList, marshal and unmarshal it.
	orig := ItemList{
		Items: []Item{
			{
				ID:    "item1",
				Title: "First Issue",
				Content: Content{
					URL:   "http://example.com/issue1",
					Type:  "Issue",
					Title: "Issue 1",
				},
				Status:     "Open",
				Kind:       "Bug",
				Workstream: "WS1",
				StartDate:  "2023-01-01",
				TargetDate: "2023-01-31",
			},
		},
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshalling ItemList failed: %v", err)
	}
	var decoded ItemList
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshalling ItemList failed: %v", err)
	}
	if !reflect.DeepEqual(orig, decoded) {
		t.Errorf("Expected %+v, got %+v", orig, decoded)
	}
}
