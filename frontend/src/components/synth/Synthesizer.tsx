import React, { useState, useEffect, useCallback } from 'react';
import { synthEngine, type SynthType } from './synthengine';
import type { NoteEvent } from '../../types';

interface SynthesizerProps {
    onSendNote: (event: NoteEvent) => void;
    incomingNote: NoteEvent | null;
    currentUserId: string;
}

const NOTES = [
    'C3', 'C#3', 'D3', 'D#3', 'E3', 'F3', 'F#3', 'G3', 'G#3', 'A3', 'A#3', 'B3',
    'C4', 'C#4', 'D4', 'D#4', 'E4', 'F4', 'F#4', 'G4', 'G#4', 'A4', 'A#4', 'B4',
    'C5'
];

const KEY_TO_NOTE: Record<string, string> = {
    'z': 'C3', 's': 'C#3', 'x': 'D3', 'd': 'D#3', 'c': 'E3', 'v': 'F3', 'g': 'F#3', 'b': 'G3', 'h': 'G#3', 'n': 'A3', 'j': 'A#3', 'm': 'B3',
    ',': 'C4', 'q': 'C4', '2': 'C#4', 'w': 'D4', '3': 'D#4', 'e': 'E4', 'r': 'F4', '5': 'F#4', 't': 'G4', '6': 'G#4', 'y': 'A4', '7': 'A#4', 'u': 'B4',
    'i': 'C5'
};

export const Synthesizer: React.FC<SynthesizerProps> = ({
    onSendNote,
    incomingNote,
    currentUserId,
}) => {
    const [synthType, setSynthType] = useState<SynthType>('subtractive');
    const [audioInitialized, setAudioInitialized] = useState(false);
    const [activeNotes, setActiveNotes] = useState<Record<string, boolean>>({});

    const handleInitAudio = async () => {
        await synthEngine.init();
        setAudioInitialized(true);
    };

    useEffect(() => {
        if (!incomingNote || !audioInitialized) return;
        if (incomingNote.userId === currentUserId) return;

        synthEngine.setSynthType(incomingNote.synthType);

        if (incomingNote.type === 'attack') {
            synthEngine.triggerAttack(incomingNote.note);
            setActiveNotes((prev) => ({ ...prev, [incomingNote.note]: true }));
        } else {
            synthEngine.triggerRelease(incomingNote.note);
            setActiveNotes((prev) => ({ ...prev, [incomingNote.note]: false }));
        }

        synthEngine.setSynthType(synthType);
    }, [incomingNote, audioInitialized, synthType, currentUserId]);

    const handleTypeChange = (type: SynthType) => {
        setSynthType(type);
        synthEngine.setSynthType(type);
    };

    const handleKeyDown = useCallback((note: string) => {
        if (!audioInitialized) return;

        synthEngine.triggerAttack(note);
        setActiveNotes((prev) => ({ ...prev, [note]: true }));

        onSendNote({
            note,
            type: 'attack',
            synthType,
            userId: currentUserId,
        });
    }, [audioInitialized, synthType, currentUserId, onSendNote]);

    const handleKeyUp = useCallback((note: string) => {
        if (!audioInitialized) return;

        synthEngine.triggerRelease(note);
        setActiveNotes((prev) => ({ ...prev, [note]: false }));

        onSendNote({
            note,
            type: 'release',
            synthType,
            userId: currentUserId,
        });
    }, [audioInitialized, synthType, currentUserId, onSendNote]);

    useEffect(() => {
        const handleWindowKeyDown = (e: KeyboardEvent) => {
            if (e.repeat) return;
            const note = KEY_TO_NOTE[e.key.toLowerCase()];
            if (note && !activeNotes[note]) {
                handleKeyDown(note);
            }
        };

        const handleWindowKeyUp = (e: KeyboardEvent) => {
            const note = KEY_TO_NOTE[e.key.toLowerCase()];
            if (note && activeNotes[note]) {
                handleKeyUp(note);
            }
        };

        window.addEventListener('keydown', handleWindowKeyDown);
        window.addEventListener('keyup', handleWindowKeyUp);

        return () => {
            window.removeEventListener('keydown', handleWindowKeyDown);
            window.removeEventListener('keyup', handleWindowKeyUp);
        };
    }, [handleKeyDown, handleKeyUp, activeNotes]);

    return (
        <div className="p-6 bg-slate-900 text-slate-100 rounded-xl shadow-lg border border-slate-800 w-full max-w-4xl">
            {!audioInitialized ? (
                <button
                    onClick={handleInitAudio}
                    className="w-full py-4 px-6 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg shadow-md transition-colors duration-200"
                >
                    🚀 Uruchom Audio (Wymagane)
                </button>
            ) : (
                <div className="space-y-6">
                    <div className="flex gap-6 p-3 bg-slate-800/50 rounded-lg border border-slate-700/50">
                        <label className="flex items-center gap-2 cursor-pointer text-sm font-medium select-none">
                            <input
                                type="radio"
                                name="synthType"
                                checked={synthType === 'subtractive'}
                                onChange={() => handleTypeChange('subtractive')}
                                className="w-4 h-4 text-indigo-600 bg-slate-700 border-slate-600 focus:ring-indigo-500"
                            />
                            Subtraktywny <span className="text-xs text-slate-400">(Sawtooth + Filter)</span>
                        </label>

                        <label className="flex items-center gap-2 cursor-pointer text-sm font-medium select-none">
                            <input
                                type="radio"
                                name="synthType"
                                checked={synthType === 'fm'}
                                onChange={() => handleTypeChange('fm')}
                                className="w-4 h-4 text-indigo-600 bg-slate-700 border-slate-600 focus:ring-indigo-500"
                            />
                            FM Synth <span className="text-xs text-slate-400">(Sine Modulated)</span>
                        </label>
                    </div>

                    <div className="overflow-x-auto bg-slate-950 rounded-lg border border-slate-800 shadow-inner p-4">
                        <div className="flex items-start justify-center min-w-max px-2">
                            {NOTES.map((note) => {
                                const isActive = activeNotes[note];
                                const isBlack = note.includes('#');

                                return (
                                    <button
                                        key={note}
                                        onMouseDown={() => handleKeyDown(note)}
                                        onMouseUp={() => handleKeyUp(note)}
                                        onMouseLeave={() => {
                                            if (isActive) handleKeyUp(note);
                                        }}
                                        className={`
                                            flex items-end justify-center font-bold text-xs select-none transition-all duration-100
                                            ${isBlack
                                                ? `w-7 md:w-8 h-24 pb-2 rounded-b-sm z-10 -mx-3.5 md:-mx-4 ${isActive ? 'bg-indigo-600 text-white translate-y-1 shadow-none' : 'bg-slate-900 text-slate-300 shadow-[0_4px_0_#0f172a]'
                                                }`
                                                : `w-11 md:w-12 h-40 pb-4 rounded-b-md z-0 ${isActive ? 'bg-indigo-200 text-slate-900 translate-y-1 shadow-none' : 'bg-white text-slate-900 hover:bg-slate-100 shadow-[0_4px_0_#cbd5e1]'
                                                }`
                                            }
                                        `}
                                    >
                                        <span className={isBlack ? 'mb-2 opacity-50' : 'opacity-70'}>{note}</span>
                                    </button>
                                );
                            })}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};