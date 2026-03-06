<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import { weatherIconEmoji, formatTime } from "$lib/utils";

    export let settlementSlug = null; // Optional: fetch for specific settlement
    export let cityName = "Miercurea Ciuc"; // Default city

    const WEATHER_CACHE_KEY =
        "weather_cache" + (settlementSlug ? `_${settlementSlug}` : "");
    const WEATHER_TTL = 30 * 60 * 1000;

    let weatherData = null;
    let loading = true;
    let error = false;

    onMount(async () => {
        const cached = localStorage.getItem(WEATHER_CACHE_KEY);
        if (cached) {
            try {
                const data = JSON.parse(cached);
                if (Date.now() - data.timestamp < WEATHER_TTL) {
                    weatherData = data;
                    loading = false;
                    return;
                }
            } catch (e) {}
        }

        try {
            const endpoint = settlementSlug
                ? `/api/weather?slug=${encodeURIComponent(settlementSlug)}`
                : `/api/proxy?url=${encodeURIComponent(`https://api.openweathermap.org/data/2.5/weather?q=${encodeURIComponent(cityName)},RO&units=metric&appid=${import.meta.env.VITE_WEATHER_API_KEY}&lang=hu`)}`;

            const data = await apiFetch(endpoint);

            weatherData = {
                temp: Math.round(data.main.temp),
                tempMin: Math.round(data.main.temp_min),
                desc: data.weather[0].description,
                icon: data.weather[0].icon,
                timestamp: Date.now(),
            };

            localStorage.setItem(
                WEATHER_CACHE_KEY,
                JSON.stringify(weatherData),
            );
        } catch (err) {
            error = true;
        } finally {
            loading = false;
        }
    });
</script>

<article id="idojaras" class="weather-card">
    {#if loading}
        <div class="weather-left">
            <span class="widget-title">Időjárás</span>
            <div class="weather-temp-row">
                <div class="skeleton weather-skeleton-temp"></div>
            </div>
            <div class="weather-desc">
                <div class="skeleton weather-skeleton-desc"></div>
            </div>
            <div class="weather-footer">
                <div class="skeleton skeleton-footer-1"></div>
                <div class="skeleton skeleton-footer-2"></div>
            </div>
        </div>
        <div class="weather-right">
            <span class="weather-icon">⛅</span>
        </div>
    {:else if error}
        <div class="weather-left">
            <span class="widget-title">Időjárás</span>
            <p class="weather-error">Időjárás adat nem elérhető.</p>
        </div>
        <div class="weather-right">
            <span class="weather-icon">⛅</span>
        </div>
    {:else if weatherData}
        <div class="weather-left">
            <span class="widget-title">Időjárás</span>
            <div class="weather-temp-row">
                <span class="weather-temp">{weatherData.temp}</span><span
                    class="weather-temp-unit">°C</span
                >
                {#if weatherData.tempMin != null}
                    <span class="weather-temp-min"
                        >/ {weatherData.tempMin}°C</span
                    >
                {/if}
            </div>
            <div class="weather-desc capitalize">
                {weatherData.desc}
            </div>
            <div class="weather-footer">
                <small class="weather-timestamp"
                    >Utoljára frissítve: {formatTime(
                        weatherData.timestamp,
                    )}</small
                >
                <small class="weather-source">Forrás: OpenWeatherMap</small>
            </div>
        </div>
        <div class="weather-right">
            <span class="weather-icon"
                >{weatherIconEmoji(weatherData.icon)}</span
            >
        </div>
    {/if}
</article>

<style>
    .capitalize {
        text-transform: capitalize;
    }
    .weather-error {
        color: var(--text-faint);
        margin: 0.5rem 0 0;
    }
    /* Skeleton sizes preserved from original */
    .skeleton-footer-1 {
        width: 120px;
        height: 0.75rem;
    }
    .skeleton-footer-2 {
        width: 90px;
        height: 0.75rem;
    }
</style>
