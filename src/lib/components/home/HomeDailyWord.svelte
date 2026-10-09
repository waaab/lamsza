<script>
    import { onMount } from "svelte";
    import { szotarUrl } from "$lib/networkOrigins.js";

    /**
     * Szótár's word of the day for the home page's "Ma" strip. Szótár decides
     * the day (R19): its /api/daily answers with today's word, read in the
     * browser like the daily mondás. While it loads, the same lines hold their
     * place (ld-reserve-space).
     */
    /** @type {{ id: number, headword: string, type: string, meaning: string } | null} */
    let word = $state(null);
    let loading = $state(true);

    onMount(async () => {
        try {
            const res = await fetch(szotarUrl("/api/daily"), { credentials: "omit" });
            if (!res.ok) throw new Error(String(res.status));
            const data = await res.json();
            const w = data?.word;
            if (w?.id && w.headword) {
                const def = Array.isArray(w.definitions) ? w.definitions[0] : null;
                word = {
                    id: w.id,
                    headword: w.headword,
                    type: def?.speech_types?.[0] || "",
                    meaning: def?.definition_hu || "",
                };
            }
        } catch {
            word = null;
        } finally {
            loading = false;
        }
    });
</script>

<div class="widget home-word">
    <div class="widget-header">
        <h3 class="widget-title">Napi székely szó</h3>
    </div>
    {#if loading}
        <div class="widget-content home-word__body" aria-busy="true">
            <span class="home-word__headword skeleton">&nbsp;</span>
            <span class="home-word__meaning skeleton">&nbsp;</span>
        </div>
    {:else if word}
        <a class="widget-content home-word__body" href={szotarUrl(`/szo/${word.id}`)}>
            <span class="home-word__headword">{word.headword}{#if word.type}<span class="home-word__type">{word.type}</span>{/if}</span>
            <span class="home-word__meaning">{word.meaning}</span>
        </a>
    {:else}
        <a class="widget-content home-word__body" href={szotarUrl("/")}>
            <span class="home-word__headword">Szótár</span>
            <span class="home-word__meaning">A mai szó most nem érhető el.</span>
        </a>
    {/if}
</div>

<style>
    .home-word__body {
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
        color: inherit;
        text-decoration: none;
        min-width: 0;
    }

    .home-word__headword {
        display: flex;
        align-items: baseline;
        gap: 0.5rem;
        font-size: var(--text-2xl);
        font-weight: 700;
        line-height: 1.2;
        min-width: 10ch;
    }

    a.home-word__body:hover .home-word__headword {
        color: var(--szekely-red);
    }

    .home-word__type {
        font-size: var(--text-sm);
        font-weight: 400;
        color: var(--text-muted);
    }

    .home-word__meaning {
        color: var(--text-muted);
        line-height: 1.4;
        min-width: 14ch;
    }
</style>
