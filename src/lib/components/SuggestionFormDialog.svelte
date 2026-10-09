<script>
    import AppIcon from "$lib/icons/AppIcon.svelte";
    import { onMount } from "svelte";
    import { SvelteSet } from "svelte/reactivity";
    import EntryHoursEditor from "$lib/components/EntryHoursEditor.svelte";
    import { apiFetch } from "$lib/api.js";
    import { normalizeHours } from "$lib/entryHours.js";
    import { offersDelivery } from "$lib/entryPublicExtras.js";
    import { diffSuggestionFields } from "$lib/suggestionDiff.js";

    const LANGUAGES = ["RO", "HU", "DE", "EN"];

    /** @type {{ slug?: string, onClose?: () => void, onSent?: () => void | Promise<void> }} */
    let { slug = "", onClose = () => {}, onSent = async () => {} } = $props();

    let loading = $state(true);
    let saving = $state(false);
    let error = $state("");
    /** @type {Record<string, any>} */
    let before = $state({});
    let entryId = $state(0);
    let name = $state("");
    let services = $state("");
    let notes = $state("");
    let locationId = $state("");
    let address = $state("");
    let url = $state("");
    let phone = $state("");
    /** @type {string[]} */
    let languages = $state(["HU"]);
    /** @type {Record<string, { open: string, close: string, closed: boolean }>} */
    let hours = $state(normalizeHours({}));
    /** @type {Record<string, { open: string, close: string, closed: boolean }>} */
    let deliveryHours = $state(normalizeHours({}));
    let hoursEnabled = $state(false);
    let showDelivery = $state(false);
    /** @type {Array<{ label: string, url: string }>} */
    let socialLinks = $state([]);
    let note = $state("");
    /** @type {Array<{ id: number, name: string }>} */
    let locations = $state([]);

    /** @param {unknown} raw */
    function tagText(raw) {
        return (Array.isArray(raw) ? raw : [])
            .map((tag) => String(tag ?? "").trim())
            .filter(Boolean)
            .sort((a, b) => a.localeCompare(b, "hu"))
            .join(", ");
    }

    /** @param {string} text */
    function tagsFromText(text) {
        return text
            .split(/[,\n]/)
            .map((tag) => tag.trim())
            .filter(Boolean)
            .filter((tag, index, all) => all.indexOf(tag) === index)
            .sort((a, b) => a.localeCompare(b, "hu"));
    }

    /** @param {unknown} raw */
    function languageList(raw) {
        const selected = new Set(
            (Array.isArray(raw) ? raw : []).map((code) => String(code).trim().toUpperCase()),
        );
        const out = LANGUAGES.filter((code) => selected.has(code));
        return out.length ? out : ["HU"];
    }

    /** @param {unknown} raw */
    function socialList(raw) {
        if (!Array.isArray(raw)) return [];
        return raw.map((link) => ({
            label: String(link?.label ?? "").trim(),
            url: String(link?.url ?? "").trim(),
        }));
    }

    function currentFields() {
        /** @type {Record<string, any>} */
        const fields = {
            name: name.trim(),
            tags: tagsFromText(services),
            notes,
            location_id: Number(locationId),
            address: address.trim(),
            url: url.trim(),
            phone: phone.trim(),
            social_links: socialLinks
                .map((link) => ({ label: link.label.trim(), url: link.url.trim() }))
                .filter((link) => link.url),
            languages: languageList(languages),
        };
        if (hoursEnabled) fields.hours = normalizeHours(hours);
        if (showDelivery) fields.delivery_hours = normalizeHours(deliveryHours);
        return fields;
    }

    onMount(() => {
        let cancelled = false;
        (async () => {
            try {
                const [form, locationRows] = await Promise.all([
                    apiFetch(`/api/entry/suggestion-form?slug=${encodeURIComponent(slug)}`),
                    apiFetch("/api/locations"),
                ]);
                if (cancelled || !form) return;
                entryId = Number(form.entry_id);
                name = String(form.name ?? "");
                services = tagText(form.tags);
                notes = String(form.notes ?? "");
                locationId = Number(form.location_id) || "";
                address = String(form.address ?? "");
                url = String(form.url ?? "");
                phone = String(form.phone ?? "");
                languages = languageList(form.languages);
                hours = normalizeHours(form.hours);
                deliveryHours = normalizeHours(form.delivery_hours);
                hoursEnabled = Boolean(form.hours_enabled);
                showDelivery =
                    Boolean(form.delivery_enabled) &&
                    offersDelivery({ name: form.name, category: form.category });
                socialLinks = socialList(form.social_links);
                locations = (Array.isArray(locationRows) ? locationRows : [])
                    .map((row) => ({
                        id: Number(row.id),
                        name: String(row.name ?? "").trim(),
                    }))
                    .filter((row) => row.id > 0 && row.name)
                    .sort((a, b) => a.name.localeCompare(b.name, "hu"));
                if (locationId > 0 && !locations.some((loc) => loc.id === locationId)) {
                    locations = [{ id: locationId, name: `Település ${locationId}` }, ...locations];
                }
                before = currentFields();
            } catch {
                if (!cancelled) error = "A mentés nem sikerült";
            } finally {
                if (!cancelled) loading = false;
            }
        })();
        return () => {
            cancelled = true;
        };
    });

    function toggleLanguage(code) {
        const next = new SvelteSet(languages);
        if (next.has(code)) next.delete(code);
        else next.add(code);
        languages = languageList([...next]);
    }

    function addSocialLink() {
        socialLinks = [...socialLinks, { label: "", url: "" }];
    }

    /** @param {number} index */
    function removeSocialLink(index) {
        socialLinks = socialLinks.filter((_, i) => i !== index);
    }

    /** @param {SubmitEvent} event */
    async function submitSuggestion(event) {
        event.preventDefault();
        error = "";
        const diff = diffSuggestionFields(before, currentFields());
        if (Object.keys(diff).length === 0) {
            error = "Nincs módosított mező.";
            return;
        }
        saving = true;
        try {
            await apiFetch("/api/entry/suggestions", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    entry_id: entryId,
                    fields: diff,
                    note: note.trim(),
                }),
            });
            await onSent();
        } catch (err) {
            const text = String(err && err.message ? err.message : "");
            error = text.includes("suggestion_pending")
                ? "Már van nyitott javaslat ehhez a bejegyzéshez."
                : "A mentés nem sikerült";
        } finally {
            saving = false;
        }
    }
