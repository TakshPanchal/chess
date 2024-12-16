import { useEffect, useState, useCallback, useRef } from "react";

interface WebSocketOptions {
  gameId?: string;
  isSpectator?: boolean;
  isPlayer?: boolean;
}

export const useSocket = (options: WebSocketOptions = {}) => {
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [isConnecting, setIsConnecting] = useState(false);
  const socketRef = useRef<WebSocket | null>(null);

  const connect = useCallback(() => {
    if (isConnecting || socketRef.current?.readyState === WebSocket.OPEN) {
      return null;
    }

    setIsConnecting(true);
    
    // Construct WebSocket URL with parameters
    const wsUrl = new URL("ws://localhost:8080/ws");
    
    if (options.gameId) {
      wsUrl.searchParams.set('gameId', options.gameId);
    }
    
    // Only set spectator mode if not explicitly marked as a player
    if (options.isSpectator && !options.isPlayer) {
      wsUrl.searchParams.set('spectator', 'true');
    }
    if (options.isPlayer) {
      wsUrl.searchParams.set('play', 'true');
    }

    console.log("Connecting to WebSocket:", wsUrl.toString());
    const ws = new WebSocket(wsUrl.toString());
    socketRef.current = ws;

    ws.onopen = () => {
      console.log("WebSocket connected successfully");
      setSocket(ws);
      setIsConnecting(false);
    };

    ws.onclose = () => {
      console.log("WebSocket disconnected");
      setSocket(null);
      setIsConnecting(false);
      socketRef.current = null;
      // Try to reconnect after a delay
      setTimeout(() => connect(), 3000);
    };

    ws.onerror = (error) => {
      console.error("WebSocket error:", error);
      setIsConnecting(false);
    };

    return ws;
  }, [options.gameId, options.isSpectator, options.isPlayer]);

  useEffect(() => {
    // Only connect if we don't have an active connection
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      const ws = connect();
      return () => {
        setIsConnecting(false);
        if (ws && ws.readyState === WebSocket.OPEN) {
          ws.close();
        }
        socketRef.current = null;
      };
    }
  }, [connect]);

  // Add event listener for beforeunload to close socket properly
  useEffect(() => {
    const handleBeforeUnload = () => {
      if (socketRef.current?.readyState === WebSocket.OPEN) {
        socketRef.current.close();
      }
    };

    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, []);

  return socket;
};
