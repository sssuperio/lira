<script lang="ts">
	import type { LiraMood, LiraSketch } from '$lib/types';
	import { MATERIAL_LABELS, MOOD_COLORS, MOOD_LABELS, SHAPE_LABELS } from '$lib/types';

	export let sketch: LiraSketch;
	export let onPlay: () => void = () => {};
	export let onEdit: () => void = () => {};
	export let onDuplicate: () => void = () => {};
	export let isPlaying = false;

	$: colors = MOOD_COLORS[sketch.mood as LiraMood];
</script>

<div
	class="group rounded-xl border p-5 transition-all hover:bg-stone-900/80"
	style="border-color: {colors.border}30; box-shadow: 0 0 30px {colors.glow}00"
	onmouseenter="this.style.boxShadow='0 0 30px {colors.glow}'; this.style.borderColor='{colors.border}60'"
	onmouseleave="this.style.boxShadow='0 0 30px {colors.glow}00'; this.style.borderColor='{colors.border}30'"
>
	<div class="mb-3 flex items-start justify-between">
		<h3 class="m-0 text-base font-semibold" style="color: {colors.text}">{sketch.name}</h3>
		<span class="shrink-0 rounded-full px-2 py-0.5 text-xs" style="background: {colors.bg}; color: {colors.text}">{sketch.durationSeconds}s</span>
	</div>

	<div class="mb-4 flex flex-wrap gap-1.5 text-xs">
		<span class="rounded px-2 py-0.5 capitalize" style="background: {colors.bg}; color: {colors.text}">{MOOD_LABELS[sketch.mood]}</span>
		<span class="rounded bg-stone-800 px-2 py-0.5 text-stone-400">{MATERIAL_LABELS[sketch.material]}</span>
		<span class="rounded bg-stone-800 px-2 py-0.5 text-stone-400">{SHAPE_LABELS[sketch.shape]}</span>
	</div>

	<div class="flex gap-2">
		<button
			on:click={onPlay}
			class="flex items-center gap-1.5 rounded-lg px-4 py-1.5 text-sm font-medium transition-all {isPlaying
				? 'text-white'
				: 'text-stone-300 hover:text-white'}"
			style="background: {isPlaying ? colors.accent : '#292524'}"
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
