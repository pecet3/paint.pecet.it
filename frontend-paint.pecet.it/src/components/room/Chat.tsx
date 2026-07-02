import React, { useEffect, useRef, useState } from "react";
import type { ChatMessage, RoomUser } from "../../types";

interface ChatProps {
    messages: ChatMessage[];
    users: RoomUser[];
    onSendMessage: (message: string) => void;
}

export const Chat: React.FC<ChatProps> = ({ messages, users, onSendMessage }) => {
    const [input, setInput] = useState("");
    const messagesEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        // messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (input.trim()) {
            onSendMessage(input.trim());
            setInput("");
        }
    };

    return (
        <div className="flex h-[600px] w-full max-w-4xl bg-white border border-gray-200 rounded-xl shadow-lg overflow-hidden">
            <div className="w-1/4 border-r border-gray-200 bg-gray-50 p-4 flex flex-col">
                <h2 className="font-bold text-gray-700 mb-4 text-sm uppercase tracking-wider">Users</h2>
                <div className="flex flex-col gap-2 overflow-y-auto">
                    {users.map((user) => (
                        <div key={user.uuid} className="flex items-center gap-2 text-sm">
                            <span className={`w-2.5 h-2.5 rounded-full ${user.is_connected ? "bg-green-500" : "bg-gray-400"}`} />
                            <span className="truncate text-gray-700">{user.name}</span>
                            {user.is_operator && <span className="text-[10px] bg-blue-100 text-blue-600 px-1 rounded">OP</span>}
                        </div>
                    ))}
                </div>
            </div>

            <div className="flex-1 flex flex-col">
                <div className="flex-1 overflow-y-auto p-4 space-y-4 bg-gray-50/50">
                    {messages.map((msg, idx) => (
                        <div key={idx} className="flex flex-col">
                            <div className="flex items-baseline gap-2">
                                <span className="font-semibold text-sm text-blue-600">{msg.name}</span>
                                <span className="text-[10px] text-gray-400">{msg.date}</span>
                            </div>
                            <p className="bg-white px-3 py-2 rounded-lg rounded-tl-none shadow-sm text-gray-800 text-sm border border-gray-100">
                                {msg.message}
                            </p>
                        </div>
                    ))}
                    <div ref={messagesEndRef} />
                </div>

                <form onSubmit={handleSubmit} className="p-3 border-t border-gray-200 bg-white flex gap-2">
                    <input
                        type="text"
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        placeholder="Type a message..."
                        className="flex-1 px-4 py-2 border border-gray-300 rounded-full focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                    />
                    <button
                        type="submit"
                        className="px-5 py-2 bg-blue-600 text-white rounded-full font-medium hover:bg-blue-700 transition-colors text-sm"
                    >
                        Send
                    </button>
                </form>
            </div>
        </div>
    );
};