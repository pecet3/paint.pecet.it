import React, { useEffect, useRef, useState } from "react";
import { PaintCanvas } from "../components/room/PaintCanvas";
import { decodeBase64ToPixels, encodePixelsToBase64 } from "../components/room/pixel";
import type { ChatMessage, Event, Pixel, RoomUser, ServerMessage, WebRTCSignalPayload } from "../types";
import { Chat } from "../components/room/Chat";
import { WebRTCManager, type WebRTCManagerHandle } from "../components/room/WebRTCManager";
import { useNavigate, useParams } from "react-router";

export const Room: React.FC = () => {
  let { roomName } = useParams();
  const navigate = useNavigate();

  const ws = useRef<WebSocket | null>(null);
  const webrtcRef = useRef<WebRTCManagerHandle>(null);

  const [incomingPixels, setIncomingPixels] = useState<Pixel[] | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [users, setUsers] = useState<RoomUser[]>([]);
  const [serverMessage, setServerMessage] = useState<ServerMessage | null>(null);

  const [isJoined, setIsJoined] = useState(false);
  const [isWebRTC, setIsWebRTC] = useState(false)

  useEffect(() => {
    ws.current = new WebSocket("/api/rooms/" + roomName);

    ws.current.onmessage = (message: MessageEvent) => {
      const data = JSON.parse(message.data);

      switch (data.type) {
        case "canvas_pixel_update":
          setIncomingPixels(decodeBase64ToPixels(data.payload))
          break;
        case "chat_message":
          setChatMessages(prev => [...prev, data.payload as ChatMessage]);
          break;
        case "server_message":
          setServerMessage(data.payload as ServerMessage);
          break;
        case "update_users_list":
          setUsers(data.payload);
          break;
        case "update_is_operator":
        case "update_ban_duration":
          break;
        case "webrtc_signal":
          if (webrtcRef.current) {
            webrtcRef.current.receiveSignal(data.payload as WebRTCSignalPayload);
          }
          break;
        case "join_confirmation":
          setIsJoined(true);
          break;
        default:
          console.warn(data.type);
      }
    };

    ws.current.onclose = (evt: CloseEvent) => {
      console.log(evt);
      navigate("/");
    };

    return () => {
      ws.current?.close();
    };
  }, [roomName, navigate]);

  useEffect(() => {
    if (isJoined) handleGetAllCanvas();
  }, [isJoined]);


  const handleSendChatMessage = (msg: string) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "chat_message", payload: { message: msg } }));
    }
  };

  const handleKick = (uuid: string) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "kick_user", payload: { uuid } }));
    }
  };

  const handleGetAllCanvas = () => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "canvas_get_all", payload: "" }));
    }
  };

  const handleSendSignal = (payload: WebRTCSignalPayload) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "webrtc_signal", payload }));
    }
  };
  const handleSendPixelUpdateWS = (pixels: Pixel[]) => {
    const base64Payload = encodePixelsToBase64(pixels);
    if (ws.current?.readyState === WebSocket.OPEN)
      ws.current.send(JSON.stringify({
        type: "canvas_pixel_update",
        payload: base64Payload
      }));
  }
  /*

    WebRTC

  */
  const receiveWebRTCData = (event: Event) => {
    switch (event.type) {
      case "canvas_pixel_update":
        setIncomingPixels(decodeBase64ToPixels(event.payload));
        break;
      default:
        console.warn("unhandled webRTC event type: ", event.type);
    }
  };

  const handleSendPixelUpdateRTC = (pixels: Pixel[]) => {
    const base64Payload = encodePixelsToBase64(pixels);
    if (isWebRTC) {
      if (webrtcRef.current) {
        webrtcRef.current.broadcastData({
          type: "canvas_pixel_update",
          payload: base64Payload
        });
      }
    };

  }


  return (
    <div className="m-2 flex flex-col items-center gap-2">
      <div className="flex gap-1 items-center justify-center w-full flex-col xl:flex-row">

        {isWebRTC
          ? <PaintCanvas
            onSendPixelUpdate={handleSendPixelUpdateWS}
            incomingPixels={incomingPixels}
            onSendPixelUpdateRTC={handleSendPixelUpdateRTC}
          />
          : <PaintCanvas
            onSendPixelUpdate={handleSendPixelUpdateWS}
            incomingPixels={incomingPixels}
          />}

        <div className="flex flex-col items-center m-auto w-full justify-between">
          {isJoined && isWebRTC ?
            (
              <>
                <WebRTCManager
                  ref={webrtcRef}
                  users={users}
                  onSendSignal={handleSendSignal}
                  onDataReceived={receiveWebRTCData}
                />
                <button className=" btn bg-red-800 text-xs" onClick={() => {
                  setIsWebRTC(false)
                }}>Disconnect</button>
              </>
            )
            :
            <button className=" btn bg-lime-700 text-xs" onClick={() => {
              setIsWebRTC(true)
            }}>Connect</button>}
          {serverMessage && (
            <div className="flex text-xl justify-center">
              {serverMessage.message}
            </div>
          )}
          <Chat
            users={users}
            messages={chatMessages}
            onSendMessage={handleSendChatMessage}
            onKick={handleKick}
          />
        </div>
      </div>
    </div>
  );
};