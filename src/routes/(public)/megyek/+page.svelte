<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";

    let pageHeader = initialPageHeader("megyek");
    let pageHeaderLoading = false;

    let locations = [];
    let loading = true;

    onMount(async () => {
        pageHeader = await loadPageMeta("megyek");
        pageHeaderLoading = false;
        try {
            const all = await apiFetch("/api/locations?type=megye");
            locations = all.sort((a, b) => a.name.localeCompare(b.name));
        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    });
</script>

<PublicPageHero
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    loading={pageHeaderLoading}
    breadcrumbLabel="Székelyföldi Megyék"
    documentTitleSuffix=" - Lámsza Index"
/>

<div class="page-inner">
    {#if loading}
        <div class="info-box"><p>Betöltés…</p></div>
    {:else if locations.length === 0}
        <div class="info-box"><p>Nincs megjeleníthető adat.</p></div>
    {:else}
        {#each locations as loc}
            <a href="/{loc.slug}-megye" class="card sm county">
                <span class="location-name">{loc.name}</span>
                <span class="location-county">{loc.type}</span>
            </a>
        {/each}
    {/if}
</div>
<nav class="page-nav">
    <h4 class="page-nav-title">Oldal navigáció</h4>
    <ul>
        <li><a class="btn nav-btn" href="/szekek">Történelmi székek</a></li>
        <li><a class="btn nav-btn" href="/varosok">Székelyföldi városok</a></li>
        <li><a class="btn nav-btn" href="/falvak">Székelyföldi falvak</a></li>
    </ul>
</nav>

<style>
    .location-county {
        color: var(--text-faint);
    }
</style>