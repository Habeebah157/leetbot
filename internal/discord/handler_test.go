package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/tn/leetbot/internal/data"
)

func createTestProblemsData() *data.ProblemsByCompany {
	testData := map[string]map[string][]data.Problem{
		"airbnb": {
			"all": []data.Problem{
				{
					ID:         1,
					URL:        "https://leetcode.com/problems/two-sum",
					Title:      "Two Sum",
					Difficulty: "Easy",
					Acceptance: 55.9,
					Frequency:  100.0,
				},
				{
					ID:         2,
					URL:        "https://leetcode.com/problems/add-two-numbers",
					Title:      "Add Two Numbers",
					Difficulty: "Medium",
					Acceptance: 46.4,
					Frequency:  75.0,
				},
			},
			"thirty-days": []data.Problem{
				{
					ID:         68,
					URL:        "https://leetcode.com/problems/text-justification",
					Title:      "Text Justification",
					Difficulty: "Hard",
					Acceptance: 48.4,
					Frequency:  100.0,
				},
			},
		},
		"amazon": {
			"all": []data.Problem{
				{
					ID:         1,
					URL:        "https://leetcode.com/problems/two-sum",
					Title:      "Two Sum",
					Difficulty: "Easy",
					Acceptance: 55.9,
					Frequency:  100.0,
				},
			},
		},
	}

	return data.NewTestProblemsByCompany(testData)
}

func TestNewHandler(t *testing.T) {
	problemsData := createTestProblemsData()
	handler := NewHandler(problemsData, "!")

	if handler.problemsData == nil {
		t.Error("NewHandler() should set problemsData")
	}

	if handler.prefix != "!" {
		t.Errorf("NewHandler() prefix = %v, want %v", handler.prefix, "!")
	}
}

func TestNormalizeTimeframe(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		input    string
		expected string
	}{
		{"30", "thirty-days"},
		{"30d", "thirty-days"},
		{"30days", "thirty-days"},
		{"thirty", "thirty-days"},
		{"thirtydays", "thirty-days"},
		{"90", "three-months"},
		{"3mo", "three-months"},
		{"3months", "three-months"},
		{"three", "three-months"},
		{"threemonths", "three-months"},
		{"180", "six-months"},
		{"6mo", "six-months"},
		{"6months", "six-months"},
		{"six", "six-months"},
		{"sixmonths", "six-months"},
		{">6mo", "more-than-six-months"},
		{"all", "all"},
		{"alltime", "all"},
		{"everything", "all"},
		{"ALL", "all"},
		{"Thirty-Days", "thirty-days"},
		{"invalid", "all"},
	}

	for _, tt := range tests {
		result := handler.NormalizeTimeframe(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeTimeframe(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatTimeframeDisplay(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		input    string
		expected string
	}{
		{"all", "all"},
		{"thirty-days", "last 30 days"},
		{"three-months", "last 3 months"},
		{"six-months", "last 6 months"},
		{"more-than-six-months", "more than 6 months"},
		{"invalid", "invalid"},
	}

	for _, tt := range tests {
		result := handler.formatTimeframeDisplay(tt.input)
		if result != tt.expected {
			t.Errorf("formatTimeframeDisplay(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatProblemsResponse(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	problems := []data.Problem{
		{
			ID:         1,
			URL:        "https://leetcode.com/problems/two-sum",
			Title:      "Two Sum",
			Difficulty: "Easy",
			Acceptance: 55.9,
			Frequency:  100.0,
		},
		{
			ID:         2,
			URL:        "https://leetcode.com/problems/add-two-numbers",
			Title:      "Add Two Numbers",
			Difficulty: "Medium",
			Acceptance: 46.4,
			Frequency:  75.0,
		},
	}

	result := handler.formatProblemsResponse("airbnb", "all", problems)

	if !contains(result, "Most Popular Problems for Airbnb (all):") {
		t.Error("formatProblemsResponse() should contain title")
	}

	if !contains(result, "Two Sum (100%)") {
		t.Error("formatProblemsResponse() should contain first problem")
	}

	if !contains(result, "Add Two Numbers (75%)") {
		t.Error("formatProblemsResponse() should contain second problem")
	}

	if len(result) > 2000 {
		t.Errorf("formatProblemsResponse() result too long: %d characters", len(result))
	}
}

func TestFormatProblemsResponse_Empty(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	result := handler.formatProblemsResponse("airbnb", "all", []data.Problem{})

	if !contains(result, "No problems found") {
		t.Error("formatProblemsResponse() should handle empty problems list")
	}
}

func TestHandleMessage(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content:   "!problems airbnb",
			ChannelID: "channel123",
		},
	}

	handler.HandleMessage(session, message)
}

func TestHandleMessage_BotMessage(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "bot123",
				Bot: true,
			},
			Content: "!problems airbnb",
		},
	}

	handler.HandleMessage(session, message)
}

