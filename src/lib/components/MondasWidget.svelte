<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import { formatHuDateLongFromYMD, localCalendarISODate } from "$lib/utils";

    /** @type {{ id: number, text: string, display_date: string }[]} */
    let quotes = [];
    let loading = true;

    onMount(async () => {
        try {
            const day = localCalendarISODate();
            const data = await apiFetch(
                `/api/mondasok?date=${encodeURIComponent(day)}`,
            );
            quotes = Array.isArray(data) ? data : [];
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
