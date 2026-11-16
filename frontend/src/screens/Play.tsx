import { useCallback, useEffect, useRef, useState } from "react";
import { Link, useBlocker, useLocation, useNavigate } from "react-router-dom";
import { Chess, Square } from "chess.js";
import AppHeader from "../components/AppHeader";
import ChessBoard from "../components/ChessBoard";
import { useSocket } from "../hooks/websocket";

type Phase = "ready" | "waiting" | "playing" | "finished";
type PieceColor = "white" | "black";
type Move = { from: Square; to: Square };
type Result = { title: string; detail: string };
type Tone = "active" | "neutral" | "alert";

const COLOR_LABELS: Record<PieceColor, string> = { white: "White", black: "Black" };

const describeOutcome = (game: Chess, outcome: string, color: PieceColor | null): Result => {
  if (outcome === "draw") {
    const detail = game.isStalemate() ? "Stalemate."
      : game.isInsufficientMaterial() ? "Draw by insufficient material."
      : game.isThreefoldRepetition() ? "Draw by threefold repetition."
      : "The game ended in a draw.";
    return { title: "Game drawn", detail };
  }
  const winner = outcome === "white" ? "white" : "black";
  return {
    title: winner === color ? "You won" : "You lost",
    detail: `${game.isCheckmate() ? "Checkmate. " : ""}${COLOR_LABELS[winner]} wins.`,
  };
};

const PlayerRow = ({ name, pieceColor, detail, toMove }: {
  name: string;
  pieceColor: PieceColor | null;
  detail: string;
  toMove: boolean;
}) => (
  <div className="player-row">
    <span className={`player-swatch player-swatch--${pieceColor ?? "unknown"}`} aria-hidden="true" />
    <span className="player-text">
      <span className="player-name">{name}</span>
      <span className="player-detail">{detail}</span>
    </span>
    {toMove && <span className="player-turn">To move</span>}
  </div>
);

