import { BrowserRouter, Route, Routes } from "react-router-dom";
import LandingPage from "./screens/Landing";
import PlayPage from "./screens/Play";

function App() {
  return (
    <div className="h-screen bg-slate-800	">
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<LandingPage />} />
          <Route path="/play" element={<PlayPage />} />
        </Routes>
      </BrowserRouter>
    </div>
  );
}

export default App;
