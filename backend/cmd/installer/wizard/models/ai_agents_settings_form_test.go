package models

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/files"
	"pentagi/cmd/installer/state"
	"pentagi/cmd/installer/wizard/controller"
	"pentagi/cmd/installer/wizard/styles"
	"pentagi/cmd/installer/wizard/window"
)

func aiAgentsFormOver(t *testing.T, vars map[string]string) (*AIAgentsSettingsFormModel, controller.Controller) {
	t.Helper()
	example, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", ".env.example"))
	if err != nil {
		t.Fatal(err)
	}
	envPath := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(envPath, example, 0o600); err != nil {
		t.Fatal(err)
	}
	st, err := state.NewState(envPath)
	if err != nil {
		t.Fatal(err)
	}
	c := controller.NewController(st, files.NewFiles(), checker.CheckResult{})
	for name, value := range vars {
		if err := c.SetVar(name, value); err != nil {
			t.Fatal(err)
		}
	}
	form := NewAIAgentsSettingsFormModel(c, styles.New(), window.New())
	form.BuildForm()
	return form, c
}

func TestAIAgentsSettingsForm_HandleSave_OptionalIntegerFields(t *testing.T) {
	limits := []struct {
		field, variable string
		min, max        int
	}{
		{"execution_monitor_same_tool_limit", "EXECUTION_MONITOR_SAME_TOOL_LIMIT", 1, 50},
		{"execution_monitor_total_tool_limit", "EXECUTION_MONITOR_TOTAL_TOOL_LIMIT", 1, 100},
		{"max_general_agent_tool_calls", "MAX_GENERAL_AGENT_TOOL_CALLS", 10, 500},
		{"max_limited_agent_tool_calls", "MAX_LIMITED_AGENT_TOOL_CALLS", 5, 200},
	}

	t.Run("save unrelated change with blank defaults", func(t *testing.T) {
		form, c := aiAgentsFormOver(t, nil)
		for i, field := range form.GetFormFields() {
			if field.Key == "ask_user" {
				form.fields[i].Input.SetValue("false")
			}
		}
		if err := form.HandleSave(); err != nil {
			t.Fatal(err)
		}
		if saved, _ := c.GetVar("ASK_USER"); saved.Value != "false" {
			t.Fatalf("ASK_USER = %q, want false", saved.Value)
		}
		for _, limit := range limits {
			if saved, _ := c.GetVar(limit.variable); saved.Value != "" {
				t.Errorf("%s = %q, want blank to use the backend default", limit.variable, saved.Value)
			}
		}
	})

	for _, limit := range limits {
		t.Run(limit.field, func(t *testing.T) {
			for _, tc := range []struct {
				name, input, want string
				wantErr           bool
			}{
				{name: "clear override"},
				{name: "whitespace", input: "   "},
				{name: "minimum", input: strconv.Itoa(limit.min), want: strconv.Itoa(limit.min)},
				{name: "maximum", input: strconv.Itoa(limit.max), want: strconv.Itoa(limit.max)},
				{name: "trim and normalize", input: " 0" + strconv.Itoa(limit.min) + " ", want: strconv.Itoa(limit.min)},
				{name: "below minimum", input: strconv.Itoa(limit.min - 1), wantErr: true},
				{name: "above maximum", input: strconv.Itoa(limit.max + 1), wantErr: true},
				{name: "noninteger", input: "abc", wantErr: true},
				{name: "fraction", input: "1.5", wantErr: true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					// Give every numeric field an override so rejection tests reach the target field.
					vars := map[string]string{}
					for _, other := range limits {
						vars[other.variable] = strconv.Itoa(other.min)
					}
					form, c := aiAgentsFormOver(t, vars)
					found := false
					for i, field := range form.GetFormFields() {
						if field.Key == limit.field {
							form.fields[i].Input.SetValue(tc.input)
							found = true
						}
					}
					if !found {
						t.Fatalf("missing field %s", limit.field)
					}
					err := form.HandleSave()
					if (err != nil) != tc.wantErr {
						t.Fatalf("HandleSave() = %v, wantErr %v", err, tc.wantErr)
					}
					want := tc.want
					if tc.wantErr {
						want = vars[limit.variable]
					}
					if saved, _ := c.GetVar(limit.variable); saved.Value != want {
						t.Errorf("%s = %q, want %q", limit.variable, saved.Value, want)
					}
				})
			}
		})
	}
}
