package discord

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/tn/leetbot/internal/data"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var SlashCommandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, problemsData *data.ProblemsByCompany, prefix string){}

func init() {
	SlashCommandHandlers["problems"] = handleProblemsSlashCommand
	SlashCommandHandlers["help"] = handleHelpSlashCommand
}

func HandleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate, problemsData *data.ProblemsByCompany) {
	data := i.ApplicationCommandData()

	if data.Name != "problems" {
		return
	}

	var choices []*discordgo.ApplicationCommandOptionChoice
	var currentInput string

	for _, option := range data.Options {
		if option.Name == "company" && option.Focused {
			currentInput = option.StringValue()
			choices = getCompanyAutocompleteChoices(currentInput, problemsData)
			break
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		fmt.Printf("Error responding to autocomplete: %v\n", err)
	}
}

func GetSlashCommands(problemsData *data.ProblemsByCompany) []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "problems",
			Description: "Show popular coding interview problems by company",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:         discordgo.ApplicationCommandOptionString,
					Name:         "company",
					Description:  "Company name (start typing to search)",
					Required:     true,
					Autocomplete: true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "timeframe",
					Description: "Time period (optional)",
					Required:    false,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{
							Name:  "All Time",
							Value: "all",
						},
						{
							Name:  "Last 30 Days",
							Value: "thirty-days",
						},
						{
							Name:  "Last 3 Months",
							Value: "three-months",
						},
						{
							Name:  "Last 6 Months",
							Value: "six-months",
						},
						{
							Name:  "More than 6 Months",
							Value: "more-than-six-months",
						},
					},
				},
			},
		},
		{
			Name:        "help",
			Description: "Show available bot commands and usage",
		},
	}
}

func formatCompanyName(company string) string {
	words := strings.Split(company, "-")
	caser := cases.Title(language.English)
	for i, word := range words {
		words[i] = caser.String(word)
	}
	return strings.Join(words, " ")
}

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

func findCompanyByFuzzySearch(input string, problemsData *data.ProblemsByCompany) (string, bool) {
	if input == "" {
		return "", false
	}

	companies := problemsData.GetAvailableCompanies()
	if len(companies) == 0 {
		return "", false
	}

	normalizedInput := strings.ToLower(input)
	normalizedInput = strings.ReplaceAll(normalizedInput, " ", "-")
	for _, company := range companies {
		if company == normalizedInput {
			return company, true
		}
	}

	for _, company := range companies {
		if strings.Contains(company, normalizedInput) {
			return company, true
		}
	}

	var displayNames []string
	for _, company := range companies {
		displayNames = append(displayNames, formatCompanyName(company))
	}
	matches := fuzzy.RankFindNormalizedFold(input, displayNames)
	if len(matches) > 0 {
		bestMatchIdx := matches[0].OriginalIndex
		if bestMatchIdx >= 0 && bestMatchIdx < len(companies) {
			return companies[bestMatchIdx], true
		}
	}

	matches = fuzzy.RankFindNormalizedFold(normalizedInput, companies)
	if len(matches) > 0 {
		return matches[0].Target, true
	}

	return "", false
}

func getCompanyAutocompleteChoices(input string, problemsData *data.ProblemsByCompany) []*discordgo.ApplicationCommandOptionChoice {
	companies := problemsData.GetAvailableCompanies()
	var choices []*discordgo.ApplicationCommandOptionChoice

	if input == "" {
		popularCompanies := []string{
			"amazon", "google", "facebook", "microsoft", "apple", "netflix",
			"uber", "meta", "tesla", "nvidia", "openai", "anthropic",
		}
		for _, company := range popularCompanies {
			if problemsData.CompanyExists(company) {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  formatCompanyName(company),
					Value: company,
				})
			}
		}

		companiesSorted := make([]string, len(companies))
		copy(companiesSorted, companies)
		for i := 0; i < len(companiesSorted)-1; i++ {
			for j := i + 1; j < len(companiesSorted); j++ {
				if companiesSorted[i] > companiesSorted[j] {
					companiesSorted[i], companiesSorted[j] = companiesSorted[j], companiesSorted[i]
				}
			}
		}

		for _, company := range companiesSorted {
			isPopular := false
			for _, pop := range popularCompanies {
				if company == pop {
					isPopular = true
					break
				}
			}
			if !isPopular {
				choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
					Name:  formatCompanyName(company),
					Value: company,
				})
				if len(choices) >= 25 {
					break
				}
			}
		}
		return choices
	}

	normalizedInput := strings.ToLower(input)
	normalizedInput = strings.ReplaceAll(normalizedInput, " ", "-")
	type companyMatch struct {
		slug        string
		displayName string
	}
	var companyList []companyMatch
	for _, company := range companies {
		companyList = append(companyList, companyMatch{
			slug:        company,
			displayName: formatCompanyName(company),
		})
	}

	var displayNames []string
	for _, cm := range companyList {
		displayNames = append(displayNames, cm.displayName)
	}
	matches := fuzzy.RankFindNormalizedFold(input, displayNames)

	for i, match := range matches {
		if i >= 25 {
			break
		}
		if match.OriginalIndex >= 0 && match.OriginalIndex < len(companyList) {
			cm := companyList[match.OriginalIndex]
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  cm.displayName,
				Value: cm.slug,
			})
		}
	}

	if len(choices) == 0 {
		slugMatches := fuzzy.RankFindNormalizedFold(normalizedInput, companies)
		for i, match := range slugMatches {
			if i >= 25 {
				break
			}
			choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
				Name:  formatCompanyName(match.Target),
				Value: match.Target,
			})
		}
	}

	return choices
}

