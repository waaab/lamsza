<script>
    import { onMount } from "svelte";
    import { browser } from "$app/environment";
    import { page } from "$app/stores";
    import { apiFetch } from "$lib/api";
    import PageHeader from "$lib/components/PageHeader.svelte";
    import RegionNav from "$lib/components/szekelyfold/RegionNav.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";
    import { COUNTY_ORDER, countyColor, inRegionOrder } from "$lib/szekelyfold.js";

    /** Placeholder tiles while the list loads (UI_BASELINE "ld-reserve-space"). */
    const PLACEHOLDERS = 16;

    let pageHeader = $state(initialPageHeader("falvak"));

    /** @type {Array<{ name: string, slug: string, county: string, county_slug: string }>} */
    let villages = $state([]);
    let loading = $state(true);

    // The county filter lives in the address (?megye=hargita), so a filtered
    // list can be shared. The page is prerendered, so it is read in the
    // browser only.
    const megye = $derived(browser ? ($page.url.searchParams.get("megye") ?? "") : "");

    /** The counties that have villages, in the map's order and colours. */
    const counties = $derived(
        inRegionOrder(
            [...new Map(villages.map((v) => [v.county_slug, { slug: v.county_slug, name: v.county }])).values()],
            COUNTY_ORDER,
        ),
    );
    const shown = $derived(megye ? villages.filter((v) => v.county_slug === megye) : villages);

    onMount(async () => {
        pageHeader = await loadPageMeta("falvak");
        try {
            const all = await apiFetch("/api/locations");
            villages = (Array.isArray(all) ? all : [])
                .filter((l) => ["falu", "község"].includes(l.type?.toLowerCase() ?? ""))
                .sort((a, b) => a.name.localeCompare(b.name, "hu"));
        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    });
</script>

<PageHeader
    landscape
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    breadcrumbLabel="Székelyföldi Falvak"
    documentTitleSuffix=" - Lámsza Index"
/>

<nav class="region-chips" aria-label="Megye szerint">
    <a class="btn btn-sm region-chip" href="/falvak" aria-current={megye === "" ? "true" : undefined} data-sveltekit-noscroll data-sveltekit-replacestate>Mind</a>
    {#if loading}
        {#each COUNTY_ORDER as [slug, color] (slug)}
            <span class="btn btn-sm region-chip region-tile--placeholder" style:--dot={color} aria-hidden="true"><span class="region-dot"></span>••••••</span>
        {/each}
    {:else}
        {#each counties as c (c.slug)}
            <a
                class="btn btn-sm region-chip"
                href="/falvak?megye={c.slug}"
                style:--dot={c.color}
                aria-current={megye === c.slug ? "true" : undefined}
                data-sveltekit-noscroll
                data-sveltekit-replacestate
            ><span class="region-dot"></span>{c.name}</a>
        {/each}
    {/if}
</nav>

{#if loading}
    <ul class="region-tiles" aria-busy="true">
        {#each { length: PLACEHOLDERS } as _, i (i)}
            <li aria-hidden="true"><span class="card region-tile region-tile--placeholder"><span>&nbsp;</span></span></li>
        {/each}
    </ul>
{:else if shown.length === 0}
    <div class="info-box"><p>Nincs megjeleníthető adat.</p></div>
{:else}
    <ul class="region-tiles">
        {#each shown as v (v.slug)}
            <li>
                <a class="card region-tile" href="/{v.county_slug}-megye/{v.slug}" style:--tile-color={countyColor(v.county_slug)}>
                    <span>{v.name}</span>
                    <span class="region-tile__tag" style:--dot={countyColor(v.county_slug)}><span class="region-dot"></span>{v.county}</span>
                </a>
            </li>
        {/each}
    </ul>
{/if}

<RegionNav current="falvak" row />
