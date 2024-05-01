import React from "react";

const Button: React.FC<{
  onClick: () => void | undefined;
  children: React.ReactNode;
  disabled?: boolean;
}> = ({ onClick = () => {}, children, disabled = false }) => {
  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className="px-8 py-4 text-lg font-bold text-white bg-green-600 rounded hover:bg-green-800"
    >
      {children}
    </button>
  );
};

export default Button;
