<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import EntryCard from "$lib/components/EntryCard.svelte";

    let searchInputValue = "";
    let searchResults = null;
    let suggestions = [];
    let loading = false;

    async function executeSearch() {
        if (!searchInputValue.trim()) return;

        loading = true;
        try {
            const data = await apiFetch(
                `/api/directory?q=${encodeURIComponent(searchInputValue)}`,
            );
            searchResults = data;

            // Fetch suggestions in parallel
            const suggestionsData = await apiFetch(
                `/api/autosuggest?q=${encodeURIComponent(searchInputValue)}`,
            );
            suggestions = suggestionsData || [];
        } catch (err) {
            console.error("Search error:", err);
        } finally {
            loading = false;
        }
    }

    function clearSearch() {
        searchInputValue = "";
        searchResults = null;
        suggestions = [];
    }

    function handleKeydown(e) {
        if (e.key === "Enter") executeSearch();
        if (e.key === "Escape") clearSearch();
    }
</script>

<section class="search-container">
    <input
        type="text"
        class="search-input {searchResults ? 'search-input--active' : ''}"
        placeholder="Mit keresel máma...?"
        bind:value={searchInputValue}
        on:keydown={handleKeydown}
        autocomplete="off"
    />
    <div class="search-buttons">
        {#if searchInputValue !== ""}
            <button
                class="clear-search-btn"
                on:click={clearSearch}
                aria-label="Keresés törlése"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="20"
                    height="20"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><line x1="18" y1="6" x2="6" y2="18"></line><line
                        x1="6"
                        y1="6"
                        x2="18"
                        y2="18"
                    ></line></svg
                >
            </button>
        {/if}
        <button class="btn btn-primary" on:click={executeSearch}
            >Na lámsza!</button
        >
    </div>
</section>

{#if loading}
    <div class="skeleton skeleton-text search-loading-skeleton">Keresés...</div>
{:else if searchResults}
    <section class="results-container">
        <div class="filter-actions">
            <div class="info-box">
                <p>
                    {#if searchResults.length === 0}
                        <span>Nincs találat az Indexben erre a keresésre.</span>
                    {:else}
                        🔍 Keresés: <span class="active"
                            >{searchInputValue}</span
                        >
                    {/if}
                </p>
                <p><span>({searchResults.length} találat)</span></p>
            </div>
        </div>

        {#if searchResults.length > 0}
            <div class="list flex">
                {#each searchResults as entry}
                    <EntryCard {entry} />
                {/each}
            </div>
        {/if}

        <!-- External links simplified for the component -->
        <div class="external-search-links">
            <span class="ext-label">Reakereshetsz máshol es:</span>
            <a
                href="https://www.google.com/search?q=${encodeURIComponent(
                    searchInputValue,
                )}"
                target="_blank"
                rel="nofollow noopener">Google</a
            >
            <a
                href="https://www.bing.com/search?q=${encodeURIComponent(
                    searchInputValue,
                )}"
                target="_blank"
                rel="nofollow noopener">Bing</a
            >
            <a
                href="https://duckduckgo.com/?q=${encodeURIComponent(
                    searchInputValue,
                )}"
                target="_blank"
                rel="nofollow noopener">DuckDuckGo</a
            >
        </div>
    </section>
{/if}

<style>
    /* Search specific styles that aren't global yet */
    .search-loading-skeleton {
        margin: 2rem auto;
        max-width: 600px;
    }
</style>
