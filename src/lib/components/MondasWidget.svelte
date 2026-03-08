<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";

    const MONDAS_CACHE_KEY = "mondas_cache";
    const MONDAS_TTL = 24 * 60 * 60 * 1000;

    let text = "";
    let loading = true;
    let error = false;

    onMount(async () => {
        localStorage.removeItem(MONDAS_CACHE_KEY);
        const cached = localStorage.getItem(MONDAS_CACHE_KEY);
        if (cached) {
            try {
                const { text: cachedText, timestamp } = JSON.parse(cached);
                if (Date.now() - timestamp < MONDAS_TTL) {
                    text = cachedText;
                    loading = false;
                    return;
                }
            } catch (e) {}
        }

        try {
            const data = await apiFetch("/api/admin/mondasok");
            if (data && data.length > 0) {
                const randomIndex = Math.floor(Math.random() * data.length);
                text = data[randomIndex].text;
                localStorage.setItem(
                    MONDAS_CACHE_KEY,
                    JSON.stringify({
                        text,
                        timestamp: Date.now(),
                    }),
                );
            }
        } catch (err) {
            error = true;
        } finally {
            loading = false;
        }
    });
</script>

{#if loading || text !== "" || error}
    <section id="szekely-mondasok">
        <div class="mondas-inner">
            <div class="mondas-label-row">
                <span class="heading-label">Napi Székely Mondás: Aszongya, hogy...</span>
            </div>
            {#if loading}
                <blockquote class="mondas-quote" style="opacity:0.5">adat betöltés...</blockquote>
            {:else if error}
                <p class="mondas-quote">A mondás jelenleg nem elérhető.</p>
            {:else}
                <blockquote class="mondas-quote">{text}</blockquote>
            {/if}
        </div>
    </section>
{/if}

<style>
    /* Styles are already in global.css, but we keep the structure here */
</style>
