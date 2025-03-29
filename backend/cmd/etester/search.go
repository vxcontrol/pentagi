package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"pentagi/pkg/terminal"

	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

// SearchOptions represents the options for vector search
type SearchOptions struct {
	Query      string
	DocType    string
	FlowID     int64
	AnswerType string
	GuideType  string
	Limit      int
	Threshold  float32
}

// Validates and fills in default values for search options
func validateSearchOptions(opts *SearchOptions) error {
	// Query is required
	if opts.Query == "" {
		return fmt.Errorf("query parameter is required")
	}

	// Validate doc_type if provided
	if opts.DocType != "" {
		validDocTypes := map[string]bool{
			"answer": true,
			"memory": true,
			"guide":  true,
			"code":   true,
		}
		if !validDocTypes[opts.DocType] {
			return fmt.Errorf("invalid doc_type: %s. Valid values are: answer, memory, guide, code", opts.DocType)
		}
	}

	// Validate flow_id if provided
	if opts.FlowID < 0 {
		return fmt.Errorf("flow_id must be a positive number")
	}

	// Validate answer_type if provided
	if opts.AnswerType != "" {
		validAnswerTypes := map[string]bool{
			"guide":         true,
			"vulnerability": true,
			"code":          true,
			"tool":          true,
			"other":         true,
		}
		if !validAnswerTypes[opts.AnswerType] {
			return fmt.Errorf("invalid answer_type: %s. Valid values are: guide, vulnerability, code, tool, other", opts.AnswerType)
		}
	}

	// Validate guide_type if provided
	if opts.GuideType != "" {
		validGuideTypes := map[string]bool{
			"install":     true,
			"configure":   true,
			"use":         true,
			"pentest":     true,
			"development": true,
			"other":       true,
		}
		if !validGuideTypes[opts.GuideType] {
			return fmt.Errorf("invalid guide_type: %s. Valid values are: install, configure, use, pentest, development, other", opts.GuideType)
		}
	}

	// Validate limit
	if opts.Limit <= 0 {
		opts.Limit = 3 // Default limit
	}

	// Validate threshold
	if opts.Threshold <= 0 || opts.Threshold > 1 {
		opts.Threshold = 0.7 // Default threshold
	}

	return nil
}

// ParseSearchArgs parses command line arguments specific for search
func parseSearchArgs(args []string) (*SearchOptions, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no arguments provided")
	}

	opts := &SearchOptions{}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if !strings.HasPrefix(arg, "-") {
			continue
		}

		paramName := strings.TrimPrefix(arg, "-")

		if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
			return nil, fmt.Errorf("missing value for parameter: %s", paramName)
		}

		paramValue := args[i+1]
		i++

		switch paramName {
		case "query":
			opts.Query = paramValue
		case "doc_type":
			opts.DocType = paramValue
		case "flow_id":
			flowID, err := strconv.ParseInt(paramValue, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid flow_id value: %v", err)
			}
			opts.FlowID = flowID
		case "answer_type":
			opts.AnswerType = paramValue
		case "guide_type":
			opts.GuideType = paramValue
		case "limit":
			limit, err := strconv.Atoi(paramValue)
			if err != nil {
				return nil, fmt.Errorf("invalid limit value: %v", err)
			}
			opts.Limit = limit
		case "threshold":
			threshold, err := strconv.ParseFloat(paramValue, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid threshold value: %v", err)
			}
			opts.Threshold = float32(threshold)
		default:
			return nil, fmt.Errorf("unknown parameter: %s", paramName)
		}
	}

	if err := validateSearchOptions(opts); err != nil {
		return nil, err
	}

	return opts, nil
}

