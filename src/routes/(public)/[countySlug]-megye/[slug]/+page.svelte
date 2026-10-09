<script>
    import { onMount } from "svelte";
    import { get } from "svelte/store";
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import { auth } from "$lib/stores/auth";
    import { openLogin } from "$lib/openLogin.js";
    import FavoriteButton from "$lib/components/FavoriteButton.svelte";
    import {
        addFavorite,
        favoriteListWithToggle,
        isFavorite,
        loadFavoriteList,
        removeFavorite,
    } from "$lib/favorites.js";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";
    import NewsWidget from "$lib/components/NewsWidget.svelte";
    import EventsWidget from "$lib/components/EventsWidget.svelte";
    import { apiFetch, getApiBase } from "$lib/api";
    import Markdown from "$lib/components/Markdown.svelte";
    import CrestShieldPlaceholder from "$lib/components/CrestShieldPlaceholder.svelte";
    import EntryPhotoGallery from "$lib/components/EntryPhotoGallery.svelte";
    import { attractionGallerySlides } from "$lib/entryPhotos.js";
    import { kindLabel } from "$lib/venueKindLabels.js";
    import AttractionSuggestionDialog from "$lib/components/AttractionSuggestionDialog.svelte";
    import NoticeDialog from "$lib/components/NoticeDialog.svelte";
    import LandscapeBackdrop from "$lib/components/LandscapeBackdrop.svelte";
    import PlaceWeatherCard from "$lib/components/PlaceWeatherCard.svelte";
    import { attractionFacts } from "$lib/attractionFacts.js";

    let settlementData = null;
    let attractionData = null;
    let countyAttractions = [];
    let nearbyAttractions = [];
    let nearbyScope = "";
    /** @type {Record<string, unknown>[]} */
    let settlementVenues = [];
    let entries = [];
    let loading = true;
    let entriesError = null;
    /** @type {{ type: string, id: number }[]} */
    let favoriteList = [];

    async function initFavorites() {
        await auth.init();
        if (!get(auth).loggedIn) {
            favoriteList = [];
            return;
        }
        try {
            favoriteList = await loadFavoriteList();
        } catch {
            favoriteList = [];
        }
    }

    /** @param {string} type @param {number} id @param {boolean} currentlyActive */
    async function handleFavoriteToggle(type, id, currentlyActive) {
        if (!get(auth).loggedIn) {
            openLogin();
            return;
        }
        try {
            if (currentlyActive) {
                await removeFavorite(type, id);
                favoriteList = favoriteListWithToggle(
                    favoriteList,
                    type,
                    id,
                    false,
                );
            } else {
                await addFavorite(type, id);
                favoriteList = favoriteListWithToggle(
                    favoriteList,
                    type,
                    id,
                    true,
                );
            }
        } catch (err) {
            console.error(err);
        }
    }

    onMount(() => {
        initFavorites();
    });

    let viewMode = "grid";
    let sortMode = "title";
    let visibleCount = 12;
    let sortOpen = false;
    let suggestionOpen = false;
    let suggestionSent = false;

    const sortLabels = { title: "Név (A→Z)", newest: "Legújabb" };

    /** @param {unknown} crest */
    function hasCrestUrl(crest) {
        const s = String(crest ?? "").trim();
        return s.length > 5 && s !== "–";
    }

    function setSortMode(mode) {
        sortMode = mode;
        sortOpen = false;
    }

    $: town = $page.params.slug;
    $: sortedEntries = [...entries].sort((a, b) => {
        if (sortMode === "newest") return b.id - a.id;
        return a.name.localeCompare(b.name);
    });
    $: totalCount = sortedEntries.length;
    $: displayItems = sortedEntries.slice(0, visibleCount);

    function openAttractionSuggestion() {
        if (!get(auth).loggedIn) {
            openLogin();
            return;
        }
        if (attractionData?.suggestion_pending) return;
        suggestionOpen = true;
    }

    function contributorInitial(name) {
        const text = String(name ?? "").trim();
        return text ? text.charAt(0).toLocaleUpperCase("hu") : "?";
    }

    function loadMore() {
        visibleCount += 12;
    }

    $: if (browser && $page.params.slug) {
        fetchData();
    }

    $: pageTitle = settlementData?.name || attractionData?.name || town;
    $: isAttraction = !!attractionData;
    $: attractionStats = attractionFacts(attractionData);
    /** The crest URL that failed to load (then the default shield shows). */
    let crestFailedFor = "";
    $: attractionActivities = Array.isArray(attractionData?.activities)
        ? attractionData.activities.map((item) => String(item ?? "").trim()).filter(Boolean)
        : [];
    $: attractionProhibitions = Array.isArray(attractionData?.prohibitions)
        ? attractionData.prohibitions.map((item) => String(item ?? "").trim()).filter(Boolean)
        : [];
    $: attractionSlides = attractionData
        ? attractionGallerySlides(attractionData, { apiBase: getApiBase() })
        : [];

    async function fetchData() {
        loading = true;
        settlementData = null;
        attractionData = null;
        countyAttractions = [];
        nearbyAttractions = [];
        nearbyScope = "";
        settlementVenues = [];
        entries = [];
        entriesError = null;
        try {
            const countySlug = $page.params.countySlug.toLowerCase();
            const slug = town.toLowerCase();

            // 1. Try settlement first (locations = counties + settlements)
            const locations = await apiFetch(
                `/api/locations?county_slug=${encodeURIComponent(countySlug)}`,
            );
            const locData = locations.find((l) => l.slug === slug && l.type !== "megye");
            if (locData) {
                settlementData = locData;
                if (locData.parent_id) {
                    const parent = locations.find((l) => l.id === locData.parent_id);
                    if (parent) settlementData.parent = parent;
                }
                const res = await apiFetch(
                    `/api/directory?location_slug=${encodeURIComponent(slug)}`,
                );
                entries = res || [];
                try {
                    countyAttractions = await apiFetch(
                        `/api/attractions?county_slug=${encodeURIComponent(countySlug)}`,
                    );
                    if (!Array.isArray(countyAttractions)) countyAttractions = [];
                } catch {
                    countyAttractions = [];
                }
                try {
                    const vlist = await apiFetch(
                        `/api/venues?settlement_id=${locData.id}`,
                    );
                    settlementVenues = Array.isArray(vlist) ? vlist : [];
                } catch {
                    settlementVenues = [];
                }
            } else {
                // 2. Try attraction
                const att = await apiFetch(
                    `/api/attractions?county_slug=${encodeURIComponent(countySlug)}&slug=${encodeURIComponent(slug)}`,
                );
                if (att && att.id) {
                    attractionData = att;
                    try {
                        const near = await apiFetch(`/api/attractions?near_id=${encodeURIComponent(att.id)}`);
                        nearbyAttractions = Array.isArray(near?.attractions) ? near.attractions : [];
                        nearbyScope = typeof near?.scope === "string" ? near.scope : "";
                    } catch {
                        nearbyAttractions = [];
                        nearbyScope = "";
                    }
                }
            }
        } catch (err) {
            console.error(err);
            entriesError = "Nem sikerült betölteni az adatokat.";
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head>
    <title>{pageTitle} - Index</title>
</svelte:head>

{#if attractionData}
  <header class="page-header page-landscape page-landscape--action">
    <LandscapeBackdrop place={$page.params.slug} />
    <Breadcrumbs
        label={attractionData.name}
        parentLabel={attractionData.county_name}
        parentUrl="/{$page.params.countySlug}-megye"
        countyName={attractionData.county_name}
        countySlug={$page.params.countySlug}
    />

    <div class="location-title-row">
        <h1 class="page-title">{attractionData.name}</h1>
        <FavoriteButton
            type="attraction"
            id={attractionData.id}
            active={isFavorite(favoriteList, "attraction", attractionData.id)}
            ontoggle={() =>
                handleFavoriteToggle(
                    "attraction",
                    attractionData.id,
                    isFavorite(favoriteList, "attraction", attractionData.id),
                )}
        />
    </div>
    <h2 class="greeting">
        {#if attractionData.description}
            {attractionData.description}
        {:else}
            Látnivaló {attractionData.county_name} megyében.
        {/if}
    </h2>
  </header>

    <div class="widgets-box widgets-box--panels widgets-box--attraction">
        <div id="attekintes" class="widget">
            <div class="widget-header">
                <h3 class="widget-title">Áttekintés</h3>
            </div>
            <div class="more-info">
                {#if attractionData.name_ro}
                    <span>Románul: <span>{attractionData.name_ro}</span></span>
                {/if}
                {#if attractionData.name_de}
                    <span>Németül: <span>{attractionData.name_de}</span></span>
                {/if}
                {#if attractionData.latitude && attractionData.longitude}
                    <span>Koordináták: <span>{attractionData.latitude.toFixed(4)}, {attractionData.longitude.toFixed(4)}</span></span>
                {/if}
                <span>Megye: <span><a href="/{$page.params.countySlug}-megye" class="parent-city-link">{attractionData.county_name}</a></span></span>
            </div>
            {#if attractionStats.length}
                <dl class="attraction-stats">
                    {#each attractionStats as f (f.key)}
                        <div class="attraction-stat"><dt>{f.label}</dt><dd>{f.value}</dd></div>
                    {/each}
                </dl>
            {/if}
        </div>

        <PlaceWeatherCard slug={$page.params.slug} />
    </div>

    {#if attractionSlides.length}
        <section class="attraction-photos" aria-labelledby="attraction-photos-title">
            <h3 id="attraction-photos-title" class="widget-title">Fotók</h3>
            <EntryPhotoGallery
                slides={attractionSlides}
                label={`${attractionData.name} fotói`}
            />
        </section>
    {/if}

    {#if attractionData.content}
        <div class="attraction-content">
            <Markdown source={attractionData.content} />
        </div>
    {/if}

    <section id="tevekenysegek" aria-labelledby="attraction-activities-title">
        <div class="event-widget widget">
            <div class="widget-header">
                <h3 id="attraction-activities-title" class="widget-title">
                    Mit lehet itt csinálni?
                </h3>
            </div>
            <div class="widget-content">
                {#if attractionActivities.length === 0}
                    <span class="info-box"><p>Még nincs megadott tevékenység.</p></span>
                {:else}
                    <ul class="place-chips">
                        {#each attractionActivities as activity}
                            <li><span class="place-chip place-chip--ok"><svg class="place-chip__mark" viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12.5l4.5 4.5L19 7" /></svg>{activity}</span></li>
                        {/each}
                    </ul>
                {/if}
            </div>
        </div>
    </section>

    <section id="tiltasok" aria-labelledby="attraction-prohibitions-title">
        <div class="event-widget widget">
            <div class="widget-header">
                <h3 id="attraction-prohibitions-title" class="widget-title">
                    Mit nem szabad itt csinálni?
                </h3>
            </div>
            <div class="widget-content">
                {#if attractionProhibitions.length === 0}
                    <span class="info-box"><p>Még nincs megadva, mit nem szabad.</p></span>
                {:else}
                    <ul class="place-chips">
                        {#each attractionProhibitions as item}
                            <li><span class="place-chip place-chip--no"><svg class="place-chip__mark" viewBox="0 0 24 24" aria-hidden="true"><path d="M7 7l10 10M17 7L7 17" /></svg>{item}</span></li>
                        {/each}
                    </ul>
                {/if}
            </div>
        </div>
    </section>

    <section id="kozeli-latnivalok" aria-labelledby="nearby-attractions-title">
        <div class="event-widget widget">
            <div class="widget-header">
                <h3 id="nearby-attractions-title" class="widget-title">Közeli látnivalók</h3>
            </div>
            <div class="widget-content">
                {#if nearbyScope === "vicinity"}
                    <p class="attraction-contributors-lead">A környék falvaiból és községeiből.</p>
                {:else if nearbyScope === "area"}
                    <p class="attraction-contributors-lead">A környező városokból és a tágabb környékről.</p>
                {:else if nearbyScope === "county"}
                    <p class="attraction-contributors-lead">A megyéből, mert közelebb nem találtunk másikat.</p>
                {/if}
                {#if nearbyAttractions.length === 0}
                    <span class="info-box"><p>Nincs más látnivaló a közelben vagy a megyében.</p></span>
                {:else}
                    <ul class="place-chips">
                        {#each nearbyAttractions as att (att.id)}
                            <li>
                                <a class="btn btn-sm place-chip" href="/{att.county_slug}-megye/{att.slug}">{att.name}{#if att.county_slug !== $page.params.countySlug}<span class="place-chip__sub">&nbsp;·&nbsp;{att.county_name}</span>{/if}</a>
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
        </div>
    </section>

    <EventsWidget
        attractionId={attractionData.id}
        nearLat={attractionData.latitude}
        nearLon={attractionData.longitude}
        locationName={attractionData.name}
    />

    <section id="kozremukodok" aria-labelledby="attraction-contributors-title">
        <div class="event-widget widget">
            <div class="widget-header">
                <h3 id="attraction-contributors-title" class="widget-title">
                    Közreműködők<span class="widget-title-count">({(attractionData.contributors || []).length})</span>
                </h3>
                <button
                    type="button"
                    class="btn"
                    disabled={!!attractionData.suggestion_pending}
                    on:click={openAttractionSuggestion}
                >
                    {attractionData.suggestion_pending ? "Javaslat elküldve" : "Javaslat módosításra"}
                </button>
            </div>
            <div class="widget-content">
                <p class="attraction-contributors-lead">
                    Akik szerkesztették ezt az oldalt, vagy elfogadott javaslatukkal hozzájárultak hozzá.
                </p>
                {#if !(attractionData.contributors || []).length}
                    <span class="info-box"><p>Még nincs közreműködő.</p></span>
                {:else}
                    <ul class="attraction-contributor-list">
                        {#each attractionData.contributors as person (person.id)}
                            <li>
                                <span class="attraction-contributor" title={person.name}>
                                    {#if person.picture}
                                        <img src={person.picture} alt="" />
                                    {:else}
                                        <span class="attraction-contributor-fallback" aria-hidden="true">{contributorInitial(person.name)}</span>
                                    {/if}
                                    <span class="sr-only">{person.name}</span>
                                </span>
                            </li>
                        {/each}
                    </ul>
                {/if}
            </div>
        </div>
    </section>

    {#if suggestionOpen}
        <AttractionSuggestionDialog
            attraction={attractionData}
            onClose={() => (suggestionOpen = false)}
            onSent={async () => {
                suggestionOpen = false;
                suggestionSent = true;
                await fetchData();
            }}
        />
    {/if}
    {#if suggestionSent}
        <NoticeDialog onClose={() => (suggestionSent = false)} />
    {/if}
{:else if settlementData}
  <header class="page-header page-landscape page-landscape--action">
    <LandscapeBackdrop place={town} />
    <Breadcrumbs
        label={settlementData.name}
        settlementType={settlementData.type}
        countyName={settlementData.county}
        countySlug={$page.params.countySlug}
    />

    <div class="location-title-row">
        <h1 class="page-title">
            {settlementData.name}
            {settlementData.type} és környéke
        </h1>
        <FavoriteButton
            type="settlement"
            id={settlementData.id}
            active={isFavorite(favoriteList, "settlement", settlementData.id)}
            ontoggle={() =>
                handleFavoriteToggle(
                    "settlement",
                    settlementData.id,
                    isFavorite(favoriteList, "settlement", settlementData.id),
                )}
        />
    </div>
    <p class="greeting">
        Helyi események, hírek, időjárás és címtár {settlementData.name} területén.
    </p>
  </header>

    <div class="widgets-box widgets-box--panels">
        <div id="attekintes" class="widget">
            <div class="widget-header">
                <h3 class="widget-title">Áttekintés</h3>
            </div>
            <div class="more-info">
                <span>Románul: <span>{settlementData.name_ro || "-"}</span></span>
                <span>Németül: <span>{settlementData.name_de || "-"}</span></span>
                <span>Irányítószám: <span>{settlementData.post_code || "-"}</span></span>
                <span>Koordináták: <span>{settlementData.coordinates || "-"}</span></span>
                <span>Lakosság: <span>{settlementData.population ? settlementData.population + " fő" : "-"}</span></span>
                <span>Terület: <span>{settlementData.area ? settlementData.area + " km²" : "-"}</span></span>
                <span>Közigazgatási forma: <span class="capitalize">{settlementData.type || "-"}</span></span>
                <span>Kapcsolódó település: <span>{#if settlementData.parent}<a href="/{settlementData.parent.county_slug}-megye/{settlementData.parent.slug}" class="parent-city-link">{settlementData.parent.name}</a>{:else}-{/if}</span></span>
                <span>Megye: <span>{#if settlementData.county}<a href="/{settlementData.county_slug}-megye" class="parent-city-link">{settlementData.county}</a>{:else}-{/if}</span></span>
            </div>
        </div>

        <div id="cimer" class="crest-card widget">
            <div class="widget-header">
                <h3 class="widget-title">{settlementData.name} címere</h3>
            </div>
            <div class="crest-container">
                {#if hasCrestUrl(settlementData.crest) && crestFailedFor !== settlementData.crest}
                    <!-- A crest the host refuses or that fails to load gives way
                         to the default shield, never a broken image. -->
                    <img
                        src={`${getApiBase()}/api/proxy?url=${encodeURIComponent(settlementData.crest)}`}
                        alt="{settlementData.name} címere"
                        class="crest-img"
                        on:error={() => (crestFailedFor = settlementData.crest)}
                    />
                {:else}
                    <CrestShieldPlaceholder
                        label="{settlementData.name} címere - nincs feltöltött kép, helyőrző pajzs"
                    />
                {/if}
            </div>
        </div>

        <PlaceWeatherCard slug={town} />
    </div>

    <EventsWidget settlementSlug={town} locationName={settlementData.name} />

    {#if settlementVenues.length > 0}
        <section id="helyszinek" class="widget settlement-venues component-box"
            aria-label="Helyszínek"
        >
            <div class="widget-header" title="{settlementData.name}i helyszínek">
                <h3 class="widget-title">Helyszínek</h3>
            </div>

            <div class="widget-content">
                {#if loading}
                    <span class="info-box"><p>Betöltés...</p></span>
                {:else if settlementVenues.length === 0}
                    <span class="info-box"><p>Nincs megjeleníthető rendezvényhelyszín.</p></span>
                {:else}
                
                <ul class="settlements-grid">
                    {#each settlementVenues as ven (ven.id)}
                        <a
                            href="/{$page.params.countySlug}-megye/{town}/helyszin/{ven.slug}"
                            class="card sm settlement"
                        >
                            {ven.name}
                            {#if kindLabel(ven.kind, ven.kind_label)}
                                {' '}{kindLabel(ven.kind, ven.kind_label)}
                            {/if}
                        </a>
                    {/each}
                </ul>
                {/if}
            </div>
        </section>
    {/if}

    <NewsWidget settlementSlug={town} ticker={true} />

    {#if countyAttractions.length > 0}
        <section class="component-box">
            <h2 class="aside-title">
                Látnivalók {settlementData.county} megyében
            </h2>
            <ul class="place-chips">
                {#each countyAttractions as att (att.id)}
                    <li><a class="btn btn-sm place-chip" href="/{$page.params.countySlug}-megye/{att.slug}">{att.name}</a></li>
                {/each}
            </ul>
        </section>
    {/if}

    <h2>{settlementData.name}i címtár - Helyi Index</h2>

    {#if loading}
        <div class="list grid">
            {#each Array(6) as _}
                <article class="card entry-placeholder">
                    <span class="entry-placeholder-cat">adat betöltés...</span>
                    <span class="entry-placeholder-title">adat betöltés...</span>
                    <span class="entry-placeholder-loc">adat betöltés...</span>
                </article>
            {/each}
        </div>
    {:else if entriesError}
        <span class="info-box error">
            <p>{entriesError}</p>
        </span>
    {:else if entries.length === 0}
        <span class="info-box error">
            <p>
                Nincs megjeleníthető bejegyzés {settlementData.name} területén.
            </p>
        </span>
    {:else}
        <div class="filter-actions">
            <span class="info-box">
                <p>💡 Összesen:</p>
                <p><span>({displayItems.length}/{totalCount})</span></p>
            </span>

            <div class="view-mode-toggle">
                <div class="sort-toggle">
                    <button
                        class="btn btn-sm"
                        on:click={() => (sortOpen = !sortOpen)}
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
                            ><line x1="4" y1="6" x2="16" y2="6"></line><line
                                x1="4"
                                y1="12"
                                x2="12"
                                y2="12"
                            ></line><line x1="4" y1="18" x2="8" y2="18"
                            ></line><polyline points="15 15 18 18 21 15"
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
                                on:click={() => setSortMode("title")}
                                >Név (A→Z)</button
                            >
                            <button
                                class:active={sortMode === "newest"}
                                on:click={() => setSortMode("newest")}
                                >Legújabb</button
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
                        ></rect><rect x="14" y="14" width="7" height="7"
                        ></rect><rect x="3" y="14" width="7" height="7"
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
                        ></line><line x1="8" y1="18" x2="21" y2="18"
                        ></line><line x1="3" y1="6" x2="3.01" y2="6"
                        ></line><line x1="3" y1="12" x2="3.01" y2="12"
                        ></line><line x1="3" y1="18" x2="3.01" y2="18"
                        ></line></svg
                    >
                    <span>Lista</span>
                </button>
            </div>
        </div>

        <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
            {#each displayItems as entry}
                <EntryCard {entry} showBadge={false} layout={viewMode === "grid" ? "grid" : "list"} />
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
{:else if loading}
    <!-- Title and lead placeholders over a screen's height, so the footer
         stays below the fold and nothing visible moves when the page arrives
         (UI_BASELINE "ld-reserve-space"). -->
    <div class="location-loading" aria-busy="true">
        <div class="skeleton location-loading__title"></div>
        <div class="skeleton location-loading__lead"></div>
    </div>
{:else}
    <p class="greeting">A keresett oldal nem található.</p>
{/if}

<style>
    .location-loading {
        min-height: 100vh;
        display: flex;
        flex-direction: column;
        gap: 1rem;
        padding-top: 2.5rem;
    }
    .location-loading__title {
        height: 2.5rem;
        width: min(60%, 24rem);
    }
    .location-loading__lead {
        height: 1.2rem;
        width: min(80%, 32rem);
    }

    .more-info {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .capitalize {
        text-transform: capitalize;
    }

    .widgets-box {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 2rem;
        margin-bottom: 2rem;
    }
    /* The overview, coat of arms and weather as cards (UI_BASELINE "szf-pages");
       an attraction has two: overview and weather. */
    .widgets-box--panels {
        gap: 1rem;
    }
    .widgets-box--panels > :global(.widget) {
        /* Its own stacking context, so a child with a negative z-index stays
           above the card's background instead of vanishing behind it. */
        isolation: isolate;
        min-width: 0;
        margin: 0;
        padding: 1.25rem;
        border: 1px solid var(--border-color);
        border-radius: 12px;
        background: var(--card-bg);
    }
    .widgets-box--attraction {
        grid-template-columns: minmax(0, 1.2fr) minmax(0, 1fr);
    }
    .attraction-stats {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(7rem, 1fr));
        gap: 0.6rem;
        margin: 1.1rem 0 0;
    }
    .attraction-stat {
        display: flex;
        flex-direction: column-reverse;
        padding: 0.7rem 0.8rem;
        border: 1px solid var(--border-color);
        border-radius: 12px;
    }
    .attraction-stat dd {
        margin: 0;
        font-size: var(--text-xl);
        font-weight: 700;
        letter-spacing: -0.02em;
        color: var(--szekely-red);
        white-space: nowrap;
    }
    .attraction-stat dt {
        font-size: var(--text-xs);
        color: var(--text-muted);
    }
    /* Activities (✓), prohibitions (✕) and places as chips. */
    .place-chips {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
        margin: 0;
        padding: 0;
        list-style: none;
    }
    .place-chip {
        display: inline-flex;
        align-items: center;
        gap: 0.4rem;
        text-decoration: none;
    }
    span.place-chip {
        padding: 0.4rem 0.9rem;
        border: 1px solid currentColor;
        border-radius: 999px;
        font-size: var(--text-sm);
        font-weight: 600;
    }
    .place-chip--ok {
        color: var(--szekely-green);
    }
    .place-chip--no {
        color: var(--szekely-red);
    }
    .place-chip__mark {
        flex: none;
        width: 1rem;
        height: 1rem;
        fill: none;
        stroke: currentColor;
        stroke-width: 2.5;
        stroke-linecap: round;
        stroke-linejoin: round;
    }
    .place-chip__sub {
        font-weight: 400;
        color: var(--text-muted);
    }
    :global(.news-widget),
    :global(.event-widget) {
        grid-column: span 3;
    }
    @media (max-width: 992px) {
        .widgets-box {
            grid-template-columns: 1fr;
        }
        .widgets-box--attraction {
            grid-template-columns: 1fr;
        }
        :global(.news-widget),
        :global(.event-widget) {
            grid-column: span 1;
        }
    }

    .entry-placeholder {
        display: flex;
        flex-direction: column;
        padding: 1rem;
        gap: 0.5rem;
    }
    .entry-placeholder-cat,
    .entry-placeholder-loc {
        color: var(--text-faint);
    }
    .entry-placeholder-title {
        color: var(--text-faint);
        margin-top: 0.5rem;
    }

    .attraction-photos {
        margin: 1.5rem 0 1rem;
    }
    .attraction-content {
        margin: 1.5rem 0;
    }
    .attraction-contributors-lead {
        margin: 0 0 0.75rem;
        color: var(--text-secondary);
    }
    .attraction-contributor-list {
        display: flex;
        flex-wrap: wrap;
        gap: 0.55rem;
        margin: 0;
        padding: 0;
        list-style: none;
    }
    .attraction-contributor {
        display: inline-flex;
        width: 2.5rem;
        height: 2.5rem;
        border-radius: 999px;
        overflow: hidden;
        border: 1px solid var(--border-color);
        background: var(--card-bg);
    }
    .attraction-contributor img,
    .attraction-contributor-fallback {
        width: 100%;
        height: 100%;
        object-fit: cover;
    }
    .attraction-contributor-fallback {
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 700;
    }

    .component-box {
        padding: 1.5rem;
        background: var(--card-bg);
        border-radius: 12px;
        border: 1px solid var(--border-color);
    }
    .aside-title {
        margin-top: 0;
    }
    .settlements-grid {
        display: flex;
        flex-wrap: wrap;
        gap: 0.8rem;
        margin: 0;
        padding: 0;
    }
    .settlement {
        background: var(--bg-body);
        font-weight: 500;
    }
    .location-title-row {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: 1rem;
    }
    .location-title-row .page-title {
        margin: 0;
        flex: 1 1 auto;
    }
</style>