package main

// mhr-team-match: Match CSV team names against home_teams labels in mhr_locations table.
// Usage: mhr-team-match <input-csv-file> [output-csv-file]
// Input CSV must have a 'name' column (second column) containing team names.
// Output CSV will have two additional columns: mhr_id, matched_team_name.
// Matching is case-insensitive substring match: if label is contained in team name.

import (
	"calendar-scrapper/config"
	"calendar-scrapper/pkg/repository"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: mhr-team-match <input-csv-file> [output-csv-file]")
	}
	inputFile := os.Args[1]
	outputFile := inputFile
	if strings.HasSuffix(outputFile, ".csv") {
		outputFile = strings.TrimSuffix(outputFile, ".csv") + "_matched.csv"
	} else {
		outputFile = outputFile + "_matched.csv"
	}
	if len(os.Args) >= 3 {
		outputFile = os.Args[2]
	}

	log.Printf("Matching teams from %s, writing to %s", inputFile, outputFile)

	// Load config and connect to database
	config.Init("config", ".")
	cfg := config.MustReadConfig()
	repo := repository.NewRepository(cfg)

	// Query all mhr_locations with home_teams
	type dbRow struct {
		MhrID     int32
		HomeTeams string
	}
	var rows []dbRow
	err := repo.DB.Raw(`SELECT mhr_id, home_teams FROM mhr_locations WHERE home_teams IS NOT NULL AND home_teams != ''`).Scan(&rows).Error
	if err != nil {
		log.Fatal("Failed to query mhr_locations:", err)
	}
	log.Printf("Loaded %d mhr_locations with home_teams", len(rows))

	// Build label index: slice of {MhrID, Label, LowerLabel}
	type labelEntry struct {
		MhrID      int32
		Label      string
		LowerLabel string
	}
	var labels []labelEntry
	for _, row := range rows {
		if row.HomeTeams == "" {
			continue
		}
		var homeTeams []map[string]string
		if err := json.Unmarshal([]byte(row.HomeTeams), &homeTeams); err != nil {
			log.Printf("Warning: failed to unmarshal home_teams JSON for mhr_id %d: %v", row.MhrID, err)
			continue
		}
		for _, ht := range homeTeams {
			if label, ok := ht["label"]; ok && label != "" {
				trimmed := strings.TrimSpace(label)
				if trimmed == "" {
					continue
				}
				labels = append(labels, labelEntry{
					MhrID:      row.MhrID,
					Label:      trimmed,
					LowerLabel: strings.ToLower(trimmed),
				})
			}
		}
	}
	log.Printf("Extracted %d label entries from home_teams", len(labels))

	// Open input CSV
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatal("Failed to open input file:", err)
	}
	defer inFile.Close()

	csvReader := csv.NewReader(inFile)
	headers, err := csvReader.Read()
	if err != nil {
		log.Fatal("Failed to read CSV header:", err)
	}

	// Determine name column index (second column per spec, zero-based index 1)
	nameColIdx := 1
	if len(headers) <= nameColIdx {
		log.Fatal("CSV does not have enough columns (need at least 2)")
	}
	if headers[nameColIdx] != "name" {
		log.Printf("Warning: second column header is '%s', expected 'name'", headers[nameColIdx])
	}

	// Prepare output CSV
	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatal("Failed to create output file:", err)
	}
	defer outFile.Close()

	csvWriter := csv.NewWriter(outFile)
	defer csvWriter.Flush()

	// Write new headers: original headers + mhr_id + matched_team_name
	newHeaders := append(headers, "mhr_id", "matched_team_name")
	if err := csvWriter.Write(newHeaders); err != nil {
		log.Fatal("Failed to write output header:", err)
	}

	// Process rows
	rowNum := 0
	matchedCount := 0
	for {
		record, err := csvReader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Warning: failed to read CSV row %d: %v", rowNum, err)
			continue
		}
		rowNum++

		// Ensure record length matches headers (pad if necessary)
		if len(record) < len(headers) {
			// Pad with empty strings
			for i := len(record); i < len(headers); i++ {
				record = append(record, "")
			}
		}

		teamName := strings.TrimSpace(record[nameColIdx])
		var mhrID string
		var matchedLabel string

		if teamName != "" {
			lowerName := strings.ToLower(teamName)
			// Find first matching label
			for _, entry := range labels {
				if strings.Contains(lowerName, entry.LowerLabel) {
					mhrID = fmt.Sprintf("%d", entry.MhrID)
					matchedLabel = entry.Label
					matchedCount++
					break
				}
			}
		}

		// Append new columns
		outputRecord := append(record, mhrID, matchedLabel)
		if err := csvWriter.Write(outputRecord); err != nil {
			log.Fatal("Failed to write CSV row:", err)
		}
	}

	log.Printf("Processed %d rows, matched %d teams", rowNum, matchedCount)
}
