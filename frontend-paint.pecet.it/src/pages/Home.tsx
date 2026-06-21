import React, { useState, useEffect, useRef, type ChangeEvent, type FormEvent } from 'react';

// Interfejs odzwierciedlający strukturę Message z backendu w Go
interface ChatMessage {
  room: string;
  sender: string;
  content: string;
}

export const Home: React.FC = () => {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [inputMessage, setInputMessage] = useState<string>('');
  const [room, setRoom] = useState<string>('general');
  const [sender, setSender] = useState<string>(`User_${Math.floor(Math.random() * 1000)}`);

  // Otypowanie referencji na WebSocket (może być nullem przed połączeniem)
  const ws = useRef<WebSocket | null>(null);

  useEffect(() => {
    // Inicjalizacja połączenia WebSocket
    ws.current = new WebSocket(`ws://localhost:8080/ws?room=${room}`);

    ws.current.onopen = () => {
      console.log(`Połączono z pokojem: ${room}`);
      setMessages([]); // Czyszczenie czatu przy zmianie pokoju
    };

    ws.current.onmessage = (event: MessageEvent) => {
      try {
        const incomingMessage: ChatMessage = JSON.parse(event.data);
        setMessages((prev) => [...prev, incomingMessage]);
      } catch (error) {
        console.error("Błąd parsowania wiadomości:", error);
      }
    };

    ws.current.onclose = () => {
      console.log('Rozłączono z WebSocket');
    };

    // Sprzątanie połączenia przy unmouncie lub zmianie pokoju
    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, [room]);

  const sendMessage = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!inputMessage.trim()) return;

    const messagePayload: ChatMessage = {
      room: room,
      sender: sender,
      content: inputMessage,
    };

    if (ws.current && ws.current.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(messagePayload));
      setInputMessage('');
    }
  };

  const handleRoomChange = (e: ChangeEvent<HTMLSelectElement>) => {
    setRoom(e.target.value);
  };

  const handleSenderChange = (e: ChangeEvent<HTMLInputElement>) => {
    setSender(e.target.value);
  };

  const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
    setInputMessage(e.target.value);
  };

  return (
    <div style={{ maxWidth: '600px', margin: '20px auto', fontFamily: 'Arial, sans-serif' }}>
      <h2>TypeScript WebSocket Chat</h2>

      {/* Konfiguracja użytkownika i pokoju */}
      <div style={{ display: 'flex', gap: '15px', marginBottom: '20px' }}>
        <label style={{ flex: 1 }}>
          Twój Nick:
          <input
            type="text"
            value={sender}
            onChange={handleSenderChange}
            style={{ width: '100%', marginTop: '5px', padding: '8px', boxSizing: 'border-box' }}
          />
        </label>

        <label style={{ width: '150px' }}>
          Pokój:
          <select
            value={room}
            onChange={handleRoomChange}
            style={{ width: '100%', marginTop: '5px', padding: '8px', boxSizing: 'border-box' }}
          >
            <option value="general">Ogólny</option>
            <option value="gaming">Gracze</option>
            <option value="tech">Technologia</option>
          </select>
        </label>
      </div>

      {/* Okno z wiadomościami */}
      <div style={{ border: '1px solid #ccc', height: '350px', overflowY: 'scroll', padding: '15px', background: '#fdfdfd', borderRadius: '4px' }}>
        {messages.map((msg, index) => {
          const isMe = msg.sender === sender;
          return (
            <div key={index} style={{ textAlign: isMe ? 'right' : 'left', margin: '10px 0' }}>
              <span style={{ fontSize: '11px', color: '#777', display: 'block' }}>
                {msg.sender} <span style={{ color: '#aaa' }}>({msg.room})</span>
              </span>
              <div style={{
                display: 'inline-block',
                background: isMe ? '#007bff' : '#e4e6eb',
                color: isMe ? '#fff' : '#000',
                padding: '8px 14px',
                borderRadius: '18px',
                marginTop: '3px',
                maxWidth: '70%',
                wordBreak: 'break-word'
              }}>
                {msg.content}
              </div>
            </div>
          );
        })}
      </div>

      {/* Formularz wysyłania */}
      <form onSubmit={sendMessage} style={{ display: 'flex', marginTop: '10px' }}>
        <input
          type="text"
          value={inputMessage}
          onChange={handleInputChange}
          placeholder="Napisz coś..."
          style={{ flex: 1, padding: '12px', borderRadius: '4px 0 0 4px', border: '1px solid #ccc', outline: 'none' }}
        />
        <button
          type="submit"
          style={{ padding: '12px 24px', background: '#007bff', color: '#fff', border: 'none', borderRadius: '0 4px 4px 0', cursor: 'pointer', fontWeight: 'bold' }}
        >
          Wyślij
        </button>
      </form>
    </div>
  );
};

