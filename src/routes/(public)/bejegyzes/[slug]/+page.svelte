<script>
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import { get } from "svelte/store";
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
    import EventsWidget from "$lib/components/EventsWidget.svelte";
    import EntryHistoryStrip from "$lib/components/EntryHistoryStrip.svelte";
    import EntryProfile from "$lib/components/EntryProfile.svelte";
    import EntryRelatedLinks from "$lib/components/EntryRelatedLinks.svelte";
    import { apiFetch } from "$lib/api";
    import {
        historyForDisplay,
        readHistory,
        recordAccountHistory,
    } from "$lib/entryHistory.js";
    import { auth } from "$lib/stores/auth";
    import { normalizePhotos } from "$lib/entryPhotos.js";

    let entry = null;
    let loading = true;
    let error = null;
    let nearby = [];
    let related = [];
    let historyItems = [];
    let fetchGen = 0;
    /** @type {{ type: string, id: number }[]} */
    let favoriteList = [];
    /** @type {{ owned: Array<{ id: number }>, member: Array<{ id: number }>, pending: Array<{ id: number }> }} */
    let accountListings = {
        owned: [],
        member: [],
        pending: [],
    };
    let claimError = "";
    let claimBusy = false;

    /** @param {number | null | undefined} entryId */
    function listingMembership(entryId) {
        const id = Number(entryId);
        if (!id) return null;
        if (accountListings.owned.some((row) => Number(row.id) === id)) {
            return "owner";
        }
        if (accountListings.member.some((row) => Number(row.id) === id)) {
            return "member";
        }
        if (accountListings.pending.some((row) => Number(row.id) === id)) {
            return "pending";
        }
        return null;
    }

    async function loadAccountListings() {
        await auth.init();
        if (!get(auth).loggedIn) {
            accountListings = { owned: [], member: [], pending: [] };
            return;
        }
        try {
            const payload = (await apiFetch("/api/account/listings")) || {};
            accountListings = {
                owned: Array.isArray(payload.owned) ? payload.owned : [],
                member: Array.isArray(payload.member) ? payload.member : [],
                pending: Array.isArray(payload.pending) ? payload.pending : [],
            };
        } catch {
            accountListings = { owned: [], member: [], pending: [] };
        }
    }

    async function submitListingClaim() {
        if (!entry?.id) return;
        if (!get(auth).loggedIn) {
            openLogin();
            return;
        }
        claimError = "";
        claimBusy = true;
        try {
            await apiFetch("/api/account/listings/claim", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ entry_id: entry.id }),
            });
            await loadAccountListings();
            await fetchEntry();
        } catch {
            claimError = "A mentés nem sikerült";
        } finally {
            claimBusy = false;
        }
    }

    $: membership = entry ? listingMembership(entry.id) : null;
    $: showClaimButton =
        $auth.loggedIn && entry && !entry.claimed && membership == null;
    $: showJoinButton =
        $auth.loggedIn &&
        entry &&
        entry.claimed &&
        membership == null;

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
        loadAccountListings();
    });

    $: slug = $page.params.slug;

    $: if (browser && slug) {
        fetchEntry();
    }

    async function fetchEntry() {
        const requested = slug;
        const gen = ++fetchGen;
        loading = true;
        error = null;
        try {
            const data = await apiFetch(
                `/api/entry?slug=${encodeURIComponent(requested)}`,
            );
            if (gen !== fetchGen) return;

            if (!data || !data.name) {
                error = "A bejegyzés nem található.";
                entry = null;
                nearby = [];
                related = [];
                historyItems = [];
            } else {
                entry = data;
                loading = false;
                await auth.init();
                await recordAccountHistory(
                    {
                        slug: data.slug,
                        name: data.name,
                        category: data.category,
                        location: data.location,
                        photo: normalizePhotos(data.photos)[0]?.url || "",
                    },
                    get(auth),
                );
                historyItems = historyForDisplay(readHistory(), data.slug);

                nearby = [];
                related = [];
                try {
                    const rel = await apiFetch(
                        `/api/entry/related?slug=${encodeURIComponent(data.slug || requested)}`,
                    );
                    if (gen !== fetchGen) return;
                    nearby = Array.isArray(rel?.nearby) ? rel.nearby : [];
                    related = Array.isArray(rel?.related) ? rel.related : [];
                } catch {
                    if (gen !== fetchGen) return;
                    nearby = [];
                    related = [];
                }
            }
        } catch (err) {
            if (gen !== fetchGen) return;
            console.error(err);
            error = "Hiba történt a szerver kapcsolat közben.";
            entry = null;
            nearby = [];
            related = [];
            historyItems = [];
        } finally {
            if (gen === fetchGen) loading = false;
        }
    }
