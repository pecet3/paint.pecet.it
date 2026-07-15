import React, { useRef, useState, useEffect } from 'react';
import type { Pixel, RoomConfig } from '../../types';

interface PaintCanvasProps {
    onSendPixelUpdate: (pixels: Pixel[]) => void;
    onSendPixelUpdateRTC?: (pixels: Pixel[]) => void;
    incomingPixels: Pixel[] | null;
    config: RoomConfig;
}

export const PaintCanvas: React.FC<PaintCanvasProps> = ({
    onSendPixelUpdate,
    incomingPixels,
    onSendPixelUpdateRTC,
    config,
}) => {
    const mainCanvasRef = useRef<HTMLCanvasElement>(null);
    const bufferCanvasRef = useRef<HTMLCanvasElement>(null);
    const bufferCanvasRTCRef = useRef<HTMLCanvasElement>(null);
    const [tool, setTool] = useState<'draw' | 'text'>('draw');
    const [textValue, setTextValue] = useState<string>('Hello world');
    const [fontSize, setFontSize] = useState<number>(24);

    const [color, setColor] = useState<string>('#000000');
    const [brushSize, setBrushSize] = useState<number>(1);
    const [isDrawing, setIsDrawing] = useState<boolean>(false);
    const lastPos = useRef<{ x: number; y: number } | null>(null);

    const [mouseCoords, setMouseCoords] = useState<{ x: number; y: number }>({ x: 0, y: 0 });

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

    const drawText = (ctx: CanvasRenderingContext2D, x: number, y: number, text: string, c: string, size: number) => {
        ctx.fillStyle = c;
        ctx.font = `italic small-caps bold ${size}px 'Courier New', monospace`

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
        if (!incomingPixels && mainCanvasRef.current) {
            const ctx = mainCanvasRef.current.getContext('2d');
            if (!ctx) return;
            ctx.reset()
        }
    }, [incomingPixels]);

    useEffect(() => {
        if (onSendPixelUpdateRTC !== undefined) {
            const interval = setInterval(() => {
                const bufferCanvas = bufferCanvasRTCRef.current;
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
                    onSendPixelUpdateRTC(changedPixels);
                    ctx.clearRect(0, 0, width, height);
                }
            }, 30);

            return () => clearInterval(interval);
        }

    }, [onSendPixelUpdateRTC]);

    const handleMouseEvent = (e: React.MouseEvent<HTMLCanvasElement>, action: 'start' | 'draw' | 'end') => {
        const mainCanvas = mainCanvasRef.current;
        const bufferCanvas = bufferCanvasRef.current;
        const bufferCanvasRTC = bufferCanvasRTCRef.current;
        if (!mainCanvas || !bufferCanvas || !bufferCanvasRTC) return;

        const rect = mainCanvas.getBoundingClientRect();
        const x = Math.round(e.clientX - rect.left);
        const y = Math.round(e.clientY - rect.top);

        setMouseCoords({ x, y });

        const mainCtx = mainCanvas.getContext('2d');
        const bufferCtx = bufferCanvas.getContext('2d', { willReadFrequently: true });
        const bufferRTCCtx = bufferCanvasRTC.getContext('2d', { willReadFrequently: true });
        if (!mainCtx || !bufferCtx || !bufferRTCCtx) return;

        if (action === 'start') {
            if (tool === 'draw') {
                setIsDrawing(true);
                lastPos.current = { x, y };
                drawLine(mainCtx, x, y, x, y, color, brushSize);
                drawLine(bufferCtx, x, y, x, y, color, brushSize);
                drawLine(bufferRTCCtx, x, y, x, y, color, brushSize);
            } else if (tool === 'text') {
                drawText(mainCtx, x, y, textValue, color, fontSize);
                drawText(bufferCtx, x, y, textValue, color, fontSize);
                drawText(bufferRTCCtx, x, y, textValue, color, fontSize);
            }
            return;
        }

        if (action === 'end') {
            setIsDrawing(false);
            lastPos.current = null;

            const { width, height } = bufferCanvas;
            const imageData = bufferCtx.getImageData(0, 0, width, height);
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
                    const px = pixelIndex % width;
                    const py = Math.floor(pixelIndex / width);
                    changedPixels.push({ x: px, y: py, color: hexColor });
                }
            }
            if (changedPixels.length > 0) {
                onSendPixelUpdate(changedPixels);
                bufferCtx.clearRect(0, 0, width, height);
            }
            return;
        }

        if (action === 'draw' && isDrawing && lastPos.current && tool === 'draw') {
            drawLine(mainCtx, lastPos.current.x, lastPos.current.y, x, y, color, brushSize);
            drawLine(bufferCtx, lastPos.current.x, lastPos.current.y, x, y, color, brushSize);
            drawLine(bufferRTCCtx, lastPos.current.x, lastPos.current.y, x, y, color, brushSize);
            lastPos.current = { x, y };
        }
    };

    return (
        <div className="bg-slate-700 p-2 rounded-lg w-full border border-black flex flex-col items-center max-w-4xl max-h-[92vh] h-full">

            <div className='flex items-end justify-between w-full m-auto flex-wrap'>
                <div className="flex gap-1">
                    <label className="flex items-center">
                        Tool:
                        <select value={tool} onChange={(e) => setTool(e.target.value as 'draw' | 'text')}
                            className="ml-2 border rounded ">
                            <option value="draw" className='text-black'>Draw</option>
                            <option value="text" className='text-black'>Text</option>
                        </select>
                    </label>

                    <label className="flex items-center">
                        Color:
                        <input type="color" value={color} onChange={(e) => setColor(e.target.value)}
                            className="ml-2 cursor-pointer h-8 w-8 p-0 border-0" />
                    </label>

                    {tool === 'draw' && (
                        <label className="flex items-center">
                            Brush size: {brushSize}px
                            <input type="range" min="1" max="24" value={brushSize} onChange={(e) => setBrushSize(Number(e.target.value))}
                                className="" />
                        </label>
                    )}

                    {tool === 'text' && (
                        <>
                            <label className="flex items-center">
                                Text:
                                <input type="text" value={textValue} onChange={(e) => setTextValue(e.target.value)} className="ml-2 border rounded p-1" />
                            </label>
                            <label className="flex items-center">
                                Size: {fontSize}px
                                <input type="range" min="5" max="20" value={fontSize} onChange={(e) => setFontSize(Number(e.target.value))} className="ml-2" />
                            </label>
                        </>
                    )}

                </div>
                <div className=" text-sm font-mono tracking-wide">
                    X: {mouseCoords.x.toString().padStart(3, '0')} | Y: {mouseCoords.y.toString().padStart(3, '0')}
                </div>
            </div>
            <div className="flex-1 w-full h-full overflow-auto bg-gray-200">

                <div
                    className="relative shadow-lg mx-auto"
                    style={{ width: config.width, height: config.height }}
                >
                    <canvas
                        ref={mainCanvasRef}
                        width={config.width}
                        height={config.height}
                        className={`absolute top-0 left-0
                             border-gray-800 bg-white ${tool === 'text' ? 'cursor-text' : 'cursor-crosshair'
                            }`}
                        onMouseDown={(e) => handleMouseEvent(e, 'start')}
                        onMouseMove={(e) => handleMouseEvent(e, 'draw')}
                        onMouseUp={(e) => handleMouseEvent(e, 'end')}
                        onMouseOut={(e) => handleMouseEvent(e, 'end')}
                    />

                    <canvas
                        ref={bufferCanvasRef}
                        width={config.width}
                        height={config.height}
                        className="hidden"
                    />

                    <canvas
                        ref={bufferCanvasRTCRef}
                        width={config.width}
                        height={config.height}
                        className="hidden"
                    />
                </div>
            </div>
        </div>
    );
};