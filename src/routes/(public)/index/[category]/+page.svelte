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
        {#each Array(6) as _}
            <div class="btn-skeleton"></div>
        {/each}
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
    <div class="info-box">
        <p>
            💡 Leszűrve: <span class="active">{currentCategory}</span>
            <button
                on:click={() => goto("/index")}
                class="clear-filters btn btn-xs">Szűrő törlése</button
            >
        </p>
        <p>({displayItems.length}/{totalCount})</p>
    </div>

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
                <div class="sort-toggle-menu">
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
            on:click={() => (viewMode = "grid")}>Rács</button
        >
        <button
            class="btn btn-sm {viewMode === 'flex' ? 'active' : ''}"
            on:click={() => (viewMode = "flex")}>Lista</button
        >
    </div>
</div>

{#if loading}
    <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
        {#each Array(6) as _}
            <article class="card entry--skeleton">
                <div class="skeleton skeleton-text skeleton-cat"></div>
                <div class="skeleton skeleton-text skeleton-title"></div>
                <div class="skeleton skeleton-text skeleton-loc"></div>
            </article>
        {/each}
    </div>
{:else if error}
    <div class="note error">{error}</div>
{:else if entries.length === 0}
    <div class="note info">
        Nincs megjeleníthető bejegyzés ebben a kategóriában.
    </div>
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
    .entry--skeleton {
        height: 150px;
        display: flex;
        flex-direction: column;
        padding: 1rem;
        gap: 0.5rem;
    }
    .skeleton-cat {
        width: 30%;
    }
    .skeleton-title {
        width: 80%;
        margin-top: 0.5rem;
        height: 1.2rem;
    }
    .skeleton-loc {
        width: 60%;
        margin-top: auto;
    }
</style>
