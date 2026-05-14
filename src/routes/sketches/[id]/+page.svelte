<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import {
		sketches,
		loadSketches,
		saveSketch,
		loadProject,
		project,
		recordExport
	} from '$lib/stores/project';
	import { generateAudio, playAudioBuffer } from '$lib/audio/engine';
	import { downloadWav, downloadSketchJSON, downloadZip, makeExportRecord } from '$lib/audio/export';
	import ParamControl from '$lib/components/ParamControl.svelte';
	import SketchCanvas from '$lib/components/SketchCanvas.svelte';
	import {
		MOODS,
		MATERIALS,
		SHAPES,
		MOOD_LABELS,
		MATERIAL_LABELS,
		SHAPE_LABELS,
		MOOD_COLORS
	} from '$lib/types';
	import type { LiraMood, LiraShape, LiraSketch } from '$lib/types';

	let sketch: LiraSketch | null = null;
	let isPlaying = false;
	let stopFn: (() => void) | null = null;
	let audioBuffer: AudioBuffer | null = null;
	let isGenerating = false;
	let showCanvas = true;

	const sketchId = $page.params.id;
	$: moodColors = sketch ? MOOD_COLORS[sketch.mood as LiraMood] : MOOD_COLORS.magic;

	onMount(async () => {
		await loadProject();
		await loadSketches();
		const found = $sketches.find((s) => s.id === sketchId);
		if (found) {
			sketch = { ...found };
		} else {
			goto('/sketches');
		}
	});

	function updateField(field: keyof LiraSketch, value: number | string) {
		if (!sketch) return;
		sketch = { ...sketch, [field]: value, updatedAt: new Date().toISOString() };
	}

	function applyDrawingParams(params: {
		shape: LiraShape;
		density: number;
		brightness: number;
		softness: number;
		movement: number;
		pictureSeed: string;
	}) {
		if (!sketch) return;
		sketch = {
			...sketch,
			shape: params.shape,
			density: params.density,
			brightness: params.brightness,
			softness: params.softness,
			movement: params.movement,
			seed: params.pictureSeed,
			updatedAt: new Date().toISOString()
		};
	}

	async function generate() {
		if (!sketch) return;
		isGenerating = true;
		try {
			audioBuffer = await generateAudio(sketch);
		} finally {
			isGenerating = false;
		}
	}

	function playStop() {
		if (isPlaying) {
			stopFn?.();
			stopFn = null;
			isPlaying = false;
			return;
		}
		if (!audioBuffer) return;
		stopFn = playAudioBuffer(audioBuffer);
		isPlaying = true;
		setTimeout(() => {
			isPlaying = false;
			stopFn = null;
		}, sketch?.durationSeconds ? sketch.durationSeconds * 1000 + 500 : 3500);
	}

	async function save() {
		if (!sketch) return;
		sketch.updatedAt = new Date().toISOString();
		await saveSketch(sketch);
		await loadSketches();
	}

	async function exportWav() {
		if (!sketch || !audioBuffer) {
			await generate();
		}
		if (!sketch || !audioBuffer) return;
		const filename = `${sketch.name.replace(/\s+/g, '-').toLowerCase()}.wav`;
		downloadWav(audioBuffer, filename);
		recordExport(makeExportRecord(sketch, filename, 'wav', sketch.durationSeconds, 44100));
	}

	function exportJSON() {
		if (!sketch) return;
		downloadSketchJSON(sketch);
		recordExport(
			makeExportRecord(
				sketch,
				`${sketch.name.replace(/\s+/g, '-').toLowerCase()}.json`,
				'json',
				sketch.durationSeconds,
				44100
			)
		);
	}

	async function exportZip() {
		if (!sketch || !audioBuffer) {
			await generate();
		}
		if (!sketch || !audioBuffer) return;
		await downloadZip(sketch, audioBuffer);
		recordExport(
			makeExportRecord(
				sketch,
				`${sketch.name.replace(/\s+/g, '-').toLowerCase()}.zip`,
				'zip',
				sketch.durationSeconds,
				44100
			)
		);
	}

	function viewVariants() {
		goto(`/sketches/${sketchId}/variants`);
	}
</script>

