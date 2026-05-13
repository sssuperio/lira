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
