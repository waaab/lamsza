<script>
    import { onMount } from "svelte";

    let entries = [];
    let loading = true;
    let error = null;

    // Filter to retain explicit Entry categories
    const retainedCategories = [
        "mesteremberek",
        "egészségügy",
        "oktatás",
        "hivatalok",
    ];

    onMount(async () => {
        try {
            const apiBase =
                import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";
            const res = await fetch(`${apiBase}/api/directory`);
            if (!res.ok) throw new Error("Hiba a címtár betöltésekor");

            const allEntries = (await res.json()) || [];

            // Filter down exclusively to the legacy entries
            entries = allEntries.filter(
                (s) =>
                    s.category &&
                    retainedCategories.includes(s.category.toLowerCase()),
            );
        } catch (err) {
            console.error(err);
            error = "Nem sikerült betölteni a bejegyzéseket.";
        } finally {
            loading = false;
        }
    });
</script>

<svelte:head>
    <title>Bejegyzések - Index</title>
</svelte:head>

<div class="breadcrumbs">
    <a href="/">Főoldal</a> &rsaquo;
    <a href="/index">Index</a> &rsaquo;
    <span class="active">Kiemelt Bejegyzések</span>
</div>

<h1 class="page-title">Bejegyzések</h1>
<p class="greeting">
    Kiemelt helyi bejegyzések, mesteremberek és egészségügyi intézmények indexe.
</p>

{#if loading}
    <div class="list grid">
        {#each Array(6) as _}
            <article class="card entry-placeholder">
                <span class="entry-placeholder-cat">adat betöltés...</span>
                <span class="entry-placeholder-title">adat betöltés...</span>
                <span class="entry-placeholder-loc">adat betöltés...</span>
            </article>
        {/each}
    </div>
{:else if error}
    <span class="info-box error">
        <p>{error}</p>
    </span>
{:else if entries.length === 0}
    <span class="info-box error">
        <p>Jelenleg nincs listázott bejegyzés ebben a kategóriában.</p>
    </span>
{:else}
    <div class="list grid">
        {#each entries as entry}
            <article class="card entry entry-card">
                <span class="badge entry-badge">{entry.category}</span>
                <h3 class="entry-name">
                    <a href="/bejegyzes/{entry.slug}" class="entry-link"
                        >{entry.name}</a
                    >
                </h3>
                {#if entry.url}
                    <div class="entry-info entry-url-wrap">
                        <span class="entry-url-icon">🔗</span>
                        <a
                            href={entry.url}
                            target="_blank"
                            rel="nofollow noopener"
                            class="entry-url-link">{entry.url}</a
                        >
                    </div>
                {/if}
                <div class="entry-info">📍 
                <a href="/{entry.county_slug}-megye/{entry.location_slug}" class="entry-link">
                    {[entry.location, entry.location_ro, entry.location_de]
                        .filter(Boolean)
                        .join(" | ")} - {entry.address}
                </a>
                </div>
                {#if entry.phone}
                    <div class="entry-info entry-phone">
                        📞 {entry.phone}
                    </div>
                {/if}
            </article>
        {/each}
    </div>
{/if}

<style>
    .entry-placeholder {
        display: flex;
        flex-direction: column;
        padding: 1rem;
        gap: 0.5rem;
    }
    .entry-placeholder-cat,
    .entry-placeholder-loc {
        font-size: 0.75rem;
        color: var(--text-faint);
    }
    .entry-placeholder-title {
        font-size: 0.95rem;
        color: var(--text-faint);
        margin-top: 0.5rem;
    }
    .entry-card {
        cursor: pointer;
        transition:
            transform 0.2s,
            box-shadow 0.2s;
    }
    .entry-badge {
        margin-bottom: 0.5rem;
        display: inline-block;
    }
    .entry-link {
        color: inherit;
        text-decoration: none;
    }
    .entry-url-wrap {
        margin-bottom: 0.5rem;
    }
    .entry-url-icon {
        color: var(--text-faint);
        margin-right: 0.3rem;
    }
    .entry-url-link {
        color: var(--primary-color);
        text-decoration: none;
    }
    .entry-phone {
        color: var(--text-faint);
        margin-top: 0.25rem;
    }
</style>
