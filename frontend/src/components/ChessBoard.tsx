import { Chess, PieceSymbol, Square } from "chess.js";
import { useEffect, useState } from "react";
import "./ChessBoard.css";

type BoardColor = "w" | "b";

type ChessBoardProps = {
  chess: Chess;
  color: BoardColor;
  interactive: boolean;
  onMove: (from: Square, to: Square) => void;
  previousMove: { from: Square; to: Square } | null;
};

const FILES = ["a", "b", "c", "d", "e", "f", "g", "h"] as const;
const RANKS = ["8", "7", "6", "5", "4", "3", "2", "1"] as const;

const PIECE_NAMES: Record<PieceSymbol, string> = {
  p: "pawn",
  n: "knight",
  b: "bishop",
  r: "rook",
  q: "queen",
  k: "king",
};

const COLOR_NAMES: Record<BoardColor, string> = {
  w: "white",
  b: "black",
};

const ChessBoard = ({
  chess,
  color,
  interactive,
  onMove,
  previousMove,
}: ChessBoardProps) => {
  const [selectedSquare, setSelectedSquare] = useState<Square | null>(null);
  const position = chess.fen();
  const files = color === "w" ? FILES : [...FILES].reverse();
  const ranks = color === "w" ? RANKS : [...RANKS].reverse();
  const legalDestinations = new Set(
    selectedSquare
      ? chess
          .moves({ square: selectedSquare, verbose: true })
          .map((move) => move.to)
      : []
  );
  const checkedKing = chess.inCheck()
    ? chess.board().flat().find((piece) => piece?.type === "k" && piece.color === chess.turn())?.square
    : undefined;

  useEffect(() => {
    setSelectedSquare(null);
  }, [interactive, position]);

  const handleSquareClick = (square: Square) => {
    if (!interactive) {
      return;
    }

    const piece = chess.get(square);

    if (selectedSquare === null) {
      if (piece?.color === color) {
        setSelectedSquare(square);
      }
      return;
    }

    if (selectedSquare === square) {
      setSelectedSquare(null);
      return;
    }

    if (legalDestinations.has(square)) {
      setSelectedSquare(null);
      onMove(selectedSquare, square);
      return;
    }

    setSelectedSquare(piece?.color === color ? square : null);
  };

  return (
    <section
      className="chess-board"
      aria-label={`Chess board, ${COLOR_NAMES[color]} perspective`}
    >
      <div className="chess-board__grid">
        {ranks.flatMap((rank, rankIndex) =>
          files.map((file, fileIndex) => {
            const square = `${file}${rank}` as Square;
            const piece = chess.get(square);
            const isLightSquare = (fileIndex + rankIndex) % 2 === 0;
            const isSelected = selectedSquare === square;
            const isLegalDestination = legalDestinations.has(square);
            const isPreviousMove =
              previousMove?.from === square || previousMove?.to === square;
            const squareColor = isLightSquare ? "light" : "dark";
            const pieceDescription = piece
              ? `${COLOR_NAMES[piece.color]} ${PIECE_NAMES[piece.type]}`
              : "empty";
            const isCheckedKing = square === checkedKing;
            const label = `${squareColor} square ${square}, ${pieceDescription}${
              isCheckedKing ? ", in check" : ""
            }${
              isSelected
                ? ", selected"
                : isLegalDestination
                  ? ", legal move destination"
                  : ""
            }`;

            return (
              <button
                key={square}
                type="button"
                className={[
                  "chess-board__square",
                  isLightSquare
                    ? "chess-board__square--light"
                    : "chess-board__square--dark",
                  isSelected ? "chess-board__square--selected" : "",
                  isLegalDestination
                    ? "chess-board__square--legal"
                    : "",
                  isLegalDestination && piece
                    ? "chess-board__square--capture"
                    : "",
                  isPreviousMove ? "chess-board__square--previous" : "",
                  isCheckedKing ? "chess-board__square--check" : "",
                ]
                  .filter(Boolean)
                  .join(" ")}
                aria-label={label}
                aria-pressed={isSelected}
                disabled={!interactive}
                onClick={() => handleSquareClick(square)}
              >
                {fileIndex === 0 && (
                  <span
                    aria-hidden="true"
                    className="chess-board__coordinate chess-board__coordinate--rank"
                  >
                    {rank}
                  </span>
                )}
                {rankIndex === ranks.length - 1 && (
                  <span
                    aria-hidden="true"
                    className="chess-board__coordinate chess-board__coordinate--file"
                  >
                    {file}
                  </span>
                )}
                {piece && (
                  <img
                    className="chess-board__piece"
                    src={`/${piece.type}${piece.color === "w" ? "_w" : ""}.png`}
                    alt=""
                  />
                )}
              </button>
            );
          })
        )}
      </div>
    </section>
  );
};

export default ChessBoard;
