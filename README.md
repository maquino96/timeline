# Timeline

Personal chronological feed aggregator. Pulls content from Reddit, Hacker News, RSS feeds, and SEC EDGAR filings into one unified timeline. Topic-based keyword matching tags items for filterable views.

**No algorithms. No recommendations. Just one chronological feed of what you care about.**

## Architecture

- **Backend**: Go API server (pure Go, zero CGo) with SQLite + FTS5
- **Frontend**: Next.js 16 + React 19 + Tailwind 4
- **Single binary deploy**: Go binary serves both API and static frontend

## Quick Start

```bash
# Backend
cd backend && go run .
# → http://localhost:8080

# Frontend (separate terminal)
cd frontend && npm install && npm run dev
# → http://localhost:3000
```

## Production Build

```bash
cd frontend && npm run build
cd backend && go build -o timeline-server .
cd backend && ./timeline-server
# → http://localhost:8080 (serves API + static frontend)
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server listen port |
| `TIMELINE_DB_PATH` | `timeline.db` | SQLite database path |
| `REDDIT_CLIENT_ID` | — | Reddit script app client ID |
| `REDDIT_CLIENT_SECRET` | — | Reddit script app secret |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Backend URL (frontend dev) |

## License

MIT
