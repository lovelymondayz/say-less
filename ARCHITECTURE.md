# Say Less — Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Cloudflare Edge                          │
│                    sayless.arjism.com (HTTPS)                   │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Cloudflare Tunnel (cf-tunnel)                │
│              http://192.168.88.101:8086 (plain HTTP)            │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Nginx Reverse Proxy                      │
│                    :8086 → :8086 (backend)                      │
│                    :3006 → :3006 (frontend)                     │
└────────────────────────────┬────────────────────────────────────┘
                             │
              ┌──────────────┴──────────────┐
              ▼                              ▼
┌──────────────────────┐        ┌──────────────────────┐
│   Go + GIN Backend   │        │  React + Vite + TS   │
│   :8086 (internal)   │        │  :3006 (internal)    │
│                      │        │                      │
│  - JWT Auth          │        │  - Tailwind CSS      │
│  - pgx + Postgres    │        │  - Text Editor       │
│  - AI Integration    │        │  - Preview           │
│  - Rate Limiting     │        │  - Export            │
└──────────┬───────────┘        └──────────────────────┘
           │
           ▼
┌──────────────────────┐
│   PostgreSQL :5437   │
│                      │
│  - Users             │
│  - Documents         │
│  - AI Conversations  │
└──────────────────────┘
```

## Tech Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Language | Go | 1.22+ |
| Web Framework | Gin | v1.10 |
| Database Driver | pgx | v4 |
| Auth | JWT (golang-jwt/jwt) | v5 |
| AI Integration | OpenAI / 9Router | - |
| Frontend | React + Vite + TypeScript | Vite 5, React 18 |
| Styling | Tailwind CSS | v3 |
| Deployment | Docker Compose | v3.8 |
| Reverse Proxy | Nginx | - |
| Tunnel | Cloudflare Tunnel | - |

## Key Design Decisions

### 1. AI-Powered Text Simplification
- Users paste text → AI suggests simpler alternatives
- Multiple tone options (professional, casual, concise)
- Preserves meaning while reducing word count

### 2. Document Versioning
- Every edit creates a new version
- Full history with diff view
- Restore to any previous version

### 3. Real-time Preview
- Side-by-side original vs simplified
- Instant AI suggestions
- Markdown support

### 4. Export Options
- Plain text, Markdown, PDF
- Copy to clipboard
- Share via link

## API Endpoints

### Public
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/health` | Health check |

### Authenticated (JWT required)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/simplify` | Simplify text |
| GET | `/api/documents` | List documents |
| POST | `/api/documents` | Create document |
| GET | `/api/documents/:id` | Get document |
| PUT | `/api/documents/:id` | Update document |
| DELETE | `/api/documents/:id` | Delete document |

## Ports

| Service | External | Internal |
|---------|----------|----------|
| Backend | `:8086` | `:8086` |
| Frontend | `:3006` | `:80` |
| DB | `:5437` | `:5432` |
