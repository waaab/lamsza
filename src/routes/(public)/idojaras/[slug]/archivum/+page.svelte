<!--
    /idojaras/<settlement>/archivum: the weather of past days, from launch
    day on. Each day is folded from MET Norway's hourly values (not a station
    reading); a day opens to its hours.
-->
<script>
    import { page } from "$app/stores";
    import { apiCall, canReachApi } from "$lib/api";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import WeatherSymbol from "$lib/icons/weather/WeatherSymbol.svelte";
    import WeatherCredit from "$lib/components/WeatherCredit.svelte";
    import { deg, dayName, monthDay, hourLabel } from "$lib/weatherFormat.js";
    import { compassHU } from "$lib/weatherSymbols.js";
    import { formatHuDateLongFromYMD } from "$lib/utils";

    const RANGES = [7, 30, 90, 365];

    /** @type {any} */
    let data = $state(null);
    let range = $state(30);
    let status = $state(/** @type {"loading" | "ok" | "notfound" | "error"} */ ("loading"));
    /** @type {string | null} */
    let openDay = $state(null);
    /** @type {any[]} */
    let openHours = $state([]);
    let hoursLoading = $state(false);

    const slug = $derived($page.params.slug);

    $effect(() => {
        const s = slug;
        const n = range;
        if (!canReachApi() || !s) return;
        let live = true;
        status = "loading";
        apiCall(`/api/weather/archive?slug=${encodeURIComponent(s)}&days=${n}`)
            .then(async (res) => {
                if (!live) return;
                if (res.ok) {
                    data = await res.json();
                    status = "ok";
                } else {
                    status = res.status === 404 ? "notfound" : "error";
                }
            })
            .catch(() => {
                if (live) status = "error";
            });
        return () => {
            live = false;
        };
    });

    /** @param {string} date */
    async function toggleDay(date) {
        if (openDay === date) {
            openDay = null;
            return;
        }
        openDay = date;
        openHours = [];
        hoursLoading = true;
        try {
            const res = await apiCall(`/api/weather/archive?slug=${encodeURIComponent(slug)}&date=${date}`);
            if (res.ok && openDay === date) openHours = (await res.json()).hours ?? [];
        } finally {
            hoursLoading = false;
        }
    }

    const days = $derived(data?.days ?? []);
    const place = $derived(data?.place ?? null);

    // Chart: oldest left. Temperatures as two lines, rain as bars.
    const W = 640;
    const H = 180;
    const PAD = 24;
    const chart = $derived.by(() => {
        const pts = [...days].reverse();
        if (pts.length < 2) return null;
        const temps = pts.flatMap((/** @type {any} */ d) => [d.tmin, d.tmax]).filter((/** @type {any} */ v) => v != null);
        const lo = Math.floor(Math.min(...temps) - 1);
        const hi = Math.ceil(Math.max(...temps) + 1);
        const rainMax = Math.max(5, ...pts.map((/** @type {any} */ d) => d.precip_mm || 0));
        const x = (/** @type {number} */ i) => PAD + (i * (W - 2 * PAD)) / (pts.length - 1);
        const y = (/** @type {number} */ v) => PAD + ((hi - v) * (H - 2 * PAD)) / (hi - lo);
        const line = (/** @type {string} */ key) =>
            pts
                .map((/** @type {any} */ d, /** @type {number} */ i) => (d[key] == null ? null : `${x(i).toFixed(1)},${y(d[key]).toFixed(1)}`))
                .filter(Boolean)
                .join(" ");
        const barW = Math.max(2, (W - 2 * PAD) / pts.length - 2);
        return {
            max: line("tmax"),
            min: line("tmin"),
            bars: pts.map((/** @type {any} */ d, /** @type {number} */ i) => {
                const h = ((d.precip_mm || 0) / rainMax) * (H - 2 * PAD) * 0.5;
                return { x: x(i) - barW / 2, y: H - PAD - h, w: barW, h };
            }),
            lo,
            hi,
            zeroY: lo < 0 && hi > 0 ? y(0) : null,
        };
    });
