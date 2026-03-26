# Reference Configuration Pattern

> Reference implementation of configuration using ScraperConfig as an example for all future configurations in StateController.

## 🎯 **Core Principles**

### 1. **Use loader.EnvVar for direct form mapping**
```go
type ReferenceConfig struct {
    // ✅ Direct form mapping - use loader.EnvVar
    DirectField1 loader.EnvVar // SOME_ENV_VAR
    DirectField2 loader.EnvVar // ANOTHER_ENV_VAR

    // ✅ Computed fields - simple types
    ComputedMode string // computed from DirectField1

    // ✅ Temporary data for processing
    TempData string // not saved, used for logic
}
```

### 2. **Reference configuration structure**
```go
// ScraperConfig represents scraper configuration settings
type ScraperConfig struct {
    // direct form field mappings using loader.EnvVar
    // these fields directly correspond to environment variables and form inputs (not computed)
    PublicURL             loader.EnvVar // SCRAPER_PUBLIC_URL
    PrivateURL            loader.EnvVar // SCRAPER_PRIVATE_URL
    LocalUsername         loader.EnvVar // LOCAL_SCRAPER_USERNAME
    LocalPassword         loader.EnvVar // LOCAL_SCRAPER_PASSWORD
    MaxConcurrentSessions loader.EnvVar // LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS

    // computed fields (not directly mapped to env vars)
    // these are derived from the above EnvVar fields
    Mode string // "embedded", "external", "disabled" - computed from PrivateURL

    // parsed credentials for external mode (extracted from URLs)
    PublicUsername  string
    PublicPassword  string
    PrivateUsername string
    PrivatePassword string
}
```

### 3. **Constants for default values**
```go
const (
    DefaultScraperBaseURL = "https://scraper/"
    DefaultScraperDomain  = "scraper"
    DefaultScraperSchema  = "https"
)
```

## 🔧 **Configuration Methods**

### 1. **GetConfig() - Retrieve configuration**
```go
func (c *StateController) GetScraperConfig() *ScraperConfig {
    // get all environment variables using the state controller
    publicURL, _ := c.GetVar("SCRAPER_PUBLIC_URL")
    privateURL, _ := c.GetVar("SCRAPER_PRIVATE_URL")
    localUsername, _ := c.GetVar("LOCAL_SCRAPER_USERNAME")
    localPassword, _ := c.GetVar("LOCAL_SCRAPER_PASSWORD")
    maxSessions, _ := c.GetVar("LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS")

    config := &ScraperConfig{
        PublicURL:             publicURL,
        PrivateURL:            privateURL,
        LocalUsername:         localUsername,
        LocalPassword:         localPassword,
        MaxConcurrentSessions: maxSessions,
    }

    // compute derived fields using multiple inputs
    config.Mode = c.determineScraperMode(privateURL.Value, publicURL.Value)

    // for external mode, extract credentials from URLs
    if config.Mode == "external" {
        config.PublicUsername, config.PublicPassword = c.extractCredentialsFromURL(publicURL.Value)
        config.PrivateUsername, config.PrivatePassword = c.extractCredentialsFromURL(privateURL.Value)
    }

    return config
}
```

