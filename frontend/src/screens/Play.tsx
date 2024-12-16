import { useEffect, useState, useRef } from "react";
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
  const socket = useSocket();
  const [gameStarted, setGameStarted] = useState(false);
  const [color, setColor] = useState<"black" | "white" | null>(null);
  const [result, setResult] = useState("*");
  const [previousMove, setPreviousMove] = useState<{ from: Square; to: Square } | null>(null);
  const [isMyTurn, setIsMyTurn] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const chessBoardRef = useRef<HTMLDivElement>(null);
  const [chess] = useState<Chess>(new Chess());
  const [pgn, setPgn] = useState<string>(chess.pgn());

  // Clear error message after 3 seconds
  useEffect(() => {
    if (errorMessage) {
      const timer = setTimeout(() => setErrorMessage(null), 3000);
      return () => clearTimeout(timer);
    }
  }, [errorMessage]);

  useEffect(() => {
    if (!socket) return;

    socket.onmessage = (event) => {
      const message = JSON.parse(event.data);
      console.log("Received message:", message);

      switch (message.type) {
        case INIT: {
          const { color: assignedColor } = message.data;
          setColor(assignedColor);
          console.log(`Game Started! You are playing as ${assignedColor}`);
          setGameStarted(true);
          // White moves first
          setIsMyTurn(assignedColor === 'white');
          break;
        }
        case MOVE: {
          const { from, to, outcome, turn } = message.data;
          console.log("Move received:", { from, to, outcome, turn });
          
          try {
            // Make the move on the chess instance
            chess.move({ from, to });
            setPgn(chess.pgn());
            setPreviousMove({ from, to });
            
            if (outcome === "*") {
              // Set turn based on server's turn information
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
    };
  }, [socket, chess, color]);

  const startGame = () => {
    if(socket) {
      chess.reset();
      setPgn(chess.pgn());
      setGameStarted(true);
      setResult("*");
      setPreviousMove(null);
      setErrorMessage(null);
      const initMessage = { 
        type: INIT,
        data: {} // Match backend InitData struct
      };
      console.log("Sending init message:", initMessage);
      socket.send(JSON.stringify(initMessage));
    }
  };

  const endGame = () => {
    setGameStarted(false);
    setColor(null);
    setIsMyTurn(false);
    chess.reset();
    setPgn(chess.pgn());
    setPreviousMove(null);
    setErrorMessage(null);
    const endMessage = { type: GAME_OVER };
    console.log("Sending end game message:", endMessage);
    socket?.send(JSON.stringify(endMessage));
  };

  const onMove = (from: Square, to: Square) => {
    if (!gameStarted) {
      setErrorMessage("Please start game first");
      return;
    }

    if (!isMyTurn) {
      setErrorMessage("It's not your turn!");
      return;
    }

    try {
      // Validate move locally without making it
      const moves = chess.moves({ square: from, verbose: true });
      const isValidMove = moves.some(move => move.to === to);
      
      if (!isValidMove) {
        setErrorMessage("Invalid move!");
        return;
      }

      // Send move to server
      const moveMessage = {
        type: MOVE,
        data: {
          to: to,
          from: from,
          color: color, // Matches PlayerType in backend (black/white)
          outcome: "*"  // Default outcome, server will update if game ends
        }
      };
      console.log("Sending move:", moveMessage);
      socket?.send(JSON.stringify(moveMessage));
      
      // Don't update local state until server confirms
    } catch (error) {
      console.error("Move error:", error);
      setErrorMessage("Invalid move!");
    }
  };

  const getGameStatus = () => {
    if (!gameStarted) return "Waiting for game to start...";
    if (!color) return "Connecting...";
    if (result !== "*") return `Game Over - ${result === "white" ? "White" : "Black"} won!`;
    return `Playing as ${color} - ${isMyTurn ? "Your turn" : "Opponent's turn"}`;
  };

  return (
    <div className="p-4">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-6">
        <div className="md:col-span-4">
          <div className="flex justify-center" ref={chessBoardRef}>
            <ChessBoard
              pgn={pgn}
              color={color === "white" ? "w" : "b"}
              onMove={onMove}
              previousMove={previousMove}
            />
          </div>
        </div>
        <div className="md:col-span-2 flex flex-col items-center pt-4 bg-slate-950" style={{ maxWidth: "100%", overflowX: "hidden" }}>
          <div className="text-white mb-4 text-center">
            <div className="text-xl font-bold mb-2">Game Status</div>
            <div>{getGameStatus()}</div>
            {errorMessage && (
              <div className="text-red-500 mt-2">{errorMessage}</div>
            )}
          </div>
          <div className="mb-4">
            {gameStarted ? (
              <Button onClick={endGame}>End Game</Button>
            ) : (
              <Button onClick={startGame}>Start Game</Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default PlayPage;
