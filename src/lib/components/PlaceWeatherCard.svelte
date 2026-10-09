<script>
    import { browser } from "$app/environment";
    import WeatherSymbol from "$lib/icons/weather/WeatherSymbol.svelte";
    import WeatherCredit from "$lib/components/WeatherCredit.svelte";
    import { loadPlaceForecast } from "$lib/placeForecast.js";
    import { compassHU, uvLevel } from "$lib/weatherSymbols.js";
    import { deg, fmtClock, dayNameShort } from "$lib/weatherFormat.js";

    /**
     * The weather card on a settlement's or attraction's page (UI_BASELINE
     * "szf-pages"): now (symbol, temperature, description, feels-like), wind,
     * humidity, rain and UV, sunrise and sunset, and the next three days, with
     * a link to the place's /idojaras page. Same-size placeholders while it
     * loads (ld-reserve-space).
     *
     * @type {{ slug: string }}
     */
    let { slug } = $props();

    /** @type {any} */
    let data = $state(null);
    let loading = $state(true);

    $effect(() => {
        if (!browser || !slug) return;
        const s = slug;
        loading = true;
        loadPlaceForecast(s).then((d) => {
            if (s !== slug) return;
            data = d;
            loading = false;
        });
    });

    const cur = $derived(data?.current ?? null);
    const next = $derived((data?.daily ?? []).slice(1, 4));
    const astro = $derived(data?.astro ?? null);
</script>

<div id="idojaras" class="widget place-weather" aria-busy={loading}>
    <div class="widget-header">
        <h3 class="widget-title"><a href="/idojaras/{slug}">Időjárás · előrejelzés ›</a></h3>
    </div>
    {#if loading}
        <!-- The loaded card's own markup with dashes, dimmed, so it has the
             loaded card's height (the credit is the real one, invisible). -->
        <div class="place-weather__placeholder" aria-hidden="true">
            <div class="place-weather__now">
                <span class="skeleton place-weather__symbol-slot"></span>
                <span class="place-weather__now-text">
                    <span class="place-weather__temp">–°</span>
                    <span class="place-weather__desc">&nbsp;</span>
                    <span class="place-weather__feels">Hőérzet –°</span>
                </span>
            </div>
            <dl class="place-weather__facts">
                {#each ["Szél", "Pára", "Csapadék", "UV", "Napkelte", "Napnyugta"] as label (label)}
                    <div><dt>{label}</dt><dd>–</dd></div>
                {/each}
            </dl>
            <ul class="place-weather__days">
                {#each [1, 2, 3] as i (i)}
                    <li>
                        <span class="place-weather__day">&nbsp;</span>
                        <span class="skeleton place-weather__day-slot"></span>
                        <span class="place-weather__range"><b>–°</b> –°</span>
                    </li>
                {/each}
            </ul>
            <div class="place-weather__credit-slot"><WeatherCredit source="metno" fetchedAt={Date.now()} /></div>
        </div>
    {:else if !cur}
        <p class="place-weather__none">Az időjárás most nem érhető el.</p>
    {:else}
        <div class="place-weather__now">
            <WeatherSymbol symbol={cur.symbol} size={72} label={cur.desc} />
            <span class="place-weather__now-text">
                <span class="place-weather__temp">{deg(cur.temp)}</span>
                <span class="place-weather__desc">{cur.desc}</span>
                {#if cur.feels != null}<span class="place-weather__feels">Hőérzet {deg(cur.feels)}</span>{/if}
            </span>
        </div>
        <dl class="place-weather__facts">
            {#if cur.wind_kph != null}<div><dt>Szél</dt><dd>{Math.round(cur.wind_kph)} km/h {compassHU(cur.wind_dir)}</dd></div>{/if}
            {#if cur.humidity != null}<div><dt>Pára</dt><dd>{Math.round(cur.humidity)}%</dd></div>{/if}
            {#if cur.precip_mm != null}<div><dt>Csapadék</dt><dd>{cur.precip_mm} mm</dd></div>{/if}
            {#if cur.uv != null}<div><dt>UV</dt><dd>{Math.round(cur.uv)} · {uvLevel(cur.uv).label}</dd></div>{/if}
            {#if astro?.sunrise}<div><dt>Napkelte</dt><dd>{fmtClock(astro.sunrise)}</dd></div>{/if}
            {#if astro?.sunset}<div><dt>Napnyugta</dt><dd>{fmtClock(astro.sunset)}</dd></div>{/if}
        </dl>
        {#if next.length}
            <ul class="place-weather__days" aria-label="A következő napok">
                {#each next as d (d.date)}
                    <li>
                        <span class="place-weather__day">{dayNameShort(d.date)}</span>
                        <WeatherSymbol symbol={d.symbol} size={32} animated={false} label={d.desc} />
                        <span class="place-weather__range"><b>{deg(d.tmax)}</b> {deg(d.tmin)}</span>
                    </li>
                {/each}
            </ul>
        {/if}
        <WeatherCredit source={data.source} fetchedAt={data.fetched_at} />
    {/if}
</div>

<style>
    .place-weather {
        gap: 0.9rem;
    }
    .place-weather .widget-title a {
        color: inherit;
        text-decoration: none;
    }
    .place-weather .widget-title a:hover {
        color: var(--szekely-red);
    }
    .place-weather__now {
        display: flex;
        align-items: center;
        gap: 0.9rem;
        min-height: 4.5rem;
    }
    .place-weather__now-text {
        display: flex;
        flex-direction: column;
        gap: 0.15rem;
        min-width: 0;
        flex: 1;
    }
    .place-weather__temp {
        font-size: var(--text-2xl);
        font-weight: 700;
        line-height: 1;
    }
    .place-weather__desc {
        text-transform: lowercase;
    }
    .place-weather__feels {
        font-size: var(--text-sm);
        color: var(--text-muted);
    }
    .place-weather__facts {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: 0.35rem 1rem;
        margin: 0;
        font-size: var(--text-sm);
        min-height: 4.2rem;
    }
    .place-weather__facts div {
        display: flex;
        justify-content: space-between;
        gap: 0.5rem;
    }
    .place-weather__facts dt {
        color: var(--text-muted);
    }
    .place-weather__facts dd {
        margin: 0;
        font-weight: 600;
        white-space: nowrap;
    }
    .place-weather__days {
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        gap: 0.5rem;
        margin: 0;
        padding: 0.6rem 0 0;
        list-style: none;
        border-top: 1px solid var(--border-color);
    }
    .place-weather__days li {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 0.2rem;
        font-size: var(--text-sm);
    }
    .place-weather__day {
        text-transform: capitalize;
        color: var(--text-muted);
    }
    .place-weather__range b {
        font-weight: 700;
    }
    .place-weather__range {
        color: var(--text-muted);
        white-space: nowrap;
    }
    .place-weather__range b {
        color: var(--text-primary);
    }
    .place-weather__none {
        margin: 0;
        color: var(--text-muted);
    }
    .place-weather__placeholder {
        display: flex;
        flex-direction: column;
        gap: 0.9rem;
        opacity: 0.4;
    }
    .place-weather__symbol-slot {
        width: 72px;
        height: 72px;
        flex: none;
        border-radius: 50%;
    }
    .place-weather__day-slot {
        width: 32px;
        height: 32px;
        border-radius: 50%;
    }
    .place-weather__credit-slot {
        visibility: hidden;
    }
</style>
