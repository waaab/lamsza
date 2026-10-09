<script>
    import AppIcon from "$lib/icons/AppIcon.svelte";
    import { REGION_PAGES } from "$lib/szekelyfold.js";

    /**
     * "Oldal navigáció" on the Székelyföld pages (UI_BASELINE "szf-pages"):
     * a card for each of the other three, with the toolbar's icon. `row`
     * lays them side by side (under a list) instead of stacked (beside the
     * map).
     *
     * @type {{ current: string, row?: boolean }}
     */
    let { current, row = false } = $props();

    const others = $derived(REGION_PAGES.filter((p) => p.key !== current));
</script>

<nav class="region-nav" aria-labelledby="region-nav-title">
    <h2 id="region-nav-title" class="widget-title">Oldal navigáció</h2>
    <ul class="region-nav__list" class:region-nav__list--row={row}>
        {#each others as p (p.key)}
            <li>
                <a class="card region-nav__card" href={p.href}>
                    <span class="region-nav__icon"><AppIcon name={p.icon} size={26} /></span>
                    <span class="region-nav__label">{p.label}</span>
                    <span class="region-nav__more" aria-hidden="true">›</span>
                </a>
            </li>
        {/each}
    </ul>
</nav>

<style>
    .region-nav__list {
        display: grid;
        gap: 0.75rem;
        list-style: none;
        margin: 0;
        padding: 0;
    }

    .region-nav__list--row {
        grid-template-columns: repeat(auto-fit, minmax(min(100%, 14rem), 1fr));
    }

    .region-nav__card {
        display: flex;
        align-items: center;
        gap: 1rem;
        margin: 0;
        padding: 1rem 1.15rem;
        font-weight: 600;
        color: inherit;
        text-decoration: none;
        transition: transform 0.15s ease, border-color 0.15s ease;
    }

    .region-nav__card:hover,
    .region-nav__card:focus-visible {
        transform: translateX(4px);
        border-color: var(--szekely-red);
    }

    .region-nav__icon {
        display: grid;
        place-items: center;
        flex: none;
        color: var(--szekely-red);
    }

    .region-nav__label {
        flex: 1;
    }

    .region-nav__more {
        color: var(--text-muted);
    }

    @media (prefers-reduced-motion: reduce) {
        .region-nav__card {
            transition: none;
        }
    }
</style>
