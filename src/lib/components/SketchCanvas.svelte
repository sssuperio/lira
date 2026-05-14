<script lang="ts">
	import type { LiraMood, LiraShape } from '$lib/types';
	import { MOOD_COLORS, SHAPES } from '$lib/types';

	export let mood: LiraMood = 'magic';
	export let onApply: (params: {
		shape: LiraShape;
		density: number;
		brightness: number;
		softness: number;
		movement: number;
		pictureSeed: string;
	}) => void = () => {};

	$: colors = MOOD_COLORS[mood];

	type Point = { x: number; y: number; t: number };
	type Stroke = { points: Point[] };

	let strokes: Stroke[] = [];
	let currentStroke: Point[] = [];
	let isDrawing = false;
	let canvasWidth = 400;
	let canvasHeight = 280;
	let detectedShape: LiraShape = 'rise';
	let detectedParams = { density: 5, brightness: 7, softness: 6, movement: 5 };
	let pictureSeed = '';

	function toLocal(e: PointerEvent): Point {
		return {
			x: e.offsetX,
			y: e.offsetY,
			t: Date.now()
		};
	}

	function onPointerDown(e: PointerEvent) {
		(e.currentTarget as Element).setPointerCapture(e.pointerId);
		currentStroke = [];
		isDrawing = true;
		currentStroke = [toLocal(e)];
	}

	function onPointerMove(e: PointerEvent) {
		if (!isDrawing) return;
		currentStroke = [...currentStroke, toLocal(e)];
	}

	function onPointerUp(_e: PointerEvent) {
		if (!isDrawing) return;
		isDrawing = false;
		if (currentStroke.length > 1) {
			strokes = [...strokes, { points: currentStroke }];
			analyzeStrokes();
		}
		currentStroke = [];
	}

	class ClassNameFix {
		random() {
			return Math.random();
		}
	}
	const _fix = new ClassNameFix();

	function analyzeStrokes() {
		const allPoints = strokes.flatMap((s) => s.points);
		if (allPoints.length < 5) {
			const allPts = [...allPoints, ...currentStroke];
			if (allPts.length < 5) return;
			doAnalyze(allPts);
			return;
		}
		doAnalyze(allPoints);
	}

	function doAnalyze(pts: Point[]) {
		const xs = pts.map((p) => p.x);
		const ys = pts.map((p) => p.y);
		const minX = Math.min(...xs);
		const maxX = Math.max(...xs);
		const minY = Math.min(...ys);
		const maxY = Math.max(...ys);
		const rangeX = maxX - minX || 1;
		const rangeY = maxY - minY || 1;

		// Normalize
		const norm = pts.map((p) => ({
			x: (p.x - minX) / rangeX,
			y: (p.y - minY) / rangeY
		}));

		// Compute shape by analyzing contour
		detectedShape = classifyShape(norm, rangeY / rangeX);

		// Density: number of direction changes normalized to 1-10
		let turns = 0;
		for (let i = 2; i < norm.length; i++) {
			const dy1 = norm[i - 1].y - norm[i - 2].y;
			const dy2 = norm[i].y - norm[i - 1].y;
			if (Math.sign(dy1) !== Math.sign(dy2) && Math.abs(dy2) > 0.02) turns++;
		}
		const density = Math.max(1, Math.min(10, Math.round(strokes.length * 2 + turns * 0.7)));

		// Brightness: average y (inverted — higher strokes = brighter)
		const avgY = norm.reduce((s, p) => s + p.y, 0) / norm.length;
		const brightness = Math.max(1, Math.min(10, Math.round((1 - avgY) * 8 + 2)));

		// Movement: horizontal variance
		const meanX = norm.reduce((s, p) => s + p.x, 0) / norm.length;
		const varX = norm.reduce((s, p) => s + (p.x - meanX) ** 2, 0) / norm.length;
		const movement = Math.max(1, Math.min(10, Math.round(varX * 30 + 3)));

		// Softness: smoothness of curves (low angular change = soft)
		let totalAngle = 0;
		for (let i = 2; i < norm.length; i++) {
			const dx1 = norm[i - 1].x - norm[i - 2].x;
			const dy1 = norm[i - 1].y - norm[i - 2].y;
			const dx2 = norm[i].x - norm[i - 1].x;
			const dy2 = norm[i].y - norm[i - 1].y;
			const dot = dx1 * dx2 + dy1 * dy2;
			const mag = Math.sqrt(dx1 * dx1 + dy1 * dy1) * Math.sqrt(dx2 * dx2 + dy2 * dy2);
			if (mag > 0.0001) totalAngle += Math.acos(Math.max(-1, Math.min(1, dot / mag)));
		}
		const avgAngle = norm.length > 2 ? totalAngle / (norm.length - 2) : 0;
		const softness = Math.max(1, Math.min(10, Math.round(10 - avgAngle * 8)));

		detectedParams = { density, brightness, softness, movement };

		// Hash the drawing into a seed
		const sample = pts.filter((_, i) => i % 3 === 0);
		const hashStr = sample
			.map((p) => `${Math.round(p.x / 3)},${Math.round(p.y / 3)}`)
			.join('|');
		pictureSeed = hashString(hashStr).slice(0, 12);
	}

	function hashString(str: string): string {
		let h = 0;
		for (let i = 0; i < str.length; i++) {
			h = ((h << 5) - h + str.charCodeAt(i)) | 0;
		}
		return Math.abs(h).toString(36);
	}

	function classifyShape(norm: { x: number; y: number }[], aspectRatio: number): LiraShape {
		const mid = Math.floor(norm.length / 2);
		const firstHalf = norm.slice(0, mid);
		const secondHalf = norm.slice(mid);

		const avgY1 = firstHalf.reduce((s, p) => s + p.y, 0) / firstHalf.length;
		const avgY2 = secondHalf.reduce((s, p) => s + p.y, 0) / secondHalf.length;
		const trend = avgY2 - avgY1;

		// Count oscillations
		let crossings = 0;
		let prevSign = 0;
		for (let i = 1; i < norm.length; i++) {
			const dy = norm[i].y - norm[i - 1].y;
			const sign = Math.sign(dy);
			if (sign !== 0 && sign !== prevSign && prevSign !== 0) crossings++;
			if (sign !== 0) prevSign = sign;
		}

		// Max y range
		const ys = norm.map((p) => p.y);
		const yRange = Math.max(...ys) - Math.min(...ys);

		if (crossings > 6) return 'sparkle';
		if (crossings > 3 && yRange > 0.4) return 'bounce';
		if (crossings > 3) return 'wave';
		if (yRange < 0.2) return 'pulse';
		if (trend < -0.2) return 'fall';
		if (trend > 0.2) return 'rise';
		if (aspectRatio < 0.6) return 'orbit';
		if (yRange > 0.7) return 'swell';
		if (norm.length < 8) return 'pop';
		return 'rise';
	}

	function strokePath(points: Point[]): string {
		if (points.length === 0) return '';
		const scaled = points.map((p) => `${p.x},${p.y}`);
		return `M ${scaled[0]} ${scaled.slice(1).map((s) => `L ${s}`).join(' ')}`;
	}

	function applyDrawing() {
		pictureSeed = pictureSeed || hashString(Date.now().toString()).slice(0, 12);
		onApply({
			shape: detectedShape,
			...detectedParams,
			pictureSeed
		});
	}

	function clearCanvas() {
		strokes = [];
		currentStroke = [];
		detectedParams = { density: 5, brightness: 7, softness: 6, movement: 5 };
		detectedShape = 'rise';
		pictureSeed = '';
	}
