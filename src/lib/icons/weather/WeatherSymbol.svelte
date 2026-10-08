<!--
    The animated weather icon (UI_BASELINE "wx-icons"): one drawing for each
    of MET Norway's 41 symbol codes, day and night, built from a few parts,
    the sun or the moon, one or two clouds, rain, sleet or snow, lightning
    and fog, so every icon matches the others.

    Inline SVG on the theme's tokens, so it follows light and dark mode with
    no colours of its own: sun and lightning --warm-light, rain
    --szekely-blue, clouds from --text-muted. The sun's rays turn, clouds
    drift, drops fall, flakes spin, lightning flashes; all of it CSS only,
    and still when the visitor prefers reduced motion. It reserves its size
    before it draws, so nothing shifts.

    Decorative by default (aria-hidden): the caller shows the text beside it.
    Pass `label` when the icon stands alone.
-->
<script>
    import { symbolParts } from "$lib/weatherSymbols.js";

    let { symbol = "cloudy", size = 48, animated = true, label = "" } = $props();

    const parts = $derived(symbolParts(symbol));
    // With rain or snow under it the cloud sits higher to make room.
    const cloudY = $derived(parts.precip ? 10 : 20);
    // With lightning the bolt takes the middle, so the drops go either side.
    const dropXs = $derived(
        (parts.thunder
            ? { 2: [21, 47], 3: [19, 45, 53], 4: [15, 22, 46, 53] }
            : { 2: [27, 39], 3: [23, 33, 43], 4: [19, 28, 37, 46] })[parts.drops] ?? [],
    );
    // Clear: the sun or moon alone, large. Fair: large, a small cloud in
    // front. Otherwise small, top left, behind the cloud (higher when rain
    // or snow pushes the cloud up).
    const fair = $derived(parts.sky === "fair");
    const sunAt = $derived(
        parts.bodyFull
            ? { x: 32, y: 32, s: 1 }
            : fair
              ? { x: 27, y: 26, s: 0.86 }
              : { x: 21, y: parts.precip ? 15 : 21, s: 0.62 },
    );
    const cloudTransform = $derived(
        fair
            ? "translate(27 31) scale(0.66)"
            : parts.body
              ? `translate(14 ${cloudY})`
              : `translate(8 ${cloudY})`,
    );

    const CLOUD = "M12 28 A9 9 0 1 1 13.39 10.11 A11 11 0 0 1 34.59 10.01 A9 9 0 1 1 35 28 Z";
    const MOON = "M30.44 19.09 A13 13 0 1 0 44.52 35.52 A11 11 0 0 1 30.44 19.09 Z";
    const RAYS = [0, 45, 90, 135, 180, 225, 270, 315];
</script>

<svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 64 64"
    width={size}
    height={size}
    class="wx"
    class:wx-anim={animated}
    role={label ? "img" : undefined}
    aria-label={label || undefined}
    aria-hidden={label ? undefined : "true"}
    focusable="false"
