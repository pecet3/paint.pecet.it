import React, { useState, useEffect } from 'react';
import { synthEngine, type SynthType } from './synthengine'
import type { NoteEvent } from '../../types';


interface SynthesizerProps {
    onSendNote: (event: NoteEvent) => void;
    incomingNote: NoteEvent | null;
    currentUserId: string;
}

const NOTES = ['C4', 'D4', 'E4', 'F4', 'G4', 'A4', 'B4', 'C5'];

export const Synthesizer: React.FC<SynthesizerProps> = ({
    onSendNote,
    incomingNote,
    currentUserId,
}) => {
    const [synthType, setSynthType] = useState<SynthType>('subtractive');
    const [audioInitialized, setAudioInitialized] = useState(false);
    // Stan do śledzenia aktywnych nut (dla efektu wizualnego)
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

    const handleKeyDown = (note: string) => {
        if (!audioInitialized) return;

        synthEngine.triggerAttack(note);
        setActiveNotes((prev) => ({ ...prev, [note]: true }));

        onSendNote({
            note,
            type: 'attack',
            synthType,
            userId: currentUserId,
        });
    };

    const handleKeyUp = (note: string) => {
        if (!audioInitialized || !activeNotes[note]) return;

        synthEngine.triggerRelease(note);
        setActiveNotes((prev) => ({ ...prev, [note]: false }));

        onSendNote({
            note,
            type: 'release',
            synthType,
            userId: currentUserId,
        });
    };

    return (
        <div className="p-6 bg-slate-900 text-slate-100 rounded-xl shadow-lg border border-slate-800 max-w-2xl">
            <h3 className="text-xl font-bold mb-4 tracking-wide text-indigo-400">
                Multiplayer Synth Engine
            </h3>

            {!audioInitialized ? (
                <button
                    onClick={handleInitAudio}
                    className="w-full py-4 px-6 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg shadow-md transition-colors duration-200"
                >
                    🚀 Uruchom Audio (Wymagane)
                </button>
            ) : (
                <div className="space-y-6">
                    {/* Wybór metody syntezy */}
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

                    {/* Wizualna Klawiatura */}
                    <div className="flex gap-1.5 p-4 bg-slate-950 rounded-lg border border-slate-800 shadow-inner overflow-x-auto justify-center">
                        {NOTES.map((note) => {
                            const isActive = activeNotes[note];
                            return (
                                <button
                                    key={note}
                                    onMouseDown={() => handleKeyDown(note)}
                                    onMouseUp={() => handleKeyUp(note)}
                                    onMouseLeave={() => handleKeyUp(note)}
                                    className={`
                    w-14 h-40 pb-4 flex items-end justify-center 
                    font-bold text-xs rounded-b-md transition-all duration-700 select-none
                    ${isActive
                                            ? 'bg-indigo-500 text-white translate-y-0.5 shadow-none'
                                            : 'bg-white text-slate-900 hover:bg-slate-100 shadow-[0_4px_0_#cbd5e1]'
                                        }
                  `}
                                >
                                    {note}
                                </button>
                            );
                        })}
                    </div>
                </div>
            )}
        </div>
    );
};