import * as Tone from 'tone';

export type SynthType = 'subtractive' | 'fm';

class SynthEngine {
    // Dwa osobne syntezatory polifoniczne
    private subtractiveSynth: Tone.PolySynth;
    private fmSynth: Tone.PolySynth;
    private currentType: SynthType = 'subtractive';

    constructor() {
        // 1. Konfiguracja Syntezatora Subtraktywnego (Oscylator bogaty w harmonika + Filtr)
        this.subtractiveSynth = new Tone.PolySynth(Tone.Synth, {
            oscillator: { type: 'sawtooth' }, // Piła ma dużo wyższych składowych, idealna do filtrowania
            envelope: {
                attack: 0.05,
                decay: 0.2,
                sustain: 0.6,
                release: 0.8,
            },
        }).toDestination();

        // Dodajemy filtr dolnoprzepustowy (Lowpass) charakterystyczny dla syntezy subtraktywnej
        const filter = new Tone.Filter(2000, 'lowpass').toDestination();
        this.subtractiveSynth.connect(filter);

        // 2. Konfiguracja Syntezatora FM (Częstotliwość modulowana innym oscylatorem)
        this.fmSynth = new Tone.PolySynth(Tone.FMSynth, {
            harmonicity: 3, // Stosunek częstotliwości modulatora do nośnej
            modulationIndex: 10, // Głębokość modulacji (jasność brzmienia)
            oscillator: { type: 'sine' },
            envelope: {
                attack: 0.01,
                decay: 0.3,
                sustain: 0.4,
                release: 0.5,
            },
        }).toDestination();
    }

    // Aktywacja kontekstu audio (musi być wywołana po interakcji użytkownika)
    public async init() {
        if (Tone.context.state !== 'running') {
            await Tone.start();
            console.log('Audio Context started');
        }
    }

    // Zmiana typu syntezatora
    public setSynthType(type: SynthType) {
        this.currentType = type;
    }

    // Granie nuty (wywoływane lokalnie oraz po odebraniu pakietu przez sieć)
    public triggerAttack(note: string, time?: number) {
        const activeSynth = this.currentType === 'subtractive' ? this.subtractiveSynth : this.fmSynth;
        activeSynth.triggerAttack(note, time ?? Tone.now());
    }

    // Puszczenie nuty
    public triggerRelease(note: string, time?: number) {
        const activeSynth = this.currentType === 'subtractive' ? this.subtractiveSynth : this.fmSynth;
        activeSynth.triggerRelease(note, time ?? Tone.now());
    }

    // Dynamiczna zmiana parametrów (np. przez suwaki udostępniane przez sieć)
    public updateParameter(param: string, value: number) {
        if (this.currentType === 'subtractive') {
            if (param === 'cutoff') {
                // Przykład zmiany odcięcia filtra w syntezatorze subtraktywnym
                // Wymagałoby to wyciągnięcia referencji do filtra (dla uproszczenia modyfikujemy detune oscylatora)
                this.subtractiveSynth.set({ detune: value });
            }
        } else {
            if (param === 'modulationIndex') {
                this.fmSynth.set({ detune: value });
            }
        }
    }
}

// Eksportujemy pojedynczą instancję (Singleton)
export const synthEngine = new SynthEngine();