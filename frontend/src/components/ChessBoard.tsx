import { Chess, Color, PieceSymbol, Square } from "chess.js";
import { useState, useEffect, useMemo } from "react";

const ChessBoard = ({
  pgn,
  color,
  onMove,
  previousMove,
}: {
  pgn: string;
  color: Color;
  onMove: (from: Square, to: Square) => void;
  previousMove: { from: Square; to: Square } | null;
}) => {
  const [selectedSquare, setSelectedSquare] = useState<Square | null>(null);
  const [validMoves, setValidMoves] = useState<Square[]>([]);

  // Memoize chess instance and board setup
  const { chess, board, cols, rows } = useMemo(() => {
    const chessInstance = new Chess();
    chessInstance.loadPgn(pgn);
    let boardSetup = chessInstance.board();
    let colsSetup = "abcdefgh";
    let rowsSetup = "87654321";

    // Rotate board based on player's color
    if (color === "b") {
      colsSetup = colsSetup.split("").reverse().join("");
      rowsSetup = rowsSetup.split("").reverse().join("");
      boardSetup = boardSetup.map((row) => row.reverse());
      boardSetup = boardSetup.reverse();
    }

    return {
      chess: chessInstance,
      board: boardSetup,
      cols: colsSetup,
      rows: rowsSetup,
    };
  }, [pgn, color]); // Only recalculate when pgn or color changes

  // Calculate valid moves when a piece is selected
  useEffect(() => {
    if (selectedSquare) {
      try {
        const moves = chess.moves({ square: selectedSquare, verbose: true });
        setValidMoves(moves.map(move => move.to as Square));
      } catch (error) {
        console.error("Error calculating valid moves:", error);
        setValidMoves([]);
      }
    } else {
      setValidMoves([]);
    }
  }, [selectedSquare, pgn]); // Depend on pgn instead of chess instance

  const handleSquareClick = (
    i: number,
    j: number,
    piece: { square: Square; type: PieceSymbol; color: Color } | null
  ) => {
    const clickedSquare = cols[j] + rows[i] as Square;

    // If no square is selected yet
    if (!selectedSquare) {
      // Can only select pieces of your color
      if (piece && piece.color === color) {
        setSelectedSquare(clickedSquare);
      }
      return;
    }

    // If clicking the same square, deselect it
    if (clickedSquare === selectedSquare) {
      setSelectedSquare(null);
      return;
    }

    // If clicking a different square
    if (validMoves.includes(clickedSquare)) {
      // Make the move
      onMove(selectedSquare, clickedSquare);
      setSelectedSquare(null);
    } else if (piece && piece.color === color) {
      // If clicking another piece of same color, select it instead
      setSelectedSquare(clickedSquare);
    } else {
      // Invalid move, deselect
      setSelectedSquare(null);
    }
  };

  // Memoize the isValidMove function
  const isValidMove = useMemo(() => {
    return (square: string): boolean => validMoves.includes(square as Square);
  }, [validMoves]);

  // Memoize the square component generator
  const squareComponent = useMemo(() => {
    return (
      i: number,
      j: number,
      piece: { square: Square; type: PieceSymbol; color: Color } | null
    ) => {
      const squareName = cols[j] + rows[i];
      let className = "flex items-center justify-center w-20 h-20 cursor-pointer " +
        "active:outline-none active:ring active:ring-violet-300 " +
        (i % 2 === j % 2 ? "bg-board-white" : "bg-board-black");

      // Highlight logic
      if (selectedSquare === squareName) {
        // Selected piece
        className += " bg-blue-400 bg-opacity-50";
      } else if (isValidMove(squareName)) {
        // Valid move square
        className += " bg-green-400 bg-opacity-30";
      } else if (previousMove && (previousMove.from === squareName || previousMove.to === squareName)) {
        // Previous move highlight
        className += " bg-yellow-500 bg-opacity-40";
      }

      // Rank numbers (1-8)
      const rankColor = i % 2 !== 0 ? "text-board-white" : "text-board-black";

      return (
        <div key={j} className="relative">
          {j === 0 && (
            <div className={`absolute z-10 left-1 text-lg ${rankColor}`}>
              {rows[i]}
            </div>
          )}
          <div 
            className={className + " z-0 relative"} 
            onClick={() => handleSquareClick(i, j, piece)}
          >
            {piece && (
              <img
                src={`/${piece.type}${piece.color === "w" ? "_w" : ""}.png`}
                className={`w-16 ${piece.color === color ? 'cursor-pointer' : ''}`}
                alt={`${piece.color} ${piece.type}`}
              />
            )}
            {isValidMove(squareName) && (
              <div className={`absolute inset-0 flex items-center justify-center ${piece ? 'bg-red-500 bg-opacity-30' : ''}`}>
                {!piece && <div className="w-4 h-4 rounded-full bg-green-500 bg-opacity-50" />}
              </div>
            )}
          </div>
        </div>
      );
    };
  }, [cols, rows, selectedSquare, isValidMove, previousMove, color]);

  return (
    <div className="relative">
      {/* Game board */}
      <div className="relative">
        {board.map((row, i) => (
          <div key={i} className="flex select-none">
            {row.map((piece, j) => squareComponent(i, j, piece))}
          </div>
        ))}
      </div>

      {/* File letters (a-h) */}
      <div className="relative flex text-right -top-6 right-1">
        {cols.split("").map((col, i) => (
          <div className="w-20" key={col}>
            <div className={`text-lg ${i % 2 === 0 ? "text-board-white" : "text-board-black"}`}>
              {col}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ChessBoard;
