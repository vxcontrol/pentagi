package main

import (
	"database/sql"
	"fmt"
	"strings"

	"pentagi/pkg/terminal"
)

// info displays statistics about the embedding database
func (t *Tester) info() error {
	terminal.PrintHeader("Database Information:")
	terminal.PrintThinSeparator()

	// Get total document count
	var docCount int
	err := t.conn.QueryRow(t.ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s", t.embeddingTableName)).Scan(&docCount)
	if err != nil {
		return fmt.Errorf("failed to get document count: %w", err)
	}
	terminal.PrintKeyValueFormat("Total documents", "%d", docCount)

	if docCount == 0 {
		terminal.Info("No documents in the database.")
		return nil
	}

	// Get average document size
	var avgSize float64
	err = t.conn.QueryRow(t.ctx,
		fmt.Sprintf("SELECT AVG(LENGTH(document)) FROM %s", t.embeddingTableName)).Scan(&avgSize)
	if err != nil {
		return fmt.Errorf("failed to get average document size: %w", err)
	}
	terminal.PrintKeyValueFormat("Average document size", "%.2f bytes", avgSize)

	// Get total document size
	var totalSize int64
	err = t.conn.QueryRow(t.ctx,
		fmt.Sprintf("SELECT SUM(LENGTH(document)) FROM %s", t.embeddingTableName)).Scan(&totalSize)
	if err != nil {
		return fmt.Errorf("failed to get total document size: %w", err)
	}
	terminal.PrintKeyValue("Total document size", formatSize(totalSize))

	// Get document type distribution
	terminal.PrintHeader("\nDocument Type Distribution:")
	rows, err := t.conn.Query(t.ctx,
		fmt.Sprintf("SELECT cmetadata->>'doc_type' as type, COUNT(*) FROM %s GROUP BY type ORDER BY COUNT(*) DESC",
			t.embeddingTableName))
	if err != nil {
		return fmt.Errorf("failed to get document type distribution: %w", err)
	}
	defer rows.Close()

	printTableHeader("Type", "Count")

	for rows.Next() {
		var docType sql.NullString
		var count int
		if err := rows.Scan(&docType, &count); err != nil {
			return fmt.Errorf("failed to scan document type row: %w", err)
		}
		typeStr := "unknown"
		if docType.Valid {
			typeStr = docType.String
		}
		printTableRow(typeStr, count)
	}

	// Get flow_id distribution
	terminal.PrintHeader("\nFlow ID Distribution:")
	rows, err = t.conn.Query(t.ctx,
		fmt.Sprintf("SELECT cmetadata->>'flow_id' as flow_id, COUNT(*) FROM %s GROUP BY flow_id ORDER BY COUNT(*) DESC",
			t.embeddingTableName))
	if err != nil {
		return fmt.Errorf("failed to get flow ID distribution: %w", err)
	}
	defer rows.Close()

	printTableHeader("Flow ID", "Count")

	for rows.Next() {
		var flowID sql.NullString
		var count int
		if err := rows.Scan(&flowID, &count); err != nil {
			return fmt.Errorf("failed to scan flow ID row: %w", err)
		}
		flowStr := "unknown"
		if flowID.Valid {
			flowStr = flowID.String
		}
		printTableRow(flowStr, count)
	}

	// Get guide_type distribution for doc_type = 'guide'
	terminal.PrintHeader("\nGuide Type Distribution (for doc_type = 'guide'):")
	rows, err = t.conn.Query(t.ctx,
		fmt.Sprintf("SELECT cmetadata->>'guide_type' as guide_type, COUNT(*) FROM %s "+
			"WHERE cmetadata->>'doc_type' = 'guide' GROUP BY guide_type ORDER BY COUNT(*) DESC",
			t.embeddingTableName))
	if err != nil {
		return fmt.Errorf("failed to get guide type distribution: %w", err)
	}
	defer rows.Close()

	printTableHeader("Guide Type", "Count")

	hasRows := false
	for rows.Next() {
		hasRows = true
		var guideType sql.NullString
		var count int
		if err := rows.Scan(&guideType, &count); err != nil {
			return fmt.Errorf("failed to scan guide type row: %w", err)
		}
		typeStr := "unknown"
		if guideType.Valid {
			typeStr = guideType.String
		}
		printTableRow(typeStr, count)
	}
	if !hasRows {
		terminal.Info("No guide documents found.")
	}

	// Get code_lang distribution for doc_type = 'code'
	terminal.PrintHeader("\nCode Language Distribution (for doc_type = 'code'):")
	rows, err = t.conn.Query(t.ctx,
		fmt.Sprintf("SELECT cmetadata->>'code_lang' as code_lang, COUNT(*) FROM %s "+
			"WHERE cmetadata->>'doc_type' = 'code' GROUP BY code_lang ORDER BY COUNT(*) DESC",
			t.embeddingTableName))
	if err != nil {
		return fmt.Errorf("failed to get code language distribution: %w", err)
	}
	defer rows.Close()

	printTableHeader("Code Language", "Count")

	hasRows = false
	for rows.Next() {
		hasRows = true
		var codeLang sql.NullString
		var count int
		if err := rows.Scan(&codeLang, &count); err != nil {
			return fmt.Errorf("failed to scan code language row: %w", err)
		}
		langStr := "unknown"
		if codeLang.Valid {
			langStr = codeLang.String
		}
		printTableRow(langStr, count)
	}
	if !hasRows {
		terminal.Info("No code documents found.")
	}

	// Get answer_type distribution for doc_type = 'answer'
	terminal.PrintHeader("\nAnswer Type Distribution (for doc_type = 'answer'):")
	rows, err = t.conn.Query(t.ctx,
		fmt.Sprintf("SELECT cmetadata->>'answer_type' as answer_type, COUNT(*) FROM %s "+
			"WHERE cmetadata->>'doc_type' = 'answer' GROUP BY answer_type ORDER BY COUNT(*) DESC",
			t.embeddingTableName))
	if err != nil {
		return fmt.Errorf("failed to get answer type distribution: %w", err)
	}
	defer rows.Close()

	printTableHeader("Answer Type", "Count")

	hasRows = false
	for rows.Next() {
		hasRows = true
		var answerType sql.NullString
		var count int
		if err := rows.Scan(&answerType, &count); err != nil {
			return fmt.Errorf("failed to scan answer type row: %w", err)
		}
		typeStr := "unknown"
		if answerType.Valid {
			typeStr = answerType.String
		}
		printTableRow(typeStr, count)
	}
	if !hasRows {
		terminal.Info("No answer documents found.")
	}

	return nil
}

// printTableHeader prints a formatted table header row
func printTableHeader(column1, column2 string) {
	fmt.Printf("%-20s | %s\n", column1, column2)
	fmt.Printf("%-20s-+-%s\n", strings.Repeat("-", 20), strings.Repeat("-", 10))
}

// printTableRow prints a table row with data
func printTableRow(value string, count int) {
	fmt.Printf("%-20s | %d\n", value, count)
}
