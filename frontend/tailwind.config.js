/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        transparent: "transparent",
        current: "currentColor",
        "board-white": "#E5E5E5",
        "board-black": "#B7C0D8",
      },
    },
  },
  plugins: [],
};
