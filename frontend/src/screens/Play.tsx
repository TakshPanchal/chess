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
  const [previousMove, setPreviousMove] = useState<{ from: Square; to: Square } | null>(null);


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
          const { from, to } = message.data;
          chess.move({ from, to });
          setPgn(chess.pgn());
          setPreviousMove({ from, to }); // Update previous move
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

  // if (!socket) {
  //   return <div>Connecting...</div>;
  // }

  const startGame = () => {
    setGameStarted(true);
    socket?.send(JSON.stringify({ type: INIT }));
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
      socket?.send(JSON.stringify({ type: MOVE, data: { from, to } }));
    }
    else{
      alert("Please start game first");
    }
  };
  

  return (
    <div className="p-10">
      <div className="grid grid-cols-6 min-w-fit">
        <div className="flex justify-center col-span-4 ">
        <ChessBoard
  pgn={pgn}
  color={color === "white" ? "w" : "b"}
  onMove={onMove}
  previousMove={previousMove} // Pass previous move
/>

        </div>
        <div className="flex flex-col items-center col-span-2 pt-10 bg-slate-950">
  <div className="text-white mb-4">
    {gameStarted ? (
      <div>Game is in progress</div>
    ) : (
      <div></div>
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


