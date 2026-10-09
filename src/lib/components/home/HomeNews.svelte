<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import { bucharestYMD, serverNowMs } from "$lib/bucharestTime.js";

    /**
     * "Friss hírek Erdélyből" on the home page: six news cards, at most two per
     * source (as the news widget always did). Six same-size placeholders hold
     * the grid while the feed loads (ld-reserve-space). "ma" and "tegnap" are
     * Bucharest days from the server's clock (R19).
     */
    const COUNT = 6;
    const PER_SOURCE = 2;
    const MONTHS = ["jan.", "febr.", "márc.", "ápr.", "máj.", "jún.", "júl.", "aug.", "szept.", "okt.", "nov.", "dec."];

    /** @type {Array<{ title: string, link: string, pubDate: number, source: string, bgColor?: string }>} */
    let items = $state([]);
    let loading = $state(true);
    let failed = $state(false);

    onMount(async () => {
        try {
            const all = await apiFetch("/api/news?limit=30");
            const counts = {};
            items = (Array.isArray(all) ? all : [])
                .filter((i) => i?.title && i?.link)
                .filter((i) => (counts[i.source] = (counts[i.source] || 0) + 1) <= PER_SOURCE)
                .slice(0, COUNT);
            failed = items.length === 0;
        } catch {
            failed = true;
        } finally {
            loading = false;
        }
    });

    /** @param {number} ms The feed's pubDate, in milliseconds. */
    function when(ms) {
        if (!ms) return "";
        const day = bucharestYMD(ms);
        const now = serverNowMs();
        if (day === bucharestYMD(now)) return "ma";
        if (day === bucharestYMD(now - 86400000)) return "tegnap";
        const [y, m, d] = day.split("-").map(Number);
        const thisYear = Number(bucharestYMD(now).slice(0, 4));
        return `${y === thisYear ? "" : `${y}. `}${MONTHS[m - 1]} ${d}.`;
    }

    /**
     * The source badge's letter: dark on a light feed colour, white on a dark one.
     * @param {string | undefined} hex
     */
    function ink(hex) {
        const m = /^#?([0-9a-f]{6})$/i.exec(String(hex || ""));
        if (!m) return "var(--text-primary)";
        const n = parseInt(m[1], 16);
        const [r, g, b] = [n >> 16, (n >> 8) & 255, n & 255].map((v) => {
            const c = v / 255;
            return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
        });
        return 0.2126 * r + 0.7152 * g + 0.0722 * b > 0.4 ? "#1a1a1a" : "#ffffff";
    }

    /** @param {string} source */
    function initial(source) {
        return String(source || "?").trim().charAt(0).toUpperCase();
    }
</script>

<section class="home-section home-news" aria-labelledby="home-news-title">
    <div class="home-section__head">
        <h2 id="home-news-title" class="widget-title">Friss hírek Erdélyből</h2>
        <a class="home-section__more" href="/hirek">Összes hír ›</a>
    </div>
    {#if failed}
        <p class="info-box home-news__empty">A hírek jelenleg nem elérhetők.</p>
    {:else}
        <ul class="home-news__grid" aria-busy={loading}>
            {#if loading}
                {#each { length: COUNT } as _, i (i)}
                    <li class="card home-news__card" aria-hidden="true">
                        <span class="home-news__title skeleton">&nbsp;<br />&nbsp;</span>
                        <span class="home-news__source"><span class="home-news__dot skeleton"></span><span class="skeleton">&nbsp;</span></span>
                    </li>
                {/each}
            {:else}
                {#each items as n (n.link)}
                    <li>
                        <a class="card home-news__card" href={n.link} target="_blank" rel="noopener noreferrer">
                            <span class="home-news__title">{n.title}</span>
                            <span class="home-news__source">
                                <span class="home-news__dot" style:background={n.bgColor || "var(--border-color)"} style:color={ink(n.bgColor)}>{initial(n.source)}</span>
                                {n.source}{#if when(n.pubDate)}&nbsp;·&nbsp;{when(n.pubDate)}{/if}
                            </span>
                        </a>
                    </li>
                {/each}
            {/if}
        </ul>
    {/if}
</section>

<style>
    .home-news__grid {
        display: grid;
        grid-template-columns: repeat(3, minmax(0, 1fr));
        gap: 1rem;
        list-style: none;
        margin: 0;
        padding: 0;
    }

    .home-news__card {
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        gap: 0.75rem;
        height: 100%;
        min-height: 6.5rem;
        box-sizing: border-box;
    }

    .home-news__title {
        font-weight: 600;
        line-height: 1.35;
        display: -webkit-box;
        -webkit-line-clamp: 3;
        line-clamp: 3;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    a.home-news__card:hover .home-news__title {
        color: var(--szekely-red);
    }

    .home-news__source {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        color: var(--text-muted);
        font-size: var(--text-sm);
    }

    .home-news__dot {
        display: grid;
        place-items: center;
        flex: none;
        width: 1.25rem;
        height: 1.25rem;
        border-radius: 6px;
        font-size: var(--text-xs);
        font-weight: 700;
    }

    /* The placeholders' text lines keep a width while they hold no words. */
    .home-news__title.skeleton {
        display: block;
    }

    .home-news__source .skeleton:last-child {
        width: 8rem;
    }

    .home-news__empty {
        margin: 0;
    }

    @media (max-width: 900px) {
        .home-news__grid {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (max-width: 560px) {
        .home-news__grid {
            grid-template-columns: 1fr;
        }
    }
</style>
