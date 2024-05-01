import { useEffect, useState } from "react";

const WS_URL = "ws://localhost:8080/ws";

export const useSocket = () => {
  const [socket, setSocket] = useState<WebSocket | null>(null);

  useEffect(() => {
    const ws = new WebSocket(WS_URL);

    ws.onopen = () => {
      console.log("Connected to server");
      setSocket(ws);
    };

    ws.onclose = () => {
      console.log("Disconnected from server");
      setSocket(null);
    };

    return () => {
      ws.close();
    };
  }, []);

  return socket;
};
