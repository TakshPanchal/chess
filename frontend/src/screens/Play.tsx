import { useEffect, useState, useRef } from "react";
import ChessBoard from "../components/ChessBoard";
import { useSocket } from "../hooks/websocket";
import { Chess, Square } from "chess.js";
import Button from "../components/Button";

// TODO: Create Proper structure for Request and response

// MEssage types
const INIT = "init";
const MOVE = "move";
const GAME_OVER = "over";
const GAME_NOT_STARTED = "game_not_started";

const PlayPage = () => {
  const socket = useSocket();
  const [chess, setChess] = useState<Chess>(new Chess());
  const [pgn, setPgn] = useState<string>(chess.pgn());
  const [color, setColor] = useState<"black" | "white" | null>(null);
  const [gameStarted, setGameStarted] = useState(false);
  const [result, setResult] = useState("*");
  const [previousMove, setPreviousMove] = useState<{ from: Square; to: Square } | null>(null);
  const chessBoardRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!socket) return;

    socket.onmessage = (event) => {
      console.log(event);
      const message = JSON.parse(event.data);
      console.log(message);

      switch (message.type) {
        case INIT: {
          const { _, color } = message.data;
          setColor(color);
          console.log("Games is Started !!");
          setGameStarted(true);
          break;
        }
        case MOVE:
          const { from, to, outcome } = message.data;
          if(outcome == "*") {
            chess.move({ from, to });
          } 
          else {
            setResult(outcome);
          }
          setPgn(chess.pgn());
          setPreviousMove({ from, to }); // Update previous move
          console.log("Incoming Move", message);
          break;
        case GAME_OVER:
          console.log("Game Over", message);
          endGame();
          break;
        case GAME_NOT_STARTED:
          setGameStarted(false);
          break;
        default:
          console.log("Unknown message", message);
      }
    };
  }, [socket]);


  const startGame = () => {
    if(socket) {
      setGameStarted(true);
      socket?.send(JSON.stringify({ type: INIT }));
    }
  };
  const endGame = () => {
    setGameStarted(false);
    let newChess = new Chess();
    setPgn(newChess.pgn());
    setPreviousMove(null);
    socket?.send(JSON.stringify({type: GAME_OVER}))
  }
  const onMove = (from: Square, to: Square) => {
    if(gameStarted) {
      chess.move({ from, to });
      setPgn(chess.pgn());
      setPreviousMove({ from, to }); // Update previous move
      socket?.send(JSON.stringify({ type: MOVE, data: { from, to, color } }));
    }
    else{
      alert("Please start game first");
    }
  };
  

  return (
    <div className="p-4" >
    <div className="grid grid-cols-1 gap-4 md:grid-cols-6">
      <div className="md:col-span-4">
        <div className="flex justify-center" ref={chessBoardRef}>
          <ChessBoard
            pgn={pgn}
            color={color === "white" ? "w" : "b"}
            onMove={onMove}
            previousMove={previousMove} // Pass previous move
          />
        </div>
      </div>
      <div className="md:col-span-2 flex flex-col items-center pt-4 bg-slate-950" style={{ maxWidth: "100%", overflowX: "hidden" }}>
        <div className="text-white mb-4">
          {!gameStarted ? (
            <div></div>
          ) : (
            <div>Game is in progress</div>
          )}
        </div>
        <div className="mb-4">
          {gameStarted ? (
            <Button onClick={endGame}>End Game</Button>
          ) : (
            <Button onClick={startGame}>Start Game</Button>
          )}
          {
            result != "*"  ? "${result}"  : "no one won"
          }
        </div>
      </div>
    </div>
  </div>
  
  );
};

export default PlayPage;


