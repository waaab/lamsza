<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import PageHeader from "$lib/components/PageHeader.svelte";
    import RegionNav from "$lib/components/szekelyfold/RegionNav.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";

    /** Placeholder tiles while the list loads (UI_BASELINE "ld-reserve-space"). */
    const PLACEHOLDERS = 12;

    let pageHeader = $state(initialPageHeader("varosok"));

    /** @type {Array<{ name: string, slug: string, county_slug: string, type: string }>} */
    let towns = $state([]);
    let loading = $state(true);

    onMount(async () => {
        pageHeader = await loadPageMeta("varosok");
        try {
            const all = await apiFetch("/api/locations");
            towns = (Array.isArray(all) ? all : [])
                .filter((l) => ["város", "municípium"].includes(l.type?.toLowerCase() ?? ""))
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
    breadcrumbLabel="Székelyföldi Városok"
    documentTitleSuffix=" - Lámsza Index"
/>

{#if loading}
    <ul class="region-tiles" aria-busy="true">
        {#each { length: PLACEHOLDERS } as _, i (i)}
            <li aria-hidden="true"><span class="card region-tile region-tile--placeholder"><span>&nbsp;</span></span></li>
        {/each}
    </ul>
{:else if towns.length === 0}
    <div class="info-box"><p>Nincs megjeleníthető adat.</p></div>
{:else}
    <ul class="region-tiles">
        {#each towns as t (t.slug)}
            <li>
                <a class="card region-tile" href="/{t.county_slug}-megye/{t.slug}">
                    <span>{t.name}</span>
                    <span class="region-tile__tag" class:region-tile__tag--strong={t.type?.toLowerCase() === "municípium"}>{t.type}</span>
                </a>
            </li>
        {/each}
    </ul>
{/if}

<RegionNav current="varosok" row />
