<!--
    /idojaras/<settlement>: the weather now, the next 48 hours, the days
    ahead (MET Norway reaches about nine), and today's sun and moon. All of it
    comes from the backend's cache (GET /api/weather/forecast), which the
    weather worker keeps fresh; this page never asks a provider itself.
-->
<script>
    import { page } from "$app/stores";
    import { apiCall, canReachApi } from "$lib/api";
    import PageHeader from "$lib/components/PageHeader.svelte";
    import WeatherSymbol from "$lib/icons/weather/WeatherSymbol.svelte";
    import WeatherGlyph from "$lib/icons/weather/WeatherGlyph.svelte";
    import WeatherCredit from "$lib/components/WeatherCredit.svelte";
    import { compassHU, uvLevel, moonPhaseName } from "$lib/weatherSymbols.js";
    import { deg, fmtClock, dayName, monthDay, duration } from "$lib/weatherFormat.js";
    import HourlyStrip from "$lib/components/HourlyStrip.svelte";

    /** @type {any} */
    let data = $state(null);
    let status = $state(/** @type {"loading" | "ok" | "notready" | "notfound" | "error"} */ ("loading"));

    const slug = $derived($page.params.slug);

    $effect(() => {
        const s = slug;
        if (!canReachApi() || !s) return;
        status = "loading";
        let live = true;
        apiCall(`/api/weather/forecast?slug=${encodeURIComponent(s)}`)
            .then(async (res) => {
                if (!live) return;
                if (res.ok) {
                    data = await res.json();
                    status = "ok";
                } else {
                    status = res.status === 404 ? "notfound" : res.status === 503 ? "notready" : "error";
                }
            })
            .catch(() => {
                if (live) status = "error";
            });
        return () => {
            live = false;
        };
    });

    const cur = $derived(data?.current ?? null);
    const days = $derived(data?.daily ?? []);
    const hours = $derived(data?.hourly ?? []);
    const astro = $derived(data?.astro ?? null);
    const place = $derived(data?.place ?? null);
    const title = $derived(place ? `${place.name} időjárása` : "Időjárás");
    const isFallback = $derived(data && data.source !== "metno");
    // Bar scale for the days' ranges: the week's lowest low to highest high.
    const span = $derived.by(() => {
        const lows = days.map((/** @type {any} */ d) => d.tmin).filter((/** @type {any} */ v) => v != null);
        const highs = days.map((/** @type {any} */ d) => d.tmax).filter((/** @type {any} */ v) => v != null);
        const lo = Math.min(...lows);
        const hi = Math.max(...highs);
        return Number.isFinite(lo) && Number.isFinite(hi) && hi > lo ? { lo, hi } : { lo: 0, hi: 1 };
    });
    const pct = (/** @type {number} */ v) => ((v - span.lo) / (span.hi - span.lo)) * 100;
    const uv = $derived(uvLevel(cur?.uv));
</script>

<PageHeader
    {title}
    greeting={place?.county ? `${place.county} megye · ${place.type}` : place?.kind === "attraction" ? "Látnivaló" : ""}
    breadcrumbLabel={place?.name ?? "…"}
    breadcrumbParentLabel="Időjárás"
    breadcrumbParentUrl="/idojaras"
/>

