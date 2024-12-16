import React from "react";

const Button: React.FC<{
  onClick: () => void | Promise<void>;
  children: React.ReactNode;
  disabled?: boolean;
  className?: string;
}> = ({ onClick = () => {}, children, disabled = false, className = "" }) => {
  const handleClick = async () => {
    try {
      await onClick();
    } catch (error) {
      console.error('Button click error:', error);
    }
  };

  return (
    <button
      onClick={handleClick}
      disabled={disabled}
      className={`px-8 py-4 text-lg font-bold text-white rounded transition-colors ${className || "bg-green-600 hover:bg-green-800"}`}
    >
      {children}
    </button>
  );
};

export default Button;
