package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Problem struct {
	ID         int
	URL        string
	Title      string
	Difficulty string
	Acceptance float64
	Frequency  float64
}

func normalizeVarName(name string) string {
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	reg := regexp.MustCompile("[^a-zA-Z0-9_]")
	name = reg.ReplaceAllString(name, "")

	parts := strings.Split(name, "_")
	for i, part := range parts {
		if part != "" {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	name = strings.Join(parts, "")

	if len(name) > 0 && unicode.IsDigit(rune(name[0])) {
		name = "_" + name
	}

	return name
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run scripts/generate_embedded/main.go <data-directory>")
		os.Exit(1)
	}

	dataDir := os.Args[1]
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Printf("Data directory %s does not exist\n", dataDir)
		os.Exit(1)
	}

	fmt.Printf("Generating embedded data from %s...\n", dataDir)
	companies, err := findCompanies(dataDir)
	if err != nil {
		fmt.Printf("Error finding companies: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d companies\n", len(companies))

	output := `package data


`

	for _, company := range companies {
		companyData, err := loadCompanyData(dataDir, company)
		if err != nil {
			fmt.Printf("Error loading data for %s: %v\n", company, err)
			continue
		}

		for timeframe, problems := range companyData {
			companyName := normalizeVarName(company)
			timeframeName := normalizeVarName(timeframe)
			varName := fmt.Sprintf("%s%sCSV", companyName, timeframeName)
			csvContent := generateCSVContent(problems)

			output += fmt.Sprintf("var %s = `%s`\n\n", varName, csvContent)
		}
	}

	output += `// embeddedCSVs maps company and timeframe to their embedded CSV data
var embeddedCSVs = map[string]map[string][]byte{
`

	for _, company := range companies {
		output += fmt.Sprintf("\t%q: {\n", company)

		companyData, err := loadCompanyData(dataDir, company)
		if err != nil {
			continue
		}

		for timeframe := range companyData {
			companyName := normalizeVarName(company)
			timeframeName := normalizeVarName(timeframe)
			varName := fmt.Sprintf("%s%sCSV", companyName, timeframeName)
			output += fmt.Sprintf("\t\t%q: []byte(%s),\n", timeframe, varName)
		}

		output += "\t},\n"
	}

	output += "}\n"

	err = os.WriteFile("internal/data/parser_generated.go", []byte(output), 0644)
	if err != nil {
		fmt.Printf("Error writing generated file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated embedded data for %d companies\n", len(companies))
	fmt.Printf("Generated file: internal/data/parser_generated.go\n")
}

func findCompanies(dataDir string) ([]string, error) {
	var companies []string

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			companies = append(companies, entry.Name())
		}
	}

	sort.Strings(companies)
	return companies, nil
}

func loadCompanyData(dataDir, company string) (map[string][]Problem, error) {
	companyDir := filepath.Join(dataDir, company)
	timeframes := map[string][]Problem{}

	entries, err := os.ReadDir(companyDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".csv") {
			timeframe := strings.TrimSuffix(entry.Name(), ".csv")
			problems, err := loadCSV(filepath.Join(companyDir, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("error loading %s: %v", entry.Name(), err)
			}

			sort.Slice(problems, func(i, j int) bool {
				return problems[i].Frequency > problems[j].Frequency
			})

			timeframes[timeframe] = problems
		}
	}

	return timeframes, nil
}

func loadCSV(filename string) ([]Problem, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("CSV file must have at least header and one data row")
	}

	expectedHeader := []string{"ID", "URL", "Title", "Difficulty", "Acceptance %", "Frequency %"}
	headerFields := strings.Split(lines[0], ",")
	for i, field := range headerFields {
		if i < len(expectedHeader) {
			field = strings.TrimSpace(field)
			if field != expectedHeader[i] {
				return nil, fmt.Errorf("invalid header[%d]: expected %q, got %q", i, expectedHeader[i], field)
			}
		}
	}

	var problems []Problem

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		problem, err := parseCSVLine(line)
		if err != nil {
			return nil, fmt.Errorf("error parsing line %d: %v", i+1, err)
		}

		problems = append(problems, problem)
	}

	return problems, nil
}

func parseCSVLine(line string) (Problem, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 6 {
		return Problem{}, fmt.Errorf("line has fewer than 6 fields")
	}

	difficulty := strings.TrimSpace(parts[len(parts)-3])
	acceptanceStr := strings.TrimSpace(parts[len(parts)-2])
	frequencyStr := strings.TrimSpace(parts[len(parts)-1])
	id, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return Problem{}, fmt.Errorf("invalid ID: %v", err)
	}
	url := strings.TrimSpace(parts[1])

	titleStart := 2
	titleEnd := len(parts) - 4
	if titleEnd < titleStart {
		return Problem{}, fmt.Errorf("cannot determine title boundaries")
	}
	var titleParts []string
	for i := titleStart; i <= titleEnd; i++ {
		titleParts = append(titleParts, strings.TrimSpace(parts[i]))
	}
	title := strings.Join(titleParts, ", ")

	acceptance, err := parsePercentage(acceptanceStr)
	if err != nil {
		return Problem{}, fmt.Errorf("invalid acceptance percentage: %v", err)
	}

	frequency, err := parsePercentage(frequencyStr)
	if err != nil {
		return Problem{}, fmt.Errorf("invalid frequency percentage: %v", err)
	}

	return Problem{
		ID:         id,
		URL:        url,
		Title:      title,
		Difficulty: difficulty,
		Acceptance: acceptance,
		Frequency:  frequency,
	}, nil
}

func parsePercentage(s string) (float64, error) {
	s = strings.TrimSuffix(s, "%")
	return strconv.ParseFloat(s, 64)
}

func generateCSVContent(problems []Problem) string {
	var lines []string
	lines = append(lines, "ID,URL,Title,Difficulty,Acceptance %,Frequency %")

	for _, problem := range problems {
		// escape backticks by replacing them with a unicode lookalike or removing them
		// backticks would break Go's raw string literals
		escapedTitle := strings.ReplaceAll(problem.Title, "`", "'")
		escapedTitle = strings.ReplaceAll(escapedTitle, "\n", " ")
		escapedTitle = strings.ReplaceAll(escapedTitle, "\r", " ")
		escapedTitle = strings.ReplaceAll(escapedTitle, "\t", " ")
		// escape quotes for CSV format
		escapedTitle = strings.ReplaceAll(escapedTitle, "\"", "\"\"")

		line := fmt.Sprintf("%d,%s,\"%s\",%s,%.1f%%,%.1f%%",
			problem.ID, problem.URL, escapedTitle, problem.Difficulty,
			problem.Acceptance, problem.Frequency)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
