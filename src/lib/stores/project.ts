import { writable } from 'svelte/store';
import type { LiraSketch, LiraVariant, LiraExport, LiraProject } from '$lib/types';

export const sketches = writable<LiraSketch[]>([]);
export const project = writable<LiraProject | null>(null);
export const exports = writable<LiraExport[]>([]);

const API_BASE = '';

async function apiGet<T>(url: string): Promise<T> {
	const res = await fetch(`${API_BASE}${url}`);
	if (!res.ok) throw new Error(`GET ${url}: ${res.status}`);
	return res.json();
}

async function apiPost<T>(url: string, body: unknown): Promise<T> {
	const res = await fetch(`${API_BASE}${url}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	if (!res.ok) throw new Error(`POST ${url}: ${res.status}`);
	return res.json();
}

async function apiPut<T>(url: string, body: unknown): Promise<T> {
	const res = await fetch(`${API_BASE}${url}`, {
		method: 'PUT',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(body)
	});
	if (!res.ok) throw new Error(`PUT ${url}: ${res.status}`);
	return res.json();
}

async function apiDelete(url: string): Promise<void> {
	const res = await fetch(`${API_BASE}${url}`, { method: 'DELETE' });
	if (!res.ok) throw new Error(`DELETE ${url}: ${res.status}`);
}

export async function loadProject(): Promise<void> {
	const proj = await apiGet<LiraProject>('/api/project?project=default');
	project.set(proj);
}

export async function loadSketches(): Promise<void> {
	const list = await apiGet<LiraSketch[]>('/api/sketches?project=default');
	sketches.set(list);
}

export async function loadExports(): Promise<void> {
	const list = await apiGet<LiraExport[]>('/api/exports?project=default');
	exports.set(list);
}

export async function saveSketch(sketch: LiraSketch): Promise<LiraSketch> {
	const existing = await apiGet<LiraSketch[]>(`/api/sketches?project=default`).catch(() => []);
	const found = existing.find((s) => s.id === sketch.id);

	if (found) {
		return apiPut<LiraSketch>(`/api/sketches/${sketch.id}?project=default`, sketch);
	}
	return apiPost<LiraSketch>('/api/sketches?project=default', sketch);
}

export async function deleteSketch(id: string): Promise<void> {
	await apiDelete(`/api/sketches/${id}?project=default`);
	await loadSketches();
}

export async function loadVariants(sketchId: string): Promise<LiraVariant[]> {
	return apiGet<LiraVariant[]>(`/api/sketches/${sketchId}/variants?project=default`);
}

export async function saveVariant(sketchId: string, variant: LiraVariant): Promise<LiraVariant> {
	return apiPost<LiraVariant>(`/api/sketches/${sketchId}/variants?project=default`, variant);
}

export async function saveProjectSettings(proj: LiraProject): Promise<LiraProject> {
	const result = await apiPut<LiraProject>('/api/project?project=default', proj);
	project.set(result);
	return result;
}

export async function recordExport(exp: LiraExport): Promise<void> {
	await apiPost<LiraExport>('/api/exports?project=default', exp);
	await loadExports();
}

export function createSketch(overrides?: Partial<LiraSketch>): LiraSketch {
	const now = new Date().toISOString();
	const id = `sketch-${Date.now()}`;
	return {
		schema: 'lira.sketch.v1',
		id,
		name: 'New sketch',
		durationSeconds: 3,
		mood: 'happy',
		material: 'glass',
		shape: 'rise',
		density: 5,
		brightness: 7,
		softness: 6,
		movement: 4,
		seed: id,
		createdAt: now,
		updatedAt: now,
		...overrides
	};
}

export function createVariant(sketchId: string, seed: string, index: number): LiraVariant {
	return {
		schema: 'lira.variant.v1',
		id: `var-${sketchId}-${index}`,
		sketchId,
		seed,
		index,
		createdAt: new Date().toISOString(),
		exportIds: []
	};
}
