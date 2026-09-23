<script>
    import { apiCall } from "$lib/api.js";
    import { openLogin } from "$lib/openLogin.js";
    import { auth } from "$lib/stores/auth.js";

    /** @type {{ onClose: () => void }} */
    let { onClose } = $props();

    let domain = $state("");
    let title = $state("");
    let description = $state("");
    let error = $state("");
    let pending = $state(false);
    let submitted = $state(false);

    function fieldError(field) {
        if (field === "domain") return "A webcím nem érvényes.";
        if (field === "title") return "A cím kötelező, legfeljebb 120 karakter.";
        if (field === "description") return "A rövid leírás kötelező, legfeljebb 300 karakter.";
        return "Ellenőrizd a mezőket.";
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
                body: JSON.stringify({ domain, title, description }),
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
                    error = `Ezt a domaint már használja: ${data.listing.name}`;
                } else if (data.website) {
                    error = `Ezt a domaint már használja: ${data.website.title} (${data.website.domain})`;
                } else {
                    error = "Ezt a domaint már átvették.";
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
        <h3 id="add-website-title">Add hozzá a weboldalad</h3>
        <p class="add-website-form__explain">
            Ingyenes. A webcím, a cím és egy rövid leírás kell. Az admin jóváhagyása után a weboldal megjelenik az indexen. Bejegyzés csak akkor lesz belőle, ha valaki később átveszi.
        </p>

        {#if submitted}
            <p class="add-website-form__success">Várakozás az admin jóváhagyására.</p>
            <div class="link-dialog-actions">
                <button type="button" class="btn btn-md" onclick={onClose}>Bezárás</button>
            </div>
        {:else}
            <form class="link-dialog-form add-website-form__fields" onsubmit={handleSubmit}>
                <label for="website-domain">Webcím</label>
                <input
                    id="website-domain"
                    name="domain"
                    type="text"
                    bind:value={domain}
                    autocomplete="url"
                    required
                />

                <label for="website-title">Cím</label>
                <input
                    id="website-title"
                    name="title"
                    type="text"
                    bind:value={title}
                    maxlength="120"
                    required
                />

                <label for="website-description">Rövid leírás</label>
                <textarea
                    id="website-description"
                    name="description"
                    bind:value={description}
                    maxlength="300"
                    rows="4"
                    required
                ></textarea>

                {#if error}
                    <p class="add-website-form__error">{error}</p>
                {/if}

                <div class="link-dialog-actions">
                    <button type="submit" class="link-dialog-submit" disabled={pending}>
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

    .add-website-form__explain {
        margin: 0 0 1rem;
        color: var(--text-muted);
        line-height: 1.45;
    }

    .add-website-form__success {
        margin: 0 0 1rem;
        color: var(--szekely-green);
        font-weight: 600;
    }
</style>
