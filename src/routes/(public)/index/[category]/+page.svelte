<script>
    import { browser } from "$app/environment";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import { onMount } from "svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";

    let dynamicCategories = [{ id: "osszes", label: "Összes", url: "/index" }];
    let entries = [];
    let loading = true;
    let error = null;

    let viewMode = "grid";
    let currentCategory = "";
    let visibleCount = 12;
    let sortMode = "title";
    let sortOpen = false;

    const sortLabels = { title: "Név (A→Z)", newest: "Legújabb" };

    function setSortMode(mode) {
        sortMode = mode;
        sortOpen = false;
    }

    $: filteredEntries = entries.filter(
        (e) =>
            currentCategory === "osszes" ||
            e.category.toLowerCase() === currentCategory.toLowerCase(),
    );
    $: sortedEntries = [...filteredEntries].sort((a, b) => {
        if (sortMode === "newest") return b.id - a.id;
        return a.name.localeCompare(b.name);
    });
    $: totalCount = sortedEntries.length;
    $: displayItems = sortedEntries.slice(0, visibleCount);

    function loadMore() {
        visibleCount += 12;
    }

    $: if (currentCategory) {
        visibleCount = 12;
    }

    $: if (browser) {
        const categoryId = $page.params.category;
        currentCategory = categoryId;
        fetchData(categoryId);
    }

    async function fetchData(categoryId) {
        loading = true;
        error = null;
        try {
            const apiBase = import.meta.env.VITE_API_BASE_URL;
            const baseUrl = apiBase || "http://localhost:3000";
            const res = await fetch(`${baseUrl}/api/directory`);
            if (!res.ok) throw new Error("Hálózati hiba");
            const allEntries = (await res.json()) || [];

            const uniqueCats = new Set(
                allEntries.map((e) => e.category).filter((c) => c),
            );
            const generatedCats = Array.from(uniqueCats).map((catName) => {
                return {
                    id: catName,
                    label: catName,
                    url: "/index/" + encodeURIComponent(catName),
                };
            });
            dynamicCategories = [
                { id: "osszes", label: "Összes", url: "/index" },
                ...generatedCats,
            ];

            entries = allEntries.filter((e) => {
                return (
                    e.category.toLowerCase() === categoryId.toLowerCase() ||
                    (categoryId === "egeszsegugy" &&
                        e.category === "Egészségügy") ||
                    (categoryId === "oktatas" && e.category === "Oktatás") ||
                    (categoryId === "mesteremberek" &&
                        e.category === "Mesteremberek") ||
                    (categoryId === "hivatalok" &&
                        e.category === "Hivatalok") ||
                    (categoryId === "egyeb" && e.category === "Egyéb") ||
                    e.category_id === categoryId
                );
            });
        } catch (err) {
            console.error(err);
            error = "Hiba történt az adatok betöltésekor.";
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head>
    <title>Szekely Gugel - Index</title>
</svelte:head>

<h1 class="page-title">Index</h1>
<p class="greeting">Keresd meg a helyi szakembereket és intézményeket!</p>

<div class="header-tabs">
    <span class="header-tabs-label">Kiemelt Kategóriák:</span>
    {#if loading}
        <span class="btn btn-md" style="opacity:0.5">adat betöltés...</span>
    {:else}
        {#each dynamicCategories as cat}
            <button
                class="btn btn-md {cat.id === currentCategory ||
                cat.id.toLowerCase() === currentCategory.toLowerCase()
                    ? 'active'
                    : ''}"
                on:click={() => goto(cat.url)}>{cat.label}</button
            >
        {/each}
    {/if}
</div>

<div class="filter-actions">
    <span class="info-box">
        <p>
            💡 Leszűrve: <span class="active">{currentCategory}</span>
            <button
                on:click={() => goto("/index")}
                class="clear-filters btn btn-xs">Szűrő törlése</button
            >
        </p>
        <p>({displayItems.length}/{totalCount})</p>
    </span>

    <div class="view-mode-toggle">
        <div class="sort-toggle">
            <button class="btn btn-sm" on:click={() => (sortOpen = !sortOpen)}>
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><line x1="4" y1="6" x2="16" y2="6"></line><line
                        x1="4"
                        y1="12"
                        x2="12"
                        y2="12"
                    ></line><line x1="4" y1="18" x2="8" y2="18"></line><polyline
                        points="15 15 18 18 21 15"
                    ></polyline><line x1="18" y1="10" x2="18" y2="18"
                    ></line></svg
                >
                <span>{sortLabels[sortMode]}</span>
            </button>
            {#if sortOpen}
                <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
                <div class="sort-toggle-menu" on:click|stopPropagation>
                    <button
                        class:active={sortMode === "title"}
                        on:click={() => setSortMode("title")}>Név (A→Z)</button
                    >
                    <button
                        class:active={sortMode === "newest"}
                        on:click={() => setSortMode("newest")}>Legújabb</button
                    >
                </div>
            {/if}
        </div>

        <button
            class="btn btn-sm {viewMode === 'grid' ? 'active' : ''}"
            on:click={() => (viewMode = "grid")}
            title="Rács nézet"
        >
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><rect x="3" y="3" width="7" height="7"></rect><rect
                    x="14"
                    y="3"
                    width="7"
                    height="7"
                ></rect><rect x="14" y="14" width="7" height="7"></rect><rect
                    x="3"
                    y="14"
                    width="7"
                    height="7"
                ></rect></svg
            >
            <span>Rács</span>
        </button>
        <button
            class="btn btn-sm {viewMode === 'flex' ? 'active' : ''}"
            on:click={() => (viewMode = "flex")}
            title="Lista nézet"
        >
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><line x1="8" y1="6" x2="21" y2="6"></line><line
                    x1="8"
                    y1="12"
                    x2="21"
                    y2="12"
                ></line><line x1="8" y1="18" x2="21" y2="18"></line><line
                    x1="3"
                    y1="6"
                    x2="3.01"
                    y2="6"
                ></line><line x1="3" y1="12" x2="3.01" y2="12"></line><line
                    x1="3"
                    y1="18"
                    x2="3.01"
                    y2="18"
                ></line></svg
            >
            <span>Lista</span>
        </button>
    </div>
</div>

{#if loading}
    <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
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
    <span class="info-box info">
        <p>Nincs megjeleníthető bejegyzés ebben a kategóriában.</p>
    </span>
{:else}
    <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
        {#each displayItems as entry}
            <EntryCard {entry} />
        {/each}
    </div>
    {#if visibleCount < totalCount}
        <div class="load-more">
            <button class="nav-btn" on:click={loadMore}>Több betöltése ↓</button
            >
        </div>
    {/if}
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
</style>
