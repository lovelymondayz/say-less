# Say Less — AI Text Simplification

An AI-powered text simplification tool that helps you write clearer, more concise content.

## Quick Start

```bash
# Clone
git clone https://github.com/lovelymondayz/say-less.git
cd say-less

# Start all services
docker compose up -d --build

# Frontend: http://localhost:3006
# Backend API: http://localhost:8086
# DB: localhost:5437
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        NGINX (80/443)                        │
│                   sayless.arjism.com → :3006                │
├─────────────────────────────────────────────────────────────┤
│  React + Vite + TS + Tailwind  │  Go + GIN + pgx + Postgres │
│        (Frontend :3006)        │       (Backend :8086)      │
├─────────────────────────────────────────────────────────────┤
│              PostgreSQL :5437  │  AI Provider (9Router)     │
└─────────────────────────────────────────────────────────────┘
```

## Features

- **AI Text Simplification**: Paste text → get simpler alternatives
- **Multiple Tones**: Professional, casual, concise modes
- **Document Management**: Save and organize simplified documents
- **Version History**: Track all changes with restore capability
- **Real-time Preview**: Side-by-side comparison
- **Export Options**: Plain text, Markdown, PDF
- **Responsive UI**: Mobile-first design with Tailwind CSS

## API Endpoints

### Public
- `GET /api/health` — Health check

### Authenticated
- `POST /api/simplify` — Simplify text
- `GET /api/documents` — List documents
- `POST /api/documents` — Create document
- `GET /api/documents/:id` — Get document
- `PUT /api/documents/:id` — Update document
- `DELETE /api/documents/:id` — Delete document

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8086 | Backend port |
| DATABASE_URL | postgres://... | DB connection |
| JWT_SECRET | - | JWT signing key |
| AI_BASE_URL | https://9router.nousresearch.com/v1 | AI API base URL |
| AI_MODEL | - | AI model name |

## Development

```bash
# Backend only
cd backend
go run .

# Frontend only
cd frontend
npm install
npm run dev
```

## Deployment

1. Push to `main` → GitHub Action auto-deploys
2. Or manually: `ssh vps && cd /root/say-less && ./update.sh`

## License

MIT
