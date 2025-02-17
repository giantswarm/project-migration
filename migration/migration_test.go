package migration

import (
	"testing"

	"project-migration/cli"
	"project-migration/types"
)

func dummyFields() *types.Fields {
	return &types.Fields{
		Fields: []types.Field{
			{ID: "1", Name: "Status", Options: []types.Option{{ID: "opt1", Name: "Open"}}},
			{ID: "2", Name: "Kind", Options: []types.Option{{ID: "opt2", Name: "Bug"}}},
			{ID: "3", Name: "Workstream", Options: []types.Option{{ID: "opt3", Name: "WS1"}}},
			{ID: "4", Name: "Team", Options: []types.Option{{ID: "opt4", Name: "TestTeam"}}},
			{ID: "5", Name: "SIG", Options: []types.Option{{ID: "opt5", Name: "TestSIG"}}},
			{ID: "6", Name: "Working Group", Options: []types.Option{{ID: "opt6", Name: "TestWG"}}},
			{ID: "7", Name: "Area", Options: []types.Option{{ID: "opt7", Name: "TestArea"}}},
			{ID: "8", Name: "Function", Options: []types.Option{{ID: "opt8", Name: "TestFunction"}}},
		},
	}
}

func TestValidateFieldsSuccess(t *testing.T) {
	opts := cli.Options{
		Project:  "301",
		Type:     "team",
		Name:     "TestTeam",
		Area:     "TestArea",
		Function: "TestFunction",
	}
	fields := dummyFields()
	if err := ValidateFields(fields, fields, opts); err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidateFieldsFailure(t *testing.T) {
	opts := cli.Options{
		Project: "301",
		Type:    "team",
		Name:    "NonExistent",
	}
	fields := dummyFields()
	if err := ValidateFields(fields, fields, opts); err == nil {
		t.Errorf("Expected error due to missing team option, got nil")
	}
}

// Since GetFields and MigrateItems rely on external commands, further tests could use dependency injection or mocks.
// For brevity, additional integration tests are omitted.
