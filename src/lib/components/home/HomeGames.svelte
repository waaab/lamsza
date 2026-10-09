<script>
    import { onMount } from "svelte";
    import GameIcon from "$lib/games/GameIcon.svelte";
    import { GAMES, LISTED_GAMES } from "$lib/games/catalog.js";
    import { jatszoterUrl } from "$lib/networkOrigins.js";

    /**
     * Játszótér's games on the home page: the first four that Játszótér's
     * /api/games lists as enabled, in its order. The colours and icons are the
     * shared catalog's (synced to Játszótér), so a tile looks the same in both
     * apps. Four tiles show from the first paint (the listed games), so
     * nothing moves when the server's list arrives (ld-reserve-space).
     */
    const COUNT = 4;

    /** @type {Array<{ slug: string, title_hu: string, description_hu: string, color: string }>} */
    let games = $state(LISTED_GAMES.slice(0, COUNT));

    onMount(async () => {
        try {
            const res = await fetch(jatszoterUrl("/api/games"), { credentials: "omit" });
            if (!res.ok) return;
            const data = await res.json();
            const enabled = (Array.isArray(data?.games) ? data.games : [])
                .filter((g) => g.enabled)
                .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0));
            const list = enabled
                .map((g) => {
                    const meta = GAMES.find((m) => m.slug === g.slug);
                    return meta ? { ...meta, title_hu: g.title_hu || meta.title_hu, description_hu: g.description_hu || meta.description_hu } : null;
                })
                .filter((g) => g !== null)
                .slice(0, COUNT);
            if (list.length) games = list;
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
    <ul class="home-games__grid">
        {#each games as g (g.slug)}
            <li>
                <a class="card home-game" href={jatszoterUrl(`/jatszok/${g.slug}`)} style:--game-color={g.color}>
                    <span class="home-game__art"><GameIcon slug={g.slug} size={56} /></span>
                    <span class="home-game__text">
                        <span class="home-game__name">{g.title_hu}</span>
                        <span class="home-game__notes">{g.description_hu}</span>
                    </span>
                </a>
            </li>
        {/each}
    </ul>
</section>

<style>
    .home-games__grid {
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        gap: 1rem;
        list-style: none;
        margin: 0;
        padding: 0;
    }

    .home-game {
        display: flex;
        flex-direction: column;
        height: 100%;
        padding: 0;
    }

    .home-game__art {
        display: grid;
        place-items: center;
        height: 5.25rem;
        background: var(--game-color);
        color: #fff;
    }

    .home-game__text {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        padding: 0.75rem 1rem 1rem;
    }

    .home-game__name {
        font-weight: 700;
    }

    .home-game:hover .home-game__name {
        color: var(--game-color);
    }

    .home-game__notes {
        color: var(--text-muted);
        font-size: var(--text-sm);
        line-height: 1.4;
    }

    @media (max-width: 760px) {
        .home-games__grid {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }
</style>
