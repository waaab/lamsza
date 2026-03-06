<script>
    import { onMount } from "svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";

    let dynamicCategories = [{ id: "osszes", label: "Összes", url: "/index" }];
    let entries = [];
    let loading = true;
    let error = null;

    let viewMode = "grid";
    let currentCategory = "osszes";
    let visibleCount = 12;
    let sortMode = "title";
    let sortOpen = false;

    const sortLabels = { title: "Név (A→Z)", newest: "Legújabb" };

    function setSortMode(mode) {
        sortMode = mode;
        sortOpen = false;
    }

    $: filteredEntries = entries.filter(
        (e) => currentCategory === "osszes" || e.category === currentCategory,
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

    onMount(async () => {
        try {
            const apiBase = import.meta.env.VITE_API_BASE_URL;
            const baseUrl = apiBase || "http://localhost:3000";
            const res = await fetch(`${baseUrl}/api/directory`);
            if (!res.ok) throw new Error("Hálózati hiba");
            entries = (await res.json()) || [];

            const uniqueCats = new Set(
                entries.map((e) => e.category).filter((c) => c),
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
        } catch (err) {
            console.error(err);
            error = "Hiba történt az adatok betöltésekor.";
        } finally {
            loading = false;
        }
    });
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
                class="btn btn-md {cat.id === currentCategory ? 'active' : ''}"
                on:click={() => (currentCategory = cat.id)}>{cat.label}</button
            >
        {/each}
    {/if}
</div>

<div class="filter-actions">
    <div class="info-box">
        <p>
            {#if currentCategory === "osszes"}
                💡 Leszűrve: <span class="active">Összes</span>
            {:else}
                🔍 Szűrő kiválasztva: <span class="active"
                    >{currentCategory}</span
                >
                <button
                    class="clear-filters btn btn-xs"
                    on:click={() => (currentCategory = "osszes")}
                    >Szűrő törlése</button
                >
            {/if}
        </p>
        <p><span>({displayItems.length}/{totalCount})</span></p>
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
{:else if displayItems.length === 0}
    <div class="note info">Nincs megjeleníthető bejegyzés.</div>
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
