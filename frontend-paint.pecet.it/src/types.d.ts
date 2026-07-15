export interface NoteEvent {
    note: string;
    type: 'attack' | 'release';
    synthType: SynthType;
    userId: string;
}

export interface Pixel {
    x: number;
    y: number;
    color: string;
}
