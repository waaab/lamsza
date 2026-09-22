<script>
    import { onMount } from "svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import IndexTagAside from "$lib/components/IndexTagAside.svelte";
    import {
        entryMatchesAsideFilters,
        normalizeTagKey,
        displayTagLabel,
    } from "$lib/directoryTagCloud.js";
    import {
        canonicalEntryType,
        canonicalEntryTypeKey,
    } from "$lib/entryType.js";
    import {
        canonicalEntryCategory,
        directoryCategoryTabs,
        entryMatchesCategory,
    } from "$lib/entryCategory.js";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";
    import { apiFetch } from "$lib/api.js";

    let pageHeader = initialPageHeader("index");
    let pageHeaderLoading = false;

    let dynamicCategories = [{ id: "osszes", label: "Összes", url: "/index" }];
    let entries = [];
    let loading = true;
    let error = null;

    /** @type {string | null} */
    let selectedTypeKey = null;
    /** @type {string | null} */
    let selectedTagKey = null;

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

    function scrollToTop() {
        if (typeof window !== "undefined") {
            window.scrollTo({ top: 0, behavior: "smooth" });
        }
    }

    /** Kategória (nem Összes) vagy oldalsáv típus/címke szűrő */
    $: hasActiveFilters =
        currentCategory !== "osszes" ||
        selectedTypeKey != null ||
        selectedTagKey != null;

    function clearAllFilters() {
        currentCategory = "osszes";
        selectedTypeKey = null;
        selectedTagKey = null;
        scrollToTop();
    }

    $: filteredEntries = entries.filter(
        (e) =>
            (currentCategory === "osszes" ||
                entryMatchesCategory(e, currentCategory)) &&
            entryMatchesAsideFilters(e, selectedTypeKey, selectedTagKey),
    );

    $: categoryFilterLabel =
        currentCategory === "osszes"
            ? null
            : dynamicCategories.find((c) => c.id === currentCategory)?.label ||
              canonicalEntryCategory(currentCategory);

    $: typeFilterLabel =
        selectedTypeKey &&
        canonicalEntryType(
            entries.find(
                (e) => canonicalEntryTypeKey(e.type) === selectedTypeKey,
            )?.type,
        );

    $: tagFilterLabel = (() => {
        if (!selectedTagKey) return null;
        for (const e of entries) {
            for (const raw of e.tags || []) {
                if (normalizeTagKey(raw) === selectedTagKey) {
                    return displayTagLabel(raw);
                }
            }
        }
        return selectedTagKey;
    })();
    $: sortedEntries = [...filteredEntries].sort((a, b) => {
        if (sortMode === "newest") return b.id - a.id;
        return a.name.localeCompare(b.name);
    });
    $: totalCount = sortedEntries.length;
    $: displayItems = sortedEntries.slice(0, visibleCount);

    function loadMore() {
        visibleCount += 12;
    }

    $: {
        currentCategory;
        selectedTypeKey;
        selectedTagKey;
        visibleCount = 12;
    }

    onMount(async () => {
        loadPageMeta("index").then((p) => {
            pageHeader = p;
        });
        try {
            entries = (await apiFetch("/api/directory")) || [];
            dynamicCategories = directoryCategoryTabs(entries);
        } catch (err) {
            console.error(err);
            error = "Hiba történt az adatok betöltésekor.";
        } finally {
            loading = false;
        }
    });
</script>

<PublicPageHero
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    loading={pageHeaderLoading}
    breadcrumbLabel="Index"
    breadcrumbParentLabel=""
    breadcrumbParentUrl=""
    documentTitleSuffix=" - Székely Gugel"
/>

