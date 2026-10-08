<!--
    The hour-by-hour strip of /idojaras/<place>: one card per hour (time,
    weather, rain, wind) with the temperature drawn as one line across all
    of them, so the day's ups and downs read at a glance. Each day's high and
    low get a coloured dot and a bold label; a dashed line marks midnight and
    names the new day.

    The line is an SVG laid over a band every card leaves empty, in rem units,
    so it lines up with the cards at any font size and scrolls and drags with
    them (lib/dragScroll.js; touch swipes natively). Geometry: lib/tempCurve.js.
-->
<script>
    import WeatherSymbol from "$lib/icons/weather/WeatherSymbol.svelte";
    import { dragScroll } from "$lib/dragScroll.js";
    import { compassHU } from "$lib/weatherSymbols.js";
    import { deg, hourLabel, dayName } from "$lib/weatherFormat.js";
    import { layoutCurve } from "$lib/tempCurve.js";

    /** @type {{ hours?: Array<{ time: string, symbol: string, desc: string, temp?: number | null, precip_mm?: number | null, wind_kph?: number | null, wind_dir?: number | null }> }} */
    let { hours = [] } = $props();

    // In rem. Keep in step with the card CSS below (--hs-card, --hs-gap) and
    // the band's height (--hs-band).
    const CARD = 3.6;
    const GAP = 0.25;
    const PITCH = CARD + GAP;
    const BAND = 5;
    // Room above the line for the labels, below it for the low dots.
    const box = { pitch: PITCH, card: CARD, top: 1.55, bottom: BAND - 0.55 };

    const uid = $props.id();
    const curve = $derived(layoutCurve(hours, box));
    const width = $derived(Math.max(0, hours.length * PITCH - GAP));
</script>