type Handler struct {
	problemsData *data.ProblemsByCompany
	prefix       string
}

func NewHandler(problemsData *data.ProblemsByCompany, prefix string) *Handler {
	return &Handler{
		problemsData: problemsData,
		prefix:       prefix,
	}
}

func (h *Handler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if !strings.HasPrefix(m.Content, h.prefix) {
		return
	}
	content := strings.TrimPrefix(m.Content, h.prefix)
	content = strings.TrimSpace(content)

	if content == "" {
		return
	}

	parts := strings.Fields(content)
	command := strings.ToLower(parts[0])
	args := parts[1:]

	switch command {
	case "problems":
		h.handleProblemsCommand(s, m, args)
	case "help":
		h.handleHelpCommand(s, m)
	default:
		h.sendErrorMessage(s, m.ChannelID, fmt.Sprintf("Unknown command '%s'. Use `%shelp` for available commands.", command, h.prefix))
	}
}

func (h *Handler) handleProblemsCommand(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		h.sendErrorMessage(s, m.ChannelID, "Please specify a company. Usage: !problems <company> [timeframe]")
		return
	}

	companyInput := args[0]
	timeframeArg := ""

	if len(args) >= 2 {
		lastArg := strings.ToLower(args[len(args)-1])
		if h.isTimeframeKeyword(lastArg) {
			timeframeArg = lastArg
			if len(args) > 2 {
				companyInput = strings.Join(args[:len(args)-1], " ")
			}
		} else {
			companyInput = strings.Join(args, " ")
		}
	}
	company, found := findCompanyByFuzzySearch(companyInput, h.problemsData)
	if !found {
		h.sendErrorMessage(s, m.ChannelID, fmt.Sprintf("Could not find company matching '%s'. Try using more specific terms or check the spelling.", companyInput))
		return
	}

	var problems []data.Problem
	var timeframe string

	if timeframeArg != "" {

		timeframe = h.NormalizeTimeframe(timeframeArg)
		problems = h.problemsData.GetProblems(company, timeframe)
	} else {

		problems, timeframe = h.problemsData.GetProblemsWithPriority(company)
	}

	if problems == nil {

		availableTimeframes := h.problemsData.GetAvailableTimeframes(company)
		if len(availableTimeframes) > 0 && timeframeArg != "" {

			suggestion := h.formatAvailableTimeframesSuggestion(company, timeframe, availableTimeframes)
			h.sendMessage(s, m.ChannelID, suggestion)
		} else {

			h.sendMessage(s, m.ChannelID, fmt.Sprintf("No data found for company '%s'", formatCompanyName(company)))
		}
		return
	}

	if shouldUsePagination(len(problems)) {
		err := sendPaginatedProblemsMessage(s, m.ChannelID, company, timeframe, problems)
		if err != nil {
			fmt.Printf("Error sending paginated message: %v\n", err)

			response := h.formatProblemsResponse(company, timeframe, problems)
			h.sendMessage(s, m.ChannelID, response)
		}
		return
	}

	response := h.formatProblemsResponse(company, timeframe, problems)
	h.sendMessage(s, m.ChannelID, response)
}

func (h *Handler) isTimeframeKeyword(s string) bool {
	s = strings.ToLower(s)
	timeframeKeywords := []string{
		"all", "alltime", "everything",
		"30", "30d", "30days", "thirty", "thirtydays",
		"90", "3mo", "90days", "3months", "three", "threemonths",
		"180", "6mo", "180days", "6months", "six", "sixmonths",
		">6mo", "more-than-six-months",
		"thirty-days", "three-months", "six-months", "more-than-six-months",
	}

	for _, keyword := range timeframeKeywords {
		if s == keyword || strings.Contains(s, keyword) {
			return true
		}
	}
	return false
}

