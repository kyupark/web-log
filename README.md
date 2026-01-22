# web-log

A CLI tool that summarizes your Safari and Chrome browsing history into a structured, tag-based journal using AI.

## Why?

Browsing history is a goldmine of personal context - what you researched, watched, shopped for, read about. But raw history is noisy and hard to review. `web-log` transforms it into a readable summary perfect for:

- Personal journaling
- Weekly reviews
- Recalling "what did I do last week?"
- Tracking research rabbit holes

## Features

- Reads from both **Safari** and **Chrome** history on macOS
- Groups browsing by **topic/tag**, not by site
- Uses AI (via OpenRouter) to intelligently categorize and summarize
- Outputs clean **Markdown** with hierarchical tags
- Provides **specific details** (not generic descriptions) for better recall
- Supports custom date ranges

## Installation

```bash
go install github.com/user/web-log/cmd/web-log@latest
```

Or build from source:

```bash
git clone https://github.com/user/web-log.git
cd web-log
go install ./cmd/web-log
```

## Setup

Set your OpenRouter API key:

```bash
export OPENROUTER_API_KEY="your-api-key"
```

Optionally set a specific model (default: `google/gemini-2.5-flash`):

```bash
export OPENROUTER_MODEL="google/gemini-2.5-flash"
```

## Usage

```bash
# Last 7 days (default)
web-log

# Last N days
web-log --days 3

# Specific date range
web-log --from 2026-01-01 --to 2026-01-15
```

## Example Output

```markdown
# Browsing Summary - 2026-01-19 to 2026-01-22 (3 days)

**Development**
#ai-agents (48) explored various AI agents and tools
  - #clawdbot (21) researched personal AI assistant, cost optimization, remote command execution [x.com/clawdbot, clawd.bot]
  - #ralph (11) investigated autonomous coding loops [github.com/michaelshimeles/ralphy]

**Finance**
#crypto (15) followed Bitcoin discussions on X
  - #saylor (3) MicroStrategy acquired 22,305 BTC for ~$2.13B at ~$95,284 per bitcoin [x.com/saylor]
  - #bitcoin (6) debates on BTC as risk-on vs risk-off asset, cycle top metrics [x.com/JoeConsorti]

**Shopping**
#garmin (11) compared Venu X1 vs Forerunner 570, checked prices [toppreise.ch, ricardo.ch]
```

## How It Works

1. Reads browsing history from Safari (`~/Library/Safari/History.db`) and Chrome (`~/Library/Application Support/Google/Chrome/Default/History`)
2. Deduplicates entries, keeping the most recent visit
3. Filters out noise (gmail, login pages, auth pages)
4. Formats history as a time-ordered table grouped by date
5. Sends to AI model via OpenRouter for structured summarization
6. Returns a tag-based Markdown summary with specific details

## Supported Models

Any model available on [OpenRouter](https://openrouter.ai), including:
- `google/gemini-2.5-flash` (default)
- `openai/gpt-4o-mini`
- `openai/gpt-oss-120b`
- `anthropic/claude-3-haiku`

## Requirements

- macOS (for Safari/Chrome history access)
- Full Disk Access permission for terminal app (to read Safari history)
- OpenRouter API key

## Privacy

- All processing happens locally + via your own OpenRouter API key
- No data is stored or sent anywhere except to OpenRouter for summarization
- History databases are read-only (copied to temp file before reading)

## License

MIT
