package main

import (
	"fmt"
	"net/http"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/tn/leetbot/internal/config"
	"github.com/tn/leetbot/internal/data"
	"github.com/tn/leetbot/internal/discord"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	fmt.Printf("Starting leetbot with prefix '%s'...\n", cfg.BotPrefix)
	fmt.Println("Loading problems data...")
	problemsData, err := data.LoadAllProblems()
	if err != nil {
		log.Fatalf("Failed to load problems data: %v", err)
	}

	fmt.Printf("Loaded data for %d companies\n", len(problemsData.GetAvailableCompanies()))

	// Start a minimal HTTP health server so platforms can probe readiness/liveness.
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("leetbot running"))
		})

		port := 8080
		if p := os.Getenv("PORT"); p != "" {
			if v, err := strconv.Atoi(p); err == nil {
				port = v
			}
		}
		addr := fmt.Sprintf(":%d", port)
		log.Printf("Health server listening on %s", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("health server stopped: %v", err)
		}
	}()

	handler := discord.NewHandler(problemsData, cfg.BotPrefix)
	dg, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		log.Fatalf("Failed to create Discord session: %v", err)
	}

	dg.AddHandler(handler.HandleMessage)

	dg.AddHandler(discord.PaginatorManager.OnInteractionCreate)

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			if h, ok := discord.SlashCommandHandlers[i.ApplicationCommandData().Name]; ok {
				h(s, i, problemsData, cfg.BotPrefix)
			}
		case discordgo.InteractionApplicationCommandAutocomplete:
			discord.HandleAutocomplete(s, i, problemsData)
		}
	})

	dg.Identify.Intents = discordgo.IntentsGuildMessages
	dg.AddHandler(func(s *discordgo.Session, event *discordgo.Ready) {
		log.Printf("Leetbot logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)

		commands := discord.GetSlashCommands(problemsData)

		log.Println("Registering slash commands...")
		for _, v := range commands {
			_, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
			if err != nil {
				log.Printf("Cannot create '%v' command: %v", v.Name, err)
			} else {
				log.Printf("✓ Registered command: /%v", v.Name)
			}
		}
		log.Printf("Finished registering slash commands")
	})

	err = dg.Open()
	if err != nil {
		log.Fatalf("Failed to open Discord connection: %v", err)
	}
	defer dg.Close()

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	fmt.Println("Shutting down bot...")
}
