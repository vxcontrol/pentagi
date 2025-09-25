package state

import (
	"fmt"
	"os"
	"path/filepath"
)

// Example demonstrates the complete workflow of state management for .env files
func ExampleState_transactionWorkflow() {
	// Setup: Create a test .env file
	tmpDir, _ := os.MkdirTemp("", "state_example")
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	initialContent := `# Database Configuration
DATABASE_URL=postgres://localhost:5432/olddb
DATABASE_PASSWORD=old_password
# API Configuration
API_HOST=localhost
API_PORT=8080`

	os.WriteFile(envPath, []byte(initialContent), 0644)

	fmt.Println("=== PentAGI Configuration Manager ===")
	fmt.Println("Starting configuration process...")

	// Step 1: Initialize state management
	state, err := NewState(envPath)
	if err != nil {
		panic(err)
	}

	// Step 2: Multi-step configuration process
	fmt.Println("\n--- Step 1: Database Configuration ---")
	state.SetStack([]string{"configure_database"})

	// User makes changes gradually
	state.SetVar("DATABASE_URL", "postgres://prod-server:5432/pentagidb")
	state.SetVar("DATABASE_PASSWORD", "secure_prod_password")
	state.SetVar("DATABASE_POOL_SIZE", "20")

	fmt.Printf("Current step: %s\n", state.GetStack()[0])
	fmt.Printf("Modified variables: %d\n", countChangedVars(state))

	// Step 3: Continue with API configuration
	fmt.Println("\n--- Step 2: API Configuration ---")
	state.SetStack([]string{"configure_api"})

	state.SetVar("API_HOST", "0.0.0.0")
	state.SetVar("API_PORT", "443")
	state.SetVar("API_SSL_ENABLED", "true")

	fmt.Printf("Current step: %s\n", state.GetStack()[0])
	fmt.Printf("Total modified variables: %d\n", countChangedVars(state))

	// Show current state
	fmt.Println("\n--- Current Configuration ---")
	showCurrentConfig(state)

	// Step 4: User can choose to commit or reset
	fmt.Println("\n--- Decision: Commit Changes ---")

	// Commit applies all changes to .env file and cleans up state
	err = state.Commit()
	if err != nil {
		panic(err)
	}

	fmt.Println("Changes committed successfully!")
	fmt.Printf("State file exists: %v\n", state.Exists())

	// Output:
	// === PentAGI Configuration Manager ===
	// Starting configuration process...
	//
	// --- Step 1: Database Configuration ---
	// Current step: configure_database
	// Modified variables: 3
	//
	// --- Step 2: API Configuration ---
	// Current step: configure_api
	// Total modified variables: 6
	//
	// --- Current Configuration ---
	// DATABASE_URL: postgres://prod-server:5432/pentagidb [CHANGED]
	// DATABASE_PASSWORD: secure_prod_password [CHANGED]
	// DATABASE_POOL_SIZE: 20 [NEW]
	// API_HOST: 0.0.0.0 [CHANGED]
	// API_PORT: 443 [CHANGED]
	// API_SSL_ENABLED: true [NEW]
	//
	// --- Decision: Commit Changes ---
	// Changes committed successfully!
	// State file exists: true
}

// Example demonstrates rollback functionality
func ExampleState_rollbackWorkflow() {
	tmpDir, _ := os.MkdirTemp("", "rollback_example")
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	originalContent := "IMPORTANT_SETTING=production_value"
	os.WriteFile(envPath, []byte(originalContent), 0644)

	fmt.Println("=== Configuration Rollback Example ===")

	state, _ := NewState(envPath)

	// User starts making risky changes
	fmt.Println("Making risky changes...")
	state.SetStack([]string{"risky_configuration"})
	state.SetVar("IMPORTANT_SETTING", "experimental_value")
	state.SetVar("DANGEROUS_SETTING", "could_break_system")

	fmt.Printf("Changes pending: %d\n", countChangedVars(state))

	// User realizes they made a mistake
	fmt.Println("Oops! These changes might break the system...")
	fmt.Println("Rolling back all changes...")

	// Reset discards all changes and preserves original file
	err := state.Reset()
	if err != nil {
		panic(err)
	}

	fmt.Println("All changes discarded!")
	fmt.Printf("State file exists: %v\n", state.Exists())

	// Verify original file is unchanged
	content, _ := os.ReadFile(envPath)
	fmt.Printf("Original file preserved: %s", string(content))

	// Output:
	// === Configuration Rollback Example ===
	// Making risky changes...
	// Changes pending: 2
	// Oops! These changes might break the system...
	// Rolling back all changes...
	// All changes discarded!
	// State file exists: true
	// Original file preserved: IMPORTANT_SETTING=production_value
}

// Example demonstrates persistence across sessions
func ExampleState_persistenceWorkflow() {
	tmpDir, _ := os.MkdirTemp("", "persistence_example")
	defer os.RemoveAll(tmpDir)

	envPath := filepath.Join(tmpDir, ".env")
	os.WriteFile(envPath, []byte("VAR1=value1"), 0644)

	fmt.Println("=== Session Persistence Example ===")

	// Session 1: User starts configuration
	fmt.Println("Session 1: Starting configuration...")
	state1, _ := NewState(envPath)
	state1.SetStack([]string{"partial_configuration"})
	state1.SetVar("VAR1", "modified_value")
	state1.SetVar("VAR2", "new_value")

	fmt.Printf("Session 1 - Step: %s, Changes: %d\n",
		state1.GetStack()[0], countChangedVars(state1))

	// Session 1 ends (simulated by network disconnection, etc.)
	fmt.Println("Session 1 ended unexpectedly...")

	// Session 2: User reconnects and resumes
	fmt.Println("\nSession 2: Resuming configuration...")
	state2, _ := NewState(envPath) // Automatically loads saved state

	fmt.Printf("Session 2 - Restored Step: %s, Changes: %d\n",
		state2.GetStack()[0], countChangedVars(state2))

	// Continue from where left off
	state2.SetStack([]string{"complete_configuration"})
	state2.SetVar("VAR3", "final_value")

	fmt.Printf("Session 2 - Final Step: %s, Changes: %d\n",
		state2.GetStack()[0], countChangedVars(state2))

	// Commit when ready
	state2.Commit()
	fmt.Println("Configuration completed successfully!")

	// Output:
	// === Session Persistence Example ===
	// Session 1: Starting configuration...
	// Session 1 - Step: partial_configuration, Changes: 2
	// Session 1 ended unexpectedly...
	//
	// Session 2: Resuming configuration...
	// Session 2 - Restored Step: partial_configuration, Changes: 2
	// Session 2 - Final Step: complete_configuration, Changes: 3
	// Configuration completed successfully!
}

func countChangedVars(state State) int {
	count := 0
	for _, envVar := range state.GetAllVars() {
		if envVar.IsChanged {
			count++
		}
	}
	return count
}

func showCurrentConfig(state State) {
	// Show specific variables in fixed order for consistent output
	vars := []string{"DATABASE_URL", "DATABASE_PASSWORD", "DATABASE_POOL_SIZE",
		"API_HOST", "API_PORT", "API_SSL_ENABLED"}

	allVars := state.GetAllVars()
	for _, name := range vars {
		if envVar, exists := allVars[name]; exists && envVar.IsChanged {
			status := " [CHANGED]"
			if !envVar.IsPresent() {
				status = " [NEW]"
			}
			fmt.Printf("%s: %s%s\n", name, envVar.Value, status)
		}
	}
}