func (h *Handler) NormalizeTimeframe(timeframe string) string {
	timeframe = strings.ToLower(strings.TrimSpace(timeframe))
	timeframe = strings.ReplaceAll(timeframe, " ", "-")
	switch timeframe {
	case "30", "30d", "90d", "30days", "30-days", "thirty", "thirtydays", "thirty-days":
		return "thirty-days"
	case "90", "3mo", "90days", "90-days", "three", "threemonths", "three-months", "3months", "3-months":
		return "three-months"
	case "180", "6mo", "180days", "180-days", "six", "sixmonths", "six-months", "6months", "6-months":
		return "six-months"
	case ">6mo", ">6months", "more-than-six-months", "more-than-6-months", "morethan6months":
		return "more-than-six-months"
	case "all", "alltime", "all-time", "everything", "":
		return "all"
	default:
		for _, tf := range []string{"all", "thirty-days", "three-months", "six-months", "more-than-six-months"} {
			if timeframe == tf {
				return tf
			}
		}
		return "all"
	}
}

func (h *Handler) formatProblemsResponse(company, timeframe string, problems []data.Problem) string {
	if len(problems) == 0 {
		return fmt.Sprintf("No problems found for %s (%s)", formatCompanyName(company), h.formatTimeframeDisplay(timeframe))
	}

	displayTimeframe := h.formatTimeframeDisplay(timeframe)

	title := fmt.Sprintf("Most Popular Problems for %s (%s):", formatCompanyName(company), displayTimeframe)

	var message strings.Builder
	message.WriteString(title + "\n")
	maxProblems := 20
	if len(problems) < maxProblems {
		maxProblems = len(problems)
	}

	for i := 0; i < maxProblems; i++ {
		problem := problems[i]
		difficultyIndicator := getDifficultyIndicator(problem.Difficulty)
		problemLine := fmt.Sprintf("%s %s (%.0f%%): %s\n",
			difficultyIndicator, problem.Title, problem.Frequency, problem.URL)
		message.WriteString(problemLine)
	}

	return message.String()
}

func (h *Handler) formatTimeframeDisplay(timeframe string) string {
	switch timeframe {
	case "all":
		return "all"
	case "thirty-days":
		return "last 30 days"
	case "three-months":
		return "last 3 months"
	case "six-months":
		return "last 6 months"
	case "more-than-six-months":
		return "more than 6 months"
	default:
		return strings.ToLower(strings.ReplaceAll(timeframe, "-", " "))
	}
}

func (h *Handler) formatAvailableTimeframesSuggestion(company, requestedTimeframe string, availableTimeframes []string) string {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("No data found for %s (%s).\n\n",
		formatCompanyName(company),
		h.formatTimeframeDisplay(requestedTimeframe)))

	message.WriteString(fmt.Sprintf("Available timeframes for %s:\n", formatCompanyName(company)))

	priorityOrder := map[string]int{
		"thirty-days":          1,
		"three-months":         2,
		"six-months":           3,
		"more-than-six-months": 4,
		"all":                  5,
	}

	type timeframeWithPriority struct {
		name     string
		priority int
	}

	var sortedTimeframes []timeframeWithPriority
	for _, tf := range availableTimeframes {
		priority := priorityOrder[tf]
		if priority == 0 {
			priority = 999
		}
		sortedTimeframes = append(sortedTimeframes, timeframeWithPriority{name: tf, priority: priority})
	}

	for i := 0; i < len(sortedTimeframes)-1; i++ {
		for j := i + 1; j < len(sortedTimeframes); j++ {
			if sortedTimeframes[i].priority > sortedTimeframes[j].priority {
				sortedTimeframes[i], sortedTimeframes[j] = sortedTimeframes[j], sortedTimeframes[i]
			}
		}
	}

	for _, tf := range sortedTimeframes {

		shortAlias := h.getTimeframeShortAlias(tf.name)
		message.WriteString(fmt.Sprintf("• **%s** (%s)\n", shortAlias, h.formatTimeframeDisplay(tf.name)))
	}

	message.WriteString(fmt.Sprintf("\nTry: `%sproblems %s <timeframe>`", h.prefix, company))

	return message.String()
}

func (h *Handler) getTimeframeShortAlias(timeframe string) string {
	switch timeframe {
	case "thirty-days":
		return "30d"
	case "three-months":
		return "3mo"
	case "six-months":
		return "6mo"
	case "more-than-six-months":
		return ">6mo"
	case "all":
		return "all"
	default:
		return timeframe
	}
}

