<script lang="ts">
	import { onMount } from 'svelte';
	import { exports, loadExports } from '$lib/stores/project';
	import type { LiraExport } from '$lib/types';

	onMount(() => {
		loadExports();
	});

	function formatDate(iso: string): string {
		try {
			return new Date(iso).toLocaleString();
		} catch {
			return iso;
		}
	}
</script>

<div class="mx-auto max-w-4xl p-6">
	<div class="mb-8">
		<h1 class="m-0 text-2xl font-bold text-stone-100">Exports</h1>
		<p class="mt-1 text-sm text-stone-500">History of all exported sounds and sketches.</p>
	</div>

	<div class="overflow-hidden rounded-xl border border-stone-800">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b border-stone-800 bg-stone-900 text-left text-xs text-stone-500">
					<th class="px-4 py-3 font-medium">Filename</th>
					<th class="px-4 py-3 font-medium">Format</th>
					<th class="px-4 py-3 font-medium">Duration</th>
					<th class="px-4 py-3 font-medium">Sketch</th>
					<th class="px-4 py-3 font-medium">Created</th>
				</tr>
			</thead>
			<tbody>
				{#each $exports as exp (exp.id)}
					<tr class="border-b border-stone-800/50 hover:bg-stone-900/50">
						<td class="px-4 py-3 font-medium text-stone-300">{exp.filename}</td>
						<td class="px-4 py-3">
							<span
								class="rounded-full px-2 py-0.5 text-xs uppercase {exp.format === 'wav'
									? 'bg-emerald-900 text-emerald-300'
									: exp.format === 'json'
										? 'bg-blue-900 text-blue-300'
										: 'bg-violet-900 text-violet-300'}"
							>
								{exp.format}
							</span>
						</td>
						<td class="px-4 py-3 text-stone-500">{exp.durationSeconds}s</td>
						<td class="px-4 py-3 text-stone-500">{exp.sketchId}</td>
						<td class="px-4 py-3 text-stone-500">{formatDate(exp.createdAt)}</td>
					</tr>
				{/each}
			</tbody>
		</table>

		{#if $exports.length === 0}
			<div class="p-12 text-center">
				<p class="text-stone-500">No exports yet.</p>
				<p class="mt-2 text-sm text-stone-600">Export a sound sketch to see it here.</p>
			</div>
		{/if}
	</div>
</div>
