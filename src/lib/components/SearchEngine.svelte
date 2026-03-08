<script>
    import { createEventDispatcher } from "svelte";
    import { apiFetch } from "$lib/api";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import { formatDateShort } from "$lib/utils";

    const dispatch = createEventDispatcher();

    let showDiscover = false;

    let searchInputValue = "";
    let searchResults = null; // { locations, entries, events, news }
    let suggestions = [];
    let loading = false;
    let searchInputEl;

    $: hasResults = searchResults && (
        (searchResults.locations && searchResults.locations.length > 0) ||
        (searchResults.entries && searchResults.entries.length > 0) ||
        (searchResults.events && searchResults.events.length > 0) ||
        (searchResults.news && searchResults.news.length > 0)
    );
    $: totalCount = searchResults
        ? (searchResults.locations?.length || 0) +
          (searchResults.entries?.length || 0) +
          (searchResults.events?.length || 0) +
          (searchResults.news?.length || 0)
        : 0;

    async function executeSearch() {
        if (!searchInputValue.trim()) return;

        loading = true;
        showDiscover = true;
        try {
            const data = await apiFetch(
                `/api/search?q=${encodeURIComponent(searchInputValue)}`,
            );
            searchResults = data;

            const suggestionsData = await apiFetch(
                `/api/autosuggest?q=${encodeURIComponent(searchInputValue)}`,
            );
            suggestions = suggestionsData || [];
        } catch (err) {
            console.error("Search error:", err);
            searchResults = { locations: [], entries: [], events: [], news: [] };
        } finally {
            loading = false;
        }
    }

    function openKapu() {
        showDiscover = true;
        dispatch("discoverOpen");
        setTimeout(() => searchInputEl?.focus(), 50);
    }

    function closeDiscover() {
        showDiscover = false;
        searchInputValue = "";
        searchResults = null;
        suggestions = [];
        dispatch("discoverClose");
    }

    function clearSearch() {
        searchInputValue = "";
        searchResults = null;
        suggestions = [];
    }

    function handleKeydown(e) {
        if (e.key === "Enter") executeSearch();
        if (e.key === "Escape") closeDiscover();
    }

    function locationToEntry(loc) {
        return {
            entity_type: "settlement",
            name: loc.name,
            slug: loc.slug,
            county_slug: loc.county_slug,
            location: loc.county,
        };
    }
</script>

<div
    class="search-discover-wrapper"
    class:search-discover-wrapper--expanded={showDiscover}
