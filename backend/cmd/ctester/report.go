package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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
			if note := failureText(test); !test.Success && note != "" {
				fmt.Printf("    Error: %s\n", note)
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
			if note := failureText(test); !test.Success && note != "" {
				fmt.Printf("    Error: %s\n", note)
			}
		}
	} else if result.SkippedAdvanced {
		fmt.Println("\nAdvanced Tests:")
		fmt.Printf("    %s\n", result.SkippedReason)
	}

	// Capability tests section (adaptive thinking / reasoning off / structured
	// output): only present when the model was actually classified as
	// supporting the capability, so its absence here means "not applicable to
	// this model", not "not tested".
	if len(result.CapabilityTests) > 0 {
		fmt.Println("\nCapability Tests:")
		for _, test := range result.CapabilityTests {
			status := "✓"
			switch {
			case test.Unsupported:
				status = "⊘"
			case !test.Success:
				status = "✗"
			}
			fmt.Printf("[%s] %s (%s) (%.3fs)\n", status, test.Name, test.Capability, float64(test.LatencyMs)/1000)
			if test.Unsupported {
				fmt.Printf("    Not supported by this model/provider: %v\n", test.Error)
			} else if note := failureText(test); !test.Success && note != "" {
				fmt.Printf("    Error: %s\n", note)
			}
		}
	}

	// Summary
	var successRate float64
	if result.TotalTests > 0 {
		successRate = float64(result.TotalSuccess) / float64(result.TotalTests) * 100
	}
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

	var totalSuccess, totalTests, totalFiltered int
	var totalLatency time.Duration

	for _, result := range results {
		success := result.TotalSuccess
		total := result.TotalTests
		totalFiltered += result.TotalFiltered
		var successRate float64
		if total > 0 {
			successRate = float64(success) / float64(total) * 100
		}
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
		if totalFiltered > 0 {
			fmt.Printf("Refused by a vendor's content filter: %d of the failed tests\n", totalFiltered)
		}
		fmt.Printf("Overall average latency: %.3fs\n", overallAvgLatency.Seconds())
	}
}

type errWriter struct {
	dst io.Writer
	err error
}

func (ew *errWriter) Write(p []byte) (int, error) {
	if ew.err != nil {
		return 0, ew.err
	}

	n, err := ew.dst.Write(p)
	ew.err = err

	return n, err
}

// WriteReportToFile writes the test results to a report file in Markdown format
func WriteReportToFile(results []AgentTestResult, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	return writeReport(file, results)
}

