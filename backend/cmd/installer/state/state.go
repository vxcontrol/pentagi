package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"pentagi/cmd/installer/loader"
)

const EULAConsentFile = "eula-consent"

type State interface {
	Exists() bool
	Reset() error
	Commit() error
	IsDirty() bool

	GetEulaConsent() bool
	SetEulaConsent() error

	SetStack(stack []string) error
	GetStack() []string

	GetVar(name string) (loader.EnvVar, bool)
	SetVar(name, value string) error
	ResetVar(name string) error

	GetVars(names []string) (map[string]loader.EnvVar, map[string]bool)
	SetVars(vars map[string]string) error
	ResetVars(names []string) error

	GetAllVars() map[string]loader.EnvVar
	GetEnvPath() string
}

type stateData struct {
	Stack []string                 `json:"stack"`
	Vars  map[string]loader.EnvVar `json:"vars"`
}

type state struct {
	mx        *sync.Mutex
	envPath   string
	statePath string
	stateDir  string
	stack     []string
	envFile   loader.EnvFile
}

func NewState(envPath string) (State, error) {
	envFile, err := loader.LoadEnvFile(envPath)
	if err != nil {
		return nil, err
	}

	stateDir := filepath.Join(filepath.Dir(envPath), ".state")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, err
	}

	envFileName := filepath.Base(envPath)
	statePath := filepath.Join(stateDir, fmt.Sprintf("%s.state", envFileName))
	s := &state{
		mx:        &sync.Mutex{},
		envPath:   envPath,
		statePath: statePath,
		stateDir:  stateDir,
		envFile:   envFile,
	}

	if info, err := os.Stat(statePath); err == nil && info.IsDir() {
		return nil, fmt.Errorf("'%s' is a directory", statePath)
	} else if err == nil {
		if err := s.loadState(statePath); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (s *state) Exists() bool {
	info, err := os.Stat(s.statePath)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func (s *state) Reset() error {
	s.mx.Lock()
	defer s.mx.Unlock()

	return s.resetState()
}

func (s *state) Commit() error {
	s.mx.Lock()
	defer s.mx.Unlock()

	if err := s.envFile.Save(s.envPath); err != nil {
		return err
	}

	return s.resetState()
}

func (s *state) IsDirty() bool {
	s.mx.Lock()
	defer s.mx.Unlock()

	info, err := os.Stat(s.statePath)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return false
	}

	for _, envVar := range s.envFile.GetAll() {
		if envVar.IsChanged {
			return true
		}
	}

	return false
}

func (s *state) GetEulaConsent() bool {
	s.mx.Lock()
	defer s.mx.Unlock()

	consentFile := filepath.Join(s.stateDir, EULAConsentFile)
	if _, err := os.Stat(consentFile); os.IsNotExist(err) {
		return false
	}

	return true
}

func (s *state) SetEulaConsent() error {
	s.mx.Lock()
	defer s.mx.Unlock()

	currentTime := time.Now().Format(time.RFC3339)
	consentFile := filepath.Join(s.stateDir, EULAConsentFile)
	if err := os.WriteFile(consentFile, []byte(currentTime), 0644); err != nil {
		return fmt.Errorf("failed to write eula consent file: %w", err)
	}

	return nil
}

func (s *state) SetStack(stack []string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.stack = stack

	return s.flushState()
}

func (s *state) GetStack() []string {
	s.mx.Lock()
	defer s.mx.Unlock()

	return s.stack
}

func (s *state) GetVar(name string) (loader.EnvVar, bool) {
	s.mx.Lock()
	defer s.mx.Unlock()

	return s.envFile.Get(name)
}

func (s *state) SetVar(name, value string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.envFile.Set(name, value)

	return s.flushState()
}

func (s *state) ResetVar(name string) error {
	return s.ResetVars([]string{name})
}

func (s *state) GetVars(names []string) (map[string]loader.EnvVar, map[string]bool) {
	s.mx.Lock()
	defer s.mx.Unlock()

	result := make(map[string]loader.EnvVar, len(names))
	present := make(map[string]bool, len(names))

	for _, name := range names {
		envVar, ok := s.envFile.Get(name)
		result[name] = envVar
		present[name] = ok
	}

	return result, present
}

func (s *state) SetVars(vars map[string]string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	for name, value := range vars {
		s.envFile.Set(name, value)
	}

	return s.flushState()
}

func (s *state) ResetVars(names []string) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	envFile, err := loader.LoadEnvFile(s.envPath)
	if err != nil {
		return err
	}

	for _, name := range names {
		// try to keep valuable variables that are not present in the env file
		// but have default value and its default value can be used in the future
		if envVar, ok := envFile.Get(name); ok && (envVar.IsPresent() || envVar.Default != "") {
			s.envFile.Set(name, envVar.Value)
		} else {
			s.envFile.Del(name)
		}
	}

	return s.flushState()
}

func (s *state) GetAllVars() map[string]loader.EnvVar {
	s.mx.Lock()
	defer s.mx.Unlock()

	return s.envFile.GetAll()
}

func (s *state) GetEnvPath() string {
	return s.envPath
}

func (s *state) loadState(stateFile string) error {
	file, err := os.Open(stateFile)
	if err != nil {
		return fmt.Errorf("failed to open state file: %w", err)
	}

	var data stateData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		// if the state file is corrupted, reset it
		data.Stack = []string{}
		data.Vars = make(map[string]loader.EnvVar)
	}

	s.stack = data.Stack
	s.envFile.SetAll(data.Vars)

	return nil
}

func (s *state) flushState() error {
	data := stateData{
		Stack: s.stack,
		Vars:  s.envFile.GetAll(),
	}

	file, err := os.OpenFile(s.statePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create state file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		return fmt.Errorf("failed to encode state file: %w", err)
	}

	return nil
}

func (s *state) resetState() error {
	if err := os.Remove(s.statePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove state file: %w", err)
	}

	envFile, err := loader.LoadEnvFile(s.envPath)
	if err != nil {
		return fmt.Errorf("failed to load state after reset: %w", err)
	}

	s.envFile = envFile
	if err := s.flushState(); err != nil {
		return fmt.Errorf("failed to flush state after reset: %w", err)
	}

	return nil
}
