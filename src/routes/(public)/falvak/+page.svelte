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
                    .filter((l) => l.type.toLowerCase() === "falu")
                    .sort((a, b) => a.name.localeCompare(b.name));
            }
        } catch (e) {
            console.error(e);
        }
        loading = false;
    });
</script>

<svelte:head>
    <title>Falvak - Index</title>
</svelte:head>

<div class="container" style="min-height: calc(100vh - 120px)">
    <Breadcrumbs label="Falvak" />
    <h1 class="page-title">Erdélyi Falvak</h1>

    <div style="display: flex; flex-wrap: wrap; gap: 0.8rem; margin-top:2rem;">
        {#if loading}
            <div
                class="skeleton skeleton-text"
                style="width: 100px; height: 30px;"
            ></div>
        {:else}
            {#each locations as loc}
                <a
                    href="/{loc.county_slug}-megye/{loc.slug}"
                    class="badge"
                    style="text-decoration: none; color: var(--primary-color); background: var(--card-bg); font-weight: 500; padding: 0.8rem 1.5rem; border: 1px solid var(--border-color); font-size: 1.1rem;"
                >
                    {loc.name}
                    <span
                        style="font-size:0.8em; color:var(--text-faint); margin-left:0.5rem;"
                        >{loc.county}</span
                    >
                </a>
            {/each}
        {/if}
    </div>
</div>
