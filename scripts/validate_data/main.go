package main

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Problem struct {
	ID         int     `csv:"ID"`
	URL        string  `csv:"URL"`
	Title      string  `csv:"Title"`
	Difficulty string  `csv:"Difficulty"`
	Acceptance float64 `csv:"Acceptance %"`
	Frequency  float64 `csv:"Frequency %"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run scripts/validate_data/main.go <data-directory>")
		os.Exit(1)
	}

	dataDir := os.Args[1]
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Printf("Data directory %s does not exist\n", dataDir)
		os.Exit(1)
	}

	fmt.Printf("Validating CSV data in %s...\n", dataDir)

	var totalFiles, validFiles, invalidFiles int
	var totalProblems int

	err := filepath.WalkDir(dataDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".csv") || strings.Contains(path, ".git") {
			return nil
		}

		totalFiles++
		fmt.Printf("Validating %s... ", filepath.Base(path))

		if err := validateCSV(path); err != nil {
			fmt.Printf("❌ INVALID: %v\n", err)
			invalidFiles++
			return nil
		}

		fmt.Printf("✅ VALID\n")
		validFiles++
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nValidation Summary:\n")
	fmt.Printf("  Total CSV files: %d\n", totalFiles)
	fmt.Printf("  Valid files: %d\n", validFiles)
	fmt.Printf("  Invalid files: %d\n", invalidFiles)
	fmt.Printf("  Total problems: %d\n", totalProblems)

	if invalidFiles > 0 {
		fmt.Printf("❌ Validation failed: %d files have errors\n", invalidFiles)
		os.Exit(1)
	}

	fmt.Printf("✅ All CSV files are valid!\n")
}

func validateCSV(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("cannot read header: %v", err)
	}
	expectedHeader := []string{"ID", "URL", "Title", "Difficulty", "Acceptance %", "Frequency %"}
	if len(header) != len(expectedHeader) {
		return fmt.Errorf("header has %d columns, expected %d", len(header), len(expectedHeader))
	}

	for i, expected := range expectedHeader {
		if header[i] != expected {
			return fmt.Errorf("header[%d] = %q, expected %q", i, header[i], expected)
		}
	}

	lineNum := 1
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("error reading line %d: %v", lineNum+1, err)
		}

		lineNum++

		if err := validateRecord(record); err != nil {
			return fmt.Errorf("line %d: %v", lineNum, err)
		}
	}

	return nil
}

func validateRecord(record []string) error {
	if len(record) != 6 {
		return fmt.Errorf("record has %d fields, expected 6", len(record))
	}

	if record[0] == "" {
		return fmt.Errorf("ID is empty")
	}

	if record[1] == "" {
		return fmt.Errorf("URL is empty")
	}
	if !strings.HasPrefix(record[1], "https://leetcode.com/") {
		return fmt.Errorf("URL does not start with https://leetcode.com/")
	}

	if record[2] == "" {
		return fmt.Errorf("Title is empty")
	}

	difficulty := strings.ToLower(record[3])
	if difficulty != "easy" && difficulty != "medium" && difficulty != "hard" {
		return fmt.Errorf("Difficulty %q is not valid (must be Easy, Medium, or Hard)", record[3])
	}

	if record[4] == "" {
		return fmt.Errorf("Acceptance %% is empty")
	}
	if _, err := parsePercentage(record[4]); err != nil {
		return fmt.Errorf("Acceptance %% %q is not a valid number: %v", record[4], err)
	}

	if record[5] == "" {
		return fmt.Errorf("Frequency %% is empty")
	}
	frequency, err := parsePercentage(record[5])
	if err != nil {
		return fmt.Errorf("Frequency %% %q is not a valid number", record[5])
	}
	if frequency < 0 || frequency > 100 {
		return fmt.Errorf("Frequency %% %.1f is out of range [0, 100]", frequency)
	}

	return nil
}

func parsePercentage(s string) (float64, error) {
	if !strings.HasSuffix(s, "%") {
		return 0, fmt.Errorf("percentage must end with %%")
	}
	s = strings.TrimSuffix(s, "%")
	return strconv.ParseFloat(s, 64)
}
