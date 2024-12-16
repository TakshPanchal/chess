import { useEffect, useState, useRef, useCallback } from "react";
import { useParams, useLocation } from "react-router-dom";
import ChessBoard from "../components/ChessBoard";
import { useSocket } from "../hooks/websocket";
import { Chess, Square } from "chess.js";
import Button from "../components/Button";

// Message types
const INIT = "init";
const MOVE = "move";
const GAME_OVER = "over";
const GAME_NOT_STARTED = "game_not_started";
const ERROR = "error";

const PlayPage = () => {
  const { gameId } = useParams();
  const location = useLocation();
  const searchParams = new URLSearchParams(location.search);
  const isPlayer = searchParams.get("play") === "true";
  // Only set spectator mode if not explicitly marked as a player
  const isSpectator = !isPlayer && searchParams.get("spectator") === "true";

  const [gameStarted, setGameStarted] = useState(false);
  const [color, setColor] = useState<"black" | "white" | null>(null);
  const [result, setResult] = useState("*");
  const [previousMove, setPreviousMove] = useState<{ from: Square; to: Square } | null>(null);
  const [isMyTurn, setIsMyTurn] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [shareableLinks, setShareableLinks] = useState<{ play: string, spectate: string } | null>(null);
  const [isConnected, setIsConnected] = useState(false);
  const [isViewerMode, setIsViewerMode] = useState(false);
  const chessBoardRef = useRef<HTMLDivElement>(null);
  const [chess] = useState<Chess>(new Chess());
  const [pgn, setPgn] = useState<string>(chess.pgn());

  // Initialize WebSocket with game ID and spectator mode
  const socket = useSocket({ gameId, isSpectator, isPlayer });

  // Track socket connection state
  useEffect(() => {
    if (!socket) {
      setIsConnected(false);
      return;
    }

    const handleOpen = () => {
      console.log("Socket connected");
      setIsConnected(true);
    };

    const handleClose = () => {
      console.log("Socket disconnected");
      setIsConnected(false);
    };

    socket.addEventListener('open', handleOpen);
    socket.addEventListener('close', handleClose);

    // Check initial state
    setIsConnected(socket.readyState === WebSocket.OPEN);

    return () => {
      socket.removeEventListener('open', handleOpen);
      socket.removeEventListener('close', handleClose);
    };
  }, [socket]);

  // Handle socket messages
  const handleMessage = useCallback((event: MessageEvent) => {
    const message = JSON.parse(event.data);
    console.log("Received message:", message);

    switch (message.type) {
      case INIT: {
        const { color: assignedColor, gameId: newGameId, isSpectator: isViewerMode } = message.data;
        console.log(`Received INIT: color=${assignedColor}, gameId=${newGameId}, isSpectator=${isViewerMode}`);
        setColor(assignedColor);
        setGameStarted(true);
        setIsMyTurn(assignedColor === 'white');
        setIsViewerMode(isViewerMode);

        // Generate shareable links if this is a new game
        if (newGameId) {
          const baseUrl = `${window.location.origin}/play/${newGameId}`;
          setShareableLinks({
            play: `${baseUrl}?play=true`,
            spectate: `${baseUrl}?spectator=true`
          });
        }
        break;
      }
      case MOVE: {
        const { from, to, outcome, turn } = message.data;
        console.log("Move received:", { from, to, outcome, turn });
        
        try {
          chess.move({ from, to });
          setPgn(chess.pgn());
          setPreviousMove({ from, to });
          
          if (outcome === "*") {
            setIsMyTurn(turn === color);
            console.log("Turn update:", {
              serverTurn: turn,
              myColor: color,
              isMyTurn: turn === color
            });
          } else {
            setResult(outcome);
            setGameStarted(false);
          }
        } catch (error) {
          console.error("Invalid move received:", error);
          setErrorMessage("Invalid move received from server");
        }
        break;
      }
      case ERROR: {
        const { message: errMsg } = message.data;
        console.error("Error:", errMsg);
        setErrorMessage(errMsg);
        break;
      }
      case GAME_OVER:
        console.log("Game Over");
        endGame();
        break;
      case GAME_NOT_STARTED:
        setGameStarted(false);
        break;
      default:
        console.log("Unknown message type:", message.type);
    }
  }, [chess, color]);

  // Set up message handler
  useEffect(() => {
    if (!socket) return;

    socket.addEventListener('message', handleMessage);

    return () => {
      socket.removeEventListener('message', handleMessage);
    };
  }, [socket, handleMessage]);

  const startGame = () => {
    if (!socket || !isConnected) {
      setErrorMessage("Not connected to server");
      return;
    }

    console.log("Starting new game");
    chess.reset();
    setPgn(chess.pgn());
    setResult("*");
    setPreviousMove(null);
    setErrorMessage(null);

    const initMessage = { 
      type: INIT,
      data: {} 
    };
    console.log("Sending init message:", initMessage);
    socket.send(JSON.stringify(initMessage));
  };

  const endGame = () => {
    if (!socket) return;

    setGameStarted(false);
    setColor(null);
    setIsMyTurn(false);
    chess.reset();
    setPgn(chess.pgn());
    setPreviousMove(null);
    setErrorMessage(null);
    setShareableLinks(null);
    
    const endMessage = { type: GAME_OVER };
    console.log("Sending end game message:", endMessage);
    socket.send(JSON.stringify(endMessage));
  };

  const onMove = (from: Square, to: Square) => {
    if (!socket || !isConnected) {
      setErrorMessage("Not connected to server");
      return;
    }

    if (!gameStarted) {
      setErrorMessage("Please start game first");
      return;
    }

    if (!isMyTurn) {
      setErrorMessage("It's not your turn!");
      return;
    }

    if (isViewerMode) {
      setErrorMessage("Spectators cannot make moves!");
      return;
    }

    try {
      const moves = chess.moves({ square: from, verbose: true });
      const isValidMove = moves.some(move => move.to === to);
      
      if (!isValidMove) {
        setErrorMessage("Invalid move!");
        return;
      }

      const moveMessage = {
        type: MOVE,
        data: {
          to: to,
          from: from,
          color: color,
          outcome: "*"
        }
      };
      console.log("Sending move:", moveMessage);
      socket.send(JSON.stringify(moveMessage));
    } catch (error) {
      console.error("Move error:", error);
      setErrorMessage("Invalid move!");
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setErrorMessage("Link copied to clipboard!");
    } catch (err) {
      console.error("Failed to copy text: ", err);
      setErrorMessage("Failed to copy link");
    }
  };

  const getGameStatus = () => {
    if (!socket || !isConnected) return "Connecting to server...";
    if (!gameStarted) return "Click 'Start New Game' to begin";
    if (!color) return "Connecting...";
    if (result !== "*") return `Game Over - ${result === "white" ? "White" : "Black"} won!`;
    if (isViewerMode) return `Spectating - ${isMyTurn ? "White" : "Black"}'s turn`;
    return `Playing as ${color} - ${isMyTurn ? "Your turn" : "Opponent's turn"}`;
  };

  // Debug log button conditions
  const shouldShowStartButton = !gameStarted && !gameId && !isViewerMode && isConnected;
  console.log("Start button conditions:", {
    gameStarted,
    gameId,
    isViewerMode,
    isConnected,
    shouldShow: shouldShowStartButton
  });

  return (
    <div className="min-h-screen bg-slate-800 p-4">
      <div className="max-w-7xl mx-auto">
        <div className="grid grid-cols-1 gap-8 md:grid-cols-3">
          {/* Chess Board */}
          <div className="md:col-span-2">
            <div className="flex justify-center" ref={chessBoardRef}>
              <ChessBoard
                pgn={pgn}
                color={color === "white" ? "w" : "b"}
                onMove={onMove}
                previousMove={previousMove}
              />
            </div>
          </div>

          {/* Game Controls */}
          <div className="flex flex-col gap-6 bg-slate-900 p-6 rounded-lg">
            {/* Game Status */}
            <div className="text-white text-center">
              <h2 className="text-2xl font-bold mb-2">Game Status</h2>
              <p className="text-lg">{getGameStatus()}</p>
              {errorMessage && (
                <div className={`mt-4 p-3 rounded ${
                  errorMessage === "Link copied to clipboard!" 
                    ? "bg-green-500/20 text-green-300" 
                    : "bg-red-500/20 text-red-300"
                }`}>
                  {errorMessage}
                </div>
              )}
            </div>

            {/* Game Controls */}
            <div className="flex flex-col gap-4">
              {gameStarted ? (
                <Button onClick={endGame} className="bg-red-600 hover:bg-red-800">
                  End Game
                </Button>
              ) : shouldShowStartButton && (
                <Button onClick={startGame} className="bg-green-600 hover:bg-green-800">
                  Start New Game
                </Button>
              )}
            </div>

            {/* Share Links */}
            {shareableLinks && !isViewerMode && (
              <div className="text-white">
                <h3 className="text-xl font-bold mb-4">Share Game</h3>
                <div className="flex flex-col gap-3">
                  <Button
                    onClick={() => copyToClipboard(shareableLinks.play)}
                    className="bg-blue-600 hover:bg-blue-800"
                  >
                    Copy Play Link
                  </Button>
                  <Button
                    onClick={() => copyToClipboard(shareableLinks.spectate)}
                    className="bg-green-600 hover:bg-green-800"
                  >
                    Copy Spectate Link
                  </Button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default PlayPage;