</script>

<svelte:head>
    <title>{entry ? entry.name : "Bejegyzés"} - Index</title>
</svelte:head>

{#if loading}
    <div aria-busy="true" aria-label="Bejegyzés betöltése">
        <div class="entry-page-skel-crumbs">
            {#each { length: 4 }}
                <span class="skeleton skeleton-text entry-page-skel-crumb"></span>
            {/each}
        </div>
        <article class="profile-detail">
            <EntryProfile placeholder />
            <section class="entry-page-skel-events" aria-hidden="true">
                <div class="skeleton skeleton-text entry-page-skel-events-title"></div>
                <div class="skeleton entry-page-skel-events-card"></div>
            </section>
        </article>
    </div>
{:else if error}
    <span class="info-box error">
        <p>{error}</p>
    </span>
    <a href="/" class="btn back-to-home">Vissza a főoldalra</a>
{:else if entry}
    <Breadcrumbs
        label={entry.name}
        countySlug={entry.county_slug}
        countyName={entry.location_county || entry.county}
        settlementSlug={entry.location_slug}
        settlementName={entry.location}
        settlementType={entry.location_type}
    />

    <article class="profile-detail">
        <EntryProfile {entry} />

        <div class="page-actions">
            <FavoriteButton
                type="entry"
                id={entry.id}
                active={isFavorite(favoriteList, "entry", entry.id)}
                ontoggle={() =>
                    handleFavoriteToggle(
                        "entry",
                        entry.id,
                        isFavorite(favoriteList, "entry", entry.id),
                    )}
            />
            {#if showClaimButton}
                <button
                    type="button"
                    class="btn"
                    disabled={claimBusy}
                    onclick={() => submitListingClaim()}
                >
                    Sajátnak jelölöm
                </button>
            {/if}
            {#if showJoinButton}
                <button
                    type="button"
                    class="btn"
                    disabled={claimBusy}
                    onclick={() => submitListingClaim()}
                >
                    Tagság kérése
                </button>
            {/if}
        </div>
        {#if claimError}
            <p class="entry-claim-error">{claimError}</p>
        {/if}

        <EventsWidget organizerName={entry.name} />
        <EntryRelatedLinks {nearby} {related} currentLocationSlug={entry.location_slug} />
        <EntryHistoryStrip items={historyItems} />
    </article>
{/if}

<style>
    .back-to-home {
        margin-top: 1rem;
        display: inline-block;
    }
    .profile-detail {
        display: flex;
        flex-direction: column;
        gap: 2rem;
    }
    .entry-page-skel-crumbs {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.45rem;
        margin-bottom: 1.25rem;
    }
    .entry-page-skel-crumb {
        width: 5.5rem;
        height: 0.85rem;
        margin: 0;
    }
    .entry-page-skel-events-title {
        width: 10rem;
        height: 1.1rem;
        margin: 0 0 0.85rem;
    }
    .entry-page-skel-events-card {
        width: 100%;
        height: 7.5rem;
        border-radius: 12px;
    }
    .entry-claim-error {
        margin: 0;
        color: #b00020;
    }
</style>
