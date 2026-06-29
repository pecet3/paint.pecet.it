import React, { useEffect, useRef, useState } from "react";
import { PaintCanvas } from "../components/room/PaintCanvas";
import { decodeBase64ToPixels, encodePixelsToBase64 } from "../components/room/pixel";
import type { ChatMessage, Pixel, RoomUser, ServerMessage } from "../types";
import { Chat } from "../components/room/Chat";


export const Home: React.FC = () => {
  const ws = useRef<WebSocket | null>(null);
  const [incomingPixels, setIncomingPixels] = useState<Pixel[] | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [users, setUsers] = useState<RoomUser[]>([]);

  const [serverMessage, setServerMessage] = useState<ServerMessage | null>(null)
  useEffect(() => {
    ws.current = new WebSocket("/ws");

    ws.current.onmessage = (message: MessageEvent) => {
      console.log(message);
      const data = JSON.parse(message.data);
      switch (data.type) {
        case "canvas_pixel_update":
          setIncomingPixels(decodeBase64ToPixels(data.payload));
          break;
        case "chat_message":
          setChatMessages(prev => [...prev, data.payload as ChatMessage]);
          break;
        case "server_message":
          setServerMessage(data.payload as ServerMessage)
          break;
        case "update_users_list":
          setUsers(data.payload);
          break;
        case "update_is_operator":
        case "update_ban_duration":
          break;
        default:
          console.warn(data.type);
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
  const handleSendChatMessage = (msg: string) => {
    console.log(msg)
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(
        JSON.stringify({
          type: "chat_message",
          payload: { message: msg },
        })
      );
    }
  };
  return (
    <div >
      <div className="flex">
        <PaintCanvas
          onSendPixelUpdate={handleSendPixelUpdate}
          incomingPixels={incomingPixels}
        />
        <Chat messages={chatMessages} onSendMessage={handleSendChatMessage} />

      </div>
      <div className="flex text-2xl">
        {serverMessage?.message}
      </div>
    </div>
  );
};