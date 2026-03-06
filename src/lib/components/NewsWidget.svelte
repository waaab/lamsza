<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import { formatDate } from "$lib/utils";

    export let settlementSlug = null; // Optional: filter by settlement
    export let limit = 6;

    const NEWS_CACHE_KEY =
        "news_cache" + (settlementSlug ? `_${settlementSlug}` : "");
    const NEWS_TTL = 30 * 60 * 1000;

    let items = [];
    let loading = true;
    let error = false;

    // Helper to extract items from RSS XML
    function parseRSS(xmlText, sourceName, bgColor) {
        const parser = new DOMParser();
        const xmlDoc = parser.parseFromString(xmlText, "text/xml");
        const nodes = Array.from(xmlDoc.querySelectorAll("item")).slice(0, 10);

        return nodes.map((node) => ({
            title: node.querySelector("title")?.textContent || "Cím nélkül",
            link: node.querySelector("link")?.textContent || "#",
            description: node.querySelector("description")?.textContent || "",
            pubDate:
                new Date(
                    node.querySelector("pubDate")?.textContent,
                ).getTime() || 0,
            source: sourceName,
            bgColor: bgColor,
        }));
    }

    onMount(async () => {
        const cached = localStorage.getItem(NEWS_CACHE_KEY);
        if (cached) {
            try {
                const data = JSON.parse(cached);
                if (Date.now() - data.timestamp < NEWS_TTL) {
                    items = data.items;
                    loading = false;
                    return;
                }
            } catch (e) {}
        }

        try {
            const dbFeeds = await apiFetch("/api/admin/news_feeds");
            if (!dbFeeds || dbFeeds.length === 0) {
                error = true;
                loading = false;
                return;
            }

            let allItems = [];
            for (const feed of dbFeeds) {
                try {
                    const xmlText = await apiFetch(
                        `/api/proxy?url=${encodeURIComponent(feed.feed_url)}`,
                        { responseType: "text" },
                    );
                    // Note: apiFetch needs to handle text responses, or we use a separate helper.
                    // For now, let's assume apiFetch handles it or we use a raw fetch for text.
                    const response = await fetch(
                        `${import.meta.env.VITE_API_BASE_URL || "http://localhost:3000"}/api/proxy?url=${encodeURIComponent(feed.feed_url)}`,
                    );
                    const text = await response.text();

                    const parsed = parseRSS(text, feed.title, feed.bg_color);

                    if (settlementSlug) {
                        allItems.push(
                            ...parsed.filter(
                                (i) =>
                                    i.title
                                        .toLowerCase()
                                        .includes(settlementSlug) ||
                                    i.description
                                        .toLowerCase()
                                        .includes(settlementSlug),
                            ),
                        );
                    } else {
                        allItems.push(...parsed);
                    }
                } catch (e) {
                    console.error("RSS feed error:", feed.feed_url, e);
                }
            }

            allItems.sort((a, b) => b.pubDate - a.pubDate);
            items = allItems.slice(0, limit);

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
        }
    });

    // Teaser logic (2 per source) for homepage
    $: teaserItems = (() => {
        if (settlementSlug) return items;
        const counts = {};
        return items.filter((item) => {
            counts[item.source] = (counts[item.source] || 0) + 1;
            return counts[item.source] <= 2;
        });
    })();
</script>

<article id="hirek" class="news news-widget">
    <h3 class="widget-title">
        {settlementSlug ? "Helyi hírek" : "Friss hírek"}
    </h3>

    {#if loading}
        <div class="news-loading-box">
            <div class="skeleton skeleton-text skeleton-full"></div>
            <div class="skeleton skeleton-text skeleton-80"></div>
        </div>
    {:else if error || items.length === 0}
        <span class="info-box"
            ><p>
                {settlementSlug
                    ? "Helyi hírek nem elérhetőek."
                    : "A hírek jelenleg nem elérhetők."}
            </p></span
        >
    {:else}
        <ul class="news-list">
            {#each settlementSlug ? items : teaserItems as item}
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

<style>
    .news-widget {
        display: flex;
        flex-direction: column;
    }
    .news-loading-box {
        padding: 1rem 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .skeleton-full {
        height: 1.2rem;
    }
    .skeleton-80 {
        height: 1.2rem;
        width: 80%;
    }

    .news-list {
        list-style: none;
        padding: 0;
        margin: 0;
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .news-title {
        text-decoration: none;
        color: inherit;
        font-weight: 500;
        font-size: 1.05rem;
    }
    .news-meta {
        font-size: 0.8em;
        color: var(--text-faint);
        margin-top: 0.2rem;
    }
</style>
