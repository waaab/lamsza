<!--
    /idojaras: every settlement's weather now, by county, with the site's own
    place (`my_location_slug`) first. Each card opens that place's forecast.
-->
<script>
    import { onMount } from "svelte";
    import { apiFetch, canReachApi } from "$lib/api";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import WeatherSymbol from "$lib/icons/weather/WeatherSymbol.svelte";
    import WeatherCredit from "$lib/components/WeatherCredit.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";
    import { deg } from "$lib/weatherFormat.js";

    let pageHeader = $state(initialPageHeader("idojaras"));
    /** @type {Array<{ slug: string, name: string, county: string, county_slug: string, type: string, is_county_seat: boolean, now: any }>} */
    let places = $state([]);
    let homeSlug = $state("csikszereda");
    let loading = $state(true);
    let failed = $state(false);

    const home = $derived(places.find((p) => p.slug === homeSlug && p.now) ?? null);
    const byCounty = $derived.by(() => {
        /** @type {Map<string, { county: string, slug: string, items: typeof places }>} */
        const groups = new Map();
        for (const p of places) {
            if (!groups.has(p.county_slug)) groups.set(p.county_slug, { county: p.county, slug: p.county_slug, items: [] });
            groups.get(p.county_slug)?.items.push(p);
        }
        for (const g of groups.values()) {
            // Towns first, then the rest, by name.
            g.items.sort((a, b) => Number(b.is_county_seat) - Number(a.is_county_seat) || a.name.localeCompare(b.name, "hu"));
        }
        return [...groups.values()];
    });
    const sources = $derived([...new Set(places.map((p) => p.now?.source).filter(Boolean))]);
    const sourceIds = { "MET Norway": "metno", "WeatherAPI.com": "weatherapi_com", OpenWeather: "openweathermap" };

    onMount(async () => {
        if (!canReachApi()) return;
        loadPageMeta("idojaras").then((m) => (pageHeader = m));
        try {
            const [cfg, list] = await Promise.all([
                apiFetch("/api/config/public").catch(() => null),
                apiFetch("/api/weather/places"),
            ]);
            if (cfg?.my_location_slug) homeSlug = cfg.my_location_slug;
            places = Array.isArray(list) ? list : [];
        } catch {
            failed = true;
        } finally {
            loading = false;
        }
    });
</script>

<PublicPageHero
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    breadcrumbLabel="Időjárás"
    breadcrumbParentLabel=""
    breadcrumbParentUrl=""
/>

{#if loading}
    <div class="wx-home card skeleton-card" aria-hidden="true"></div>
{:else if failed}
    <div class="info-box"><p>Az időjárás most nem érhető el. Próbáld újra később.</p></div>
{:else}
    {#if home}
        <a class="wx-home card" href="/idojaras/{home.slug}">
            <WeatherSymbol symbol={home.now.symbol} size={112} />
            <div class="wx-home-text">
                <span class="wx-home-name">{home.name}</span>
                <span class="wx-home-temp">{deg(home.now.temp)}</span>
                <span class="wx-home-desc">{home.now.desc}</span>
                <span class="wx-home-range">
                    max. {deg(home.now.temp_max)} · min. {deg(home.now.temp_min)}
                </span>
            </div>
            <span class="wx-home-more">Részletes előrejelzés &rsaquo;</span>
        </a>
    {/if}

    {#each byCounty as group (group.slug)}
        <section class="wx-county">
            <h3 class="widget-title">{group.county} megye</h3>
            <div class="wx-grid">
                {#each group.items as p (p.slug)}
                    <a class="card sm wx-place" href="/idojaras/{p.slug}">
                        <span class="wx-place-name">{p.name}</span>
                        {#if p.now}
                            <span class="wx-place-row">
                                <WeatherSymbol symbol={p.now.symbol} size={44} />
                                <span class="wx-place-temp">{deg(p.now.temp)}</span>
                            </span>
                            <span class="wx-place-desc">{p.now.desc}</span>
                            <span class="wx-place-range">{deg(p.now.temp_max)} / {deg(p.now.temp_min)}</span>
                        {:else}
                            <span class="wx-place-desc">Hamarosan…</span>
                        {/if}
                    </a>
                {/each}
            </div>
        </section>
    {/each}

    {#each sources as src (src)}
        <WeatherCredit source={sourceIds[src] ?? "metno"} />
    {/each}
{/if}

<style>
    .wx-home {
        display: flex;
        align-items: center;
        gap: 1.25rem;
        flex-wrap: wrap;
        margin: 1rem 0 1.5rem;
        min-height: 9rem;
    }
    .skeleton-card {
        background: var(--skeleton-bg);
        border: none;
    }
    .wx-home-text {
        display: flex;
        flex-direction: column;
        gap: 0.15rem;
        flex: 1 1 12rem;
    }
    .wx-home-name {
        font-weight: 600;
        font-size: 1.1rem;
    }
    .wx-home-temp {
        font-size: 3rem;
        line-height: 1;
        font-weight: 300;
    }
    .wx-home-desc {
        color: var(--text-secondary);
        text-transform: capitalize;
    }
    .wx-home-range {
        color: var(--text-faint);
        font-size: 0.9rem;
    }
    .wx-home-more {
        align-self: flex-end;
        color: var(--text-secondary);
        font-size: 0.9rem;
    }
    .wx-county {
        margin-top: 1.5rem;
    }
    .wx-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(9.5rem, 1fr));
        gap: 0.6rem;
        margin-top: 0.6rem;
    }
    .wx-place {
        gap: 0.2rem;
        min-height: 8.5rem;
    }
    .wx-place-name {
        font-weight: 600;
    }
    .wx-place-row {
        display: flex;
        align-items: center;
        gap: 0.4rem;
    }
    .wx-place-temp {
        font-size: 1.6rem;
        font-weight: 300;
    }
    .wx-place-desc {
        color: var(--text-secondary);
        font-size: 0.85rem;
    }
    .wx-place-desc::first-letter {
        text-transform: uppercase;
    }
    .wx-place-range {
        color: var(--text-faint);
        font-size: 0.8rem;
    }
</style>
