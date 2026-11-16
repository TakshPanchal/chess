import { ReactNode } from "react";
import { Link } from "react-router-dom";

const AppHeader = ({ children }: { children?: ReactNode }) => (
  <header className="app-header">
    <Link className="wordmark" to="/">
      Still Chess
    </Link>
    {children}
  </header>
);

export default AppHeader;