func writeReport(dst io.WriteCloser, results []AgentTestResult) (err error) {
	defer func() {
		err = errors.Join(err, dst.Close())
	}()

	w := &errWriter{dst: dst}

	// Write header
	_, _ = io.WriteString(w, "# LLM Agent Testing Report\n\n")
	fmt.Fprintf(w, "Generated: %s\n\n", time.Now().UTC().Format(time.RFC1123))

	// Create a table for overall results
	_, _ = io.WriteString(w, "## Overall Results\n\n")
	_, _ = io.WriteString(w, "| Agent | Model | Reasoning | Success Rate | Average Latency |\n")
	_, _ = io.WriteString(w, "|-------|-------|-----------|--------------|-----------------|\n")

	var totalSuccess, totalTests, totalFiltered int
	var totalLatency time.Duration

	for _, result := range results {
		success := result.TotalSuccess
		total := result.TotalTests
		totalFiltered += result.TotalFiltered
		var successRate float64
		if total > 0 {
			successRate = float64(success) / float64(total) * 100
		}
		fmt.Fprintf(w, "| %s | %s | %t | %d/%d (%.2f%%) | %.3fs |\n",
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

	// Write summary
	if totalTests > 0 {
		overallSuccessRate := float64(totalSuccess) / float64(totalTests) * 100
		overallAvgLatency := totalLatency / time.Duration(totalTests)
		fmt.Fprintf(w, "\n**Total**: %d/%d (%.2f%%) successful tests\n",
			totalSuccess, totalTests, overallSuccessRate)
		if totalFiltered > 0 {
			fmt.Fprintf(w, "**Refused by a vendor's content filter**: %d of the failed tests\n", totalFiltered)
		}
		fmt.Fprintf(w, "**Overall average latency**: %.3fs\n\n", overallAvgLatency.Seconds())
	}

	// Write detailed results for each agent
	_, _ = io.WriteString(w, "## Detailed Results\n\n")

	for _, result := range results {
		fmt.Fprintf(w, "### %s (%s)\n\n", result.AgentType, result.ModelName)

		// Basic tests
		if len(result.BasicTests) > 0 {
			_, _ = io.WriteString(w, "#### Basic Tests\n\n")
			_, _ = io.WriteString(w, "| Test | Result | Latency | Error |\n")
			_, _ = io.WriteString(w, "|------|--------|---------|-------|\n")

			for _, test := range result.BasicTests {
				status, errorMsg := rowStatus(test)
				name := test.Name
				if test.Streaming {
					name = fmt.Sprintf("Streaming %s", name)
				}

				fmt.Fprintf(w, "| %s | %s | %.3fs | %s |\n",
					name,
					status,
					float64(test.LatencyMs)/1000,
					errorMsg)
			}
			_, _ = io.WriteString(w, "\n")
		}

		// Advanced tests
		if len(result.AdvancedTests) > 0 {
			_, _ = io.WriteString(w, "#### Advanced Tests\n\n")
			_, _ = io.WriteString(w, "| Test | Result | Latency | Error |\n")
			_, _ = io.WriteString(w, "|------|--------|---------|-------|\n")

			for _, test := range result.AdvancedTests {
				status, errorMsg := rowStatus(test)
				name := test.Name
				if test.Streaming {
					name = fmt.Sprintf("Streaming %s", name)
				}

				fmt.Fprintf(w, "| %s | %s | %.3fs | %s |\n",
					name,
					status,
					float64(test.LatencyMs)/1000,
					errorMsg)
			}
			_, _ = io.WriteString(w, "\n")
		} else if result.SkippedAdvanced {
			_, _ = io.WriteString(w, "#### Advanced Tests\n\n")
			fmt.Fprintf(w, "*%s*\n\n", result.SkippedReason)
		}

		// Capability tests (adaptive thinking / reasoning off / structured
		// output) - absent entirely for a model not classified as supporting
		// the capability at all, distinguished from "attempted but rejected"
		// (Unsupported) and from a genuine failure.
		if len(result.CapabilityTests) > 0 {
			_, _ = io.WriteString(w, "#### Capability Tests\n\n")
			_, _ = io.WriteString(w, "| Test | Capability | Result | Latency | Note |\n")
			_, _ = io.WriteString(w, "|------|------------|--------|---------|------|\n")

			for _, test := range result.CapabilityTests {
				status, note := rowStatus(test)

				fmt.Fprintf(w, "| %s | %s | %s | %.3fs | %s |\n",
					test.Name,
					test.Capability,
					status,
					float64(test.LatencyMs)/1000,
					note)
			}
			_, _ = io.WriteString(w, "\n")
		}

		// Summary
		var successRate float64
		if result.TotalTests > 0 {
			successRate = float64(result.TotalSuccess) / float64(result.TotalTests) * 100
		}
		fmt.Fprintf(w, "**Summary**: %d/%d (%.2f%%) successful tests\n\n",
			result.TotalSuccess, result.TotalTests, successRate)
		if result.TotalFiltered > 0 {
			fmt.Fprintf(w, "**Refused by the vendor's content filter**: %d of the failed tests\n\n", result.TotalFiltered)
		}
		fmt.Fprintf(w, "**Average latency**: %.3fs\n\n", result.AverageLatency.Seconds())
		_, _ = io.WriteString(w, "---\n\n")
	}

	return w.err
}

func rowStatus(test TestResult) (status, note string) {
	switch {
	case test.Unsupported:
		if test.Error != nil {
			note = TruncateString(EscapeMarkdown(test.Error.Error()), 150)
		}
		return "⊘ Not Supported", note
	case test.Success:
		return "✅ Pass", ""
	default:
		return failureStatus(test), TruncateString(EscapeMarkdown(failureText(test)), 150)
	}
}

func failureStatus(test TestResult) string {
	if test.ContentFiltered {
		return "⛔ Filtered"
	}
	return "❌ Fail"
}

func failureText(test TestResult) string {
	var text string
	if test.Error != nil {
		text = test.Error.Error()
	}
	if test.StopReason != "" && !strings.Contains(text, test.StopReason) {
		if text != "" {
			text += "; "
		}
		text += "stop reason: " + test.StopReason
	}

	return text
}
