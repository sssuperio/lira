<script lang="ts">
	import { onMount } from 'svelte';
	import {
		project,
		loadProject,
		saveProjectSettings,
		createSketch,
		saveSketch,
		loadSketches,
		sketches
	} from '$lib/stores/project';
	import type { LiraProject, LiraSketch } from '$lib/types';
	import { MOODS, MATERIALS, SHAPES, MOOD_LABELS, MATERIAL_LABELS, SHAPE_LABELS } from '$lib/types';

	let proj: LiraProject | null = null;
	let statusMessage = '';

	onMount(async () => {
		await loadProject();
		proj = $project ? { ...$project } : null;
	});

	function updateField(field: keyof LiraProject, value: unknown) {
		if (!proj) return;
		proj = { ...proj, [field]: value };
	}

	async function save() {
		if (!proj) return;
		await saveProjectSettings(proj);
		statusMessage = 'Settings saved.';
		setTimeout(() => (statusMessage = ''), 2000);
	}

	async function seedPresets() {
		await loadSketches();
		if ($sketches.length > 0) {
			statusMessage = 'Sketches already exist — skipping preset seeding.';
			setTimeout(() => (statusMessage = ''), 3000);
			return;
		}

		const presets: Partial<LiraSketch>[] = [
			{
				id: 'tiny-victory',
				name: 'tiny victory',
				durationSeconds: 2.2,
				mood: 'happy',
				material: 'bell',
				shape: 'rise',
				density: 5,
				brightness: 8,
				softness: 5,
				movement: 6
			},
			{
				id: 'soft-notification',
				name: 'soft notification',
				durationSeconds: 1.4,
				mood: 'calm',
				material: 'glass',
				shape: 'pulse',
				density: 3,
				brightness: 6,
				softness: 8,
				movement: 2
			},
			{
				id: 'error-bump',
				name: 'error bump',
				durationSeconds: 0.8,
				mood: 'tense',
				material: 'wood',
				shape: 'pop',
				density: 4,
				brightness: 5,
				softness: 3,
				movement: 5
			},
			{
				id: 'curious-sparkle',
				name: 'curious sparkle',
				durationSeconds: 3,
				mood: 'curious',
				material: 'bubble',
				shape: 'sparkle',
				density: 8,
				brightness: 9,
				softness: 6,
				movement: 8
			},
			{
				id: 'calm-intro',
				name: 'calm intro',
				durationSeconds: 4,
				mood: 'calm',
				material: 'breath',
				shape: 'swell',
				density: 3,
				brightness: 4,
				softness: 9,
				movement: 3
			},
			{
				id: 'game-coin',
				name: 'game coin',
				durationSeconds: 0.6,
				mood: 'happy',
				material: 'glass',
				shape: 'pop',
				density: 6,
				brightness: 9,
				softness: 3,
				movement: 4
			},
			{
				id: 'magic-reveal',
				name: 'magic reveal',
				durationSeconds: 3.5,
				mood: 'magic',
				material: 'bell',
				shape: 'sparkle',
				density: 7,
				brightness: 8,
				softness: 6,
				movement: 7
			},
			{
				id: 'warm-logo',
				name: 'warm logo',
				durationSeconds: 4,
				mood: 'warm',
				material: 'pluck',
				shape: 'swell',
				density: 4,
				brightness: 5,
				softness: 7,
				movement: 4
			},
			{
				id: 'little-machine',
				name: 'little machine',
				durationSeconds: 2.5,
				mood: 'mechanical',
				material: 'wood',
				shape: 'orbit',
				density: 7,
				brightness: 6,
				softness: 3,
				movement: 7
			},
			{
				id: 'ocean-button',
				name: 'ocean button',
				durationSeconds: 1.8,
				mood: 'aquatic',
				material: 'soft-noise',
				shape: 'wave',
				density: 5,
				brightness: 5,
				softness: 7,
				movement: 6
			}
		];

		for (const preset of presets) {
			const sketch = createSketch(preset);
			await saveSketch(sketch);
		}
		await loadSketches();
		statusMessage = `Seeded ${presets.length} preset sketches!`;
		setTimeout(() => (statusMessage = ''), 3000);
	}