const PlayPage = () => {
  const location = useLocation();
  const navigate = useNavigate();
  const { socket, connection, reconnect } = useSocket();
  const [chess, setChess] = useState(() => new Chess());
  const chessRef = useRef(chess);
  const [phase, setPhase] = useState<Phase>("ready");
  const [color, setColor] = useState<PieceColor | null>(null);
  const colorRef = useRef<PieceColor | null>(null);
  const [previousMove, setPreviousMove] = useState<Move | null>(null);
  const previousMoveRef = useRef<Move | null>(null);
  const [result, setResult] = useState<Result | null>(null);
  const [searchRequested, setSearchRequested] = useState(
    () => (location.state as { findOpponent?: boolean } | null)?.findOpponent === true,
  );
  const [confirmResign, setConfirmResign] = useState(false);
  const resigned = useRef(false);
  const [notice, setNotice] = useState("");
  const [movesOpen, setMovesOpen] = useState(false);
  const moveListRef = useRef<HTMLOListElement>(null);
  const keepPlayingRef = useRef<HTMLButtonElement>(null);

  const activeGame = phase === "playing" && connection === "connected";
  const blocker = useBlocker(({ currentLocation, nextLocation }) =>
    activeGame && currentLocation.pathname !== nextLocation.pathname);

  const resetGame = useCallback((nextColor: PieceColor | null) => {
    const fresh = new Chess();
    chessRef.current = fresh;
    setChess(fresh);
    colorRef.current = nextColor;
    setColor(nextColor);
    previousMoveRef.current = null;
    setPreviousMove(null);
    setResult(null);
    setConfirmResign(false);
    resigned.current = false;
  }, []);

  // The welcome screen's request is consumed once; a refresh of /play must not search again.
  useEffect(() => {
    if (location.state) navigate(location.pathname, { replace: true, state: null });
  }, [location.pathname, location.state, navigate]);

  useEffect(() => {
    if (!searchRequested || connection !== "connected" || !socket) return;
    setSearchRequested(false);
    resetGame(null);
    setPhase("waiting");
    socket.send(JSON.stringify({ type: "init" }));
  }, [searchRequested, connection, socket, resetGame]);

  useEffect(() => {
    if (!socket) return;
    socket.onmessage = (event) => {
      let message;
      try {
        message = JSON.parse(event.data);
      } catch {
        setNotice("The server sent an unexpected response.");
        return;
      }
      if (message.type === "init") {
        resetGame(message.data.color);
        setNotice("");
        setPhase("playing");
      } else if (message.type === "move") {
        const { from, to, outcome } = message.data;
        const previous = previousMoveRef.current;
        const alreadyPlayed = previous?.from === from && previous?.to === to && !chessRef.current.get(from);
        if (!alreadyPlayed) {
          const next = new Chess();
          next.loadPgn(chessRef.current.pgn());
          try {
            next.move({ from, to, promotion: "q" });
          } catch {
            setNotice("The game position could not be updated.");
            return;
          }
          chessRef.current = next;
          setChess(next);
        }
        previousMoveRef.current = { from, to };
        setPreviousMove({ from, to });
        if (outcome !== "*") {
          setResult(describeOutcome(chessRef.current, outcome, colorRef.current));
          setConfirmResign(false);
          setPhase("finished");
        }
      } else if (message.type === "over") {
        if (!resigned.current) {
          setResult({ title: "Game ended", detail: "Your opponent resigned or left the game." });
        }
        setConfirmResign(false);
        setPhase("finished");
      } else if (message.type === "ERROR" || message.type === "game_not_started") {
        setNotice("The server didn't accept that move. Start a new game if the board stops updating.");
      }
    };
    return () => { socket.onmessage = null; };
  }, [socket, resetGame]);

  useEffect(() => {
    if (!activeGame) return;
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", warn);
    return () => window.removeEventListener("beforeunload", warn);
  }, [activeGame]);

  useEffect(() => {
    if (blocker.state === "blocked" && !activeGame) blocker.reset();
  }, [blocker, activeGame]);

  const confirming = blocker.state === "blocked" ? "leave" : confirmResign ? "resign" : null;
  useEffect(() => {
    if (confirming) keepPlayingRef.current?.focus();
  }, [confirming]);

  const history = chess.history();
  useEffect(() => {
    const list = moveListRef.current;
    if (list) list.scrollTop = list.scrollHeight;
  }, [history.length]);

  const requestSearch = () => {
    setNotice("");
    setSearchRequested(true);
    // The server ignores `init` from a socket still attached to a game and closes it once the game ends,
    // so a rematch always waits for a new connection; the request is kept until that socket opens.
    if (connection === "disconnected" || phase === "finished" || phase === "playing") reconnect();
  };

  const cancelSearch = () => {
    if (searchRequested) {
      setSearchRequested(false);
      return;
    }
    socket?.send(JSON.stringify({ type: "cancel" }));
    setPhase("ready");
  };

  const resign = () => {
    resigned.current = true;
    socket?.send(JSON.stringify({ type: "over" }));
    setResult({ title: "You resigned", detail: "Your opponent wins by resignation." });
    setConfirmResign(false);
    setPhase("finished");
  };

  const leave = () => {
    socket?.send(JSON.stringify({ type: "over" }));
    blocker.proceed?.();
  };

  const onMove = (from: Square, to: Square) => {
    if (!socket || !activeGame || chessRef.current.turn() !== color?.[0]) return;
    const next = new Chess();
    next.loadPgn(chessRef.current.pgn());
    try {
      next.move({ from, to, promotion: "q" });
    } catch {
      setNotice("That move isn't available.");
      return;
    }
    chessRef.current = next;
    setChess(next);
    previousMoveRef.current = { from, to };
    setPreviousMove({ from, to });
    setNotice("");
    socket.send(JSON.stringify({ type: "move", data: { from, to, color } }));
  };

  const sideToMove: PieceColor = chess.turn() === "w" ? "white" : "black";
  const isMyTurn = activeGame && sideToMove === color;
  const opponentColor: PieceColor | null = color === "white" ? "black" : color === "black" ? "white" : null;

  let status: { title: string; detail: string; tone: Tone };
  let action: { label: string; kind: "primary" | "secondary"; onClick: () => void } | null;
  // A pending search request survives a failed connection, so "Try again" resumes the search.
  if (connection === "disconnected" && (searchRequested || phase === "ready" || phase === "waiting")) {
    status = { title: "Can't reach the game server", detail: "Check your connection, then try again.", tone: "alert" };
    action = { label: "Try again", kind: "primary", onClick: () => { setPhase("ready"); reconnect(); } };
  } else if (searchRequested) {
    status = { title: "Connecting…", detail: "You'll start searching as soon as the connection is ready.", tone: "neutral" };
    action = { label: "Cancel", kind: "secondary", onClick: cancelSearch };
  } else if (phase === "finished" && result) {
    status = { ...result, tone: "neutral" };
    action = { label: "Find another opponent", kind: "primary", onClick: requestSearch };
  } else if (phase === "playing" && connection !== "connected") {
    status = { title: "Connection lost", detail: "This game has ended because the connection closed.", tone: "alert" };
    action = { label: "Find a new opponent", kind: "primary", onClick: requestSearch };
  } else if (phase === "waiting") {
    status = { title: "Finding an opponent", detail: "You'll be matched with the next player who is searching.", tone: "active" };
    action = { label: "Cancel search", kind: "secondary", onClick: cancelSearch };
  } else if (phase === "playing" && color) {
    const check = chess.inCheck() ? (isMyTurn ? " Your king is in check." : " Their king is in check.") : "";
    status = isMyTurn
      ? { title: "Your turn", detail: `You're playing ${color}.${check} Select a piece to see its moves.`, tone: "active" }
      : { title: "Opponent's turn", detail: `You're playing ${color}.${check} Waiting for their move.`, tone: "neutral" };
    action = null;
  } else {
    status = {
      title: connection === "connecting" ? "Connecting…" : "Ready to play",
      detail: "Find an opponent to start a game. Colors are assigned at random.",
      tone: "neutral",
    };
    action = { label: "Find an opponent", kind: "primary", onClick: requestSearch };
  }

  const showTurn = phase === "playing" && connection === "connected";

  return (
    <div className="page">
      <AppHeader>
        <Link className="header-link" to="/">Home</Link>
      </AppHeader>
      <main className="play">
        <section className="play-board" aria-label="Board">
          <PlayerRow
            name="Opponent"
            pieceColor={opponentColor}
            detail={opponentColor ? COLOR_LABELS[opponentColor] : phase === "waiting" ? "Searching…" : "Not matched yet"}
            toMove={showTurn && sideToMove === opponentColor}
          />
          <ChessBoard
            chess={chess}
            color={color === "black" ? "b" : "w"}
            interactive={isMyTurn}
            onMove={onMove}
            previousMove={previousMove}
          />
          <PlayerRow
            name="You"
            pieceColor={color}
            detail={color ? COLOR_LABELS[color] : "Color assigned when matched"}
            toMove={showTurn && sideToMove === color}
          />
        </section>
        <div className="play-panel">
          <div className="play-status" aria-live="polite">
            <h1 className="status-title">
              <span className={`status-dot status-dot--${status.tone}`} aria-hidden="true" />
              {status.title}
            </h1>
            <p className="status-detail">{status.detail}</p>
            {notice && <p className="notice" role="alert">{notice}</p>}
          </div>
          {action && (
            <div className="play-actions">
              <button type="button" className={`button button--${action.kind} button--full`} onClick={action.onClick}>
                {action.label}
              </button>
            </div>
          )}
          <section className={`moves${movesOpen ? " moves--open" : ""}`} aria-labelledby="moves-heading">
            <div className="moves-header">
              <h2 id="moves-heading" className="moves-title">Moves</h2>
              <button
                type="button"
                className="moves-toggle"
                aria-expanded={movesOpen}
                aria-controls="move-list"
                onClick={() => setMovesOpen((open) => !open)}
              >
                {movesOpen ? "Hide" : "Show"} moves ({history.length})
              </button>
            </div>
            <div id="move-list" className="moves-body">
              {history.length === 0 ? (
                <p className="moves-empty">No moves yet.</p>
              ) : (
                <ol className="move-list" ref={moveListRef} tabIndex={0} aria-label="Move history">
                  {Array.from({ length: Math.ceil(history.length / 2) }, (_, i) => (
                    <li key={i} className="move-row">
                      <span className="move-number">{i + 1}.</span>
                      <span className="move-san">{history[i * 2]}</span>
                      <span className="move-san">{history[i * 2 + 1] ?? ""}</span>
                    </li>
                  ))}
                </ol>
              )}
            </div>
          </section>
          {(activeGame || confirming) && (
            <div className="game-actions">
              {confirming ? (
                <div className="confirm" role="group" aria-labelledby="confirm-title">
                  <p id="confirm-title" className="confirm-title">
                    {confirming === "leave" ? "Leave this game?" : "Resign this game?"}
                  </p>
                  <p className="confirm-detail">
                    {confirming === "leave" ? "Leaving ends the game for both players." : "Your opponent will win the game."}
                  </p>
                  <div className="confirm-buttons">
                    <button
                      ref={keepPlayingRef}
                      type="button"
                      className="button button--secondary"
                      onClick={() => (confirming === "leave" ? blocker.reset?.() : setConfirmResign(false))}
                    >
                      Keep playing
                    </button>
                    <button type="button" className="button button--danger" onClick={confirming === "leave" ? leave : resign}>
                      {confirming === "leave" ? "Leave game" : "Resign"}
                    </button>
                  </div>
                </div>
              ) : (
                <button type="button" className="button button--quiet button--full" onClick={() => setConfirmResign(true)}>
                  Resign game
                </button>
              )}
            </div>
          )}
        </div>
      </main>
    </div>
  );
};

export default PlayPage;
