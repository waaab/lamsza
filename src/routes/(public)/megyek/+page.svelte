<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import PageHeader from "$lib/components/PageHeader.svelte";
    import HexMap from "$lib/components/szekelyfold/HexMap.svelte";
    import RegionNav from "$lib/components/szekelyfold/RegionNav.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";
    import { COUNTY_ORDER, inRegionOrder } from "$lib/szekelyfold.js";

    let pageHeader = $state(initialPageHeader("megyek"));

    /** @type {Array<{ name: string, href: string, color: string, sub: string }>} */
    let counties = $state([]);
    let loading = $state(true);

    onMount(async () => {
        pageHeader = await loadPageMeta("megyek");
        try {
            const rows = await apiFetch("/api/locations?type=megye");
            counties = inRegionOrder(Array.isArray(rows) ? rows : [], COUNTY_ORDER).map((c) => ({
                name: c.name,
                href: `/${c.slug}-megye`,
                color: c.color,
                sub: "megye",
            }));
        } catch (e) {
            console.error(e);
            counties = [];
        } finally {
            loading = false;
        }
    });
</script>

<PageHeader
    landscape
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    breadcrumbLabel="Székelyföldi Megyék"
    documentTitleSuffix=" - Lámsza Index"
/>

<div class="region-top">
    {#if !loading && counties.length === 0}
        <div class="info-box"><p>Nincs megjeleníthető adat.</p></div>
    {:else}
        <HexMap items={counties} label="Székelyföldi megyék (sematikus ábra)" placeholders={COUNTY_ORDER.length} />
    {/if}
    <RegionNav current="megyek" />
</div>
