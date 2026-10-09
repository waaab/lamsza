<script>
    import { onMount } from "svelte";
    import GameCard from "$lib/games/GameCard.svelte";
    import { LISTED_GAMES, featuredGames } from "$lib/games/catalog.js";
    import { jatszoterUrl } from "$lib/networkOrigins.js";

    /**
     * Játszótér's games on the home page: every game Játszótér's /api/games
     * lists as enabled, in its order, as the shared GameCard (the same card
     * as in Játszótér and Szótár; UI_BASELINE "game-look"). The listed games
     * show from the first paint, so nothing moves when the server's list
     * arrives (ld-reserve-space).
     */

    /** @type {Array<{ slug: string, title_hu: string, description_hu: string }>} */
    let games = $state(LISTED_GAMES);

    onMount(async () => {
        try {
            const res = await fetch(jatszoterUrl("/api/games"), { credentials: "omit" });
            if (!res.ok) return;
            games = featuredGames(await res.json(), Infinity) ?? games;
        } catch {
            // Keep the listed games.
        }
    });
</script>

<section class="home-section home-games" aria-labelledby="home-games-title">
    <div class="home-section__head">
        <h2 id="home-games-title" class="widget-title">Játszótér · Mai kihívások</h2>
        <a class="home-section__more" href={jatszoterUrl("/")}>Összes játék ›</a>
    </div>
    <div class="home-games__grid">
        {#each games as g (g.slug)}
            <GameCard slug={g.slug} title_hu={g.title_hu} description_hu={g.description_hu} href={jatszoterUrl(`/jatszok/${g.slug}`)} />
        {/each}
    </div>
</section>

<style>
    .home-games__grid {
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        gap: 1.5rem 1.25rem;
    }

    @media (max-width: 760px) {
        .home-games__grid {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }
</style>
