import React, { useEffect, useRef, useState, useCallback } from "react";
import type { RoomUser, WebRTCSignalPayload } from "../../types"
import { useStore } from "../../Store";

const STUN_SERVERS = {
    iceServers: [
        { urls: "stun:stun.l.google.com:19302" },
        { urls: "stun:stun1.l.google.com:19302" },
    ],
};

interface WebRTCManagerProps {
    users: RoomUser[];
    incomingSignal: WebRTCSignalPayload | null;
    onSendSignal: (payload: WebRTCSignalPayload) => void;
}

export const WebRTCManager: React.FC<WebRTCManagerProps> = ({
    users,
    incomingSignal,
    onSendSignal,
}) => {
    const { user } = useStore()
    const localUserUuid = user?.uuid || ""

    const localVideoRef = useRef<HTMLVideoElement>(null);
    const [localStream, setLocalStream] = useState<MediaStream | null>(null);
    const [remoteStreams, setRemoteStreams] = useState<Record<string, MediaStream>>({});

    const peerConnections = useRef<Map<string, RTCPeerConnection>>(new Map());

    useEffect(() => {
        navigator.mediaDevices
            .getUserMedia({ video: true, audio: true })
            .then((stream) => {
                setLocalStream(stream);
                if (localVideoRef.current) {
                    localVideoRef.current.srcObject = stream;
                }
            })
            .catch((err) => console.error("WebRTC getUserMedia error:", err));

        return () => {
            localStream?.getTracks().forEach((track) => track.stop());
            peerConnections.current.forEach((pc) => pc.close());
        };
    }, []);

    const createPeerConnection = useCallback((targetUuid: string, stream: MediaStream) => {
        const pc = new RTCPeerConnection(STUN_SERVERS);
        console.log(pc)
        stream.getTracks().forEach((track) => {
            pc.addTrack(track, stream);
        });

        pc.onicecandidate = (event) => {
            if (event.candidate) {
                onSendSignal({
                    targetUuid,
                    senderUuid: localUserUuid,
                    signalType: "ice",
                    data: event.candidate,
                });
            }
        };

        pc.ontrack = (event) => {
            setRemoteStreams((prev) => ({
                ...prev,
                [targetUuid]: event.streams[0],
            }));
        };

        peerConnections.current.set(targetUuid, pc);
        return pc;
    }, [localUserUuid, onSendSignal]);

    useEffect(() => {
        if (!localStream) return;

        users.forEach(async (user) => {
            if (user.uuid === localUserUuid) return;
            if (peerConnections.current.has(user.uuid)) return;

            const pc = createPeerConnection(user.uuid, localStream);


            const offer = await pc.createOffer();
            await pc.setLocalDescription(offer);

            onSendSignal({
                targetUuid: user.uuid,
                senderUuid: localUserUuid,
                signalType: "offer",
                data: offer,
            });

        });

        const currentUserUuids = new Set(users.map(u => u.uuid));
        for (const [uuid, pc] of peerConnections.current.entries()) {
            if (!currentUserUuids.has(uuid)) {
                pc.close();
                peerConnections.current.delete(uuid);
                setRemoteStreams(prev => {
                    const next = { ...prev };
                    delete next[uuid];
                    return next;
                });
            }
        }
    }, [users, localStream, localUserUuid, createPeerConnection, onSendSignal]);

    useEffect(() => {
        if (!incomingSignal || !localStream) return;

        if (incomingSignal.targetUuid !== localUserUuid) return;

        const { senderUuid, signalType, data } = incomingSignal;

        let pc = peerConnections.current.get(senderUuid);
        if (!pc) {
            pc = createPeerConnection(senderUuid, localStream);
        }

        const processSignal = async () => {
            try {
                if (signalType === "offer") {
                    await pc!.setRemoteDescription(new RTCSessionDescription(data));
                    const answer = await pc!.createAnswer();
                    await pc!.setLocalDescription(answer);

                    onSendSignal({
                        targetUuid: senderUuid,
                        senderUuid: localUserUuid,
                        signalType: "answer",
                        data: answer,
                    });
                } else if (signalType === "answer") {
                    await pc!.setRemoteDescription(new RTCSessionDescription(data));
                } else if (signalType === "ice") {
                    await pc!.addIceCandidate(new RTCIceCandidate(data));
                }
            } catch (err) {
                console.error("Signal processing error:", err);
            }
        };

        processSignal();
    }, [incomingSignal, localStream, localUserUuid, createPeerConnection, onSendSignal]);

    return (
        <div className="flex flex-col gap-4 p-4 bg-gray-900 rounded-lg shadow-md">
            <h3 className="text-white font-bold">Kamera (Ty)</h3>
            <video
                ref={localVideoRef}
                autoPlay
                muted
                playsInline
                className="w-48 h-36 bg-black rounded border border-gray-600 object-cover"
            />

            {Object.entries(remoteStreams).length > 0 && (
                <>
                    <h3 className="text-white font-bold mt-2">Inni użytkownicy</h3>
                    <div className="flex flex-wrap gap-2">
                        {Object.entries(remoteStreams).map(([uuid, stream]) => (
                            <RemoteVideo key={uuid} stream={stream} uuid={uuid} />
                        ))}
                    </div>
                </>
            )}
        </div>
    );
};

const RemoteVideo: React.FC<{ stream: MediaStream; uuid: string }> = ({ stream, uuid }) => {
    const ref = useRef<HTMLVideoElement>(null);

    useEffect(() => {
        if (ref.current) {
            ref.current.srcObject = stream;
        }
    }, [stream]);

    return (
        <div className="relative w-48 h-36">
            <video
                ref={ref}
                autoPlay
                playsInline
                className="w-full h-full bg-black rounded border border-blue-500 object-cover"
            />
            <span className="absolute bottom-1 left-1 bg-black/60 text-white text-xs px-1 rounded">
                {uuid.slice(0, 5)}...
            </span>
        </div>
    );
};