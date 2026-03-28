<script>
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import { apiFetch } from "$lib/api";

    let seats = [];
    let loading = true;

    onMount(async () => {
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

<svelte:head>
    <title>Történelmi székek - Lámsza Index</title>
</svelte:head>

<Breadcrumbs label="Történelmi székek" />
<h1 class="page-title">Székelyföld történelmi székei</h1>

<div class="page-inner">
    {#if loading}
        <span class="badge county-badge" style="opacity:0.5">adat betöltés...</span>
    {:else if seats.length === 0}
        <span class="info-box">Nincs megjeleníthető adat.</span>
    {:else}
        {#each seats as seat}
            <a href="/szekek/{seat.slug}" class="badge county-badge szek-list-link">
                {seat.name}
                {#if seat.name_ro}
                    <span class="szek-list-ro">({seat.name_ro})</span>
                {/if}
            </a>
        {/each}
    {/if}
    <p class="szek-megyek-link">
        <a href="/megyek">Székelyföldi megyék (Hargita, Kovászna, Maros)</a>
    </p>
</div>

<style>
    .szek-megyek-link {
        margin-top: 1.5rem;
        width: 100%;
        flex-basis: 100%;
    }
    .szek-megyek-link a {
        color: var(--szekely-blue, #1565c0);
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
    .szek-list-link {
        display: inline-flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 0.15rem;
    }
    .szek-list-ro {
        font-size: 0.85em;
        color: var(--text-muted);
        font-weight: 400;
    }
</style>
