<script>
    import { onMount } from "svelte";
    import { formatHuDateLongFromYMD } from "$lib/utils";
    import { szotarUrl } from "$lib/networkOrigins";

    /** @type {{ id: number, text: string, display_date: string }[]} */
    let quotes = [];
    let loading = true;

    // The daily mondás lives in Szótár, its only store (OPEN_ITEMS, Mondások).
    // Szótár decides today in Europe/Bucharest (WAYS_OF_WORKING R19), so no
    // date is sent. A public read: no cookies go along.
    onMount(async () => {
        try {
            const res = await fetch(szotarUrl("/api/proverbs/today"), { credentials: "omit" });
            const data = res.ok ? await res.json() : null;
            quotes = data?.proverb ? [data.proverb] : [];
        } catch {
            quotes = [];
        } finally {
            loading = false;
        }
    });
</script>

{#if !loading && quotes.length > 0}
    <section id="szekely-mondasok">
        <div class="mondas-inner">
            <div class="mondas-label-row">
                <span class="heading-label"
                    >Napi Székely Mondás: Aszongya, hogy…
                    {#if quotes[0]?.display_date}
                        <span class="mondas-date-label"
                            >· {formatHuDateLongFromYMD(quotes[0].display_date)}</span
                        >
                    {/if}</span
                >
            </div>
            {#each quotes as q (q.id)}
                <blockquote class="mondas-quote">{q.text}</blockquote>
            {/each}
        </div>
    </section>
{/if}

<style>
    .mondas-date-label {
        font-weight: 500;
        opacity: 0.85;
    }
</style>