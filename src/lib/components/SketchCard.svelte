<script lang="ts">
	import type { LiraSketch } from '$lib/types';
	import { MATERIAL_LABELS, MOOD_LABELS, SHAPE_LABELS } from '$lib/types';

	export let sketch: LiraSketch;
	export let onPlay: () => void = () => {};
	export let onEdit: () => void = () => {};
	export let onDuplicate: () => void = () => {};
	export let isPlaying = false;
</script>

<div
	class="group rounded-xl border border-stone-800 bg-stone-900 p-5 transition-all hover:border-violet-700/50 hover:bg-stone-900/80"
>
	<div class="mb-3 flex items-start justify-between">
		<h3 class="m-0 text-base font-semibold text-stone-100">{sketch.name}</h3>
		<span class="shrink-0 rounded-full bg-stone-800 px-2 py-0.5 text-xs text-stone-400"
			>{sketch.durationSeconds}s</span
		>
	</div>

	<div class="mb-4 flex flex-wrap gap-1.5 text-xs text-stone-400">
		<span class="rounded bg-stone-800 px-2 py-0.5 capitalize">{MOOD_LABELS[sketch.mood]}</span>
		<span class="rounded bg-stone-800 px-2 py-0.5">{MATERIAL_LABELS[sketch.material]}</span>
		<span class="rounded bg-stone-800 px-2 py-0.5">{SHAPE_LABELS[sketch.shape]}</span>
	</div>

	<div class="flex gap-2">
		<button
			on:click={onPlay}
			class="flex items-center gap-1.5 rounded-lg px-4 py-1.5 text-sm font-medium transition-all {isPlaying
				? 'bg-violet-700 text-white'
				: 'bg-stone-800 text-stone-300 hover:bg-violet-800 hover:text-white'}"
		>
			{isPlaying ? '⏹ Stop' : '▶ Play'}
		</button>
		<button
			on:click={onEdit}
			class="rounded-lg bg-stone-800 px-3 py-1.5 text-sm text-stone-400 transition-all hover:bg-stone-700 hover:text-stone-200"
		>
			Edit
		</button>
		<button
			on:click={onDuplicate}
			class="rounded-lg bg-stone-800 px-3 py-1.5 text-sm text-stone-400 transition-all hover:bg-stone-700 hover:text-stone-200"
		>
			Clone
		</button>
	</div>
</div>
