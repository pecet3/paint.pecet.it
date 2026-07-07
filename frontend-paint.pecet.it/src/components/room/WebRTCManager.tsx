import React, { useEffect, useRef, useState, useCallback, forwardRef, useImperativeHandle } from "react";
import type { Event, RoomUser, WebRTCSignalPayload } from "../../types";
import { useStore } from "../../Store";

const STUN_SERVERS = {
    iceServers: [
        { urls: "stun:stun.l.google.com:19302" },
        { urls: "stun:stun1.l.google.com:19302" },
    ],
};

interface WebRTCManagerProps {
    users: RoomUser[];
    onSendSignal: (payload: WebRTCSignalPayload) => void;
    onDataReceived?: (payload: Event) => void;
}
export interface WebRTCManagerHandle {
    receiveSignal: (signal: WebRTCSignalPayload) => void;
    broadcastData: (payload: Event) => void;
}
interface PeerData {
    pc: RTCPeerConnection;
    iceQueue: RTCIceCandidateInit[];
    dataChannel?: RTCDataChannel;
}

// ----------------------------------------------------------------------
// Pomocniczy komponent wideo
// ----------------------------------------------------------------------
const VideoPlayer: React.FC<{ stream: MediaStream | null; muted?: boolean; label: string }> = ({ stream, muted = false, label }) => {
    const videoRef = useRef<HTMLVideoElement>(null);

    useEffect(() => {
        if (videoRef.current && stream) {
            videoRef.current.srcObject = stream;
        }
    }, [stream]);

    return (
        <div className="relative bg-gray-800 rounded-xl overflow-hidden border
         border-gray-700 flex items-center justify-center aspect-video shadow-sm">
            {stream ? (
                <video
                    ref={videoRef}
                    autoPlay
                    playsInline
                    muted={muted}
                    className="w-full h-full object-cover"
                />
            ) : (
                <span className="text-gray-500 text-sm font-medium animate-pulse">Connecting</span>
            )}
            <div className="absolute bottom-2 left-2 bg-black/70 backdrop-blur-sm text-gray-200 text-xs px-2 py-1 rounded-md max-w-[80%] truncate">
                {label}
            </div>
        </div>
    );
};

