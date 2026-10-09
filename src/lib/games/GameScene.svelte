<!--
	A game's animated scene on its card and its starter (UI_BASELINE "game-look"; owner,
	2026-10-09, after words.com's game thumbnails): a small looping scene in
	white on the game's colour that acts out how the game is played. Tiles
	flip, honeycomb cells pop, a word is dragged out of a letter grid, letters drop into
	their blanks, rovás signs get their reading, an arrow-word fills in. Only
	the scene moves; the card stays still.

	CSS motion only, one 6 s loop; each piece starts at its own delay (--d).
	When the visitor prefers reduced motion the finished scene simply stands.
	GameCard keeps it still until the card is hovered or focused; the game
	page's starter (Játszótér's GameShell) lets it loop. Shared from lamsza
	to Játszótér and Szótár. The static GameIcon is no longer drawn; it is
	kept until its drawings move to AppIcon (OPEN_ITEMS).
-->
<script>
	let { slug, size = 160 } = $props();

	const TILE = 22;
	const GAP = 6;
	/** x of tile i in a centred row of n tiles in the 160-wide scene. */
	const rowX = (/** @type {number} */ n, /** @type {number} */ i) => 80 - (n * TILE + (n - 1) * GAP) / 2 + i * (TILE + GAP);

	const LASKA = ['L', 'A', 'S', 'K', 'A'];
	const KALAN = ['K', 'A', 'L', 'Á', 'N'];

	// Kaptár: the centre cell and six around it (pointy-top hexagons).
	const HEX_R = 13;
	const hexPath = (/** @type {number} */ cx, /** @type {number} */ cy) =>
		Array.from({ length: 6 }, (_, k) => {
			const a = (Math.PI / 3) * k - Math.PI / 2;
			return `${k ? 'L' : 'M'}${(cx + HEX_R * Math.cos(a)).toFixed(2)} ${(cy + HEX_R * Math.sin(a)).toFixed(2)}`;
		}).join(' ') + 'Z';
	const HEX_STEP = HEX_R * Math.sqrt(3) + 2;
	const HIVE = ['K', 'L', 'P', 'T', 'R', 'S'].map((letter, k) => {
		const a = (Math.PI / 3) * k;
		return { letter, x: 80 + HEX_STEP * Math.cos(a), y: 50 + HEX_STEP * Math.sin(a) };
	});

	// Szókereső: a 6 × 4 letter grid of cells, as in the game; KACOR hides in the
	// second row. Its cells turn amber one by one as if dragged (the game's live
	// selection), then all turn found together: white with the game's colour
	// for the letter, the card's white-on-colour version of the game's found cells.
	const GRID = ['TÉLÓSZ', 'MKACOR', 'ÁBLEÍV', 'GYÚPÖN'];
	const CELL = 18;
	const GX = 80 - (6 * CELL) / 2;
	const GY = 50 - (4 * CELL) / 2;
	const KACOR_ROW = 1;

	// Rovásfejtő: the five signs of "lámsza" (GameIcon's outlines, 24-unit),
	// drawn 5.5x and mirrored as rovás runs right to left.
	const ROVAS = [
		'M0.52 15.95L-0.18 15.95L2.22 10.01L2.82 10.01L5.18 15.95L4.49 15.95L3.99 14.66L3.48 15.95L2.87 15.95L3.68 13.89L3.26 12.81L2.02 15.95L1.39 15.95L2.94 12.00Q2.81 11.65 2.71 11.36Q2.60 11.07 2.52 10.79L2.52 10.79L2.50 10.79Q2.41 11.07 2.29 11.41Q2.17 11.74 2.02 12.12L2.02 12.12L0.52 15.95Z',
		'M8.20 15.95L7.57 15.95L7.57 14.19L5.56 12.47L5.56 11.94L7.55 10.01L8.20 10.01L8.20 15.95ZM6.15 12.20L7.57 13.38L7.57 10.74L6.15 12.20Z',
		'M13.28 15.95L12.62 15.95L10.31 14.57L10.31 13.96L11.82 12.96L10.31 11.95L10.31 11.33L12.62 10.01L13.28 10.01L13.28 15.95ZM10.90 11.65L12.65 12.84L12.65 10.65L10.90 11.65ZM10.90 14.23L12.65 15.25L12.65 13.07L10.90 14.23Z',
		'M17.05 15.95L16.42 15.95L16.42 10.01L17.05 10.01L17.05 15.95Z',
		'M22.45 15.95L21.82 15.95L21.82 12.63L19.72 12.41L19.72 11.68L21.77 10.01L22.45 10.01L22.45 15.95ZM20.29 11.91L21.82 12.05L21.82 10.67L20.29 11.91Z'
	];
	const ROVAS_LATIN = ['L', 'Á', 'M', 'SZ', 'A'];
	const RS = 5.5;
	const ROVAS_X = (/** @type {number} */ i) => 2.5 + i * 4.75;

	// Tájszórejtvény: a 6 × 3 arrow-word grid; the clue cell's word runs right.
	const TX = 80 - (6 * CELL) / 2;
	const TY = 50 - (3 * CELL) / 2;
	const ACROSS = ['K', 'A', 'C', 'O', 'R'];
