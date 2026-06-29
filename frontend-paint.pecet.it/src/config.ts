const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const host = window.location.host;

export const wsAddr = `${protocol}//${host}/ws`;

// export const wsAddr = `ws://localhost:8080/ws`
export const paintDataSendTimestampMs = 30