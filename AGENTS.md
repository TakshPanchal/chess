## Project layout
- `src/main/`: HTTP server entry point; `/health` and `/ws` listen on `$PORT`, default 8080.
- `src/ws/`: WebSocket upgrade and player connection setup.
- `src/game/`: matchmaking, player pumps, chess game state, and JSON message types.
- `frontend/src/screens/` and `frontend/src/components/`: React screens and board UI; `frontend/src/hooks/websocket.ts` connects to the server.
- The WebSocket protocol is represented on both sides in `src/game/message.go` (plus `cancelRequest` in `src/game/player.go`) and `frontend/src/screens/Play.tsx`; keep message types and payloads compatible when changing it.

## Commands
Run from the repository root unless noted. Go 1.23 is specified in `go.mod`; the frontend uses npm and `frontend/package-lock.json`.

| Task | Command |
| --- | --- |
| Start server | `go run ./src/main` |
| Test one Go package | `go test ./src/game` (replace with the affected package) |
| Test all Go packages | `go test ./...` |
| Install frontend dependencies | `cd frontend && npm ci` |
| Start frontend | `cd frontend && npm run dev` |
| Lint frontend file | `cd frontend && npx eslint src/screens/Play.tsx` (replace with the affected file) |
| Build frontend | `cd frontend && npm run build` |

## Notes
- `frontend/src/hooks/websocket.ts` connects to `wss://$VITE_WS_HOST/ws` when `VITE_WS_HOST` is set at build time, else `ws://localhost:8080/ws`.
- `render.yaml` deploys both services; the frontend gets `VITE_WS_HOST` from the server's Render hostname.
- Games and matchmaking live in server memory; a restart ends active games.

## External References
| Need | File |
| --- | --- |
| Setup and deploy | `README.md` |
| Frontend commands and layout | `frontend/README.md` |
