<script>
    import { onMount } from "svelte";
    import CategoryMultiSelect from "$lib/components/CategoryMultiSelect.svelte";
    import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
    import EntryHoursEditor from "$lib/components/EntryHoursEditor.svelte";
    import WebsiteCard from "$lib/components/WebsiteCard.svelte";
    import { canonicalDomain } from "$lib/websiteDomain.js";
    import { emptyWeekHours, normalizeHours, withDefaultWeekHours } from "$lib/entryHours.js";
    import { offersDelivery } from "$lib/entryPublicExtras.js";
    import {
        DEFAULT_PHOTO_HEIGHT,
        DEFAULT_PHOTO_WIDTH,
        MAX_ENTRY_PHOTOS,
        normalizePhotos,
    } from "$lib/entryPhotos.js";
    import { apiFetch } from "$lib/api.js";
    import AppIcon from "$lib/icons/AppIcon.svelte";
    import { LISTING_CREATE_NOTE, LISTING_CREATE_TITLE } from "$lib/indexCreateCopy.js";

    const LISTING_LANGUAGES = ["RO", "HU", "DE", "EN"];

    /**
     * @type {{
     *   mode?: "create" | "edit",
     *   entryId?: number,
     *   onClose: () => void,
     *   onSaved?: () => void | Promise<void>,
     *   onNotice?: (message: string) => void,
     *   onError?: (message: string) => void,
     * }}
     */
    let {
        mode = "create",
        entryId = 0,
        onClose,
        onSaved = async () => {},
        onNotice = () => {},
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
            location_id: -1,
            category_id: 0,
            category_extra: [],
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
            tags: "",
            social_links: [],
        };
    }

    /** @param {unknown} raw */
    function socialList(raw) {
        return (Array.isArray(raw) ? raw : [])
            .map((row) => ({
                label: String(row?.label ?? "").trim(),
                url: String(row?.url ?? "").trim(),
            }))
            .filter((row) => row.url);
    }

    function addSocialLink() {
        listingForm.social_links = [...listingForm.social_links, { label: "", url: "" }];
    }

    /** @param {number} index */
    function removeSocialLink(index) {
        listingForm.social_links = listingForm.social_links.filter((_, i) => i !== index);
    }

    let listingsError = $state("");
    let formNotice = $state("");
    let confirmOpen = $state(false);
    let confirmMessage = $state("");
    /** @type {((accepted: boolean) => void) | null} */
    let confirmResolve = null;
    let listingCatalogLoading = $state(false);
    /** @type {Array<{ id: number, name: string }>} */
    let listingLocations = $state([]);
    /** @type {Array<{ id: number, name: string, parent_id?: number | null }>} */
    let listingCategories = $state([]);
    /** @type {Array<{ id: number, name: string }>} */
    let listingTypes = $state([]);
    let listingForm = $state(emptyListingForm());
    let hoursEnabled = $state(false);
    let deliveryEnabled = $state(false);
    let listingCategoryParents = $derived(
        listingCategories.filter((row) => row.parent_id == null || row.parent_id === 0),
    );
    let listingCategoryChildren = $derived(
        listingCategories.filter((row) => row.parent_id != null && row.parent_id > 0),
    );
    let listingOffersDelivery = $derived(
        offersDelivery({
            name: listingForm.name,
            category:
                listingCategories.find((row) => row.id === Number(listingForm.category_id))?.name ??
                "",
        }),
    );
    let listingWebsiteId = $state(0);
    /** @type {null | { id: number, title: string, description: string, domain: string, url: string, status: string, claimed: boolean, entry_id: number, published: boolean, membership: string }} */
    let matchedWebsite = $state(null);
    let listingUrlLookup = 0;
    /** @type {ReturnType<typeof setTimeout> | null} */
    let listingUrlTimer = null;
    let listingPhotoLimit = $derived(mode === "create" ? 1 : MAX_ENTRY_PHOTOS);
    let listingFormReady = $derived(
        String(listingForm.name ?? "").trim() !== "" &&
            Number(listingForm.category_id) > 0 &&
            Number(listingForm.type_id) > 0 &&
            listingUrlOk(listingForm.url),
    );
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
                    parent_id:
                        row.parent_id == null || row.parent_id === ""
                            ? null
                            : Number(row.parent_id),
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

    /** @param {string} message */
    function askConfirm(message) {
        if (confirmResolve) {
            const previous = confirmResolve;
            confirmResolve = null;
            previous(false);
        }
        confirmMessage = message;
        confirmOpen = true;
        return new Promise((resolve) => {
            confirmResolve = resolve;
        });
    }

    /** @param {boolean} accepted */
    function closeConfirm(accepted) {
        confirmOpen = false;
        const resolve = confirmResolve;
        confirmResolve = null;
        resolve?.(accepted);
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
    async function removeListingPhoto(index) {
        const yes = await askConfirm("Biztosan eltávolítod ezt a képet?");
        if (!yes) return;
        listingForm = {
            ...listingForm,
            photos: normalizePhotos(listingForm.photos).filter((_, i) => i !== index),
        };
        listingsError = "";
        formNotice = "A kép eltávolítva a listából. A Mentés gomb menti a változást.";
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

    /** @param {number} index @param {string} copyright */
    function updateListingPhotoCopyright(index, copyright) {
        listingForm = {
            ...listingForm,
            photos: normalizePhotos(listingForm.photos).map((photo, i) =>
                i === index ? { ...photo, copyright } : photo,
            ),
        };
    }

    async function addListingPhotoFromUrl() {
        const url = String(newListingPhotoUrl ?? "").trim();
        if (!isHttpPhotoUrl(url)) {
            listingsError = "";
            formNotice = "";
            listingsError = "A kép URL-je http vagy https kell legyen.";
            return;
        }
        const current = normalizePhotos(listingForm.photos);
        if (current.length >= listingPhotoLimit) return;
        if (current.some((photo) => photo.url === url)) {
            listingsError = "";
            formNotice = "";
            listingsError = "Ez a kép már szerepel a listában.";
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
        listingsError = "";
        formNotice = "A kép hozzáadva. A Mentés gomb menti a változást.";
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
            const yes = await askConfirm("Biztosan átveszed ezt a bejegyzést?");
            if (!yes) return;
            listingsError = "";
            formNotice = "";
            try {
                await apiFetch("/api/account/listings/claim", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ entry_id: matchedWebsite.entry_id }),
                });
                await onSaved();
                onNotice("Az átvétel sikerült.");
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
        listingsError = "";
        formNotice = "A weboldal kijelölve. A Beküldés gomb hozza létre a bejegyzést.";
    }

    /** @param {unknown} raw */
    function tagsToField(raw) {
        return (Array.isArray(raw) ? raw : [])
            .map((tag) => String(tag ?? "").trim().replace(/^#/, ""))
            .filter(Boolean)
            .map((tag) => `#${tag}`)
            .join(" ");
    }

    /** @param {unknown} raw */
    function tagsFromField(raw) {
        return String(raw ?? "")
            .split(/[\s,]+/)
            .map((tag) => tag.replace(/^#/, "").trim())
            .filter(Boolean);
    }

    /** @param {Record<string, unknown>} form */
    function listingRequestBody(form) {
        const languages = listingLanguages(form.languages);
        /** @type {number[]} */
        const categoryIds = [];
        const seen = new Set();
        for (const raw of [
            form.category_id,
            ...(Array.isArray(form.category_extra) ? form.category_extra : []),
        ]) {
            const id = Number(raw);
            if (id <= 0 || seen.has(id)) continue;
            seen.add(id);
            categoryIds.push(id);
        }
        return {
            name: String(form.name ?? "").trim(),
            location_id: Number(form.location_id) > 0 ? Number(form.location_id) : 0,
            category_id: categoryIds[0] || 0,
            category_ids: categoryIds,
            type_id: Number(form.type_id) || 0,
            url: String(form.url ?? "").trim(),
            phone: String(form.phone ?? "").trim(),
            address: String(form.address ?? "").trim(),
            notes: String(form.notes ?? "").trim(),
            languages: languages.length ? [...languages] : ["HU"],
            hours: normalizeHours(form.hours),
            hours_enabled: Boolean(hoursEnabled),
            delivery_hours: normalizeHours(form.delivery_hours),
            delivery_enabled: listingOffersDelivery ? Boolean(deliveryEnabled) : false,
            photos: normalizePhotos(form.photos).slice(0, listingPhotoLimit),
            ratings_enabled: mode === "edit" ? Boolean(form.ratings_enabled) : false,
            tags: tagsFromField(form.tags),
            ...(mode === "edit" ? { social_links: socialList(form.social_links) } : {}),
            ...(mode === "create" && listingWebsiteId > 0
                ? { website_id: Number(listingWebsiteId) }
                : {}),
        };
    }

    /** @param {SubmitEvent} event */
    async function saveListingDialog(event) {
        event.preventDefault();
        const yes = await askConfirm(
            mode === "create" ? "Biztosan beküldöd a bejegyzést?" : "Biztosan mented a bejegyzést?",
        );
        if (!yes) return;
        listingsError = "";
        formNotice = "";
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
            onNotice("A mentés sikerült.");
            onClose();
        } catch (err) {
            listingsError = listingSaveErrorMessage(err);
        }
    }

    /** @param {unknown} err */
    function listingSaveErrorMessage(err) {
        const text = String(
            err && typeof err === "object" && "message" in err ? err.message : err ?? "",
        ).trim();
        try {
            const data = JSON.parse(text);
            if (data?.error === "domain_taken") {
                return "Ez a domain már másik bejegyzéshez tartozik.";
            }
            if (data?.field === "category_id") {
                return "A kategória érvénytelen. Válassz újra a listából.";
            }
            if (data?.field === "type_id") {
                return "A típus érvénytelen.";
            }
            if (data?.field === "tags") {
                return "A címke nem lehet megegyező egy kategória nevével.";
            }
            if (data?.error === "invalid") {
                return "A mentés nem sikerült: érvénytelen mező.";
            }
        } catch {
            /* plain text body */
        }
        if (text.includes("entries_unique_entry") || text.includes("duplicate key")) {
            return "Ezzel a névvel már van bejegyzés ezen a településen.";
        }
        if (text.includes("name, category_id, and type_id required")) {
            return "A név, a kategória és a típus kötelező.";
        }
        if (text.includes("invalid url")) {
            return "A weboldal URL érvénytelen.";
        }
        if (text.includes("invalid json")) {
            return "A mentés nem sikerült: hibás adat a űrlapon.";
        }
        if (text.includes("forbidden") || text.includes("unauthorized")) {
            return "Nincs jogosultságod a mentéshez. Jelentkezz be újra.";
        }
        if (text && text !== "A mentés nem sikerült") {
            const short = text.length > 180 ? `${text.slice(0, 180)}...` : text;
            return `A mentés nem sikerült: ${short}`;
        }
        return "A mentés nem sikerült";
    }

    async function loadEntry(id) {
        try {
            const detail = await apiFetch(
                `/api/account/listings?id=${encodeURIComponent(String(id))}`,
            );
            const categoryIDs = (Array.isArray(detail.category_ids) ? detail.category_ids : [])
                .map((id) => Number(id))
                .filter((id) => id > 0);
            const categoryID = Number(detail.category_id) || categoryIDs[0] || 0;
            listingForm = {
                id: Number(detail.id),
                name: String(detail.name ?? ""),
                location_id: Number(detail.location_id) > 0 ? Number(detail.location_id) : 0, // 0 = Nincs település
                category_id: categoryID,
                category_extra: categoryIDs.filter((id) => id !== categoryID),
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
                tags: tagsToField(detail.tags),
                social_links: socialList(detail.social_links),
            };
            hoursEnabled = Boolean(detail.hours_enabled);
            deliveryEnabled = Boolean(detail.delivery_enabled);
        } catch {
            onError("A betöltés nem sikerült");
            onClose();
        }
    }

    /** @param {string} url */
    function listingUrlOk(url) {
        const value = String(url ?? "").trim();
        if (value === "") return true;
        const lower = value.toLowerCase();
        return lower.startsWith("http://") || lower.startsWith("https://");
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
            {mode === "create" ? LISTING_CREATE_TITLE : "Bejegyzés szerkesztése"}
        </h3>
        {#if mode === "create"}
            <p class="create-form-note">{LISTING_CREATE_NOTE}</p>
        {/if}
        {#if listingCatalogLoading}
            <p>Katalógus betöltése…</p>
        {:else}
            <form class="link-dialog-form" onsubmit={saveListingDialog}>
                <label for="profile_listing_name">Név <span class="field-required" aria-hidden="true">*</span></label>
                <input id="profile_listing_name" type="text" bind:value={listingForm.name} required />

                <label for="profile_listing_location">Település</label>
                <select id="profile_listing_location" bind:value={listingForm.location_id}>
                    <option value={-1}>Válassz...</option>
                    <option value={0}>Nincs település</option>
                    {#each listingLocations as loc (loc.id)}
                        <option value={loc.id}>{loc.name}</option>
                    {/each}
                </select>

                <CategoryMultiSelect
                    parents={listingCategoryParents}
                    children={listingCategoryChildren}
                    bind:primary={listingForm.category_id}
                    bind:extra={listingForm.category_extra}
                    primaryInputId="profile_listing_category"
                />

                <label for="profile_listing_type">Típus <span class="field-required" aria-hidden="true">*</span></label>
                <select id="profile_listing_type" bind:value={listingForm.type_id} required>
                    <option value={0}>Válassz...</option>
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
                {#if formNotice}
                    <p class="profile-ok">{formNotice}</p>
                {/if}

                <label for="profile_listing_phone">Telefon</label>
                <input id="profile_listing_phone" type="text" bind:value={listingForm.phone} />

                <label for="profile_listing_address">Cím</label>
                <input id="profile_listing_address" type="text" bind:value={listingForm.address} />

                <label for="profile_listing_notes">Bemutatkozás</label>
                <textarea id="profile_listing_notes" bind:value={listingForm.notes} rows="3"></textarea>

                <label for="profile_listing_tags">Címkék (#cimke1 #cimke2)</label>
                <input id="profile_listing_tags" type="text" bind:value={listingForm.tags} />

                {#if mode === "edit"}
                    <fieldset class="link-dialog-choices">
                        <legend>Közösségi oldalak</legend>
                        {#each listingForm.social_links as link, index (index)}
                            <div class="listing-social">
                                <input type="text" name={`social-label-${index}`} placeholder="Név" aria-label="Közösségi oldal neve" bind:value={link.label} />
                                <input type="url" name={`social-url-${index}`} placeholder="https://" aria-label="Közösségi oldal címe" bind:value={link.url} />
                                <button type="button" class="btn btn-xs" onclick={() => removeSocialLink(index)}>
                                    Eltávolítás
                                </button>
                            </div>
                        {/each}
                        <button type="button" class="btn btn-xs" onclick={addSocialLink}><AppIcon name="plus" size={14} />Hozzáadás</button>
                    </fieldset>
                {/if}

                <fieldset class="link-dialog-choices">
                    <legend>Nyelvek</legend>
                    {#each LISTING_LANGUAGES as code (code)}
                        <label>
                            <input
                                type="checkbox"
                                name={`listing-language-${code}`}
                                checked={listingLanguages(listingForm.languages).includes(code)}
                                onchange={() => toggleListingLanguage(code)}
                            />
                            {code}
                        </label>
                    {/each}
                </fieldset>

                <label class="link-dialog-check">
                    <input
                        type="checkbox"
                        name="hours_enabled"
                        bind:checked={hoursEnabled}
                        onchange={() => {
                            if (hoursEnabled) listingForm.hours = withDefaultWeekHours(listingForm.hours);
                        }}
                    />
                    Nyitvatartás / Program
                </label>
                {#if hoursEnabled}
                    <EntryHoursEditor bind:hours={listingForm.hours} />
                {/if}

                {#if listingOffersDelivery}
                    <label class="link-dialog-check">
                        <input
                            type="checkbox"
                            name="delivery_enabled"
                            bind:checked={deliveryEnabled}
                            onchange={() => {
                                if (deliveryEnabled) {
                                    listingForm.delivery_hours = withDefaultWeekHours(listingForm.delivery_hours);
                                }
                            }}
                        />
                        Kiszállítási idő
                    </label>
                    {#if deliveryEnabled}
                        <EntryHoursEditor bind:hours={listingForm.delivery_hours} />
                    {/if}
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
                                            name={`listing-photo-alt-${i}`}
                                            value={photo.alt}
                                            oninput={(e) =>
                                                updateListingPhotoAlt(i, e.currentTarget.value)}
                                        />
                                    </label>
                                    <label class="profile-listing-photos-alt">
                                        Szerzői jog
                                        <input
                                            type="text"
                                            name={`listing-photo-copyright-${i}`}
                                            value={photo.copyright}
                                            oninput={(e) =>
                                                updateListingPhotoCopyright(i, e.currentTarget.value)}
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
                                class="btn"
                                disabled={!isHttpPhotoUrl(newListingPhotoUrl)}
                                onclick={addListingPhotoFromUrl}
                            >
                                <AppIcon name="upload" size={16} />
                                Hozzáadás
                            </button>
                        </div>
                    {/if}
                </div>

                {#if mode === "edit"}
                    <label class="link-dialog-check">
                        <input type="checkbox" name="ratings_enabled" bind:checked={listingForm.ratings_enabled} />
                        Értékelések
                    </label>
                {/if}

                {#if listingsError}
                    <p class="profile-error" role="alert">{listingsError}</p>
                {/if}

                <div class="link-dialog-actions">
                    <button type="submit" class="link-dialog-submit" disabled={!listingFormReady}>
                        {mode === "create" ? "Beküldés" : "Mentés"}
                    </button>
                    <button type="button" class="link-dialog-cancel" onclick={closeDialog}>
                        Mégse
                    </button>
                </div>
            </form>
        {/if}
    </div>
</div>

<ConfirmDialog
    open={confirmOpen}
    message={confirmMessage}
    onYes={() => closeConfirm(true)}
    onNo={() => closeConfirm(false)}
/>

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
    .profile-ok {
        margin: 0;
        color: #3ddc97;
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
    .profile-listing-photos-add .btn {
        align-self: flex-start;
    }
    .listing-social {
        display: grid;
        grid-template-columns: minmax(6rem, 8rem) 1fr auto;
        gap: 0.4rem;
        align-items: center;
        width: 100%;
    }
</style>