>
    <section class="search-container">
        <input
            type="text"
            name="search"
            id="search"
            class="search-input {showDiscover ? 'search-input--active' : ''}"
            placeholder="Na mit keresel...?"
            bind:value={searchInputValue}
            bind:this={searchInputEl}
            on:focus={() => (showDiscover = true)}
            on:keydown={handleKeydown}
            autocomplete="off"
        />
        <div class="search-buttons">
            <button class="btn btn-primary" on:click={executeSearch} name="search" id="search-button">
                Na lámsza!
            </button>
            <button
                class="btn btn-kapu"
                class:btn-kapu--active={showDiscover}
                on:click={openKapu}
                aria-label="Kapu megnyitása"
                title="Kapu"
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
              >
                <path d="M3 21V5a2 2 0 0 1 2-2h1a2 2 0 0 1 2 2v16"></path>
                <path d="M16 21V5a2 2 0 0 1 2-2h1a2 2 0 0 1 2 2v16"></path>
                <path d="M8 6l6 3"></path>
                <path d="M8 10l6 3"></path>
                <path d="M8 14l6 3"></path>
                <path d="M8 18l6 3"></path>
              </svg>
                <span class="btn-kapu-label">Kapu</span>
            </button>
            {#if searchInputValue !== ""}
                <button
                    class="btn clear-search-btn"
                    on:click={clearSearch}
                    aria-label="Keresés törlése"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                </button>
            {/if}
        </div>
    </section>

    <section
        class="discover-container"
        class:discover-container--visible={showDiscover}
        aria-label="Keresési eredmények"
    >
        {#if showDiscover}
            <div class="discover-header">
                <span class="info-box">
                    {#if loading}
                        <p>Keresés...</p>
                    {:else if searchResults && searchInputValue}
                        <p>
                            {#if totalCount === 0}
                                <span>Nincs találat erre a keresésre.</span>
                            {:else}
                                🔍 Keresés: <span class="active">{searchInputValue}</span>
                            {/if}
                        </p>
                        <p><span>({totalCount} találat)</span></p>
                    {:else}
                        <p>Írd be a keresett szót, majd kattints a „Na lámsza!" gombra.</p>
                    {/if}
                </span>
                <button
                    class="btn discover-close-btn"
                    on:click={closeDiscover}
                    aria-label="Bezárás"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <line x1="18" y1="6" x2="6" y2="18"></line>
                        <line x1="6" y1="6" x2="18" y2="18"></line>
                    </svg>
                </button>
            </div>

            {#if !loading && searchResults && totalCount > 0}
                <div class="discover-sections">
                    {#if searchResults.locations?.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">📍 Települések</h4>
                            <div class="list flex">
                                {#each searchResults.locations as loc}
                                    {@const entry = locationToEntry(loc)}
                                    <EntryCard entry={entry} />
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if searchResults.entries?.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">📋 Index</h4>
                            <div class="list flex">
                                {#each searchResults.entries as entry}
                                    <EntryCard {entry} />
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if searchResults.events?.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">📅 Események</h4>
                            <div class="discover-event-list">
                                {#each searchResults.events as ev}
                                    <a href="/esemenyek/{ev.id}" class="discover-event-card">
                                        <span class="discover-event-title">{ev.title}</span>
                                        <span class="discover-event-meta">
                                            {formatDateShort(ev.start_date)}
                                            {#if ev.location_name} · {ev.location_name}{/if}
                                        </span>
                                    </a>
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if searchResults.news?.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">📰 Hírek</h4>
                            <div class="discover-news-list">
                                {#each searchResults.news as item}
                                    <a href={item.link} target="_blank" rel="nofollow noopener" class="discover-news-card">
                                        <span class="discover-news-title">{item.title}</span>
                                        <span class="discover-news-source">{item.source}</span>
                                    </a>
                                {/each}
                            </div>
                        </div>
                    {/if}
                </div>
            {/if}

            
                <div class="external-search-links">{#if !loading && searchResults && totalCount === 0 && searchInputValue}
                    <span class="ext-label">Kereshetsz máshol is:</span>
                    <div class="btn-group">
                        <a class="btn btn-md google" href="https://www.google.com/search?q={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">Google</a>
                        <a class="btn btn-md bing" href="https://www.bing.com/search?q={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">Bing</a>
                        <a class="btn btn-md duckduckgo" href="https://duckduckgo.com/?q={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">DuckDuckGo</a>
                        <a class="btn btn-md yahoo" href="https://search.yahoo.com/search?p={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">Yahoo</a>
                    </div>
                    <div class="btn-group">
                        <a class="btn btn-md index" href="/index">Lámsza Index</a>
                    </div>
               {/if}  </div>
           
        {/if}
    </section>
</div>

<style>
/* External search engine links inside results */
.external-search-links {
    margin-top: 1.5rem;
    border-top: 1px solid var(--border-color);
    display: flex;
    flex-wrap: wrap;
    gap: 0 2rem;
    align-items: center;
    justify-content: space-between;
    padding: 1rem;
}

.external-search-links .ext-label {
    font-size: 0.8rem;
    color: var(--text-muted);
    flex-basis: 100%;
    margin-bottom: 1rem;
}

.external-search-links a.google {
    border-color: var(--google-green);
}
.external-search-links a.bing {
    border-color: var(--bing-blue);
}
.external-search-links a.duckduckgo {
    border-color: var(--duckduckgo-orange);
}
.external-search-links a.yahoo {
    border-color: var(--yahoo-purple);
}
.external-search-links a.index {
    border-color: var(--szekely-green);
}

.btn-group {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
}

.btn-kapu:hover {
    background: var(--tab-hover-bg);
    color: var(--text-primary);
}
.btn-kapu--active {
    background: var(--tab-hover-bg);
    color: var(--text-primary);
    border-color: var(--text-muted);
}

/* Discover: hidden by default, revealed smoothly */
.discover-container {
    visibility: hidden;
    opacity: 0;
    max-height: 0;
    overflow: hidden;
    padding: 0;
    border: none;
    box-shadow: none;
    transition: visibility 0.2s, opacity 0.25s ease, max-height 0.3s ease, padding 0.25s ease;
}
.discover-container--visible {
    visibility: visible;
    opacity: 1;
    max-height: 80vh;
    overflow-y: auto;
    border: 1px solid var(--border-color);
    border-top: 0;
    box-shadow: 0 4px 12px var(--shadow-md);
    border-radius: 0 0 12px 12px;
}

.discover-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 1.5rem;
    gap: 1rem;
}
.discover-close-btn {
    flex-shrink: 0;
    background: none;
    border: none;
    color: var(--text-muted);
    cursor: pointer;
    padding: 0.25rem;
    border-radius: 50%;
}
.discover-close-btn:hover {
    background: var(--skeleton-bg);
    color: var(--szekely-red);
}

.discover-sections {
    padding: 0 1.5rem 0;
}
.discover-section {
    margin-bottom: 1.5rem;
}
.discover-section:last-child {
    margin-bottom: 0;
}
.discover-section-title {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-secondary);
    margin: 0 0 0.75rem 0;
}

.discover-event-list,
.discover-news-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}
.discover-event-card,
.discover-news-card {
    display: block;
    padding: 0.6rem 0.8rem;
    background: var(--bg-body);
    border-radius: 8px;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    text-decoration: none;
    transition: background 0.2s, border-color 0.2s;
}
.discover-event-card:hover,
.discover-news-card:hover {
    background: var(--tab-hover-bg);
    border-color: var(--text-muted);
}
.discover-event-title,
.discover-news-title {
    display: block;
    font-weight: 500;
}
.discover-event-meta,
.discover-news-source {
    font-size: 0.85rem;
    color: var(--text-faint);
}
</style>
