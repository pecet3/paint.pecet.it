import React, { useEffect, useRef, useState } from "react";
import type { ChatMessage, RoomUser } from "../../types";
import { useStore } from "../../Store";


interface OperatorHandlers {
    onKick: (uuid: string) => void;
    onOp: (uuid: string) => void;
    onDrawing: (uuid: string) => void;
}
interface ChatProps {
    messages: ChatMessage[];
    users: RoomUser[];
    onSendMessage: (message: string) => void;
    operatorHandlers: OperatorHandlers;
}

function formatMessageDate(dateString: string): string {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return dateString;

    return date.toLocaleString('pl-PL', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

export const Chat: React.FC<ChatProps> = ({ messages, users, onSendMessage, operatorHandlers }) => {
    const { user } = useStore()
    const [isOp, setIsOp] = useState(false)
    const [input, setInput] = useState("");
    const chatContainerRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const localUser = users.find(u => u.uuid == user?.uuid)
        localUser?.is_operator && setIsOp(true)
    }, [users]);

    useEffect(() => {
        const container = chatContainerRef.current;
        if (container) {
            container.scrollTo({
                top: container.scrollHeight,
                behavior: "smooth"
            });
        }
    }, [messages]);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (input.trim()) {
            onSendMessage(input.trim());
            setInput("");
        }
    };
    const [uuidToManage, setUuidToManage] = useState("")
    return (
        <div className="bg-slate-700 rounded-lg m-auto border border-black flex max-w-2xl w-full h-64">
            <div className="w-1/3 border-r border-gray-400 flex flex-col items-center bg-slate-800 rounded-l-lg">
                <h2 className="font-bold text-sm border-b w-full border-b-gray-400">Users</h2>
                <div className="flex flex-col gap-2 overflow-y-auto m-0 w-full">
                    {users.map((user) => (
                        <div key={user.uuid} className={`flex flex-col items-start pl-0.5 gap-0.5 text-xs ${uuidToManage === user.uuid && "bg-slate-900"}`}>
                            <div className="flex items-center gap-0.5 text-xs tracking-tighter">
                                <span className={`w-1.5 h-1.5 rounded-full ${user.is_connected ? "bg-green-500" : "bg-gray-400"}`} />
                                {user.is_drawing && "🎨"}
                                <span className="truncate">{user.name.slice(0, 16)}</span>
                                {user.is_operator && <span className="text-[8px] tracking-wide bg-blue-100 text-blue-600 px-1 rounded">OP</span>}
                                {isOp && <button onClick={() => {
                                    uuidToManage === user.uuid ? setUuidToManage("") : setUuidToManage(user.uuid)
                                }} className="cursor-pointer ">⚙️</button>}
                            </div>
                            {uuidToManage === user.uuid &&
                                <div className="flex w-full font-mono tracking-tighter items-center gap-0.5 text-xs bg-slate-900 m-0 justify-evenly">
                                    <button onClick={() => operatorHandlers.onKick(user.uuid)} className="cursor-pointer">Kick</button>
                                    <button onClick={() => operatorHandlers.onOp(user.uuid)} className="cursor-pointer">Op</button>
                                    <button onClick={() => operatorHandlers.onDrawing(user.uuid)} className="cursor-pointer">Drawing</button>
                                </div>}
                        </div>
                    ))}
                </div>
            </div>

            <div className="flex-1 flex flex-col">
                <div
                    ref={chatContainerRef}
                    className="flex-1 overflow-y-auto p-1 scroll-smooth"
                >
                    {messages.map((msg, idx) => (
                        <div key={idx} className="flex flex-col text-sm">
                            <div className="flex items-baseline gap-2">
                                <span className="font-semibold text-sm text-gray-300">{msg.name}</span>
                                <span className="text-[10px] text-gray-400">{formatMessageDate(msg.date)}</span>
                            </div>
                            <p className="bg-slate-300 text-left p-1 
                            rounded-lg rounded-tl-none shadow-sm text-gray-800 text-xs border border-gray-100 break-all">
                                {msg.message}
                            </p>
                        </div>
                    ))}
                </div>

                <form onSubmit={handleSubmit} className="p-1 border-t border-gray-400 flex gap-2">
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        placeholder=""
                        className="inpt w-full text-xs"
                    />
                    <button type="submit" className="btn bg-black">
                        Send
                    </button>
                </form>
            </div>
        </div>
    );
};