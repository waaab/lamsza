<script>
    import { createEventDispatcher, onDestroy, onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import WebsiteCard from "$lib/components/WebsiteCard.svelte";
    import { searchPreferredLocation, sortDirectoryEntries } from "$lib/directoryListingOrder.js";
    import {
        buildSettlementAnswer,
        pickSettlement,
        settlementFoundLine,
    } from "$lib/settlementSearchAnswer.js";
    import AppIcon from "$lib/icons/AppIcon.svelte";
    import { auth } from "$lib/stores/auth";
    import { formatDateShort, weatherIconEmoji } from "$lib/utils";

    const dispatch = createEventDispatcher();

    let showDiscover = false;

    let searchInputValue = "";
    let searchResults = null; // { locations, entries, events, news, attractions, venues, historical_seats, websites, website_query }
    let suggestions = [];
    let loading = false;
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

    $: hasResults = searchResults && (
        filteredLocations.length > 0 ||
        filteredEntries.length > 0 ||
        filteredEvents.length > 0 ||
        (searchResults.news && searchResults.news.length > 0) ||
        filteredAttractions.length > 0 ||
        filteredVenues.length > 0 ||
        (searchResults.historical_seats && searchResults.historical_seats.length > 0) ||
        (searchResults.websites && searchResults.websites.length > 0)
    );
    $: totalCount = searchResults
        ? filteredLocations.length +
          filteredEntries.length +
          filteredEvents.length +
          (searchResults.news?.length || 0) +
          filteredAttractions.length +
          filteredVenues.length +
          (searchResults.historical_seats?.length || 0) +
          (searchResults.websites?.length || 0)
        : 0;
    $: preferredLocation = searchPreferredLocation($auth.preferredLocation, $auth.loggedIn);
    $: townChoices = (locations || [])
        .filter((loc) => String(loc?.type || "") !== "megye" && String(loc?.slug || "").trim())
        .filter((loc) => loc.slug !== preferredLocation?.slug)
        .slice()
        .sort((a, b) => String(a.name || "").localeCompare(String(b.name || ""), "hu"));
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
        if (!searchInputValue.trim()) return;

        loading = true;
        showDiscover = true;
        resetAnswer();
        try {
            const data = await apiFetch(
                `/api/search?q=${encodeURIComponent(searchInputValue)}`,
            );
            searchResults = data;
            presentSettlementAnswer(data, searchInputValue);

            const suggestionsData = await apiFetch(
                `/api/autosuggest?q=${encodeURIComponent(searchInputValue)}`,
            );
            suggestions = suggestionsData || [];
        } catch (err) {
            console.error("Search error:", err);
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
        resetAnswer();
        dispatch("discoverClose");
    }

    function clearSearch() {
        searchInputValue = "";
        searchResults = null;
        suggestions = [];
        resetAnswer();
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
                    on:click={clearSearch}
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
                    <AppIcon name="x" size={20} />
                </button>
            </div>

            {#if !loading && searchResults && totalCount > 0}
                <div class="discover-sections">
                    {#if answerSettlement}
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
                    {#if searchResults.website_query && searchResults.websites?.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">Weboldalak</h4>
                            <div class="list flex">
                                {#each searchResults.websites as website (website.id)}
                                    <WebsiteCard {website} />
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if searchResults.entries?.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">📋 Index</h4>
                            <div class="list flex">
                                {#each orderedEntries as entry}
                                    <EntryCard {entry} />
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if !searchResults.website_query && searchResults.websites?.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">Weboldalak</h4>
                            <div class="list flex">
                                {#each searchResults.websites as website (website.id)}
                                    <WebsiteCard {website} />
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if filteredEvents.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">📅 Események</h4>
                            <div class="discover-event-list">
                                {#each filteredEvents as ev}
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

                    {#if filteredVenues.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">🏟 Helyszínek</h4>
                            <div class="discover-venue-list">
                                {#each filteredVenues as venue}
                                    <a
                                        href="/{venue.county_slug}-megye/{venue.settlement_slug}/helyszin/{venue.slug}"
                                        class="discover-venue-card"
                                    >
                                        <span class="discover-venue-title">{venue.name}</span>
                                        <span class="discover-venue-meta">
                                            {venue.settlement_name}{#if venue.kind_label} · {venue.kind_label}{/if}
                                        </span>
                                    </a>
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if filteredAttractions.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title discover-section-title--attractions">🏔 Látnivalók</h4>
                            <div class="discover-attraction-list">
                                {#each filteredAttractions as att}
                                    <a
                                        href="/{att.county_slug}-megye/{att.slug}"
                                        class="discover-attraction-card"
                                    >
                                        <span class="discover-attraction-title">{att.name}</span>
                                        <span class="discover-attraction-meta">{att.county_name}</span>
                                        {#if att.description}
                                            <span class="discover-attraction-desc">{att.description}</span>
                                        {/if}
                                    </a>
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if filteredLocations.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title">📍 Települések</h4>
                            <div class="list flex">
                                {#each filteredLocations as loc}
                                    {@const entry = locationToEntry(loc)}
                                    <EntryCard entry={entry} />
                                {/each}
                            </div>
                        </div>
                    {/if}

                    {#if searchResults.historical_seats?.length > 0}
                        <div class="discover-section">
                            <h4 class="discover-section-title discover-section-title--szek">⚜ Történelmi székek</h4>
                            <div class="discover-szek-list">
                                {#each searchResults.historical_seats as seat}
                                    <a href="/szekek/{seat.slug}" class="discover-szek-card">
                                        <span class="discover-szek-title">{seat.name}</span>
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

                <div class="external-search-links">
                    {#if !loading && searchResults && searchInputValue}
                    <span class="ext-label">Rákereshetsz máshol is:</span>
                    <div class="btn-group">
                        <a class="btn btn-md google" href="https://www.google.com/search?q={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">Google</a>
                        <a class="btn btn-md bing" href="https://www.bing.com/search?q={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">Bing</a>
                        <a class="btn btn-md duckduckgo" href="https://duckduckgo.com/?q={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">DuckDuckGo</a>
                        <a class="btn btn-md yahoo" href="https://search.yahoo.com/search?p={encodeURIComponent(searchInputValue)}" target="_blank" rel="nofollow noopener">Yahoo</a>
                    </div>
                    <div class="btn-group">
                        <a class="btn btn-md index" href="/index">Lámsza Index</a>
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
.external-search-links a.index {
    border-color: var(--szekely-green);
}
.external-search-links a.index:hover {
    background: var(--szekely-green);
    color: var(--white);
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
.search-location-name:hover,
.search-location-clear:hover:not(:disabled) {
    color: var(--text-primary);
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
    color: var(--text-faint);
}

.discover-venue-list,
.discover-attraction-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}
.discover-venue-card,
.discover-attraction-card {
    display: block;
    padding: 0.6rem 0.8rem;
    background: var(--bg-body);
    border-radius: 8px;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    text-decoration: none;
    transition: background 0.2s, border-color 0.2s;
}
.discover-section-title--attractions {
    color: var(--szekely-brown, #6d4c41);
}
.discover-attraction-card {
    border-color: var(--szekely-brown, #8d6e63);
}
.discover-venue-card:hover,
.discover-attraction-card:hover {
    background: var(--tab-hover-bg);
    border-color: var(--text-muted);
}
.discover-venue-title,
.discover-attraction-title {
    display: block;
    font-weight: 500;
}
.discover-venue-meta,
.discover-attraction-meta {
    color: var(--text-faint);
}
.discover-attraction-desc {
    display: block;
    color: var(--text-muted);
    margin-top: 0.25rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.discover-section-title--szek {
    color: var(--szekely-blue, #1565c0);
}
.discover-szek-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
}
.discover-szek-card {
    padding: 0.5rem 0.85rem;
    background: var(--bg-body);
    border-radius: 8px;
    border: 1px solid var(--szekely-blue, #42a5f5);
    color: var(--text-primary);
    text-decoration: none;
    font-weight: 500;
    transition: background 0.2s;
}
.discover-szek-card:hover {
    background: var(--tab-hover-bg);
}
.discover-szek-title {
}
</style>