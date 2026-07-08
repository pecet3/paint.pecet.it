import type { Pixel } from "../../types";

export const encodePixelsToBase64 = (pixels: Pixel[]): string => {
    const buffer = new ArrayBuffer(pixels.length * 8);
    const view = new DataView(buffer);

    pixels.forEach((p, i) => {
        const offset = i * 8;

        view.setUint16(offset, p.x, true);
        view.setUint16(offset + 2, p.y, true);

        const hex = p.color.replace('#', '');
        const r = parseInt(hex.substring(0, 2), 16) || 0;
        const g = parseInt(hex.substring(2, 4), 16) || 0;
        const b = parseInt(hex.substring(4, 6), 16) || 0;
        const a = hex.length === 8 ? parseInt(hex.substring(6, 8), 16) : 255;

        view.setUint8(offset + 4, r);
        view.setUint8(offset + 5, g);
        view.setUint8(offset + 6, b);
        view.setUint8(offset + 7, a);
    });

    const bytes = new Uint8Array(buffer);
    let binary = '';
    for (let i = 0; i < bytes.byteLength; i++) {
        binary += String.fromCharCode(bytes[i]);
    }

    return btoa(binary);
};

export const decodeBase64ToPixels = (base64: string): Pixel[] => {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);

    for (let i = 0; i < binary.length; i++) {
        bytes[i] = binary.charCodeAt(i);
    }

    const view = new DataView(bytes.buffer);
    const pixels: Pixel[] = [];

    for (let i = 0; i < bytes.length; i += 8) {
        const x = view.getUint16(i, true);
        const y = view.getUint16(i + 2, true);

        const r = view.getUint8(i + 4);
        const g = view.getUint8(i + 5);
        const b = view.getUint8(i + 6);
        const a = view.getUint8(i + 7);

        const color = `rgba(${r}, ${g}, ${b}, ${a / 255})`;
        pixels.push({ x, y, color });
    }

    return pixels;
};