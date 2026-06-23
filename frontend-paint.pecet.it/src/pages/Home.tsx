import React, { useEffect, useRef, useState } from "react";
import { PaintCanvas, type Pixel } from "../components/paint/PaintCanvas";

export const Home: React.FC = () => {
  const ws = useRef<WebSocket | null>(null);
  const [incomingPixels, setIncomingPixels] = useState<Pixel[] | null>(null);
  const room = "1";

  useEffect(() => {
    ws.current = new WebSocket(`ws://localhost:8080/ws?room=${room}`);

    ws.current.onmessage = (message: MessageEvent) => {
      console.log(message)
      const data = JSON.parse(message.data);
      if (data.type === "canvas_pixel_update") {
        setIncomingPixels(data.payload);
      }
    };

    return () => {
      ws.current?.close();
    };
  }, []);

  const handleSendPixelUpdate = (pixels: Pixel[]) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({
        type: "canvas_pixel_update",
        payload: pixels,
      }));
    }
  };

  return (
    <PaintCanvas
      onSendPixelUpdate={handleSendPixelUpdate}
      incomingPixels={incomingPixels}
    />
  );
};