<!--
    Small animated drawings for the weather's measurements (UI_BASELINE
    "wx-icons"), each showing its own value where it can: the thermometer
    fills to the temperature, the wind arrow points where the wind blows, the
    drop fills to the humidity, the pressure needle turns, the moon shows its
    phase.

    Same rules as WeatherSymbol: inline SVG on theme tokens (lines in
    currentColor at stroke 2, the ic-system weight), CSS motion only, still
    under prefers-reduced-motion, a fixed size. Decorative unless `label`.
-->
<script>
    import { moonLitPath } from "$lib/weatherSymbols.js";

    /**
     * kind: thermometer | feels | wind | gust | humidity | precip | precip_prob
     * | pressure | uv | cloud | fog | dewpoint | sunrise | sunset | moonrise
     * | moonset | moonphase | daylength
     */
    let { kind = "thermometer", value = null, size = 24, animated = true, label = "" } = $props();
    // Clip paths need ids unique on the page: several glyphs share a page.
    const uid = $props.id();

    const num = $derived(value == null || Number.isNaN(Number(value)) ? null : Number(value));
    const clamp = (/** @type {number} */ v, /** @type {number} */ lo, /** @type {number} */ hi) => Math.min(hi, Math.max(lo, v));

    // Thermometer: -20 °C empty, 35 °C full; the tube runs y 21 to 5.
    const mercuryTop = $derived(num == null ? 15 : 21 - (clamp(num, -20, 35) + 20) * (16 / 55));
    const cold = $derived(num != null && num <= 0);
    // Wind: MET gives where it comes from; the arrow points where it goes.
    const windTo = $derived(num == null ? 45 : (num + 180) % 360);
    // Humidity and rain chance: 0-100 % fills the shape bottom up.
    const level = $derived(num == null ? 0.5 : clamp(num, 0, 100) / 100);
    // Pressure: 970 to 1050 hPa sweeps the needle from -120° to 120°.
    const needle = $derived(num == null ? 0 : ((clamp(num, 970, 1050) - 1010) / 40) * 120);
    // Rain gauge: 0 to 20 mm fills it.
    const gauge = $derived(num == null ? 0.3 : clamp(num, 0, 20) / 20);
    const uvColor = $derived(
        num == null || num < 3
            ? "var(--google-green)"
            : num < 6
              ? "var(--warm-light)"
              : num < 8
                ? "var(--warning-orange)"
                : "var(--szekely-red)",
    );
    const litPath = $derived(moonLitPath(num ?? 180, 16, 16, 10));
</script>

<svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 32 32"
    width={size}
    height={size}
    class="wg"
    class:wg-anim={animated}
    fill="none"
    stroke="currentColor"
    stroke-width="2"
    stroke-linecap="round"
    stroke-linejoin="round"
    role={label ? "img" : undefined}
    aria-label={label || undefined}
    aria-hidden={label ? undefined : "true"}
    focusable="false"
