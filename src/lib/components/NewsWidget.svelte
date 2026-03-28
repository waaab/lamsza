<script>
    import { onMount, onDestroy } from "svelte";
    import { apiFetch } from "$lib/api";
    import { formatDate } from "$lib/utils";

    export let settlementSlug = null;
    export let limit = 20;
    export let ticker = false;

    const NEWS_CACHE_KEY =
        "news_cache" + (settlementSlug ? `_${settlementSlug}` : "");
    const NEWS_TTL = 30 * 60 * 1000;

    let items = [];
    let loading = true;
    let error = false;

    let tickerIndex = 0;
    let tickerInterval = null;
    let tickerDirection = 1;

    onMount(async () => {
        const cached = localStorage.getItem(NEWS_CACHE_KEY);
        if (cached) {
            try {
                const data = JSON.parse(cached);
                if (Date.now() - data.timestamp < NEWS_TTL) {
                    items = data.items;
                    loading = false;
                    startTicker();
                    return;
                }
            } catch (e) {}
        }

        try {
            let allItems = await apiFetch(`/api/news?limit=${limit}`);

            if (settlementSlug) {
                const slug = settlementSlug.toLowerCase();
                allItems = allItems.filter(
                    (i) => i.title.toLowerCase().includes(slug),
                );
            }

            items = allItems;

            localStorage.setItem(
                NEWS_CACHE_KEY,
                JSON.stringify({
                    items,
                    timestamp: Date.now(),
                }),
            );
        } catch (err) {
            error = true;
        } finally {
            loading = false;
            startTicker();
        }
    });

    function startTicker() {
        if (!ticker || tickerInterval) return;
        tickerInterval = setInterval(() => {
            const list = displayItems;
            if (list.length <= 1) return;
            tickerIndex = (tickerIndex + 1) % list.length;
        }, 5000);
    }

    function stopTicker() {
        if (tickerInterval) {
            clearInterval(tickerInterval);
            tickerInterval = null;
        }
    }

    onDestroy(() => {
        stopTicker();
    });

    $: teaserItems = (() => {
        if (settlementSlug) return items;
        const counts = {};
        return items.filter((item) => {
            counts[item.source] = (counts[item.source] || 0) + 1;
            return counts[item.source] <= 2;
        });
    })();

    $: displayItems = settlementSlug ? items : teaserItems;
</script>

<section id="hirek">
    <article class="news news-widget component-box widget">
        <h3 class="widget-title">
            {settlementSlug ? "Helyi hírek" : "Friss hírek erdélyből"}
        </h3>

        {#if loading}
            <div class="news-loading-placeholder">
                <span class="news-title">adat betöltés...</span>
                <div class="news-meta">adat betöltés...</div>
            </div>
        {:else if error || items.length === 0}
            <span class="info-box"
                ><p>
                    {settlementSlug
                        ? "Helyi hírek nem elérhetőek."
                        : "A hírek jelenleg nem elérhetők."}
                </p></span
            >
        {:else if ticker}
            <div class="news-ticker">
                {#key tickerIndex}
                    {@const item = displayItems[tickerIndex]}
                    {#if item}
                        <div
                            class="news-ticker-item"
                            role="region"
                            aria-label="Hír megállítása rámutatással"
                            on:mouseenter={stopTicker}
                            on:mouseleave={startTicker}
                        >
                            <a
                                href={item.link}
                                target="_blank"
                                rel="nofollow noopener"
                                class="news-title"
                            >
                                📰 {item.title}
                            </a>
                            <div class="news-meta">
                                {item.source} - {formatDate(item.pubDate)}
                            </div>
                        </div>
                    {/if}
                {/key}
                <div class="widget-nav">
                    <div class="arrows-container">
                        <button class="scroll-arrow left" on:click={() => { tickerIndex = (tickerIndex - 1 + displayItems.length) % displayItems.length; }} aria-label="Előző hír">&#8249;</button>
                        <button class="scroll-arrow right" on:click={() => { tickerIndex = (tickerIndex + 1) % displayItems.length; }} aria-label="Következő hír">&#8250;</button>
                    </div>
                    <a href="/hirek" class="nav-btn">Összes hír</a>
                </div>
            </div>
        {:else}
            <ul class="news-list">
                {#each displayItems as item}
                    <li>
                        <a
                            href={item.link}
                            target="_blank"
                            rel="nofollow noopener"
                            class="news-title"
                        >
                            📰 {item.title}
                        </a>
                        <div class="news-meta">
                            {item.source} - {formatDate(item.pubDate)}
                        </div>
                    </li>
                {/each}
            </ul>
        {/if}
    </article>
</section>

<style>
    .news-widget {
        grid-column: span 3;
    }

    .component-box {
        padding: 1.5rem;
        background: var(--card-bg);
        border-radius: 12px;
        border: 1px solid var(--border-color);
    }
    .news-loading-placeholder {
        padding: 0.5rem 0;
    }

    .news-list {
        list-style: none;
        padding: 0;
        margin: 0;
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }
    .news-title {
        text-decoration: none;
        color: inherit;
        font-weight: 500;
        font-size: 1rem;
    }
    .news-meta {
        font-size: 0.8em;
        color: var(--text-faint);
        margin-top: 0.2rem;
    }
    .news-ticker {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .news-ticker-item {
        animation: ticker-slide-in 0.35s ease-out;
    }
    .news-ticker-item a:hover, .news-list a:hover {
        color: var(--szekely-red);
    }
    @keyframes ticker-slide-in {
        from {
            opacity: 0;
            transform: translateY(8px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }
</style>
