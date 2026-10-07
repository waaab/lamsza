<script>
    import { onMount } from "svelte";
    import {
        acceptAllConsent,
        closeConsentSettings,
        consent,
        consentPanelOpen,
        consentReady,
        initConsent,
        openConsentSettings,
        rejectOptionalConsent,
        reloadConsent,
        saveConsent,
        CONSENT_STORAGE_KEY,
    } from "$lib/stores/consent.js";

    const POLICY_HREF = "/iranyelvek/sutik";

    let kenyelmi = $state(false);
    let gyorsitotar = $state(false);
    let panelHost = $state(null);

    const showBanner = $derived($consentReady && !$consent.decided && !$consentPanelOpen);
    const showPanel = $derived($consentPanelOpen);

    onMount(() => {
        initConsent();
    });

    // Load the stored values into the switches every time the panel opens.
    $effect(() => {
        if ($consentPanelOpen) {
            kenyelmi = $consent.kenyelmi;
            gyorsitotar = $consent.gyorsitotar;
            panelHost?.focus();
        }
    });

    function onKey(event) {
        if (event.key === "Escape" && $consentPanelOpen) closeConsentSettings();
    }

    /** Another tab changed the choice — follow it instead of asking twice. */
    function onStorage(event) {
        if (event.key === CONSENT_STORAGE_KEY) reloadConsent();
    }

    function acceptAll() {
        acceptAllConsent();
        closeConsentSettings();
    }

    function rejectAll() {
        rejectOptionalConsent();
        closeConsentSettings();
    }

    function saveSelection() {
        saveConsent({ kenyelmi, gyorsitotar });
        closeConsentSettings();
    }
</script>

<svelte:window onkeydown={onKey} onstorage={onStorage} />

