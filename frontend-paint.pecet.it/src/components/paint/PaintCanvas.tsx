import React, { useRef, useState, useEffect } from 'react';
import type { Pixel } from '../../types';
import { paintDataSendTimestampMs } from '../../config';


interface PaintCanvasProps {
    onSendPixelUpdate: (pixels: Pixel[]) => void;
    incomingPixels: Pixel[] | null;
}

export const PaintCanvas: React.FC<PaintCanvasProps> = ({
    onSendPixelUpdate,
    incomingPixels,
}) => {
    const mainCanvasRef = useRef<HTMLCanvasElement>(null);
    const bufferCanvasRef = useRef<HTMLCanvasElement>(null);

    // Nowe stany do obsługi tekstu i wyboru narzędzia
    const [tool, setTool] = useState<'draw' | 'text'>('draw');
    const [textValue, setTextValue] = useState<string>('Hello world');
    const [fontSize, setFontSize] = useState<number>(24);

    const [color, setColor] = useState<string>('#000000');
    const [brushSize, setBrushSize] = useState<number>(4);
    const [isDrawing, setIsDrawing] = useState<boolean>(false);
    const lastPos = useRef<{ x: number; y: number } | null>(null);

    const drawLine = (ctx: CanvasRenderingContext2D, x1: number, y1: number, x2: number, y2: number, c: string, size: number) => {
        ctx.strokeStyle = c;
        ctx.lineWidth = size;
        ctx.lineCap = 'round';
        ctx.lineJoin = 'round';
        ctx.beginPath();
        ctx.moveTo(x1, y1);
        ctx.lineTo(x2, y2);
        ctx.stroke();
        ctx.closePath();
    };

    // Nowa funkcja do rysowania tekstu
    const drawText = (ctx: CanvasRenderingContext2D, x: number, y: number, text: string, c: string, size: number) => {
        ctx.fillStyle = c;
        ctx.font = `${size}px Arial`;
        ctx.fillText(text, x, y);
    };

    useEffect(() => {
        if (incomingPixels && mainCanvasRef.current) {
            const ctx = mainCanvasRef.current.getContext('2d');
            if (!ctx) return;
            incomingPixels.forEach((p) => {
                ctx.fillStyle = p.color;
                ctx.fillRect(p.x, p.y, 1, 1);
            });
        }
    }, [incomingPixels]);

    useEffect(() => {
        const interval = setInterval(() => {
            const bufferCanvas = bufferCanvasRef.current;
            if (!bufferCanvas) return;

            const ctx = bufferCanvas.getContext('2d', { willReadFrequently: true });
            if (!ctx) return;

            const { width, height } = bufferCanvas;
            const imageData = ctx.getImageData(0, 0, width, height);
            const data = imageData.data;
            const changedPixels: Pixel[] = [];

            for (let i = 0; i < data.length; i += 4) {
                const alpha = data[i + 3];
                if (alpha > 0) {
                    const r = data[i];
                    const g = data[i + 1];
                    const b = data[i + 2];

                    const hexColor = '#' +
                        r.toString(16).padStart(2, '0') +
                        g.toString(16).padStart(2, '0') +
                        b.toString(16).padStart(2, '0') +
                        alpha.toString(16).padStart(2, '0');

                    const pixelIndex = i / 4;
                    const x = pixelIndex % width;
                    const y = Math.floor(pixelIndex / width);
                    changedPixels.push({ x, y, color: hexColor });
                }
            }
            if (changedPixels.length > 0) {
                onSendPixelUpdate(changedPixels);
                ctx.clearRect(0, 0, width, height);
            }
        }, paintDataSendTimestampMs);

        return () => clearInterval(interval);
    }, [onSendPixelUpdate]);

    const handleMouseEvent = (e: React.MouseEvent<HTMLCanvasElement>, action: 'start' | 'draw' | 'end') => {
        const mainCanvas = mainCanvasRef.current;
        const bufferCanvas = bufferCanvasRef.current;
        if (!mainCanvas || !bufferCanvas) return;

        const rect = mainCanvas.getBoundingClientRect();
        const x = Math.round(e.clientX - rect.left);
        const y = Math.round(e.clientY - rect.top);

        const mainCtx = mainCanvas.getContext('2d');
        const bufferCtx = bufferCanvas.getContext('2d');
        if (!mainCtx || !bufferCtx) return;

        if (action === 'start') {
            if (tool === 'draw') {
                setIsDrawing(true);
                lastPos.current = { x, y };
                drawLine(mainCtx, x, y, x, y, color, brushSize);
                drawLine(bufferCtx, x, y, x, y, color, brushSize);
            } else if (tool === 'text') {
                // Rysuje tekst w miejscu kliknięcia
                drawText(mainCtx, x, y, textValue, color, fontSize);
                drawText(bufferCtx, x, y, textValue, color, fontSize);
            }
            return;
        }

        if (action === 'end') {
            setIsDrawing(false);
            lastPos.current = null;
            return;
        }

        if (action === 'draw' && isDrawing && lastPos.current && tool === 'draw') {
            drawLine(mainCtx, lastPos.current.x, lastPos.current.y, x, y, color, brushSize);
            drawLine(bufferCtx, lastPos.current.x, lastPos.current.y, x, y, color, brushSize);
            lastPos.current = { x, y };
        }
    };

    return (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '1rem' }}>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: '1rem', padding: '10px', background: '#f0f0f0', borderRadius: '8px', justifyContent: 'center' }}>
                <label>
                    Narzędzie:
                    <select value={tool} onChange={(e) => setTool(e.target.value as 'draw' | 'text')} style={{ marginLeft: '5px' }}>
                        <option value="draw">Rysowanie</option>
                        <option value="text">Tekst</option>
                    </select>
                </label>

                <label>
                    Kolor:
                    <input type="color" value={color} onChange={(e) => setColor(e.target.value)} style={{ marginLeft: '5px' }} />
                </label>

                {tool === 'draw' && (
                    <label>
                        Grubość pędzla: {brushSize}px
                        <input type="range" min="1" max="10"
                            value={brushSize} onChange={(e) => setBrushSize(Number(e.target.value))} style={{ marginLeft: '5px' }} />
                    </label>
                )}

                {tool === 'text' && (
                    <>
                        <label>
                            Tekst:
                            <input type="text" value={textValue} onChange={(e) => setTextValue(e.target.value)} style={{ marginLeft: '5px' }} />
                        </label>
                        <label>
                            Rozmiar: {fontSize}px
                            <input type="range" min="5" max="20" value={fontSize} onChange={(e) => setFontSize(Number(e.target.value))} style={{ marginLeft: '5px' }} />
                        </label>
                    </>
                )}
            </div>

            <div style={{ position: 'relative', width: 800, height: 600 }}>
                <canvas
                    ref={mainCanvasRef}
                    width={800}
                    height={600}
                    style={{ position: 'absolute', top: 0, left: 0, border: '2px solid #333', cursor: tool === 'text' ? 'text' : 'crosshair', backgroundColor: '#fff' }}
                    onMouseDown={(e) => handleMouseEvent(e, 'start')}
                    onMouseMove={(e) => handleMouseEvent(e, 'draw')}
                    onMouseUp={(e) => handleMouseEvent(e, 'end')}
                    onMouseOut={(e) => handleMouseEvent(e, 'end')}
                />
                <canvas
                    ref={bufferCanvasRef}
                    width={800}
                    height={600}
                    style={{ display: 'none' }}
                />
            </div>
        </div>
    );
};