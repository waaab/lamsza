<script>
    import { onMount } from "svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { apiFetch } from "$lib/api.js";
    import { filterServiceEntries } from "$lib/entryType.js";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";

    const PAGE_SLUG = "index/szolgaltatasok";
    let pageHeader = $state(initialPageHeader(PAGE_SLUG));
    const pageHeaderLoading = false;

    let entries = $state([]);
    let loading = $state(true);
    let error = $state(/** @type {string | null} */ (null));

    onMount(async () => {
        loadPageMeta(PAGE_SLUG).then((p) => {
            pageHeader = p;
        });
        try {
            const allEntries = (await apiFetch("/api/directory")) || [];
            entries = filterServiceEntries(allEntries);
        } catch (err) {
            console.error(err);
            error = "Nem sikerült betölteni a szolgáltatásokat.";
        } finally {
            loading = false;
        }
    });
</script>

<PublicPageHero
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    loading={pageHeaderLoading}
    breadcrumbLabel={pageHeader.title || "Szolgáltatások"}
    breadcrumbParentLabel="Index"
    breadcrumbParentUrl="/index"
    documentTitleSuffix=" - Székely Gugel"
/>

{#if loading}
    <div class="list grid">
        {#each Array(6) as _, i (i)}
            <EntryCard placeholder layout="grid" />
        {/each}
    </div>
{:else if error}
    <span class="info-box error">
        <p>{error}</p>
    </span>
{:else if entries.length === 0}
    <span class="info-box info">
        <p>Jelenleg nincs szolgáltatástípusú bejegyzés.</p>
    </span>
{:else}
    <div class="list grid">
        {#each entries as entry (entry.id)}
            <EntryCard {entry} layout="grid" />
        {/each}
    </div>
{/if}
