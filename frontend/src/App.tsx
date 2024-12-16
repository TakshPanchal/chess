import { BrowserRouter, Route, Routes, Navigate } from "react-router-dom";
import PlayPage from "./screens/Play";

function App() {
  return (
    <div className="min-h-screen bg-slate-800">
      <BrowserRouter>
        <Routes>
          {/* Redirect root to play */}
          <Route path="/" element={<Navigate to="/play" replace />} />
          
          {/* Play routes */}
          <Route path="/play" element={<PlayPage />} />
          <Route path="/play/:gameId" element={<PlayPage />} />
        </Routes>
      </BrowserRouter>
    </div>
  );
}

export default App;
