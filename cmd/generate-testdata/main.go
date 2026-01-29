// Package main provides a command-line tool to generate test data files for performance benchmarks.
// This is a separate executable and is not included when importing the typedb package.
// Run with: go run ./cmd/generate-testdata
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Row counts to generate
var rowCounts = []int{10, 100, 1000, 10000, 100000, 1000000}

func main() {
	// Create testdata directories
	simpleDir := "testdata/simple"
	complexDir := "testdata/complex"

	if err := os.MkdirAll(simpleDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create simple directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(complexDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create complex directory: %v\n", err)
		os.Exit(1)
	}

	// Generate simple dataset files
	fmt.Println("Generating simple dataset files...")
	for _, count := range rowCounts {
		if count >= 100000 {
			// Split large files into chunks
			chunkSize := 10000
			if err := generateAndSaveSimpleDataSplit(count, simpleDir, chunkSize); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to generate split files for %d rows: %v\n", count, err)
				os.Exit(1)
			}
			fmt.Printf("  Generated split files for %d rows (%d files)\n", count, (count+chunkSize-1)/chunkSize)
		} else {
			filename := fmt.Sprintf("%s/%drows.json", simpleDir, count)
			if err := generateAndSaveSimpleData(count, filename); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to generate %s: %v\n", filename, err)
				os.Exit(1)
			}
			fmt.Printf("  Generated %s\n", filename)
		}
	}

	// Generate complex dataset files
	fmt.Println("Generating complex dataset files...")
	for _, count := range rowCounts {
		if count >= 100000 {
			// Split large files into chunks
			chunkSize := 10000
			if err := generateAndSaveComplexDataSplit(count, complexDir, chunkSize); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to generate split files for %d rows: %v\n", count, err)
				os.Exit(1)
			}
			fmt.Printf("  Generated split files for %d rows (%d files)\n", count, (count+chunkSize-1)/chunkSize)
		} else {
			filename := fmt.Sprintf("%s/%drows.json", complexDir, count)
			if err := generateAndSaveComplexData(count, filename); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to generate %s: %v\n", filename, err)
				os.Exit(1)
			}
			fmt.Printf("  Generated %s\n", filename)
		}
	}

	fmt.Println("All test data files generated successfully!")
}

func generateAndSaveSimpleData(rowCount int, filename string) error {
	data := generateSimpleData(rowCount)
	return saveJSONFile(filename, data)
}

func generateAndSaveComplexData(rowCount int, filename string) error {
	data := generateComplexData(rowCount)
	return saveJSONFile(filename, data)
}

func generateAndSaveSimpleDataSplit(totalRows int, dir string, chunkSize int) error {
	data := generateSimpleData(totalRows)
	return saveJSONFileSplit(totalRows, dir, chunkSize, data)
}

func generateAndSaveComplexDataSplit(totalRows int, dir string, chunkSize int) error {
	data := generateComplexData(totalRows)
	return saveJSONFileSplit(totalRows, dir, chunkSize, data)
}

func saveJSONFileSplit(totalRows int, dir string, chunkSize int, data []map[string]any) error {
	numChunks := (totalRows + chunkSize - 1) / chunkSize
	
	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(data) {
			end = len(data)
		}
		
		chunk := data[start:end]
		filename := fmt.Sprintf("%s/%drows_%03d.json", dir, totalRows, i+1)
		
		if err := saveJSONFile(filename, chunk); err != nil {
			return fmt.Errorf("failed to save chunk %d: %w", i+1, err)
		}
	}
	
	return nil
}

func saveJSONFile(filename string, data []map[string]any) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func generateSimpleData(rowCount int) []map[string]any {
	rows := make([]map[string]any, rowCount)
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < rowCount; i++ {
		rows[i] = map[string]any{
			"id":         int64(i + 1),
			"name":       fmt.Sprintf("User %d", i+1),
			"email":      fmt.Sprintf("user%d@example.com", i+1),
			"created_at": baseTime.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			"updated_at": baseTime.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			"active":     i%2 == 0,
			"age":        20 + (i % 50),
			"score":      50.0 + float64(i%50),
		}
	}
	return rows
}

