<script>
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import { get } from "svelte/store";
    import { openLogin } from "$lib/openLogin.js";
    import {
        addFavorite,
        favoriteListWithToggle,
        isFavorite,
        loadFavoriteList,
        removeFavorite,
    } from "$lib/favorites.js";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import EventsWidget from "$lib/components/EventsWidget.svelte";
    import EntryProfile from "$lib/components/EntryProfile.svelte";
    import EntryRelatedLinks from "$lib/components/EntryRelatedLinks.svelte";
    import ListingFormDialog from "$lib/components/ListingFormDialog.svelte";
    import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
    import SuggestionFormDialog from "$lib/components/SuggestionFormDialog.svelte";
    import NoticeDialog from "$lib/components/NoticeDialog.svelte";
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
    /** @type {{ owned: Array<{ id: number, role?: string }>, member: Array<{ id: number, role?: string }>, pending: Array<{ id: number, role?: string }> }} */
    let accountListings = {
        owned: [],
        member: [],
        pending: [],
    };
    let claimError = "";
    let claimBusy = false;
    let requestKind = "";
    let listingsLoaded = false;
    let loadedListingsKey = "";
    let listingsGen = 0;
    const claimRequestMessage =
        "Ezzel kéred, hogy egy admin tegyen a bejegyzés gazdájává.\n\nA kérés a bejelentkezett fiókodat és ezt a bejegyzést küldi el. Megjegyzés nem megy vele.\n\nA bejegyzés Gazdátlan marad, amíg egy admin el nem fogadja. Elfogadás után a bejegyzés Átvéve lesz, és te szerkesztheted. Elutasításkor a kérés törlődik, a bejegyzés Gazdátlan marad. Egyszerre egy átvételi kérés lehet nyitva, és a küldő nem vonhatja vissza.";
    const joinRequestMessage =
        "Ezzel kéred, hogy a bejegyzés gazdája, vagy egy admin, tagként vegyen fel.\n\nA kérés a bejelentkezett fiókodat és ezt a bejegyzést küldi el. Megjegyzés nem megy vele.\n\nA gazda a Bejegyzéseim oldalon fogadja el vagy utasítja el. Egy admin ugyanezt a tagjelölések között teheti meg. Elfogadás után ugyanazokat a mezőket szerkesztheted, mint a gazda. Törölni csak a gazda tudja. Több tagságkérés is nyitva lehet egyszerre.";
    let editingListingId = 0;
    let suggestionOpen = false;
    let suggestionSent = false;

    /**
     * @param {number | string | null | undefined} entryId
     * @param {{ owned: Array<{ id: number, role?: string }>, member: Array<{ id: number, role?: string }>, pending: Array<{ id: number, role?: string }> }} listings
     */
    function listingMembership(entryId, listings) {
        const id = Number(entryId);
        if (!id || !listings) return null;
        if (listings.owned.some((row) => Number(row.id) === id)) {
            return "owner";
        }
        if (listings.member.some((row) => Number(row.id) === id)) {
            return "member";
        }
        const pending = listings.pending.find((row) => Number(row.id) === id);
        if (pending) {
            return pending.role === "owner" ? "pending-owner" : "pending-member";
        }
        return null;
    }

    function openClaimRequest() {
        if (!get(auth).loggedIn) {
            openLogin();
            return;
        }
        claimError = "";
        requestKind = "claim";
    }

    function openJoinRequest() {
        if (!get(auth).loggedIn) {
            openLogin();
            return;
        }
        claimError = "";
        requestKind = "join";
    }

    function closeRequestDialog() {
        requestKind = "";
    }

    async function loadAccountListings() {
        const gen = ++listingsGen;
        await auth.init();
        if (gen !== listingsGen) return;
        if (!get(auth).loggedIn) {
            accountListings = { owned: [], member: [], pending: [] };
            listingsLoaded = true;
            return;
        }
        try {
            const payload = (await apiFetch("/api/account/listings")) || {};
            if (gen !== listingsGen) return;
            accountListings = {
                owned: Array.isArray(payload.owned) ? payload.owned : [],
                member: Array.isArray(payload.member) ? payload.member : [],
                pending: Array.isArray(payload.pending) ? payload.pending : [],
            };
        } catch {
            if (gen !== listingsGen) return;
            accountListings = { owned: [], member: [], pending: [] };
        } finally {
            if (gen === listingsGen) listingsLoaded = true;
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
                body: JSON.stringify({ entry_id: Number(entry.id) }),
            });
            await loadAccountListings();
            await fetchEntry();
        } catch (err) {
            const text = String(err && err.message ? err.message : "");
            if (text.includes("claim_pending")) {
                claimError = "Már van nyitott átvételi kérés ehhez a bejegyzéshez.";
                await fetchEntry();
            } else {
                claimError = "A kérés elküldése nem sikerült.";
            }
        } finally {
            claimBusy = false;
        }
    }

    $: membership = entry ? listingMembership(entry.id, accountListings) : null;
    $: isOwner = membership === "owner";
    $: isMember = membership === "member";
    $: listingsUserKey = browser && $auth.loggedIn ? $auth.email || "in" : "";
    $: if (browser && listingsUserKey !== loadedListingsKey) {
        loadedListingsKey = listingsUserKey;
        listingsLoaded = false;
        loadAccountListings();
    }
    $: claimPending = Boolean(entry && entry.claim_pending);
    $: showClaimButton =
        $auth.loggedIn &&
        listingsLoaded &&
        entry &&
        !entry.claimed &&
        membership == null &&
        !claimPending;
    $: showClaimWaiting =
        $auth.loggedIn &&
        listingsLoaded &&
        entry &&
        !entry.claimed &&
        (claimPending || membership === "pending-owner");
    $: showJoinButton =
        $auth.loggedIn &&
        listingsLoaded &&
        entry &&
        entry.claimed &&
        membership == null;
    $: showJoinWaiting =
        $auth.loggedIn &&
        listingsLoaded &&
        entry &&
        entry.claimed &&
        membership === "pending-member";
    $: showOwnerEdit = $auth.loggedIn && listingsLoaded && isOwner;
    $: suggestionState =
        !$auth.loggedIn || !listingsLoaded || isOwner || isMember
            ? "hidden"
            : entry && entry.suggestion_pending
              ? "waiting"
              : "open";

    async function sendMembershipRequest() {
        closeRequestDialog();
        await submitListingClaim();
    }

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
        <EntryProfile
            {entry}
            loggedIn={$auth.loggedIn}
            isFavorite={isFavorite(favoriteList, "entry", entry.id)}
            onFavorite={() =>
                handleFavoriteToggle(
                    "entry",
                    entry.id,
                    isFavorite(favoriteList, "entry", entry.id),
                )}
            showClaim={showClaimButton}
            showClaimWaiting={showClaimWaiting}
            showJoin={showJoinButton}
            showJoinWaiting={showJoinWaiting}
            {claimBusy}
            onClaim={openClaimRequest}
            onJoin={openJoinRequest}
            {showOwnerEdit}
            {suggestionState}
            onSuggest={() => (suggestionOpen = true)}
            onEdit={() => (editingListingId = entry.id)}
        />

        {#if claimError}
            <p class="entry-claim-error">{claimError}</p>
        {/if}

        <EventsWidget organizerName={entry.name} />
        <EntryRelatedLinks
            {nearby}
            {related}
            history={historyItems}
            currentLocationSlug={entry.location_slug}
        />
    </article>

    <ConfirmDialog
        open={requestKind === "claim" || requestKind === "join"}
        title={requestKind === "join" ? "Tagság kérése" : "Sajátnak jelölöm"}
        message={requestKind === "join" ? joinRequestMessage : claimRequestMessage}
        yesLabel="Kérés elküldése"
        noLabel="Mégse"
        onYes={sendMembershipRequest}
        onNo={closeRequestDialog}
    />

    {#if suggestionOpen && entry}
        <SuggestionFormDialog
            slug={entry.slug}
            onClose={() => (suggestionOpen = false)}
            onSent={async () => {
                suggestionOpen = false;
                suggestionSent = true;
                await fetchEntry();
            }}
        />
    {/if}
    {#if suggestionSent}
        <NoticeDialog onClose={() => (suggestionSent = false)} />
    {/if}

    {#if editingListingId > 0}
        <ListingFormDialog
            mode="edit"
            entryId={editingListingId}
            onClose={() => (editingListingId = 0)}
            onSaved={fetchEntry}
        />
    {/if}
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
