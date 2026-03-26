package loader

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"pentagi/pkg/config"

	"github.com/caarlos0/env/v10"
)

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
		if strings.HasPrefix(line, "#") {
			envVar.IsComment = true
			line = trim(strings.TrimPrefix(line, "#"))
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		envVar.Name = trim(parts[0])
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
