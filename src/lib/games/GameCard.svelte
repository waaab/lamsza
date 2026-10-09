<script>
	import GameIcon from '$lib/games/GameIcon.svelte';
	import { gameBySlug } from '$lib/games/catalog.js';

	/**
	 * A game's card, the same everywhere (UI_BASELINE "game-look"): the
	 * game's colour and icon, its name and line. Shared from lamsza to
	 * Játszótér and Szótár, so Lámsza's home, Szótár's home and Játszótér's
	 * lists draw one card. On hover or focus the icon hops and grows and a
	 * round play button appears; the card itself stays put (owner,
	 * 2026-10-09). Still when the visitor prefers reduced motion.
	 *
	 * `href`: where it leads; Játszótér's own game page by default, the
	 * caller passes Játszótér's full address from the other apps.
	 * `featured`: Játszótér's home page's first game, a larger icon (the page
	 * makes it two columns wide).
	 *
	 * @type {{ slug: string, title_hu: string, description_hu: string, featured?: boolean, href?: string }}
	 */
	let { slug, title_hu, description_hu, featured = false, href = '' } = $props();

	const color = $derived(gameBySlug(slug)?.color ?? '#3d7a6f');
	const link = $derived(href || `/jatszok/${slug}`);
</script>

<a href={link} class="game-poster" class:game-poster--featured={featured} style:--game-color={color}>
	<span class="game-poster-art">
		<span class="game-poster-icon"><GameIcon {slug} size={featured ? 96 : 72} /></span>
		<span class="game-poster-play" aria-hidden="true">
			<svg viewBox="0 0 24 24" width="16" height="16"><path d="M8 5.5v13l10.5-6.5z" fill="currentColor" /></svg>
		</span>
	</span>
	<span class="game-poster-name">{title_hu}</span>
	<span class="game-poster-notes">{description_hu}</span>
</a>

<style>
	.game-poster {
		display: flex;
		flex-direction: column;
		color: inherit;
		text-decoration: none;
	}

	.game-poster-art {
		position: relative;
		display: grid;
		place-items: center;
		height: 9.5rem;
		border-radius: 1.1rem;
		background: var(--game-color);
		color: #fff;
	}

	.game-poster-icon {
		display: grid;
		place-items: center;
		line-height: 0;
		transition: transform 0.25s ease;
	}

	/* Hover or focus: the icon hops, settles a little larger, and a round
	   play button appears. The card does not move. */
	.game-poster:hover .game-poster-icon,
	.game-poster:focus-visible .game-poster-icon {
		animation: game-icon-hop 0.55s cubic-bezier(0.3, 0.7, 0.4, 1) forwards;
	}

	@keyframes game-icon-hop {
		0% {
			transform: translateY(0) scale(1) rotate(0deg);
		}
		40% {
			transform: translateY(-8px) scale(1.14) rotate(-6deg);
		}
		70% {
			transform: translateY(0) scale(1.08) rotate(3deg);
		}
		100% {
			transform: translateY(0) scale(1.1) rotate(0deg);
		}
	}

	.game-poster-play {
		position: absolute;
		right: 0.75rem;
		bottom: 0.75rem;
		display: grid;
		place-items: center;
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.92);
		color: #222;
		opacity: 0;
		transform: translateY(8px);
		transition: opacity 0.2s ease, transform 0.2s ease;
	}

	.game-poster:hover .game-poster-play,
	.game-poster:focus-visible .game-poster-play {
		opacity: 1;
		transform: none;
	}

	.game-poster-name {
		margin-top: 0.75rem;
		font-size: var(--text-lg);
		font-weight: 700;
		text-align: center;
	}

	.game-poster-notes {
		margin-top: 0.25rem;
		color: var(--text-secondary);
		font-size: var(--text-sm);
		text-align: center;
	}

	.game-poster:hover .game-poster-name {
		color: var(--game-color);
	}

	@media (prefers-reduced-motion: reduce) {
		.game-poster:hover .game-poster-icon,
		.game-poster:focus-visible .game-poster-icon {
			animation: none;
		}

		.game-poster-play {
			transition: none;
		}
	}
</style>
