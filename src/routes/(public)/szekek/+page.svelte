<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";

    let pageHeader = initialPageHeader("szekek");
    let pageHeaderLoading = false;

    let seats = [];
    let loading = true;

    onMount(async () => {
        pageHeader = await loadPageMeta("szekek");
        pageHeaderLoading = false;
        try {
            seats = await apiFetch("/api/historical_seats");
            if (!Array.isArray(seats)) seats = [];
            seats = seats.sort((a, b) => a.name.localeCompare(b.name));
        } catch (e) {
            console.error(e);
            seats = [];
        } finally {
            loading = false;
        }
    });
</script>

<PublicPageHero
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    loading={pageHeaderLoading}
    breadcrumbLabel="Történelmi székek"
    documentTitleSuffix=" - Lámsza Index"
/>

<div class="page-inner">
    {#if loading}
        <div class="info-box"><p>Betöltés…</p></div>
    {:else if seats.length === 0}
        <div class="info-box"><p>Nincs megjeleníthető adat.</p></div>
    {:else}
        {#each seats as seat}
            <a href="/szekek/{seat.slug}" class="card sm seat">
                <span class="seat-name">{seat.name}</span>
            </a>
        {/each}
    {/if}
</div>
<nav class="page-nav">
    <h4 class="page-nav-title">Oldal navigáció</h4>
    <ul>
        <li><a class="btn nav-btn" href="/megyek">Székelyföldi megyék</a></li>
        <li><a class="btn nav-btn" href="/varosok">Székelyföldi városok</a></li>
        <li><a class="btn nav-btn" href="/falvak">Székelyföldi falvak</a></li>
    </ul>
</nav>