<!--
    The widgets' weather icon, in the style the admin picked
    (`weather_icon_style`): "svg" draws the animated WeatherSymbol, "emoji"
    an emoji. `symbol` is the MET symbol code the API sends since the move to
    MET Norway; `code` (the old OpenWeatherMap-style "02d") is the fallback
    for a reply cached in the browser before that.
-->
<script>
    import WeatherSymbol from "$lib/icons/weather/WeatherSymbol.svelte";
    import { weatherIconEmoji } from "$lib/utils";
    import { symbolEmoji } from "$lib/weatherSymbols.js";

    let { code = "02d", symbol = "", style = "emoji", size = "1.1em", animated = true } = $props();

    // An old reply has no symbol: draw its legacy code's nearest symbol.
    const LEGACY = {
        "01": "clearsky",
        "02": "fair",
        "03": "partlycloudy",
        "04": "cloudy",
        "09": "rainshowers",
        "10": "rain",
        "11": "rainandthunder",
        "13": "snow",
        "50": "fog",
    };
    const drawn = $derived.by(() => {
        if (symbol) return symbol;
        const base = LEGACY[String(code).slice(0, 2)] ?? "cloudy";
        const night = String(code).endsWith("n");
        return ["clearsky", "fair", "partlycloudy", "rainshowers"].includes(base) ? `${base}_${night ? "night" : "day"}` : base;
    });
</script>

{#if style === "svg"}
    <span class="weather-icon-svg" data-weather-symbol={drawn} aria-hidden="true">
        <WeatherSymbol symbol={drawn} {size} {animated} />
    </span>
{:else}
    <span class="weather-icon-emoji" aria-hidden="true">{symbol ? symbolEmoji(symbol) : weatherIconEmoji(code)}</span>
{/if}

<style>
    .weather-icon-svg {
        display: inline-block;
        line-height: 0;
        vertical-align: middle;
    }
    .weather-icon-emoji {
        display: inline-block;
        line-height: 1;
    }
</style>
