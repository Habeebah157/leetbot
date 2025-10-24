# leetbot

A Discord bot that emits coding interview problems (derived from [leetcode.com](https://leetcode.com/)) by company using the [discordgo](https://github.com/bwmarrin/discordgo) library.

<img src="./internal/docs/example.webp" alt="Leetbot Example With Google (30d)" width="500" height="auto">

<img src="./internal/docs/example2.webp" alt="Leetbot No Data Example (Jump)" width="500" height="auto">

- leetbot [embeds](https://pkg.go.dev/embed) `csv` files within the compiled binary keeping latency near zero ✅
- leetbot uses the [dgo-paginator](https://github.com/topi314/dgo-paginator) library to paginate the results ✅
- leetbot supports both text (`!problems google`) and slash commands (`/problems google`) ✅
- leetbot supports suggestions, autocompletion, [fuzzy search](https://pkg.go.dev/github.com/lithammer/fuzzysearch@v1.1.8), and validation for company names ✅
- leetbot supports multiple timeframes (`all`, `30d`, `3mo`, `6mo`, `>6mo`) ✅

## Commands

The bot supports both text commands (with prefix `!`) and slash commands (with prefix `/`) for native interactions.

### Slash Commands (Reccomended)
```
/problems company:<company> [timeframe:<timeframe>]
/help
```

### Text Commands (Legacy)
```
!problems <company> [timeframe]
!help
```

**Examples:**
- `!problems airbnb` or `/problems company:airbnb` - Show most popular problems for Airbnb (all time)
-  `!problems susquehanna >6mo`
-  `!problems amazon 30d`
-  `!problems HRT all`

**Supported timeframes:**
- `all` (default) - All time
- `thirty-days`, `30-days`, `30d` - Last 30 days
- `three-months`, `3-months`, `3mo` - Last 3 months
- `six-months`, `6-months`, `6mo` - Last 6 months
- `more-than-six-months`, `>6mo` - More than 6 months ago

> [!NOTE]
> Leetbot uses a priority system to determine the default timeframe if no timeframe is specified.
> It will try to use the most recent timeframe that has data first.
> If no data is found, it will use the next most recent timeframe until it reaches the default timeframe (all time).

## Setup locally

### Prerequisites

- Go 1.21 or later
- Discord Bot Token
- Bot must be added to your Discord server with `application.commands` scope

### Discord Bot Setup

1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Create a new application or use existing bot
3. Go to **OAuth2** → **URL Generator**
4. Select scopes: `bot` and `applications.commands`
5. Select permissions: `Send Messages`, `Use Slash Commands`, `Read Message History`
6. Use the generated URL to add the bot to your server

### Installation

1. Clone the repository:
```bash
git clone https://github.com/dotcomnerd/leetbot
cd leetbot
```

2. Copy environment file and configure:
```bash
cp .env.example .env
```

> [!IMPORTANT]
> Be sure to update the .env file with your actual Discord bot token.

3. Install dependencies:
```bash
make setup
```

4. Run the bot in development mode:
```bash
make dev
# or
go run ./cmd/bot
# or
make run
```


## Development

### Available Make Commands

- `make help` - Show available commands
- `make dev` - Run with live reload using air
- `make lint` - Run linter
- `make test` - Run tests
- `make test-coverage` - Run tests with coverage report
- `make build` - Build the application
- `make run` - Build and run the application
- `make clean` - Clean build artifacts
- `make docker-build` - Build Docker image
- `make docker-run` - Run Docker container
- `make validate` - Run linting and tests
- `make setup` - Setup development environment
- `make generate-embedded` - Generate embedded CSV data from actual files
- `make validate-data` - Validate all CSV files in data directory
- `make demo` - Run the bot demo

### Adding New Companies

1. Create a new directory under `data/`:
```bash
mkdir data/new-company
```

2. Add CSV files for each timeframe (all.csv, thirty-days.csv, etc.)

3. Update `internal/data/parser.go`:
   - Add embed directives for the new CSV files
   - Add the company to the `embeddedCSVs` map

> [!TIP]
> Run `make generate-embedded` to generate the embedded data automatically.

## CSV Format

CSV files should have the following columns:
- `ID` - LeetCode problem ID
- `URL` - LeetCode problem URL
- `Title` - Problem title
- `Difficulty` - Easy/Medium/Hard
- `Acceptance %` - LeetCode acceptance rate
- `Frequency %` - How often this problem appears in interviews (used for sorting)

Example:
```csv
ID,URL,Title,Difficulty,Acceptance %,Frequency %
1,https://leetcode.com/problems/two-sum,Two Sum,Easy,55.9%,100.0%
2,https://leetcode.com/problems/add-two-numbers,Add Two Numbers,Medium,46.4%,75.0%
```

## Docker

Build and run with Docker:

```bash
make docker-build
make docker-run
```

## Deployment

The bot is designed to be deployed as a compiled binary to platforms like:
- Fly.io
- Any VPS or cloud instance

For deployment, simply build the binary and run it with the appropriate environment variables. Cheers to go!
```bash
make build
./bin/leetbot
```

## License

This project is licensed under the [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.html).
