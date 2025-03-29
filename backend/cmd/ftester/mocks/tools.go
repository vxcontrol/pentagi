package mocks

import (
	"encoding/json"
	"fmt"
	"strings"

	"pentagi/pkg/terminal"
	"pentagi/pkg/tools"
)

// MockResponse generates a mock response for a function
func MockResponse(funcName string, args json.RawMessage) (string, error) {
	var resultObj any

	switch funcName {
	case tools.TerminalToolName:
		var termArgs tools.TerminalAction
		if err := json.Unmarshal(args, &termArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling terminal arguments: %w", err)
		}

		terminal.PrintMock("Would execute terminal command:")
		terminal.PrintKeyValue("Command", termArgs.Input)
		terminal.PrintKeyValue("Working directory", termArgs.Cwd)
		terminal.PrintKeyValueFormat("Timeout", "%d seconds", termArgs.Timeout.Int64())
		terminal.PrintKeyValueFormat("Detach", "%v", termArgs.Detach.Bool())

		if termArgs.Detach.Bool() {
			resultObj = "Command executed successfully in the background mode"
		} else {
			resultObj = fmt.Sprintf("Mock output for command: %s\nCommand executed successfully", termArgs.Input)
		}

	case tools.FileToolName:
		var fileArgs tools.FileAction
		if err := json.Unmarshal(args, &fileArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling file arguments: %w", err)
		}

		terminal.PrintMock("File operation:")
		terminal.PrintKeyValue("Operation", string(fileArgs.Action))
		terminal.PrintKeyValue("Path", fileArgs.Path)

		if fileArgs.Action == tools.ReadFile {
			resultObj = fmt.Sprintf("Mock content of file: %s\nThis is a sample content that would be read from the file.\nIt contains multiple lines to simulate a real file.", fileArgs.Path)
		} else {
			resultObj = fmt.Sprintf("file %s written successfully", fileArgs.Path)
		}

	case tools.BrowserToolName:
		var browserArgs tools.Browser
		if err := json.Unmarshal(args, &browserArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling browser arguments: %w", err)
		}

		terminal.PrintMock("Browser action:")
		terminal.PrintKeyValue("Action", string(browserArgs.Action))
		terminal.PrintKeyValue("URL", browserArgs.Url)

		switch browserArgs.Action {
		case tools.Markdown:
			resultObj = fmt.Sprintf("# Mock page for %s\n\n## Introduction\n\nThis is a mock page content that simulates what the real browser tool would return in markdown format.\n\n## Main Content\n\nHere is some example text that would appear on the page.\n\n* List item 1\n* List item 2\n* List item 3\n\n## Conclusion\n\nThis mock content is designed to look like real markdown content from a web page.", browserArgs.Url)
		case tools.HTML:
			resultObj = fmt.Sprintf("<!DOCTYPE html>\n<html>\n<head>\n  <title>Mock Page for %s</title>\n</head>\n<body>\n  <h1>Mock HTML Content</h1>\n  <p>This is a mock HTML page that simulates what the real browser tool would return.</p>\n  <ul>\n    <li>HTML Element 1</li>\n    <li>HTML Element 2</li>\n    <li>HTML Element 3</li>\n  </ul>\n</body>\n</html>", browserArgs.Url)
		case tools.Links:
			resultObj = fmt.Sprintf("Links list from URL '%s'\n[Homepage](https://example.com)\n[About Us](https://example.com/about)\n[Products](https://example.com/products)\n[Documentation](https://example.com/docs)\n[Contact](https://example.com/contact)", browserArgs.Url)
		}

	case tools.GoogleToolName:
		var searchArgs tools.SearchAction
		if err := json.Unmarshal(args, &searchArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling search arguments: %w", err)
		}

		terminal.PrintMock("Google search:")
		terminal.PrintKeyValue("Query", searchArgs.Query)
		terminal.PrintKeyValueFormat("Max results", "%d", searchArgs.MaxResults.Int())

		var builder strings.Builder
		for i := 1; i <= min(searchArgs.MaxResults.Int(), 5); i++ {
			builder.WriteString(fmt.Sprintf("# %d. Mock Google Result %d for '%s'\n\n", i, i, searchArgs.Query))
			builder.WriteString(fmt.Sprintf("## URL\nhttps://example.com/result%d\n\n", i))
			builder.WriteString(fmt.Sprintf("## Snippet\n\nThis is a detailed mock snippet for search result %d that matches your query '%s'. It contains relevant information that would be returned by the real Google search API.\n\n", i, searchArgs.Query))
		}
		resultObj = builder.String()

	case tools.DuckDuckGoToolName:
		var searchArgs tools.SearchAction
		if err := json.Unmarshal(args, &searchArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling search arguments: %w", err)
		}

		terminal.PrintMock("DuckDuckGo search:")
		terminal.PrintKeyValue("Query", searchArgs.Query)
		terminal.PrintKeyValueFormat("Max results", "%d", searchArgs.MaxResults.Int())

		var builder strings.Builder
		for i := 1; i <= min(searchArgs.MaxResults.Int(), 5); i++ {
			builder.WriteString(fmt.Sprintf("# %d. Mock DuckDuckGo Result %d for '%s'\n\n", i, i, searchArgs.Query))
			builder.WriteString(fmt.Sprintf("## URL\nhttps://example.com/duckduckgo/result%d\n\n", i))
			builder.WriteString(fmt.Sprintf("## Description\n\nThis is a detailed mock description for search result %d that matches your query '%s'. DuckDuckGo would provide this kind of anonymous search result.\n\n", i, searchArgs.Query))

			if i < min(searchArgs.MaxResults.Int(), 5) {
				builder.WriteString("---\n\n")
			}
		}
		resultObj = builder.String()

	case tools.TavilyToolName:
		var searchArgs tools.SearchAction
		if err := json.Unmarshal(args, &searchArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling search arguments: %w", err)
		}

		terminal.PrintMock("Tavily search:")
		terminal.PrintKeyValue("Query", searchArgs.Query)
		terminal.PrintKeyValueFormat("Max results", "%d", searchArgs.MaxResults.Int())

		var builder strings.Builder
		builder.WriteString("# Answer\n\n")
		builder.WriteString(fmt.Sprintf("This is a comprehensive answer to your query '%s' that would be generated by Tavily AI. It synthesizes information from multiple sources to provide you with the most relevant information.\n\n", searchArgs.Query))
		builder.WriteString("# Links\n\n")

		for i := 1; i <= min(searchArgs.MaxResults.Int(), 3); i++ {
			builder.WriteString(fmt.Sprintf("## %d. Mock Tavily Result %d\n\n", i, i))
			builder.WriteString(fmt.Sprintf("* URL https://example.com/tavily/result%d\n", i))
			builder.WriteString(fmt.Sprintf("* Match score %.3f\n\n", 0.95-float64(i-1)*0.1))
			builder.WriteString(fmt.Sprintf("### Short content\n\nHere is a brief summary of the content from this search result related to '%s'.\n\n", searchArgs.Query))
			builder.WriteString(fmt.Sprintf("### Content\n\nThis is the full detailed content that would be retrieved from the URL. It contains comprehensive information about '%s' that helps answer your query with specific facts and data points that would be relevant to your search.\n\n", searchArgs.Query))
		}

		resultObj = builder.String()

	case tools.TraversaalToolName:
		var searchArgs tools.SearchAction
		if err := json.Unmarshal(args, &searchArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling search arguments: %w", err)
		}

		terminal.PrintMock("Traversaal search:")
		terminal.PrintKeyValue("Query", searchArgs.Query)
		terminal.PrintKeyValueFormat("Max results", "%d", searchArgs.MaxResults.Int())

		var builder strings.Builder
		builder.WriteString("# Answer\n\n")
		builder.WriteString(fmt.Sprintf("Here is the Traversaal answer to your query '%s'. Traversaal provides concise answers based on web information with relevant links for further exploration.\n\n", searchArgs.Query))
		builder.WriteString("# Links\n\n")

		for i := 1; i <= min(searchArgs.MaxResults.Int(), 5); i++ {
			builder.WriteString(fmt.Sprintf("%d. https://example.com/traversaal/resource%d\n", i, i))
		}

		resultObj = builder.String()

	case tools.PerplexityToolName:
		var searchArgs tools.SearchAction
		if err := json.Unmarshal(args, &searchArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling search arguments: %w", err)
		}

		terminal.PrintMock("Perplexity search:")
		terminal.PrintKeyValue("Query", searchArgs.Query)
		terminal.PrintKeyValueFormat("Max results", "%d", searchArgs.MaxResults.Int())

		var builder strings.Builder
		builder.WriteString("# Answer\n\n")
		builder.WriteString(fmt.Sprintf("This is a detailed research report from Perplexity AI about '%s'. Perplexity provides comprehensive answers by synthesizing information from various sources and augmenting it with AI analysis.\n\n", searchArgs.Query))
		builder.WriteString("The query you've asked about requires examining multiple perspectives and sources. Based on recent information, here's a thorough analysis of the topic with key insights and developments.\n\n")
		builder.WriteString("First, it's important to understand the background of this subject. Several authoritative sources indicate that this is an evolving area with recent developments. The most current research suggests that...\n\n")

		builder.WriteString("\n\n# Citations\n\n")
		for i := 1; i <= min(searchArgs.MaxResults.Int(), 3); i++ {
			builder.WriteString(fmt.Sprintf("%d. https://example.com/perplexity/citation%d\n", i, i))
		}

		resultObj = builder.String()

	case tools.SearchToolName:
		var searchArgs tools.ComplexSearch
		if err := json.Unmarshal(args, &searchArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling complex search arguments: %w", err)
		}

		terminal.PrintMock("Complex search:")
		terminal.PrintKeyValue("Question", searchArgs.Question)

		resultObj = fmt.Sprintf("# Comprehensive Search Results for: '%s'\n\n## Summary\nThis is a comprehensive answer to your complex question based on multiple search engines and memory sources. The researcher team has compiled the most relevant information from various sources.\n\n## Key Findings\n1. Finding one: Important information related to your query\n2. Finding two: Additional context that helps answer your question\n3. Finding three: Specific details from technical documentation\n\n## Sources\n- Web search (Google, DuckDuckGo)\n- Technical documentation\n- Academic papers\n- Long-term memory results\n\n## Conclusion\nBased on all available information, here is the complete answer to your question with code examples, command samples, and specific technical details as requested.", searchArgs.Question)

	case tools.SearchResultToolName:
		var searchResultArgs tools.SearchResult
		if err := json.Unmarshal(args, &searchResultArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling search result arguments: %w", err)
		}

		terminal.PrintMock("Search result received:")
		terminal.PrintKeyValueFormat("Content length", "%d chars", len(searchResultArgs.Result))

		resultObj = map[string]any{
			"status":  "success",
			"message": "Search results processed and delivered successfully",
		}

	case tools.MemoristToolName:
		var memoristArgs tools.MemoristAction
		if err := json.Unmarshal(args, &memoristArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling memorist arguments: %w", err)
		}

		terminal.PrintMock("Memorist question:")
		terminal.PrintKeyValue("Question", memoristArgs.Question)

		if memoristArgs.TaskID != nil {
			terminal.PrintKeyValueFormat("Task ID", "%d", memoristArgs.TaskID.Int64())
		}
		if memoristArgs.SubtaskID != nil {
			terminal.PrintKeyValueFormat("Subtask ID", "%d", memoristArgs.SubtaskID.Int64())
		}

		resultObj = fmt.Sprintf("# Archivist Memory Results\n\n## Question\n%s\n\n## Retrieved Information\nThe archivist has searched through all past work and tasks and found the following relevant information:\n\n1. On [date], a similar task was performed with the following approach...\n2. The team previously encountered this issue and resolved it by...\n3. Related documentation was created during project [X] that explains...\n\n## Historical Context\nThis question relates to work that was done approximately [time period] ago, and involved the following components and techniques...\n\n## Recommended Next Steps\nBased on historical information, the most effective approach would be to...", memoristArgs.Question)

	case tools.MemoristResultToolName:
		var memoristResultArgs tools.MemoristResult
		if err := json.Unmarshal(args, &memoristResultArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling memorist result arguments: %w", err)
		}

		terminal.PrintMock("Memorist result received:")
		terminal.PrintKeyValueFormat("Content length", "%d chars", len(memoristResultArgs.Result))

		resultObj = map[string]any{
			"status":  "success",
			"message": "Memory search results processed and delivered successfully",
		}

	case tools.SearchInMemoryToolName:
		var searchMemoryArgs tools.SearchInMemoryAction
		if err := json.Unmarshal(args, &searchMemoryArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling search memory arguments: %w", err)
		}

		terminal.PrintMock("Search in memory:")
		terminal.PrintKeyValue("Question", searchMemoryArgs.Question)

		if searchMemoryArgs.TaskID != nil {
			terminal.PrintKeyValueFormat("Task ID filter", "%d", searchMemoryArgs.TaskID.Int64())
		}
		if searchMemoryArgs.SubtaskID != nil {
			terminal.PrintKeyValueFormat("Subtask ID filter", "%d", searchMemoryArgs.SubtaskID.Int64())
		}

		var builder strings.Builder
		builder.WriteString("# Match score 0.92\n\n")
		if searchMemoryArgs.TaskID != nil {
			builder.WriteString(fmt.Sprintf("# Task ID %d\n\n", searchMemoryArgs.TaskID.Int64()))
		}
		if searchMemoryArgs.SubtaskID != nil {
			builder.WriteString(fmt.Sprintf("# Subtask ID %d\n\n", searchMemoryArgs.SubtaskID.Int64()))
		}
		builder.WriteString("# Tool Name 'terminal'\n\n")
		builder.WriteString("# Tool Description\n\nCalls a terminal command in blocking mode with hard limit timeout 1200 seconds and optimum timeout 60 seconds\n\n")
		builder.WriteString("# Chunk\n\n")
		builder.WriteString(fmt.Sprintf("This is a memory chunk related to your question '%s'. It contains information about previous commands, outputs, and relevant context that was stored in the vector database.\n\n", searchMemoryArgs.Question))
		builder.WriteString("---------------------------\n")
		builder.WriteString("# Match score 0.85\n\n")
		builder.WriteString("# Tool Name 'file'\n\n")
		builder.WriteString("# Chunk\n\n")
		builder.WriteString("This is another memory chunk that provides additional context to your question. It contains information about file operations and relevant content changes.\n")
		builder.WriteString("---------------------------\n")

		resultObj = builder.String()

	case tools.SearchGuideToolName:
		var searchGuideArgs tools.SearchGuideAction
		if err := json.Unmarshal(args, &searchGuideArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling search guide arguments: %w", err)
		}

		terminal.PrintMock("Search guide:")
		terminal.PrintKeyValue("Question", searchGuideArgs.Question)
		terminal.PrintKeyValue("Guide type", searchGuideArgs.Type)

		if searchGuideArgs.Type == "pentest" {
			resultObj = fmt.Sprintf("# Original Guide Type: pentest\n\n# Original Guide Question\n\n%s\n\n## Penetration Testing Guide\n\nThis guide provides a step-by-step approach for conducting a penetration test on the target system.\n\n### 1. Reconnaissance\n- Gather information about the target using OSINT tools\n- Identify potential entry points and attack surfaces\n\n### 2. Scanning\n- Use tools like Nmap to scan for open ports and services\n- Identify vulnerabilities using automated scanners\n\n### 3. Exploitation\n- Attempt to exploit identified vulnerabilities\n- Document successful attack vectors\n\n### 4. Post-Exploitation\n- Maintain access and explore the system\n- Identify sensitive data and potential lateral movement paths\n\n### 5. Reporting\n- Document all findings with proof of concept\n- Provide remediation recommendations\n\n", searchGuideArgs.Question)
		} else if searchGuideArgs.Type == "install" {
			resultObj = fmt.Sprintf("# Original Guide Type: install\n\n# Original Guide Question\n\n%s\n\n## Installation Guide\n\n### Prerequisites\n- Operating System: Linux/macOS/Windows\n- Required dependencies: [list]\n\n### Installation Steps\n1. Download the software from the official repository\n   ```bash\n   git clone https://github.com/example/software.git\n   ```\n\n2. Navigate to the project directory\n   ```bash\n   cd software\n   ```\n\n3. Install dependencies\n   ```bash\n   npm install\n   ```\n\n4. Build the project\n   ```bash\n   npm run build\n   ```\n\n5. Verify installation\n   ```bash\n   npm test\n   ```\n\n### Troubleshooting\n- Common issue 1: [solution]\n- Common issue 2: [solution]\n\n", searchGuideArgs.Question)
		} else {
			resultObj = fmt.Sprintf("# Original Guide Type: %s\n\n# Original Guide Question\n\n%s\n\n## Guide Content\n\nThis is a comprehensive guide for the requested type '%s'. It contains detailed instructions, best practices, and examples tailored to your specific question.\n\n### Section 1: Getting Started\n[Detailed content would be here]\n\n### Section 2: Main Procedures\n[Step-by-step instructions would be here]\n\n### Section 3: Advanced Techniques\n[Advanced content would be here]\n\n### Section 4: Troubleshooting\n[Common issues and solutions would be here]\n\n", searchGuideArgs.Type, searchGuideArgs.Question, searchGuideArgs.Type)
		}

	case tools.StoreGuideToolName:
		var storeGuideArgs tools.StoreGuideAction
		if err := json.Unmarshal(args, &storeGuideArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling store guide arguments: %w", err)
		}

		terminal.PrintMock("Store guide:")
		terminal.PrintKeyValue("Type", storeGuideArgs.Type)
		terminal.PrintKeyValueFormat("Guide length", "%d chars", len(storeGuideArgs.Guide))
		terminal.PrintKeyValue("Guide question", storeGuideArgs.Question)

		resultObj = "guide stored successfully"

	case tools.SearchAnswerToolName:
		var searchAnswerArgs tools.SearchAnswerAction
		if err := json.Unmarshal(args, &searchAnswerArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling search answer arguments: %w", err)
		}

		terminal.PrintMock("Search answer:")
		terminal.PrintKeyValue("Question", searchAnswerArgs.Question)
		terminal.PrintKeyValue("Answer type", searchAnswerArgs.Type)

		if searchAnswerArgs.Type == "vulnerability" {
			resultObj = fmt.Sprintf("# Original Answer Type: vulnerability\n\n# Original Search Question\n\n%s\n\n## Vulnerability Details\n\n### CVE-2023-12345\n\n**Severity**: High\n\n**Affected Systems**: Linux servers running Apache 2.4.x before 2.4.56\n\n**Description**:\nA buffer overflow vulnerability in Apache HTTP Server allows attackers to execute arbitrary code via a crafted request.\n\n**Exploitation**:\nAttackers can send a specially crafted HTTP request that triggers the buffer overflow, leading to remote code execution with the privileges of the web server process.\n\n**Remediation**:\n- Update Apache HTTP Server to version 2.4.56 or later\n- Apply the security patch provided by the vendor\n- Implement network filtering to block malicious requests\n\n**References**:\n- https://example.com/cve-2023-12345\n- https://example.com/apache-advisory\n", searchAnswerArgs.Question)
		} else {
			resultObj = fmt.Sprintf("# Original Answer Type: %s\n\n# Original Search Question\n\n%s\n\n## Comprehensive Answer\n\nThis is a detailed answer to your question related to the type '%s'. The answer provides comprehensive information, examples, and best practices.\n\n### Key Points\n1. First important point about your question\n2. Second important aspect to consider\n3. Technical details relevant to your inquiry\n\n### Examples\n```\nExample code or configuration would be here\n```\n\n### Additional Resources\n- Resource 1: [description]\n- Resource 2: [description]\n\n", searchAnswerArgs.Type, searchAnswerArgs.Question, searchAnswerArgs.Type)
		}

	case tools.StoreAnswerToolName:
		var storeAnswerArgs tools.StoreAnswerAction
		if err := json.Unmarshal(args, &storeAnswerArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling store answer arguments: %w", err)
		}

		terminal.PrintMock("Store answer:")
		terminal.PrintKeyValue("Type", storeAnswerArgs.Type)
		terminal.PrintKeyValueFormat("Answer length", "%d chars", len(storeAnswerArgs.Answer))
		terminal.PrintKeyValue("Question", storeAnswerArgs.Question)

		resultObj = "answer for question stored successfully"

	case tools.SearchCodeToolName:
		var searchCodeArgs tools.SearchCodeAction
		if err := json.Unmarshal(args, &searchCodeArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling search code arguments: %w", err)
		}

		terminal.PrintMock("Search code:")
		terminal.PrintKeyValue("Question", searchCodeArgs.Question)
		terminal.PrintKeyValue("Language", searchCodeArgs.Lang)

		var mockCode string
		if searchCodeArgs.Lang == "python" {
			mockCode = "def example_function(param1, param2='default'):\n    \"\"\"This is an example Python function that demonstrates a pattern.\n    \n    Args:\n        param1: The first parameter\n        param2: The second parameter with default value\n        \n    Returns:\n        The processed result\n    \"\"\"\n    result = {}\n    \n    # Process the parameters\n    if param1 is not None:\n        result['param1'] = param1\n    \n    # Additional processing\n    if param2 != 'default':\n        result['param2'] = param2\n    \n    return result\n\n# Example usage\nif __name__ == '__main__':\n    output = example_function('test', 'custom')\n    print(output)"
		} else if searchCodeArgs.Lang == "javascript" || searchCodeArgs.Lang == "js" {
			mockCode = "/**\n * Example JavaScript function that demonstrates a pattern\n * @param {Object} options - Configuration options\n * @param {string} options.name - The name parameter\n * @param {number} [options.count=1] - Optional count parameter\n * @returns {Object} The processed result\n */\nfunction exampleFunction(options) {\n  const { name, count = 1 } = options;\n  \n  // Input validation\n  if (!name) {\n    throw new Error('Name is required');\n  }\n  \n  // Process the data\n  const result = {\n    processedName: name.toUpperCase(),\n    repeatedCount: Array(count).fill(name).join(', ')\n  };\n  \n  return result;\n}\n\n// Example usage\nconst output = exampleFunction({ name: 'test', count: 3 });\nconsole.log(output);"
		} else {
			mockCode = fmt.Sprintf("// Example code in %s language\n// This is a mock code snippet that would be returned from the vector database\n\n// Main function definition\nfunction exampleFunction(param) {\n  // Initialization\n  const result = [];\n  \n  // Processing logic\n  for (let i = 0; i < param.length; i++) {\n    result.push(processItem(param[i]));\n  }\n  \n  return result;\n}\n\n// Helper function\nfunction processItem(item) {\n  return item.transform();\n}", searchCodeArgs.Lang)
		}

		resultObj = fmt.Sprintf("# Original Code Question\n\n%s\n\n# Original Code Description\n\nThis code sample demonstrates the implementation pattern for handling the specific scenario you asked about. It includes proper error handling, input validation, and follows best practices for %s.\n\n```%s\n%s\n```\n\n", searchCodeArgs.Question, searchCodeArgs.Lang, searchCodeArgs.Lang, mockCode)

	case tools.StoreCodeToolName:
		var storeCodeArgs tools.StoreCodeAction
		if err := json.Unmarshal(args, &storeCodeArgs); err != nil {
			return "", fmt.Errorf("error unmarshaling store code arguments: %w", err)
		}

		terminal.PrintMock("Store code:")
		terminal.PrintKeyValue("Language", storeCodeArgs.Lang)
		terminal.PrintKeyValueFormat("Code length", "%d chars", len(storeCodeArgs.Code))
		terminal.PrintKeyValue("Question", storeCodeArgs.Question)
		terminal.PrintKeyValue("Description", storeCodeArgs.Description)

		resultObj = "code sample stored successfully"

	default:
		terminal.PrintMock("Generic mock response:")
		terminal.PrintKeyValue("Function", funcName)
		resultObj = map[string]any{
			"status":  "success",
			"message": fmt.Sprintf("Mock result for function: %s", funcName),
			"data":    "This is a generic mock response for testing purposes",
		}
	}

	var resultJSON string

	// Handle string results directly
	if strResult, ok := resultObj.(string); ok {
		resultJSON = strResult
	} else {
		// Marshal object results
		jsonBytes, err := json.Marshal(resultObj)
		if err != nil {
			return "", fmt.Errorf("error marshaling mock result: %w", err)
		}
		resultJSON = string(jsonBytes)
	}

	return resultJSON, nil
}
