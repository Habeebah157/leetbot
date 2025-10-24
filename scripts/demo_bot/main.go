package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/tn/leetbot/internal/data"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titleCaser = cases.Title(language.English)

func getDifficultyIndicator(difficulty string) string {
	switch strings.ToLower(difficulty) {
	case "easy":
		return "🟢"
	case "medium":
		return "🟡"
	case "hard":
		return "🔴"
	default:
		return ""
	}
}

func main() {
	fmt.Println("=== Leetbot Demo ===")
	fmt.Println()

	fmt.Println("Loading problems data...")
	problemsData, err := data.LoadAllProblems()
	if err != nil {
		log.Fatalf("Failed to load problems data: %v", err)
	}

	companies := problemsData.GetAvailableCompanies()
	fmt.Printf("Available companies: %v\n", companies)
	fmt.Println()

	fmt.Println("=== Command Examples ===")
	fmt.Println()

	fmt.Println("Command: !problems airbnb")
	demoProblemsResponse(problemsData, "airbnb", "all")
	fmt.Println()

	fmt.Println("Command: !problems amazon thirty-days")
	demoProblemsResponse(problemsData, "amazon", "thirty-days")
	fmt.Println()

	fmt.Println("Command: !problems google three-months")
	demoProblemsResponse(problemsData, "google", "three-months")
	fmt.Println()

	fmt.Println("=== Help Command ===")
	fmt.Println("Command: !help")
	fmt.Println("Response would show:")
	fmt.Println("**Leetbot Commands:**")
	fmt.Println("**!problems <company> [timeframe]** - Show popular coding interview problems by company")
	fmt.Println("• company: Company name (e.g., airbnb, amazon, google)")
	fmt.Println("• timeframe: Optional timeframe filter")
	fmt.Println("  - all (default) - All time")
	fmt.Println("  - thirty-days - Last 30 days")
	fmt.Println("  - three-months - Last 3 months")
	fmt.Println("  - six-months - Last 6 months")
	fmt.Println("  - more-than-six-months - More than 6 months ago")
	fmt.Println()
	fmt.Println("**Examples:**")
	fmt.Println("• !problems airbnb")
	fmt.Println("• !problems amazon thirty-days")
	fmt.Println("• !problems google three-months")
}

func demoProblemsResponse(problemsData *data.ProblemsByCompany, company, timeframe string) {
	problems := problemsData.GetProblems(company, timeframe)
	if problems == nil {
		fmt.Printf("❌ No data found for company '%s' and timeframe '%s'\n", company, timeframe)
		return
	}

	companyName := titleCaser.String(company)
	displayTimeframe := formatTimeframeDisplay(timeframe)
	title := fmt.Sprintf("Most Popular Problems for %s (%s):", companyName, displayTimeframe)

	fmt.Println(title)
	fmt.Println()

	const maxLength = 2000
	currentLength := len(title) + 2

	for i, problem := range problems {
		difficultyIndicator := getDifficultyIndicator(problem.Difficulty)
		problemLine := fmt.Sprintf("%.1f%%: %s %s - %s",
			problem.Frequency, difficultyIndicator, problem.Title, problem.URL)

		if currentLength+len(problemLine) > maxLength {
			fmt.Printf("\n... and %d more problems\n", len(problems)-i)
			break
		}

		fmt.Println(problemLine)
		currentLength += len(problemLine) + 1
	}
}

func formatTimeframeDisplay(timeframe string) string {
	switch timeframe {
	case "all":
		return "All Time"
	case "thirty-days":
		return "Last 30 Days"
	case "three-months":
		return "Last 3 Months"
	case "six-months":
		return "Last 6 Months"
	case "more-than-six-months":
		return "More than 6 Months"
	default:
		return titleCaser.String(strings.ReplaceAll(timeframe, "-", " "))
	}
}