func TestHandleMessage_WrongPrefix(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content: "problems airbnb",
		},
	}

	handler.HandleMessage(session, message)
}

func TestHandleMessage_UnknownCommand(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content: "!unknown command",
		},
	}

	handler.HandleMessage(session, message)
}

func TestHandleMessage_ProblemsNoArgs(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content:   "!problems",
			ChannelID: "channel123",
		},
	}

	handler.HandleMessage(session, message)
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || (len(s) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestFindCompanyByFuzzySearch(t *testing.T) {
	testData := map[string]map[string][]data.Problem{
		"palantir-technologies": {
			"all": []data.Problem{
				{ID: 1, Title: "Test", Difficulty: "Easy", Frequency: 100.0},
			},
		},
		"capital-one": {
			"all": []data.Problem{
				{ID: 2, Title: "Test", Difficulty: "Medium", Frequency: 90.0},
			},
		},
		"goldman-sachs": {
			"all": []data.Problem{
				{ID: 3, Title: "Test", Difficulty: "Hard", Frequency: 85.0},
			},
		},
		"jane-street": {
			"all": []data.Problem{
				{ID: 4, Title: "Test", Difficulty: "Medium", Frequency: 80.0},
			},
		},
	}
	problemsData := data.NewTestProblemsByCompany(testData)

	tests := []struct {
		input    string
		expected string
		found    bool
	}{
		{"palantir", "palantir-technologies", true},
		{"palantir-technologies", "palantir-technologies", true},
		{"Palantir Technologies", "palantir-technologies", true},
		{"capital one", "capital-one", true},
		{"Capital One", "capital-one", true},
		{"capital-one", "capital-one", true},
		{"goldman", "goldman-sachs", true},
		{"jane street", "jane-street", true},
		{"nonexistent-company", "", false},
	}

	for _, tt := range tests {
		company, found := findCompanyByFuzzySearch(tt.input, problemsData)
		if found != tt.found {
			t.Errorf("findCompanyByFuzzySearch(%q) found = %v, want %v", tt.input, found, tt.found)
		}
		if found && company != tt.expected {
			t.Errorf("findCompanyByFuzzySearch(%q) = %q, want %q", tt.input, company, tt.expected)
		}
	}
}

func TestGetCompanyAutocompleteChoices(t *testing.T) {
	testData := map[string]map[string][]data.Problem{
		"airbnb": {
			"all": []data.Problem{{ID: 1, Title: "Test", Difficulty: "Easy", Frequency: 100.0}},
		},
		"amazon": {
			"all": []data.Problem{{ID: 2, Title: "Test", Difficulty: "Medium", Frequency: 90.0}},
		},
		"apple": {
			"all": []data.Problem{{ID: 3, Title: "Test", Difficulty: "Hard", Frequency: 85.0}},
		},
	}
	problemsData := data.NewTestProblemsByCompany(testData)

	choices := getCompanyAutocompleteChoices("", problemsData)
	if len(choices) == 0 {
		t.Error("getCompanyAutocompleteChoices with empty input should return choices")
	}
	choices = getCompanyAutocompleteChoices("air", problemsData)
	if len(choices) == 0 {
		t.Error("getCompanyAutocompleteChoices('air') should return choices")
	}

	foundAirbnb := false
	for _, choice := range choices {
		if choice.Value == "airbnb" {
			foundAirbnb = true
			break
		}
	}
	if !foundAirbnb {
		t.Error("getCompanyAutocompleteChoices('air') should include airbnb")
	}
	if len(choices) > 25 {
		t.Errorf("getCompanyAutocompleteChoices should return max 25 choices, got %d", len(choices))
	}
}

func TestIsTimeframeKeyword(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		input    string
		expected bool
	}{
		{"30d", true},
		{"3mo", true},
		{"6mo", true},
		{">6mo", true},
		{"all", true},
		{"thirty-days", true},
		{"three-months", true},
		{"six-months", true},
		{"more-than-six-months", true},
		{"company", false},
		{"amazon", false},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		result := handler.isTimeframeKeyword(tt.input)
		if result != tt.expected {
			t.Errorf("isTimeframeKeyword(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestFormatAvailableTimeframesSuggestion(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	availableTimeframes := []string{"all", "six-months"}
	result := handler.formatAvailableTimeframesSuggestion("starbucks", "thirty-days", availableTimeframes)

	if !contains(result, "No data found for Starbucks") {
		t.Error("Suggestion should mention no data found for company")
	}
	if !contains(result, "Available timeframes for Starbucks:") {
		t.Error("Suggestion should list available timeframes")
	}
	if !contains(result, "all") {
		t.Error("Suggestion should include 'all' timeframe")
	}
	if !contains(result, "6mo") {
		t.Error("Suggestion should include short alias for six-months")
	}
	if !contains(result, "!problems starbucks") {
		t.Error("Suggestion should include example command")
	}
}

func TestGetTimeframeShortAlias(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		input    string
		expected string
	}{
		{"thirty-days", "30d"},
		{"three-months", "3mo"},
		{"six-months", "6mo"},
		{"more-than-six-months", ">6mo"},
		{"all", "all"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		result := handler.getTimeframeShortAlias(tt.input)
		if result != tt.expected {
			t.Errorf("getTimeframeShortAlias(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatProblemsResponse_LongList(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	var longList []data.Problem
	for i := 1; i <= 100; i++ {
		longList = append(longList, data.Problem{
			ID:         i,
			URL:        "https://leetcode.com/problems/test",
			Title:      "Test Problem",
			Difficulty: "Easy",
			Acceptance: 50.0,
			Frequency:  100.0,
		})
	}

	result := handler.formatProblemsResponse("test-company", "all", longList)

	if len(result) > 2000 {
		t.Errorf("formatProblemsResponse() length = %d, exceeds Discord limit of 2000", len(result))
	}
}

func TestHandleMessage_ProblemsWithTimeframe(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content:   "!problems airbnb 30d",
			ChannelID: "channel123",
		},
	}

	handler.HandleMessage(session, message)
}

func TestHandleMessage_UnknownCompany(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	session := &discordgo.Session{Token: ""}
	message := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Author: &discordgo.User{
				ID:  "user123",
				Bot: false,
			},
			Content:   "!problems unknowncompany",
			ChannelID: "channel123",
		},
	}

	handler.HandleMessage(session, message)
}

func TestNormalizeTimeframe_EdgeCases(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	tests := []struct {
		input    string
		expected string
	}{
		{"  30d  ", "thirty-days"},

		{"AlL", "all"},
		{"ThIrTy-DaYs", "thirty-days"},

		{"", "all"},

		{"thirty days", "thirty-days"},
		{"three months", "three-months"},
	}

	for _, tt := range tests {
		result := handler.NormalizeTimeframe(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeTimeframe(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatProblemsResponse_DifferentDifficulties(t *testing.T) {
	handler := NewHandler(createTestProblemsData(), "!")

	problems := []data.Problem{
		{ID: 1, Title: "Easy Problem", Difficulty: "Easy", Frequency: 100.0},
		{ID: 2, Title: "Medium Problem", Difficulty: "Medium", Frequency: 90.0},
		{ID: 3, Title: "Hard Problem", Difficulty: "Hard", Frequency: 80.0},
	}

	result := handler.formatProblemsResponse("test-company", "all", problems)

	if !contains(result, "🟢") {
		t.Error("formatProblemsResponse() should contain 🟢 for Easy")
	}
	if !contains(result, "🟡") {
		t.Error("formatProblemsResponse() should contain 🟡 for Medium")
	}
	if !contains(result, "🔴") {
		t.Error("formatProblemsResponse() should contain 🔴 for Hard")
	}
}
