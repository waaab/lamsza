<script>
    import { createEventDispatcher, onDestroy, onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import SearchResultCard from "$lib/components/SearchResultCard.svelte";
    import { searchResultCardModel } from "$lib/searchResultCard.js";
    import { formatSearchElapsed } from "$lib/searchElapsed.js";
    import { locationMenuTowns, searchPreferredLocation, sortDirectoryEntries } from "$lib/directoryListingOrder.js";
    import {
        buildSettlementAnswer,
        pickSettlement,
        settlementFoundLine,
    } from "$lib/settlementSearchAnswer.js";
    import AppIcon from "$lib/icons/AppIcon.svelte";
    import { auth } from "$lib/stores/auth";
    import { weatherIconEmoji } from "$lib/utils";

    const dispatch = createEventDispatcher();

    let showDiscover = false;

    let searchInputValue = "";
    let searchResults = null; // { locations, entries, events, news, attractions, venues, historical_seats, websites, words, website_query }
    let loading = false;
    /** @type {number | null} */
    let searchElapsedMs = null;
    let searchInputEl;
    let searchRootEl;
    let answerSettlement = null;
    let answerFull = "";
    let answerShown = "";
    let answerTimer = null;
    let answerGen = 0;
    /** @type {{ slug: string, name: string, county_slug: string } | null} */
    let selectedLocation = null;
    let locationMenuOpen = false;
    let locationFieldEl;
    /** @type {Array<Record<string, any>>} */
    let locations = [];
    /** @type {"index" | "services" | "websites" | "szotar"} */
    let resultFilter = "index";
    let searchGen = 0;
    // Results belong to the last submitted query. Clearing the box drops them
    // and retires any search still in flight, so a late response cannot redraw them.
    $: if (!searchInputValue.trim() && (searchResults || loading || searchElapsedMs != null)) {
        searchGen += 1;
        searchResults = null;
        loading = false;
        searchElapsedMs = null;
        resetAnswer();
    }
    $: preferredLocation = searchPreferredLocation($auth.preferredLocation, $auth.loggedIn);
    $: townChoices = locationMenuTowns(locations, preferredLocation);
    $: selectedSlug = selectedLocation?.slug || "";
    $: selectedCounty = selectedLocation?.county_slug || "";
    $: filteredEntries = (searchResults?.entries || []).filter((entry) =>
        placeMatches(entry.location_slug, selectedSlug),
    );
    $: orderedEntries = sortDirectoryEntries(filteredEntries, { sortMode: "title" });
    $: filteredVenues = (searchResults?.venues || []).filter((venue) =>
        placeMatches(venue.settlement_slug, selectedSlug),
    );
    $: filteredEvents = (searchResults?.events || []).filter((event) =>
        placeMatches(event.location_slug, selectedSlug),
    );
    $: filteredLocations = (searchResults?.locations || []).filter((loc) =>
        !selectedSlug || loc.slug === selectedSlug,
    );
    $: filteredAttractions = (searchResults?.attractions || []).filter((item) =>
        !selectedCounty || item.county_slug === selectedCounty,
    );
    $: shownWords =
        resultFilter === "services" || resultFilter === "websites"
            ? []
            : (searchResults?.words || []);
    $: shownWebsites =
        resultFilter === "services" || resultFilter === "szotar"
            ? []
            : (searchResults?.websites || []);
    $: indexEntries =
        resultFilter === "websites" || resultFilter === "szotar"
            ? []
            : orderedEntries;
    $: showBrowseSections = resultFilter === "index";
    $: hasResults = searchResults && (
        (showBrowseSections && filteredLocations.length > 0) ||
        indexEntries.length > 0 ||
        (showBrowseSections && filteredEvents.length > 0) ||
        (showBrowseSections && (searchResults.news?.length || 0) > 0) ||
        (showBrowseSections && filteredAttractions.length > 0) ||
        (showBrowseSections && filteredVenues.length > 0) ||
        (showBrowseSections && (searchResults.historical_seats?.length || 0) > 0) ||
        shownWebsites.length > 0 ||
        shownWords.length > 0
    );
    $: totalCount = searchResults
        ? (showBrowseSections ? filteredLocations.length : 0) +
          indexEntries.length +
          (showBrowseSections ? filteredEvents.length : 0) +
          (showBrowseSections ? (searchResults.news?.length || 0) : 0) +
          (showBrowseSections ? filteredAttractions.length : 0) +
          (showBrowseSections ? filteredVenues.length : 0) +
          (showBrowseSections ? (searchResults.historical_seats?.length || 0) : 0) +
          shownWebsites.length +
          shownWords.length
        : 0;

    function placeMatches(slug, selected) {
        if (!selected) return true;
        return String(slug || "") === selected;
    }

    function toggleLocationMenu() {
        locationMenuOpen = !locationMenuOpen;
    }

    function chooseLocation(loc) {
        const slug = String(loc?.slug || "").trim();
        if (!slug) return;
        // Local result filter only. The saved place is set in user settings.
        selectedLocation = {
            slug,
            name: String(loc?.name || "").trim() || slug,
            county_slug: String(loc?.county_slug || "").trim(),
        };
        locationMenuOpen = false;
    }

    function clearSelectedLocation(event) {
        event?.stopPropagation();
        selectedLocation = null;
        locationMenuOpen = false;
    }

    function selectResultFilter(kind) {
        resultFilter = kind;
    }

    $: searchScope = resultFilter === "services"
        ? "szolgáltatásokban:"
        : resultFilter === "websites"
          ? "weboldalakban:"
          : resultFilter === "szotar"
            ? "szótárban:"
            : selectedLocation?.name
              ? `${selectedLocation.name} és környéke:`
              : "mindenhol:";

    function stopAnswerTyping() {
        if (answerTimer) clearInterval(answerTimer);
        answerTimer = null;
    }

    function resetAnswer() {
        answerGen += 1;
        stopAnswerTyping();
        answerSettlement = null;
        answerFull = "";
        answerShown = "";
    }

    function prefersReducedMotion() {
        return typeof window !== "undefined"
            && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    }

    function syncAnswer(text) {
        answerFull = text;
        if (prefersReducedMotion()) {
            stopAnswerTyping();
            answerShown = text;
            return;
        }
        if (!text.startsWith(answerShown)) answerShown = "";
        if (answerTimer) return;
        answerTimer = setInterval(() => {
            if (answerShown.length >= answerFull.length) {
                stopAnswerTyping();
                return;
            }
            answerShown = answerFull.slice(0, answerShown.length + 1);
        }, 24);
    }

    function presentSettlementAnswer(data, query) {
        const settlement = pickSettlement(data?.locations, query);
        if (!settlement) {
            resetAnswer();
            return;
        }
        const events = (data.events || []).filter(
            (event) => event?.location_slug === settlement.slug,
        );
        answerSettlement = settlement;
        const gen = ++answerGen;
        syncAnswer(buildSettlementAnswer(settlement, { events }));
        apiFetch(`/api/weather?slug=${encodeURIComponent(settlement.slug)}`)
            .then((weather) => {
                if (gen !== answerGen || weather?.temp == null) return;
                syncAnswer(buildSettlementAnswer(settlement, {
                    events,
                    weather: {
                        temp: weather.temp,
                        desc: weather.desc || "",
                        emoji: weatherIconEmoji(weather.icon),
                    },
                }));
            })
            .catch(() => {});
    }

    async function executeSearch() {
        const query = searchInputValue.trim();
        if (!query) return;

        const gen = ++searchGen;
        const started = typeof performance !== "undefined" ? performance.now() : Date.now();
        loading = true;
        showDiscover = true;
        searchElapsedMs = null;
        resetAnswer();
        try {
            const data = await apiFetch(
                `/api/search?q=${encodeURIComponent(query)}`,
            );
            if (gen !== searchGen) return;
            searchResults = data;
            searchElapsedMs = (typeof performance !== "undefined" ? performance.now() : Date.now()) - started;
            presentSettlementAnswer(data, query);
        } catch (err) {
            if (gen !== searchGen) return;
            console.error("Search error:", err);
            searchElapsedMs = (typeof performance !== "undefined" ? performance.now() : Date.now()) - started;
            searchResults = {
                locations: [],
                entries: [],
                events: [],
                news: [],
                attractions: [],
                venues: [],
                historical_seats: [],
                websites: [],
            };
        } finally {
            if (gen === searchGen) loading = false;
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
        searchElapsedMs = null;
        resultFilter = "index";
        resetAnswer();
        dispatch("discoverClose");
    }

    onMount(() => {
        (async () => {
            try {
                const locs = await apiFetch("/api/locations");
                locations = Array.isArray(locs) ? locs : [];
            } catch {
                locations = [];
            }
        })();
    });

    onDestroy(stopAnswerTyping);

    function handleKeydown(e) {
        if (e.key === "Enter") executeSearch();
        if (e.key === "Escape") closeDiscover();
    }

    function handleWindowPointerDown(event) {
        const target = event.target;
        if (locationMenuOpen && (!(target instanceof Node) || !locationFieldEl?.contains(target))) {
            locationMenuOpen = false;
        }
        if (!showDiscover || searchInputValue.trim()) return;
        if (!(target instanceof Node) || searchRootEl?.contains(target)) return;
        closeDiscover();
    }

</script>

<svelte:window on:pointerdown={handleWindowPointerDown} />

<div
    class="search-discover-wrapper"
    class:search-discover-wrapper--expanded={showDiscover}
    bind:this={searchRootEl}
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
            <div class="search-location" bind:this={locationFieldEl}>
                <button
                    type="button"
                    class="search-location-name"
                    class:search-location-name--set={!!selectedLocation}
                    aria-expanded={locationMenuOpen}
                    aria-haspopup="listbox"
                    on:click={toggleLocationMenu}
                >
                    <svg
                        class="search-location-icon"
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        aria-hidden="true"
                    >
                        <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path>
                        <circle cx="12" cy="10" r="3"></circle>
                    </svg>
                    <span>{selectedLocation?.name || "Település"}</span>
                </button>
                <button
                    type="button"
                    class="search-location-clear"
                    aria-label="Település szűrő törlése"
                    title="Település szűrő törlése"
                    disabled={!selectedLocation}
                    on:click={clearSelectedLocation}
                >
                    <AppIcon name="x" size={14} />
                </button>
                {#if locationMenuOpen}
                    <ul class="search-location-menu" role="listbox">
                        {#if preferredLocation}
                            <li>
                                <button
                                    type="button"
                                    role="option"
                                    aria-selected={selectedLocation?.slug === preferredLocation.slug}
                                    on:click={() => chooseLocation(preferredLocation)}
                                >
                                    <span class="search-location-option-label">Településem</span>
                                    <span>{preferredLocation.name}</span>
                                </button>
                            </li>
                        {/if}
                        {#each townChoices as town (town.slug + town.county_slug)}
                            <li>
                                <button
                                    type="button"
                                    role="option"
                                    aria-selected={selectedLocation?.slug === town.slug}
                                    on:click={() => chooseLocation(town)}
                                >
                                    <span>{town.name}</span>
                                    {#if town.county}
                                        <span class="search-location-option-meta">{town.county}</span>
                                    {/if}
                                </button>
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
            <button class="btn btn-primary" on:click={executeSearch} name="search" id="search-button">
                Na lámsza!
            </button>
            {#if searchInputValue !== ""}
                <button
                    class="btn clear-search-btn"
                    on:click={closeDiscover}
                    aria-label="Keresés törlése"
                >
                    <AppIcon name="x" size={20} />
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
                            🔍 Keresés
                            {#if resultFilter === "index" && selectedLocation}
                                <span class="search-place">{searchScope}</span>
                            {:else}
                                {searchScope}
                            {/if}
                            <span class="active">{searchInputValue}</span>
                        </p>
                        <p class:search-result-count={totalCount > 0}>
                            {#if totalCount === 0}
                                <span>Nincs találat erre a keresésre.</span>
                            {:else}
                                <span>{totalCount} találat{#if searchElapsedMs != null}, {formatSearchElapsed(searchElapsedMs)} alatt{/if}</span>
                            {/if}
                        </p>
                    {:else}
                        <p>Írd be a keresett szót, majd kattints a „Na lámsza!" gombra.</p>
                    {/if}
                </span>
            </div>

            {#if !loading && searchResults && searchInputValue.trim() && totalCount > 0}
                <div class="discover-sections">
                    {#if answerSettlement && resultFilter === "index"}
                        <div class="discover-answer">
                            <p class="discover-answer-text">
                                {answerShown}<span
                                    class="discover-answer-caret"
                                    class:discover-answer-caret--done={answerShown.length >= answerFull.length}
                                    aria-hidden="true"
                                ></span>
                            </p>
                            {#if answerFull && answerShown.length >= answerFull.length}
                                <p class="discover-answer-follow">{settlementFoundLine(answerSettlement.name)}</p>
                            {/if}
                        </div>
                    {/if}
                    {#if shownWebsites.length > 0 && (resultFilter === "websites" || searchResults.website_query)}
                        <div class="discover-section">
                            <h4 class="discover-section-title">
                                <AppIcon name="websites" size={18} />
                                Weboldalak
                            </h4>
                            <div class="discover-result-list">
                                {#each shownWebsites as website (website.id)}
                                    {@const card = searchResultCardModel("website", website)}
                                    {#if card}
                                        <SearchResultCard {...card} />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if indexEntries.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">
                                <AppIcon name="entries" size={18} />
                                Szolgáltatások
                            </h4>
                            <div class="discover-result-list">
                                {#each indexEntries as entry (entry.id)}
                                    {@const card = searchResultCardModel("service", entry)}
                                    {#if card}
                                        <SearchResultCard {...card} />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if shownWebsites.length > 0 && resultFilter !== "websites" && !searchResults.website_query}
                        <div class="discover-section">
                            <h4 class="discover-section-title">
                                <AppIcon name="websites" size={18} />
                                Weboldalak
                            </h4>
                            <div class="discover-result-list">
                                {#each shownWebsites as website (website.id)}
                                    {@const card = searchResultCardModel("website", website)}
                                    {#if card}
                                        <SearchResultCard {...card} />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if showBrowseSections && filteredEvents.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">
                                <AppIcon name="events" size={18} />
                                Események
                            </h4>
                            <div class="discover-result-list">
                                {#each filteredEvents as ev (ev.id)}
                                    {@const card = searchResultCardModel("event", ev)}
                                    {#if card}
                                        <SearchResultCard {...card} />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if showBrowseSections && filteredVenues.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">
                                <AppIcon name="venues" size={18} />
                                Helyszínek
                            </h4>
                            <div class="discover-result-list">
                                {#each filteredVenues as venue (venue.id)}
                                    {@const card = searchResultCardModel("venue", venue)}
                                    {#if card}
                                        <SearchResultCard {...card} />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if showBrowseSections && filteredAttractions.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title discover-section-title--attractions">
                                <AppIcon name="attractions" size={18} />
                                Látnivalók
                            </h4>
                            <div class="discover-result-list">
                                {#each filteredAttractions as att (att.id)}
                                    {@const card = searchResultCardModel("attraction", att)}
                                    {#if card}
                                        <SearchResultCard {...card} />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if showBrowseSections && filteredLocations.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">
                                <AppIcon name="locations" size={18} />
                                Települések
                            </h4>
                            <div class="discover-result-list">
                                {#each filteredLocations as loc (loc.id)}
                                    {@const card = searchResultCardModel("settlement", loc)}
                                    {#if card}
                                        <SearchResultCard {...card} />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if showBrowseSections && searchResults.historical_seats?.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title discover-section-title--szek">
                                <AppIcon name="szekek" size={18} />
                                Történelmi székek
                            </h4>
                            <div class="discover-result-list">
                                {#each searchResults.historical_seats as seat (seat.id)}
                                    {@const card = searchResultCardModel("seat", seat)}
                                    {#if card}
                                        <SearchResultCard {...card} />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if showBrowseSections && searchResults.news?.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">
                                <AppIcon name="newsfeeds" size={18} />
                                Hírek
                            </h4>
                            <div class="discover-result-list">
                                {#each searchResults.news as item (item.link)}
                                    {@const card = searchResultCardModel("news", item)}
                                    {#if card}
                                        <SearchResultCard {...card} />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if shownWords.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">
                                <AppIcon name="szotar" size={18} />
                                Székely Szótár
                            </h4>
                            <div class="discover-result-list">
                                {#each shownWords as word (word.id)}
                                    {@const card = searchResultCardModel("word", word)}
                                    {#if card}
                                        <SearchResultCard {...card} />
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/if}
                </div>
            {/if}

                <div class="external-search-links">
                    {#if !loading && searchResults && searchInputValue}
                    <span class="ext-label">Rákereshetsz máshol is:</span>
                    <div class="btn-group">
                        <a class="btn btn-md google" href="https://www.google.com/search?q={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">Google</a>
                        <a class="btn btn-md bing" href="https://www.bing.com/search?q={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">Bing</a>
                        <a class="btn btn-md duckduckgo" href="https://duckduckgo.com/?q={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">DuckDuckGo</a>
                        <a class="btn btn-md yahoo" href="https://search.yahoo.com/search?p={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">Yahoo</a>
                    </div>
                    <div class="btn-group result-filters" role="group" aria-label="Találatok szűrése">
                        <button
                            type="button"
                            class="btn btn-md nav-btn"
                            class:active={resultFilter === "index"}
                            aria-pressed={resultFilter === "index"}
                            on:click={() => selectResultFilter("index")}
                        >
                            Lámsza Index
                        </button>
                        <button
                            type="button"
                            class="btn btn-md nav-btn"
                            class:active={resultFilter === "services"}
                            aria-pressed={resultFilter === "services"}
                            on:click={() => selectResultFilter("services")}
                        >
                            Szolgáltatások
                        </button>
                        <button
                            type="button"
                            class="btn btn-md nav-btn"
                            class:active={resultFilter === "websites"}
                            aria-pressed={resultFilter === "websites"}
                            on:click={() => selectResultFilter("websites")}
                        >
                            Weboldalak
                        </button>
                        <button
                            type="button"
                            class="btn btn-md nav-btn"
                            class:active={resultFilter === "szotar"}
                            aria-pressed={resultFilter === "szotar"}
                            on:click={() => selectResultFilter("szotar")}
                        >
                            Székely Szótár
                        </button>
                    </div>
                    {/if}
                </div>
        {/if}
    </section>
</div>

<style>
/* External search engine links inside results */
.external-search-links {
    display: flex;
    flex-wrap: wrap;
    gap: 0 2rem;
    align-items: center;
    justify-content: space-between;
    padding: 1rem 1.5rem;
}

.external-search-links .ext-label {
    color: var(--text-muted);
    flex-basis: 100%;
    margin-bottom: 1rem;
    border-top: 1px dashed var(--border-color);
    padding-top: 1rem;
}

.external-search-links a.google {
    border-color: var(--google-green);

}
.external-search-links a.google:hover {
    background: var(--google-green);
    color: var(--white);
}
.external-search-links a.bing {
    border-color: var(--bing-blue);
}
.external-search-links a.bing:hover {
    background: var(--bing-blue);
    color: var(--white);
}
.external-search-links a.duckduckgo {
    border-color: var(--duckduckgo-orange);
}
.external-search-links a.duckduckgo:hover {
    background: var(--duckduckgo-orange);
    color: var(--white);
}
.external-search-links a.yahoo {
    border-color: var(--yahoo-purple);
}
.external-search-links a.yahoo:hover {
    background: var(--yahoo-purple);
    color: var(--white);
}
.result-filters {
    margin-left: auto;
    justify-content: flex-end;
}

.btn-group {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
}

.btn-primary {
  background-color: var(--szekely-red);
  color: var(--white);
  border-color: var(--szekely-red);
}

.btn-primary:hover {
  background: var(--szekely-red);
  color: var(--white);
  border-color: var(--szekely-red);
}

.search-location {
    position: relative;
    display: flex;
    align-items: center;
    align-self: stretch;
    gap: 0.2rem;
    max-width: 12.5rem;
}
.search-location::before {
    content: "";
    align-self: stretch;
    width: 1px;
    margin: 0.65rem 0.7rem 0.65rem 0;
    background: var(--thin-grey);
}
.search-location-name,
.search-location-clear {
    background: none;
    border: none;
    padding: 0;
    color: var(--text-secondary);
    font: inherit;
    cursor: pointer;
}
.search-location-name {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    min-width: 0;
    font-size: var(--text-sm);
}
.search-location-name span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.search-location-icon {
    width: 0.9rem;
    height: 0.9rem;
    flex-shrink: 0;
}
.search-location-name--set {
    color: #4f4f4f;
    font-weight: 600;
}
.search-location-name:hover,
.search-location-clear:hover:not(:disabled) {
    color: var(--text-primary);
}
.search-location-name--set:hover {
    color: #333333;
}
.search-location-clear {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    color: var(--text-muted);
    padding: 0 0.55rem;
}
.search-location-clear:disabled {
    color: var(--thin-grey);
    cursor: default;
    opacity: 1;
}
.search-location-menu {
    position: absolute;
    top: calc(100% + 0.75rem);
    right: 0;
    z-index: 30;
    min-width: 16rem;
    max-height: 18rem;
    overflow: auto;
    margin: 0;
    padding: 0.25rem;
    list-style: none;
    background: var(--card-bg);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    box-shadow: 0 8px 24px var(--shadow-md);
}
.search-location-menu button {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    width: 100%;
    background: none;
    border: none;
    text-align: left;
    padding: 0.45rem 0.6rem;
    border-radius: 6px;
    color: var(--text-primary);
    cursor: pointer;
    font: inherit;
}
.search-location-menu button:hover,
.search-location-menu button[aria-selected="true"] {
    background: var(--tab-hover-bg);
}
.search-location-option-label,
.search-location-option-meta {
    color: var(--text-muted);
    font-size: var(--text-xs, 0.75rem);
}

/* Discover: hidden by default, revealed smoothly */
.discover-container {
    visibility: hidden;
    opacity: 0;
    padding: 0;
    border: none;
    box-shadow: none;
    transition: visibility 0.2s, opacity 0.25s ease, max-height 0.3s ease, padding 0.25s ease;
}
.discover-container--visible {
    visibility: visible;
    opacity: 1;
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
.search-place {
    font-weight: 600;
}
.search-result-count {
    color: var(--text-faintest);
    font-weight: 400;
}

.discover-sections {
    padding: 0 1.5rem 0;
}
.discover-answer {
    margin: 0 0 1.25rem;
    padding: 0.85rem 1rem;
    border-left: 3px solid var(--szekely-red, #c0392b);
    background: var(--bg-body);
    border-radius: 0 8px 8px 0;
}
.discover-answer-text {
    margin: 0;
    line-height: 1.6;
    min-height: 1.6em;
}
.discover-answer-caret {
    display: inline-block;
    width: 0.55ch;
    height: 1.05em;
    margin-left: 1px;
    background: currentColor;
    vertical-align: text-bottom;
    animation: discover-caret 1s steps(1) infinite;
}
.discover-answer-caret--done {
    display: none;
}
.discover-answer-follow {
    margin: 0.85rem 0 0;
    color: var(--text-secondary);
}
@keyframes discover-caret {
    50% { opacity: 0; }
}
.discover-section {
    margin-bottom: 1.5rem;
}
.discover-section:last-child {
    margin-bottom: 0;
}
.discover-section-title {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-weight: 600;
    color: var(--text-secondary);
    margin: 0 0 0.75rem 0;
}
.discover-section-title :global(svg) {
    stroke: currentColor;
    fill: none;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
}
.discover-section-title :global(svg polygon.app-icon-solid) {
    fill: currentColor;
    stroke: none;
}

.discover-section-title--attractions {
    color: var(--szekely-brown, #6d4c41);
}

.discover-section-title--szek {
    color: var(--szekely-blue, #1565c0);
}

.discover-result-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}
</style>