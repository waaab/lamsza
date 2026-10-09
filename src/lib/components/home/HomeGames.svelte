<script>
    import { onMount } from "svelte";
    import GameCard from "$lib/games/GameCard.svelte";
    import { LISTED_GAMES, todaysChallenges } from "$lib/games/catalog.js";
    import { jatszoterUrl } from "$lib/networkOrigins.js";

    /**
     * "Játszótér · Mai kihívások" on the home page: the games that have a
     * daily challenge today (Játszótér's /api/games `daily_today`), in its
     * order, as the shared GameCard (UI_BASELINE "game-look", "home-lamsza").
     * A card opens the game's page on today's challenge; the player starts it
     * with Kezdés. Until the answer arrives the listed games hold the space,
     * dimmed and inert (ld-reserve-space). If Játszótér cannot be reached they
     * stay, as plain links; if no game has a challenge today the section goes.
     */

    /** @type {Array<{ slug: string, title_hu: string, description_hu: string }>} */
    let games = $state(LISTED_GAMES);
    let loading = $state(true);

    onMount(async () => {
        try {
            const res = await fetch(jatszoterUrl("/api/games"), { credentials: "omit" });
            if (res.ok) games = todaysChallenges(await res.json()) ?? games;
        } catch {
            // Keep the listed games.
        } finally {
            loading = false;
        }
    });
</script>

{#if games.length}
    <section class="home-section home-games" aria-labelledby="home-games-title">
        <div class="home-section__head">
            <h2 id="home-games-title" class="widget-title">Játszótér · Mai kihívások</h2>
            <a class="home-section__more" href={jatszoterUrl("/")}>Összes játék ›</a>
        </div>
        <div class="home-games__grid" class:home-games__grid--loading={loading} inert={loading} aria-busy={loading}>
            {#each games as g (g.slug)}
                <GameCard slug={g.slug} title_hu={g.title_hu} description_hu={g.description_hu} href={jatszoterUrl(`/jatszok/${g.slug}`)} />
            {/each}
        </div>
    </section>
{/if}

<style>
    .home-games__grid {
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        gap: 1.5rem 1.25rem;
        transition: opacity 0.2s ease;
    }

    .home-games__grid--loading {
        opacity: 0.55;
    }

    @media (max-width: 760px) {
        .home-games__grid {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (prefers-reduced-motion: reduce) {
        .home-games__grid {
            transition: none;
        }
    }
</style>