<div class="header-tabs">
    <span class="header-tabs-label" aria-label="Kiemelt Kategóriák">Kiemelt Kategóriák:</span>
    {#if loading}
        <span class="btn btn--loading">Szűrők betöltése…</span>
    {:else}
        <div class="header-tabs-filters-row">
            {#each dynamicCategories as cat}
                <button
                    class="btn btn-md {cat.id === currentCategory ? 'active' : ''}"
                    on:click={() => (currentCategory = cat.id)}>{cat.label}</button
                >
            {/each}
        </div>
    {/if}
</div>
{#snippet indexFilterBar()}
    <div class="filter-actions">
        <span class="info-box">
            <p>
                {#if !hasActiveFilters}
                    💡 Leszűrve: <span class="active">Összes</span>
                {:else}
                    🔍 Szűrők:
                    {#if currentCategory !== "osszes"}
                        <span class="active">{categoryFilterLabel}</span>
                    {/if}
                    {#if typeFilterLabel}
                        {#if currentCategory !== "osszes"}<span class="filter-sep">·</span
                            >{/if}
                        <span class="active">{typeFilterLabel}</span>
                    {/if}
                    {#if tagFilterLabel}
                        {#if currentCategory !== "osszes" || typeFilterLabel}<span
                                class="filter-sep">·</span
                            >{/if}
                        <span class="active">{tagFilterLabel}</span>
                    {/if}
                    <button
                        type="button"
                        class="clear-filters btn btn-xs"
                        aria-label="Szűrők törlése"
                        title="Szűrők törlése"
                        on:click={clearAllFilters}>Szűrő törlése</button
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
{/snippet}

{@render indexFilterBar()}

<div class="list-page-layout">
    <section class="list">
        {#if loading}
            <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
                {#each Array(6) as _, i (i)}
                    <EntryCard placeholder layout={viewMode === "grid" ? "grid" : "list"} />
                {/each}
            </div>
        {:else if error}
            <span class="info-box error">
                <p>{error}</p>
            </span>
        {:else if displayItems.length === 0}
            <span class="info-box info"
                ><p>Nincs megjeleníthető bejegyzés.</p></span
            >
        {:else}
            <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
                {#each displayItems as entry}
                    <EntryCard {entry} layout={viewMode === "grid" ? "grid" : "list"} />
                {/each}
            </div>
            {#if visibleCount < totalCount}
                <div class="load-more">
                    <button class="btn nav-btn" on:click={loadMore}
                        >Több betöltése ↓</button
                    >
                </div>
            {/if}
        {/if}
    </section>

    <aside class="sidebar index-tags-sidebar" aria-label="Címkék">
        <div class="sidebar-box">
            <div class="sidebar-header">
                <h4 class="sidebar-heading">Címkék</h4>
            </div>
            {#if loading}
                <div
                    class="index-tags-aside-skeleton"
                    aria-busy="true"
                    aria-label="Címkék betöltése"
                >
                    <div class="index-tags-aside-skeleton__row">
                        {#each Array(8) as _}
                            <span class="skeleton index-tags-aside-skeleton__chip"></span>
                        {/each}
                    </div>
                    <div class="index-tags-aside-skeleton__row">
                        {#each Array(6) as _}
                            <span class="skeleton index-tags-aside-skeleton__chip"></span>
                        {/each}
                    </div>
                </div>
            {:else if error}
                <p class="index-tags-aside__empty">Nem sikerült betölteni a címkéket.</p>
            {:else}
                <IndexTagAside
                    bind:selectedTypeKey
                    bind:selectedTagKey
                    {entries}
                />
            {/if}
        </div>
    </aside>
</div>

{@render indexFilterBar()}

<section class="index-stats-section">
    <h2 class="index-stats-section-title">Index statisztikák:</h2>
    <div class="index-stats-items">
        <div class="card index-stats-item">
            <div class="index-stats-item-content">
                <div class="index-stats-item-icon"><svg xmlns="http://www.w3.org/2000/svg" height="50px" viewBox="0 -960 960 960" width="50px" fill="currentColor"><path d="m599-538 138-138-35-34-103 104-53-54-35 35 88 87ZM143-192v-72h432v72H143Zm345-296q-56-56-56-136t56-136q56-56 136-56t136 56q56 56 56 136t-56 136q-56 56-136 56t-136-56Zm-345-4v-72h224q5 20 12 37.5t17 34.5H143Zm0 150v-72h322q24 18 51.5 30.5T575-365v23H143Z"/></svg></div>
                <span class="index-stats-item-value">{totalCount}</span>
                <h3 class="index-stats-item-title">Ellenőrzött Bejegyzés</h3>
            </div>
        </div>
        <div class="card index-stats-item">
            <div class="index-stats-item-content">
                <div class="index-stats-item-icon"><svg xmlns="http://www.w3.org/2000/svg" height="50px" viewBox="0 -960 960 960" width="50px" fill="currentColor"><path d="M240-336h312v-72H240v72Zm0-120h480v-72H240v72Zm-72 264q-29 0-50.5-21.5T96-264v-432q0-29.7 21.5-50.85Q139-768 168-768h216l96 96h312q29.7 0 50.85 21.15Q864-629.7 864-600v336q0 29-21.15 50.5T792-192H168Zm0-72h624v-336H450l-96-96H168v432Zm0 0v-432 432Z"/></svg></div>
                <span class="index-stats-item-value">10</span>
                <h3 class="index-stats-item-title">Kategória</h3>
            </div>
        </div>
        <div class="card index-stats-item">
            <div class="index-stats-item-content">
                <div class="index-stats-item-icon">
                    <svg xmlns="http://www.w3.org/2000/svg" height="50px" viewBox="0 -960 960 960" width="50px" fill="currentColor"><path d="M216-144q-29.7 0-50.85-21.15Q144-186.3 144-216v-528q0-29.7 21.15-50.85Q186.3-816 216-816h528q29.7 0 50.85 21.15Q816-773.7 816-744v258q-17.1-5.76-35.1-9.92T744-502v-242H216v528h241q1.88 19.52 5.94 37.26Q467-161 473-144H216Zm0-96v24-528 242-2 264Zm72-48h172q4-19 10.19-36.97Q476.38-342.93 484-360H288v72Zm0-156h264q26-20 56-34.5t64-20.5v-17H288v72Zm0-156h384v-72H288v72ZM719.77-48Q640-48 584-104.23q-56-56.22-56-136Q528-320 584.23-376q56.22-56 136-56Q800-432 856-375.77q56 56.22 56 136Q912-160 855.77-104q-56.22 56-136 56ZM696-144h48v-72h72v-48h-72v-72h-48v72h-72v48h72v72Z"/></svg>
                </div>
                <span class="index-stats-item-value">50000</span>
                <h3 class="index-stats-item-title">Havi felhasználó</h3>
            </div>
        </div>
    </div>
</section>
<style>

.index-stats-items{
    display: flex;
    gap: 1rem;
    flex-direction: row;
    justify-content: space-around;
}

.index-stats-item{
    border: none;
    box-shadow: none;
    background: none;
}

.index-stats-section-title{
    text-align: center;
}

.index-stats-item-content{
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
}

.index-stats-item-value{
    font-weight: 700;
    color: var(--text-primary);
    line-height: 1;
    margin: 1rem 0 0;
}
    .index-tags-aside-skeleton {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        padding: 0.25rem 0 0.5rem;
    }

    .index-tags-aside-skeleton__row {
        display: flex;
        flex-wrap: wrap;
        gap: 0.4rem;
    }

    .index-tags-aside-skeleton__chip {
        display: inline-block;
        height: 1.85rem;
        border-radius: 999px;
        min-width: 3.5rem;
    }

    .index-tags-aside-skeleton__row :nth-child(3n) {
        min-width: 4.75rem;
    }

    .index-tags-aside-skeleton__row :nth-child(4n) {
        min-width: 5.25rem;
    }
</style>