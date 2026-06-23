import React, { useRef, useState, useEffect } from 'react';

export interface Pixel {
    x: number;
    y: number;
    color: string;
}

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

    const [color, setColor] = useState<string>('#000000');
    const [brushSize, setBrushSize] = useState<number>(5);
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
                    const hexColor = '#' + ((1 << 24) + (r << 16) + (g << 8) + b).toString(16).slice(1);
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
        }, 1000);

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
            setIsDrawing(true);
            lastPos.current = { x, y };
            drawLine(mainCtx, x, y, x, y, color, brushSize);
            drawLine(bufferCtx, x, y, x, y, color, brushSize);
            return;
        }

        if (action === 'end') {
            setIsDrawing(false);
            lastPos.current = null;
            return;
        }

        if (action === 'draw' && isDrawing && lastPos.current) {
            drawLine(mainCtx, lastPos.current.x, lastPos.current.y, x, y, color, brushSize);
            drawLine(bufferCtx, lastPos.current.x, lastPos.current.y, x, y, color, brushSize);
            lastPos.current = { x, y };
        }
    };

    return (
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '1rem' }}>
            <div style={{ display: 'flex', gap: '1rem', padding: '10px', background: '#f0f0f0', borderRadius: '8px' }}>
                <label>
                    Kolor:
                    <input type="color" value={color} onChange={(e) => setColor(e.target.value)} style={{ marginLeft: '5px' }} />
                </label>
                <label>
                    Grubość: {brushSize}px
                    <input type="range" min="1" max="50" value={brushSize} onChange={(e) => setBrushSize(Number(e.target.value))} style={{ marginLeft: '5px' }} />
                </label>
            </div>

            <div style={{ position: 'relative', width: 800, height: 600 }}>
                <canvas
                    ref={mainCanvasRef}
                    width={800}
                    height={600}
                    style={{ position: 'absolute', top: 0, left: 0, border: '2px solid #333', cursor: 'crosshair', backgroundColor: '#fff' }}
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