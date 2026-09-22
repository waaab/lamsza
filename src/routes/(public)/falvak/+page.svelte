<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";

    let pageHeader = initialPageHeader("falvak");
    let pageHeaderLoading = false;

    let locations = [];
    let loading = true;

    onMount(async () => {
        pageHeader = await loadPageMeta("falvak");
        pageHeaderLoading = false;
        try {
            const all = await apiFetch("/api/locations");
            locations = all
                .filter((l) =>
                    ["falu", "község"].includes(l.type.toLowerCase()),
                )
                .sort((a, b) => a.name.localeCompare(b.name));
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
    breadcrumbLabel="Székelyföldi Falvak"
    documentTitleSuffix=" - Lámsza Index"
/>

<div class="page-inner">
    {#if loading}
        <div class="info-box"><p>Betöltés…</p></div>
    {:else if locations.length === 0}
        <div class="info-box"><p>Nincs megjeleníthető adat.</p></div>
    {:else}
        {#each locations as loc}
            <a
                href="/{loc.county_slug}-megye/{loc.slug}"
                class="card sm location"
            >
                <span class="location-name">{loc.name}</span>
                <span class="location-county">{loc.county}</span>
            </a>
        {/each}
    {/if}
</div>
<nav class="page-nav">
    <h4 class="page-nav-title">Oldal navigáció</h4>
    <ul>
        <li><a class="btn nav-btn" href="/megyek">Székelyföldi megyék</a></li>
        <li><a class="btn nav-btn" href="/szekek">Történelmi székek</a></li>
        <li><a class="btn nav-btn" href="/varosok">Székelyföldi városok</a></li>
    </ul>
</nav>

<style>
    .location-county {
        color: var(--text-faint);
    }
</style>