import React, { useEffect, useRef, useState, useCallback } from "react";
import type { RoomUser, WebRTCSignalPayload } from "../../types";
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
    const { user } = useStore();
    const localUserUuid = user?.uuid || "";

    const localVideoRef = useRef<HTMLVideoElement>(null);
    const [localStream, setLocalStream] = useState<MediaStream | null>(null);
    const [remoteStreams, setRemoteStreams] = useState<Record<string, MediaStream>>({});

    const [isAudioEnabled, setIsAudioEnabled] = useState(true);
    const [isVideoEnabled, setIsVideoEnabled] = useState(true);
    const [mutedRemotes, setMutedRemotes] = useState<Record<string, boolean>>({});

    const peerConnections = useRef<Map<string, RTCPeerConnection>>(new Map());
    const pendingCandidates = useRef<Map<string, RTCIceCandidateInit[]>>(new Map());

    useEffect(() => {
        let isMounted = true;

        navigator.mediaDevices
            .getUserMedia({ video: true, audio: true })
            .then((stream) => {
                if (!isMounted) return;
                setLocalStream(stream);
                if (localVideoRef.current) {
                    localVideoRef.current.srcObject = stream;
                }
            })
            .catch((err) => console.error(err));

        return () => {
            isMounted = false;
            localStream?.getTracks().forEach((track) => track.stop());
            peerConnections.current.forEach((pc) => pc.close());
            peerConnections.current.clear();
            pendingCandidates.current.clear();
        };
    }, []);

    const toggleLocalAudio = () => {
        if (localStream) {
            const audioTracks = localStream.getAudioTracks();
            audioTracks.forEach((track) => {
                track.enabled = !track.enabled;
            });
            setIsAudioEnabled(!isAudioEnabled);
        }
    };

    const toggleLocalVideo = () => {
        if (localStream) {
            const videoTracks = localStream.getVideoTracks();
            videoTracks.forEach((track) => {
                track.enabled = !track.enabled;
            });
            setIsVideoEnabled(!isVideoEnabled);
        }
    };

    const toggleRemoteMute = (uuid: string) => {
        setMutedRemotes((prev) => ({
            ...prev,
            [uuid]: !prev[uuid],
        }));
    };

    const createPeerConnection = useCallback((targetUuid: string, stream: MediaStream | null) => {
        if (peerConnections.current.has(targetUuid)) {
            return peerConnections.current.get(targetUuid)!;
        }

        const pc = new RTCPeerConnection(STUN_SERVERS);
        peerConnections.current.set(targetUuid, pc);
        pendingCandidates.current.set(targetUuid, []);

        if (stream) {
            stream.getTracks().forEach((track) => {
                pc.addTrack(track, stream);
            });
        }

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
            setRemoteStreams((prev) => {
                const existingStream = prev[targetUuid] || new MediaStream();
                existingStream.addTrack(event.track);
                return {
                    ...prev,
                    [targetUuid]: existingStream,
                };
            });
        };

        pc.oniceconnectionstatechange = () => {
            const state = pc.iceConnectionState;
            if (state === "failed" || state === "disconnected" || state === "closed") {
                pc.close();
                peerConnections.current.delete(targetUuid);
                pendingCandidates.current.delete(targetUuid);

                setRemoteStreams((prev) => {
                    const next = { ...prev };
                    delete next[targetUuid];
                    return next;
                });
            }
        };

        pc.onnegotiationneeded = async () => {
            try {
                const offer = await pc.createOffer();
                await pc.setLocalDescription(offer);
                onSendSignal({
                    targetUuid,
                    senderUuid: localUserUuid,
                    signalType: "offer",
                    data: pc.localDescription,
                });
            } catch (err) {
                console.error(err);
            }
        };

        return pc;
    }, [localUserUuid, onSendSignal]);

    useEffect(() => {
        if (!localStream) return;

        users.forEach((user) => {
            console.log(user)
            createPeerConnection(user.uuid, localStream);
        });

        const currentUserUuids = new Set(users.map((u) => u.uuid));
        for (const [uuid, pc] of peerConnections.current.entries()) {
            if (!currentUserUuids.has(uuid)) {
                pc.close();
                peerConnections.current.delete(uuid);
                pendingCandidates.current.delete(uuid);

                setRemoteStreams((prev) => {
                    const next = { ...prev };
                    delete next[uuid];
                    return next;
                });
                setMutedRemotes((prev) => {
                    const next = { ...prev };
                    delete next[uuid];
                    return next;
                });
            }
        }
    }, [users, localStream, localUserUuid, createPeerConnection]);

    useEffect(() => {
        if (!incomingSignal) return;
        if (incomingSignal.targetUuid !== localUserUuid) return;

        const { senderUuid, signalType, data } = incomingSignal;
        const pc = createPeerConnection(senderUuid, localStream);

        const processSignal = async () => {
            try {
                if (signalType === "offer") {
                    await pc.setRemoteDescription(new RTCSessionDescription(data));

                    const candidates = pendingCandidates.current.get(senderUuid) || [];
                    for (const candidate of candidates) {
                        await pc.addIceCandidate(new RTCIceCandidate(candidate));
                    }
                    pendingCandidates.current.set(senderUuid, []);

                    const answer = await pc.createAnswer();
                    await pc.setLocalDescription(answer);

                    onSendSignal({
                        targetUuid: senderUuid,
                        senderUuid: localUserUuid,
                        signalType: "answer",
                        data: pc.localDescription,
                    });
                } else if (signalType === "answer") {
                    if (pc.signalingState === "have-local-offer") {
                        await pc.setRemoteDescription(new RTCSessionDescription(data));
                    }
                } else if (signalType === "ice") {
                    if (pc.remoteDescription) {
                        await pc.addIceCandidate(new RTCIceCandidate(data));
                    } else {
                        const candidates = pendingCandidates.current.get(senderUuid) || [];
                        candidates.push(data);
                        pendingCandidates.current.set(senderUuid, candidates);
                    }
                }
            } catch (err) {
                console.error(err);
            }
        };

        processSignal();
    }, [incomingSignal, localStream, localUserUuid, createPeerConnection, onSendSignal]);

    return (
        <div className="flex flex-wrap gap-2 p-2 bg-slate-900 rounded-lg shadow-md m-2">
            <div className="relative w-48 h-36 bg-black rounded border border-blue-500 overflow-hidden">
                <video
                    ref={localVideoRef}
                    autoPlay
                    playsInline
                    muted
                    className="w-full h-full object-cover"
                />
                <span className="absolute bottom-1 left-1 bg-black/60 text-white text-xs px-1 rounded">
                    you
                </span>
                <div className="absolute bottom-1 right-1 flex gap-1">
                    <button
                        onClick={toggleLocalAudio}
                        className={`text-xs px-1.5 py-0.5 rounded text-white ${isAudioEnabled ? "bg-blue-600/80" : "bg-red-600/80"}`}
                    >
                        {isAudioEnabled ? "Mute" : "Unmute"}
                    </button>
                    <button
                        onClick={toggleLocalVideo}
                        className={`text-xs px-1.5 py-0.5 rounded text-white ${isVideoEnabled ? "bg-blue-600/80" : "bg-red-600/80"}`}
                    >
                        {isVideoEnabled ? "Cam Off" : "Cam On"}
                    </button>
                </div>
            </div>

            {Object.entries(remoteStreams).length > 0 && (
                <div className="flex flex-wrap gap-2">
                    {Object.entries(remoteStreams).map(([uuid, stream]) => (
                        <RemoteVideo
                            key={uuid}
                            stream={stream}
                            user={users.find((u) => u.uuid === uuid)}
                            isMuted={!!mutedRemotes[uuid]}
                            onToggleMute={() => toggleRemoteMute(uuid)}
                        />
                    ))}
                </div>
            )}
        </div>
    );
};

const RemoteVideo: React.FC<{
    stream: MediaStream;
    user?: RoomUser;
    isMuted: boolean;
    onToggleMute: () => void;
}> = ({ stream, user, isMuted, onToggleMute }) => {
    const ref = useRef<HTMLVideoElement>(null);

    useEffect(() => {
        if (ref.current) {
            ref.current.srcObject = stream;
        }
    }, [stream]);

    return (
        <div className="relative w-48 h-36 bg-black rounded border border-blue-500 overflow-hidden">
            <video
                ref={ref}
                autoPlay
                playsInline
                muted={isMuted}
                className="w-full h-full object-cover"
            />
            <span className="absolute bottom-1 left-1 bg-black/60 text-white text-xs px-1 rounded">
                {user ? user.name.slice(0, 16) : "user"}
            </span>
            <div className="absolute bottom-1 right-1">
                <button
                    onClick={onToggleMute}
                    className={`text-xs px-1.5 py-0.5 rounded text-white ${!isMuted ? "bg-blue-600/80" : "bg-red-600/80"}`}
                >
                    {isMuted ? "Unmute" : "Mute"}
                </button>
            </div>
        </div>
    );
};