</script>

<PublicPageHero
    title={place ? `${place.name}: időjárás-archívum` : "Időjárás-archívum"}
    greeting={data?.since ? `Az archívum ${formatHuDateLongFromYMD(data.since)} óta gyűlik, naponta egy sorral.` : "Az archívum az indulás napjától gyűlik."}
    breadcrumbLabel="Archívum"
    breadcrumbParentLabel="Időjárás"
    breadcrumbParentUrl="/idojaras"
    breadcrumbExtraLabel={place?.name ?? ""}
    breadcrumbExtraUrl={place ? `/idojaras/${slug}` : ""}
/>

<div class="wx-ranges" role="group" aria-label="Időszak">
    {#each RANGES as n (n)}
        <button type="button" class="btn nav-btn" class:active={range === n} aria-pressed={range === n} onclick={() => (range = n)}>
            {n === 365 ? "1 év" : `${n} nap`}
        </button>
    {/each}
</div>

{#if status === "loading" && !data}
    <div class="wx-chart-skeleton" aria-hidden="true"></div>
{:else if status === "notfound"}
    <div class="info-box"><p>Nincs ilyen település. <a href="/idojaras">Vissza az időjáráshoz</a></p></div>
{:else if status === "error"}
    <div class="info-box"><p>Az archívum most nem érhető el. Próbáld újra később.</p></div>
{:else if days.length === 0}
    <div class="info-box">
        <p>Ebben az időszakban még nincs archivált nap. Az archívum az indulás napján kezdődött; minden éjfél után egy nappal bővül.</p>
    </div>
{:else}
    {#if chart}
        <figure class="wx-chart">
            <svg viewBox="0 0 {W} {H}" role="img" aria-label="Napi legmagasabb és legalacsonyabb hőmérséklet és csapadék">
                {#if chart.zeroY != null}
                    <line x1={PAD} x2={W - PAD} y1={chart.zeroY} y2={chart.zeroY} class="wx-zero" />
                {/if}
                {#each chart.bars as b, i (i)}
                    <rect x={b.x} y={b.y} width={b.w} height={b.h} class="wx-rainbar" />
                {/each}
                <polyline points={chart.max} class="wx-line-max" />
                <polyline points={chart.min} class="wx-line-min" />
                <text x="4" y={PAD} class="wx-axis">{chart.hi}°</text>
                <text x="4" y={H - PAD} class="wx-axis">{chart.lo}°</text>
            </svg>
            <figcaption>
                <span class="wx-key wx-key-max">napi max.</span>
                <span class="wx-key wx-key-min">napi min.</span>
                <span class="wx-key wx-key-rain">csapadék</span>
            </figcaption>
        </figure>
    {/if}

    <ul class="wx-archive">
        {#each days as d (d.date)}
            <li>
                <button type="button" class="wx-arch-row" aria-expanded={openDay === d.date} onclick={() => toggleDay(d.date)}>
                    <span class="wx-arch-date"><strong>{dayName(d.date)}</strong> <small>{monthDay(d.date)}</small></span>
                    <WeatherSymbol symbol={d.symbol} size={36} animated={false} />
                    <span class="wx-arch-desc">{d.desc}</span>
                    <span class="wx-arch-temps">{deg(d.tmax)} / {deg(d.tmin)}</span>
                    <span class="wx-arch-rain">{d.precip_mm} mm</span>
                </button>
                {#if openDay === d.date}
                    <div class="wx-arch-hours">
                        {#if hoursLoading}
                            <p class="wx-faint">Betöltés…</p>
                        {:else if openHours.length === 0}
                            <p class="wx-faint">Erről a napról nincs óránkénti adat.</p>
                        {:else}
                            {#each openHours as h (h.time)}
                                <span class="wx-arch-hour">
                                    <span class="wx-faint">{hourLabel(h.time)}</span>
                                    <WeatherSymbol symbol={h.symbol} size={28} animated={false} label={h.desc} />
                                    <span>{deg(h.temp)}</span>
                                    {#if h.wind_kph != null}<span class="wx-faint">{Math.round(h.wind_kph)} {compassHU(h.wind_dir)}</span>{/if}
                                </span>
                            {/each}
                        {/if}
                    </div>
                {/if}
            </li>
        {/each}
    </ul>
    <p class="wx-faint wx-note">A napi értékek a MET Norway előrejelző modelljének óránkénti értékeiből készülnek, nem mérőállomás adatai.</p>
    <WeatherCredit source="metno" />
{/if}

<style>
    .wx-ranges {
        display: flex;
        flex-wrap: wrap;
        gap: 0.4rem;
        margin: 1rem 0;
    }
    .wx-chart-skeleton {
        height: 12rem;
        border-radius: 12px;
        background: var(--skeleton-bg);
    }
    .wx-chart {
        margin: 0 0 1rem;
    }
    .wx-chart svg {
        width: 100%;
        height: auto;
        display: block;
    }
    .wx-line-max,
    .wx-line-min {
        fill: none;
        stroke-width: 2.5;
        stroke-linejoin: round;
        stroke-linecap: round;
    }
    .wx-line-max {
        stroke: var(--warm-light);
    }
    .wx-line-min {
        stroke: var(--szekely-blue);
    }
    .wx-rainbar {
        fill: color-mix(in srgb, var(--szekely-blue) 30%, transparent);
    }
    .wx-zero {
        stroke: var(--border-color);
        stroke-dasharray: 3 4;
    }
    .wx-axis {
        font-size: 11px;
        fill: var(--text-faint);
    }
    figcaption {
        display: flex;
        gap: 1rem;
        font-size: 0.8rem;
        color: var(--text-faint);
    }
    .wx-key::before {
        content: "";
        display: inline-block;
        width: 0.9rem;
        height: 0.25rem;
        border-radius: 2px;
        margin-right: 0.3rem;
        vertical-align: middle;
    }
    .wx-key-max::before {
        background: var(--warm-light);
    }
    .wx-key-min::before {
        background: var(--szekely-blue);
    }
    .wx-key-rain::before {
        background: color-mix(in srgb, var(--szekely-blue) 30%, transparent);
        height: 0.6rem;
    }
    .wx-archive {
        list-style: none;
        margin: 0;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
    }
    .wx-arch-row {
        width: 100%;
        display: grid;
        grid-template-columns: 8rem 36px 1fr auto 4.5rem;
        align-items: center;
        gap: 0.75rem;
        padding: 0.45rem 0.75rem;
        border-radius: 10px;
        border: 1px solid var(--border-color);
        background: var(--card-bg);
        color: inherit;
        font: inherit;
        text-align: left;
        cursor: pointer;
    }
    .wx-arch-row:hover {
        box-shadow: 0 2px 6px var(--shadow-md);
    }
    @media (max-width: 640px) {
        .wx-arch-row {
            grid-template-columns: 6rem 36px 1fr auto;
        }
        .wx-arch-desc {
            display: none;
        }
    }
    .wx-arch-date small,
    .wx-faint {
        color: var(--text-faint);
    }
    .wx-arch-desc::first-letter {
        text-transform: uppercase;
    }
    .wx-arch-temps {
        font-variant-numeric: tabular-nums;
        font-weight: 600;
    }
    .wx-arch-rain {
        color: var(--szekely-blue);
        text-align: right;
        font-size: 0.85rem;
    }
    .wx-arch-hours {
        display: flex;
        flex-wrap: wrap;
        gap: 0.4rem;
        padding: 0.5rem 0.75rem 0.75rem;
    }
    .wx-arch-hour {
        display: inline-flex;
        flex-direction: column;
        align-items: center;
        gap: 0.1rem;
        min-width: 3.2rem;
        font-size: 0.8rem;
    }
    .wx-note {
        font-size: 0.8rem;
        margin-top: 1rem;
    }
</style>
