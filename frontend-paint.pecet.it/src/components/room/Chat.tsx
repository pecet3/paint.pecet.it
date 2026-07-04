import React, { useEffect, useRef, useState } from "react";
import type { ChatMessage, RoomUser } from "../../types";

interface ChatProps {
    messages: ChatMessage[];
    users: RoomUser[];
    onSendMessage: (message: string) => void;
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

export const Chat: React.FC<ChatProps> = ({ messages, users, onSendMessage }) => {
    const [input, setInput] = useState("");
    const chatContainerRef = useRef<HTMLDivElement>(null);

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

    return (
        <div className="bg-slate-700 rounded-lg m-auto border border-black flex max-w-xl w-full h-96">
            <div className="w-1/4 border-r border-gray-400 flex flex-col items-center bg-slate-800 rounded-l-lg">
                <h2 className="font-bold">Users</h2>
                <div className="flex flex-col gap-2 overflow-y-auto">
                    {users.map((user) => (
                        <div key={user.uuid} className="flex items-center gap-0.5 text-xs">
                            <span className={`w-2.5 h-2.5 rounded-full ${user.is_connected ? "bg-green-500" : "bg-gray-400"}`} />
                            <span className="truncate">{user.name}</span>
                            {user.is_operator && <span className="text-[8px] bg-blue-100 text-blue-600 px-1 rounded">OP</span>}
                        </div>
                    ))}
                </div>
            </div>

            <div className="flex-1 flex flex-col pt-1 pr-1">
                <div
                    ref={chatContainerRef}
                    className="flex-1 overflow-y-auto p-2 scroll-smooth"
                >
                    {messages.map((msg, idx) => (
                        <div key={idx} className="flex flex-col mb-2">
                            <div className="flex items-baseline gap-2">
                                <span className="font-semibold text-sm text-gray-300">{msg.name}</span>
                                <span className="text-[10px] text-gray-400">{formatMessageDate(msg.date)}</span>
                            </div>
                            <p className="bg-slate-300 text-left p-1 rounded-lg rounded-tl-none shadow-sm text-gray-800 text-sm border border-gray-100 break-all">
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
                        className="inpt w-full"
                    />
                    <button type="submit" className="btn bg-black">
                        Send
                    </button>
                </form>
            </div>
        </div>
    );
};