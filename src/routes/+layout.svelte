<script lang="ts">
	import '../app.postcss';
	import '$lib/app.css';
	import { onMount } from 'svelte';
	import { loadProject, loadSketches, loadExports } from '$lib/stores/project';

	onMount(() => {
		console.log(
			'%c🎵 lira %csound sketchbook %c🖊️',
			'font-size:1.5em',
			'font-size:1em;color:#a78bfa',
			'font-size:1.5em'
		);
		console.log('%cgenerate → listen → adjust → export', 'color:#94a3b8');
		console.log('%cLira is not a DAW. It is a sound sketchbook.', 'color:#94a3b8');

		loadProject();
		loadSketches();
		loadExports();
	});

	const links = [
		{ href: '/sketches', text: 'Gallery' },
		{ href: '/exports', text: 'Exports' },
		{ href: '/settings', text: 'Settings' }
	];
</script>

<div class="flex h-screen w-screen flex-col overflow-hidden bg-stone-950">
	<nav class="flex shrink-0 flex-row items-center gap-1 space-x-1 bg-stone-900 px-4 py-2">
		<a href="/sketches" class="mr-3 font-bold text-violet-400 text-lg no-underline"> lira </a>
		{#each links as link}
			<a
				href={link.href}
				class="rounded-lg px-3 py-1.5 text-sm text-stone-400 no-underline transition-colors hover:bg-stone-800 hover:text-stone-200"
			>
				{link.text}
			</a>
		{/each}
		<div class="ml-auto text-xs text-stone-600">sound sketchbook</div>
	</nav>

	<main class="h-0 grow overflow-y-auto">
		<slot />
	</main>
</div>
