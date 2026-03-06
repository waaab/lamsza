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
                    .filter((l) => l.type.toLowerCase() === "megye")
                    .sort((a, b) => a.name.localeCompare(b.name));
            }
        } catch (e) {
            console.error(e);
        }
        loading = false;
    });
</script>

<svelte:head>
    <title>Székelyföldi Megyék - Lámsza Index</title>
</svelte:head>

<div class="container main-container">
    <Breadcrumbs label="Székelyföldi Megyék" />
    <h1 class="page-title">Székelyföldi Megyék</h1>

    <div class="page-inner">
        {#if loading}
            <div class="skeleton skeleton-text skeleton-badge"></div>
        {:else}
            {#each locations as loc}
                <a href="/{loc.slug}-megye" class="badge county-badge">
                    {loc.name}
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
    .county-badge {
        text-decoration: none;
        color: var(--primary-color);
        background: var(--card-bg);
        font-weight: 500;
        padding: 0.8rem 1.5rem;
        border: 1px solid var(--border-color);
        font-size: 1.1rem;
    }
</style>
