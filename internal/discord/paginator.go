package discord

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/tn/leetbot/internal/data"
	paginator "github.com/topi314/dgo-paginator"
)

const (
	problemsPerPage     = 10
	paginationThreshold = 20
)

var PaginatorManager *paginator.Manager

func init() {

	PaginatorManager = paginator.NewManager(

		paginator.WithButtonsConfig(paginator.ButtonsConfig{
			First: &paginator.ComponentOptions{
				Emoji: &discordgo.ComponentEmoji{
					Name: "⏮",
				},
				Style: discordgo.PrimaryButton,
			},
			Back: &paginator.ComponentOptions{
				Emoji: &discordgo.ComponentEmoji{
					Name: "◀",
				},
				Style: discordgo.PrimaryButton,
			},
			Stop: nil,
			Next: &paginator.ComponentOptions{
				Emoji: &discordgo.ComponentEmoji{
					Name: "▶",
				},
				Style: discordgo.PrimaryButton,
			},
			Last: &paginator.ComponentOptions{
				Emoji: &discordgo.ComponentEmoji{
					Name: "⏩",
				},
				Style: discordgo.PrimaryButton,
			},
		}),

		paginator.WithNotYourPaginatorMessage("This paginator can only be used by the person who requested it."),
	)
}

func shouldUsePagination(problemCount int) bool {
	return problemCount > paginationThreshold
}

func createProblemsPaginator(company, timeframe string, problems []data.Problem) *paginator.Paginator {
	totalPages := (len(problems) + problemsPerPage - 1) / problemsPerPage

	return &paginator.Paginator{
		PageFunc: func(page int, embed *discordgo.MessageEmbed) {

			start := page * problemsPerPage
			end := start + problemsPerPage
			if end > len(problems) {
				end = len(problems)
			}

			pageProblems := problems[start:end]

			embed.Title = fmt.Sprintf("Most Popular Problems for %s (%s)",
				formatCompanyName(company),
				formatTimeframeDisplay(timeframe))
			embed.Color = 0x5865F2
			embed.Footer = &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Page %d/%d • Total: %d problems", page+1, totalPages, len(problems)),
			}
			embed.Timestamp = time.Now().Format(time.RFC3339)

			description := ""
			for i, problem := range pageProblems {
				problemNumber := start + i + 1
				difficultyIndicator := getDifficultyIndicator(problem.Difficulty)
				description += fmt.Sprintf("**%d.** %s [%s](<%s>) `%.0f%%`\n",
					problemNumber,
					difficultyIndicator,
					problem.Title,
					problem.URL,
					problem.Frequency)
			}

			embed.Description = description
		},
		MaxPages:        totalPages,
		ExpiryLastUsage: true,
		Expiry:          time.Now().Add(10 * time.Minute),
	}
}

func sendPaginatedProblems(s *discordgo.Session, i *discordgo.InteractionCreate, company, timeframe string, problems []data.Problem) error {
	pg := createProblemsPaginator(company, timeframe, problems)

	return PaginatorManager.CreateInteraction(s, i.Interaction, pg, false)
}

func sendPaginatedProblemsMessage(s *discordgo.Session, channelID, company, timeframe string, problems []data.Problem) error {
	pg := createProblemsPaginator(company, timeframe, problems)

	return PaginatorManager.CreateMessage(s, channelID, pg)
}
