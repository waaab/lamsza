<script>
    import { onMount } from "svelte";
    import { get } from "svelte/store";
    import ConfirmDialog from "$lib/components/ConfirmDialog.svelte";
    import AppIcon from "$lib/icons/AppIcon.svelte";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { userSettingsTabIds } from "$lib/accountPrefs.js";
    import { apiFetch } from "$lib/api.js";
    import { openLogin } from "$lib/openLogin.js";
    import {
        isWideQuicklinkLayout,
        readSlotCount,
        slotsForQuicklinkLayout,
        writeSlotCount,
    } from "$lib/quickLinksDisplay.js";
    import { auth } from "$lib/stores/auth";
    import { applyThemeLocal, LABELS, theme } from "$lib/stores/theme";

    const TAB_LABELS = {
        tema: "Téma beállítások",
        linkbeallitasok: "Link beállítások",
        location: "Település beállítások",
    };

    let activeTab = $state(userSettingsTabIds[0]);
    let confirmOpen = $state(false);
    let confirmMessage = $state("");
    /** @type {((accepted: boolean) => void) | null} */
    let confirmResolve = null;
    let saveError = $state("");
    let saveOk = $state("");
    let themeNotice = $state("");
    let themeSaving = $state(false);
    let linkSaveError = $state("");
    let linkSaveOk = $state("");
    let linkNotice = $state("");
    let linkSaving = $state(false);
    let selectedTheme = $state("system");
    let savedTheme = $state("system");
    let preferredSettlementId = $state("");
    /** @type {Array<{ id: number, name: string, county: string, type: string }>} */
    let locationChoices = $state([]);
    let locationSaveError = $state("");
    let locationSaveOk = $state("");
    let locationNotice = $state("");
    let locationSaving = $state(false);
    let quicklinkLayoutWide = $state(false);

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

    /** @param {unknown} value */
    function isThemeId(value) {
        return value === "light" || value === "dark" || value === "system";
    }

    async function refreshAccount() {
        const preview = get(theme);
        await auth.refresh();
        if (isThemeId(preview) && preview !== savedTheme) {
            applyThemeLocal(preview);
        }
    }
    function initSettingsFromAuth() {
        const state = get(auth);
        quicklinkLayoutWide = isWideQuicklinkLayout(
            typeof state.quicklinkSlots === "number"
                ? state.quicklinkSlots
                : readSlotCount(),
        );
    }

    async function loadLocationChoices() {
        try {
            const rows = await apiFetch("/api/locations");
            locationChoices = (Array.isArray(rows) ? rows : [])
                .filter((row) => String(row?.type || "") !== "megye")
                .map((row) => ({
                    id: Number(row.id),
                    name: String(row.name ?? "").trim(),
                    county: String(row.county ?? "").trim(),
                    type: String(row.type ?? "").trim(),
                }))
                .filter((row) => row.id > 0 && row.name)
                .sort((a, b) => a.name.localeCompare(b.name, "hu"));
        } catch {
            locationChoices = [];
            locationSaveError = "A települések betöltése nem sikerült";
        }
    }

    async function savePreferredLocation() {
        const yes = await askConfirm("Biztosan mented a települést?");
        if (!yes) return;
        locationSaveError = "";
        locationSaveOk = "";
        locationNotice = "";
        locationSaving = true;
        const id = Number(preferredSettlementId);
        try {
            await apiFetch("/api/account/preferences", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    preferred_settlement_id: Number.isFinite(id) && id > 0 ? id : 0,
                }),
            });
            await refreshAccount();
            locationSaveOk = "A mentés sikerült.";
        } catch {
            locationSaveError = "A mentés nem sikerült";
        } finally {
            locationSaving = false;
        }
    }

    onMount(() => {
        let accountDataLoaded = false;
        selectedTheme = get(theme);
        const unsubscribeAuth = auth.subscribe((state) => {
            if (!state.loggedIn) {
                accountDataLoaded = false;
                return;
            }
            if (accountDataLoaded) return;
            accountDataLoaded = true;
            preferredSettlementId = state.preferredLocation?.id
                ? String(state.preferredLocation.id)
                : "";
            savedTheme = isThemeId(state.theme) ? state.theme : get(theme);
            initSettingsFromAuth();
            void loadLocationChoices();
        });
        const unsubscribeTheme = theme.subscribe((value) => {
            selectedTheme = value;
        });

        void (async () => {
            await auth.init();
            if (!get(auth).loggedIn) {
                openLogin();
            }
        })();

        return () => {
            unsubscribeAuth();
            unsubscribeTheme();
        };
    });

    async function saveTheme() {
        const yes = await askConfirm("Biztosan mented a témát?");
        if (!yes) return;
        saveError = "";
        saveOk = "";
        themeNotice = "";
        themeSaving = true;
        const themeToSave = get(theme);
        const prevTheme = savedTheme;
        try {
            await apiFetch("/api/account/preferences", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ theme: themeToSave }),
            });
            savedTheme = isThemeId(themeToSave) ? themeToSave : prevTheme;
            applyThemeLocal(savedTheme);
            await refreshAccount();
            saveOk = "A mentés sikerült.";
        } catch {
            saveError = "A mentés nem sikerült";
        } finally {
            themeSaving = false;
        }
    }

    async function saveLinkSettings() {
        const yes = await askConfirm("Biztosan mented a gyorslinkek elrendezését?");
        if (!yes) return;
        linkSaveError = "";
        linkSaveOk = "";
        linkNotice = "";
        linkSaving = true;
        const slots = slotsForQuicklinkLayout(quicklinkLayoutWide);
        try {
            await apiFetch("/api/account/preferences", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ quicklink_slots: slots }),
            });
            writeSlotCount(slots);
            await refreshAccount();
            linkSaveOk = "A mentés sikerült.";
        } catch {
            linkSaveError = "A mentés nem sikerült";
        } finally {
            linkSaving = false;
        }
    }

    /** @param {string} next */
    function selectTheme(next) {
        applyThemeLocal(next);
        saveError = "";
        saveOk = "";
        themeNotice = `${LABELS[next] || next} kiválasztva. A Mentés gomb menti.`;
    }

    /** @param {boolean} wide */
    function chooseQuicklinkLayout(wide) {
        quicklinkLayoutWide = wide;
        linkSaveError = "";
        linkSaveOk = "";
        linkNotice = wide
            ? "Széles elrendezés kiválasztva. A Mentés gomb menti."
            : "Keskeny elrendezés kiválasztva. A Mentés gomb menti.";
    }

    function noteLocationChoice() {
        locationSaveError = "";
        locationSaveOk = "";
        locationNotice = "A település kiválasztva. A Mentés gomb menti.";
    }

