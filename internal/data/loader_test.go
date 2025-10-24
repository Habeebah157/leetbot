package data

import (
	"testing"
)

func TestNewTestProblemsByCompany(t *testing.T) {
	testData := map[string]map[string][]Problem{
		"test-company": {
			"all": []Problem{
				{ID: 1, Title: "Test Problem", Difficulty: "Easy", Frequency: 100.0},
			},
		},
	}

	pbc := NewTestProblemsByCompany(testData)

	if pbc == nil {
		t.Fatal("NewTestProblemsByCompany() returned nil")
	}

	if pbc.data == nil {
		t.Error("NewTestProblemsByCompany() data should not be nil")
	}

	if len(pbc.data) != 1 {
		t.Errorf("NewTestProblemsByCompany() data length = %d, want 1", len(pbc.data))
	}
}

func TestGetProblems(t *testing.T) {
	testData := map[string]map[string][]Problem{
		"airbnb": {
			"all": []Problem{
				{ID: 1, Title: "Two Sum", Difficulty: "Easy", Frequency: 100.0},
				{ID: 2, Title: "Add Two Numbers", Difficulty: "Medium", Frequency: 75.0},
			},
			"thirty-days": []Problem{
				{ID: 68, Title: "Text Justification", Difficulty: "Hard", Frequency: 100.0},
			},
		},
	}
	pbc := NewTestProblemsByCompany(testData)

	tests := []struct {
		name      string
		company   string
		timeframe string
		wantCount int
		wantNil   bool
	}{
		{
			name:      "valid company and timeframe",
			company:   "airbnb",
			timeframe: "all",
			wantCount: 2,
			wantNil:   false,
		},
		{
			name:      "valid company with specific timeframe",
			company:   "airbnb",
			timeframe: "thirty-days",
			wantCount: 1,
			wantNil:   false,
		},
		{
			name:      "nonexistent company",
			company:   "nonexistent",
			timeframe: "all",
			wantNil:   true,
		},
		{
			name:      "valid company, invalid timeframe defaults to all",
			company:   "airbnb",
			timeframe: "invalid-timeframe",
			wantCount: 2,
			wantNil:   false,
		},
		{
			name:      "case insensitive company",
			company:   "AIRBNB",
			timeframe: "all",
			wantCount: 2,
			wantNil:   false,
		},
		{
			name:      "company with whitespace",
			company:   "  airbnb  ",
			timeframe: "all",
			wantCount: 2,
			wantNil:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := pbc.GetProblems(tt.company, tt.timeframe)

			if tt.wantNil {
				if problems != nil {
					t.Errorf("GetProblems() = %v, want nil", problems)
				}
			} else {
				if problems == nil {
					t.Fatal("GetProblems() returned nil, want problems")
				}
				if len(problems) != tt.wantCount {
					t.Errorf("GetProblems() count = %d, want %d", len(problems), tt.wantCount)
				}
			}
		})
	}
}

func TestGetAvailableCompanies(t *testing.T) {
	testData := map[string]map[string][]Problem{
		"airbnb": {
			"all": []Problem{{ID: 1, Title: "Test", Difficulty: "Easy", Frequency: 100.0}},
		},
		"amazon": {
			"all": []Problem{{ID: 2, Title: "Test", Difficulty: "Medium", Frequency: 90.0}},
		},
		"google": {
			"all": []Problem{{ID: 3, Title: "Test", Difficulty: "Hard", Frequency: 85.0}},
		},
	}
	pbc := NewTestProblemsByCompany(testData)

	companies := pbc.GetAvailableCompanies()

	if len(companies) != 3 {
		t.Errorf("GetAvailableCompanies() count = %d, want 3", len(companies))
	}

	if len(companies) >= 3 {
		if companies[0] != "airbnb" || companies[1] != "amazon" || companies[2] != "google" {
			t.Errorf("GetAvailableCompanies() not sorted correctly: %v", companies)
		}
	}
}

