import { createBrowserRouter, RouterProvider } from "react-router-dom";
import LandingPage from "./screens/Landing";
import PlayPage from "./screens/Play";

const router = createBrowserRouter([
  { path: "/", element: <LandingPage /> },
  { path: "/play", element: <PlayPage /> },
]);

function App() {
  return (
    <div className="app-shell">
      <RouterProvider router={router} />
    </div>
  );
}

export default App;
