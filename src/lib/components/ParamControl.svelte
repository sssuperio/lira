<script lang="ts">
	export let label: string;
	export let value: number | string;
	export let min: number | undefined = undefined;
	export let max: number | undefined = undefined;
	export let step: number | string = 1;
	export let type: 'slider' | 'select' | 'text' = 'slider';
	export let options: { value: string; label: string }[] = [];
	export let onChange: (val: number | string) => void = () => {};

	function handleInput(e: Event) {
		const target = e.target as HTMLInputElement | HTMLSelectElement;
		if (type === 'slider') {
			onChange(Number(target.value));
		} else {
			onChange(target.value);
		}
	}
</script>

<div class="flex flex-col gap-1.5">
	<div class="flex items-baseline justify-between">
		<label class="text-xs font-medium text-stone-400">{label}</label>
		<span class="text-xs text-stone-500">
			{#if type === 'slider'}
				{value}
			{:else if type === 'select'}
				{options.find((o) => o.value === value)?.label ?? value}
			{:else}
				{value}
			{/if}
		</span>
	</div>

	{#if type === 'slider' && typeof min === 'number' && typeof max === 'number'}
		<input
			type="range"
			{min}
			{max}
			{step}
			value={typeof value === 'number' ? value : 0}
			on:input={handleInput}
			class="custom-slider w-full"
		/>
	{:else if type === 'select'}
		<select
			value={value}
			on:change={handleInput}
			class="w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-1.5 text-sm text-stone-200"
		>
			{#each options as opt}
				<option value={opt.value}>{opt.label}</option>
			{/each}
		</select>
	{:else}
		<input
			type="text"
			value={value}
			on:input={handleInput}
			class="w-full rounded-lg border border-stone-700 bg-stone-800 px-3 py-1.5 text-sm text-stone-200"
		/>
	{/if}
</div>

<style>
	.custom-slider {
		-webkit-appearance: none;
		appearance: none;
		height: 6px;
		background: #44403c;
		border-radius: 3px;
		outline: none;
		cursor: pointer;
	}
	.custom-slider::-webkit-slider-thumb {
		-webkit-appearance: none;
		appearance: none;
		width: 16px;
		height: 16px;
		border-radius: 50%;
		background: #a78bfa;
		cursor: pointer;
		border: none;
	}
	.custom-slider::-moz-range-thumb {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		background: #a78bfa;
		cursor: pointer;
		border: none;
	}
</style>
