<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { sketches, loadSketches, saveVariant, loadVariants, createVariant, recordExport } from '$lib/stores/project';
	import { generateAudio, playAudioBuffer, seedFromString } from '$lib/audio/engine';
	import { downloadWav, makeExportRecord } from '$lib/audio/export';
	import type { LiraSketch, LiraVariant } from '$lib/types';

	let sketch: LiraSketch | null = null;
	let variants: LiraVariant[] = [];
	let playingVariantId: string | null = null;
	let stopFn: (() => void) | null = null;
	let audioBuffers: Map<string, AudioBuffer> = new Map();
	let variantCount = 8;
	let isGenerating = false;

	const sketchId = $page.params.id;

	onMount(async () => {
		await loadSketches();
		sketch = $sketches.find((s) => s.id === sketchId) ?? null;
		if (!sketch) {
			goto('/sketches');
			return;
		}
		const loaded = await loadVariants(sketchId);
		variants = loaded;
	});

	async function generateVariants() {
		if (!sketch) return;
		isGenerating = true;
		variants = [];
		audioBuffers = new Map();

		try {
			for (let i = 0; i < variantCount; i++) {
				const variantSeed = `${sketch.seed}-v${i}-${Date.now()}`;
				const variant = createVariant(sketchId, variantSeed, i);
				const buffer = await generateAudio(sketch, variantSeed);
				audioBuffers.set(variant.id, buffer);
				variants = [...variants, variant];
				await saveVariant(sketchId, variant);
			}
		} finally {
			isGenerating = false;
		}
	}

	function playStop(variant: LiraVariant) {
		if (playingVariantId === variant.id) {
			stopFn?.();
			playingVariantId = null;
			stopFn = null;
			return;
		}
		stopFn?.();
		const buffer = audioBuffers.get(variant.id);
		if (!buffer) return;
		stopFn = playAudioBuffer(buffer);
		playingVariantId = variant.id;
		setTimeout(() => {
			if (playingVariantId === variant.id) {
				playingVariantId = null;
				stopFn = null;
			}
		}, sketch?.durationSeconds ? sketch.durationSeconds * 1000 + 500 : 3500);
	}

	async function exportVariantWav(variant: LiraVariant) {
		if (!sketch) return;
		let buffer = audioBuffers.get(variant.id);
		if (!buffer) {
			buffer = await generateAudio(sketch, variant.seed);
			audioBuffers.set(variant.id, buffer);
		}
		const filename = `${sketch.name.replace(/\s+/g, '-').toLowerCase()}-v${variant.index}.wav`;
		downloadWav(buffer, filename);
		recordExport(
			makeExportRecord(sketch, filename, 'wav', sketch.durationSeconds, 44100, variant.id)
		);
	}

	async function promoteToMain(variant: LiraVariant) {
		if (!sketch) return;
		sketch = { ...sketch, seed: variant.seed, updatedAt: new Date().toISOString() };
		goto(`/sketches/${sketchId}`);
	}
</script>

{#if sketch}
	<div class="mx-auto max-w-4xl p-6">
		<div class="mb-6 flex items-center gap-3">
			<button
				on:click={() => goto(`/sketches/${sketchId}`)}
				class="rounded-lg bg-stone-800 px-3 py-1 text-sm text-stone-400 hover:bg-stone-700 hover:text-stone-200"
			>
				← Editor
			</button>
			<h1 class="m-0 text-xl font-bold text-stone-100">Variants of {sketch.name}</h1>
		</div>

		<div class="mb-6 flex items-center gap-4">
			<label class="text-sm text-stone-400">
				Count:
				<select
					bind:value={variantCount}
					class="ml-2 rounded bg-stone-800 px-2 py-1 text-stone-200 text-sm"
				>
					{#each [1, 2, 4, 8, 12, 16] as n}
						<option value={n}>{n}</option>
					{/each}
				</select>
			</label>
			<button
				on:click={generateVariants}
				disabled={isGenerating}
				class="rounded-xl bg-violet-700 px-5 py-2 text-sm font-semibold text-white transition-all hover:bg-violet-600 disabled:opacity-50"
			>
				{isGenerating ? 'Generating...' : '🎲 Generate variants'}
			</button>
		</div>

		<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
			{#each variants as variant (variant.id)}
				<div
					class="rounded-xl border border-stone-800 bg-stone-900 p-4 transition-all hover:border-violet-700/40"
				>
					<div class="mb-2 flex items-center justify-between">
						<span class="text-sm font-medium text-stone-300">Variant {variant.index + 1}</span>
						<span class="text-xs text-stone-500">{variant.seed.slice(0, 8)}...</span>
					</div>
					<div class="flex flex-wrap gap-2">
						<button
							on:click={() => playStop(variant)}
							class="rounded-lg px-3 py-1 text-xs font-medium transition-all {playingVariantId ===
							variant.id
								? 'bg-amber-700 text-white'
								: 'bg-stone-800 text-stone-300 hover:bg-stone-700'}"
						>
							{playingVariantId === variant.id ? '⏹ Stop' : '▶ Play'}
						</button>
						<button
							on:click={() => exportVariantWav(variant)}
							class="rounded-lg bg-stone-800 px-3 py-1 text-xs text-stone-400 transition-all hover:bg-stone-700"
						>
							WAV
						</button>
						<button
							on:click={() => promoteToMain(variant)}
							class="rounded-lg bg-stone-800 px-3 py-1 text-xs text-violet-400 transition-all hover:bg-stone-700"
							title="Use this variant's seed as the main sketch seed"
						>
							⬆ Promote
						</button>
					</div>
				</div>
			{/each}
		</div>

		{#if variants.length === 0}
			<div class="rounded-xl border border-dashed border-stone-800 p-12 text-center">
				<p class="text-stone-500">No variants generated yet.</p>
				<p class="mt-2 text-sm text-stone-600">
					Generate variants to explore different sounds from the same parameters.
				</p>
			</div>
		{/if}
	</div>
{/if}
