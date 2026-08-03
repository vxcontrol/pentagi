package loader

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"pentagi/pkg/config"

	"github.com/caarlos0/env/v10"
)

// envVarNameRe matches valid environment variable identifiers (e.g. OPEN_AI_KEY).
// Used to distinguish a real commented-out assignment (e.g. "#SOME_VAR=value")
// from a plain human-readable comment that merely contains a stray '=' character
// (e.g. "<=" comparisons or "empty = disabled" phrasing in descriptive text).
var envVarNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func LoadEnvFile(path string) (EnvFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat '%s' file: %w", path, err)
	} else if info.IsDir() {
		return nil, fmt.Errorf("'%s' is a directory", path)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read '%s' file: %w", path, err)
	}

	envFile := &envFile{
		vars: loadVars(string(raw)),
		perm: info.Mode(),
		raw:  string(raw),
		mx:   &sync.Mutex{},
	}

	if err := setDefaultVars(envFile); err != nil {
		return nil, fmt.Errorf("failed to set default vars: %w", err)
	}

	return envFile, nil
}

func loadVars(raw string) map[string]*EnvVar {
	lines := strings.Split(string(raw), "\n")
	vars := make(map[string]*EnvVar, len(lines))

	for ldx, line := range lines {
		envVar := &EnvVar{Line: ldx}
		line = trim(line)
		if line == "" {
			continue
		}
		isComment := false
		if strings.HasPrefix(line, "#") {
			isComment = true
			line = trim(strings.TrimPrefix(line, "#"))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		name := trim(parts[0])
		// a commented line only represents a disabled variable assignment
		// (e.g. "#SOME_VAR=value") if what precedes '=' is a valid identifier;
		// otherwise it's just descriptive text that happens to contain '='
		// (e.g. "<=" comparisons or "empty = disabled" phrasing) and must not
		// be treated as a pseudo-variable that never resolves as "unchanged"
		if isComment && !envVarNameRe.MatchString(name) {
			continue
		}
		envVar.IsComment = isComment
		envVar.Name = name
		envVar.Value = trim(stripComments(parts[1]))
		envVar.IsChanged = envVar.Value != parts[1] || envVar.Name != parts[0]
		if envVar.Name != "" {
			vars[envVar.Name] = envVar
		}
	}

	return vars
}

func stripComments(value string) string {
	parts := strings.SplitN(value, " # ", 2)
	if len(parts) == 2 {
		return parts[0]
	}

	return value
}

func setDefaultVars(envFile *envFile) error {
	var defaultConfig config.Config
	if err := env.ParseWithOptions(&defaultConfig, env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(&url.URL{}): func(s string) (any, error) {
				if s == "" {
					return nil, nil
				}
				return url.Parse(s)
			},
		},
		OnSet: func(tag string, value any, isDefault bool) {
			if !isDefault {
				return
			}

			var valueStr string
			switch v := value.(type) {
			case string:
				valueStr = v
			case *url.URL:
				if v != nil {
					valueStr = v.String()
				}
			case int:
				valueStr = strconv.Itoa(v)
			case bool:
				valueStr = strconv.FormatBool(v)
			default:
				valueStr = fmt.Sprintf("%v", v)
			}

			if envVar, ok := envFile.vars[tag]; ok {
				envVar.Default = valueStr
			} else {
				envFile.vars[tag] = &EnvVar{
					Name:    tag,
					Value:   "",
					Default: valueStr,
					Line:    -1,
				}
			}
		},
	}); err != nil {
		return fmt.Errorf("failed to parse env file: %w", err)
	}

	return nil
}
