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
        <span class="btn btn-md" style="opacity:0.5">adat betöltés...</span>
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
    <span class="info-box">
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
            <article class="card entry--skeleton">
                <div class="skeleton skeleton-text skeleton-cat"></div>
                <div class="skeleton skeleton-text skeleton-title"></div>
                <div class="skeleton skeleton-text skeleton-loc"></div>
            </article>
        {/each}
    </div>
{:else if error}
    <span class="info-box error">
        <p>{error}</p>
    </span>
{:else if displayItems.length === 0}
    <span class="info-box info"><p>Nincs megjeleníthető bejegyzés.</p></span>
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
