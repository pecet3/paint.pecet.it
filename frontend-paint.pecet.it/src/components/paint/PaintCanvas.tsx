import React, { useRef, useState, useEffect, useCallback } from 'react';
import { type DrawEvent, type DrawAction, type Point } from '../../dto';

interface MultiplayerPaintProps {
    // Funkcja wywoływana, gdy użytkownik coś narysuje (podłącz to do ws.send)
    onSendDrawEvent: (event: DrawEvent) => void;
    // Najnowsze zdarzenie odebrane z serwera (podłącz to do ws.onmessage)
    incomingDrawEvent?: DrawEvent | null;
}

export const PaintCanvas: React.FC<MultiplayerPaintProps> = ({
    onSendDrawEvent,
    incomingDrawEvent,
}) => {
    const canvasRef = useRef<HTMLCanvasElement>(null);

    // Stan narzędzi
    const [color, setColor] = useState<string>('#000000');
    const [brushSize, setBrushSize] = useState<number>(5);
    const [isDrawing, setIsDrawing] = useState<boolean>(false);

    // Funkcja pomocnicza do pobierania współrzędnych myszy
    const getCoordinates = (e: React.MouseEvent | MouseEvent): Point | null => {
        if (!canvasRef.current) return null;
        const canvas = canvasRef.current;
        const rect = canvas.getBoundingClientRect();
        return {
            x: e.clientX - rect.left,
            y: e.clientY - rect.top,
        };
    };

    // Główna funkcja wykonująca rysowanie na Canvas (używana lokalnie i zdalnie)
    const drawOnCanvas = useCallback((event: DrawEvent) => {
        const canvas = canvasRef.current;
        if (!canvas) return;
        const ctx = canvas.getContext('2d');
        if (!ctx) return;

        ctx.strokeStyle = event.color;
        ctx.lineWidth = event.size;
        ctx.lineCap = 'round';
        ctx.lineJoin = 'round';

        switch (event.action) {
            case 'start':
                ctx.beginPath();
                ctx.moveTo(event.point.x, event.point.y);
                break;
            case 'draw':
                ctx.lineTo(event.point.x, event.point.y);
                ctx.stroke();
                break;
            case 'end':
                ctx.closePath();
                break;
        }
    }, []);

    // Obsługa zdarzeń z serwera
    useEffect(() => {
        if (incomingDrawEvent) {
            drawOnCanvas(incomingDrawEvent);
        }
    }, [incomingDrawEvent, drawOnCanvas]);

    // --- Handlery zdarzeń myszy (Lokalne rysowanie) ---

    const handleMouseDown = (e: React.MouseEvent<HTMLCanvasElement>) => {
        const point = getCoordinates(e);
        if (!point) return;

        setIsDrawing(true);
        const event: DrawEvent = { action: 'start', point, color, size: brushSize };

        drawOnCanvas(event); // Rysuj u siebie
        onSendDrawEvent(event); // Wyślij do innych
    };

    const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
        if (!isDrawing) return;
        const point = getCoordinates(e);
        if (!point) return;

        const event: DrawEvent = { action: 'draw', point, color, size: brushSize };

        drawOnCanvas(event); // Rysuj u siebie
        onSendDrawEvent(event); // Wyślij do innych
    };

    const handleMouseUp = (e: React.MouseEvent<HTMLCanvasElement>) => {
        if (!isDrawing) return;
        setIsDrawing(false);

        const point = getCoordinates(e);
        if (!point) return;

        const event: DrawEvent = { action: 'end', point, color, size: brushSize };

        drawOnCanvas(event);
        onSendDrawEvent(event);
    };

    const handleMouseOut = (e: React.MouseEvent<HTMLCanvasElement>) => {
        if (isDrawing) {
            handleMouseUp(e);
        }
    };

    return (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '1rem' }}>

            {/* Pasek narzędzi */}
            <div style={{ display: 'flex', gap: '1rem', padding: '10px', background: '#f0f0f0', borderRadius: '8px' }}>
                <label>
                    Kolor:
                    <input
                        type="color"
                        value={color}
                        onChange={(e) => setColor(e.target.value)}
                        style={{ marginLeft: '5px' }}
                    />
                </label>

                <label>
                    Grubość pędzla: {brushSize}px
                    <input
                        type="range"
                        min="1"
                        max="50"
                        value={brushSize}
                        onChange={(e) => setBrushSize(Number(e.target.value))}
                        style={{ marginLeft: '5px' }}
                    />
                </label>
            </div>

            {/* Obszar roboczy */}
            <canvas
                ref={canvasRef}
                width={800}
                height={600}
                style={{ border: '2px solid #333', cursor: 'crosshair', backgroundColor: '#fff' }}
                onMouseDown={handleMouseDown}
                onMouseMove={handleMouseMove}
                onMouseUp={handleMouseUp}
                onMouseOut={handleMouseOut}
            />

        </div>
    );
};