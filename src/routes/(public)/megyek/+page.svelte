<script>
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import { apiFetch } from "$lib/api";

    let locations = [];
    let loading = true;

    onMount(async () => {
        try {
            const all = await apiFetch("/api/locations");
            locations = all
                .filter((l) => l.type.toLowerCase() === "megye")
                .sort((a, b) => a.name.localeCompare(b.name));
        } catch (e) {
            console.error(e);
        } finally {
            loading = false;
        }
    });
</script>

<svelte:head>
    <title>Székelyföldi Megyék - Lámsza Index</title>
</svelte:head>

<Breadcrumbs label="Székelyföldi Megyék" />
<h1 class="page-title">Székelyföldi Megyék</h1>

<div class="page-inner">
    {#if loading}
        <div class="skeleton-badge skeleton"></div>
    {:else}
        {#each locations as loc}
            <a href="/{loc.slug}-megye" class="badge county-badge">
                {loc.name}
            </a>
        {/each}
    {/if}
</div>

<style>
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