</script>

<div class="flex flex-col gap-3">
	<div class="flex items-center justify-between">
		<span class="text-xs font-medium text-stone-400">Sound sketch</span>
		<button
			on:click={clearCanvas}
			class="rounded-lg px-2 py-0.5 text-xs text-stone-500 transition-colors hover:bg-stone-800 hover:text-stone-300"
		>
			Clear
		</button>
	</div>

	<div class="relative inline-block overflow-hidden rounded-xl border" style="border-color: {colors.border}40; line-height: 0;">
		<svg
			width={canvasWidth}
			height={canvasHeight}
			style="background: #1c1917; display: block;"
		>
			<!-- Grid -->
			{#each Array(12) as _, i}
				<line
					x1={0}
					y1={(i / 12) * canvasHeight}
					x2={canvasWidth}
					y2={(i / 12) * canvasHeight}
					stroke="#292524"
					stroke-width="0.5"
				/>
			{/each}
			{#each Array(16) as _, i}
				<line
					x1={(i / 16) * canvasWidth}
					y1={0}
					x2={(i / 16) * canvasWidth}
					y2={canvasHeight}
					stroke="#292524"
					stroke-width="0.5"
				/>
			{/each}

			<!-- Past strokes -->
			{#each strokes as stroke}
				<path
					d={strokePath(stroke.points)}
					fill="none"
					stroke={colors.accent}
					stroke-width="3"
					stroke-linecap="round"
					stroke-linejoin="round"
					style="filter: drop-shadow(0 0 6px {colors.glow})"
				/>
			{/each}

			<!-- Current stroke -->
			{#if currentStroke.length > 1}
				<path
					d={strokePath(currentStroke)}
					fill="none"
					stroke={colors.accent}
					stroke-width="3"
					stroke-linecap="round"
					stroke-linejoin="round"
					opacity="0.8"
					style="filter: drop-shadow(0 0 8px {colors.glow})"
				/>
			{/if}

			<!-- Placeholder text -->
			{#if strokes.length === 0 && currentStroke.length === 0}
				<text
					x={canvasWidth / 2}
					y={canvasHeight / 2}
					text-anchor="middle"
					fill="#44403c"
					font-size="14"
				>
					Draw a sound shape here
				</text>
			{/if}
		</svg>

		<!-- Transparent drawing surface overlaid on top of the SVG -->
		<div
			class="absolute inset-0 cursor-crosshair"
			style="touch-action: none;"
			role="img"
			aria-label="Drawing canvas - sketch a sound shape"
			on:pointerdown={onPointerDown}
			on:pointermove={onPointerMove}
			on:pointerup={onPointerUp}
			on:pointercancel={onPointerUp}
		/>
	</div>

	<!-- Detected params -->
	<div class="grid grid-cols-5 gap-2 text-center text-xs">
		<div class="rounded-lg px-1 py-1" style="background:{colors.bg}; color:{colors.text}">
			<span class="block text-[10px] opacity-60">shape</span>
			<span class="font-medium capitalize">{detectedShape}</span>
		</div>
		<div class="rounded-lg bg-stone-800 px-1 py-1 text-stone-400">
			<span class="block text-[10px] opacity-60">density</span>
			<span class="font-medium text-stone-200">{detectedParams.density}</span>
		</div>
		<div class="rounded-lg bg-stone-800 px-1 py-1 text-stone-400">
			<span class="block text-[10px] opacity-60">bright</span>
			<span class="font-medium text-stone-200">{detectedParams.brightness}</span>
		</div>
		<div class="rounded-lg bg-stone-800 px-1 py-1 text-stone-400">
			<span class="block text-[10px] opacity-60">soft</span>
			<span class="font-medium text-stone-200">{detectedParams.softness}</span>
		</div>
		<div class="rounded-lg bg-stone-800 px-1 py-1 text-stone-400">
			<span class="block text-[10px] opacity-60">move</span>
			<span class="font-medium text-stone-200">{detectedParams.movement}</span>
		</div>
	</div>

	<button
		on:click={applyDrawing}
		disabled={strokes.length === 0}
		class="rounded-xl py-2 text-sm font-semibold text-white transition-all disabled:opacity-30"
		style="background: {strokes.length > 0 ? colors.accent : '#44403c'}"
	>
		Apply drawing to sketch
	</button>
</div>
