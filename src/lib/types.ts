export type LiraMood =
	| 'calm'
	| 'happy'
	| 'curious'
	| 'tense'
	| 'magic'
	| 'mechanical'
	| 'aquatic'
	| 'warm';

export type LiraMaterial =
	| 'glass'
	| 'wood'
	| 'bell'
	| 'bubble'
	| 'pluck'
	| 'breath'
	| 'metal'
	| 'soft-noise'
	| 'toy';

export type LiraShape =
	| 'rise'
	| 'fall'
	| 'bounce'
	| 'pulse'
	| 'wave'
	| 'sparkle'
	| 'swell'
	| 'pop'
	| 'orbit';

export interface LiraSketch {
	schema: 'lira.sketch.v1';
	id: string;
	name: string;
	durationSeconds: number;
	mood: LiraMood;
	material: LiraMaterial;
	shape: LiraShape;
	density: number;
	brightness: number;
	softness: number;
	movement: number;
	seed: string;
	createdAt: string;
	updatedAt: string;
}

export interface LiraVariant {
	schema: 'lira.variant.v1';
	id: string;
	sketchId: string;
	seed: string;
	index: number;
	createdAt: string;
	exportIds: string[];
}

export interface LiraExport {
	schema: 'lira.export.v1';
	id: string;
	sketchId: string;
	variantId?: string;
	filename: string;
	format: 'wav' | 'json' | 'zip';
	createdAt: string;
	durationSeconds: number;
	sampleRate: number;
}

export interface LiraProject {
	schema: string;
	name: string;
	author: string;
	description: string;
	sampleRate: number;
	exportFormat: string;
	defaultDuration: number;
	outputDir: string;
	namingPattern: string;
	version: number;
	updatedAt: string;
}

export const MOODS: LiraMood[] = [
	'calm',
	'happy',
	'curious',
	'tense',
	'magic',
	'mechanical',
	'aquatic',
	'warm'
];

export const MATERIALS: LiraMaterial[] = [
	'glass',
	'wood',
	'bell',
	'bubble',
	'pluck',
	'breath',
	'metal',
	'soft-noise',
	'toy'
];

export const SHAPES: LiraShape[] = [
	'rise',
	'fall',
	'bounce',
	'pulse',
	'wave',
	'sparkle',
	'swell',
	'pop',
	'orbit'
];

export const MOOD_LABELS: Record<LiraMood, string> = {
	calm: 'Calm',
	happy: 'Happy',
	curious: 'Curious',
	tense: 'Tense',
	magic: 'Magic',
	mechanical: 'Mechanical',
	aquatic: 'Aquatic',
	warm: 'Warm'
};

export const MATERIAL_LABELS: Record<LiraMaterial, string> = {
	glass: 'Glass',
	wood: 'Wood',
	bell: 'Bell',
	bubble: 'Bubble',
	pluck: 'Pluck',
	breath: 'Breath',
	metal: 'Metal',
	'soft-noise': 'Soft Noise',
	toy: 'Toy'
};

export const MOOD_COLORS: Record<LiraMood, { accent: string; bg: string; border: string; text: string; glow: string }> = {
	calm:       { accent: '#60a5fa', bg: '#1e3a5f', border: '#3b82f6', text: '#93bbfd', glow: 'rgba(96,165,250,0.25)' },
	happy:      { accent: '#fbbf24', bg: '#5c4a0a', border: '#f59e0b', text: '#fcd34d', glow: 'rgba(251,191,36,0.25)' },
	curious:    { accent: '#34d399', bg: '#0f3d2e', border: '#10b981', text: '#6ee7b7', glow: 'rgba(52,211,153,0.25)' },
	tense:      { accent: '#f87171', bg: '#4a1a1a', border: '#ef4444', text: '#fca5a5', glow: 'rgba(248,113,113,0.25)' },
	magic:      { accent: '#a78bfa', bg: '#2d1b69', border: '#8b5cf6', text: '#c4b5fd', glow: 'rgba(167,139,250,0.25)' },
	mechanical: { accent: '#a8a29e', bg: '#292524', border: '#78716c', text: '#d6d3d1', glow: 'rgba(168,162,158,0.25)' },
	aquatic:    { accent: '#22d3ee', bg: '#0f3d4a', border: '#06b6d4', text: '#67e8f9', glow: 'rgba(34,211,238,0.25)' },
	warm:       { accent: '#fb923c', bg: '#4a2a0a', border: '#f97316', text: '#fdba74', glow: 'rgba(251,146,60,0.25)' },
};

export const SHAPE_LABELS: Record<LiraShape, string> = {
	rise: 'Rise',
	fall: 'Fall',
	bounce: 'Bounce',
	pulse: 'Pulse',
	wave: 'Wave',
	sparkle: 'Sparkle',
	swell: 'Swell',
	pop: 'Pop',
	orbit: 'Orbit'
};
