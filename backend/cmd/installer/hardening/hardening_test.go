package hardening

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"

	"pentagi/cmd/installer/checker"
	"pentagi/cmd/installer/loader"

	"github.com/vxcontrol/cloud/system"
)

const hardeningUUID = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`

// hardeningEnvExample loads the .env.example shipped at the repository root.
func hardeningEnvExample(t *testing.T) map[string]loader.EnvVar {
	t.Helper()
	envFile, err := loader.LoadEnvFile("../../../../.env.example")
	if err != nil {
		t.Fatalf("load .env.example: %v", err)
	}
	return envFile.GetAll()
}

func TestHardening_RandomString_GeneratesWhatThePolicyAsks(t *testing.T) {
	tests := []struct {
		name    string
		policy  HardeningPolicy
		want    string // a regular expression for the whole value
		random  bool   // a second call returns another value
		wantErr string
	}{
		{"empty alphanumeric", HardeningPolicy{Type: HardeningPolicyTypeDefault}, `^$`, false, ""},
		{"alphanumeric", HardeningPolicy{Type: HardeningPolicyTypeDefault, Length: 10}, `^[0-9a-zA-Z]{10}$`, true, ""},
		{"empty hex", HardeningPolicy{Type: HardeningPolicyTypeHex}, `^$`, false, ""},
		{"even hex", HardeningPolicy{Type: HardeningPolicyTypeHex, Length: 16}, `^[0-9a-f]{16}$`, true, ""},
		{"odd hex", HardeningPolicy{Type: HardeningPolicyTypeHex, Length: 7}, `^[0-9a-f]{7}$`, true, ""},
		{"uuid without a prefix", HardeningPolicy{Type: HardeningPolicyTypeUUID}, `^` + hardeningUUID, true, ""},
		{"uuid with a prefix", HardeningPolicy{Type: HardeningPolicyTypeUUID, Prefix: "pk-lf-"}, `^pk-lf-` + hardeningUUID, true, ""},
		{"bool true", HardeningPolicy{Type: HardeningPolicyTypeBoolTrue}, `^true$`, false, ""},
		{"bool false", HardeningPolicy{Type: HardeningPolicyTypeBoolFalse}, `^false$`, false, ""},
		{"an unknown type", HardeningPolicy{Type: "invalid"}, "", false, "invalid hardening policy type: invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := randomString(tt.policy)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("randomString: %v", err)
			}
			if !regexp.MustCompile(tt.want).MatchString(got) {
				t.Errorf("randomString = %q, want a match for %s", got, tt.want)
			}
			if again, err := randomString(tt.policy); tt.random && (err != nil || again == got) {
				t.Errorf("a second randomString = %q (err %v), want a value other than %q", again, err, got)
			}
		})
	}
}

func TestHardening_UpdateDefaultValues_FillsOnlyAnEmptyDefault(t *testing.T) {
	vars := map[string]loader.EnvVar{
		"COOKIE_SIGNING_SALT":        {Name: "COOKIE_SIGNING_SALT"},
		"LANGFUSE_POSTGRES_PASSWORD": {Name: "LANGFUSE_POSTGRES_PASSWORD", Default: "custom"},
		"UNKNOWN_VAR":                {Name: "UNKNOWN_VAR"},
	}

	updateDefaultValues(vars)

	got := map[string]string{}
	for name, envVar := range vars {
		got[name] = envVar.Default
	}
	want := map[string]string{"COOKIE_SIGNING_SALT": "salt", "LANGFUSE_POSTGRES_PASSWORD": "custom", "UNKNOWN_VAR": ""}
	if !maps.Equal(got, want) {
		t.Errorf("defaults = %v, want %v", got, want)
	}
}

func TestHardening_ReplaceDefaultValues_ReplacesOnlyADefaultValue(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		policy  HardeningPolicyType
		want    string // a regular expression for the resulting value
		changed bool
		wantErr string
	}{
		{"a default value is regenerated", "default", HardeningPolicyTypeDefault, `^[0-9a-zA-Z]{10}$`, true, ""},
		{"a custom value is kept", "custom", HardeningPolicyTypeDefault, `^custom$`, false, ""},
		{"an unknown policy type fails", "default", "invalid", "", false, "failed to generate random string for TEST_VAR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &mockState{}
			vars := map[string]loader.EnvVar{"TEST_VAR": {Name: "TEST_VAR", Value: tt.value, Default: "default"}}

			changed, err := replaceDefaultValues(st, vars, map[string]HardeningPolicy{"TEST_VAR": {Type: tt.policy, Length: 10}})

			if changed != tt.changed {
				t.Errorf("changed = %t, want %t", changed, tt.changed)
			}
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("replaceDefaultValues: %v", err)
			}
			got := vars["TEST_VAR"]
			if !regexp.MustCompile(tt.want).MatchString(got.Value) || got.IsChanged != tt.changed {
				t.Errorf("TEST_VAR = %q (changed %t), want a match for %s", got.Value, got.IsChanged, tt.want)
			}
			if staged, ok := st.vars["TEST_VAR"]; ok != tt.changed || (ok && staged.Value != got.Value) {
				t.Errorf("state holds %q for TEST_VAR (staged %t), want the new value staged only when replaced", staged.Value, ok)
			}
		})
	}
}

func TestHardening_SyncValueToState_KeepsThePreviousValueAsDefault(t *testing.T) {
	for _, newValue := range []string{"new_value", ""} {
		st := &mockState{}

		got, err := syncValueToState(st, loader.EnvVar{Name: "TEST_VAR", Value: "old_value"}, newValue)
		if err != nil {
			t.Fatalf("syncValueToState(%q): %v", newValue, err)
		}

		if staged := st.vars["TEST_VAR"]; staged.Value != newValue || !staged.IsChanged {
			t.Errorf("state holds %q (changed %t), want %q staged", staged.Value, staged.IsChanged, newValue)
		}
		if got.Value != newValue || got.Default != "old_value" {
			t.Errorf("returned %q with default %q, want %q with default %q", got.Value, got.Default, newValue, "old_value")
		}
	}
}

// A .env.example value that differs from its hardening default is never regenerated on a fresh install.
func TestHardening_VarsForHardening_ShipInEnvExampleWithTheirDefaults(t *testing.T) {
	example := hardeningEnvExample(t)

	for area, names := range varsForHardening {
		for _, name := range names {
			def, hasDefault := varsForHardeningDefault[name]
			envVar, loaded := example[name]
			switch {
			case !hasDefault:
				t.Errorf("%s (%s) has no hardening default", name, area)
			case !loaded || !envVar.IsPresent():
				t.Errorf("%s (%s) is missing from .env.example", name, area)
			case envVar.Value != def:
				t.Errorf(".env.example ships %s=%q, hardening treats %q as its default", name, envVar.Value, def)
			}
		}
	}
}

func TestHardening_DoHardening_RegeneratesOnlyTheSecretsOfAFreshStack(t *testing.T) {
	secrets := []struct{ area, name, shape string }{
		{"pentagi", "COOKIE_SIGNING_SALT", `^[0-9a-f]{32}$`},
		{"pentagi", "PENTAGI_POSTGRES_PASSWORD", `^[0-9a-zA-Z]{18}$`},
		{"pentagi", "LOCAL_SCRAPER_USERNAME", `^[0-9a-zA-Z]{10}$`},
		{"pentagi", "LOCAL_SCRAPER_PASSWORD", `^[0-9a-zA-Z]{12}$`},
		{"graphiti", "NEO4J_PASSWORD", `^[0-9a-zA-Z]{18}$`},
		{"langfuse", "LANGFUSE_POSTGRES_PASSWORD", `^[0-9a-zA-Z]{18}$`},
		{"langfuse", "LANGFUSE_CLICKHOUSE_PASSWORD", `^[0-9a-zA-Z]{18}$`},
		{"langfuse", "LANGFUSE_S3_ACCESS_KEY_ID", `^[0-9a-zA-Z]{20}$`},
		{"langfuse", "LANGFUSE_S3_SECRET_ACCESS_KEY", `^[0-9a-zA-Z]{40}$`},
		{"langfuse", "LANGFUSE_REDIS_AUTH", `^[0-9a-f]{48}$`},
		{"langfuse", "LANGFUSE_SALT", `^[0-9a-f]{28}$`},
		{"langfuse", "LANGFUSE_ENCRYPTION_KEY", `^[0-9a-f]{64}$`},
		{"langfuse", "LANGFUSE_NEXTAUTH_SECRET", `^[0-9a-f]{32}$`},
		{"langfuse", "LANGFUSE_INIT_PROJECT_PUBLIC_KEY", `^pk-lf-` + hardeningUUID},
		{"langfuse", "LANGFUSE_INIT_PROJECT_SECRET_KEY", `^sk-lf-` + hardeningUUID},
		{"langfuse", "LANGFUSE_AUTH_DISABLE_SIGNUP", `^true$`},
	}
	langfuseSync := map[string]string{
		"LANGFUSE_PROJECT_ID": "LANGFUSE_INIT_PROJECT_ID",
		"LANGFUSE_PUBLIC_KEY": "LANGFUSE_INIT_PROJECT_PUBLIC_KEY",
		"LANGFUSE_SECRET_KEY": "LANGFUSE_INIT_PROJECT_SECRET_KEY",
	}
	installationID := system.GetInstallationID().String()
	example := hardeningEnvExample(t)

	tests := []struct {
		name     string
		check    checker.CheckResult
		custom   map[string]string // values the user changed away from .env.example
		hardened string            // the areas whose secrets are regenerated
	}{
		{"a fresh langfuse is regenerated", checker.CheckResult{PentagiInstalled: true, GraphitiInstalled: true}, nil, "langfuse"},
		{"a fresh pentagi is regenerated", checker.CheckResult{GraphitiInstalled: true, LangfuseInstalled: true}, nil, "pentagi"},
		{"a fresh graphiti is regenerated", checker.CheckResult{PentagiInstalled: true, LangfuseInstalled: true}, nil, "graphiti"},
		{"a fresh installation regenerates every stack", checker.CheckResult{}, nil, "pentagi graphiti langfuse"},
		{"installed stacks keep their secrets", checker.CheckResult{
			PentagiInstalled: true, GraphitiInstalled: true, LangfuseInstalled: true,
			PentagiVolumesExist: true, GraphitiVolumesExist: true, LangfuseVolumesExist: true,
		}, nil, ""},
		{"langfuse volumes left behind keep its secrets", checker.CheckResult{
			PentagiInstalled: true, GraphitiInstalled: true, LangfuseVolumesExist: true,
		}, nil, ""},
		{"pentagi volumes left behind keep its secrets", checker.CheckResult{
			GraphitiInstalled: true, LangfuseInstalled: true, PentagiVolumesExist: true,
		}, nil, ""},
		{"graphiti volumes left behind keep its secrets", checker.CheckResult{
			PentagiInstalled: true, LangfuseInstalled: true, GraphitiVolumesExist: true,
		}, nil, ""},
		{"volumes left behind by every stack keep every secret", checker.CheckResult{
			PentagiVolumesExist: true, GraphitiVolumesExist: true, LangfuseVolumesExist: true,
		}, nil, ""},
		{"a first run records the installation id", checker.CheckResult{
			PentagiInstalled: true, GraphitiInstalled: true, LangfuseInstalled: true,
		}, map[string]string{"INSTALLATION_ID": ""}, ""},
		{"a custom langfuse secret is kept", checker.CheckResult{PentagiInstalled: true, GraphitiInstalled: true},
			map[string]string{"LANGFUSE_SALT": "custom"}, "langfuse"},
		{"custom scraper credentials and url are kept", checker.CheckResult{GraphitiInstalled: true, LangfuseInstalled: true},
			map[string]string{
				"LOCAL_SCRAPER_USERNAME": "customuser", "LOCAL_SCRAPER_PASSWORD": "custompass",
				"SCRAPER_PRIVATE_URL": "https://customuser:custompass@scraper/",
			}, "pentagi"},
		{"a custom scraper url is kept while its default password is regenerated",
			checker.CheckResult{GraphitiInstalled: true, LangfuseInstalled: true},
			map[string]string{
				"LOCAL_SCRAPER_USERNAME": "customuser", "SCRAPER_PRIVATE_URL": "https://customuser:somepass@scraper/",
			}, "pentagi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := &mockState{vars: maps.Clone(example)}
			st.vars["INSTALLATION_ID"] = loader.EnvVar{Name: "INSTALLATION_ID", Value: installationID}
			for name, value := range tt.custom {
				envVar := st.vars[name]
				envVar.Value = value
				st.vars[name] = envVar
			}
			before := maps.Clone(st.vars)
			hardened := map[string]bool{}
			for _, area := range strings.Fields(tt.hardened) {
				hardened[area] = true
			}

			if err := DoHardening(st, tt.check); err != nil {
				t.Fatalf("DoHardening: %v", err)
			}

			regenerated := map[string]bool{"INSTALLATION_ID": true}
			if got := st.vars["INSTALLATION_ID"].Value; got != installationID {
				t.Errorf("INSTALLATION_ID = %q, want %q", got, installationID)
			}
			for _, secret := range secrets {
				if _, custom := tt.custom[secret.name]; custom || !hardened[secret.area] {
					continue
				}
				regenerated[secret.name] = true
				got := st.vars[secret.name]
				if !regexp.MustCompile(secret.shape).MatchString(got.Value) || !got.IsChanged || got.Value == before[secret.name].Value {
					t.Errorf("%s = %q (changed %t, shipped %q), want a fresh value matching %s",
						secret.name, got.Value, got.IsChanged, before[secret.name].Value, secret.shape)
				}
			}
			if hardened["langfuse"] {
				for target, source := range langfuseSync {
					regenerated[target] = true
					if got, want := st.vars[target].Value, st.vars[source].Value; got != want {
						t.Errorf("%s = %q, want %s's %q", target, got, source, want)
					}
				}
			}
			if _, custom := tt.custom["SCRAPER_PRIVATE_URL"]; hardened["pentagi"] && !custom {
				regenerated["SCRAPER_PRIVATE_URL"] = true
				want := "https://" + st.vars["LOCAL_SCRAPER_USERNAME"].Value + ":" +
					st.vars["LOCAL_SCRAPER_PASSWORD"].Value + "@scraper/"
				if got := st.vars["SCRAPER_PRIVATE_URL"].Value; got != want {
					t.Errorf("SCRAPER_PRIVATE_URL = %q, want %q", got, want)
				}
			}
			for name, got := range st.vars {
				if !regenerated[name] && got.Value != before[name].Value {
					t.Errorf("%s changed from %q to %q", name, before[name].Value, got.Value)
				}
			}
			wantCommits := 0
			if _, firstRun := tt.custom["INSTALLATION_ID"]; firstRun || tt.hardened != "" {
				wantCommits = 1
			}
			if st.commits != wantCommits {
				t.Errorf("committed %d times, want %d", st.commits, wantCommits)
			}
		})
	}
}

func TestHardening_SyncLangfuseState_CopiesTheInitKeysIntoEmptyVars(t *testing.T) {
	const (
		projectID = "cm47619l0000872mcd2dlbqwb"
		publicKey = "pk-lf-12345678-1234-1234-1234-123456789abc"
		secretKey = "sk-lf-87654321-4321-4321-4321-cba987654321"
	)
	tests := []struct {
		name    string
		vars    map[string]string
		want    map[string]string
		synced  []string
		changed bool
	}{
		{
			name: "empty vars take the init keys",
			vars: map[string]string{
				"LANGFUSE_PROJECT_ID": "", "LANGFUSE_PUBLIC_KEY": "", "LANGFUSE_SECRET_KEY": "",
				"LANGFUSE_INIT_PROJECT_ID": projectID, "LANGFUSE_INIT_PROJECT_PUBLIC_KEY": publicKey,
				"LANGFUSE_INIT_PROJECT_SECRET_KEY": secretKey,
			},
			want: map[string]string{
				"LANGFUSE_PROJECT_ID": projectID, "LANGFUSE_PUBLIC_KEY": publicKey, "LANGFUSE_SECRET_KEY": secretKey,
				"LANGFUSE_INIT_PROJECT_ID": projectID, "LANGFUSE_INIT_PROJECT_PUBLIC_KEY": publicKey,
				"LANGFUSE_INIT_PROJECT_SECRET_KEY": secretKey,
			},
			synced:  []string{"LANGFUSE_PROJECT_ID", "LANGFUSE_PUBLIC_KEY", "LANGFUSE_SECRET_KEY"},
			changed: true,
		},
		{
			name: "a value the user set is kept",
			vars: map[string]string{
				"LANGFUSE_PROJECT_ID": "existing-project-id", "LANGFUSE_PUBLIC_KEY": "",
				"LANGFUSE_INIT_PROJECT_ID": projectID, "LANGFUSE_INIT_PROJECT_PUBLIC_KEY": publicKey,
			},
			want: map[string]string{
				"LANGFUSE_PROJECT_ID": "existing-project-id", "LANGFUSE_PUBLIC_KEY": publicKey,
				"LANGFUSE_INIT_PROJECT_ID": projectID, "LANGFUSE_INIT_PROJECT_PUBLIC_KEY": publicKey,
			},
			synced:  []string{"LANGFUSE_PUBLIC_KEY"},
			changed: true,
		},
		{
			name: "no init key leaves the var empty",
			vars: map[string]string{"LANGFUSE_PROJECT_ID": ""},
			want: map[string]string{"LANGFUSE_PROJECT_ID": ""},
		},
		{
			name: "an absent var is not created",
			vars: map[string]string{"LANGFUSE_INIT_PROJECT_ID": projectID},
			want: map[string]string{"LANGFUSE_INIT_PROJECT_ID": projectID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := map[string]loader.EnvVar{}
			for name, value := range tt.vars {
				vars[name] = loader.EnvVar{Name: name, Value: value}
			}

			changed, err := syncLangfuseState(&mockState{}, vars)
			if err != nil {
				t.Fatalf("syncLangfuseState: %v", err)
			}

			if changed != tt.changed {
				t.Errorf("changed = %t, want %t", changed, tt.changed)
			}
			got := map[string]string{}
			for name, envVar := range vars {
				got[name] = envVar.Value
				if synced := slices.Contains(tt.synced, name); envVar.IsChanged != synced {
					t.Errorf("%s changed = %t, want %t", name, envVar.IsChanged, synced)
				}
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("vars = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHardening_SyncScraperState_RebuildsADefaultURLFromChangedCredentials(t *testing.T) {
	const defaultURL = "https://someuser:somepass@scraper/"
	tests := []struct {
		name                     string
		user, pass, url          string // empty means the variable is absent
		userChanged, passChanged bool
		urlDefault               string // defaultURL when empty
		wantURL                  string
		changed                  bool
		wantErr                  string
	}{
		{"hardened credentials rebuild a default url", "newuser", "newpass", defaultURL, true, true, "",
			"https://newuser:newpass@scraper/", true, ""},
		{"a custom url is kept", "newuser", "newpass", "https://customuser:custompass@scraper/", true, true, "",
			"https://customuser:custompass@scraper/", false, ""},
		{"unchanged credentials keep the url", "someuser", "somepass", defaultURL, false, false, "",
			defaultURL, false, ""},
		{"one changed credential is enough", "newuser", "somepass", defaultURL, true, false, "",
			"https://newuser:somepass@scraper/", true, ""},
		{"missing variables change nothing", "newuser", "", "", true, false, "", "", false, ""},
		{"an unparsable url fails", "newuser", "newpass", "://invalid-url", true, true, "://invalid-url",
			"://invalid-url", false, "failed to parse scraper private URL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := map[string]loader.EnvVar{
				"LOCAL_SCRAPER_USERNAME": {Name: "LOCAL_SCRAPER_USERNAME", Value: tt.user, Default: "someuser", IsChanged: tt.userChanged},
			}
			if tt.pass != "" {
				vars["LOCAL_SCRAPER_PASSWORD"] = loader.EnvVar{Name: "LOCAL_SCRAPER_PASSWORD", Value: tt.pass, Default: "somepass", IsChanged: tt.passChanged}
			}
			if tt.url != "" {
				def := tt.urlDefault
				if def == "" {
					def = defaultURL
				}
				vars["SCRAPER_PRIVATE_URL"] = loader.EnvVar{Name: "SCRAPER_PRIVATE_URL", Value: tt.url, Default: def}
			}

			changed, err := syncScraperState(&mockState{}, vars)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("err = %v, want %q", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("syncScraperState: %v", err)
			}
			if changed != tt.changed {
				t.Errorf("changed = %t, want %t", changed, tt.changed)
			}
			got, exists := vars["SCRAPER_PRIVATE_URL"]
			if got.Value != tt.wantURL || exists != (tt.wantURL != "") || got.IsChanged != tt.changed {
				t.Errorf("SCRAPER_PRIVATE_URL = %q (present %t, changed %t), want %q", got.Value, exists, got.IsChanged, tt.wantURL)
			}
		})
	}
}
