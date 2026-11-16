import { Link } from "react-router-dom";
import AppHeader from "../components/AppHeader";

const LandingPage = () => (
  <div className="page">
    <AppHeader />
    <main className="welcome">
      <section className="welcome-intro" aria-labelledby="welcome-title">
        <h1 id="welcome-title" className="welcome-title">Play chess online</h1>
        <p className="welcome-lead">Get matched with the next available player. No account required.</p>
        <Link className="button button--primary button--large" to="/play" state={{ findOpponent: true }}>
          Find an opponent
        </Link>
        <p className="welcome-meta">Untimed games · Colors assigned at random</p>
      </section>
      <section className="welcome-guide" aria-labelledby="guide-title">
        <h2 id="guide-title" className="guide-title">How to play</h2>
        <ol className="guide-steps">
          <li>
            <span className="guide-number" aria-hidden="true">1</span>
            <span>
              <strong>Find an opponent.</strong>
              You'll be paired with the next player who is searching.
            </span>
          </li>
          <li>
            <span className="guide-number" aria-hidden="true">2</span>
            <span>
              <strong>Select one of your pieces.</strong>
              Its legal moves are marked on the board.
            </span>
          </li>
          <li>
            <span className="guide-number" aria-hidden="true">3</span>
            <span>
              <strong>Select a marked square to move.</strong>
              Pawns that reach the last rank become queens.
            </span>
          </li>
        </ol>
        <p className="guide-note">A game ends if either player leaves or loses their connection.</p>
      </section>
    </main>
  </div>
);

export default LandingPage;