</script>

{#if proj}
	<div class="mx-auto max-w-3xl p-6">
		<h1 class="mb-8 text-2xl font-bold text-stone-100">Project Settings</h1>

		<div class="space-y-5">
			<div>
				<label class="text-xs font-medium text-stone-400">Project name</label>
				<input
					type="text"
					bind:value={proj.name}
					on:input={() => updateField('name', proj?.name ?? '')}
					class="mt-1 w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-2 text-stone-200"
				/>
			</div>

			<div>
				<label class="text-xs font-medium text-stone-400">Author</label>
				<input
					type="text"
					bind:value={proj.author}
					on:input={() => updateField('author', proj?.author ?? '')}
					class="mt-1 w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-2 text-stone-200"
				/>
			</div>

			<div>
				<label class="text-xs font-medium text-stone-400">Description</label>
				<textarea
					bind:value={proj.description}
					on:input={() => updateField('description', proj?.description ?? '')}
					rows={3}
					class="mt-1 w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-2 text-stone-200"
				/>
			</div>

			<div>
				<label class="text-xs font-medium text-stone-400">Sample rate</label>
				<select
					value={proj.sampleRate}
					on:change={(e) => updateField('sampleRate', Number(e.currentTarget.value))}
					class="mt-1 w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-2 text-stone-200"
				>
					<option value={22050}>22050 Hz</option>
					<option value={44100}>44100 Hz</option>
					<option value={48000}>48000 Hz</option>
				</select>
			</div>

			<div>
				<label class="text-xs font-medium text-stone-400">Default export format</label>
				<select
					value={proj.exportFormat}
					on:change={(e) => updateField('exportFormat', e.currentTarget.value)}
					class="mt-1 w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-2 text-stone-200"
				>
					<option value="wav">WAV</option>
					<option value="json">JSON</option>
					<option value="zip">ZIP</option>
				</select>
			</div>

			<div>
				<label class="text-xs font-medium text-stone-400">Default duration (seconds)</label>
				<input
					type="number"
					bind:value={proj.defaultDuration}
					min={0.2}
					max={10}
					step={0.1}
					on:input={() => updateField('defaultDuration', proj?.defaultDuration ?? 2)}
					class="mt-1 w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-2 text-stone-200"
				/>
			</div>

			<div>
				<label class="text-xs font-medium text-stone-400">Output directory</label>
				<input
					type="text"
					bind:value={proj.outputDir}
					on:input={() => updateField('outputDir', proj?.outputDir ?? '')}
					class="mt-1 w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-2 text-stone-200"
				/>
			</div>

			<div>
				<label class="text-xs font-medium text-stone-400">Naming pattern</label>
				<input
					type="text"
					bind:value={proj.namingPattern}
					on:input={() => updateField('namingPattern', proj?.namingPattern ?? '')}
					class="mt-1 w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-2 text-stone-200"
				/>
				<p class="mt-1 text-xs text-stone-600">
					&#123;project&#125;-&#123;sketch&#125;-&#123;variant&#125;-&#123;seed&#125;
				</p>
			</div>
		</div>

		<div class="mt-8 flex items-center gap-4 border-t border-stone-800 pt-6">
			<button
				on:click={save}
				class="rounded-xl bg-violet-700 px-6 py-2.5 text-sm font-semibold text-white transition-all hover:bg-violet-600"
			>
				Save settings
			</button>

			<button
				on:click={seedPresets}
				class="rounded-xl bg-stone-800 px-6 py-2.5 text-sm font-medium text-stone-300 transition-all hover:bg-stone-700"
			>
				🌱 Seed preset sketches
			</button>

			{#if statusMessage}
				<span class="text-sm text-emerald-400">{statusMessage}</span>
			{/if}
		</div>
	</div>
{/if}
