import { useEffect, useState } from "react";

const WS_URL = import.meta.env.VITE_WS_HOST
  ? `wss://${import.meta.env.VITE_WS_HOST}/ws`
  : "ws://localhost:8080/ws";

export const useSocket = () => {
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [connection, setConnection] = useState<"connecting" | "connected" | "disconnected">("connecting");
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    const ws = new WebSocket(WS_URL);
    setConnection("connecting");
    ws.onopen = () => {
      setSocket(ws);
      setConnection("connected");
    };
    ws.onclose = () => {
      setSocket(null);
      setConnection("disconnected");
    };
    return () => {
      ws.onclose = null;
      ws.close();
    };
  }, [attempt]);

  return { socket, connection, reconnect: () => setAttempt((value) => value + 1) };
};