{#if status === "loading"}
    <div class="wx-now card wx-skeleton" aria-hidden="true"></div>
{:else if status === "notfound"}
    <div class="info-box"><p>Nincs ilyen település. <a href="/idojaras">Vissza az időjáráshoz</a></p></div>
{:else if status === "notready"}
    <div class="info-box"><p>Ennek a településnek az előrejelzése még készül. Nézz vissza pár perc múlva.</p></div>
{:else if status === "error" || !cur}
    <div class="info-box"><p>Az időjárás most nem érhető el. Próbáld újra később.</p></div>
{:else}
    {#if data.stale}
        <p class="note warn">Az időjárás-szolgáltatók egy ideje nem válaszolnak; ez a legutóbbi ismert előrejelzés.</p>
    {:else if isFallback}
        <p class="note">A fő forrásunk (MET Norway) most nem érhető el, ezért egy rövidebb, {days.length} napos előrejelzést mutatunk.</p>
    {/if}

    <section class="wx-now card" aria-label="Most">
        <div class="wx-now-main">
            <WeatherSymbol symbol={cur.symbol} size={128} />
            <div class="wx-now-text">
                <span class="wx-now-temp">{deg(cur.temp)}</span>
                <span class="wx-now-desc">{cur.desc}</span>
                {#if cur.feels != null}
                    <span class="wx-now-feels">Hőérzet {deg(cur.feels)}</span>
                {/if}
                {#if days[0]}
                    <span class="wx-now-feels">Ma max. {deg(days[0].tmax)}, min. {deg(days[0].tmin)}</span>
                {/if}
            </div>
        </div>
        <ul class="wx-facts">
            {#if cur.wind_kph != null}
                <li>
                    <WeatherGlyph kind="wind" value={cur.wind_dir} />
                    <span class="wx-fact-label">Szél</span>
                    <span class="wx-fact-value">{Math.round(cur.wind_kph)} km/h {compassHU(cur.wind_dir)}</span>
                </li>
            {/if}
            {#if cur.gust_kph != null}
                <li>
                    <WeatherGlyph kind="gust" />
                    <span class="wx-fact-label">Széllökés</span>
                    <span class="wx-fact-value">{Math.round(cur.gust_kph)} km/h</span>
                </li>
            {/if}
            {#if cur.humidity != null}
                <li>
                    <WeatherGlyph kind="humidity" value={cur.humidity} />
                    <span class="wx-fact-label">Páratartalom</span>
                    <span class="wx-fact-value">{Math.round(cur.humidity)}%</span>
                </li>
            {/if}
            {#if cur.precip_mm != null}
                <li>
                    <WeatherGlyph kind="precip" value={cur.precip_mm} />
                    <span class="wx-fact-label">Csapadék (1 óra)</span>
                    <span class="wx-fact-value">{cur.precip_mm} mm</span>
                </li>
            {/if}
            {#if cur.pressure_hpa != null}
                <li>
                    <WeatherGlyph kind="pressure" value={cur.pressure_hpa} />
                    <span class="wx-fact-label">Légnyomás</span>
                    <span class="wx-fact-value">{Math.round(cur.pressure_hpa)} hPa</span>
                </li>
            {/if}
            {#if cur.uv != null}
                <li>
                    <WeatherGlyph kind="uv" value={cur.uv} />
                    <span class="wx-fact-label">UV-index</span>
                    <span class="wx-fact-value">{Math.round(cur.uv)} · {uv.label}</span>
                </li>
            {/if}
            {#if cur.cloud_pct != null}
                <li>
                    <WeatherGlyph kind="cloud" value={cur.cloud_pct} />
                    <span class="wx-fact-label">Felhőzet</span>
                    <span class="wx-fact-value">{Math.round(cur.cloud_pct)}%</span>
                </li>
            {/if}
            {#if cur.dew_point != null}
                <li>
                    <WeatherGlyph kind="dewpoint" value={cur.humidity} />
                    <span class="wx-fact-label">Harmatpont</span>
                    <span class="wx-fact-value">{deg(cur.dew_point)}</span>
                </li>
            {/if}
            {#if cur.fog_pct != null && cur.fog_pct > 0}
                <li>
                    <WeatherGlyph kind="fog" />
                    <span class="wx-fact-label">Köd</span>
                    <span class="wx-fact-value">{Math.round(cur.fog_pct)}%</span>
                </li>
            {/if}
        </ul>
    </section>

    {#if astro}
        <section class="wx-section">
            <h3 class="widget-title">Nap és Hold ma</h3>
            <ul class="wx-astro">
                <li><WeatherGlyph kind="sunrise" size={36} /><span class="wx-fact-label">Napkelte</span><span class="wx-fact-value">{fmtClock(astro.sunrise) || "-"}</span></li>
                <li><WeatherGlyph kind="sunset" size={36} /><span class="wx-fact-label">Napnyugta</span><span class="wx-fact-value">{fmtClock(astro.sunset) || "-"}</span></li>
                <li><WeatherGlyph kind="daylength" size={36} /><span class="wx-fact-label">Nappal hossza</span><span class="wx-fact-value">{duration(astro.sunrise, astro.sunset) || "-"}</span></li>
                <li><WeatherGlyph kind="moonrise" size={36} /><span class="wx-fact-label">Holdkelte</span><span class="wx-fact-value">{fmtClock(astro.moonrise) || "-"}</span></li>
                <li><WeatherGlyph kind="moonset" size={36} /><span class="wx-fact-label">Holdnyugta</span><span class="wx-fact-value">{fmtClock(astro.moonset) || "-"}</span></li>
                {#if astro.moon_phase != null}
                    <li><WeatherGlyph kind="moonphase" value={astro.moon_phase} size={36} /><span class="wx-fact-label">Holdfázis</span><span class="wx-fact-value">{moonPhaseName(astro.moon_phase)}</span></li>
                {/if}
            </ul>
        </section>
    {/if}

    <section class="wx-section">
        <h3 class="widget-title">A következő 48 óra</h3>
        <HourlyStrip {hours} />
    </section>

    <section class="wx-section">
        <h3 class="widget-title">{days.length} napos előrejelzés</h3>
        <ul class="wx-days">
            {#each days as d (d.date)}
                <li class="wx-day">
                    <span class="wx-day-name">
                        <strong>{dayName(d.date)}</strong>
                        <small>{monthDay(d.date)}</small>
                    </span>
                    <WeatherSymbol symbol={d.symbol} size={48} />
                    <span class="wx-day-desc">{d.desc}</span>
                    <span class="wx-day-range" aria-label="max. {deg(d.tmax)}, min. {deg(d.tmin)}">
                        <span class="wx-day-min">{deg(d.tmin)}</span>
                        <span class="wx-bar" aria-hidden="true">
                            {#if d.tmin != null && d.tmax != null}
                                <span class="wx-bar-fill" style="left: {pct(d.tmin)}%; right: {100 - pct(d.tmax)}%"></span>
                            {/if}
                        </span>
                        <span class="wx-day-max">{deg(d.tmax)}</span>
                    </span>
                    <span class="wx-day-extra">
                        {#if d.precip_mm > 0}
                            <span><WeatherGlyph kind="precip" value={d.precip_mm} size={18} animated={false} /> {d.precip_mm} mm</span>
                        {/if}
                        {#if d.precip_prob != null}
                            <span><WeatherGlyph kind="precip_prob" size={18} animated={false} /> {Math.round(d.precip_prob)}%</span>
                        {/if}
                        {#if d.wind_max_kph != null}
                            <span><WeatherGlyph kind="wind" size={18} animated={false} /> {Math.round(d.wind_max_kph)} km/h</span>
                        {/if}
                    </span>
                </li>
            {/each}
        </ul>
    </section>

    <nav class="page-nav">
        <h4 class="page-nav-title">Oldal navigáció</h4>
        <ul>
            {#if place?.kind === "settlement"}
                <li><a class="btn nav-btn" href="/idojaras/{slug}/archivum">Időjárás-archívum</a></li>
            {/if}
            {#if place?.county_slug}
                <li><a class="btn nav-btn" href="/{place.county_slug}-megye/{slug}">{place.name} oldala</a></li>
            {/if}
            <li><a class="btn nav-btn" href="/idojaras">Minden település</a></li>
        </ul>
    </nav>

    <WeatherCredit source={data.source} fetchedAt={data.fetched_at} />
{/if}

<style>
    .wx-skeleton {
        min-height: 14rem;
        background: var(--skeleton-bg);
        border: none;
    }
    .wx-now {
        display: grid;
        grid-template-columns: minmax(16rem, 1fr) minmax(16rem, 1.4fr);
        gap: 1rem 2rem;
        margin: 1rem 0;
    }
    @media (max-width: 760px) {
        .wx-now {
            grid-template-columns: 1fr;
        }
    }
    .wx-now-main {
        display: flex;
        align-items: center;
        gap: 1rem;
    }
    .wx-now-text {
        display: flex;
        flex-direction: column;
        gap: 0.2rem;
    }
    .wx-now-temp {
        font-size: 3.5rem;
        line-height: 1;
        font-weight: 300;
    }
    .wx-now-desc {
        font-size: 1.1rem;
        color: var(--text-secondary);
    }
    .wx-now-desc::first-letter {
        text-transform: uppercase;
    }
    .wx-now-feels {
        color: var(--text-faint);
        font-size: 0.9rem;
    }
    .wx-facts,
    .wx-astro {
        list-style: none;
        margin: 0;
        padding: 0;
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(11rem, 1fr));
        gap: 0.5rem 1rem;
        align-content: center;
    }
    .wx-facts li,
    .wx-astro li {
        display: grid;
        grid-template-columns: auto 1fr;
        grid-template-rows: auto auto;
        column-gap: 0.6rem;
        align-items: center;
        color: var(--text-secondary);
    }
    .wx-facts li :global(svg),
    .wx-astro li :global(svg) {
        grid-row: span 2;
    }
    .wx-fact-label {
        font-size: 0.8rem;
        color: var(--text-faint);
    }
    .wx-fact-value {
        font-weight: 600;
        color: var(--text-primary);
    }
    .wx-section {
        margin-top: 1.75rem;
    }
    .wx-days {
        list-style: none;
        margin: 0.5rem 0 0;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 0.4rem;
    }
    .wx-day {
        display: grid;
        grid-template-columns: 7.5rem 48px minmax(7rem, 1fr) minmax(9rem, 14rem) minmax(0, auto);
        align-items: center;
        gap: 0.75rem;
        padding: 0.5rem 0.75rem;
        border-radius: 12px;
        background: var(--card-bg);
        border: 1px solid var(--border-color);
    }
    @media (max-width: 760px) {
        .wx-day {
            grid-template-columns: 5.5rem 48px 1fr;
        }
        .wx-day-range,
        .wx-day-extra {
            grid-column: 1 / -1;
        }
    }
    .wx-day-name {
        display: flex;
        flex-direction: column;
    }
    .wx-day-name small {
        color: var(--text-faint);
    }
    .wx-day-desc {
        color: var(--text-secondary);
    }
    .wx-day-desc::first-letter {
        text-transform: uppercase;
    }
    .wx-day-range {
        display: grid;
        grid-template-columns: 2.5rem 1fr 2.5rem;
        align-items: center;
        gap: 0.4rem;
        font-variant-numeric: tabular-nums;
    }
    .wx-day-min {
        color: var(--text-faint);
        text-align: right;
    }
    .wx-day-max {
        font-weight: 600;
    }
    .wx-bar {
        position: relative;
        height: 6px;
        border-radius: 3px;
        background: var(--border-color);
    }
    .wx-bar-fill {
        position: absolute;
        top: 0;
        bottom: 0;
        border-radius: 3px;
        background: linear-gradient(90deg, var(--szekely-blue), var(--warm-light));
    }
    .wx-day-extra {
        display: flex;
        flex-wrap: wrap;
        gap: 0.2rem 0.8rem;
        color: var(--text-faint);
        font-size: 0.8rem;
    }
    .wx-day-extra span {
        display: inline-flex;
        align-items: center;
        gap: 0.2rem;
    }
</style>
