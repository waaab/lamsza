<script>
    import { onMount } from "svelte";
    import { apiCall } from "$lib/api.js";

    /** @type {{ attraction?: Record<string, any>, onClose?: () => void, onSent?: () => void | Promise<void> }} */
    let { attraction = {}, onClose = () => {}, onSent = async () => {} } = $props();

    let saving = $state(false);
    let error = $state("");
    let description = $state("");
    let content = $state("");
    let nameRo = $state("");
    let nameDe = $state("");
    let activities = $state("");
    let prohibitions = $state("");
    let note = $state("");

    onMount(() => {
        description = String(attraction.description ?? "");
        content = String(attraction.content ?? "");
        nameRo = String(attraction.name_ro ?? "");
        nameDe = String(attraction.name_de ?? "");
        activities = Array.isArray(attraction.activities) ? attraction.activities.join("\n") : "";
        prohibitions = Array.isArray(attraction.prohibitions) ? attraction.prohibitions.join("\n") : "";
    });

    function activityList(text) {
        return String(text || "")
            .split("\n")
            .map((line) => line.trim())
            .filter(Boolean);
    }

    async function submit(event) {
        event.preventDefault();
        if (saving) return;
        saving = true;
        error = "";
        try {
            const res = await apiCall("/api/attraction-suggestions", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    attraction_id: attraction.id,
                    note,
                    fields: {
                        description: description.trim(),
                        content: content.trim(),
                        name_ro: nameRo.trim(),
                        name_de: nameDe.trim(),
                        activities: activityList(activities),
                        prohibitions: activityList(prohibitions),
                    },
                }),
            });
            if (!res.ok) {
                const text = await res.text();
                error = text.includes("suggestion_pending")
                    ? "Már van nyitott javaslat ehhez a látnivalóhoz."
                    : text.includes("empty_diff")
                      ? "Nincs módosítás a jelenlegi szöveghez képest."
                      : "A javaslatot nem sikerült elküldeni.";
                return;
            }
            await onSent();
        } catch (err) {
            error = "A javaslatot nem sikerült elküldeni.";
        } finally {
            saving = false;
        }
    }
</script>

<div
    class="link-dialog-overlay"
    role="dialog"
    aria-labelledby="attraction-suggest-title"
    tabindex="-1"
    onclick={(e) => e.target === e.currentTarget && onClose()}
    onkeydown={(e) => e.key === "Escape" && onClose()}
>
    <div class="link-dialog" role="presentation" onclick={(e) => e.stopPropagation()}>
        <h3 id="attraction-suggest-title">Javaslat módosításra</h3>
        <p class="create-form-note">
            A módosítás nem jelenik meg azonnal. Egy admin ellenőrzi, és elfogadás után kerül az oldalra.
            Az elfogadott javaslat közreműködői közé tesz.
        </p>
        <form class="link-dialog-form" onsubmit={submit}>
            <label for="attraction_suggest_description">Rövid leírás</label>
            <textarea id="attraction_suggest_description" name="description" rows="3" bind:value={description}></textarea>

            <label for="attraction_suggest_name_ro">Román név</label>
            <input id="attraction_suggest_name_ro" name="name_ro" type="text" bind:value={nameRo} />

            <label for="attraction_suggest_name_de">Német név</label>
            <input id="attraction_suggest_name_de" name="name_de" type="text" bind:value={nameDe} />

            <label for="attraction_suggest_activities">Tevékenységek / aktivitások (soronként egy)</label>
            <textarea id="attraction_suggest_activities" name="activities" rows="4" bind:value={activities}></textarea>

            <label for="attraction_suggest_prohibitions">Mit nem szabad itt csinálni? (soronként egy)</label>
            <textarea id="attraction_suggest_prohibitions" name="prohibitions" rows="4" bind:value={prohibitions}></textarea>

            <label for="attraction_suggest_content">Tartalom (Markdown)</label>
            <textarea id="attraction_suggest_content" name="content" rows="8" bind:value={content}></textarea>

            <label for="attraction_suggest_note">Megjegyzés az adminnak (opcionális)</label>
            <textarea id="attraction_suggest_note" name="note" rows="2" bind:value={note}></textarea>

            {#if error}
                <p class="login-error" role="alert">{error}</p>
            {/if}
            <div class="link-dialog-actions">
                <button type="submit" class="link-dialog-submit" disabled={saving}>
                    {saving ? "Küldés..." : "Javaslat elküldése"}
                </button>
                <button type="button" class="link-dialog-cancel" onclick={onClose}>Mégse</button>
            </div>
        </form>
    </div>
</div>
