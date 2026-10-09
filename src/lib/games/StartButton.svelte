<!--
	The one start button for every game (UI_BASELINE "game-start"; owner,
	2026-10-09): on a game's page a red pill with the play triangle and
	"Kezdés", on its card the same triangle in a red circle, without text.
	Shared from lamsza to Játszótér and Szótár.

	`iconOnly`: the card's version, a decorative span (the whole card is the
	link, so it cannot hold a button). `label`: the pill's text; a game may
	say "Folytatás" or "Eredmény" when that is what the button does.
-->
<script>
	/** @type {{ label?: string, iconOnly?: boolean, disabled?: boolean, onclick?: (event: MouseEvent) => void }} */
	let { label = 'Kezdés', iconOnly = false, disabled = false, onclick } = $props();
</script>

{#snippet triangle()}
	<svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true" focusable="false">
		<path d="M5 2.4 21.2 12 5 21.6z" />
	</svg>
{/snippet}

{#if iconOnly}
	<span class="start-btn start-btn--icon" aria-hidden="true">{@render triangle()}</span>
{:else}
	<button type="button" class="btn btn-lg start-btn" {disabled} {onclick}>
		{@render triangle()}
		{label}
	</button>
{/if}

<style>
	.start-btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.55rem;
		border-radius: 999px;
		background: var(--szekely-red);
		border-color: var(--szekely-red);
		color: #fff;
		font-weight: 700;
	}

	button.start-btn {
		min-width: 11rem;
	}

	button.start-btn:hover,
	button.start-btn:focus {
		background: var(--szekely-red);
		border-color: var(--szekely-red);
		color: #fff;
	}

	.start-btn svg {
		width: 1.4em;
		height: 1.4em;
	}

	/* The card's version: the same red and triangle in a circle, with a thin
	   white ring so it stands out on every game colour. */
	.start-btn--icon {
		width: 2.5rem;
		height: 2.5rem;
		box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.9);
	}

	.start-btn--icon svg {
		width: 1.05rem;
		height: 1.05rem;
		margin-left: 0.1rem;
	}
</style>
