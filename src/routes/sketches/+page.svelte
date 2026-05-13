<script lang="ts">
	import { onMount } from 'svelte';
	import { sketches, loadSketches, saveSketch, deleteSketch, createSketch } from '$lib/stores/project';
	import { generateAudio, playAudioBuffer } from '$lib/audio/engine';
	import SketchCard from '$lib/components/SketchCard.svelte';
	import { goto } from '$app/navigation';
	import type { LiraSketch } from '$lib/types';

	let playingId: string | null = null;
	let stopFn: (() => void) | null = null;

	onMount(() => {
		loadSketches();
	});

	function play(sketch: LiraSketch) {
		if (playingId === sketch.id) {
			stopFn?.();
			playingId = null;
			stopFn = null;
			return;
		}
		stopFn?.();

		generateAudio(sketch).then((buffer) => {
			if (playingId === sketch.id) {
				stopFn = playAudioBuffer(buffer);
			}
		});
		playingId = sketch.id;
	}

	function editSketch(sketch: LiraSketch) {
		goto(`/sketches/${sketch.id}`);
	}

	async function duplicateSketch(sketch: LiraSketch) {
		const clone = createSketch({
			...sketch,
			id: `sketch-${Date.now()}`,
			name: `${sketch.name} (copy)`,
			createdAt: new Date().toISOString(),
			updatedAt: new Date().toISOString()
		});
		await saveSketch(clone);
		await loadSketches();
	}

	async function addNewSketch() {
		const sketch = createSketch();
		await saveSketch(sketch);
		await loadSketches();
		goto(`/sketches/${sketch.id}`);
	}
</script>

<div class="mx-auto max-w-4xl p-6">
	<div class="mb-8 flex items-center justify-between">
		<div>
			<h1 class="m-0 text-2xl font-bold text-stone-100">Sound Sketch Gallery</h1>
			<p class="mt-1 text-sm text-stone-500">
				Generate tiny sounds, jingles, and sonic logos without knowing music theory.
			</p>
		</div>
		<button
			on:click={addNewSketch}
			class="rounded-xl bg-violet-700 px-5 py-2.5 text-sm font-semibold text-white transition-all hover:bg-violet-600"
		>
			+ New sketch
		</button>
	</div>

	<div class="grid gap-4 sm:grid-cols-2">
		{#each $sketches as sketch (sketch.id)}
			<SketchCard
				{sketch}
				isPlaying={playingId === sketch.id}
				onPlay={() => play(sketch)}
				onEdit={() => editSketch(sketch)}
				onDuplicate={() => duplicateSketch(sketch)}
			/>
		{/each}
	</div>

	{#if $sketches.length === 0}
		<div class="rounded-xl border border-dashed border-stone-800 p-12 text-center">
			<p class="text-stone-500">No sketches yet.</p>
			<p class="mt-2 text-sm text-stone-600">Create your first sound sketch to get started.</p>
		</div>
	{/if}
</div>
