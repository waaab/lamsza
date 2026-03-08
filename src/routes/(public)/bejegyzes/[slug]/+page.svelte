<script>
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import EventsWidget from "$lib/components/EventsWidget.svelte";
    import { apiFetch } from "$lib/api";

    let entry = null;
    let loading = true;
    let error = null;

    $: slug = $page.params.slug;

    $: if (browser && slug) {
        fetchEntry();
    }

    async function fetchEntry() {
        loading = true;
        error = null;
        try {
            const data = await apiFetch(
                `/api/entry?slug=${encodeURIComponent(slug)}`,
            );
            if (!data || !data.name) {
                error = "A bejegyzés nem található.";
            } else {
                entry = data;
            }
        } catch (err) {
            console.error(err);
            error = "Hiba történt a szerver kapcsolat közben.";
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head>
    <title>{entry ? entry.name : "Bejegyzés"} - Index</title>
</svelte:head>

{#if loading}
    <p class="loading-placeholder">adat betöltés...</p>
{:else if error}
    <span class="info-box error">
        <p>{error}</p>
    </span>
    <a href="/" class="btn back-to-home">Vissza a főoldalra</a>
{:else if entry}
    <Breadcrumbs
        label={entry.name}
        countySlug={entry.county_slug}
        countyName={entry.county}
        settlementSlug={entry.location_slug}
        settlementName={entry.location}
        settlementType={entry.location_type}
    />

    <div class="entry-content">
        <div class="entry-header">
            <div class="badge">Index: {entry.category}</div>
            <h1 class="entry-title">{entry.name}</h1>

            {#if entry.url}
                <div class="entry-url-row">
                    <span class="entry-url-label">🔗 Weboldal:</span>
                    <a
                        href={entry.url}
                        target="_blank"
                        rel="nofollow noopener"
                        class="entry-url-link">{entry.url}</a
                    >
                </div>
            {/if}
        </div>

        <div class="contact-card">
            <h3 class="contact-title">Kapcsolat</h3>
            <div class="contact-grid">
                {#if entry.location || entry.address}
                    <div class="contact-item">
                        <span class="contact-icon">📍</span>
                        <div>
                            {#if entry.location || entry.location_ro || entry.location_de}
                                <strong
                                    >{[
                                        entry.location,
                                        entry.location_ro,
                                        entry.location_de,
                                    ]
                                        .filter(Boolean)
                                        .join(" | ")}</strong
                                >
                            {/if}
                            {#if (entry.location || entry.location_ro || entry.location_de) && entry.address}
                                -
                            {/if}
                            {#if entry.address}{entry.address}{/if}
                        </div>
                    </div>
                {/if}

                {#if entry.phone}
                    <div class="contact-item-center">
                        <span class="contact-icon">📞</span>
                        <a
                            href={`tel:${entry.phone.replace(/[^0-9+]/g, "")}`}
                            class="contact-link">{entry.phone}</a
                        >
                    </div>
                {/if}
            </div>
        </div>

        <div class="details-section">
            <h3 class="details-title">Részletek & Megjegyzések</h3>
            {#if entry.notes}
                <div class="entry-notes">{entry.notes}</div>
            {/if}

            {#if entry.tags && entry.tags.length > 0}
                <div class="entry-tags">
                    {#each entry.tags as t}
                        <span class="entry-tag tag-padded"
                            >{t.startsWith("#") ? t : "#" + t}</span
                        >
                    {/each}
                </div>
            {/if}
        </div>

        <EventsWidget organizerName={entry.name} />
    </div>
{/if}

<style>
    .loading-placeholder {
        color: var(--text-faint);
        margin-bottom: 2rem;
    }
    .back-to-home {
        margin-top: 1rem;
        display: inline-block;
    }
    .entry-content {
        margin-top: 2rem;
    }
    .entry-header {
        margin-bottom: 2rem;
    }
    .entry-title {
        margin: 0 0 1rem;
    }
    .entry-url-row {
        margin-bottom: 1.5rem;
        font-size: 1.1rem;
    }
    .entry-url-label {
        color: var(--text-faint);
        margin-right: 0.5rem;
    }
    .entry-url-link {
        color: var(--primary-color);
        text-decoration: none;
        font-weight: 500;
    }

    .contact-card {
        padding: 1.5rem;
        background: var(--card-bg);
        border-radius: 12px;
        border: 1px solid var(--border-color);
        margin-bottom: 2rem;
    }
    .contact-title {
        margin-top: 0;
        color: var(--text-color);
        font-size: 1.2rem;
        margin-bottom: 1rem;
    }
    .contact-grid {
        display: grid;
        gap: 1rem;
    }
    .contact-item {
        display: flex;
        align-items: flex-start;
        gap: 0.8rem;
        font-size: 1.1rem;
    }
    .contact-item-center {
        display: flex;
        align-items: center;
        gap: 0.8rem;
        font-size: 1.1rem;
    }
    .contact-icon {
        font-size: 1.3rem;
    }
    .contact-link {
        color: var(--text-color);
        text-decoration: none;
    }

    .details-section {
        margin-bottom: 2rem;
    }
    .details-title {
        margin-top: 0;
        color: var(--text-color);
        font-size: 1.2rem;
        margin-bottom: 1rem;
    }
    .entry-notes {
        font-size: 1rem;
        line-height: 1.6;
        margin-bottom: 1.5rem;
        white-space: pre-wrap;
        color: var(--text-faint);
    }
    .tag-padded {
        padding: 0.4rem 0.8rem;
        font-size: 0.95rem;
    }
</style>