func TestGetAvailableTimeframes(t *testing.T) {
	testData := map[string]map[string][]Problem{
		"airbnb": {
			"all":          []Problem{{ID: 1, Title: "Test", Difficulty: "Easy", Frequency: 100.0}},
			"thirty-days":  []Problem{{ID: 2, Title: "Test", Difficulty: "Medium", Frequency: 90.0}},
			"three-months": []Problem{{ID: 3, Title: "Test", Difficulty: "Hard", Frequency: 85.0}},
		},
	}
	pbc := NewTestProblemsByCompany(testData)

	timeframes := pbc.GetAvailableTimeframes("airbnb")

	if len(timeframes) != 3 {
		t.Errorf("GetAvailableTimeframes() count = %d, want 3", len(timeframes))
	}

	if len(timeframes) >= 3 {
		if timeframes[0] != "all" || timeframes[1] != "thirty-days" || timeframes[2] != "three-months" {
			t.Errorf("GetAvailableTimeframes() not sorted correctly: %v", timeframes)
		}
	}

	timeframes = pbc.GetAvailableTimeframes("nonexistent")
	if timeframes != nil {
		t.Errorf("GetAvailableTimeframes() for nonexistent company = %v, want nil", timeframes)
	}
}

func TestCompanyExists(t *testing.T) {
	testData := map[string]map[string][]Problem{
		"airbnb": {
			"all": []Problem{{ID: 1, Title: "Test", Difficulty: "Easy", Frequency: 100.0}},
		},
	}
	pbc := NewTestProblemsByCompany(testData)

	tests := []struct {
		name     string
		company  string
		expected bool
	}{
		{"existing company", "airbnb", true},
		{"nonexistent company", "nonexistent", false},
		{"case insensitive", "AIRBNB", true},
		{"with whitespace", "  airbnb  ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pbc.CompanyExists(tt.company)
			if result != tt.expected {
				t.Errorf("CompanyExists(%q) = %v, want %v", tt.company, result, tt.expected)
			}
		})
	}
}

func TestGetProblemsWithPriority(t *testing.T) {
	testData := map[string]map[string][]Problem{
		"company-with-recent": {
			"thirty-days": []Problem{{ID: 1, Title: "Recent", Difficulty: "Easy", Frequency: 100.0}},
			"all":         []Problem{{ID: 2, Title: "Old", Difficulty: "Medium", Frequency: 90.0}},
		},
		"company-with-no-recent": {
			"six-months": []Problem{{ID: 3, Title: "Six Months", Difficulty: "Hard", Frequency: 85.0}},
			"all":        []Problem{{ID: 4, Title: "All", Difficulty: "Easy", Frequency: 80.0}},
		},
		"company-with-all-only": {
			"all": []Problem{{ID: 5, Title: "All Only", Difficulty: "Medium", Frequency: 75.0}},
		},
	}
	pbc := NewTestProblemsByCompany(testData)

	tests := []struct {
		name          string
		company       string
		wantTimeframe string
		wantProblemID int
		wantNil       bool
	}{
		{
			name:          "prioritizes thirty-days",
			company:       "company-with-recent",
			wantTimeframe: "thirty-days",
			wantProblemID: 1,
			wantNil:       false,
		},
		{
			name:          "falls back to six-months",
			company:       "company-with-no-recent",
			wantTimeframe: "six-months",
			wantProblemID: 3,
			wantNil:       false,
		},
		{
			name:          "uses all as final fallback",
			company:       "company-with-all-only",
			wantTimeframe: "all",
			wantProblemID: 5,
			wantNil:       false,
		},
		{
			name:    "returns nil for nonexistent company",
			company: "nonexistent",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems, timeframe := pbc.GetProblemsWithPriority(tt.company)

			if tt.wantNil {
				if problems != nil {
					t.Errorf("GetProblemsWithPriority() problems = %v, want nil", problems)
				}
				if timeframe != "" {
					t.Errorf("GetProblemsWithPriority() timeframe = %q, want empty", timeframe)
				}
			} else {
				if problems == nil {
					t.Fatal("GetProblemsWithPriority() returned nil problems")
				}
				if timeframe != tt.wantTimeframe {
					t.Errorf("GetProblemsWithPriority() timeframe = %q, want %q", timeframe, tt.wantTimeframe)
				}
				if len(problems) > 0 && problems[0].ID != tt.wantProblemID {
					t.Errorf("GetProblemsWithPriority() first problem ID = %d, want %d", problems[0].ID, tt.wantProblemID)
				}
			}
		})
	}
}

