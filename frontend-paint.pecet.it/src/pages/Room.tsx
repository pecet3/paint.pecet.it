import React, { useEffect, useRef, useState } from "react";
import { PaintCanvas } from "../components/paint/PaintCanvas";
import { decodeBase64ToPixels, encodePixelsToBase64 } from "../components/paint/pixel";
import type { ChatMessage, Event, RoomConfig, RoomInfo, RoomUser, ServerMessage, SignalPayload } from "../gengotypes";
import { Chat } from "../components/room/Chat";
import { WebRTCManager, type WebRTCManagerHandle } from "../components/room/WebRTCManager";
import { useNavigate, useParams } from "react-router";
import { useStore } from "../Store";
import type { Pixel } from "../types";
import { Synthesizer } from "../components/synth/Synthesizer";


export const PaintRoom: React.FC<{ roomInfo: RoomInfo }> = ({ roomInfo }) => {
  const { user } = useStore()
  const navigate = useNavigate();

  const ws = useRef<WebSocket | null>(null);
  const webrtcRef = useRef<WebRTCManagerHandle>(null);

  const [incomingPixels, setIncomingPixels] = useState<Pixel[] | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [users, setUsers] = useState<RoomUser[]>([]);
  const [serverMessage, setServerMessage] = useState<ServerMessage | null>({ date: "", message: "" });

  const [isJoined, setIsJoined] = useState(false);
  const [isWebRTC, setIsWebRTC] = useState(false)
  const [reconnectCounter, setReconnectCounter] = useState(0)
  const [resetKey, setResetKey] = useState(0);

  const [isConnect, setIsConnect] = useState(true)
  const [localUser, setLocalUser] = useState<RoomUser>({
    uuid: "",
    name: "",
    is_operator: false,
    is_connected: false,
    is_draw: false,
    is_kicked: false,
  })

  useEffect(() => {
    ws.current = new WebSocket("/api/join-room/" + roomInfo.name);

    ws.current.onmessage = (message: MessageEvent) => {
      const data = JSON.parse(message.data);
      console.log(data.type)
      switch (data.type) {
        case "canvas_pixel_update":
          setIncomingPixels(decodeBase64ToPixels(data.payload))
          break;
        case "canvas_reset":
          setResetKey(p => p + 1)
          setIncomingPixels(null)
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
        case "webrtc_signal":
          if (webrtcRef.current) {
            webrtcRef.current.receiveSignal(data.payload as SignalPayload);
          }
          break;
        case "join_confirmation":
          setIsJoined(true);
          break;

        case "kick_confirmation":
          alert("you are kicked from this room")
          navigate("/");
          break;
        default:
          console.warn(data.type);
      }
    };

    ws.current.onclose = () => {
      if (reconnectCounter < 5) {
        console.log("ws conn closed, reconnecting: ", reconnectCounter)
        ws.current = new WebSocket("/api/join-room/" + roomInfo.name);
        setReconnectCounter(reconnectCounter + 1)
      } else {

        navigate("/");
      }

    }
    return () => {
      ws.current?.close();
    };
  }, [roomInfo, navigate]);

  useEffect(() => {
    if (isJoined) {
      handleGetAllCanvas();
      let timer1 = setTimeout(() => setIsWebRTC(true), 1000);
      return () => {
        clearTimeout(timer1)
      }
    }
  }, [isJoined]);

  useEffect(() => {
    const usr = users.find(u => u.uuid == user?.uuid)
    if (usr !== undefined) {
      setLocalUser(usr)
    }
  }, [users]);



  const handleSendChatMessage = (msg: string) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "chat_message", payload: { message: msg } }));
    }
  };

  const handleUserKick = (uuid: string) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "user_kick", payload: { uuid } }));
    }
  };
  const handleUserOperator = (uuid: string) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "user_operator", payload: { uuid } }));
    }
  };
  const handleUserCanDrawing = (uuid: string) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "user_draw", payload: { uuid } }));
    }
  };

  const handleGetAllCanvas = () => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "canvas_get_all", payload: "" }));
    }
  };
  const handleCanvasReset = () => {
    confirm("Are you sure?")
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "canvas_reset", payload: "" }));
    }
  };
  const handleSendSignal = (payload: SignalPayload) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify({ type: "webrtc_signal", payload }));
    }
  };
  const handleSendPixelUpdateWS = (pixels: Pixel[]) => {
    if (!localUser.is_draw) return;
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

  // const [incomingNote, setIncomingNote] = useState<NoteEvent | null>(null)
  const receiveWebRTCData = (event: Event) => {
    switch (event.type) {
      case "canvas_pixel_update":
        setIncomingPixels(decodeBase64ToPixels(event.payload));
        break;
      case "synth_note_update":
        // setIncomingNote(event.payload)
        break;
      default:
        console.warn("unhandled webRTC event type: ", event.type);
    }
  };

  const handleSendPixelUpdateRTC = (pixels: Pixel[]) => {
    if (!localUser.is_draw) return;
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
  // const handleSendNoteRTC = (note: NoteEvent) => {
  //   if (isWebRTC) {
  //     if (webrtcRef.current) {
  //       webrtcRef.current.broadcastData({
  //         type: "synth_note_update",
  //         payload: note,
  //       });
  //     }
  //   };
  // }


  return (
    <>
      {isConnect ? <div className="m-2 flex flex-col items-center gap-2">
        < div className="flex gap-1 items-center justify-center w-full flex-col xl:flex-row" >
          {
            isWebRTC
              ? <div key={resetKey}>
                <PaintCanvas
                  config={roomInfo.config}
                  onSendPixelUpdate={handleSendPixelUpdateWS}
                  incomingPixels={incomingPixels}
                  onSendPixelUpdateRTC={handleSendPixelUpdateRTC}
                />
                {localUser.is_operator && <button onClick={() => handleCanvasReset()}>Reset Canvas</button>}
              </div>
              : <PaintCanvas
                key={resetKey}
                config={roomInfo.config}
                onSendPixelUpdate={handleSendPixelUpdateWS}
                incomingPixels={incomingPixels}
              />}

          <div className="flex flex-col items-center m-auto h-full w-full justify-between">
            {isWebRTC && isConnect ?
              (
                <>
                  <WebRTCManager
                    ref={webrtcRef}
                    users={users}
                    onSendSignal={handleSendSignal}
                    onDataReceived={receiveWebRTCData}
                  />

                </>
              )
              : null}

            <div className="flex text-xl justify-center">
              {serverMessage!.message && serverMessage!.message}
            </div>

            <Chat
              users={users}
              messages={chatMessages}
              onSendMessage={handleSendChatMessage}
              operatorHandlers={{
                onDrawing: handleUserCanDrawing,
                onKick: handleUserKick,
                onOp: handleUserOperator
              }}
            />
            {roomInfo.config.is_synth && <div>.</div>}
          </div>
        </div >

      </div > : <button className="btn" onClick={async () => {
        setIsConnect(true)

      }}>Connect</button>}
    </>
  );
};



export const Room = () => {
  let navigate = useNavigate();
  let { roomName } = useParams();
  const [room, setRoom] = useState<RoomInfo | null>(null)

  const fetchRoom = async () => {
    try {
      const response = await fetch('/api/rooms/' + roomName);
      if (response.ok) {
        const data = await response.json();
        console.log(data)
        setRoom(data);
      } else {
        navigate("/")
      }
    } catch (error) {
      navigate("/")
    }
  };

  useEffect(() => {
    roomName !== undefined && fetchRoom();
  }, [roomName]);


  return (
    <>
      {room && <>
        <PaintRoom roomInfo={room} />
      </>}
    </>
  );
};