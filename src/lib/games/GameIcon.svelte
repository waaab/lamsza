<script>
	import AppIcon from '$lib/icons/AppIcon.svelte';

	let { slug, size = 24 } = $props();

	/**
	 * Rovás for Lámsza, one sign each: l, á, m, sz, a (U+10C96, U+10C81, U+10C98,
	 * U+10CA5, U+10C80), the same mapping as the game. The outlines are Noto Sans
	 * Old Hungarian's (SIL OFL), traced at 7px so the icon needs no font: Lámsza
	 * loads no rovás font. Each sign is mirrored around its centre, as rovás runs
	 * right to left.
	 */
	const lamsza = [
		'M0.52 15.95L-0.18 15.95L2.22 10.01L2.82 10.01L5.18 15.95L4.49 15.95L3.99 14.66L3.48 15.95L2.87 15.95L3.68 13.89L3.26 12.81L2.02 15.95L1.39 15.95L2.94 12.00Q2.81 11.65 2.71 11.36Q2.60 11.07 2.52 10.79L2.52 10.79L2.50 10.79Q2.41 11.07 2.29 11.41Q2.17 11.74 2.02 12.12L2.02 12.12L0.52 15.95Z',
		'M8.20 15.95L7.57 15.95L7.57 14.19L5.56 12.47L5.56 11.94L7.55 10.01L8.20 10.01L8.20 15.95ZM6.15 12.20L7.57 13.38L7.57 10.74L6.15 12.20Z',
		'M13.28 15.95L12.62 15.95L10.31 14.57L10.31 13.96L11.82 12.96L10.31 11.95L10.31 11.33L12.62 10.01L13.28 10.01L13.28 15.95ZM10.90 11.65L12.65 12.84L12.65 10.65L10.90 11.65ZM10.90 14.23L12.65 15.25L12.65 13.07L10.90 14.23Z',
		'M17.05 15.95L16.42 15.95L16.42 10.01L17.05 10.01L17.05 15.95Z',
		'M22.45 15.95L21.82 15.95L21.82 12.63L19.72 12.41L19.72 11.68L21.77 10.01L22.45 10.01L22.45 15.95ZM20.29 11.91L21.82 12.05L21.82 10.67L20.29 11.91Z'
	];
</script>

{#if slug === 'tajszorejtveny'}
	<!-- From the shared icon set (UI_BASELINE ic-system); the other tiles move there later (OPEN_ITEMS). -->
	<AppIcon name="tajszorejtveny" {size} class="game-icon" />
{:else}
<svg
	class="game-icon"
	width={size}
	height={size}
	viewBox="0 0 24 24"
	fill="currentColor"
	aria-hidden="true"
>
	{#if slug === 'rovasfejto'}
		{#each lamsza as d, i (i)}
			{@const x = 2.5 + i * 4.75}
			<path {d} transform="translate({x} 0) scale(-1 1) translate({-x} 0)" />
		{/each}
	{:else if slug === 'kaptar'}
		<g opacity="0.5">
			<path d="M15.73 1.39 19.33 3.46 19.33 7.62 15.73 9.7 12.13 7.62 12.13 3.46z" />
			<path d="M19.46 7.84 23.06 9.92 23.06 14.08 19.46 16.16 15.86 14.08 15.86 9.92z" />
			<path d="M15.73 14.3 19.33 16.38 19.33 20.54 15.73 22.61 12.13 20.54 12.13 16.38z" />
			<path d="M8.27 14.3 11.87 16.38 11.87 20.54 8.27 22.61 4.67 20.54 4.67 16.38z" />
			<path d="M4.54 7.84 8.14 9.92 8.14 14.08 4.54 16.16 0.94 14.08 0.94 9.92z" />
			<path d="M8.27 1.39 11.87 3.46 11.87 7.62 8.27 9.7 4.67 7.62 4.67 3.46z" />
		</g>
		<path d="M12 7.84 15.6 9.92 15.6 14.08 12 16.16 8.4 14.08 8.4 9.92z" />
	{:else if slug === 'szorejto'}
		<g opacity="0.4">
			<rect x="2" y="9.5" width="5" height="5" rx="1.1" />
			<rect x="9.5" y="9.5" width="5" height="5" rx="1.1" />
			<rect x="17" y="9.5" width="5" height="5" rx="1.1" />
			<rect x="2" y="17" width="5" height="5" rx="1.1" />
			<rect x="9.5" y="17" width="5" height="5" rx="1.1" />
			<rect x="17" y="17" width="5" height="5" rx="1.1" />
		</g>
		<rect x="2" y="2" width="5" height="5" rx="1.1" />
		<rect x="9.5" y="2" width="5" height="5" rx="1.1" />
		<rect x="17" y="2" width="5" height="5" rx="1.1" />
	{:else if slug === 'szokereso'}
		<g opacity="0.4">
			<rect x="2" y="2" width="5" height="5" rx="1.1" />
			<rect x="9.5" y="2" width="5" height="5" rx="1.1" />
			<rect x="17" y="2" width="5" height="5" rx="1.1" />
			<rect x="2" y="9.5" width="5" height="5" rx="1.1" />
			<rect x="17" y="9.5" width="5" height="5" rx="1.1" />
			<rect x="2" y="17" width="5" height="5" rx="1.1" />
			<rect x="9.5" y="17" width="5" height="5" rx="1.1" />
			<rect x="17" y="17" width="5" height="5" rx="1.1" />
		</g>
		<rect x="9.5" y="9.5" width="5" height="5" rx="1.1" />
		<rect x="2" y="2" width="5" height="5" rx="1.1" />
		<rect x="17" y="17" width="5" height="5" rx="1.1" />
	{:else if slug === 'akasztofa'}
		<path d="M4 3h12.2v1.7H5.7V21H4z" />
		<path d="M13.4 4.7h1.7V9h-1.7z" />
		<circle cx="14.25" cy="11.3" r="2.15" />
	{/if}
</svg>
{/if}