func TestNormalizeTimeframe(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"30", "thirty-days"},
		{"30d", "thirty-days"},
		{"30days", "thirty-days"},
		{"30-days", "thirty-days"},
		{"thirty", "thirty-days"},
		{"thirtydays", "thirty-days"},
		{"thirty-days", "thirty-days"},
		{"  30d  ", "thirty-days"},

		{"90", "three-months"},
		{"90d", "three-months"},
		{"3mo", "three-months"},
		{"90days", "three-months"},
		{"90-days", "three-months"},
		{"three", "three-months"},
		{"threemonths", "three-months"},
		{"three-months", "three-months"},
		{"3months", "three-months"},
		{"3-months", "three-months"},

		{"180", "six-months"},
		{"6mo", "six-months"},
		{"180days", "six-months"},
		{"180-days", "six-months"},
		{"six", "six-months"},
		{"sixmonths", "six-months"},
		{"six-months", "six-months"},
		{"6months", "six-months"},
		{"6-months", "six-months"},

		{"all", "all"},
		{"alltime", "all"},
		{"all-time", "all"},
		{"everything", "all"},
		{"", "all"},

		{"more-than-six-months", "more-than-six-months"},
		{"morethan6months", "more-than-six-months"},
		{"more-than-6-months", "more-than-six-months"},
		{">6mo", "more-than-six-months"},
		{">6months", "more-than-six-months"},

		{"ALL", "all"},
		{"Thirty-Days", "thirty-days"},
		{"THREE-MONTHS", "three-months"},

		{"invalid", "all"},
		{"random", "all"},
		{"xyz", "all"},
	}

	for _, tt := range tests {
		result := normalizeTimeframe(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeTimeframe(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParsePercentage(t *testing.T) {
	tests := []struct {
		input     string
		expected  float64
		wantError bool
	}{
		{"50.5%", 50.5, false},
		{"100%", 100.0, false},
		{"0%", 0.0, false},
		{"  75.3%  ", 75.3, false},
		{"50.5", 50.5, false},
		{"invalid", 0.0, true},
		{"", 0.0, true},
	}

	for _, tt := range tests {
		result, err := parsePercentage(tt.input)

		if tt.wantError {
			if err == nil {
				t.Errorf("parsePercentage(%q) expected error, got nil", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("parsePercentage(%q) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("parsePercentage(%q) = %f, want %f", tt.input, result, tt.expected)
			}
		}
	}
}

func TestParseCSV(t *testing.T) {
	tests := []struct {
		name      string
		csvData   string
		wantCount int
		wantErr   bool
	}{
		{
			name: "valid CSV",
			csvData: `ID,URL,Title,Difficulty,Acceptance %,Frequency %
1,https://leetcode.com/problems/two-sum,"Two Sum",Easy,55.9%,100.0%
2,https://leetcode.com/problems/add-two-numbers,"Add Two Numbers",Medium,46.4%,75.0%`,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:      "empty CSV (only header)",
			csvData:   `ID,URL,Title,Difficulty,Acceptance %,Frequency %`,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "CSV with quoted fields containing commas",
			csvData: `ID,URL,Title,Difficulty,Acceptance %,Frequency %
1,https://leetcode.com/problems/test,"Test, with comma",Easy,55.9%,100.0%`,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "empty string",
			csvData:   "",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems, err := parseCSV([]byte(tt.csvData))

			if tt.wantErr {
				if err == nil {
					t.Error("parseCSV() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("parseCSV() unexpected error: %v", err)
				}
				if len(problems) != tt.wantCount {
					t.Errorf("parseCSV() count = %d, want %d", len(problems), tt.wantCount)
				}
			}
		})
	}
}

func TestProblemSorting(t *testing.T) {
	csvData := `ID,URL,Title,Difficulty,Acceptance %,Frequency %
1,https://leetcode.com/problems/low,"Low Frequency",Easy,55.9%,50.0%
2,https://leetcode.com/problems/high,"High Frequency",Medium,46.4%,100.0%
3,https://leetcode.com/problems/medium,"Medium Frequency",Hard,40.0%,75.0%`

	problems, err := parseCSV([]byte(csvData))
	if err != nil {
		t.Fatalf("parseCSV() error = %v", err)
	}

	if len(problems) != 3 {
		t.Fatalf("parseCSV() count = %d, want 3", len(problems))
	}

	if problems[0].Frequency != 100.0 {
		t.Errorf("First problem frequency = %f, want 100.0", problems[0].Frequency)
	}
	if problems[1].Frequency != 75.0 {
		t.Errorf("Second problem frequency = %f, want 75.0", problems[1].Frequency)
	}
	if problems[2].Frequency != 50.0 {
		t.Errorf("Third problem frequency = %f, want 50.0", problems[2].Frequency)
	}
}
