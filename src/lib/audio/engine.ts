import type { LiraMaterial, LiraShape, LiraSketch } from '$lib/types';

// --- Seeded PRNG (mulberry32) ---
export function createRNG(seed: number): () => number {
	let s = seed | 0;
	return () => {
		s = (s + 0x6d2b79f5) | 0;
		let t = Math.imul(s ^ (s >>> 15), 1 | s);
		t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}

export function seedFromString(str: string): number {
	let hash = 0;
	for (let i = 0; i < str.length; i++) {
		const ch = str.charCodeAt(i);
		hash = ((hash << 5) - hash + ch) | 0;
	}
	return hash;
}

// --- Audio generation ---

interface SoundEvent {
	startTime: number;
	duration: number;
	frequency: number;
	pan: number;
	gain: number;
	oscType: OscillatorType;
}

function materialOscTypes(material: LiraMaterial): OscillatorType[] {
	switch (material) {
		case 'glass':
			return ['sine', 'sine'];
		case 'bell':
			return ['triangle', 'sine'];
		case 'bubble':
			return ['sine'];
		case 'pluck':
			return ['triangle'];
		case 'breath':
			return ['sawtooth'];
		case 'metal':
			return ['square', 'square'];
		case 'wood':
			return ['sine'];
		case 'soft-noise':
			return ['sawtooth'];
		case 'toy':
			return ['square'];
	}
}

function materialBaseFreq(material: LiraMaterial): number {
	switch (material) {
		case 'glass':
			return 1200;
		case 'bell':
			return 800;
		case 'bubble':
			return 600;
		case 'pluck':
			return 400;
		case 'breath':
			return 300;
		case 'metal':
			return 500;
		case 'wood':
			return 350;
		case 'soft-noise':
			return 200;
		case 'toy':
			return 700;
	}
}

function shapeContour(shape: LiraShape, t: number, duration: number): number {
	const phase = t / duration;
	switch (shape) {
		case 'rise':
			return phase;
		case 'fall':
			return 1 - phase;
		case 'bounce':
			return Math.abs(Math.sin(phase * Math.PI * 3)) * (1 - phase * 0.5);
		case 'pulse':
			return Math.sin(phase * Math.PI * 2) * 0.5 + 0.5;
		case 'wave':
			return Math.sin(phase * Math.PI * 4) * 0.5 + 0.5;
		case 'sparkle': {
			const spike = Math.sin(phase * Math.PI * 8) * Math.exp(-phase * 3);
			return Math.abs(spike);
		}
		case 'swell':
			return Math.sin(phase * Math.PI) * (1 - Math.exp(-phase * 5));
		case 'pop':
			return Math.exp(-phase * 6);
		case 'orbit':
			return Math.sin(phase * Math.PI * 2) * 0.5 + 0.5;
	}
}

function moodMultiplier(mood: string): number {
	switch (mood) {
		case 'calm':
			return 0.7;
		case 'happy':
			return 1.1;
		case 'curious':
			return 1.0;
		case 'tense':
			return 0.9;
		case 'magic':
			return 1.05;
		case 'mechanical':
			return 0.85;
		case 'aquatic':
			return 0.75;
		case 'warm':
			return 0.8;
		default:
			return 1.0;
	}
}

export function generateSoundEvents(sketch: LiraSketch, variantSeed?: string): SoundEvent[] {
	const seedVal = seedFromString(variantSeed ?? sketch.seed);
	const rng = createRNG(seedVal);

	const duration = Math.max(0.2, Math.min(10, sketch.durationSeconds));
	const density = Math.max(1, Math.min(10, sketch.density));
	const brightness = Math.max(1, Math.min(10, sketch.brightness));
	const softness = Math.max(1, Math.min(10, sketch.softness));
	const movement = Math.max(1, Math.min(10, sketch.movement));
	const moodFactor = moodMultiplier(sketch.mood);

	const oscTypes = materialOscTypes(sketch.material);
	const baseFreq = materialBaseFreq(sketch.material);

	const eventCount = Math.floor(density * 3 + rng() * 3);
	const events: SoundEvent[] = [];

	for (let i = 0; i < eventCount; i++) {
		const t = i / Math.max(1, eventCount - 1);
		const shapeVal = shapeContour(sketch.shape, t * duration, duration);

		const freqRange = brightness * 80;
		const freqOffset = (rng() - 0.5) * 2 * freqRange * (movement / 10);
		const frequency = Math.max(100, baseFreq * moodFactor + freqOffset * shapeVal);

		const startJitter = (rng() - 0.5) * 0.08 * (movement / 10);
		const startTime = Math.max(0, t * duration + startJitter);

		const eventDuration = Math.max(0.05, (duration / eventCount) * (1.5 - softness * 0.1) + rng() * 0.1);

		const pan = (rng() - 0.5) * 2 * (movement / 10) * shapeVal;
		const gain = Math.max(0.05, (0.6 - density * 0.03) * (0.5 + shapeVal * 0.5));
		const oscType = oscTypes[Math.floor(rng() * oscTypes.length)];

		events.push({
			startTime,
			duration: eventDuration,
			frequency,
			pan,
			gain,
			oscType
		});
	}

	return events;
}

export async function renderToAudioBuffer(
	events: SoundEvent[],
	totalDuration: number,
	sampleRate: number = 44100
): Promise<AudioBuffer> {
	const ctx = new OfflineAudioContext(2, Math.ceil(totalDuration * sampleRate), sampleRate);

	for (const evt of events) {
		const osc = ctx.createOscillator();
		const gainNode = ctx.createGain();
		const panner = ctx.createStereoPanner();

		osc.type = evt.oscType;
		osc.frequency.value = evt.frequency;

		const attack = Math.min(0.02, evt.duration * 0.15);
		const release = Math.min(0.15, evt.duration * 0.4);
		gainNode.gain.setValueAtTime(0, evt.startTime);
		gainNode.gain.linearRampToValueAtTime(evt.gain, evt.startTime + attack);
		gainNode.gain.linearRampToValueAtTime(0, evt.startTime + evt.duration);

		panner.pan.value = evt.pan;

		osc.connect(gainNode);
		gainNode.connect(panner);
		panner.connect(ctx.destination);

		osc.start(evt.startTime);
		osc.stop(evt.startTime + evt.duration + release);
	}

	return ctx.startRendering();
}

export async function generateAudio(
	sketch: LiraSketch,
	variantSeed?: string,
	sampleRate: number = 44100
): Promise<AudioBuffer> {
	const events = generateSoundEvents(sketch, variantSeed);
	return renderToAudioBuffer(events, sketch.durationSeconds, sampleRate);
}

export function playAudioBuffer(buffer: AudioBuffer): () => void {
	const ctx = new AudioContext();
	const source = ctx.createBufferSource();
	source.buffer = buffer;
	source.connect(ctx.destination);
	source.start();

	return () => {
		source.stop();
		ctx.close();
	};
}
