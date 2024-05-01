import { Link } from "react-router-dom";
import Button from "../components/Button";

const LandingPage = () => {
  return (
    <div>
      <div className="grid max-w-screen-lg grid-cols-1 gap-4 pt-8 mx-auto md:grid-cols-2">
        <div className="flex justify-center">
          <img src="/chess.png" alt="chess" className="max-w-100" />
        </div>
        <div className="pt-16 text-white">
          <h1 className="font-mono text-6xl text-center">
            Play Chess Online on the #0 chess site!
          </h1>
          <div className="flex justify-center mt-10">
            <Link to="/play">
              <Button onClick={() => {}}>Play Online</Button>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default LandingPage;
