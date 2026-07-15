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
// Pomocniczy komponent wideo z ikonką mutowania wewnątrz kafelka
// ----------------------------------------------------------------------
const VideoPlayer: React.FC<{
    stream: MediaStream | null;
    muted?: boolean;
    label: string;
    showRemoteControls?: boolean;
    onToggleRemoteMute?: () => void;
}> = ({ stream, muted = false, label, showRemoteControls = false, onToggleRemoteMute }) => {
    const videoRef = useRef<HTMLVideoElement>(null);

    useEffect(() => {
        if (videoRef.current && stream) {
            videoRef.current.srcObject = stream;
        }
    }, [stream]);

    return (
        <div className="relative bg-gray-800 rounded-lg overflow-hidden border
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

            {/* Lewy dolny róg: Nazwa użytkownika */}
            <div className="absolute bottom-1 left-1 bg-black/70 backdrop-blur-sm text-gray-200 text-[10px]
             p-0.5 rounded-md max-w-[60%] truncate">
                {label} {muted && label !== "You" && " (Muted)"}
            </div>

            {/* Prawy dolny róg: Sama ikonka Unicode do mutowania lokatnego */}
            {showRemoteControls && onToggleRemoteMute && (
                <button
                    onClick={onToggleRemoteMute}
                    className={`absolute bottom-1 right-1 p-1.5 rounded-md text-xs backdrop-blur-sm transition-all active:scale-95 ${muted
                        ? "bg-red-600/80 text-white hover:bg-red-600"
                        : "bg-black/70 text-gray-200 hover:bg-black/90"
                        }`}
                    title={muted ? "Unmute user" : "Mute user"}
                >
                    {muted ? "🔈" : "🔊"}
                </button>
            )}
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

    // Stan kontrolek dla samego siebie
    const [isMicEnabled, setIsMicEnabled] = useState(true);
    const [isCamEnabled, setIsCamEnabled] = useState(true);

    // Stan lokalnego wyciszenia zdalnych użytkowników { [uuid]: boolean }
    const [mutedRemoteUsers, setMutedRemoteUsers] = useState<Record<string, boolean>>({});

    // 1. Inicjalizacja lokalnego strumienia
    useEffect(() => {
        let isMounted = true;

        const initMedia = async () => {
            try {
                const stream = await navigator.mediaDevices.getUserMedia({
                    video: { width: { ideal: 320, max: 480 }, height: { ideal: 240, max: 360 }, frameRate: { ideal: 10, max: 15 } },
                    audio: { sampleRate: 16000, echoCancellation: true, noiseSuppression: true },
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

    // Przełączanie własnego mikrofonu
    const toggleLocalMic = () => {
        if (localStreamRef.current) {
            localStreamRef.current.getAudioTracks().forEach(track => {
                track.enabled = !isMicEnabled;
            });
            setIsMicEnabled(!isMicEnabled);
        }
    };

    // Przełączanie własnej kamery
    const toggleLocalCam = () => {
        if (localStreamRef.current) {
            localStreamRef.current.getVideoTracks().forEach(track => {
                track.enabled = !isCamEnabled;
            });
            setIsCamEnabled(!isCamEnabled);
        }
    };

    // Przełączanie wyciszenia kogoś u siebie
    const toggleRemoteMute = (uuid: string) => {
        setMutedRemoteUsers(prev => ({
            ...prev,
            [uuid]: !prev[uuid]
        }));
    };

    // 2. Tworzenie PeerConnection
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

        const dataChannel = pc.createDataChannel("general_data_channel");
        setupDataChannel(dataChannel);
        peerData.dataChannel = dataChannel;

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

        pc.oniceconnectionstatechange = () => {
            if (pc.iceConnectionState === 'failed' || pc.iceConnectionState === 'disconnected') {
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
                        onSendSignal({
                            targetUuid: u.uuid, senderUuid: localUserUuid, signalType: "offer",
                            data: peerData.pc.localDescription
                        });
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

    // 4. Procesor sygnałów z MUTEXEM
    const processSignalQueue = useCallback(async () => {
        if (!isMediaReady || isProcessingRef.current || signalQueueRef.current.length === 0) return;

        isProcessingRef.current = true;

        try {
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
            isProcessingRef.current = false;
        }
    }, [isMediaReady, localUserUuid, createPeer, onSendSignal]);

    useImperativeHandle(ref, () => ({
        receiveSignal: (signal: WebRTCSignalPayload) => {
            if (signal.targetUuid !== localUserUuid) return;
            signalQueueRef.current.push(signal);
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

            {/* Siatka z samymi kafelkami wideo */}
            <div className="grid grid-cols-2 md:grid-cols-3  gap-4">

                {/* Twoje wideo */}
                <VideoPlayer stream={localStreamDisplay} muted={true} label="You" />

                {/* Wideo innych użytkowników z przyciskiem wyciszania w rogu */}
                {Object.entries(remoteStreams).map(([uuid, stream]) => {
                    const userMeta = users.find(u => u.uuid === uuid);
                    const displayName = userMeta ? userMeta.name : `Gość (${uuid.substring(0, 4)})`;
                    const isMutedLocally = !!mutedRemoteUsers[uuid];

                    return (
                        <VideoPlayer
                            key={uuid}
                            stream={stream}
                            label={displayName}
                            muted={isMutedLocally}
                            showRemoteControls={true}
                            onToggleRemoteMute={() => toggleRemoteMute(uuid)}
                        />
                    );
                })}
            </div>

            {/* Panel kontrolny pod WSZYSTKIMI kafelkami wideo (Mutowanie Siebie) */}
            <div className="flex justify-center text-xs items-center space-x-3 pt-2 border-gray-800">
                <button
                    onClick={toggleLocalMic}
                    className={`px-4 py-2  font-medium rounded-xl transition-all active:scale-95 flex items-center border
                         space-x-2 ${isMicEnabled ? "bg-gray-800 text-gray-200 hover:bg-gray-700"
                            : "bg-red-900/60 text-red-200 hover:bg-red-800 border border-red-700"
                        }`}
                >
                    <span>🔈</span>
                    <span>{isMicEnabled ? "Mute Me" : "Unmute Me"}</span>
                </button>

                <button
                    onClick={toggleLocalCam}
                    className={`px-4 py-2  font-medium rounded-xl transition-all border
                        active:scale-95 flex items-center space-x-2
                         ${isCamEnabled ? "bg-gray-800 text-gray-200 hover:bg-gray-700"
                            : "bg-red-900/60 text-red-200 hover:bg-red-800 border border-red-700"
                        }`}
                >
                    <span>📷</span>
                    <span>{isCamEnabled ? "Turn Cam Off" : "Turn Cam On"}</span>
                </button>
            </div>
        </div>
    );
});