>
    {#if parts.stars}
        <g class="wx-stars">
            <path class="wx-star" d="M14 13 l1.2 2.8 2.8 1.2 -2.8 1.2 -1.2 2.8 -1.2 -2.8 -2.8 -1.2 2.8 -1.2 Z" />
            <path class="wx-star wx-d2" d="M52 46 l0.9 2.1 2.1 0.9 -2.1 0.9 -0.9 2.1 -0.9 -2.1 -2.1 -0.9 2.1 -0.9 Z" />
            <path class="wx-star wx-d3" d="M53 11 l0.7 1.6 1.6 0.7 -1.6 0.7 -0.7 1.6 -0.7 -1.6 -1.6 -0.7 1.6 -0.7 Z" />
        </g>
    {/if}

    {#if parts.body === "sun"}
        <g transform="translate({sunAt.x} {sunAt.y}) scale({sunAt.s})">
            <g class="wx-rays">
                {#each RAYS as a (a)}
                    <line x1="0" y1="-16" x2="0" y2="-21" transform="rotate({a})" />
                {/each}
            </g>
            <circle class="wx-sun" r="11" />
        </g>
    {:else if parts.body === "moon"}
        <g transform="translate({sunAt.x} {sunAt.y}) scale({sunAt.s}) translate(-32 -32)">
            <path class="wx-moon" d={MOON} />
        </g>
    {/if}

    {#if parts.clouds === 2}
        <g class="wx-drift-b">
            <path class="wx-cloud wx-cloud-back" d={CLOUD} transform="translate(19 9) scale(0.78)" />
        </g>
    {/if}
    {#if parts.clouds > 0}
        <g class="wx-drift">
            <path
                class="wx-cloud"
                class:wx-cloud-dark={parts.dark}
                d={CLOUD}
                transform={cloudTransform}
            />
        </g>
    {/if}

    {#if parts.thunder}
        <path class="wx-bolt" d="M35 33 L28 45 H33 L30 56 L41 41 H35.5 L39 33 Z" />
    {/if}

    {#if parts.precip}
        {#each dropXs as x, i (x)}
            {@const snow = parts.precip === "snow" || (parts.precip === "sleet" && i % 2 === 1)}
            {#if snow}
                <g class="wx-flake" style="animation-delay: {(i * 0.7).toFixed(2)}s">
                    <g transform="translate({x} 48)">
                        <path d="M0 -4 V4 M-3.5 -2 L3.5 2 M-3.5 2 L3.5 -2" />
                    </g>
                </g>
            {:else}
                <line
                    class="wx-drop"
                    class:wx-drop-heavy={parts.drops === 4}
                    x1={x}
                    y1="43"
                    x2={x - 2.6}
                    y2="52"
                    style="animation-delay: {(i * 0.27 + (parts.showers ? 0.1 : 0)).toFixed(2)}s"
                />
            {/if}
        {/each}
    {/if}

    {#if parts.fog}
        <g class="wx-fog">
            <line x1="12" y1="24" x2="46" y2="24" />
            <line class="wx-fog-alt" x1="18" y1="32" x2="54" y2="32" />
            <line x1="10" y1="40" x2="48" y2="40" />
            <line class="wx-fog-alt" x1="20" y1="48" x2="52" y2="48" />
        </g>
    {/if}
</svg>

<style>
    .wx {
        display: block;
        flex: none;
        overflow: visible;
        --wx-sun: var(--warm-light);
        --wx-moon: color-mix(in srgb, var(--warm-light) 62%, var(--text-muted));
        --wx-cloud-line: var(--text-muted);
        --wx-cloud-fill: color-mix(in srgb, var(--text-muted) 14%, var(--card-bg));
        --wx-cloud-dark: color-mix(in srgb, var(--text-muted) 42%, var(--card-bg));
        --wx-rain: var(--szekely-blue);
        --wx-snow: color-mix(in srgb, var(--szekely-blue) 45%, var(--text-muted));
        --wx-fog: var(--text-faint);
    }
    .wx-sun {
        fill: var(--wx-sun);
    }
    .wx-rays line {
        stroke: var(--wx-sun);
        stroke-width: 2.6;
        stroke-linecap: round;
    }
    .wx-moon {
        fill: var(--wx-moon);
    }
    .wx-star {
        fill: var(--wx-moon);
    }
    .wx-cloud {
        fill: var(--wx-cloud-fill);
        stroke: var(--wx-cloud-line);
        stroke-width: 2;
        stroke-linejoin: round;
    }
    .wx-cloud-back {
        opacity: 0.75;
    }
    .wx-cloud-dark {
        fill: var(--wx-cloud-dark);
        stroke: var(--text-secondary);
    }
    .wx-bolt {
        fill: var(--wx-sun);
        stroke: color-mix(in srgb, var(--wx-sun) 70%, var(--text-primary));
        stroke-width: 0.8;
        stroke-linejoin: round;
    }
    .wx-drop {
        stroke: var(--wx-rain);
        stroke-width: 2.2;
        stroke-linecap: round;
    }
    .wx-drop-heavy {
        stroke-width: 2.6;
    }
    .wx-flake path {
        stroke: var(--wx-snow);
        stroke-width: 1.9;
        stroke-linecap: round;
        fill: none;
    }
    .wx-fog line {
        stroke: var(--wx-fog);
        stroke-width: 2.6;
        stroke-linecap: round;
    }

    /* Motion. transform-box makes each part turn and move about itself. */
    .wx-anim .wx-rays,
    .wx-anim .wx-moon,
    .wx-anim .wx-star,
    .wx-anim .wx-drift,
    .wx-anim .wx-drift-b,
    .wx-anim .wx-drop,
    .wx-anim .wx-flake,
    .wx-anim .wx-bolt,
    .wx-anim .wx-fog line {
        transform-box: fill-box;
        transform-origin: center;
    }
    .wx-anim .wx-rays {
        animation: wx-spin 36s linear infinite;
    }
    .wx-anim .wx-sun {
        animation: wx-breathe 5s ease-in-out infinite;
        transform-box: fill-box;
        transform-origin: center;
    }
    .wx-anim .wx-moon {
        animation: wx-glow 6s ease-in-out infinite;
    }
    .wx-anim .wx-star {
        animation: wx-twinkle 3.2s ease-in-out infinite;
    }
    .wx-anim .wx-d2 {
        animation-delay: 1.1s;
    }
    .wx-anim .wx-d3 {
        animation-delay: 2.1s;
    }
    .wx-anim .wx-drift {
        animation: wx-drift 8s ease-in-out infinite alternate;
    }
    .wx-anim .wx-drift-b {
        animation: wx-drift 11s ease-in-out infinite alternate-reverse;
    }
    .wx-anim .wx-drop {
        animation: wx-fall 1.1s linear infinite;
    }
    .wx-anim .wx-flake {
        animation: wx-snowfall 2.8s linear infinite;
    }
    .wx-anim .wx-bolt {
        animation: wx-flash 3.6s linear infinite;
    }
    .wx-anim .wx-fog line {
        animation: wx-slide 6s ease-in-out infinite alternate;
    }
    .wx-anim .wx-fog .wx-fog-alt {
        animation-direction: alternate-reverse;
    }

    @keyframes wx-spin {
        to {
            transform: rotate(360deg);
        }
    }
    @keyframes wx-breathe {
        0%,
        100% {
            transform: scale(1);
        }
        50% {
            transform: scale(1.05);
        }
    }
    @keyframes wx-glow {
        0%,
        100% {
            opacity: 1;
        }
        50% {
            opacity: 0.78;
        }
    }
    @keyframes wx-twinkle {
        0%,
        100% {
            opacity: 0.2;
            transform: scale(0.7);
        }
        50% {
            opacity: 1;
            transform: scale(1);
        }
    }
    @keyframes wx-drift {
        from {
            transform: translateX(-1.5px);
        }
        to {
            transform: translateX(1.5px);
        }
    }
    @keyframes wx-fall {
        0% {
            transform: translate(1px, -3px);
            opacity: 0;
        }
        20% {
            opacity: 1;
        }
        100% {
            transform: translate(-2px, 9px);
            opacity: 0;
        }
    }
    @keyframes wx-snowfall {
        0% {
            transform: translate(0, -4px) rotate(0deg);
            opacity: 0;
        }
        20% {
            opacity: 1;
        }
        100% {
            transform: translate(-1.5px, 10px) rotate(120deg);
            opacity: 0;
        }
    }
    @keyframes wx-flash {
        0%,
        62%,
        68%,
        74%,
        100% {
            opacity: 0.7;
        }
        64%,
        72% {
            opacity: 1;
        }
    }
    @keyframes wx-slide {
        from {
            transform: translateX(-3px);
        }
        to {
            transform: translateX(3px);
        }
    }

    @media (prefers-reduced-motion: reduce) {
        .wx-anim *,
        .wx-anim {
            animation: none !important;
        }
        .wx-anim .wx-drop,
        .wx-anim .wx-flake,
        .wx-anim .wx-star {
            opacity: 1;
        }
        .wx-anim .wx-bolt {
            opacity: 1;
        }
    }
</style>
