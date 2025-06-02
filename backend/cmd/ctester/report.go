package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

// PrintAgentResults prints the test results for a single agent
func PrintAgentResults(result AgentTestResult) {
	fmt.Println("\nTest Results:")

	// Basic tests section
	if len(result.BasicTests) > 0 {
		fmt.Println("\nBasic Tests:")
		for _, test := range result.BasicTests {
			status := "✓"
			if !test.Success {
				status = "✗"
			}
			name := test.Name
			if test.Streaming {
				name = fmt.Sprintf("Streaming %s", name)
			}
			fmt.Printf("[%s] %s (%.3fs)\n", status, name, float64(test.LatencyMs)/1000)
			if !test.Success && test.Error != nil {
				fmt.Printf("    Error: %v\n", test.Error)
			}
		}
	}

	// Advanced tests section
	if len(result.AdvancedTests) > 0 {
		fmt.Println("\nAdvanced Tests:")
		for _, test := range result.AdvancedTests {
			status := "✓"
			if !test.Success {
				status = "✗"
			}
			name := test.Name
			if test.Streaming {
				name = fmt.Sprintf("Streaming %s", name)
			}
			fmt.Printf("[%s] %s (%.3fs)\n", status, name, float64(test.LatencyMs)/1000)
			if !test.Success && test.Error != nil {
				fmt.Printf("    Error: %v\n", test.Error)
			}
		}
	} else if result.SkippedAdvanced {
		fmt.Println("\nAdvanced Tests:")
		fmt.Printf("    %s\n", result.SkippedReason)
	}

	// Summary
	successRate := float64(result.TotalSuccess) / float64(result.TotalTests) * 100
	fmt.Printf("\nSummary: %d/%d (%.2f%%) successful tests\n",
		result.TotalSuccess, result.TotalTests, successRate)
	fmt.Printf("Average latency: %.3fs\n", result.AverageLatency.Seconds())
}

// PrintSummaryReport prints the overall summary table of results
func PrintSummaryReport(results []AgentTestResult) {
	fmt.Println("\nOverall Testing Summary:")
	fmt.Println("=================================================")

	// Create a tabwriter for aligned columns
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Agent\tModel\tReasoning\tSuccess Rate\tAvg Latency\t")
	fmt.Fprintln(w, "-----\t-----\t----------\t-----------\t-----------\t")

	var totalSuccess, totalTests int
	var totalLatency time.Duration

	for _, result := range results {
		success := result.TotalSuccess
		total := result.TotalTests
		successRate := float64(success) / float64(total) * 100
		fmt.Fprintf(w, "%s\t%s\t%t\t%d/%d (%.2f%%)\t%.3fs\t\n",
			result.AgentType,
			result.ModelName,
			result.Reasoning,
			success,
			total,
			successRate,
			result.AverageLatency.Seconds())

		totalSuccess += success
		totalTests += total
		totalLatency += result.AverageLatency * time.Duration(total)
	}

	w.Flush()

	if totalTests > 0 {
		overallSuccessRate := float64(totalSuccess) / float64(totalTests) * 100
		overallAvgLatency := totalLatency / time.Duration(totalTests)
		fmt.Printf("\nTotal: %d/%d (%.2f%%) successful tests\n", totalSuccess, totalTests, overallSuccessRate)
		fmt.Printf("Overall average latency: %.3fs\n", overallAvgLatency.Seconds())
	}
}

// WriteReportToFile writes the test results to a report file in Markdown format
func WriteReportToFile(results []AgentTestResult, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	file.WriteString("# LLM Agent Testing Report\n\n")
	file.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().UTC().Format(time.RFC1123)))

	// Create a table for overall results
	file.WriteString("## Overall Results\n\n")
	file.WriteString("| Agent | Model | Reasoning | Success Rate | Average Latency |\n")
	file.WriteString("|-------|-------|-----------|--------------|-----------------|\n")

	var totalSuccess, totalTests int
	var totalLatency time.Duration

	for _, result := range results {
		success := result.TotalSuccess
		total := result.TotalTests
		successRate := float64(success) / float64(total) * 100
		file.WriteString(fmt.Sprintf("| %s | %s | %t | %d/%d (%.2f%%) | %.3fs |\n",
			result.AgentType,
			result.ModelName,
			result.Reasoning,
			success,
			total,
			successRate,
			result.AverageLatency.Seconds()))

		totalSuccess += success
		totalTests += total
		totalLatency += result.AverageLatency * time.Duration(total)
	}

	// Write summary
	if totalTests > 0 {
		overallSuccessRate := float64(totalSuccess) / float64(totalTests) * 100
		overallAvgLatency := totalLatency / time.Duration(totalTests)
		file.WriteString(fmt.Sprintf("\n**Total**: %d/%d (%.2f%%) successful tests\n",
			totalSuccess, totalTests, overallSuccessRate))
		file.WriteString(fmt.Sprintf("**Overall average latency**: %.3fs\n\n", overallAvgLatency.Seconds()))
	}

	// Write detailed results for each agent
	file.WriteString("## Detailed Results\n\n")

	for _, result := range results {
		file.WriteString(fmt.Sprintf("### %s (%s)\n\n", result.AgentType, result.ModelName))

		// Basic tests
		if len(result.BasicTests) > 0 {
			file.WriteString("#### Basic Tests\n\n")
			file.WriteString("| Test | Result | Latency | Error |\n")
			file.WriteString("|------|--------|---------|-------|\n")

			for _, test := range result.BasicTests {
				status := "✅ Pass"
				errorMsg := ""
				if !test.Success {
					status = "❌ Fail"
					if test.Error != nil {
						errorMsg = EscapeMarkdown(test.Error.Error())
					}
				}
				name := test.Name
				if test.Streaming {
					name = fmt.Sprintf("Streaming %s", name)
				}

				file.WriteString(fmt.Sprintf("| %s | %s | %.3fs | %s |\n",
					name,
					status,
					float64(test.LatencyMs)/1000,
					errorMsg))
			}
			file.WriteString("\n")
		}

		// Advanced tests
		if len(result.AdvancedTests) > 0 {
			file.WriteString("#### Advanced Tests\n\n")
			file.WriteString("| Test | Result | Latency | Error |\n")
			file.WriteString("|------|--------|---------|-------|\n")

			for _, test := range result.AdvancedTests {
				status := "✅ Pass"
				errorMsg := ""
				if !test.Success {
					status = "❌ Fail"
					if test.Error != nil {
						errorMsg = EscapeMarkdown(test.Error.Error())
					}
				}
				name := test.Name
				if test.Streaming {
					name = fmt.Sprintf("Streaming %s", name)
				}

				file.WriteString(fmt.Sprintf("| %s | %s | %.3fs | %s |\n",
					name,
					status,
					float64(test.LatencyMs)/1000,
					errorMsg))
			}
			file.WriteString("\n")
		} else if result.SkippedAdvanced {
			file.WriteString("#### Advanced Tests\n\n")
			file.WriteString(fmt.Sprintf("*%s*\n\n", result.SkippedReason))
		}

		// Summary
		successRate := float64(result.TotalSuccess) / float64(result.TotalTests) * 100
		file.WriteString(fmt.Sprintf("**Summary**: %d/%d (%.2f%%) successful tests\n\n",
			result.TotalSuccess, result.TotalTests, successRate))
		file.WriteString(fmt.Sprintf("**Average latency**: %.3fs\n\n", result.AverageLatency.Seconds()))
		file.WriteString("---\n\n")
	}

	return nil
}