// search performs vector search in the embedding database
func (t *Tester) search(args []string) error {
	// Display usage if no arguments provided
	if len(args) == 0 {
		printSearchUsage()
		return nil
	}

	// Parse search options
	opts, err := parseSearchArgs(args)
	if err != nil {
		terminal.Error("Error parsing search arguments: %v", err)
		printSearchUsage()
		return nil
	}

	// Create pgvector store if needed for search
	store, err := t.createVectorStore()
	if err != nil {
		return fmt.Errorf("failed to create vector store: %w", err)
	}

	// Prepare filters
	filters := make(map[string]any)
	if opts.DocType != "" {
		filters["doc_type"] = opts.DocType
	}
	if opts.FlowID > 0 {
		filters["flow_id"] = strconv.FormatInt(opts.FlowID, 10)
	}
	if opts.AnswerType != "" {
		filters["answer_type"] = opts.AnswerType
	}
	if opts.GuideType != "" {
		filters["guide_type"] = opts.GuideType
	}

	// Prepare search options
	searchOpts := []vectorstores.Option{
		vectorstores.WithScoreThreshold(opts.Threshold),
	}

	if len(filters) > 0 {
		searchOpts = append(searchOpts, vectorstores.WithFilters(filters))
	}

	// Perform the search
	terminal.Info("Searching for: %s", opts.Query)
	terminal.Info("Threshold: %.2f, Limit: %d", opts.Threshold, opts.Limit)
	if len(filters) > 0 {
		terminal.Info("Filters: %v", filters)
	}

	docs, err := store.SimilaritySearch(
		t.ctx,
		opts.Query,
		opts.Limit,
		searchOpts...,
	)

	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Display results
	if len(docs) == 0 {
		terminal.Info("No matching documents found.")
		return nil
	}

	terminal.Success("Found %d matching documents:", len(docs))
	terminal.PrintThinSeparator()

	for i, doc := range docs {
		terminal.PrintHeader(fmt.Sprintf("Result #%d (similarity score: %.4f)", i+1, doc.Score))

		// Print metadata
		terminal.Info("Metadata:")
		keys := []string{}
		for k := range doc.Metadata {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			terminal.PrintKeyValueFormat(fmt.Sprintf("%-12s  ", k), "%v", doc.Metadata[k])
		}

		// Print content with markdown rendering
		terminal.PrintThinSeparator()
		terminal.PrintResult(doc.PageContent)
		terminal.PrintThickSeparator()
	}

	return nil
}

// createVectorStore creates a pgvector store instance using the current connection and embedder
func (t *Tester) createVectorStore() (*pgvector.Store, error) {
	// Create pgvector store
	store, err := pgvector.New(
		t.ctx,
		pgvector.WithConn(t.conn),
		pgvector.WithEmbedder(t.embedder),
		pgvector.WithCollectionName("langchain"),
		pgvector.WithEmbeddingTableName(t.embeddingTableName),
		pgvector.WithCollectionTableName(t.collectionTableName),
	)

	if err != nil {
		return nil, err
	}

	return &store, nil
}

// printSearchUsage prints the usage information for the search command
func printSearchUsage() {
	terminal.PrintHeader("Search Command Usage:")
	terminal.Info("Performs vector search in the embedding database")
	terminal.Info("\nSyntax:")
	terminal.Info("  ./etester search [OPTIONS]")
	terminal.Info("\nOptions:")
	terminal.PrintKeyValue("  -query STRING", "Search query text (required)")
	terminal.PrintKeyValue("  -doc_type STRING", "Filter by document type (answer, memory, guide, code)")
	terminal.PrintKeyValue("  -flow_id NUMBER", "Filter by flow ID (positive number)")
	terminal.PrintKeyValue("  -answer_type STRING", "Filter by answer type (guide, vulnerability, code, tool, other)")
	terminal.PrintKeyValue("  -guide_type STRING", "Filter by guide type (install, configure, use, pentest, development, other)")
	terminal.PrintKeyValue("  -limit NUMBER", "Maximum number of results (default: 3)")
	terminal.PrintKeyValue("  -threshold NUMBER", "Similarity threshold (0.0-1.0, default: 0.7)")
	terminal.Info("\nExamples:")
	terminal.Info("  ./etester search -query \"How to install PostgreSQL\" -limit 5")
	terminal.Info("  ./etester search -query \"Security vulnerability\" -doc_type guide -threshold 0.8")
	terminal.Info("  ./etester search -query \"Code examples\" -doc_type code -flow_id 42")
}
