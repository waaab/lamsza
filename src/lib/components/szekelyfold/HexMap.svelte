<script>
    import { HEX_POINTS, honeycomb } from "$lib/szekelyfold.js";

    /**
     * A schematic honeycomb map of the seats or counties (UI_BASELINE
     * "szf-pages"): one hexagon each, in its colour, linking to its page.
     * Not to scale or shape; the order follows the map roughly. While the
     * list loads, `placeholders` empty hexagons hold the same size.
     *
     * @type {{ items: Array<{ name: string, href: string, color: string, sub?: string }>, label: string, placeholders?: number }}
     */
    let { items, label, placeholders = 0 } = $props();

    const count = $derived(items.length || placeholders);
    const layout = $derived(honeycomb(count));
</script>

<svg class="hex-map" viewBox="0 0 {layout.width} {layout.height}" role="group" aria-label={label}>
    {#if items.length}
        {#each items as item, i (item.href)}
            {@const c = layout.cells[i]}
            <a href={item.href} class="hex-map__cell" aria-label={item.sub ? `${item.name} ${item.sub}` : item.name}>
                <g transform="translate({c.x} {c.y})" style:--hex={item.color}>
                    <polygon points={HEX_POINTS} />
                    <circle cy={item.sub ? -18 : -12} r="4" />
                    <text y={item.sub ? 6 : 12}>{item.name}</text>
                    {#if item.sub}<text class="hex-map__sub" y="24">{item.sub}</text>{/if}
                </g>
            </a>
        {/each}
    {:else}
        {#each layout.cells as c, i (i)}
            <g transform="translate({c.x} {c.y})" class="hex-map__placeholder">
                <polygon points={HEX_POINTS} />
            </g>
        {/each}
    {/if}
</svg>

<style>
    .hex-map {
        display: block;
        width: 100%;
        max-width: 30rem;
        height: auto;
        margin: 0 auto;
        overflow: visible;
    }

    .hex-map__cell polygon {
        fill: var(--hex);
        fill-opacity: 0.18;
        stroke: var(--hex);
        stroke-width: 2;
        transform-box: fill-box;
        transform-origin: center;
        transition: transform 0.18s ease, fill-opacity 0.18s ease;
    }

    .hex-map__cell circle {
        fill: var(--hex);
    }

    .hex-map__cell text {
        fill: var(--text-primary);
        font-size: 12.5px;
        font-weight: 600;
        text-anchor: middle;
        pointer-events: none;
    }

    .hex-map__cell text.hex-map__sub {
        fill: var(--text-muted);
        font-size: 12px;
        font-weight: 400;
    }

    .hex-map__cell:hover polygon,
    .hex-map__cell:focus-visible polygon {
        fill-opacity: 0.38;
        transform: scale(1.05);
    }

    .hex-map__cell:focus-visible {
        outline: none;
    }

    .hex-map__cell:focus-visible polygon {
        stroke-width: 4;
    }

    .hex-map__placeholder polygon {
        fill: var(--border-color);
        fill-opacity: 0.35;
        stroke: var(--border-color);
        stroke-width: 2;
    }

    @media (prefers-reduced-motion: reduce) {
        .hex-map__cell polygon {
            transition: none;
        }
    }
</style>
