package loader

import (
	"fmt"
	"os"
	"path/filepath"
)

// Example demonstrates the full workflow of loading, modifying, and saving .env files
func ExampleEnvFile_workflow() {
	// Create a temporary .env file
	tmpDir, _ := os.MkdirTemp("", "example")
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	initialContent := `# PentAGI Configuration
DATABASE_URL=postgres://localhost:5432/db
DEBUG=false
# API Settings
API_KEY=old_key`

	os.WriteFile(envPath, []byte(initialContent), 0644)

	// Step 1: Load existing .env file
	envFile, err := LoadEnvFile(envPath)
	if err != nil {
		panic(err)
	}

	// Step 2: Display current values and defaults (only variables from file)
	fmt.Println("Current configuration:")
	fileVars := []string{"DATABASE_URL", "DEBUG", "API_KEY"}
	for _, name := range fileVars {
		envVar, exists := envFile.Get(name)
		if !exists {
			fmt.Printf("%s = (not present)\n", name)
			continue
		}
		if envVar.IsPresent() && !envVar.IsComment {
			fmt.Printf("%s = %s", name, envVar.Value)
			if envVar.Default != "" && envVar.Default != envVar.Value {
				fmt.Printf(" (default: %s)", envVar.Default)
			}
			if envVar.IsChanged {
				fmt.Printf(" [modified]")
			}
			fmt.Println()
		}
	}

	// Step 3: User modifies values
	envFile.Set("DEBUG", "true")
	envFile.Set("API_KEY", "new_secret_key")
	envFile.Set("NEW_SETTING", "added_value")

	// Step 4: Save changes (creates backup automatically)
	err = envFile.Save(envPath)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nConfiguration saved successfully!")
	fmt.Println("Backup created in .bak directory")

	// Output:
	// Current configuration:
	// DATABASE_URL = postgres://localhost:5432/db (default: postgres://pentagiuser:pentagipass@pgvector:5432/pentagidb?sslmode=disable)
	// DEBUG = false
	// API_KEY = old_key
	//
	// Configuration saved successfully!
	// Backup created in .bak directory
}