</script>

<svg
	class="game-scene"
	viewBox="0 0 160 100"
	width={size}
	height={Math.round((size * 100) / 160)}
	aria-hidden="true"
	focusable="false"
>
	{#if slug === 'szorejto'}
		{#each LASKA as letter, i (i)}
			<rect class="gs-line" x={rowX(5, i)} y="39" width={TILE} height={TILE} rx="5" />
			<g class="gs-flip" style="--d: {(i * 0.35).toFixed(2)}s">
				<rect class="gs-solid" x={rowX(5, i)} y="39" width={TILE} height={TILE} rx="5" />
				<text class="gs-letter gs-on-solid" x={rowX(5, i) + TILE / 2} y="55.5">{letter}</text>
			</g>
		{/each}
	{:else if slug === 'kaptar'}
		{#each HIVE as cell, k (k)}
			<g class="gs-pop" style="--d: {(0.25 + k * 0.22).toFixed(2)}s">
				<path class="gs-line" d={hexPath(cell.x, cell.y)} />
				<text class="gs-letter" x={cell.x} y={cell.y + 5}>{cell.letter}</text>
			</g>
		{/each}
		<g class="gs-pulse">
			<path class="gs-solid" d={hexPath(80, 50)} />
			<text class="gs-letter gs-on-solid" x="80" y="55">A</text>
		</g>
	{:else if slug === 'szokereso'}
		{#each GRID as row, r (r)}
			{#each Array.from(row) as letter, c (c)}
				{@const x = GX + c * CELL}
				{@const y = GY + r * CELL}
				<rect class="gs-cell" x={x + 1} y={y + 1} width={CELL - 2} height={CELL - 2} rx="3" />
				<text class="gs-letter gs-small" x={x + CELL / 2} y={y + CELL / 2 + 4}>{letter}</text>
				{#if r === KACOR_ROW && c >= 1}
					<g class="gs-trace" style="--d: {(0.6 + (c - 1) * 0.28).toFixed(2)}s">
						<rect class="gs-trace-fill" x={x + 1} y={y + 1} width={CELL - 2} height={CELL - 2} rx="3" />
						<text class="gs-letter gs-small gs-on-trace" x={x + CELL / 2} y={y + CELL / 2 + 4}>{letter}</text>
					</g>
					<g class="gs-hit">
						<rect class="gs-solid" x={x + 1} y={y + 1} width={CELL - 2} height={CELL - 2} rx="3" />
						<text class="gs-letter gs-small gs-on-solid" x={x + CELL / 2} y={y + CELL / 2 + 4}>{letter}</text>
					</g>
				{/if}
			{/each}
		{/each}
	{:else if slug === 'akasztofa'}
		{#each KALAN as letter, i (i)}
			<line class="gs-line" x1={rowX(5, i) + 2} y1="66" x2={rowX(5, i) + TILE - 2} y2="66" />
			<text class="gs-letter gs-big gs-drop" style="--d: {(0.3 + i * 0.4).toFixed(2)}s" x={rowX(5, i) + TILE / 2} y="60">{letter}</text>
		{/each}
	{:else if slug === 'rovasfejto'}
		<g transform="translate({80 - 12 * RS} {-6 * RS}) scale({RS})">
			{#each ROVAS as d, i (i)}
				<path class="gs-glyph" {d} transform="translate({ROVAS_X(i)} 0) scale(-1 1) translate({-ROVAS_X(i)} 0)" />
			{/each}
		</g>
		{#each ROVAS_LATIN as letter, i (i)}
			<text class="gs-letter gs-rise" style="--d: {(0.4 + i * 0.35).toFixed(2)}s" x={80 - 12 * RS + ROVAS_X(i) * RS} y="78">{letter}</text>
		{/each}
	{:else if slug === 'tajszorejtveny'}
		{#each [0, 1, 2] as r (r)}
			{#each [0, 1, 2, 3, 4, 5] as c (c)}
				{#if !(r === 0 && c === 0)}
					<rect class="gs-line gs-thin" x={TX + c * CELL} y={TY + r * CELL} width={CELL} height={CELL} rx="2" />
				{/if}
			{/each}
		{/each}
		<rect class="gs-solid" x={TX} y={TY} width={CELL} height={CELL} rx="2" />
		<path class="gs-on-solid-fill" d="M{TX + 5} {TY + 9}h6m-2.5 -3l3 3l-3 3" />
		{#each ACROSS as letter, i (i)}
			<text class="gs-letter gs-small gs-pop" style="--d: {(0.4 + i * 0.32).toFixed(2)}s" x={TX + (i + 1) * CELL + CELL / 2} y={TY + CELL / 2 + 4}>{letter}</text>
		{/each}
	{/if}
</svg>

<style>
	.game-scene {
		display: block;
		max-width: 100%;
		height: auto;
		overflow: visible;
		--gs-ink: #ffffff;
	}

	.gs-line {
		fill: none;
		stroke: var(--gs-ink);
		stroke-width: 2.5;
		stroke-linecap: round;
		stroke-linejoin: round;
		opacity: 0.85;
	}

	.gs-thin {
		stroke-width: 1.6;
		opacity: 0.6;
	}

	.gs-solid {
		fill: var(--gs-ink);
	}

	.gs-letter {
		fill: var(--gs-ink);
		font-family: var(--font-sans, system-ui, sans-serif);
		font-size: 14px;
		font-weight: 800;
		text-anchor: middle;
	}

	.gs-small {
		font-size: 11px;
	}

	.gs-big {
		font-size: 18px;
	}

	.gs-on-solid {
		fill: var(--game-color, #333);
	}

	.gs-on-solid-fill {
		fill: none;
		stroke: var(--game-color, #333);
		stroke-width: 1.8;
		stroke-linecap: round;
		stroke-linejoin: round;
	}

	/* Szókereső's grid cells, its amber live selection (the game's --warm-light
	   with dark letters) and its found cells (gs-solid). */
	.gs-cell {
		fill: none;
		stroke: var(--gs-ink);
		stroke-width: 1;
		opacity: 0.45;
	}

	.gs-trace-fill {
		fill: var(--warm-light, #f2b44f);
	}

	.gs-on-trace {
		fill: #1f1f1f;
	}

	.gs-glyph {
		fill: var(--gs-ink);
	}

	/* Motion: one 6 s loop; each piece waits for its --d. At rest (and under
	   reduced motion) the scene shows its finished state. */
	.gs-flip,
	.gs-pop,
	.gs-pulse,
	.gs-trace,
	.gs-hit,
	.gs-drop,
	.gs-rise {
		transform-box: fill-box;
		transform-origin: center;
	}

	@media (prefers-reduced-motion: no-preference) {
		.gs-flip {
			animation: gs-flip 6s ease-in-out var(--d, 0s) infinite both;
		}
		.gs-pop {
			animation: gs-pop 6s cubic-bezier(0.3, 1.4, 0.5, 1) var(--d, 0s) infinite both;
		}
		.gs-pulse {
			animation: gs-pulse 3s ease-in-out infinite;
		}
		.gs-trace {
			animation: gs-trace 6s ease-out var(--d, 0s) infinite both;
		}
		.gs-hit {
			animation: gs-hit 6s ease-out infinite both;
		}
		.gs-drop {
			animation: gs-drop 6s cubic-bezier(0.3, 1.3, 0.5, 1) var(--d, 0s) infinite both;
		}
		.gs-rise {
			animation: gs-rise 6s ease-out var(--d, 0s) infinite both;
		}
	}

	@keyframes gs-flip {
		0%,
		4% {
			transform: scaleY(0);
		}
		10%,
		78% {
			transform: scaleY(1);
		}
		84%,
		100% {
			transform: scaleY(0);
		}
	}

	@keyframes gs-pop {
		0% {
			transform: scale(0);
			opacity: 0;
		}
		8%,
		78% {
			transform: scale(1);
			opacity: 1;
		}
		86%,
		100% {
			transform: scale(0);
			opacity: 0;
		}
	}

	@keyframes gs-pulse {
		0%,
		100% {
			transform: scale(1);
		}
		50% {
			transform: scale(1.07);
		}
	}

	/* A KACOR cell turns amber at its --d (0.6 s to 1.72 s) and stays amber for
	   2.5 s, under the found cell from 2.2 s; gone before the found cells go. */
	@keyframes gs-trace {
		0% {
			transform: scale(0.7);
			opacity: 0;
		}
		3%,
		42% {
			transform: scale(1);
			opacity: 1;
		}
		43%,
		100% {
			transform: scale(1);
			opacity: 0;
		}
	}

	/* The drag ends at 2.1 s: the five cells turn found together, a small pop,
	   until 5.2 s; then the grid stands plain until the next loop. */
	@keyframes gs-hit {
		0%,
		35% {
			transform: scale(1);
			opacity: 0;
		}
		38% {
			transform: scale(1.12);
			opacity: 1;
		}
		42%,
		86% {
			transform: scale(1);
			opacity: 1;
		}
		91%,
		100% {
			transform: scale(1);
			opacity: 0;
		}
	}

	@keyframes gs-drop {
		0% {
			transform: translateY(-34px);
			opacity: 0;
		}
		8%,
		78% {
			transform: translateY(0);
			opacity: 1;
		}
		86%,
		100% {
			transform: translateY(0);
			opacity: 0;
		}
	}

	@keyframes gs-rise {
		0% {
			transform: translateY(8px);
			opacity: 0;
		}
		8%,
		78% {
			transform: translateY(0);
			opacity: 1;
		}
		86%,
		100% {
			transform: translateY(0);
			opacity: 0;
		}
	}
</style>
