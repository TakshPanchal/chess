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
      "flex items-center justify-center w-20 h-20 cursor-pointer active:outline-none active:ring active:ring-violet-300 " +
      (i % 2 === j % 2 ? "bg-board-white" : "bg-board-black");

    className += from === piece?.square ? " active" : "";
    const rankColor = i % 2 !== 0 ? "text-board-white" : "text-board-black";

    return (
      <div key={j} className="relative">
        {j == 0 && (
          <div className={"absolute z-10 left-1 text-lg " + rankColor}>
            {rows[i]}
          </div>
        )}
        <div
          className={className + " z-0"}
          onClick={
            piece == null || piece?.color === color ? onClick : undefined
          }
        >
          {piece ? (
            <img
              src={`/${piece.type}${piece.color === "w" ? "_w" : ""}.png`}
              className="w-16"
            />
          ) : (
            ""
          )}
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
        {cols.split("").map((col, i) => {
          const classNames =
            "text-lg " + (i % 2 == 0 ? "text-board-white" : "text-board-black");
          return (
            <div className="w-20" key={col}>
              <div className={classNames}>{col}</div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default ChessBoard;
