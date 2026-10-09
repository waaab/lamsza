<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import PageHeader from "$lib/components/PageHeader.svelte";
    import HexMap from "$lib/components/szekelyfold/HexMap.svelte";
    import RegionNav from "$lib/components/szekelyfold/RegionNav.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";
    import { SEAT_ORDER, inRegionOrder } from "$lib/szekelyfold.js";

    let pageHeader = $state(initialPageHeader("szekek"));

    /** @type {Array<{ name: string, href: string, color: string }>} */
    let seats = $state([]);
    let loading = $state(true);

    onMount(async () => {
        pageHeader = await loadPageMeta("szekek");
        try {
            const rows = await apiFetch("/api/historical_seats");
            seats = inRegionOrder(Array.isArray(rows) ? rows : [], SEAT_ORDER).map((s) => ({
                name: s.name,
                href: `/szekek/${s.slug}`,
                color: s.color,
            }));
        } catch (e) {
            console.error(e);
            seats = [];
        } finally {
            loading = false;
        }
    });
</script>

<PageHeader
    landscape
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    breadcrumbLabel="Történelmi székek"
    documentTitleSuffix=" - Lámsza Index"
/>

<div class="region-top">
    {#if !loading && seats.length === 0}
        <div class="info-box"><p>Nincs megjeleníthető adat.</p></div>
    {:else}
        <HexMap items={seats} label="Székelyföld történelmi székei (sematikus ábra)" placeholders={SEAT_ORDER.length} />
    {/if}
    <RegionNav current="szekek" />
</div>