### 2. **UpdateConfig() - Update configuration**
```go
func (c *StateController) UpdateScraperConfig(config *ScraperConfig) error {
    if config == nil {
        return fmt.Errorf("config cannot be nil")
    }

    switch config.Mode {
    case "disabled":
        // clear scraper URLs, preserve local settings
        if err := c.SetVar("SCRAPER_PUBLIC_URL", ""); err != nil {
            return fmt.Errorf("failed to clear SCRAPER_PUBLIC_URL: %w", err)
        }
        if err := c.SetVar("SCRAPER_PRIVATE_URL", ""); err != nil {
            return fmt.Errorf("failed to clear SCRAPER_PRIVATE_URL: %w", err)
        }

    case "external":
        // construct URLs with credentials if provided
        publicURL := config.PublicURL.Value
        if config.PublicUsername != "" && config.PublicPassword != "" {
            publicURL = c.addCredentialsToURL(config.PublicURL.Value, config.PublicUsername, config.PublicPassword)
        }

        privateURL := config.PrivateURL.Value
        if config.PrivateUsername != "" && config.PrivatePassword != "" {
            privateURL = c.addCredentialsToURL(config.PrivateURL.Value, config.PrivateUsername, config.PrivatePassword)
        }

        if err := c.SetVar("SCRAPER_PUBLIC_URL", publicURL); err != nil {
            return fmt.Errorf("failed to set SCRAPER_PUBLIC_URL: %w", err)
        }
        if err := c.SetVar("SCRAPER_PRIVATE_URL", privateURL); err != nil {
            return fmt.Errorf("failed to set SCRAPER_PRIVATE_URL: %w", err)
        }

    case "embedded":
        // handle embedded mode with credential mapping
        publicURL := config.PublicURL.Value
        if config.PublicUsername != "" && config.PublicPassword != "" {
            // fallback to private URL if public URL is not set
            if privateURL := config.PrivateURL.Value; privateURL != "" && publicURL == "" {
                publicURL = privateURL
            }
            publicURL = c.addCredentialsToURL(publicURL, config.PublicUsername, config.PublicPassword)
        }

        privateURL := config.PrivateURL.Value
        if config.PrivateUsername != "" && config.PrivatePassword != "" {
            privateURL = c.addCredentialsToURL(privateURL, config.PrivateUsername, config.PrivatePassword)
        }

        // update all relevant variables
        if err := c.SetVar("SCRAPER_PUBLIC_URL", publicURL); err != nil {
            return fmt.Errorf("failed to set SCRAPER_PUBLIC_URL: %w", err)
        }
        if err := c.SetVar("SCRAPER_PRIVATE_URL", privateURL); err != nil {
            return fmt.Errorf("failed to set SCRAPER_PRIVATE_URL: %w", err)
        }

        // map credentials to local settings
        if err := c.SetVar("LOCAL_SCRAPER_USERNAME", config.PrivateUsername); err != nil {
            return fmt.Errorf("failed to set LOCAL_SCRAPER_USERNAME: %w", err)
        }
        if err := c.SetVar("LOCAL_SCRAPER_PASSWORD", config.PrivatePassword); err != nil {
            return fmt.Errorf("failed to set LOCAL_SCRAPER_PASSWORD: %w", err)
        }
        if err := c.SetVar("LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS", config.MaxConcurrentSessions.Value); err != nil {
            return fmt.Errorf("failed to set LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS: %w", err)
        }
    }

    return nil
}
```

### 3. **ResetConfig() - Reset to defaults**
```go
func (c *StateController) ResetScraperConfig() *ScraperConfig {
    // reset all scraper-related environment variables to their defaults
    vars := []string{
        "SCRAPER_PUBLIC_URL",
        "SCRAPER_PRIVATE_URL",
        "LOCAL_SCRAPER_USERNAME",
        "LOCAL_SCRAPER_PASSWORD",
        "LOCAL_SCRAPER_MAX_CONCURRENT_SESSIONS",
    }

    if err := c.ResetVars(vars); err != nil {
        return nil
    }

    return c.GetScraperConfig()
}
```

## 🧩 **Helper Methods**

### 1. **Mode determination with multiple inputs**
```go
func (c *StateController) determineScraperMode(privateURL, publicURL string) string {
    if privateURL == "" && publicURL == "" {
        return "disabled"
    }

    parsedURL, err := url.Parse(privateURL)
    if err != nil {
        return "external"
    }

    if parsedURL.Scheme == DefaultScraperSchema && parsedURL.Hostname() == DefaultScraperDomain {
        return "embedded"
    }

    return "external"
}
```

### 2. **URL credential handling with defaults**
```go
func (c *StateController) addCredentialsToURL(urlStr, username, password string) string {
    if username == "" || password == "" {
        return urlStr
    }

    if urlStr == "" {
        urlStr = DefaultScraperBaseURL
    }

    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        return urlStr
    }

    // set user info
    parsedURL.User = url.UserPassword(username, password)

    return parsedURL.String()
}
```

### 3. **Public method for safe display**
```go
// RemoveCredentialsFromURL removes credentials from URL - public method for form display
func (c *StateController) RemoveCredentialsFromURL(urlStr string) string {
    if urlStr == "" {
        return urlStr
    }

    parsedURL, err := url.Parse(urlStr)
    if err != nil {
        return urlStr
    }

    parsedURL.User = nil
    return parsedURL.String()
}
```

## 📋 **Form Integration**

### 1. **Field creation**
```go
func (m *FormModel) createURLField(key, title, description, placeholder string) FormField {
    input := textinput.New()
    input.Placeholder = placeholder

    var value string
    switch key {
    case "public_url":
        value = m.config.PublicURL.Value
    case "private_url":
        value = m.config.PrivateURL.Value
    }

    if value != "" {
        input.SetValue(value)
    }

    return FormField{
        Key:         key,
        Title:       title,
        Description: description,
        Input:       input,
        Value:       input.Value(),
    }
}
```

