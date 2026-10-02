<script>
    import { onMount } from "svelte";
    import { apiCall, apiFetch } from "$lib/api.js";
    import { WEBSITE_CREATE_NOTE, WEBSITE_CREATE_TITLE } from "$lib/indexCreateCopy.js";
    import { openLogin } from "$lib/openLogin.js";
    import { auth } from "$lib/stores/auth.js";
    import { canonicalDomain, plainText } from "$lib/websiteDomain.js";

    /** @type {{ onClose: () => void }} */
    let { onClose } = $props();

    let domain = $state("");
    let title = $state("");
    let description = $state("");
    let categoryId = $state(0);
    let catalogLoading = $state(false);
    /** @type {Array<{ id: number, name: string, parent_id?: number | null }>} */
    let catalogCategories = $state([]);
    let error = $state("");
    let pending = $state(false);
    let submitted = $state(false);

    let categoryParents = $derived(
        catalogCategories.filter((row) => row.parent_id == null || row.parent_id === 0),
    );
    let categoryChildren = $derived(
        catalogCategories.filter((row) => row.parent_id != null && row.parent_id > 0),
    );

    let websiteReady = $derived(
        canonicalDomain(domain) !== "" &&
            plainText(title, 120) !== "" &&
            plainText(description, 300) !== "" &&
            Number(categoryId) > 0,
    );

    function fieldError(field) {
        if (field === "domain") return "A webcím nem érvényes.";
        if (field === "title") return "A cím kötelező, legfeljebb 120 karakter.";
        if (field === "description") return "A rövid leírás kötelező, legfeljebb 300 karakter.";
        if (field === "category_id") return "Válassz alkategóriát.";
        return "Ellenőrizd a mezőket.";
    }

    async function loadWebsiteCatalog() {
        if (!$auth.loggedIn) return;
        catalogLoading = true;
        try {
            const catalog = await apiFetch("/api/account/listings/catalog");
            catalogCategories = (Array.isArray(catalog?.categories) ? catalog.categories : [])
                .map((row) => ({
                    id: Number(row.id),
                    name: String(row.name ?? "").trim(),
                    parent_id:
                        row.parent_id == null || row.parent_id === ""
                            ? null
                            : Number(row.parent_id),
                }))
                .filter((row) => row.id > 0 && row.name);
        } catch {
            catalogCategories = [];
        } finally {
            catalogLoading = false;
        }
    }

    async function handleSubmit(event) {
        event.preventDefault();
        if (!$auth.loggedIn) {
            openLogin();
            return;
        }

        pending = true;
        error = "";
        try {
            const res = await apiCall("/api/websites", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    domain,
                    title,
                    description,
                    category_id: Number(categoryId),
                }),
            });

            if (res.status === 201) {
                submitted = true;
                return;
            }

            const data = await res.json().catch(() => ({}));

            if (res.status === 400) {
                error = fieldError(data.field);
            } else if (res.status === 403) {
                error = "Nem adhatsz hozzá weboldalt.";
            } else if (res.status === 409 && data.error === "domain_pending") {
                error = "Ez a domain már jóváhagyásra vár.";
            } else if (res.status === 409 && data.error === "domain_taken") {
                if (data.listing?.name) {
                    error = `Ez a domain már használatban van: ${data.listing.name}`;
                } else if (data.website) {
                    error = `Ez a domain már használatban van: ${data.website.title} (${data.website.domain})`;
                } else {
                    error = "Ez a domain már használatban van.";
                }
            } else {
                error = "A küldés nem sikerült.";
            }
        } catch {
            error = "A küldés nem sikerült.";
        } finally {
            pending = false;
        }
    }

    onMount(() => {
        void loadWebsiteCatalog();
    });
</script>

<div
    class="link-dialog-overlay"
    role="dialog"
    aria-labelledby="add-website-title"
    tabindex="-1"
    onclick={(event) => {
        if (event.target === event.currentTarget) onClose();
    }}
    onkeydown={(event) => {
        if (event.key === "Escape") onClose();
    }}
>
    <div class="link-dialog add-website-form">
        <h3 id="add-website-title">{WEBSITE_CREATE_TITLE}</h3>
        <p class="create-form-note">{WEBSITE_CREATE_NOTE}</p>

        {#if submitted}
            <p class="add-website-form__success">Várakozás a jóváhagyásra.</p>
            <div class="link-dialog-actions">
                <button type="button" class="btn btn-md" onclick={onClose}>Bezárás</button>
            </div>
        {:else}
            <form class="link-dialog-form add-website-form__fields" onsubmit={handleSubmit}>
                {#if catalogLoading}
                    <p>Kategóriák betöltése…</p>
                {/if}
                <label for="website-domain">Webcím <span class="field-required" aria-hidden="true">*</span></label>
                <input
                    id="website-domain"
                    name="domain"
                    type="text"
                    bind:value={domain}
                    autocomplete="url"
                    required
                />

                <label for="website-title">Cím <span class="field-required" aria-hidden="true">*</span></label>
                <input
                    id="website-title"
                    name="title"
                    type="text"
                    bind:value={title}
                    maxlength="120"
                    required
                />

                <label for="website-description">Rövid leírás <span class="field-required" aria-hidden="true">*</span></label>
                <textarea
                    id="website-description"
                    name="description"
                    bind:value={description}
                    maxlength="300"
                    rows="4"
                    required
                ></textarea>

                <label for="website-category">Alkategória <span class="field-required" aria-hidden="true">*</span></label>
                <select id="website-category" name="category_id" bind:value={categoryId} required>
                    <option value={0} disabled>Válassz alkategóriát</option>
                    {#each categoryParents as parent (parent.id)}
                        <optgroup label={parent.name}>
                            {#each categoryChildren.filter((row) => row.parent_id === parent.id) as child (child.id)}
                                <option value={child.id}>{child.name}</option>
                            {/each}
                        </optgroup>
                    {/each}
                </select>

                {#if error}
                    <p class="add-website-form__error">{error}</p>
                {/if}

                <div class="link-dialog-actions">
                    <button type="submit" class="link-dialog-submit" disabled={pending || !websiteReady}>
                        {pending ? "Küldés…" : "Beküldés"}
                    </button>
                    <button type="button" class="btn btn-md" onclick={onClose}>Mégse</button>
                </div>
            </form>
        {/if}
    </div>
</div>

<style>
    .add-website-form__fields textarea {
        padding: 0.6rem 0.8rem;
        border: 1px solid var(--border-color);
        border-radius: 6px;
        font-size: var(--text-base);
        background: var(--bg-body);
        color: var(--text-primary);
        resize: vertical;
        min-height: 5rem;
        font-family: inherit;
    }

    .add-website-form__error {
        margin: 0;
        color: var(--szekely-red);
        font-weight: 600;
    }

    .add-website-form__success {
        margin: 0 0 1rem;
        color: var(--szekely-green);
        font-weight: 600;
    }
</style>