func generateComplexData(rowCount int) []map[string]any {
	rows := make([]map[string]any, rowCount)
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	cities := []string{"San Francisco", "New York", "Los Angeles"}
	states := []string{"CA", "NY", "CA"}
	statuses := []string{"active", "inactive", "pending"}
	statusCodes := []string{"ACTIVE", "INACTIVE", "PENDING"}
	themes := []string{"light", "dark"}
	languages := []string{"en", "es", "fr"}
	timezones := []string{"UTC", "PST", "EST"}
	currencies := []string{"USD", "EUR", "GBP"}
	sources := []string{"web", "mobile", "api"}

	for i := 0; i < rowCount; i++ {
		// Generate nested address
		address := map[string]any{
			"street":   fmt.Sprintf("%d Main St", 100+i),
			"city":     cities[i%len(cities)],
			"state":    states[i%len(states)],
			"zip_code": fmt.Sprintf("%05d", 10000+i),
		}

		// Generate roles array
		roles := []map[string]any{
			{
				"id":          int64(1),
				"name":        "user",
				"permissions": []string{"read"},
				"granted_at":  baseTime.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			},
		}
		if i%3 == 0 {
			roles = append(roles, map[string]any{
				"id":          int64(2),
				"name":        "admin",
				"permissions": []string{"read", "write", "delete"},
				"granted_at":  baseTime.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			})
		}

		// Generate settings
		settings := map[string]any{
			"theme": themes[i%len(themes)],
			"notifications": map[string]bool{
				"email": i%2 == 0,
				"sms":   i%3 == 0,
				"push":  i%2 == 1,
			},
			"preferences": map[string]any{
				"auto_save":     i%2 == 0,
				"show_tooltips": i%3 == 0,
			},
		}

		// Serialize nested structs as JSON strings (as typedb expects from database)
		addressJSON, _ := json.Marshal(address)
		rolesJSON, _ := json.Marshal(roles)
		settingsJSON, _ := json.Marshal(settings)

		row := map[string]any{
			"id":         int64(i + 1),
			"name":       fmt.Sprintf("User %d", i+1),
			"email":      fmt.Sprintf("user%d@example.com", i+1),
			"created_at": baseTime.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			"updated_at": baseTime.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			"active":     i%2 == 0,
			"age":        20 + (i % 50),
			"score":      50.0 + float64(i%50),
			"metadata": map[string]any{
				"source":   sources[i%len(sources)],
				"campaign": fmt.Sprintf("campaign-%d", i%10),
				"referrer": fmt.Sprintf("referrer-%d.com", i%5),
			},
			"tags": []string{
				[]string{"premium", "verified"}[i%2],
				[]string{"early-adopter", "beta"}[i%2],
			},
			"preferences": map[string]string{
				"language": languages[i%len(languages)],
				"timezone": timezones[i%len(timezones)],
				"currency": currencies[i%len(currencies)],
			},
			"address":     string(addressJSON),  // JSON string for struct field
			"roles":       string(rolesJSON),    // JSON string for struct array field
			"settings":    string(settingsJSON), // JSON string for struct pointer field
			"last_login":  baseTime.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			"balance":     1000.0 + float64(i%1000),
			"status":      statuses[i%len(statuses)],
			"notes":       fmt.Sprintf("Notes for user %d", i+1),
			"user_id":     int64(100 + i),
			"status_code": statusCodes[i%len(statusCodes)],
		}

		// Make some fields nullable (nil) for some rows
		if i%5 == 0 {
			row["last_login"] = nil
		}
		if i%7 == 0 {
			row["balance"] = nil
		}
		if i%11 == 0 {
			row["settings"] = nil
		}

		rows[i] = row
	}
	return rows
}