### 2. **Save handling with validation**
```go
func (m *ScraperFormModel) HandleSave() error {
    mode := m.getSelectedMode()
    fields := m.GetFormFields()

    // create a working copy of the current config to modify
    newConfig := &controllers.ScraperConfig{
        Mode: mode,
        // copy current EnvVar fields - they preserve metadata like Line, IsPresent, etc.
        PublicURL:             m.config.PublicURL,
        PrivateURL:            m.config.PrivateURL,
        LocalUsername:         m.config.LocalUsername,
        LocalPassword:         m.config.LocalPassword,
        MaxConcurrentSessions: m.config.MaxConcurrentSessions,
    }

    // update field values based on form input
    for _, field := range fields {
        value := strings.TrimSpace(field.Input.Value())

        switch field.Key {
        case "public_url":
            newConfig.PublicURL.Value = value
        case "private_url":
            newConfig.PrivateURL.Value = value
        case "local_username":
            newConfig.LocalUsername.Value = value
        case "local_password":
            newConfig.LocalPassword.Value = value
        case "max_sessions":
            // validate numeric input
            if value != "" {
                if _, err := strconv.Atoi(value); err != nil {
                    return fmt.Errorf("invalid number for max concurrent sessions: %s", value)
                }
            }
            newConfig.MaxConcurrentSessions.Value = value
        }
    }

    // set defaults for embedded mode if needed
    if mode == "embedded" {
        if newConfig.LocalUsername.Value == "" {
            newConfig.LocalUsername.Value = "someuser"
        }
        if newConfig.LocalPassword.Value == "" {
            newConfig.LocalPassword.Value = "somepass"
        }
        if newConfig.MaxConcurrentSessions.Value == "" {
            newConfig.MaxConcurrentSessions.Value = "10"
        }
    }

    // save the configuration
    if err := m.GetController().UpdateScraperConfig(newConfig); err != nil {
        return err
    }

    // reload config to get updated state
    m.config = m.GetController().GetScraperConfig()
    return nil
}
```

### 3. **Safe display in overview**
```go
func (m *ScraperFormModel) GetFormOverview() string {
    config := m.GetController().GetScraperConfig()

    var sections []string
    sections = append(sections, "Current Configuration:")

    switch config.Mode {
    case "embedded":
        sections = append(sections, "• Mode: Embedded")
        if config.PublicURL.Value != "" {
            sections = append(sections, "• Public URL: " + config.PublicURL.Value)
        }
        if config.LocalUsername.Value != "" {
            sections = append(sections, "• Local Username: " + config.LocalUsername.Value)
        }

    case "external":
        sections = append(sections, "• Mode: External")
        if config.PublicURL.Value != "" {
            // show clean URL without credentials for security
            cleanURL := m.GetController().RemoveCredentialsFromURL(config.PublicURL.Value)
            sections = append(sections, "• Public URL: " + cleanURL)
        }
        if config.PrivateURL.Value != "" {
            cleanURL := m.GetController().RemoveCredentialsFromURL(config.PrivateURL.Value)
            sections = append(sections, "• Private URL: " + cleanURL)
        }

    case "disabled":
        sections = append(sections, "• Mode: Disabled")
    }

    return strings.Join(sections, "\n")
}
```

## ✅ **Benefits of Reference Approach**

1. **State tracking**: `loader.EnvVar` tracks changes, file presence, default values
2. **Metadata preservation**: Information about changes, presence, defaults
3. **Security**: Public methods for safe display of sensitive data
4. **Consistency**: Uniform behavior across all configurations
5. **Reliability**: Minimizes errors in different usage scenarios
6. **URL handling**: Uses `net/url` package for robust URL parsing
7. **Default management**: Constants for maintainable default values

## 🔄 **Key Patterns**

### 1. **Data Types**
| Field Type | Data Type | Usage |
|------------|-----------|-------|
| **Direct mapping** | `loader.EnvVar` | Form fields, env variables |
| **Computed** | `string`/`bool`/`int` | Modes, status, flags |
| **Temporary** | `string` | Parsing, processing |

### 2. **Method signatures**
- `GetConfig() *Config` - retrieves with metadata
- `UpdateConfig(config *Config) error` - saves with validation
- `ResetConfig() *Config` - resets to defaults
- `PublicMethod()` - exported for form usage

### 3. **Error handling**
- Always validate input parameters
- Use `fmt.Errorf` with context
- Handle URL parsing errors gracefully
- Provide meaningful error messages

## 🚀 **Creating New Configurations**

```go
// 1. Define structure
type NewServiceConfig struct {
    // direct fields
    APIKey    loader.EnvVar // NEW_SERVICE_API_KEY
    BaseURL   loader.EnvVar // NEW_SERVICE_BASE_URL
    Enabled   loader.EnvVar // NEW_SERVICE_ENABLED

    // computed fields
    IsConfigured bool
}

// 2. Add constants
const (
    DefaultNewServiceURL = "https://api.newservice.com"
)

// 3. Implement methods
func (c *StateController) GetNewServiceConfig() *NewServiceConfig { /* ... */ }
func (c *StateController) UpdateNewServiceConfig(config *NewServiceConfig) error { /* ... */ }
func (c *StateController) ResetNewServiceConfig() *NewServiceConfig { /* ... */ }

// 4. Create form following ScraperFormModel pattern
```

This reference approach ensures reliable and consistent operation of all configurations in the system.
