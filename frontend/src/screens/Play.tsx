import { useEffect, useState } from "react";
import ChessBoard from "../components/ChessBoard";
import { useSocket } from "../hooks/websocket";
import { Chess, Square } from "chess.js";
import Button from "../components/Button";

// TODO: Create Proper structure for Request and response

// MEssage types
const INIT = "init";
const MOVE = "move";
const GAME_OVER = "over";

const PlayPage = () => {
  const socket = useSocket();
  const [chess, setChess] = useState<Chess>(new Chess());
  const [pgn, setPgn] = useState<string>(chess.pgn());
  const [color, setColor] = useState<"black" | "white" | null>(null);
  const [gameStarted, setGameStarted] = useState(false);

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
          chess.move(message.data);
          setPgn(chess.pgn());
          console.log("Incoming Move", message);
          break;
        case GAME_OVER:
          console.log("Game Over", message);
          break;
        default:
          console.log("Unknown message", message);
      }
    };
  }, [socket]);

  if (!socket) {
    return <div>Connecting...</div>;
  }

  const startGame = () => {
    socket.send(JSON.stringify({ type: INIT }));
  };

  const onMove = (from: Square, to: Square) => {
    console.log({ from, to });
    chess.move({ from: from, to: to });

    setPgn(chess.pgn());

    socket.send(JSON.stringify({ type: MOVE, data: { from, to } }));
    console.log({ from, to });
  };

  return (
    <div className="p-10">
      <div className="grid grid-cols-6 min-w-fit">
        <div className="flex justify-center col-span-4 ">
          <ChessBoard
            pgn={pgn}
            color={color == "white" ? "w" : "b"}
            onMove={onMove}
          />
        </div>
        <div className="flex justify-center col-span-2 pt-10 bg-slate-950h">
          {gameStarted && <div>Game is started</div>}
          <div>
            <Button onClick={startGame}>Start Game</Button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PlayPage;