<div class="hs" role="list" aria-label="Óránkénti előrejelzés" {@attach dragScroll}>
    <div class="hs-track" style="--hs-card: {CARD}rem; --hs-gap: {GAP}rem; --hs-band: {BAND}rem">
        {#each hours as h (h.time)}
            <div class="hs-card" role="listitem">
                <span class="hs-time">{hourLabel(h.time)}</span>
                <WeatherSymbol symbol={h.symbol} size={40} animated={false} label={h.desc} />
                <span class="sr-only">{deg(h.temp)}</span>
                <span class="hs-band" aria-hidden="true"></span>
                <span class="hs-rain" class:hs-dim={!h.precip_mm}>{h.precip_mm ? `${h.precip_mm} mm` : "·"}</span>
                {#if h.wind_kph != null}
                    <span class="hs-wind" title="Szél: {compassHU(h.wind_dir)}">
                        <span class="hs-arrow" style="transform: rotate({((h.wind_dir ?? 0) + 180) % 360}deg)" aria-hidden="true">↑</span>
                        {Math.round(h.wind_kph)}
                    </span>
                {:else}
                    <span class="hs-wind" aria-hidden="true"></span>
                {/if}
            </div>
        {/each}

        {#if curve.points.length > 1}
            <svg
                class="hs-curve"
                viewBox="0 0 {width} {BAND}"
                style="width: {width}rem"
                aria-hidden="true"
                focusable="false"
            >
                <defs>
                    <linearGradient id="hs-line-{uid}" gradientUnits="userSpaceOnUse" x1="0" y1={box.top} x2="0" y2={box.bottom}>
                        <stop offset="0" style="stop-color: var(--warm-light)" />
                        <stop offset="1" style="stop-color: var(--szekely-blue)" />
                    </linearGradient>
                    <linearGradient id="hs-fill-{uid}" gradientUnits="userSpaceOnUse" x1="0" y1={box.top} x2="0" y2={BAND}>
                        <stop offset="0" style="stop-color: var(--warm-light); stop-opacity: 0.22" />
                        <stop offset="1" style="stop-color: var(--szekely-blue); stop-opacity: 0" />
                    </linearGradient>
                </defs>

                {#each curve.midnights as m (m.i)}
                    <line class="hs-midnight" x1={m.x} y1="0.15" x2={m.x} y2={BAND} />
                    <text class="hs-day" x={m.x + 0.18} y="0.6">{dayName(m.day)}</text>
                {/each}

                <path class="hs-area" d={curve.area} fill="url(#hs-fill-{uid})" />
                <path class="hs-line" d={curve.path} stroke="url(#hs-line-{uid})" />

                {#each curve.points as p (p.i)}
                    <circle
                        class="hs-dot"
                        class:hs-high={p.extreme === "high"}
                        class:hs-low={p.extreme === "low"}
                        cx={p.x}
                        cy={p.y}
                        r={p.extreme ? 0.2 : 0.11}
                    />
                    <text
                        class="hs-temp"
                        class:hs-high={p.extreme === "high"}
                        class:hs-low={p.extreme === "low"}
                        x={p.x}
                        y={p.y - 0.4}>{deg(p.temp)}</text
                    >
                {/each}
            </svg>
        {/if}
    </div>
</div>

<style>
    .hs {
        overflow-x: auto;
        padding: 0.5rem 0 0.75rem;
        scroll-snap-type: x proximity;
        overscroll-behavior-x: contain;
        scrollbar-width: thin;
    }
    /* Mouse drag (lib/dragScroll.js); touch swipes natively. */
    .hs:global(.drag-scroll) {
        cursor: grab;
    }
    .hs:global(.is-dragging) {
        cursor: grabbing;
        scroll-snap-type: none;
        user-select: none;
    }
    .hs-track {
        position: relative;
        display: flex;
        gap: var(--hs-gap);
        width: max-content;
        /* The band's top: card padding + time row + icon row + two gaps. */
        --hs-pad: 0.4rem;
        --hs-row-time: 1.2rem;
        --hs-row-icon: 2.5rem;
        --hs-row-gap: 0.2rem;
    }
    .hs-card {
        flex: 0 0 var(--hs-card);
        display: grid;
        grid-template-rows: var(--hs-row-time) var(--hs-row-icon) var(--hs-band) 1.1rem 1.1rem;
        row-gap: var(--hs-row-gap);
        justify-items: center;
        align-items: center;
        padding: var(--hs-pad) 0;
        border-radius: 10px;
        background: var(--card-bg);
        border: 1px solid var(--border-color);
        scroll-snap-align: start;
        font-size: 0.85rem;
        box-sizing: border-box;
    }
    /* sr-only spans sit outside the grid flow. */
    .hs-card :global(.sr-only) {
        position: absolute;
    }
    .hs-time {
        color: var(--text-faint);
    }
    .hs-rain {
        color: var(--szekely-blue);
        font-size: 0.75rem;
    }
    .hs-dim {
        color: var(--text-faintest);
    }
    .hs-wind {
        color: var(--text-faint);
        font-size: 0.75rem;
        display: inline-flex;
        align-items: center;
        gap: 0.1rem;
    }
    .hs-arrow {
        display: inline-block;
        line-height: 1;
    }
    .hs-curve {
        position: absolute;
        left: 0;
        /* Card border (1px) + padding + time + gap + icon + gap. */
        top: calc(1px + var(--hs-pad) + var(--hs-row-time) + var(--hs-row-gap) + var(--hs-row-icon) + var(--hs-row-gap));
        height: var(--hs-band);
        overflow: visible;
        pointer-events: none;
    }
    .hs-line {
        fill: none;
        stroke-width: 0.12;
        stroke-linecap: round;
        stroke-linejoin: round;
    }
    .hs-area {
        stroke: none;
    }
    .hs-dot {
        fill: var(--card-bg);
        stroke: var(--text-muted);
        stroke-width: 0.06;
    }
    .hs-dot.hs-high {
        fill: var(--warm-light);
        stroke: var(--card-bg);
    }
    .hs-dot.hs-low {
        fill: var(--szekely-blue);
        stroke: var(--card-bg);
    }
    .hs-temp {
        font-size: 0.78px;
        text-anchor: middle;
        fill: var(--text-primary);
        font-weight: 600;
        font-variant-numeric: tabular-nums;
    }
    .hs-temp.hs-high {
        fill: color-mix(in srgb, var(--warm-light) 75%, var(--text-primary));
        font-weight: 800;
    }
    .hs-temp.hs-low {
        fill: var(--szekely-blue);
        font-weight: 800;
    }
    .hs-midnight {
        stroke: var(--text-faintest);
        stroke-width: 0.04;
        stroke-dasharray: 0.15 0.15;
    }
    .hs-day {
        font-size: 0.62px;
        fill: var(--text-faint);
        font-weight: 600;
    }
</style>
