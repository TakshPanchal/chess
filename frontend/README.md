# Chess frontend

React and TypeScript client for the chess application, built with Vite and styled with Tailwind CSS. See the [root README](../README.md) for the full application setup.

## Getting started

Install Node.js 18 or later and npm. From this directory:

```sh
npm ci
npm run dev
```

Open the URL printed by Vite. Start the Go server separately as described in the root README; the client connects to `ws://localhost:8080/ws`, configured in `src/hooks/websocket.ts`.

To play locally, open the app in two browser windows or tabs. In each, select **Find an opponent**; this opens the board and joins matchmaking in one step.

## Commands

Run these from `frontend/`:

| Command | Purpose |
| --- | --- |
| `npm run dev` | Start the Vite development server |
| `npm run lint` | Run ESLint |
| `npm run build` | Type-check and build for production |
| `npm run preview` | Preview the production build locally |

## Source layout

- `src/App.tsx` — routes for the landing and play pages.
- `src/screens/` — landing and play screens, including game message handling.
- `src/components/` — chess board and shared UI components.
- `src/hooks/websocket.ts` — WebSocket connection to the Go server.

The client uses `chess.js` for board state. WebSocket message types and payloads must remain compatible with `../src/game/message.go` on the server.
