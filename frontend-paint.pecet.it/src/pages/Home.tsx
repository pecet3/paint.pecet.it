import React, { useEffect, useRef, useState } from "react";
import { PaintCanvas } from "../components/paint/PaintCanvas";
import { decodeBase64ToPixels, encodePixelsToBase64 } from "../components/paint/pixel";
import type { Pixel } from "../types";
import { wsAddr } from "../config";


export const Home: React.FC = () => {
  const ws = useRef<WebSocket | null>(null);
  const [incomingPixels, setIncomingPixels] = useState<Pixel[] | null>(null);
  const room = "1";

  useEffect(() => {
    ws.current = new WebSocket(wsAddr);

    ws.current.onmessage = (message: MessageEvent) => {
      console.log(message);
      const data = JSON.parse(message.data);
      if (data.type === "canvas_pixel_update") {
        const decodedPixels = decodeBase64ToPixels(data.payload);
        setIncomingPixels(decodedPixels);
      }
    };

    return () => {
      ws.current?.close();
    };
  }, []);

  const handleSendPixelUpdate = (pixels: Pixel[]) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      const base64Payload = encodePixelsToBase64(pixels);

      ws.current.send(JSON.stringify({
        type: "canvas_pixel_update",
        payload: base64Payload,
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