<script>
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";

    let slug = "";
    let entry = null;
    let loading = true;
    let error = null;

    $: if (browser && $page.params.slug) {
        slug = $page.params.slug;
        fetchEntry();
    }

    async function fetchEntry() {
        loading = true;
        error = null;
        const apiBase =
            import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

        try {
            const res = await fetch(
                `${apiBase}/api/service?slug=${encodeURIComponent(slug)}`,
            );
            if (!res.ok) throw new Error("Hiba az adat lekérésekor");

            const data = await res.json();
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
    {#if entry}
        <title>{entry.name} - Index</title>
    {:else}
        <title>Bejegyzés - Index</title>
    {/if}
</svelte:head>

<div class="container" style="min-height: calc(100vh - 120px);">
    {#if loading}
        <div
            class="skeleton skeleton-text"
            style="width: 40%; height: 2rem; margin-bottom: 2rem;"
        ></div>
    {:else if error}
        <div class="error-msg">{error}</div>
        <a href="/" class="btn" style="margin-top: 1rem; display: inline-block;"
            >Vissza a főoldalra</a
        >
    {:else if entry}
        <Breadcrumbs
            label={entry.name}
            countySlug={entry.county_slug}
            countyName={entry.county}
            settlementSlug={entry.location_slug}
            settlementName={entry.location}
            settlementType={entry.location_type}
        />

        <div style="margin-top: 2rem;">
            <!-- Remodeled Full-Width Profile -->
            <div style="margin-bottom: 2rem;">
                <div
                    class="badge"
                    style="font-size: 1rem; margin-bottom: 1rem; display: inline-block;"
                >
                    Index: {entry.category}
                </div>
                <h1
                    style="margin: 0 0 1rem; font-size: 2.5rem; color: var(--text-color);"
                >
                    {entry.name}
                </h1>

                {#if entry.url}
                    <div style="margin-bottom: 1.5rem; font-size: 1.1rem;">
                        <span
                            style="color: var(--text-faint); margin-right: 0.5rem;"
                            >🔗 Weboldal:</span
                        >
                        <a
                            href={entry.url}
                            target="_blank"
                            rel="nofollow noopener"
                            style="color: var(--primary-color); text-decoration: none; font-weight: 500;"
                        >
                            {entry.url}
                        </a>
                    </div>
                {/if}
            </div>

            <!-- Distinct Contact Block -->
            <div
                style="padding: 1.5rem; background: var(--card-bg); border-radius: 12px; border: 1px solid var(--border-color); margin-bottom: 2rem;"
            >
                <h3
                    style="margin-top: 0; color: var(--text-color); font-size: 1.2rem; margin-bottom: 1rem;"
                >
                    Kapcsolat
                </h3>
                <div style="display: grid; gap: 1rem;">
                    {#if entry.location || entry.address}
                        <div
                            style="display: flex; align-items: flex-start; gap: 0.8rem; font-size: 1.1rem;"
                        >
                            <span style="font-size: 1.3rem;">📍</span>
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
                        <div
                            style="display: flex; align-items: center; gap: 0.8rem; font-size: 1.1rem;"
                        >
                            <span style="font-size: 1.3rem;">📞</span>
                            <a
                                href={`tel:${entry.phone.replace(/[^0-9+]/g, "")}`}
                                style="color: var(--text-color); text-decoration: none;"
                                >{entry.phone}</a
                            >
                        </div>
                    {/if}
                </div>
            </div>

            <!-- Details Block -->
            <div
                style="padding: 1.5rem; background: var(--bg-body); border-radius: 12px; margin-bottom: 2rem;"
            >
                <h3
                    style="margin-top: 0; color: var(--text-color); font-size: 1.2rem; margin-bottom: 1rem;"
                >
                    Részletek & Megjegyzések
                </h3>
                {#if entry.notes}
                    <div
                        class="service-notes"
                        style="font-size: 1.05rem; line-height: 1.6; margin-bottom: 1.5rem; white-space: pre-wrap; color: var(--text-faint);"
                    >
                        {entry.notes}
                    </div>
                {/if}

                {#if entry.tags && entry.tags.length > 0}
                    <div class="service-tags">
                        {#each entry.tags as t}
                            <span
                                class="service-tag"
                                style="padding: 0.4rem 0.8rem; font-size: 0.95rem;"
                                >{t.startsWith("#") ? t : "#" + t}</span
                            >
                        {/each}
                    </div>
                {/if}
            </div>
        </div>
    {/if}
</div>
