<script>
    import { onMount } from "svelte";
    import EntryHoursEditor from "$lib/components/EntryHoursEditor.svelte";
    import WebsiteCard from "$lib/components/WebsiteCard.svelte";
    import { canonicalDomain } from "$lib/websiteDomain.js";
    import { emptyWeekHours, hoursConfigured, normalizeHours } from "$lib/entryHours.js";
    import {
        DEFAULT_PHOTO_HEIGHT,
        DEFAULT_PHOTO_WIDTH,
        MAX_ENTRY_PHOTOS,
        normalizePhotos,
    } from "$lib/entryPhotos.js";
    import { apiFetch } from "$lib/api.js";

    const LISTING_LANGUAGES = ["RO", "HU", "DE", "EN"];

    /**
     * @type {{
     *   mode?: "create" | "edit",
     *   entryId?: number,
     *   onClose: () => void,
     *   onSaved?: () => void | Promise<void>,
     *   onError?: (message: string) => void,
     * }}
     */
    let {
        mode = "create",
        entryId = 0,
        onClose,
        onSaved = async () => {},
        onError = () => {},
    } = $props();

    /** @param {unknown} raw */
    function listingLanguages(raw) {
        const values = Array.isArray(raw) ? raw : String(raw ?? "").split(/[,;]+/);
        const selected = new Set(values.map((part) => String(part).trim().toUpperCase()));
        return LISTING_LANGUAGES.filter((code) => selected.has(code));
    }

    function emptyListingForm() {
        return {
            id: 0,
            name: "",
            location_id: 0,
            category_id: 0,
            type_id: 0,
            url: "",
            phone: "",
            address: "",
            notes: "",
            languages: ["HU"],
            hours: emptyWeekHours(),
            delivery_hours: emptyWeekHours(),
            photos: [],
            ratings_enabled: false,
        };
    }

    let listingsError = $state("");
    let listingCatalogLoading = $state(false);
    /** @type {Array<{ id: number, name: string }>} */
    let listingLocations = $state([]);
    /** @type {Array<{ id: number, name: string }>} */
    let listingCategories = $state([]);
    /** @type {Array<{ id: number, name: string }>} */
    let listingTypes = $state([]);
    let listingForm = $state(emptyListingForm());
    let hoursEnabled = $state(false);
    let listingWebsiteId = $state(0);
    /** @type {null | { id: number, title: string, description: string, domain: string, url: string, status: string, claimed: boolean, entry_id: number, published: boolean, membership: string }} */
    let matchedWebsite = $state(null);
    let listingUrlLookup = 0;
    /** @type {ReturnType<typeof setTimeout> | null} */
    let listingUrlTimer = null;
    let listingPhotoLimit = $derived(mode === "create" ? 1 : MAX_ENTRY_PHOTOS);
    let listingPhotoHeading = $derived(listingPhotoLimit === 1 ? "Fotó" : "Fotók");
    let newListingPhotoUrl = $state("");
    let newListingPhotoAlt = $state("");

    async function loadListingCatalog() {
        listingCatalogLoading = true;
        try {
            const [locations, catalog] = await Promise.all([
                apiFetch("/api/locations"),
                apiFetch("/api/account/listings/catalog"),
            ]);
            listingLocations = (Array.isArray(locations) ? locations : [])
                .map((row) => ({
                    id: Number(row.id),
                    name: String(row.name ?? "").trim(),
                }))
                .filter((row) => row.id > 0 && row.name)
                .sort((a, b) => a.name.localeCompare(b.name, "hu"));
            listingCategories = (Array.isArray(catalog?.categories) ? catalog.categories : [])
                .map((row) => ({
                    id: Number(row.id),
                    name: String(row.name ?? "").trim(),
                }))
                .filter((row) => row.id > 0 && row.name);
            listingTypes = (Array.isArray(catalog?.types) ? catalog.types : [])
                .map((row) => ({
                    id: Number(row.id),
                    name: String(row.name ?? "").trim(),
                }))
                .filter((row) => row.id > 0 && row.name);
        } catch {
            listingLocations = [];
            listingCategories = [];
            listingTypes = [];
        } finally {
            listingCatalogLoading = false;
        }
    }

    function clearNewListingPhotoFields() {
        newListingPhotoUrl = "";
        newListingPhotoAlt = "";
    }

    /** @param {string} url */
    function isHttpPhotoUrl(url) {
        const s = String(url ?? "").trim();
        const lower = s.toLowerCase();
        return lower.startsWith("http://") || lower.startsWith("https://");
    }

    /** @param {number} index */
    function removeListingPhoto(index) {
        listingForm = {
            ...listingForm,
            photos: normalizePhotos(listingForm.photos).filter((_, i) => i !== index),
        };
    }

    /** @param {number} index @param {string} alt */
    function updateListingPhotoAlt(index, alt) {
        listingForm = {
            ...listingForm,
            photos: normalizePhotos(listingForm.photos).map((photo, i) =>
                i === index ? { ...photo, alt } : photo,
            ),
        };
    }

    function addListingPhotoFromUrl() {
        const url = String(newListingPhotoUrl ?? "").trim();
        if (!isHttpPhotoUrl(url)) return;
        const current = normalizePhotos(listingForm.photos);
        if (current.length >= listingPhotoLimit) return;
        if (current.some((photo) => photo.url === url)) {
            clearNewListingPhotoFields();
            return;
        }
        listingForm = {
            ...listingForm,
            photos: [
                ...current,
                {
                    url,
                    alt: String(newListingPhotoAlt ?? "").trim(),
                    title: "",
                    description: "",
                    width: DEFAULT_PHOTO_WIDTH,
                    height: DEFAULT_PHOTO_HEIGHT,
                },
            ],
        };
        clearNewListingPhotoFields();
    }

    function toggleListingLanguage(code) {
        const current = listingLanguages(listingForm.languages);
        const next = current.includes(code)
            ? current.filter((item) => item !== code)
            : [...current, code];
        listingForm = {
            ...listingForm,
            languages: LISTING_LANGUAGES.filter((item) => next.includes(item)),
        };
    }

    function clearListingWebsiteMatch() {
        listingWebsiteId = 0;
        matchedWebsite = null;
        listingUrlLookup += 1;
        if (listingUrlTimer) {
            clearTimeout(listingUrlTimer);
            listingUrlTimer = null;
        }
    }

    /** @param {Event} event */
    function scheduleListingWebsiteLookup(event) {
        if (mode !== "create") return;
        const value = event.currentTarget instanceof HTMLInputElement ? event.currentTarget.value : "";
        listingWebsiteId = 0;
        const domain = canonicalDomain(value);
        const token = ++listingUrlLookup;
        if (!domain) {
            matchedWebsite = null;
            return;
        }
        if (listingUrlTimer) clearTimeout(listingUrlTimer);
        listingUrlTimer = setTimeout(() => {
            void lookupListingWebsite(domain, token);
        }, 300);
    }

    /** @param {string} domain @param {number} token */
    async function lookupListingWebsite(domain, token) {
        try {
            const data = await apiFetch(
                `/api/account/websites/lookup?url=${encodeURIComponent(domain)}`,
            );
            if (token !== listingUrlLookup) return;
            matchedWebsite = data?.website ?? null;
        } catch {
            if (token !== listingUrlLookup) return;
            matchedWebsite = null;
        }
    }

    function matchedWebsiteCanClaim() {
        return (
            matchedWebsite?.status === "approved" &&
            Number(matchedWebsite.entry_id) === 0 &&
            !matchedWebsite.membership
        );
    }

    function matchedWebsiteCanJoin() {
        return Boolean(matchedWebsite?.published) && !matchedWebsite?.membership;
    }

    async function claimMatchedWebsite() {
        if (!matchedWebsite) return;
        if (matchedWebsiteCanJoin()) {
            listingsError = "";
            try {
                await apiFetch("/api/account/listings/claim", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ entry_id: matchedWebsite.entry_id }),
                });
                await onSaved();
                onClose();
            } catch {
                listingsError = "Az átvétel nem sikerült";
            }
            return;
        }
        if (!matchedWebsiteCanClaim()) return;
        listingWebsiteId = matchedWebsite.id;
        listingForm = {
            ...listingForm,
            url: matchedWebsite.url,
            name: String(listingForm.name ?? "").trim() ? listingForm.name : matchedWebsite.title,
        };
    }

    /** @param {Record<string, unknown>} form */
    function listingRequestBody(form) {
        const languages = listingLanguages(form.languages);
        return {
            name: String(form.name ?? "").trim(),
            location_id: Number(form.location_id),
            category_id: Number(form.category_id),
            type_id: Number(form.type_id),
            url: String(form.url ?? "").trim(),
            phone: String(form.phone ?? "").trim(),
            address: String(form.address ?? "").trim(),
            notes: String(form.notes ?? "").trim(),
            languages: languages.length ? languages : ["HU"],
            hours: hoursEnabled ? normalizeHours(form.hours) : emptyWeekHours(),
            delivery_hours: normalizeHours(form.delivery_hours),
            photos: normalizePhotos(form.photos).slice(0, listingPhotoLimit),
            ratings_enabled: mode === "edit" ? Boolean(form.ratings_enabled) : false,
            ...(mode === "create" && listingWebsiteId > 0 ? { website_id: listingWebsiteId } : {}),
        };
    }

    /** @param {SubmitEvent} event */
    async function saveListingDialog(event) {
        event.preventDefault();
        listingsError = "";
        if (mode === "create" && matchedWebsiteCanClaim() && listingWebsiteId <= 0) {
            listingsError = "Ez a weboldal már létezik. Az Átveszem gombbal veheted át.";
            return;
        }
        if (mode === "create" && matchedWebsiteCanJoin()) {
            listingsError = "Ehhez a weboldalhoz már van bejegyzés. Az Átveszem gombbal csatlakozhatsz.";
            return;
        }
        if (mode === "create" && matchedWebsite?.status === "pending") {
            listingsError = "Ez a weboldal jóváhagyásra vár.";
            return;
        }
        const body = listingRequestBody(listingForm);
        try {
            if (mode === "create") {
                await apiFetch("/api/account/listings", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify(body),
                });
            } else {
                await apiFetch(
                    `/api/account/listings?id=${encodeURIComponent(String(listingForm.id))}`,
                    {
                        method: "PATCH",
                        headers: { "Content-Type": "application/json" },
                        body: JSON.stringify(body),
                    },
                );
            }
            await onSaved();
            onClose();
        } catch {
            listingsError = "A mentés nem sikerült";
        }
    }

    async function loadEntry(id) {
        try {
            const detail = await apiFetch(
                `/api/account/listings?id=${encodeURIComponent(String(id))}`,
            );
            listingForm = {
                id: Number(detail.id),
                name: String(detail.name ?? ""),
                location_id: Number(detail.location_id),
                category_id: Number(detail.category_id),
                type_id: Number(detail.type_id),
                url: String(detail.url ?? ""),
                phone: String(detail.phone ?? ""),
                address: String(detail.address ?? ""),
                notes: String(detail.notes ?? ""),
                languages: listingLanguages(detail.languages),
                hours: normalizeHours(detail.hours),
                delivery_hours: normalizeHours(detail.delivery_hours),
                photos: normalizePhotos(detail.photos),
                ratings_enabled: Boolean(detail.ratings_enabled),
            };
            hoursEnabled = hoursConfigured(listingForm.hours);
        } catch {
            onError("A mentés nem sikerült");
            onClose();
        }
    }

    function closeDialog() {
        clearListingWebsiteMatch();
        clearNewListingPhotoFields();
        onClose();
    }

    onMount(() => {
        void (async () => {
            await loadListingCatalog();
            if (mode === "edit" && entryId > 0) {
                await loadEntry(entryId);
            }
        })();
    });
