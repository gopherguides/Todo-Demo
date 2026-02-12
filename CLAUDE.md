# Todo-Demo

Kanban-style TODO app built with Go + Echo v4, Templ, Tailwind CSS v4, HTMX, SortableJS, SQLite, and Clerk auth.

## Quick Start

```bash
make setup      # install deps
make generate   # run sqlc + templ
make dev        # start with hot reload (air)
```

## Stack

- **Go** with Echo v4 web framework
- **Templ** for type-safe HTML templates
- **Tailwind CSS v4** (CLI, no PostCSS)
- **HTMX** for server-driven interactivity
- **SortableJS** for drag-and-drop
- **SQLite** via modernc.org/sqlite (CGO-free)
- **sqlc** for type-safe SQL queries
- **goose** for database migrations
- **Clerk** for authentication (cookie-based sessions)
- **Air** for hot reload during development

## Project Structure

- `cmd/server/` - application entry point
- `internal/config/` - environment configuration
- `internal/database/` - DB init, migrations, generated sqlc code
- `internal/handler/` - HTTP handlers and route registration
- `internal/middleware/` - Echo middleware (auth, logging, etc.)
- `internal/ctxkeys/` - typed context keys
- `internal/meta/` - page metadata helpers
- `templates/` - Templ templates (layouts, pages, components)
- `static/` - CSS and JS assets
- `sqlc/` - sqlc config and SQL queries

## Code Generation

After modifying `.templ` files or `sqlc/queries/*.sql`:

```bash
make generate
```

## Database

Migrations are in `internal/database/migrations/`. Run with:

```bash
make migrate
```

## Build Errors

Always check `tmp/air-combined.log` for build errors during development.

## Conventions

- Conventional commits: `feat(scope): subject`
- No comments unless explicitly requested
- Follow existing patterns in the codebase
- All code, comments, and docs in English
