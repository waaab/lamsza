<script>
	import GameScene from '$lib/games/GameScene.svelte';
	import StartButton from '$lib/games/StartButton.svelte';
	import { gameBySlug } from '$lib/games/catalog.js';

	/**
	 * A game's card, the same everywhere (UI_BASELINE "game-look"): the
	 * game's colour and icon, its name and line. Shared from lamsza to
	 * Játszótér and Szótár, so Lámsza's home, Szótár's home and Játszótér's
	 * lists draw one card. Its art is the game's animated scene (GameScene,
	 * after words.com's thumbnails; owner, 2026-10-09): at rest it shows the
	 * finished scene, and only on hover or keyboard focus does it play, from
	 * the start (owner, 2026-10-09). Only the scene moves, the card never
	 * does; the start button's round, text-less version fades in
	 * (StartButton, UI_BASELINE "game-start").
	 *
	 * `href`: where it leads; Játszótér's own game page by default, the
	 * caller passes Játszótér's full address from the other apps.
	 * `featured`: Játszótér's home page's first game, a larger scene (the page
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
		<span class="game-poster-scene"><GameScene {slug} size={featured ? 210 : 160} /></span>
		<span class="game-poster-play"><StartButton iconOnly /></span>
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

	.game-poster-scene {
		display: grid;
		place-items: center;
		width: 100%;
		padding: 0 0.75rem;
		box-sizing: border-box;
		line-height: 0;
	}

	/* The scene stands finished until the card is hovered or focused; then
	   its loop starts from the beginning. */
	.game-poster:not(:hover):not(:focus-visible) .game-poster-scene :global(*) {
		animation: none;
	}

	.game-poster-play {
		position: absolute;
		right: 0.75rem;
		bottom: 0.75rem;
		display: grid;
		line-height: 0;
		opacity: 0;
		transition: opacity 0.2s ease;
	}

	.game-poster:hover .game-poster-play,
	.game-poster:focus-visible .game-poster-play {
		opacity: 1;
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
		.game-poster-play {
			transition: none;
		}
	}
</style>