{#if showBanner}
    <section class="consent-bar" aria-label="Sütik és helyi tárolás">
        <div class="consent-bar-text">
            <h2>Sütik és helyi tárolás</h2>
            <p>
                A Lámsza annyit tárol a böngésződben, amennyi a működéshez kell:
                belépés, a választott téma és a saját gyorslinkjeid. Ezen felül két
                dolgot kérünk külön — a megnézett találatok megjegyzését és a
                tartalmak gyorsítótárazását. Nem használunk analitikát, hirdetést
                vagy követő sütiket.
            </p>
            <p class="consent-bar-link">
                Részletek: <a href={POLICY_HREF}>Süti tájékoztató</a>
            </p>
        </div>
        <div class="consent-bar-actions">
            <button type="button" class="consent-btn consent-btn-primary" onclick={acceptAll}>
                Mindent elfogadok
            </button>
            <button type="button" class="consent-btn consent-btn-primary" onclick={rejectAll}>
                Csak a szükségeseket
            </button>
            <button type="button" class="consent-btn consent-btn-quiet" onclick={openConsentSettings}>
                Beállítások
            </button>
        </div>
    </section>
{/if}

{#if showPanel}
    <div
        class="consent-overlay"
        role="presentation"
        onclick={closeConsentSettings}
        onkeydown={onKey}
    >
        <div
            class="consent-panel"
            role="dialog"
            aria-modal="true"
            aria-labelledby="consent-panel-title"
            tabindex="-1"
            bind:this={panelHost}
            onclick={(event) => event.stopPropagation()}
            onkeydown={(event) => event.stopPropagation()}
        >
            <h2 id="consent-panel-title">Süti beállítások</h2>
            <p class="consent-panel-lead">
                Az alábbi lista azt mutatja, mit tárol a Lámsza a böngésződben.
                Analitikát, hirdetést és követő sütit nem használunk.
            </p>

            <ul class="consent-list">
                <li class="consent-item">
                    <label class="consent-item-head">
                        <input type="checkbox" checked disabled />
                        <span class="consent-item-title">Működéshez szükséges</span>
                        <span class="consent-badge">Mindig be van kapcsolva</span>
                    </label>
                    <p class="consent-item-desc">
                        Belépés és munkamenet (<code>lamsza_session</code> süti,
                        <code>lamsza_auth_session</code>), a választott téma
                        (<code>theme</code>) és a saját gyorslinkjeid
                        (<code>user_quick_links</code>,
                        <code>quick_links_display_count</code>), valamint ez a
                        süti-döntés (<code>lamsza_cookie_consent</code>). Ezek nélkül
                        nem tudjuk nyújtani, amit kérsz az oldaltól, ezért nem
                        kapcsolhatók ki.
                    </p>
                </li>

                <li class="consent-item">
                    <label class="consent-item-head">
                        <input type="checkbox" bind:checked={kenyelmi} />
                        <span class="consent-item-title">Kényelmi funkciók</span>
                    </label>
                    <p class="consent-item-desc">
                        Megjegyzi, mely találatokat nézted meg utoljára, és
                        visszamutatja őket (<code>lamsza_entry_history</code>). Ha
                        kikapcsolod, a lista törlődik.
                    </p>
                </li>

                <li class="consent-item">
                    <label class="consent-item-head">
                        <input type="checkbox" bind:checked={gyorsitotar} />
                        <span class="consent-item-title">Tartalom gyorsítótár</span>
                    </label>
                    <p class="consent-item-desc">
                        Rövid időre elmenti az időjárást, a híreket és a kiemelt
                        linkeket (<code>weather_cache_*</code>,
                        <code>news_cache*</code>, <code>hirek_cache</code>,
                        <code>promoted_links_cache</code>), hogy az oldal gyorsabban
                        töltsön. Ha kikapcsolod, minden látogatáskor újra letöltjük
                        ezeket.
                    </p>
                </li>
            </ul>

            <p class="consent-panel-link">
                Részletes leírás: <a href={POLICY_HREF}>Süti tájékoztató</a>
            </p>

            <div class="consent-panel-actions">
                <button type="button" class="consent-btn consent-btn-primary" onclick={saveSelection}>
                    Választás mentése
                </button>
                <button type="button" class="consent-btn consent-btn-primary" onclick={acceptAll}>
                    Mindent elfogadok
                </button>
                <button type="button" class="consent-btn consent-btn-primary" onclick={rejectAll}>
                    Csak a szükségeseket
                </button>
                {#if $consent.decided}
                    <button type="button" class="consent-btn consent-btn-quiet" onclick={closeConsentSettings}>
                        Mégse
                    </button>
                {/if}
            </div>
        </div>
    </div>
{/if}

<style>
    .consent-bar {
        position: fixed;
        left: 0;
        right: 0;
        bottom: 0;
        z-index: 1200;
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: space-between;
        gap: 1rem;
        padding: 1rem 1.25rem;
        background: var(--card-bg, #fff);
        color: var(--text-primary, #111);
        border-top: 1px solid var(--border-color, #d1d1d1);
        box-shadow: 0 -4px 16px var(--shadow-lg, rgba(0, 0, 0, 0.14));
    }

    .consent-bar-text {
        flex: 1 1 22rem;
        min-width: 0;
    }

    .consent-bar h2 {
        margin: 0 0 0.35rem;
        font-size: 1rem;
    }

    .consent-bar p {
        margin: 0;
        font-size: 0.9rem;
        line-height: 1.45;
        color: var(--text-secondary, #595959);
    }

    .consent-bar-link {
        margin-top: 0.35rem !important;
    }

    .consent-bar-actions,
    .consent-panel-actions {
        display: flex;
        flex-wrap: wrap;
        gap: 0.5rem;
    }

    .consent-btn {
        padding: 0.55rem 1rem;
        border-radius: 6px;
        border: 1px solid var(--border-color, #d1d1d1);
        background: var(--card-bg, #fff);
        color: var(--text-primary, #111);
        font-size: 0.9rem;
        font-weight: 600;
        cursor: pointer;
    }

    /* Accept and reject look the same on purpose — no nudging toward "accept". */
    .consent-btn-primary {
        border-color: var(--szekely-green, #2f4f4f);
        background: var(--szekely-green, #2f4f4f);
        color: var(--white, #fff);
    }

    .consent-btn-quiet {
        font-weight: 500;
        color: var(--text-secondary, #595959);
    }

    .consent-btn:hover {
        filter: brightness(1.08);
    }

    .consent-btn:focus-visible {
        outline: 2px solid var(--szekely-blue, #0059b3);
        outline-offset: 2px;
    }

    .consent-overlay {
        position: fixed;
        inset: 0;
        z-index: 1300;
        display: flex;
        align-items: center;
        justify-content: center;
        padding: 1rem;
        background: rgba(0, 0, 0, 0.62);
    }

    .consent-panel {
        width: min(34rem, 100%);
        max-height: 85vh;
        overflow-y: auto;
        padding: 1.25rem;
        border-radius: 10px;
        background: var(--card-bg, #fff);
        color: var(--text-primary, #111);
        box-shadow: 0 12px 32px var(--shadow-xl, rgba(0, 0, 0, 0.15));
    }

    .consent-panel h2 {
        margin: 0 0 0.5rem;
        font-size: 1.15rem;
    }

    .consent-panel-lead,
    .consent-panel-link {
        margin: 0 0 1rem;
        font-size: 0.9rem;
        color: var(--text-secondary, #595959);
    }

    .consent-list {
        list-style: none;
        margin: 0 0 1rem;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 0.9rem;
    }

    .consent-item {
        padding: 0.75rem;
        border: 1px solid var(--border-color, #d1d1d1);
        border-radius: 8px;
    }

    .consent-item-head {
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 0.5rem;
        cursor: pointer;
    }

    .consent-item-title {
        font-weight: 600;
    }

    .consent-badge {
        padding: 0.1rem 0.45rem;
        border-radius: 999px;
        background: var(--badge-bg, #e0e0e0);
        color: var(--badge-text, #111);
        font-size: 0.72rem;
    }

    .consent-item-desc {
        margin: 0.45rem 0 0;
        font-size: 0.85rem;
        line-height: 1.5;
        color: var(--text-secondary, #595959);
    }

    .consent-item-desc code {
        font-size: 0.8rem;
        word-break: break-all;
    }

    @media (max-width: 600px) {
        .consent-bar-actions .consent-btn,
        .consent-panel-actions .consent-btn {
            flex: 1 1 100%;
        }
    }
</style>