func (h *Handler) sendMessage(s *discordgo.Session, channelID, message string) {
	if s == nil {
		fmt.Printf("[TEST] Would send to %s: %s\n", channelID, message)
		return
	}

	if s.Token == "" {
		fmt.Printf("[TEST] Would send to %s: %s\n", channelID, message)
		return
	}
	_, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: message,
		Flags:   discordgo.MessageFlagsSuppressEmbeds,
	})
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
	}
}

func (h *Handler) handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	helpMsg := fmt.Sprintf(`**Leetbot Commands:**

**Text Commands (prefix: %s):**
• **%sproblems <company> [timeframe]** - Show interview problems

**Slash Commands:**
• **/problems** - Show interview problems (with dropdown options)
• **/help** - Show this help message

**Problems Command Usage:**
• *company*: Company name (e.g., airbnb, amazon, google)
• *timeframe*: Optional timeframe filter (if not specified, uses smart priority system)

**Timeframe Options:**
• **all** - All time
• **30d** or **thirty-days** - Last 30 days
• **3mo** or **three-months** - Last 3 months
• **6mo** or **six-months** - Last 6 months
• **>6mo** or **more-than-six-months** - More than 6 months ago

**Smart Priority System:**
When no timeframe is specified, the bot automatically tries:
1. Last 30 days (most recent)
2. Last 3 months (if 30d has no data)
3. Last 6 months (if 3mo has no data)
4. More than 6 months (if 6mo has no data)
5. All time (fallback)

**Examples:**
• %sproblems airbnb (uses smart priority)
• %sproblems amazon 30d (forces 30 days)
• %sproblems google 3mo (forces 3 months)
• /problems company:airbnb (uses smart priority)
• /problems company:amazon timeframe:thirty-days

**Supported Companies:**
• Use the slash command dropdown to see all available companies!

*Problems are sorted by interview frequency (most popular first)*`, h.prefix, h.prefix, h.prefix, h.prefix, h.prefix)

	h.sendMessage(s, m.ChannelID, helpMsg)
}

func handleProblemsSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate, problemsData *data.ProblemsByCompany, prefix string) {
	options := i.ApplicationCommandData().Options
	optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
	for _, opt := range options {
		optionMap[opt.Name] = opt
	}

	companyOpt, ok := optionMap["company"]
	if !ok {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Company is required!",
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}
	company := strings.ToLower(companyOpt.StringValue())

	var problems []data.Problem
	var timeframe string

	if timeframeOpt, ok := optionMap["timeframe"]; ok {

		timeframe = timeframeOpt.StringValue()
		problems = problemsData.GetProblems(company, timeframe)
	} else {

		problems, timeframe = problemsData.GetProblemsWithPriority(company)
	}

	if problems == nil {

		availableTimeframes := problemsData.GetAvailableTimeframes(company)
		var responseContent string

		if len(availableTimeframes) > 0 {

			_, specifiedTimeframe := optionMap["timeframe"]
			if specifiedTimeframe {

				responseContent = formatAvailableTimeframesSuggestionSlash(company, timeframe, availableTimeframes)
			} else {

				responseContent = fmt.Sprintf("No data found for %s", formatCompanyName(company))
			}
		} else {

			responseContent = fmt.Sprintf("No data found for %s", formatCompanyName(company))
		}

		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: responseContent,
				Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	if shouldUsePagination(len(problems)) {
		err := sendPaginatedProblems(s, i, company, timeframe, problems)
		if err != nil {
			fmt.Printf("Error sending paginated response: %v\n", err)

			response := formatProblemsResponse(company, timeframe, problems)
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: response,
					Flags:   discordgo.MessageFlagsSuppressEmbeds,
				},
			})
			if err != nil {
				fmt.Printf("Error responding to interaction: %v\n", err)
			}
		}
		return
	}

	response := formatProblemsResponse(company, timeframe, problems)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: response,
			Flags:   discordgo.MessageFlagsSuppressEmbeds,
		},
	})
	if err != nil {
		fmt.Printf("Error responding to interaction: %v\n", err)
	}
}