</script>

<div
    class="link-dialog-overlay"
    role="dialog"
    aria-modal="true"
    aria-labelledby="suggestion-form-title"
    tabindex="-1"
    onclick={(e) => e.target === e.currentTarget && onClose()}
    onkeydown={(e) => e.key === "Escape" && onClose()}
>
    <div class="link-dialog suggestion-dialog" role="presentation" onclick={(e) => e.stopPropagation()}>
        <h3 id="suggestion-form-title">Javaslat módosításra</h3>
        <p class="suggestion-lead">
            A mentés csak a megváltozott mezőket küldi el, egy opcionális megjegyzéssel. Egy admin megnézi a javaslatot, majd az egészet elfogadja vagy elutasítja. Elfogadáskor ezek a mezők felülíródnak. A fényképek nem részei a javaslatnak.
        </p>
        {#if loading}
            <p>Betöltés…</p>
        {:else}
            <form class="link-dialog-form" onsubmit={submitSuggestion}>
                <label for="suggestion-name">Név</label>
                <input id="suggestion-name" type="text" bind:value={name} />

                <label for="suggestion-services">Címkék</label>
                <textarea id="suggestion-services" rows="3" bind:value={services}></textarea>

                <label for="suggestion-notes">Bemutatkozás</label>
                <textarea id="suggestion-notes" rows="4" bind:value={notes}></textarea>

                <label for="suggestion-location">Település</label>
                <select id="suggestion-location" bind:value={locationId}>
<option value="">Válassz...</option>
                    {#each locations as loc (loc.id)}
                        <option value={loc.id}>{loc.name}</option>
                    {/each}
                </select>

                <label for="suggestion-address">Cím</label>
                <input id="suggestion-address" type="text" bind:value={address} />

                {#if hoursEnabled}
                    <p class="suggestion-label">Nyitvatartás</p>
                    <EntryHoursEditor bind:hours />
                {/if}

                {#if showDelivery}
                    <p class="suggestion-label">Kiszállítási idő</p>
                    <EntryHoursEditor bind:hours={deliveryHours} />
                {/if}

                <label for="suggestion-url">Weboldal</label>
                <input id="suggestion-url" type="url" bind:value={url} />

                <label for="suggestion-phone">Telefon</label>
                <input id="suggestion-phone" type="text" bind:value={phone} />

                <fieldset class="link-dialog-choices">
                    <legend>Közösségi oldalak</legend>
                    {#each socialLinks as link, index (index)}
                        <div class="suggestion-social">
                            <input type="text" name={`social-label-${index}`} placeholder="Név" aria-label="Közösségi oldal neve" bind:value={link.label} />
                            <input type="url" name={`social-url-${index}`} placeholder="https://" aria-label="Közösségi oldal címe" bind:value={link.url} />
                            <button type="button" class="btn btn-xs" onclick={() => removeSocialLink(index)}>
                                Eltávolítás
                            </button>
                        </div>
                    {/each}
                    <button type="button" class="btn btn-xs" onclick={addSocialLink}><AppIcon name="plus" size={14} />Hozzáadás</button>
                </fieldset>

                <fieldset class="link-dialog-choices">
                    <legend>Nyelvek</legend>
                    {#each LANGUAGES as code (code)}
                        <label>
                            <input
                                type="checkbox"
                                name={`suggestion-language-${code}`}
                                checked={languages.includes(code)}
                                onchange={() => toggleLanguage(code)}
                            />
                            {code}
                        </label>
                    {/each}
                </fieldset>

                <label for="suggestion-note">Megjegyzés</label>
                <textarea id="suggestion-note" rows="2" bind:value={note}></textarea>

                {#if error}
                    <p class="profile-error">{error}</p>
                {/if}
                <div class="link-dialog-actions">
                    <button type="submit" class="link-dialog-submit" disabled={saving}>
                        Javaslat elküldése
                    </button>
                    <button type="button" class="link-dialog-cancel" onclick={() => onClose()}>Mégse</button>
                </div>
            </form>
        {/if}
    </div>
</div>

<style>
    .suggestion-dialog {
        width: min(70vw, calc(100vw - 2rem));
        max-width: min(70vw, calc(100vw - 2rem));
        min-width: 0;
        max-height: calc(100vh - 2rem);
        overflow: auto;
    }
    .suggestion-lead,
    .suggestion-label {
        margin: 0 0 0.75rem;
    }
    .suggestion-social {
        display: grid;
        grid-template-columns: 1fr 1.4fr auto;
        gap: 0.4rem;
        margin-bottom: 0.4rem;
    }
</style>
