<script>
    import { onMount } from "svelte";
    import { get } from "svelte/store";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { userSettingsTabIds } from "$lib/accountPrefs.js";
    import { apiFetch } from "$lib/api.js";
    import { openLogin } from "$lib/openLogin.js";
    import {
        clampSlotCount,
        DEFAULT_QUICKLINK_SLOTS,
        MAX_QUICKLINK_SLOTS,
        MIN_QUICKLINK_SLOTS,
        readSlotCount,
        writeSlotCount,
    } from "$lib/quickLinksDisplay.js";
    import { auth } from "$lib/stores/auth";
    import { applyTheme, LABELS, theme } from "$lib/stores/theme";

    const TAB_LABELS = {
        tema: "Téma beállítások",
        linkbeallitasok: "Link beállítások",
        location: "Település beállítások",
    };

    /** @returns {Record<string, unknown>} */
    /** @param {unknown} value */
    let activeTab = $state(userSettingsTabIds[0]);
    let saveError = $state("");
    let linkSaveError = $state("");
    let selectedTheme = $state("system");
    let preferredSettlementId = $state("");
    /** @type {Array<{ id: number, name: string, county: string, type: string }>} */
    let locationChoices = $state([]);
    let locationSaveError = $state("");
    let locationSaving = $state(false);
    let slotCount = $state(DEFAULT_QUICKLINK_SLOTS);
    function initSettingsFromAuth() {
        const state = get(auth);
        slotCount = clampSlotCount(
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
        locationSaveError = "";
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
            await auth.refresh();
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
            preferredSettlementId = state.preferredLocation?.id
                ? String(state.preferredLocation.id)
                : "";
            if (accountDataLoaded) return;
            accountDataLoaded = true;
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
        saveError = "";
        const themeToSave = get(theme);
        const prevTheme = themeToSave;
        try {
            await apiFetch("/api/account/preferences", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ theme: themeToSave }),
            });
            applyTheme(themeToSave);
            await auth.refresh();
        } catch {
            applyTheme(prevTheme);
            saveError = "A mentés nem sikerült";
        }
    }

    async function saveLinkSettings() {
        linkSaveError = "";
        const prevSlots = slotCount;
        const slots = clampSlotCount(slotCount);
        slotCount = slots;
        try {
            await apiFetch("/api/account/preferences", {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ quicklink_slots: slots }),
            });
            writeSlotCount(slots);
            await auth.refresh();
        } catch {
            slotCount = prevSlots;
            linkSaveError = "A mentés nem sikerült";
        }
    }

    function selectTheme(next) {
        applyTheme(next);
    }

    function decreaseSlots() {
        slotCount = clampSlotCount(slotCount - 1);
    }

    function increaseSlots() {
        slotCount = clampSlotCount(slotCount + 1);
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
                <button type="button" class="btn profile-save" onclick={saveTheme}>
                    Mentés
                </button>
            </div>
        {:else if activeTab === "linkbeallitasok"}
            <div class="profile-settings">
                <h3>Gyorslinkek száma a főoldalon</h3>
                <p class="profile-hint">
                    Ennyi hely jelenik meg a kezdőlap gyorslinkjei között, a saját linkjeiddel együtt.
                </p>
                <div
                    class="quicklinks-slot-stepper"
                    role="group"
                    aria-label="Megjelenített gyorslinkek száma"
                >
                    <button
                        type="button"
                        class="btn btn-xs"
                        disabled={slotCount <= MIN_QUICKLINK_SLOTS}
                        aria-label="Kevesebb hely"
                        onclick={decreaseSlots}
                    >−</button>
                    <span class="quicklinks-slot-count" aria-live="polite">{slotCount}</span>
                    <button
                        type="button"
                        class="btn btn-xs"
                        disabled={slotCount >= MAX_QUICKLINK_SLOTS}
                        aria-label="Több hely"
                        onclick={increaseSlots}
                    >+</button>
                </div>
                {#if linkSaveError}
                    <p class="profile-error">{linkSaveError}</p>
                {/if}
                <button type="button" class="btn profile-save" onclick={saveLinkSettings}>
                    Mentés
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
                >
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
</style>