func handleHelpSlashCommand(s *discordgo.Session, i *discordgo.InteractionCreate, problemsData *data.ProblemsByCompany, prefix string) {
	helpMsg := fmt.Sprintf(`**Leetbot Commands:**

**Text Commands (prefix: %s):**
• **%sproblems <company> [timeframe]** - Show interview problems

**Slash Commands:**
• **/problems** - Show interview problems (with dropdown options)
• **/help** - Show this help message

**Problems Command Usage:**
• *company*: Company name (e.g., airbnb, amazon, google)
• *timeframe*: Optional timeframe filter (if not specified, uses smart priority system)

**Timeframe Options:**
• **all** - All time
• **30d** or **thirty-days** - Last 30 days
• **3mo** or **three-months** - Last 3 months
• **6mo** or **six-months** - Last 6 months
• **>6mo** or **more-than-six-months** - More than 6 months ago

**Smart Priority System:**
When no timeframe is specified, the bot automatically tries:
1. Last 30 days (most recent)
2. Last 3 months (if 30d has no data)
3. Last 6 months (if 3mo has no data)
4. More than 6 months (if 6mo has no data)
5. All time (fallback)

**Examples:**
• %sproblems airbnb (uses smart priority)
• %sproblems amazon 30d (forces 30 days)
• %sproblems google 3mo (forces 3 months)
• /problems company:airbnb (uses smart priority)
• /problems company:amazon timeframe:thirty-days

**Supported Companies:**
• Use the slash command dropdown to see all available companies!

*Problems are sorted by interview frequency (most popular first)*`, prefix, prefix, prefix, prefix, prefix)

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: helpMsg,
			Flags:   discordgo.MessageFlagsEphemeral | discordgo.MessageFlagsSuppressEmbeds,
		},
	})
	if err != nil {
		fmt.Printf("Error responding to interaction: %v\n", err)
	}
}

func formatProblemsResponse(company, timeframe string, problems []data.Problem) string {
	if len(problems) == 0 {
		return fmt.Sprintf("No problems found for %s (%s)", formatCompanyName(company), formatTimeframeDisplay(timeframe))
	}

	displayTimeframe := formatTimeframeDisplay(timeframe)

	title := fmt.Sprintf("Most Popular Problems for %s (%s):", formatCompanyName(company), displayTimeframe)

	var message strings.Builder
	message.WriteString(title + "\n")
	maxProblems := 20
	if len(problems) < maxProblems {
		maxProblems = len(problems)
	}

	for i := 0; i < maxProblems; i++ {
		problem := problems[i]
		difficultyIndicator := getDifficultyIndicator(problem.Difficulty)
		problemLine := fmt.Sprintf("%s %s (%.0f%%): %s\n",
			difficultyIndicator, problem.Title, problem.Frequency, problem.URL)
		message.WriteString(problemLine)
	}

	return message.String()
}

func formatTimeframeDisplay(timeframe string) string {
	switch timeframe {
	case "all":
		return "all"
	case "thirty-days":
		return "last 30 days"
	case "three-months":
		return "last 3 months"
	case "six-months":
		return "last 6 months"
	case "more-than-six-months":
		return "more than 6 months"
	default:
		return strings.ToLower(strings.ReplaceAll(timeframe, "-", " "))
	}
}

func formatAvailableTimeframesSuggestionSlash(company, requestedTimeframe string, availableTimeframes []string) string {
	var message strings.Builder
	message.WriteString(fmt.Sprintf("No data found for %s (%s).\n\n",
		formatCompanyName(company),
		formatTimeframeDisplay(requestedTimeframe)))

	message.WriteString(fmt.Sprintf("Available timeframes for %s:\n", formatCompanyName(company)))

	priorityOrder := map[string]int{
		"thirty-days":          1,
		"three-months":         2,
		"six-months":           3,
		"more-than-six-months": 4,
		"all":                  5,
	}

	type timeframeWithPriority struct {
		name     string
		priority int
	}

	var sortedTimeframes []timeframeWithPriority
	for _, tf := range availableTimeframes {
		priority := priorityOrder[tf]
		if priority == 0 {
			priority = 999
		}
		sortedTimeframes = append(sortedTimeframes, timeframeWithPriority{name: tf, priority: priority})
	}

	for i := 0; i < len(sortedTimeframes)-1; i++ {
		for j := i + 1; j < len(sortedTimeframes); j++ {
			if sortedTimeframes[i].priority > sortedTimeframes[j].priority {
				sortedTimeframes[i], sortedTimeframes[j] = sortedTimeframes[j], sortedTimeframes[i]
			}
		}
	}

	for _, tf := range sortedTimeframes {

		message.WriteString(fmt.Sprintf("• **%s** (%s)\n", tf.name, formatTimeframeDisplay(tf.name)))
	}

	message.WriteString(fmt.Sprintf("\nTry: `/problems company:%s timeframe:<option>`", company))

	return message.String()
}

func (h *Handler) sendErrorMessage(s *discordgo.Session, channelID, message string) {

	h.sendMessage(s, channelID, message)
}
