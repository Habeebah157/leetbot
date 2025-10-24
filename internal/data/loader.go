package data

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Problem struct {
	ID         int
	URL        string
	Title      string
	Difficulty string
	Acceptance float64
	Frequency  float64
}

type ProblemsByCompany struct {
	data map[string]map[string][]Problem
}

func LoadAllProblems() (*ProblemsByCompany, error) {
	pbc := &ProblemsByCompany{
		data: make(map[string]map[string][]Problem),
	}

	for company, timeframes := range embeddedCSVs {
		pbc.data[company] = make(map[string][]Problem)
		for timeframe, csvData := range timeframes {
			problems, err := parseCSV(csvData)
			if err != nil {
				return nil, fmt.Errorf("error parsing CSV for %s/%s: %w", company, timeframe, err)
			}
			pbc.data[company][timeframe] = problems
		}
	}

	return pbc, nil
}

func (pbc *ProblemsByCompany) GetProblems(company, timeframe string) []Problem {
	company = strings.ToLower(strings.TrimSpace(company))
	timeframe = normalizeTimeframe(timeframe)

	if companyData, ok := pbc.data[company]; ok {
		if problems, ok := companyData[timeframe]; ok {
			return problems
		}
	}

	return nil
}

func (pbc *ProblemsByCompany) GetAvailableCompanies() []string {
	companies := make([]string, 0, len(pbc.data))
	for company := range pbc.data {
		companies = append(companies, company)
	}
	sort.Strings(companies)
	return companies
}

func (pbc *ProblemsByCompany) GetAvailableTimeframes(company string) []string {
	company = strings.ToLower(strings.TrimSpace(company))

	if companyData, ok := pbc.data[company]; ok {
		timeframes := make([]string, 0, len(companyData))
		for timeframe := range companyData {
			timeframes = append(timeframes, timeframe)
		}
		sort.Strings(timeframes)
		return timeframes
	}

	return nil
}

func (pbc *ProblemsByCompany) CompanyExists(company string) bool {
	company = strings.ToLower(strings.TrimSpace(company))
	_, exists := pbc.data[company]
	return exists
}

// thirty-days > three-months > six-months > more-than-six-months > all
func (pbc *ProblemsByCompany) GetProblemsWithPriority(company string) ([]Problem, string) {
	company = strings.ToLower(strings.TrimSpace(company))

	priorities := []string{"thirty-days", "three-months", "six-months", "more-than-six-months", "all"}

	if companyData, ok := pbc.data[company]; ok {
		for _, timeframe := range priorities {
			if problems, ok := companyData[timeframe]; ok && len(problems) > 0 {
				return problems, timeframe
			}
		}
	}

	return nil, ""
}

func normalizeTimeframe(timeframe string) string {
	timeframe = strings.ToLower(strings.TrimSpace(timeframe))
	timeframe = strings.ReplaceAll(timeframe, " ", "-")

	switch timeframe {
	case "30", "30days", "30-days", "thirty", "thirtydays", "thirty-days", "30d":
		return "thirty-days"
	case "90", "90days", "90-days", "three", "threemonths", "three-months", "3months", "3-months", "3mo", "90d":
		return "three-months"
	case "180", "180days", "180-days", "six", "sixmonths", "six-months", "6months", "6-months", "6mo":
		return "six-months"
	case "all", "alltime", "all-time", "everything", "":
		return "all"
	case "more-than-six-months", "morethan6months", "more-than-6-months", ">6mo", ">6months":
		return "more-than-six-months"
	}

	for _, tf := range []string{"all", "thirty-days", "three-months", "six-months", "more-than-six-months"} {
		if timeframe == tf {
			return tf
		}
	}

	return "all"
}

func parseCSV(csvData []byte) ([]Problem, error) {
	reader := csv.NewReader(strings.NewReader(string(csvData)))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return []Problem{}, nil
	}

	var problems []Problem
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 6 {
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			continue
		}

		acceptance, _ := parsePercentage(record[4])
		frequency, _ := parsePercentage(record[5])

		problem := Problem{
			ID:         id,
			URL:        record[1],
			Title:      record[2],
			Difficulty: record[3],
			Acceptance: acceptance,
			Frequency:  frequency,
		}

		problems = append(problems, problem)
	}

	sort.Slice(problems, func(i, j int) bool {
		return problems[i].Frequency > problems[j].Frequency
	})

	return problems, nil
}

func parsePercentage(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	return strconv.ParseFloat(s, 64)
}

func NewTestProblemsByCompany(testData map[string]map[string][]Problem) *ProblemsByCompany {
	pbc := &ProblemsByCompany{
		data: make(map[string]map[string][]Problem),
	}

	for company, timeframes := range testData {
		pbc.data[company] = make(map[string][]Problem)
		for timeframe, problems := range timeframes {
			pbc.data[company][timeframe] = problems
		}
	}

	return pbc
}