{#if sketch}
	<div class="mx-auto max-w-5xl p-6">
		<div class="mb-6 flex items-center gap-3">
			<button
				on:click={() => goto('/sketches')}
				class="rounded-lg bg-stone-800 px-3 py-1 text-sm text-stone-400 hover:bg-stone-700 hover:text-stone-200"
			>
				← Gallery
			</button>
			<h1 class="m-0 text-xl font-bold text-stone-100">Edit sketch</h1>
			<div class="ml-auto flex items-center gap-2 text-xs">
				<button
					on:click={() => (showCanvas = !showCanvas)}
					class="rounded-lg px-3 py-1 transition-colors"
					style="background: {showCanvas ? moodColors.bg : '#292524'}; color: {showCanvas ? moodColors.text : '#78716c'}"
				>
					{showCanvas ? '✎ Canvas' : '☰ Sliders only'}
				</button>
			</div>
		</div>

		<div class="grid gap-6 lg:grid-cols-3">
			<!-- Left column: basic params -->
			<div class="space-y-4 lg:col-span-1">
				<div>
					<label class="text-xs font-medium text-stone-400">Name</label>
					<input
						type="text"
						bind:value={sketch.name}
						on:input={() => updateField('name', sketch?.name ?? '')}
						class="mt-1 w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-2 text-stone-200"
					/>
				</div>

				<ParamControl
					label="Duration (seconds)"
					bind:value={sketch.durationSeconds}
					min={0.2}
					max={10}
					step={0.1}
					type="slider"
					onChange={(v) => updateField('durationSeconds', Number(v))}
				/>

				<ParamControl
					label="Mood"
					bind:value={sketch.mood}
					type="select"
					options={MOODS.map((m) => ({ value: m, label: MOOD_LABELS[m] }))}
					onChange={(v) => updateField('mood', String(v))}
				/>

				<ParamControl
					label="Material"
					bind:value={sketch.material}
					type="select"
					options={MATERIALS.map((m) => ({ value: m, label: MATERIAL_LABELS[m] }))}
					onChange={(v) => updateField('material', String(v))}
				/>

				<ParamControl
					label="Shape"
					bind:value={sketch.shape}
					type="select"
					options={SHAPES.map((s) => ({ value: s, label: SHAPE_LABELS[s] }))}
					onChange={(v) => updateField('shape', String(v))}
				/>
			</div>

			<!-- Middle: drawing canvas -->
			<div class="lg:col-span-1">
				<SketchCanvas
					mood={sketch.mood}
					onApply={applyDrawingParams}
				/>
			</div>

			<!-- Right column: fine-tuning sliders -->
			<div class="space-y-4 lg:col-span-1">
				<ParamControl
					label="Density"
					bind:value={sketch.density}
					min={1}
					max={10}
					step={1}
					onChange={(v) => updateField('density', Number(v))}
				/>

				<ParamControl
					label="Brightness"
					bind:value={sketch.brightness}
					min={1}
					max={10}
					step={1}
					onChange={(v) => updateField('brightness', Number(v))}
				/>

				<ParamControl
					label="Softness"
					bind:value={sketch.softness}
					min={1}
					max={10}
					step={1}
					onChange={(v) => updateField('softness', Number(v))}
				/>

				<ParamControl
					label="Movement"
					bind:value={sketch.movement}
					min={1}
					max={10}
					step={1}
					onChange={(v) => updateField('movement', Number(v))}
				/>

				<ParamControl
					label="Seed"
					bind:value={sketch.seed}
					type="text"
					onChange={(v) => updateField('seed', String(v))}
				/>
			</div>
		</div>

		<!-- Actions -->
		<div class="mt-8 flex flex-wrap items-center gap-3 border-t border-stone-800 pt-6">
			<button
				on:click={generate}
				disabled={isGenerating}
				class="rounded-xl px-6 py-2.5 text-sm font-semibold text-white transition-all disabled:opacity-50"
				style="background: {moodColors.accent}; box-shadow: 0 0 20px {moodColors.glow}"
			>
				{isGenerating ? 'Generating...' : '⚡ Generate'}
			</button>

			<button
				on:click={playStop}
				disabled={!audioBuffer}
				class="rounded-xl px-5 py-2.5 text-sm font-medium transition-all {isPlaying
					? 'bg-amber-700 text-white'
					: 'bg-stone-800 text-stone-300 hover:bg-stone-700'} disabled:opacity-30"
			>
				{isPlaying ? '⏹ Stop' : '▶ Play'}
			</button>

			<button
				on:click={save}
				class="rounded-xl bg-stone-800 px-5 py-2.5 text-sm font-medium text-stone-300 transition-all hover:bg-stone-700"
			>
				💾 Save
			</button>

			<button
				on:click={viewVariants}
				class="rounded-xl bg-stone-800 px-5 py-2.5 text-sm font-medium text-stone-300 transition-all hover:bg-stone-700"
			>
				🎲 Variants
			</button>

			<div class="ml-auto flex gap-2">
				<button
					on:click={exportWav}
					disabled={!audioBuffer}
					class="rounded-lg bg-emerald-900 px-4 py-2 text-xs font-medium text-emerald-300 transition-all hover:bg-emerald-800 disabled:opacity-30"
				>
					WAV
				</button>
				<button
					on:click={exportJSON}
					class="rounded-lg bg-stone-700 px-4 py-2 text-xs font-medium text-stone-300 transition-all hover:bg-stone-600"
				>
					JSON
				</button>
				<button
					on:click={exportZip}
					disabled={!audioBuffer}
					class="rounded-lg bg-stone-700 px-4 py-2 text-xs font-medium text-stone-300 transition-all hover:bg-stone-600 disabled:opacity-30"
				>
					ZIP
				</button>
			</div>
		</div>
	</div>
{/if}