</script>

<div
    class="link-dialog-overlay"
    role="dialog"
    aria-labelledby="listing-form-dialog-title"
    tabindex="-1"
    onclick={(e) => e.target === e.currentTarget && closeDialog()}
    onkeydown={(e) => e.key === "Escape" && closeDialog()}
>
    <div class="link-dialog" role="presentation" onclick={(e) => e.stopPropagation()}>
        <h3 id="listing-form-dialog-title">
            {mode === "create" ? "Új bejegyzés" : "Bejegyzés szerkesztése"}
        </h3>
        {#if listingCatalogLoading}
            <p>Katalógus betöltése…</p>
        {:else}
            <form class="link-dialog-form" onsubmit={saveListingDialog}>
                <label for="profile_listing_name">Név</label>
                <input id="profile_listing_name" type="text" bind:value={listingForm.name} required />

                <label for="profile_listing_location">Település</label>
                <select id="profile_listing_location" bind:value={listingForm.location_id} required>
                    <option value={0} disabled>Válassz települést</option>
                    {#each listingLocations as loc (loc.id)}
                        <option value={loc.id}>{loc.name}</option>
                    {/each}
                </select>

                <label for="profile_listing_category">Kategória</label>
                <select id="profile_listing_category" bind:value={listingForm.category_id} required>
                    <option value={0} disabled>Válassz kategóriát</option>
                    {#each listingCategories as cat (cat.id)}
                        <option value={cat.id}>{cat.name}</option>
                    {/each}
                </select>

                <label for="profile_listing_type">Típus</label>
                <select id="profile_listing_type" bind:value={listingForm.type_id} required>
                    <option value={0} disabled>Válassz típust</option>
                    {#each listingTypes as typ (typ.id)}
                        <option value={typ.id}>{typ.name}</option>
                    {/each}
                </select>

                <label for="profile_listing_url">Weboldal URL</label>
                <input
                    id="profile_listing_url"
                    type="url"
                    bind:value={listingForm.url}
                    oninput={scheduleListingWebsiteLookup}
                />
                {#if mode === "create" && matchedWebsite}
                    <div class="listing-url-match">
                        <div class="listing-url-match__card">
                            <WebsiteCard website={matchedWebsite} />
                            {#if (matchedWebsiteCanClaim() && listingWebsiteId !== matchedWebsite.id) || matchedWebsiteCanJoin()}
                                <button
                                    type="button"
                                    class="link-dialog-submit listing-url-match__claim"
                                    onclick={claimMatchedWebsite}
                                >
                                    Átveszem
                                </button>
                            {/if}
                        </div>
                        {#if listingWebsiteId === matchedWebsite.id}
                            <p class="profile-hint">Ezt a weboldalt veszed át. A mentés létrehozza a bejegyzést.</p>
                        {:else if matchedWebsite.status === "pending"}
                            <p class="profile-hint">Ez a weboldal jóváhagyásra vár.</p>
                        {:else if matchedWebsite.entry_id > 0 && !matchedWebsite.published}
                            <p class="profile-hint">Ehhez a weboldalhoz már készül bejegyzés.</p>
                        {/if}
                    </div>
                {/if}
                {#if listingsError}
                    <p class="profile-error">{listingsError}</p>
                {/if}

                <label for="profile_listing_phone">Telefon</label>
                <input id="profile_listing_phone" type="text" bind:value={listingForm.phone} />

                <label for="profile_listing_address">Cím</label>
                <input id="profile_listing_address" type="text" bind:value={listingForm.address} />

                <label for="profile_listing_notes">Megjegyzés</label>
                <textarea id="profile_listing_notes" bind:value={listingForm.notes} rows="3"></textarea>

                <fieldset class="link-dialog-choices">
                    <legend>Nyelvek</legend>
                    {#each LISTING_LANGUAGES as code (code)}
                        <label>
                            <input
                                type="checkbox"
                                checked={listingLanguages(listingForm.languages).includes(code)}
                                onchange={() => toggleListingLanguage(code)}
                            />
                            {code}
                        </label>
                    {/each}
                </fieldset>

                <label class="link-dialog-check">
                    <input type="checkbox" bind:checked={hoursEnabled} />
                    Nyitvatartás / Program
                </label>
                {#if hoursEnabled}
                    <EntryHoursEditor bind:hours={listingForm.hours} />
                {/if}

                {#if mode === "edit"}
                    <h4>Kiszállítási idő</h4>
                    <EntryHoursEditor bind:hours={listingForm.delivery_hours} />
                {/if}

                <div class="profile-listing-photos">
                    <h4>{listingPhotoHeading}</h4>
                    {#if normalizePhotos(listingForm.photos).length === 0}
                        <p class="profile-listing-photos-empty">Még nincs fotó.</p>
                    {:else}
                        <ul class="profile-listing-photos-list">
                            {#each normalizePhotos(listingForm.photos) as photo, i (photo.url + i)}
                                <li class="profile-listing-photos-row">
                                    <a
                                        href={photo.url}
                                        class="profile-listing-photos-url"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                    >{photo.url}</a>
                                    <label class="profile-listing-photos-alt">
                                        Alt
                                        <input
                                            type="text"
                                            value={photo.alt}
                                            oninput={(e) =>
                                                updateListingPhotoAlt(i, e.currentTarget.value)}
                                        />
                                    </label>
                                    <button
                                        type="button"
                                        class="btn btn-xs"
                                        onclick={() => removeListingPhoto(i)}
                                    >Eltávolítás</button>
                                </li>
                            {/each}
                        </ul>
                    {/if}
                    {#if normalizePhotos(listingForm.photos).length < listingPhotoLimit}
                        <div class="profile-listing-photos-add">
                            <label for="profile_listing_photo_url">Kép URL (http vagy https)</label>
                            <input
                                id="profile_listing_photo_url"
                                type="url"
                                bind:value={newListingPhotoUrl}
                                placeholder="https://..."
                            />
                            <label for="profile_listing_photo_alt">Alt (opcionális)</label>
                            <input
                                id="profile_listing_photo_alt"
                                type="text"
                                bind:value={newListingPhotoAlt}
                            />
                            <button
                                type="button"
                                class="btn btn-xs"
                                disabled={!isHttpPhotoUrl(newListingPhotoUrl)}
                                onclick={addListingPhotoFromUrl}
                            >Hozzáadás</button>
                        </div>
                    {/if}
                </div>

                {#if mode === "edit"}
                    <label class="link-dialog-check">
                        <input type="checkbox" bind:checked={listingForm.ratings_enabled} />
                        Értékelések
                    </label>
                {/if}

                <div class="link-dialog-actions">
                    <button type="submit" class="link-dialog-submit">Mentés</button>
                    <button type="button" class="link-dialog-cancel" onclick={closeDialog}>
                        Mégse
                    </button>
                </div>
            </form>
        {/if}
    </div>
</div>

<style>
    .profile-hint {
        margin: 0;
        max-width: 36rem;
        color: var(--text-muted, #666);
    }
    .profile-error {
        margin: 0;
        color: #b00020;
    }
    .listing-url-match {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 0.75rem;
        width: 100%;
    }
    .listing-url-match__card {
        position: relative;
        width: 100%;
    }
    .listing-url-match :global(.website-card) {
        width: 100%;
        margin: 0;
    }
    .listing-url-match :global(.website-card__link) {
        padding-right: 7.5rem;
    }
    .listing-url-match__claim {
        position: absolute;
        top: 0.75rem;
        right: 0.75rem;
        z-index: 1;
    }
    .profile-listing-photos {
        width: 100%;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .profile-listing-photos-empty {
        margin: 0;
        color: var(--text-muted, #666);
        font-size: 0.9rem;
    }
    .profile-listing-photos-list {
        list-style: none;
        margin: 0;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .profile-listing-photos-row {
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
        padding: 0.5rem 0;
        border-bottom: 1px solid var(--border-color, #ddd);
    }
    .profile-listing-photos-url {
        font-size: 0.9rem;
        word-break: break-all;
    }
    .profile-listing-photos-alt {
        display: flex;
        flex-direction: column;
        gap: 0.15rem;
        font-size: 0.9rem;
        font-weight: 600;
    }
    .profile-listing-photos-add {
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
        margin-top: 0.25rem;
    }
</style>
