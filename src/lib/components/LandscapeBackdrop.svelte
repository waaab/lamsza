<script>
    import { browser } from "$app/environment";
    import WeatherSymbol from "$lib/icons/weather/WeatherSymbol.svelte";
    import { loadPlaceForecast } from "$lib/placeForecast.js";

    /**
     * The Székelyföld pages' landscape (UI_BASELINE "szf-pages"): three
     * mountain ridges in theme colours, and above them a red sun with a dashed
     * ring, or on a settlement's or attraction's page (`place`) that place's
     * current weather, drawn with the weather pages' animated symbols. The
     * symbol's slot keeps its size while the forecast loads, so nothing moves,
     * and no sun flashes first. Goes inside an element with the
     * `page-landscape` class (lamsza.css).
     *
     * @type {{ place?: string }}
     */
    let { place = "" } = $props();

    /** @type {string | null} */
    let symbol = $state(null);
    let label = $state("");

    $effect(() => {
        if (!browser || !place) return;
        const p = place;
        symbol = null;
        loadPlaceForecast(p).then((d) => {
            if (p !== place) return;
            symbol = d?.current?.symbol ?? null;
            label = d?.current?.desc ?? "";
        });
    });
</script>

{#if place}
    <span class="page-landscape__weather" aria-hidden="true">
        {#if symbol}<WeatherSymbol {symbol} size={160} {label} />{/if}
    </span>
{:else}
    <span class="page-landscape__sun" aria-hidden="true"></span>
{/if}
<svg class="page-landscape__ridge" viewBox="0 0 1200 110" preserveAspectRatio="none" aria-hidden="true" focusable="false">
    <path class="page-landscape__far" d="M0 110V60L120 38L240 64L380 24L520 58L650 18L800 54L950 32L1080 60L1200 40V110Z" />
    <path class="page-landscape__mid" d="M0 110V78L150 56L300 82L470 46L620 80L780 52L930 82L1100 60L1200 74V110Z" />
    <path class="page-landscape__near" d="M0 110V94L200 80L400 98L600 82L800 100L1000 84L1200 96V110Z" />
</svg>
