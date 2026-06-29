import React, { useEffect, useRef, useState } from "react";
import type { ChatMessage } from "../../types";

interface ChatProps {
    messages: ChatMessage[];
    onSendMessage: (message: string) => void;
}

export const Chat: React.FC<ChatProps> = ({ messages, onSendMessage }) => {
    const [input, setInput] = useState("");
    const messagesEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        if (input.trim()) {
            onSendMessage(input.trim());
            setInput("");
        }
    };

    return (
        <div style={{ display: "flex", flexDirection: "column", height: "100%", border: "1px solid #ccc", background: "#fff" }}>
            <div style={{ flex: 1, overflowY: "auto", padding: "10px", display: "flex", flexDirection: "column", gap: "8px" }}>
                {messages.map((msg, idx) => (

                    <div>
                        <strong style={{ color: "#0056b3" }}>{msg.name}: </strong>
                        <strong style={{ color: "#0056b3" }}>{msg.date}: </strong>
                        <span>{msg.message}</span>
                    </div>

                ))}
                <div ref={messagesEndRef} />
            </div>
            <form onSubmit={handleSubmit} style={{ display: "flex", borderTop: "1px solid #ccc", padding: "10px" }}>
                <input
                    type="text"
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    placeholder="Wpisz wiadomość..."
                    style={{ flex: 1, padding: "8px", marginRight: "8px" }}
                />
                <button type="submit" style={{ padding: "8px 16px" }}>
                    Wyślij
                </button>
            </form>
        </div>
    );
};