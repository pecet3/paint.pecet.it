import { useEffect, useRef, useState } from "react";
import { PaintCanvas } from "../components/paint/PaintCanvas";
import type { DrawEvent } from "../dto";


export const Home: React.FC = () => {
  const ws = useRef<WebSocket | null>(null);
  const [incomingEvent, setIncomingEvent] = useState<DrawEvent | null>(null);

  const room = "1"
  useEffect(() => {
    ws.current = new WebSocket(`ws://localhost:8080/ws?room=${room}`);
    ws.current.onmessage = (message: any) => {
      console.log(message)
      const data: DrawEvent = JSON.parse(message.data);
      setIncomingEvent(data);
    };

    return () => {
      ws.current?.close();
    };
  }, []);

  const handleSendDrawEvent = (drawEvent: DrawEvent) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      const wsEvent = {
        type: "drawing",
        payload: drawEvent,
      }
      ws.current.send(JSON.stringify(wsEvent));
    }
  };



  return (
    <PaintCanvas onSendDrawEvent={handleSendDrawEvent}
      incomingDrawEvent={incomingEvent} />
  );
};

