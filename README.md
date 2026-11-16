# Chess

A two-player chess application with a Go server and a React frontend. Players connect over WebSockets and are paired for a game without creating an account.

## Requirements

- Go 1.23 or later
- Node.js 18 or later and npm

## Run locally

Start the server from the repository root:

```sh
go run ./src/main
```

The server listens on port 8080. Its health endpoint is at `http://localhost:8080/health`.

In a second terminal, start the frontend:

```sh
cd frontend
npm ci
npm run dev
```

Open the URL printed by Vite in two browser windows or tabs. In each, choose **Find an opponent**; this opens the board and joins the matchmaking queue in one step. You can cancel a search before a match is found. Opening or refreshing `/play` directly does not join the queue until you choose **Find an opponent** there.

During a game, select a piece to see its legal moves, then select a destination. Board squares are keyboard-accessible. Pawns promote to queens automatically. Use **Resign game** to end an active game; the final position remains visible. Leaving the play screen during an active game asks for confirmation, because leaving ends the game for both players.

The frontend connects to `ws://localhost:8080/ws` by default. For a deployed backend, set `VITE_WS_HOST` to its hostname when building the frontend; the client connects using `wss://`.

## Development

| Task | Command |
| --- | --- |
| Run Go tests | `go test ./...` |
| Lint frontend | `cd frontend && npm run lint` |
| Build frontend | `cd frontend && npm run build` |
| Preview frontend build | `cd frontend && npm run preview` |

The server keeps matchmaking and games in memory; restarting it loses active state. There is no database, Redis, or Docker Compose setup in this repository.

## Deploy on Render

The included `render.yaml` defines a free Go web service and a static site. Push the repository to GitHub, then create a [Render Blueprint](https://dashboard.render.com/blueprints) from the repository and select the branch containing `render.yaml`. The Blueprint connects the frontend to the server automatically.

Render's free server sleeps after 15 minutes without inbound traffic, including WebSocket messages, and can take about a minute to wake. A restart or sleep ends active games because their state is in memory. Free services have usage limits; without a payment method Render suspends services or builds when applicable limits are reached. See [Render's free-tier details](https://render.com/docs/free).

## Project structure

- `src/main/` — HTTP server and routes.
- `src/ws/` — WebSocket connections.
- `src/game/` — matchmaking, players, and chess games.
- `frontend/` — React, TypeScript, Vite, and Tailwind CSS client; see [frontend/README.md](frontend/README.md).