// ----------------------------------------------------------------------
// Główny komponent WebRTCManager
// ----------------------------------------------------------------------
export const WebRTCManager = forwardRef<WebRTCManagerHandle, WebRTCManagerProps>(({
    users,
    onSendSignal,
    onDataReceived,
}, ref) => {
    const { user } = useStore();
    const localUserUuid = user?.uuid || "";

    const localStreamRef = useRef<MediaStream | null>(null);
    const peersRef = useRef<Map<string, PeerData>>(new Map());

    const signalQueueRef = useRef<WebRTCSignalPayload[]>([]);
    const isProcessingRef = useRef<boolean>(false);

    const [isMediaReady, setIsMediaReady] = useState(false);
    const [localStreamDisplay, setLocalStreamDisplay] = useState<MediaStream | null>(null);
    const [remoteStreams, setRemoteStreams] = useState<Record<string, MediaStream>>({});
    const [errorMsg, setErrorMsg] = useState<string | null>(null);

    // 1. Inicjalizacja lokalnego strumienia (niska jakość)
    useEffect(() => {
        let isMounted = true;

        const initMedia = async () => {
            try {
                const stream = await navigator.mediaDevices.getUserMedia({
                    video: { width: { ideal: 320, max: 480 }, height: { ideal: 240, max: 360 }, frameRate: { ideal: 10, max: 15 } },
                    audio: { sampleRate: 16000, echoCancellation: true, noiseSuppression: true }
                });

                if (isMounted) {
                    localStreamRef.current = stream;
                    setLocalStreamDisplay(stream);
                    setIsMediaReady(true);
                } else {
                    stream.getTracks().forEach(track => track.stop());
                }
            } catch (err) {
                console.error("Błąd kamery/mikrofonu:", err);
                setErrorMsg("No access to camera or microphone");
            }
        };

        if (localUserUuid) initMedia();

        return () => {
            isMounted = false;
            localStreamRef.current?.getTracks().forEach(track => track.stop());
            localStreamRef.current = null;
        };
    }, [localUserUuid]);

    // 2. Tworzenie PeerConnection i czyszczenie w razie awarii ICE
    const createPeer = useCallback((targetUuid: string) => {
        const pc = new RTCPeerConnection(STUN_SERVERS);
        const peerData: PeerData = { pc, iceQueue: [] };
        peersRef.current.set(targetUuid, peerData);

        const setupDataChannel = (channel: RTCDataChannel) => {
            channel.onmessage = (event) => {
                try {
                    const evt = JSON.parse(event.data);
                    if (onDataReceived) {
                        onDataReceived(evt as Event);
                    }
                } catch (e) {
                    console.error("Błąd parsowania danych WebRTC", e);
                }
            };
        };

        // Tworzymy kanał lokalnie (jako strona inicjująca połączenie)
        const dataChannel = pc.createDataChannel("general_data_channel");
        setupDataChannel(dataChannel);
        peerData.dataChannel = dataChannel;

        // Odbieramy kanał utworzony przez drugą stronę
        pc.ondatachannel = (event) => {
            peerData.dataChannel = event.channel;
            setupDataChannel(event.channel);
        };

        if (localStreamRef.current) {
            localStreamRef.current.getTracks().forEach(track => {
                pc.addTrack(track, localStreamRef.current!);
            });
        }

        pc.onicecandidate = (event) => {
            if (event.candidate) {
                onSendSignal({ targetUuid, senderUuid: localUserUuid, signalType: "ice", data: event.candidate });
            }
        };

        pc.ontrack = (event) => {
            if (event.streams && event.streams[0]) {
                setRemoteStreams(prev => ({ ...prev, [targetUuid]: event.streams[0] }));
            }
        };

        // Auto-leczenie: jeśli połączenie padnie na poziomie sieci, usuwamy je
        pc.oniceconnectionstatechange = () => {
            if (pc.iceConnectionState === 'failed' || pc.iceConnectionState === 'disconnected') {
                console.warn(`Połączenie z ${targetUuid} zerwane (stan: ${pc.iceConnectionState}). Czyścimy...`);
                pc.close();
                peersRef.current.delete(targetUuid);
                setRemoteStreams(prev => {
                    const newState = { ...prev };
                    delete newState[targetUuid];
                    return newState;
                });
            }
        };

        return peerData;
    }, [localUserUuid, onSendSignal]);

    // 3. Obsługa nowych użytkowników
    useEffect(() => {
        if (!localUserUuid || !isMediaReady) return;

        users.forEach(u => {
            if (u.uuid !== localUserUuid && u.is_connected && !peersRef.current.has(u.uuid)) {
                const peerData = createPeer(u.uuid);


                peerData.pc.createOffer()
                    .then(offer => peerData.pc.setLocalDescription(offer))
                    .then(() => {
                        onSendSignal({ targetUuid: u.uuid, senderUuid: localUserUuid, signalType: "offer", data: peerData.pc.localDescription });
                    })
                    .catch(err => console.error("Błąd tworzenia oferty dla", u.uuid, err));

            }
        });

        const activeUuids = new Set(users.filter(u => u.is_connected).map(u => u.uuid));
        peersRef.current.forEach((peerData, targetUuid) => {
            if (!activeUuids.has(targetUuid)) {
                peerData.pc.close();
                peersRef.current.delete(targetUuid);
                setRemoteStreams(prev => {
                    const newState = { ...prev };
                    delete newState[targetUuid];
                    return newState;
                });
            }
        });
    }, [users, localUserUuid, isMediaReady, createPeer, onSendSignal]);

    // 4. Masywna zmiana: Rygorystyczny procesor sygnałów z MUTEXEM
    const processSignalQueue = useCallback(async () => {
        // Jeśli kamera nie jest gotowa LUB już przetwarzamy kolejkę, uciekamy.
        if (!isMediaReady || isProcessingRef.current || signalQueueRef.current.length === 0) return;

        isProcessingRef.current = true; // ZAKŁADAMY BLOKADĘ (MUTEX)

        try {
            // Przetwarzaj pętlą while, upewniając się, że wykonujemy po jednym zadaniu na raz
            while (signalQueueRef.current.length > 0) {
                const signal = signalQueueRef.current.shift();
                if (!signal) continue;

                const { senderUuid, signalType, data } = signal;

                let peerData = peersRef.current.get(senderUuid);
                if (!peerData) {
                    peerData = createPeer(senderUuid);
                }

                const { pc, iceQueue } = peerData;

                try {
                    if (signalType === "offer") {
                        // Dodatkowe zabezpieczenie: ignoruj ofertę, jeśli my już wysłaliśmy (signaling state collision)
                        if (pc.signalingState !== "stable" && pc.signalingState !== "have-local-offer") {
                            console.warn("Zignorowano ofertę od", senderUuid, "z powodu stanu", pc.signalingState);
                            continue;
                        }
                        await pc.setRemoteDescription(new RTCSessionDescription(data));
                        const answer = await pc.createAnswer();
                        await pc.setLocalDescription(answer);
                        onSendSignal({ targetUuid: senderUuid, senderUuid: localUserUuid, signalType: "answer", data: pc.localDescription });

                        while (iceQueue.length > 0) {
                            const ice = iceQueue.shift();
                            if (ice) await pc.addIceCandidate(new RTCIceCandidate(ice)).catch(e => console.warn(e));
                        }
                    }
                    else if (signalType === "answer") {
                        if (pc.signalingState === "have-local-offer") {
                            await pc.setRemoteDescription(new RTCSessionDescription(data));
                            while (iceQueue.length > 0) {
                                const ice = iceQueue.shift();
                                if (ice) await pc.addIceCandidate(new RTCIceCandidate(ice)).catch(e => console.warn(e));
                            }
                        }
                    }
                    else if (signalType === "ice") {
                        const candidate = new RTCIceCandidate(data);
                        if (pc.remoteDescription) {
                            await pc.addIceCandidate(candidate).catch(e => console.warn("ICE error", e));
                        } else {
                            iceQueue.push(data);
                        }
                    }
                } catch (err) {
                    console.error(`Błąd przetwarzania sygnału ${signalType} od ${senderUuid}:`, err);
                }
            }
        } finally {
            // ZDEJMUJEMY BLOKADĘ bez względu na to czy wystąpił błąd
            isProcessingRef.current = false;
        }
    }, [isMediaReady, localUserUuid, createPeer, onSendSignal]);

    // Dodawanie do kolejki
    useImperativeHandle(ref, () => ({
        receiveSignal: (signal: WebRTCSignalPayload) => {
            if (signal.targetUuid !== localUserUuid) return;

            // Wrzucamy sygnał bezpośrednio do referencji kolejki
            signalQueueRef.current.push(signal);
            // I odpalamy procesor
            processSignalQueue();
        },
        broadcastData: (event: Event) => {
            peersRef.current.forEach((peerData) => {
                const dc = peerData.dataChannel;
                if (dc && dc.readyState === "open") {
                    dc.send(JSON.stringify({ type: event.type, payload: event.payload }));
                }
            });
        }
    }), [localUserUuid, processSignalQueue]);

    // 5. Zwracany UI
    return (
        <div className="w-full flex flex-col space-y-4">
            {errorMsg && (
                <div className="p-3 bg-red-900/50 border border-red-500 text-red-200 text-sm rounded-lg">
                    {errorMsg}
                </div>
            )}

            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
                <VideoPlayer stream={localStreamDisplay} muted={true} label="You" />

                {Object.entries(remoteStreams).map(([uuid, stream]) => {
                    const userMeta = users.find(u => u.uuid === uuid);
                    const displayName = userMeta ? userMeta.name : `Gość (${uuid.substring(0, 4)})`;
                    return <VideoPlayer key={uuid} stream={stream} label={displayName} />;
                })}
            </div>
        </div>
    );
});