>
    {#if kind === "thermometer" || kind === "feels"}
        <clipPath id="wg-tube-{uid}">
            <rect x="12" y={mercuryTop} width="6" height={27 - mercuryTop} />
        </clipPath>
        <g class="wg-rise">
            <path class="wg-mercury" class:wg-cold={cold} d="M13 4 h4 v17.5 a5 5 0 1 1 -4 0 Z" clip-path="url(#wg-tube-{uid})" stroke="none" />
        </g>
        <path d="M13 21.5 V6 a2 2 0 0 1 4 0 V21.5 a5 5 0 1 1 -4 0 Z" />
        {#if kind === "feels"}
            <path class="wg-waft" d="M23 9 q2 2 0 4 q-2 2 0 4 M27 9 q2 2 0 4 q-2 2 0 4" stroke-width="1.6" />
        {:else}
            <path d="M20 8 h3 M20 12 h2 M20 16 h3" stroke-width="1.4" opacity="0.6" />
        {/if}
    {:else if kind === "wind" || kind === "gust"}
        <g class="wg-flow">
            <path d="M3 11 H18 a3.5 3.5 0 1 0 -3.5 -3.5" />
            <path d="M3 17 H24 a3.5 3.5 0 1 1 -3.5 3.5" />
            {#if kind === "gust"}
                <path d="M3 23 H11" class="wg-gust" />
            {/if}
        </g>
        {#if kind === "wind" && num != null}
            <g transform="rotate({windTo} 26 8)">
                <path class="wg-arrow" d="M26 3.5 V12.5 M23 6.5 L26 3.5 L29 6.5" stroke-width="1.8" />
            </g>
        {/if}
    {:else if kind === "humidity" || kind === "dewpoint"}
        <clipPath id="wg-drop-{uid}">
            <rect x="0" y={28 - 22 * level} width="32" height={22 * level + 1} />
        </clipPath>
        <path class="wg-water" d="M16 4 C16 4 7 14.5 7 20 a9 9 0 0 0 18 0 C25 14.5 16 4 16 4 Z" clip-path="url(#wg-drop-{uid})" stroke="none" />
        <path class="wg-bob" d="M16 4 C16 4 7 14.5 7 20 a9 9 0 0 0 18 0 C25 14.5 16 4 16 4 Z" />
        {#if kind === "dewpoint"}
            <path d="M12 21 a4 4 0 0 0 4 4" stroke-width="1.5" opacity="0.7" />
        {/if}
    {:else if kind === "precip"}
        <path class="wg-water" d="M9 {27 - 15 * gauge} H23 V27 H9 Z" stroke="none" opacity="0.85" />
        <path d="M8 9 V26 a1.5 1.5 0 0 0 1.5 1.5 H22.5 a1.5 1.5 0 0 0 1.5 -1.5 V9" />
        <path d="M24 14 h-3 M24 19 h-3 M24 24 h-3" stroke-width="1.3" opacity="0.6" />
        <path class="wg-fall wg-water-line" d="M16 2.5 v3.5" />
    {:else if kind === "precip_prob"}
        <path d="M4 15 a12 10 0 0 1 24 0 Z" class="wg-umbrella" />
        <path d="M16 15 V25 a2.5 2.5 0 0 1 -5 0" />
        <path class="wg-fall wg-water-line" d="M7 20 v3" />
        <path class="wg-fall wg-water-line wg-d2" d="M25 19 v3" />
    {:else if kind === "pressure"}
        <path d="M5.5 22 A12 12 0 1 1 26.5 22" />
        <path d="M16 6.5 v2 M7.5 12 l1.6 1 M24.5 12 l-1.6 1" stroke-width="1.4" opacity="0.6" />
        <g class="wg-needle" style="--wg-needle: {needle}deg">
            <path d="M16 18 L16 9" class="wg-accent" />
        </g>
        <circle cx="16" cy="18" r="2" fill="currentColor" stroke="none" />
    {:else if kind === "uv"}
        <g class="wg-rays" style="color: {uvColor}">
            {#each [0, 45, 90, 135, 180, 225, 270, 315] as a (a)}
                <line x1="16" y1="3" x2="16" y2="6" transform="rotate({a} 16 16)" />
            {/each}
        </g>
        <circle cx="16" cy="16" r="6.5" fill={uvColor} stroke="none" class="wg-core" />
    {:else if kind === "cloud"}
        <path class="wg-drift" d="M9 24 A5.5 5.5 0 1 1 9.9 13.1 A7 7 0 0 1 23.3 12.9 A5.5 5.5 0 1 1 23.5 24 Z" />
    {:else if kind === "fog"}
        <path class="wg-slide" d="M5 10 H23 M9 16 H27" />
        <path class="wg-slide wg-rev" d="M4 22 H20 M12 28 H26" />
    {:else if kind === "sunrise" || kind === "sunset" || kind === "moonrise" || kind === "moonset"}
        {@const up = kind === "sunrise" || kind === "moonrise"}
        <clipPath id="wg-horizon-{uid}">
            <rect x="0" y="0" width="32" height="22" />
        </clipPath>
        <g clip-path="url(#wg-horizon-{uid})">
            <g class={up ? "wg-up" : "wg-down"}>
                {#if kind.startsWith("sun")}
                    <circle cx="16" cy="22" r="6" class="wg-sunfill" stroke="none" />
                    <path d="M16 11 v-2.5 M8.5 15 l-1.8 -1.5 M23.5 15 l1.8 -1.5" class="wg-sunline" />
                {:else}
                    <path d="M16 16 a6 6 0 1 0 6 6 a4.8 4.8 0 0 1 -6 -6 Z" class="wg-moonfill" stroke="none" />
                {/if}
            </g>
        </g>
        <path d="M3 22 H29" />
        <path d={up ? "M16 30 v-4.5 M13.5 28 l2.5 -2.5 l2.5 2.5" : "M16 25.5 v4.5 M13.5 27.5 l2.5 2.5 l2.5 -2.5"} stroke-width="1.6" />
    {:else if kind === "moonphase"}
        <circle cx="16" cy="16" r="10" class="wg-moondark" />
        {#if litPath}
            <path d={litPath} class="wg-moonfill wg-glow" stroke="none" />
        {/if}
    {:else if kind === "daylength"}
        <path d="M4 24 A12 12 0 0 1 28 24" stroke-dasharray="2 3" opacity="0.6" />
        <path d="M2 24 H30" />
        <g class="wg-orbit">
            <circle cx="16" cy="12" r="3" class="wg-sunfill" stroke="none" />
        </g>
    {/if}
</svg>

<style>
    .wg {
        display: inline-block;
        flex: none;
        vertical-align: middle;
        overflow: visible;
    }
    .wg-mercury {
        fill: var(--szekely-red);
    }
    .wg-mercury.wg-cold {
        fill: var(--szekely-blue);
    }
    .wg-water {
        fill: color-mix(in srgb, var(--szekely-blue) 55%, transparent);
    }
    .wg-water-line {
        stroke: var(--szekely-blue);
    }
    .wg-accent,
    .wg-arrow {
        stroke: var(--szekely-red);
    }
    .wg-sunfill {
        fill: var(--warm-light);
    }
    .wg-sunline {
        stroke: var(--warm-light);
    }
    .wg-moonfill {
        fill: color-mix(in srgb, var(--warm-light) 62%, var(--text-muted));
    }
    .wg-moondark {
        fill: color-mix(in srgb, var(--text-muted) 18%, var(--card-bg));
        stroke-width: 1.4;
        opacity: 0.9;
    }
    .wg-umbrella {
        fill: color-mix(in srgb, var(--szekely-blue) 18%, transparent);
    }
    .wg-needle {
        transform-box: view-box;
        transform-origin: 16px 18px;
        transform: rotate(var(--wg-needle));
    }

    .wg-anim .wg-rise {
        animation: wg-rise 1.4s ease-out both;
    }
    .wg-anim .wg-waft {
        animation: wg-waft 2.4s ease-in-out infinite;
    }
    .wg-anim .wg-flow {
        stroke-dasharray: 26 6;
        animation: wg-flow 1.8s linear infinite;
    }
    .wg-anim .wg-gust {
        animation: wg-gust 1.2s ease-in-out infinite;
    }
    .wg-anim .wg-bob {
        animation: wg-bob 3s ease-in-out infinite;
        transform-box: fill-box;
        transform-origin: center bottom;
    }
    .wg-anim .wg-fall {
        animation: wg-fall 1.3s linear infinite;
    }
    .wg-anim .wg-d2 {
        animation-delay: 0.6s;
    }
    .wg-anim .wg-needle {
        animation: wg-needle 1.6s ease-out both;
    }
    .wg-anim .wg-rays {
        animation: wg-spin 24s linear infinite;
        transform-box: view-box;
        transform-origin: 16px 16px;
    }
    .wg-anim .wg-core {
        animation: wg-pulse 3s ease-in-out infinite;
        transform-box: fill-box;
        transform-origin: center;
    }
    .wg-anim .wg-drift {
        animation: wg-drift 6s ease-in-out infinite alternate;
    }
    .wg-anim .wg-slide {
        animation: wg-drift 5s ease-in-out infinite alternate;
    }
    .wg-anim .wg-rev {
        animation-direction: alternate-reverse;
    }
    .wg-anim .wg-up {
        animation: wg-up 4s ease-in-out infinite;
    }
    .wg-anim .wg-down {
        animation: wg-down 4s ease-in-out infinite;
    }
    .wg-anim .wg-glow {
        animation: wg-glow 5s ease-in-out infinite;
    }
    .wg-anim .wg-orbit {
        animation: wg-orbit 6s ease-in-out infinite alternate;
        transform-box: view-box;
        transform-origin: 16px 24px;
    }

    @keyframes wg-rise {
        from {
            transform: translateY(8px);
            opacity: 0.4;
        }
    }
    @keyframes wg-waft {
        0%,
        100% {
            transform: translateY(1px);
            opacity: 0.5;
        }
        50% {
            transform: translateY(-1.5px);
            opacity: 1;
        }
    }
    @keyframes wg-flow {
        to {
            stroke-dashoffset: -32;
        }
    }
    @keyframes wg-gust {
        0%,
        100% {
            transform: translateX(0);
            opacity: 0.4;
        }
        50% {
            transform: translateX(4px);
            opacity: 1;
        }
    }
    @keyframes wg-bob {
        0%,
        100% {
            transform: scale(1);
        }
        50% {
            transform: scale(1.04, 0.97);
        }
    }
    @keyframes wg-fall {
        0% {
            transform: translateY(-3px);
            opacity: 0;
        }
        30% {
            opacity: 1;
        }
        100% {
            transform: translateY(5px);
            opacity: 0;
        }
    }
    @keyframes wg-needle {
        from {
            transform: rotate(-120deg);
        }
    }
    @keyframes wg-spin {
        to {
            transform: rotate(360deg);
        }
    }
    @keyframes wg-pulse {
        0%,
        100% {
            transform: scale(1);
        }
        50% {
            transform: scale(1.08);
        }
    }
    @keyframes wg-drift {
        from {
            transform: translateX(-1.5px);
        }
        to {
            transform: translateX(1.5px);
        }
    }
    @keyframes wg-up {
        0% {
            transform: translateY(7px);
        }
        60%,
        100% {
            transform: translateY(0);
        }
    }
    @keyframes wg-down {
        0%,
        40% {
            transform: translateY(0);
        }
        100% {
            transform: translateY(7px);
        }
    }
    @keyframes wg-glow {
        0%,
        100% {
            opacity: 1;
        }
        50% {
            opacity: 0.75;
        }
    }
    @keyframes wg-orbit {
        from {
            transform: rotate(-70deg);
        }
        to {
            transform: rotate(70deg);
        }
    }

    @media (prefers-reduced-motion: reduce) {
        .wg-anim *,
        .wg-anim {
            animation: none !important;
        }
    }
</style>
