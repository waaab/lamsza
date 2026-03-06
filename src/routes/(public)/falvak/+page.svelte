<script>
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";

    let locations = [];
    let loading = true;

    onMount(async () => {
        try {
            const apiBase =
                import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";
            const res = await fetch(`${apiBase}/api/locations`);
            if (res.ok) {
                const all = await res.json();
                locations = all
                    .filter((l) =>
                        ["falu", "község"].includes(l.type.toLowerCase()),
                    )
                    .sort((a, b) => a.name.localeCompare(b.name));
            }
        } catch (e) {
            console.error(e);
        }
        loading = false;
    });
</script>

<svelte:head>
    <title>Székelyföldi Falvak - Lámsza Index</title>
</svelte:head>

<div class="container main-container">
    <Breadcrumbs label="Székelyföldi Falvak" />
    <h1 class="page-title">Székelyföldi Falvak</h1>

    <div class="page-inner">
        {#if loading}
            <div class="skeleton skeleton-text skeleton-badge"></div>
        {:else}
            {#each locations as loc}
                <a
                    href="/{loc.county_slug}-megye/{loc.slug}"
                    class="badge location-badge"
                >
                    {loc.name}
                    <span class="location-county">{loc.county}</span>
                </a>
            {/each}
        {/if}
    </div>
</div>

<style>
    .main-container {
        min-height: calc(100vh - 120px);
    }
    .skeleton-badge {
        width: 100px;
        height: 30px;
    }
    .location-badge {
        text-decoration: none;
        color: var(--primary-color);
        background: var(--card-bg);
        font-weight: 500;
        padding: 0.8rem 1.5rem;
        border: 1px solid var(--border-color);
        font-size: 1.1rem;
    }
    .location-county {
        font-size: 0.8em;
        color: var(--text-faint);
        margin-left: 0.5rem;
    }
</style>
