package typedb

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"
)

var (
	// Simple data sets (basic structs, 5-10 fields)
	simpleData10   []map[string]any
	simpleData100  []map[string]any
	simpleData1k   []map[string]any
	simpleData10k  []map[string]any
	simpleData100k []map[string]any
	simpleData1M   []map[string]any

	// Complex data sets (15-20 fields, JSONB, arrays, nested types)
	complexData10   []map[string]any
	complexData100  []map[string]any
	complexData1k   []map[string]any
	complexData10k  []map[string]any
	complexData100k []map[string]any
	complexData1M   []map[string]any
)

func init() {
	// Load simple test data files into memory once
	simpleData10 = loadTestData("testdata/simple/10rows.json")
	simpleData100 = loadTestData("testdata/simple/100rows.json")
	simpleData1k = loadTestData("testdata/simple/1000rows.json")
	simpleData10k = loadTestData("testdata/simple/10000rows.json")
	simpleData100k = loadTestData("testdata/simple/100000rows.json")
	simpleData1M = loadTestData("testdata/simple/1000000rows.json")

	// Load complex test data files into memory once
	complexData10 = loadTestData("testdata/complex/10rows.json")
	complexData100 = loadTestData("testdata/complex/100rows.json")
	complexData1k = loadTestData("testdata/complex/1000rows.json")
	complexData10k = loadTestData("testdata/complex/10000rows.json")
	complexData100k = loadTestData("testdata/complex/100000rows.json")
	complexData1M = loadTestData("testdata/complex/1000000rows.json")
}

func loadTestData(filename string) []map[string]any {
	// Check if file exists
	_, err := os.Stat(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Try loading split files (e.g., 100000rows.json -> 100000rows_*.json)
			splitRows := loadSplitTestData(filename)
			if len(splitRows) > 0 {
				return splitRows
			}
			// If no split files found, panic with helpful error
			panic("Failed to load test data " + filename + ": file does not exist and no split files found")
		}
		// Other error (permission, etc.)
		panic("Failed to stat test data file " + filename + ": " + err.Error())
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		panic("Failed to load test data " + filename + ": " + err.Error())
	}
	var rows []map[string]any
	if err := json.Unmarshal(data, &rows); err != nil {
		panic("Failed to parse test data " + filename + ": " + err.Error())
	}
	return rows
}

func loadSplitTestData(baseFilename string) []map[string]any {
	// Extract directory and base name
	dir := filepath.Dir(baseFilename)
	baseName := filepath.Base(baseFilename)
	ext := filepath.Ext(baseName)
	nameWithoutExt := baseName[:len(baseName)-len(ext)]

	// Find all split files matching pattern: basename_*.json
	pattern := filepath.Join(dir, nameWithoutExt+"_*"+ext)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		// Return empty slice instead of panicking - let caller handle it
		return nil
	}
	if len(matches) == 0 {
		// Return empty slice instead of panicking - let caller handle it
		return nil
	}

	// Sort matches to ensure correct order
	sort.Strings(matches)

	// Load and combine all split files
	var allRows []map[string]any
	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			panic("Failed to load split test data " + match + ": " + err.Error())
		}
		var rows []map[string]any
		if err := json.Unmarshal(data, &rows); err != nil {
			panic("Failed to parse split test data " + match + ": " + err.Error())
		}
		allRows = append(allRows, rows...)
	}
	return allRows
}

// Simple data benchmarks - Basic structs (8 fields, primitives/strings/time)

func BenchmarkDeserialize_Simple_10rows_typedb(b *testing.B) {
	rows := simpleData10
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*SimpleUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*SimpleUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Simple_10rows_manual(b *testing.B) {
	rows := simpleData10
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*SimpleUser, 0, len(rows))
		for _, row := range rows {
			// Manual struct population (baseline)
			user := &SimpleUser{
				ID:        int64(row["id"].(float64)),
				Name:      row["name"].(string),
				Email:     row["email"].(string),
				CreatedAt: parseTimeFromString(row["created_at"].(string)),
				UpdatedAt: parseTimeFromString(row["updated_at"].(string)),
				Active:    row["active"].(bool),
				Age:       int(row["age"].(float64)),
				Score:     row["score"].(float64),
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Simple_100rows_typedb(b *testing.B) {
	rows := simpleData100
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*SimpleUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*SimpleUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Simple_1krows_typedb(b *testing.B) {
	rows := simpleData1k
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*SimpleUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*SimpleUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Simple_10krows_typedb(b *testing.B) {
	rows := simpleData10k
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*SimpleUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*SimpleUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Simple_100krows_typedb(b *testing.B) {
	rows := simpleData100k
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*SimpleUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*SimpleUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Simple_1Mrows_typedb(b *testing.B) {
	rows := simpleData1M
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*SimpleUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*SimpleUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

// Complex data benchmarks - Complex structs (20 fields, JSONB, arrays, nested types)

func BenchmarkDeserialize_Complex_10rows_typedb(b *testing.B) {
	rows := complexData10
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*ComplexUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*ComplexUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Complex_10rows_manual(b *testing.B) {
	rows := complexData10
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*ComplexUser, 0, len(rows))
		for _, row := range rows {
			// Manual struct population with complex type handling (baseline)
			// This is a simplified version - full manual implementation would be very verbose
			user := &ComplexUser{
				ID:        int64(row["id"].(float64)),
				Name:      row["name"].(string),
				Email:     row["email"].(string),
				CreatedAt: parseTimeFromString(row["created_at"].(string)),
				UpdatedAt: parseTimeFromString(row["updated_at"].(string)),
				Active:    row["active"].(bool),
				Age:       int(row["age"].(float64)),
				Score:     row["score"].(float64),
				// Note: Full manual implementation of JSONB, arrays, nested structs
				// would require extensive type assertions and unmarshaling
				// This is a simplified baseline
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Complex_100rows_typedb(b *testing.B) {
	rows := complexData100
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*ComplexUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*ComplexUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Complex_1krows_typedb(b *testing.B) {
	rows := complexData1k
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*ComplexUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*ComplexUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Complex_10krows_typedb(b *testing.B) {
	rows := complexData10k
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*ComplexUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*ComplexUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Complex_100krows_typedb(b *testing.B) {
	rows := complexData100k
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*ComplexUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*ComplexUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

func BenchmarkDeserialize_Complex_1Mrows_typedb(b *testing.B) {
	rows := complexData1M
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		users := make([]*ComplexUser, 0, len(rows))
		for _, row := range rows {
			user, err := DeserializeForBenchmark[*ComplexUser](row)
			if err != nil {
				b.Fatal(err)
			}
			users = append(users, user)
		}
		_ = users
	}
}

// Helper function to parse time from string
func parseTimeFromString(s string) time.Time {
	// Use the same parseTime function that typedb uses internally
	t, err := parseTime(s)
	if err != nil {
		panic("Failed to parse time: " + err.Error())
	}
	return t
}
