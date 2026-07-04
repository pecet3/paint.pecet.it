import React, { useEffect, useRef, useState } from "react";
import { PaintCanvas } from "../components/room/PaintCanvas";
import { decodeBase64ToPixels, encodePixelsToBase64 } from "../components/room/pixel";
import type { ChatMessage, Pixel, RoomUser, ServerMessage, WebRTCSignalPayload } from "../types";
import { Chat } from "../components/room/Chat";
import { WebRTCManager } from "../components/room/WebRTCManager";
import { useParams } from "react-router";


export const Room: React.FC = () => {
  let { roomName } = useParams();


  const ws = useRef<WebSocket | null>(null);
  const [incomingPixels, setIncomingPixels] = useState<Pixel[] | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [users, setUsers] = useState<RoomUser[]>([]);
  const [incomingSignal, setIncomingSignal] = useState<WebRTCSignalPayload | null>(null);



  const [serverMessage, setServerMessage] = useState<ServerMessage | null>(null)
  useEffect(() => {
    console.log(roomName)
    if (roomName === undefined || roomName === "") {
      return
    }
    ws.current = new WebSocket("/api/rooms/" + roomName);

    ws.current.onmessage = (message: MessageEvent) => {
      const data = JSON.parse(message.data);
      switch (data.type) {
        case "canvas_pixel_update":
          setIncomingPixels(decodeBase64ToPixels(data.payload));
          break;
        case "chat_message":
          setChatMessages(prev => [...prev, data.payload as ChatMessage]);
          break;
        case "server_message":
          console.log(data.payload)
          setServerMessage(data.payload as ServerMessage)
          break;
        case "update_users_list":
          setUsers(data.payload);
          break;
        case "update_is_operator":
        case "update_ban_duration":
          break;
        case "webrtc_signal":
          console.log(data.payload)
          setIncomingSignal(data.payload as WebRTCSignalPayload);
          break;
        default:
          console.warn(data.type);
      }
    };

    return () => {
      ws.current?.close();
    };
  }, [roomName]);

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
  const handleSendSignal = (payload: WebRTCSignalPayload) => {
    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(
        JSON.stringify({
          type: "webrtc_signal",
          payload,
        })
      );
    }
  };
  return (
    <div >
      <div className="flex m-auto items-center justify-center w-full">
        <PaintCanvas
          onSendPixelUpdate={handleSendPixelUpdate}
          incomingPixels={incomingPixels}
        />

      </div>
      <div className="flex text-3xl bg-gray-200/50 justify-center ">
        {serverMessage?.message}
      </div>
      <div className="w-full flex">
        <Chat users={users} messages={chatMessages} onSendMessage={handleSendChatMessage} />

        <WebRTCManager
          users={users}
          incomingSignal={incomingSignal}
          onSendSignal={handleSendSignal}
        />
      </div>
    </div>
  );
};