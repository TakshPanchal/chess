import { Chess, Color, PieceSymbol, Square } from "chess.js";
import { useState } from "react";

const ChessBoard = ({
  pgn,
  color,
  onMove,
}: {
  pgn: string;
  color: Color;
  onMove: (from: Square, to: Square) => void;
}) => {
  const chess = new Chess();
  chess.loadPgn(pgn);
  let board = chess.board();
  let cols = "abcdefgh";
  let rows = "87654321";

  if (color === "b") {
    cols = cols.split("").reverse().join("");
    rows = rows.split("").reverse().join("");

    board = board.map((row) => row.reverse());
    board = board.reverse();
  }

  const [from, setFrom] = useState<Square | null>(null);
  const [to, setTo] = useState<Square | null>(null);

  // build square
  const square = (
    i: number,
    j: number,
    piece: { square: Square; type: PieceSymbol; color: Color } | null,
    onClick: () => void = () => {}
  ) => {
    let className =
      "flex items-center justify-center w-16 h-16 border border-gray-800 cursor-pointer active:outline-none active:ring active:ring-violet-300 " +
      (i % 2 === j % 2 ? "bg-gray-300" : "bg-gray-500");

    className += from === piece?.square ? " active" : "";

    return (
      <div key={j} className="relative">
        {j == 0 && <div className="absolute z-10 left-1">{rows[i]}</div>}
        <div
          className={className + " z-0"}
          onClick={
            piece == null || piece?.color === color ? onClick : undefined
          }
        >
          {piece
            ? piece.color === "w"
              ? piece.type.toUpperCase()
              : piece.type
            : ""}
        </div>
      </div>
    );
  };

  return (
    <div className="">
      {board.map((row, i) => {
        return (
          <div key={i} className="flex select-none">
            {row.map((piece, j) => {
              return square(i, j, piece, () => {
                console.log("click", piece);
                console.log("from", from);

                if (!from) {
                  setFrom(piece?.square || null);
                  // light up the possible moves square
                } else {
                  const to = cols[j] + rows[i];
                  setTo(to as Square);
                  try {
                    onMove(from, to as Square);
                  } catch (error) {
                    // send invalid move UI
                    console.log(error);
                  } finally {
                    setFrom(null);
                    setTo(null);
                  }
                }
              });
            })}
          </div>
        );
      })}

      <div className="relative flex text-right white -top-6 right-1 text-black-200 ">
        {cols.split("").map((col) => {
          return (
            <div className="w-16" key={col}>
              {col}
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ChessBoard;
