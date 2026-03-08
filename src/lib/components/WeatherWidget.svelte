<script>
    import { browser } from "$app/environment";
    import { apiFetch } from "$lib/api";
    import { formatTime } from "$lib/utils";
    import WeatherIcon from "$lib/components/WeatherIcon.svelte";

    export let settlementSlug = "csikszereda";
    /** When true, show precipitation, humidity, wind in .weather-details */
    export let advanced = false;

    const DEFAULT_TTL_MS = 15 * 60 * 1000;

    let weatherData = null;
    let loading = true;
    let error = false;
    /** "emoji" | "svg" from admin setting */
    let weatherIconStyle = "emoji";

    async function fetchWeather() {
        if (!settlementSlug) return;
        const cacheKey = "weather_cache_" + settlementSlug;
        loading = true;
        error = false;
        let ttlMs = DEFAULT_TTL_MS;
        let cacheVersion = "";

        try {
            const configRes = await apiFetch("/api/config/public");
            if (configRes && configRes.weather_cache_ttl_minutes != null) {
                ttlMs = configRes.weather_cache_ttl_minutes * 60 * 1000;
            }
            if (configRes && configRes.weather_cache_version != null) {
                cacheVersion = String(configRes.weather_cache_version);
            }
            if (configRes && configRes.weather_icon_style != null) {
                weatherIconStyle = configRes.weather_icon_style === "svg" ? "svg" : "emoji";
            }
        } catch (e) {}

        if (browser) {
            const cached = localStorage.getItem(cacheKey);
            if (cached) {
                try {
                    const data = JSON.parse(cached);
                    const ageOk = Date.now() - (data.timestamp || 0) < ttlMs;
                    const versionOk = !cacheVersion || data.cache_version === cacheVersion;
                    if (ageOk && versionOk && data.temp != null) {
                        weatherData = {
                            temp: data.temp,
                            tempMin: data.temp_min,
                            desc: data.desc ?? "",
                            icon: data.icon ?? "02d",
                            source: data.source ?? "",
                            timestamp: data.timestamp,
                            humidity: data.humidity ?? null,
                            windKph: data.wind_kph ?? null,
                            precipMm: data.precip_mm ?? null,
                        };
                        loading = false;
                        return;
                    }
                } catch (e) {}
            }
        }

        try {
            const data = await apiFetch(
                `/api/weather?slug=${encodeURIComponent(settlementSlug)}`,
            );

            const ts = data.fetched_at ? data.fetched_at * 1000 : Date.now();
            weatherData = {
                temp: data.temp ?? Math.round(data.main?.temp ?? 0),
                tempMin: data.temp_min != null ? data.temp_min : (data.main?.temp_min != null ? Math.round(data.main.temp_min) : null),
                desc: data.desc ?? data.weather?.[0]?.description ?? "",
                icon: data.icon ?? data.weather?.[0]?.icon ?? "02d",
                source: data.source ?? "",
                timestamp: ts,
                humidity: data.humidity ?? null,
                windKph: data.wind_kph != null ? Math.round(data.wind_kph) : null,
                precipMm: data.precip_mm ?? null,
            };

            if (browser) {
                localStorage.setItem(
                    cacheKey,
                    JSON.stringify({
                        temp: weatherData.temp,
                        temp_min: weatherData.tempMin,
                        desc: weatherData.desc,
                        icon: weatherData.icon,
                        source: weatherData.source,
                        timestamp: weatherData.timestamp,
                        humidity: weatherData.humidity,
                        wind_kph: weatherData.windKph,
                        precip_mm: weatherData.precipMm,
                        cache_version: cacheVersion,
                    }),
                );
            }
        } catch (err) {
            error = true;
        } finally {
            loading = false;
        }
    }

    $: if (browser && settlementSlug) {
        fetchWeather();
    }
</script>

<div id="idojaras" class="weather-card {advanced ? 'complex' : 'simple'}">
    <h3 class="widget-title">Időjárás</h3>
    <div class="widget-content">
        <div class="weather-left">
            {#if error}
            <span>
                <p class="weather-error">Időjárás adat nem elérhető.</p>
            </span>
            {:else}
                <div class="weather-temp-row">
                    <span class="weather-temp">{weatherData ? weatherData.temp : '--'}</span><span class="weather-temp-unit">°C</span>
                    <span class="weather-temp-min">/ {weatherData?.tempMin != null ? weatherData.tempMin : '--'}°C</span>
                </div>

                {#if loading}
                    <span class="weather-desc capitalize">adat betöltés...</span>
                {/if}
                {#if weatherData}
                    <span class="weather-desc capitalize">
                        {weatherData.desc || 'nincs adat'}
                    </span>
                {/if}
                {#if advanced}
                    <span class="weather-details">
                        <span class="weather-detail">
                            <span class="weather-detail-label">Csapadék:</span>
                            <span class="weather-detail-value">{weatherData?.precipMm != null ? weatherData.precipMm + ' mm' : '...'}</span>
                        </span>
                        <span class="weather-detail">
                            <span class="weather-detail-label">Páratart.:</span>
                            <span class="weather-detail-value">{weatherData?.humidity != null ? weatherData.humidity + '%' : '...'}</span>
                        </span>
                        <span class="weather-detail">
                            <span class="weather-detail-label">Szél:</span>
                            <span class="weather-detail-value">{weatherData?.windKph != null ? weatherData.windKph + ' km/h' : '...'}</span>
                        </span>
                    </span>
                {/if}
            {/if}
        </div>
        <div class="weather-right">
            {#if loading}
                <div class="skeleton weather-skeleton-icon" aria-hidden="true"></div>
            {:else if weatherData}
                <span class="weather-icon">
                    <WeatherIcon code={weatherData.icon} style={weatherIconStyle} />
                </span>
            {/if}
        </div>
    </div>
    <div class="weather-footer">
        <small class="weather-timestamp">Utoljára frissítve: <span class="weather-timestamp-value">{weatherData ? formatTime(weatherData.timestamp) : '00:00'}</span></small>
        <small class="weather-source">Forrás: <span class="weather-source-value">{weatherData?.source ? weatherData.source : 'adat betöltés...'}</span></small>
    </div>
</div>

<style>
    .capitalize {
        text-transform: capitalize;
    }
    .weather-error {
        color: var(--text-faint);
        margin: 0.5rem 0 0;
    }
    .weather-skeleton-icon {
        width: 5rem;
        height: 5rem;
        min-width: 5rem;
        min-height: 5rem;
        border-radius: 50%;
    }
    .weather-details {
        display: flex;
        flex-direction: column;
        gap: 0.15rem;
    }
    .weather-detail {
        font-size: 0.75rem;
        color: var(--text-faint);
        white-space: nowrap;
    }

    .widget-content{
        display: grid;
        grid-template-columns: 1fr auto;
        grid-template-areas:
            "left right"
            "footer footer";
        column-gap: 0.5rem;
        justify-items: start;
    }
</style>