</script>

<section class="page-section profile-page">
    <PublicPageHero
        title="Felhasználói beállítások"
        greeting="Fiókod és beállításaid"
        loading={false}
        breadcrumbLabel="Felhasználói beállítások"
        documentTitleSuffix=" – Lámsza"
    />

    {#if !$auth.loggedIn}
        <div class="info-box profile-login-prompt">
            <p>Jelentkezz be a felhasználói beállításokhoz</p>
            <button type="button" class="btn" onclick={openLogin}>Belépés</button>
        </div>
    {:else}
        <nav class="header-tabs profile-tabs" aria-label="Felhasználói beállítások">
            {#each userSettingsTabIds as tabId (tabId)}
                <button
                    type="button"
                    class="btn"
                    class:active={activeTab === tabId}
                    onclick={() => (activeTab = tabId)}
                >
                    {TAB_LABELS[tabId]}
                </button>
            {/each}
        </nav>

        {#if activeTab === "tema"}
            <div class="profile-settings">
                <h3>Téma</h3>
                <div class="profile-theme-buttons" role="group" aria-label="Téma">
                    {#each ["light", "dark", "system"] as themeId (themeId)}
                        <button
                            type="button"
                            class="btn"
                            class:active={selectedTheme === themeId}
                            onclick={() => selectTheme(themeId)}
                        >
                            {LABELS[themeId]}
                        </button>
                    {/each}
                </div>
                {#if saveError}
                    <p class="profile-error">{saveError}</p>
                {/if}
                {#if saveOk}
                    <p class="profile-ok">{saveOk}</p>
                {/if}
                {#if themeNotice}
                    <p class="profile-notice">{themeNotice}</p>
                {/if}
                <button
                    type="button"
                    class="btn profile-save"
                    disabled={themeSaving}
                    onclick={saveTheme}
                >
                    {themeSaving ? "Mentés…" : "Mentés"}
                </button>
            </div>
        {:else if activeTab === "linkbeallitasok"}
            <div class="profile-settings">
                <h3>Gyorslinkek elrendezése a főoldalon</h3>
                <p class="profile-hint">
                    Keskeny elrendezésnél a gyorslinkek, a dátum és az időjárás egy sorban marad.
                    Ha az Új gombbal együtt 7-nél kevesebb gyorslink van, mindig ez a sor látszik, és a kezdőlapon nincs választógomb.
                    Széles elrendezésnél, ha már legalább 7 van, a gyorslinkek a dátum és az időjárás fölé kerülnek, és a képernyőn túlnyúló linkek új sorban folytatódnak. A dátum balra, az időjárás jobbra kerül.
                </p>
                <div class="profile-theme-buttons" role="group" aria-label="Gyorslinkek elrendezése">
                    <button
                        type="button"
                        class="btn"
                        class:active={!quicklinkLayoutWide}
                        aria-pressed={!quicklinkLayoutWide}
                        onclick={() => chooseQuicklinkLayout(false)}
                    ><AppIcon name="collapse" size={16} /> Keskeny</button>
                    <button
                        type="button"
                        class="btn"
                        class:active={quicklinkLayoutWide}
                        aria-pressed={quicklinkLayoutWide}
                        onclick={() => chooseQuicklinkLayout(true)}
                    ><AppIcon name="expand" size={16} /> Széles</button>
                </div>
                {#if linkSaveError}
                    <p class="profile-error">{linkSaveError}</p>
                {/if}
                {#if linkSaveOk}
                    <p class="profile-ok">{linkSaveOk}</p>
                {/if}
                {#if linkNotice}
                    <p class="profile-notice">{linkNotice}</p>
                {/if}
                <button
                    type="button"
                    class="btn profile-save"
                    disabled={linkSaving}
                    onclick={saveLinkSettings}
                >
                    {linkSaving ? "Mentés…" : "Mentés"}
                </button>
            </div>
        {:else if activeTab === "location"}
            <div class="profile-settings">
                <h3>Település beállítások</h3>
                <p class="profile-hint">
                    A saját településed. A kedvenc helyek nem állítják be, és a kereső sem választja ki magától.
                </p>
                <p class="profile-hint">
                    Ha ki van választva, a kezdőlap időjárása és eseménysora ezt a települést használja.
                    Ha nincs, az időjárás az admin alapértelmezett települése, az eseménysor pedig minden település eseményét mutatja.
                    Az index közelségi rendezése is ezt veszi középpontnak, amikor bekapcsolod; üresen az admin alapértelmezett települése a középpont.
                </p>
                <p class="profile-hint">
                    A kereső település nélkül indul, és mindenhol keres, amíg te nem választasz települést.
                    A településed a lista elején jelenik meg, Településem néven, hogy egy koppintással kiválaszthasd.
                    A keresőben vagy az indexen választott szűrő ezt a beállítást nem írja felül.
                </p>
                <label for="preferred_settlement">Település</label>
                <select
                    id="preferred_settlement"
                    class="profile-location-select"
                    bind:value={preferredSettlementId}
                    onchange={noteLocationChoice}
                >
<option value="">Válassz...</option>
                    <option value="">Nincs kiválasztva</option>
                    {#each locationChoices as loc (loc.id)}
                        <option value={String(loc.id)}>
                            {loc.name}{loc.county ? ` (${loc.county})` : ""}{loc.type ? ` – ${loc.type}` : ""}
                        </option>
                    {/each}
                </select>
                {#if locationSaveError}
                    <p class="profile-error">{locationSaveError}</p>
                {/if}
                {#if locationSaveOk}
                    <p class="profile-ok">{locationSaveOk}</p>
                {/if}
                {#if locationNotice}
                    <p class="profile-notice">{locationNotice}</p>
                {/if}
                <button
                    type="button"
                    class="btn profile-save"
                    disabled={locationSaving}
                    onclick={savePreferredLocation}
                >
                    {locationSaving ? "Mentés…" : "Mentés"}
                </button>
            </div>
        {/if}
    {/if}
</section>

<ConfirmDialog
    open={confirmOpen}
    message={confirmMessage}
    onYes={() => closeConfirm(true)}
    onNo={() => closeConfirm(false)}
/>

<style>
    .profile-page {
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }
    .profile-login-prompt {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 0.75rem;
    }
    .profile-tabs {
        flex-wrap: wrap;
    }
    .profile-settings {
        display: flex;
        flex-direction: column;
        gap: 0.75rem;
        align-items: flex-start;
    }
    .profile-settings h3 {
        margin: 0.5rem 0 0;
        font-size: 1rem;
    }
    .profile-hint {
        margin: 0;
        max-width: 36rem;
        color: var(--text-muted, #666);
    }
    .profile-location-select {
        min-width: min(100%, 22rem);
        max-width: 100%;
        padding: 0.45rem 0.6rem;
        border-radius: 8px;
        border: 1px solid var(--border-color);
        background: var(--card-bg, #fff);
        color: var(--text-primary);
        font: inherit;
    }
    .profile-theme-buttons {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
    }
    .profile-error {
        margin: 0;
        color: #b00020;
    }
    .profile-ok {
        margin: 0;
        color: #3ddc97;
    }
    .profile-notice {
        margin: 0;
        color: var(--text-muted, #666);
    }
</style>
