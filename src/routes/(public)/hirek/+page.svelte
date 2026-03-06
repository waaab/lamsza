<script>
    import { onMount, onDestroy } from "svelte";

    let allNewsItems = [];
    let visibleCount = 9;
    let sources = [];
    let loading = true;
    let error = false;
    let viewMode = "grid";
    let selectedSource = null;
    let selectedWord = null;
    let sourcesOpen = false;
    let cacheTimestamp = null;

    function scrollToTop() {
        if (typeof window !== "undefined") {
            window.scrollTo({ top: 0, behavior: "smooth" });
        }
    }

    const STOP_WORDS = new Set([
        "a",
        "az",
        "és",
        "is",
        "nem",
        "már",
        "csak",
        "de",
        "egy",
        "meg",
        "ez",
        "egyik",
        "nincs",
        "hogy",
        "van",
        "volt",
        "lesz",
        "el",
        "ki",
        "be",
        "fel",
        "le",
        "át",
        "vissza",
        "még",
        "sem",
        "se",
        "ha",
        "vagy",
        "mint",
        "pedig",
        "mert",
        "így",
        "úgy",
        "ami",
        "ahol",
        "aki",
        "amely",
        "azt",
        "ezt",
        "erre",
        "olyan",
        "ott",
        "itt",
        "akkor",
        "most",
        "majd",
        "után",
        "előtt",
        "között",
        "miatt",
        "ellen",
        "mellett",
        "alatt",
        "felett",
        "szerint",
        "által",
        "illetve",
        "valamint",
        "tehát",
        "ahogy",
        "melyet",
        "amelyet",
        "amelynek",
        "amelyben",
        "melyek",
        "azok",
        "ezek",
        "ők",
        "mi",
        "ti",
        "én",
        "te",
        "ő",
        "sok",
        "több",
        "kevés",
        "minden",
        "valami",
        "valaki",
        "senki",
        "semmi",
        "igen",
        "jó",
        "nagy",
        "kis",
        "új",
        "régi",
        "szép",
        "rossz",
        "gyors",
        "lasú",
        "lett",
        "lenne",
        "len",
        "két",
        "három",
        "négy",
        "öt",
        "hat",
        "kér",
        "kéri",
        "kért",
        "fog",
        "fogja",
        "fogjuk",
        "tett",
        "tesz",
        "teszi",
        "aztán",
        "rá",
        "felől",
        "irányban",
        "részén",
        "területén",
        "oldalon",
    ]);

    // Stage 1: filter by selected source
    $: sourceFilteredItems = selectedSource
        ? allNewsItems.filter((i) => i.source === selectedSource)
        : allNewsItems;

    // Stage 2: compute top-20 words (min 2 occurrences) from source-filtered titles
    $: topWords = (() => {
        const freq = {};
        for (const item of sourceFilteredItems) {
            item.title
                .toLowerCase()
                .replace(
                    /[^a-z\u00e1\u00e9\u00ed\u00f3\u00f6\u0151\u00fa\u00fc\u0171\s-]/gi,
                    " ",
                )
                .split(/\s+/)
                .map((w) => w.replace(/^-+|-+$/g, ""))
                .filter((w) => w.length > 0 && !STOP_WORDS.has(w))
                .forEach((w) => (freq[w] = (freq[w] || 0) + 1));
        }
        return Object.entries(freq)
            .filter(([, c]) => c >= 2)
            .sort((a, b) => b[1] - a[1])
            .slice(0, 10);
    })();

    // Stage 3: filter by selected word (case-insensitive substring match)
    $: filteredItems = selectedWord
        ? sourceFilteredItems.filter((i) =>
              i.title.toLowerCase().includes(selectedWord.toLowerCase()),
          )
        : sourceFilteredItems;

    $: totalCount = filteredItems.length;
    $: displayItems = filteredItems.slice(0, visibleCount);

    // Reset pagination on any filter change
    $: {
        selectedSource;
        selectedWord;
        visibleCount = 12;
    }

    const DEFAULT_IMAGE =
        "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='400' height='225' viewBox='0 0 400 225'%3E%3Crect width='400' height='225' fill='%232f4f4f'/%3E%3Ctext x='200' y='113' text-anchor='middle' dominant-baseline='middle' fill='%23a0c0b0' font-size='15' font-family='system-ui'%3E%F0%9F%93%B0 Sz%C3%A9kely Gugel%3C%2Ftext%3E%3C%2Fsvg%3E";

    function extractImage(node) {
        // 1. <enclosure> with image type
        const enc = node.querySelector("enclosure");
        if (
            enc?.getAttribute("url") &&
            enc.getAttribute("type")?.startsWith("image")
        ) {
            return enc.getAttribute("url");
        }
        // 2. <media:content> or <media:thumbnail> (namespace-agnostic query)
        const mc =
            node.querySelector("content") || node.querySelector("thumbnail");
        if (mc?.getAttribute("url")) return mc.getAttribute("url");

        // 3. <img> inside <description>
        const desc = node.querySelector("description")?.textContent || "";
        const m = desc.match(/<img[^>]+src=["']([^"']+)["']/i);
        if (m) return m[1];

        return null;
    }

    function formatDate(ts) {
        return new Date(ts).toLocaleString("hu-HU", {
            year: "numeric",
            month: "long",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
        });
    }

    function handleImgError(e) {
        e.target.src = DEFAULT_IMAGE;
    }

    function showMore() {
        visibleCount += 12;
    }

    onMount(async () => {
        const isMobile = window.innerWidth < 768;
        sourcesOpen = !isMobile;

        const NEWS_CACHE_KEY = "news_cache";
        const NEWS_TTL = 30 * 60 * 1000;

        try {
            const apiBase = import.meta.env.VITE_API_BASE_URL;
            if (!apiBase)
                console.warn(
                    "VITE_API_BASE_URL is not set. Falling back to http://localhost:3000",
                );
            const baseUrl = apiBase || "http://localhost:3000";

            const feedsRes = await fetch(`${baseUrl}/api/admin/news_feeds`);
            let dbFeeds = [];
            if (feedsRes.ok) {
                dbFeeds = await feedsRes.json();
            } else {
                throw new Error(`News feeds API hiba: ${feedsRes.status}`);
            }
            sources = dbFeeds || [];

            if (!dbFeeds || dbFeeds.length === 0) {
                error = true;
                loading = false;
                return;
            }

            // Try cache first
            const cached = localStorage.getItem(NEWS_CACHE_KEY);
            if (cached) {
                const { items, timestamp } = JSON.parse(cached);
                if (Date.now() - timestamp < NEWS_TTL && items?.length > 0) {
                    allNewsItems = items;
                    cacheTimestamp = timestamp;
                    loading = false;
                    return;
                }
            }

            // Fetch all feeds
            const allItems = [];
            const tsObj = JSON.parse(
                localStorage.getItem("news_feed_timestamps") || "{}",
            );

            for (const feedObj of dbFeeds) {
                const feedUrl = feedObj.feed_url;
                try {
                    const proxiedUrl =
                        `${baseUrl}/api/proxy?url=` +
                        encodeURIComponent(feedUrl);
                    const response = await fetch(proxiedUrl);
                    const text = await response.text();
                    const parser = new DOMParser();
                    const xml = parser.parseFromString(text, "text/xml");
                    const nodes = Array.from(
                        xml.querySelectorAll("item"),
                    ).slice(0, 20);

                    tsObj[feedUrl] = Date.now();

                    nodes.forEach((node) => {
                        try {
                            allItems.push({
                                title:
                                    node.querySelector("title")?.textContent ||
                                    "Cím nélkül",
                                link:
                                    node.querySelector("link")?.textContent ||
                                    "#",
                                pubDate:
                                    new Date(
                                        node.querySelector(
                                            "pubDate",
                                        )?.textContent,
                                    ).getTime() || 0,
                                source: feedObj.title || "Ismeretlen",
                                bgColor: feedObj.bg_color || "#222",
                                image: extractImage(node),
                            });
                        } catch (e) {
                            console.error("RSS elem hiba:", e);
                        }
                    });
                } catch (e) {
                    console.error("RSS forrás hiba:", feedUrl, e);
                }
            }

            localStorage.setItem("news_feed_timestamps", JSON.stringify(tsObj));

            if (allItems.length > 0) {
                allItems.sort((a, b) => b.pubDate - a.pubDate);
                allNewsItems = allItems;
                const now = Date.now();
                cacheTimestamp = now;
                localStorage.setItem(
                    NEWS_CACHE_KEY,
                    JSON.stringify({ items: allItems, timestamp: now }),
                );
            } else {
                error = true;
            }
        } catch (err) {
            console.error("Teljes RSS lekérési hiba:", err);
            // Stale cache fallback — show old news rather than an error
            const staleNews = localStorage.getItem(NEWS_CACHE_KEY);
            if (staleNews) {
                try {
                    const { items, timestamp } = JSON.parse(staleNews);
                    if (items && items.length > 0) {
                        allNewsItems = items;
                        cacheTimestamp = timestamp ?? null;
                        // Rebuild sources list from cached items
                        const seen = new Set();
                        sources = items
                            .filter((i) => {
                                if (seen.has(i.source)) return false;
                                seen.add(i.source);
                                return true;
                            })
                            .map((i) => ({
                                title: i.source,
                                bg_color: i.bgColor,
                            }));
                        loading = false;
                        return;
                    }
                } catch (e) {}
            }
            error = true;
        } finally {
            loading = false;
        }
    });
    let canScrollLeft = false;
    let canScrollRight = false;
    let chipsContainer;

    function checkScroll(node) {
        if (!node) return;
        canScrollLeft = node.scrollLeft > 5;
        canScrollRight =
            node.scrollLeft < node.scrollWidth - node.clientWidth - 5;
    }

    // Drag to scroll logic for chips
    function dragScroll(node) {
        let isDown = false;
        let startX;
        let scrollLeft;

        const onMouseDown = (e) => {
            isDown = true;
            node.classList.add("active");
            startX = e.pageX - node.offsetLeft;
            scrollLeft = node.scrollLeft;
        };

        const onMouseLeave = () => {
            isDown = false;
            node.classList.remove("active");
        };

        const onMouseUp = () => {
            isDown = false;
            node.classList.remove("active");
        };

        const onMouseMove = (e) => {
            if (!isDown) return;
            e.preventDefault();
            const x = e.pageX - node.offsetLeft;
            const walk = (x - startX) * 2; // Scroll-fast factor
            node.scrollLeft = scrollLeft - walk;
            checkScroll(node);
        };

        const onScroll = () => {
            checkScroll(node);
        };

        const onResize = () => {
            checkScroll(node);
        };

        node.addEventListener("mousedown", onMouseDown);
        node.addEventListener("mouseleave", onMouseLeave);
        node.addEventListener("mouseup", onMouseUp);
        node.addEventListener("mousemove", onMouseMove);
        node.addEventListener("scroll", onScroll);
        window.addEventListener("resize", onResize);

        // Initial check after a short delay for layout
        setTimeout(() => checkScroll(node), 100);

        return {
            destroy() {
                node.removeEventListener("mousedown", onMouseDown);
                node.removeEventListener("mouseleave", onMouseLeave);
                node.removeEventListener("mouseup", onMouseUp);
                node.removeEventListener("mousemove", onMouseMove);
                node.removeEventListener("scroll", onScroll);
                window.removeEventListener("resize", onResize);
            },
        };
    }

    function scrollChips(amount) {
        if (chipsContainer) {
            chipsContainer.scrollBy({ left: amount, behavior: "smooth" });
        }
    }
</script>

<svelte:head>
    <title>Friss hírek erdélyi forrásból - Székely Gugel</title>
    <meta
        name="description"
        content="Friss hírek erdélyi forrásból - helyi hírcsatornák legfrissebb hírei időrendben."
    />
</svelte:head>

<h1 class="page-title">Friss hírek erdélyi forrásból</h1>
<p class="greeting">Helyi hírcsatornák legfrissebb hírei időrendben.</p>

{#if loading}
    <div class="header-tabs chips">
        <span class="header-tabs-label">Leggyakoribb témák:</span>
        <div class="chips-scroll-wrapper">
            <div class="chips-list">
                {#each Array(6) as _}
                    <div class="btn-skeleton"></div>
                {/each}
            </div>
        </div>
    </div>
{:else if allNewsItems.length > 0}
    <div class="header-tabs chips">
        <span class="header-tabs-label">Leggyakoribb témák:</span>
        <div
            class="chips-scroll-wrapper"
            class:can-left={canScrollLeft}
            class:can-right={canScrollRight}
        >
            <button
                class="scroll-arrow left"
                aria-label="Görgetés balra"
                on:click={() => scrollChips(-200)}>‹</button
            >
            <div class="chips-list" use:dragScroll bind:this={chipsContainer}>
                {#each topWords as [word, count]}
                    <button
                        class="btn btn-md {selectedWord === word
                            ? 'active'
                            : ''}"
                        on:click={() => {
                            selectedWord = selectedWord === word ? null : word;
                            scrollToTop();
                        }}
                    >
                        {word} <span class="news-word-count">{count}</span>
                    </button>
                {/each}
            </div>
            <button
                class="scroll-arrow right"
                aria-label="Görgetés jobbra"
                on:click={() => scrollChips(200)}>›</button
            >
        </div>
    </div>
{/if}

{#if selectedSource || selectedWord}
    <div class="filter-actions">
        <div class="info-box">
            <p>
                {#if selectedSource && selectedWord}
                    💡 Leszűrve:<span class="active"
                        >{selectedSource} és {selectedWord}</span
                    >
                {:else if selectedSource}
                    💡 Leszűrve:<span class="active">{selectedSource}</span>
                {:else if selectedWord}
                    💡 Leszűrve:<span class="active">{selectedWord}</span>
                {/if}

                <button
                    class="clear-filters btn btn-sm"
                    aria-label="Szűrők törlése"
                    title="Szűrők törlése"
                    on:click={() => {
                        selectedSource = null;
                        selectedWord = null;
                        scrollToTop();
                    }}
                >
                    Szűrő törlése
                </button>
            </p>
            <p>
                <span>({displayItems.length}/{totalCount})</span>
            </p>
        </div>

        <div class="view-mode-toggle">
            <button
                class="btn btn-sm {viewMode === 'grid' ? 'active' : ''}"
                on:click={() => (viewMode = "grid")}
                title="Rács nézet"
            >
                <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    stroke="currentColor"
                    stroke-width="2"
                    fill="none"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><rect x="3" y="3" width="7" height="7"></rect><rect
                        x="14"
                        y="3"
                        width="7"
                        height="7"
                    ></rect><rect x="14" y="14" width="7" height="7"
                    ></rect><rect x="3" y="14" width="7" height="7"></rect></svg
                >
                <span>Rács</span>
            </button>
            <button
                class="btn btn-sm {viewMode === 'flex' ? 'active' : ''}"
                on:click={() => (viewMode = "flex")}
                title="Lista nézet"
            >
                <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    stroke="currentColor"
                    stroke-width="2"
                    fill="none"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><line x1="8" y1="6" x2="21" y2="6"></line><line
                        x1="8"
                        y1="12"
                        x2="21"
                        y2="12"
                    ></line><line x1="8" y1="18" x2="21" y2="18"></line><line
                        x1="3"
                        y1="6"
                        x2="3.01"
                        y2="6"
                    ></line><line x1="3" y1="12" x2="3.01" y2="12"></line><line
                        x1="3"
                        y1="18"
                        x2="3.01"
                        y2="18"
                    ></line></svg
                >
                <span>Lista</span>
            </button>
        </div>
    </div>
{:else}
    <div class="filter-actions">
        <div class="info-box">
            <p>
                💡 Leszűrve: <span class="active">Összes</span>
            </p>
            <p>
                <span>({displayItems.length}/{totalCount})</span>
            </p>
        </div>

        <div class="view-mode-toggle">
            <button
                class="btn btn-sm {viewMode === 'grid' ? 'active' : ''}"
                on:click={() => (viewMode = "grid")}
                title="Rács nézet"
            >
                <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    stroke="currentColor"
                    stroke-width="2"
                    fill="none"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><rect x="3" y="3" width="7" height="7"></rect><rect
                        x="14"
                        y="3"
                        width="7"
                        height="7"
                    ></rect><rect x="14" y="14" width="7" height="7"
                    ></rect><rect x="3" y="14" width="7" height="7"></rect></svg
                >
                <span>Rács</span>
            </button>
            <button
                class="btn btn-sm {viewMode === 'flex' ? 'active' : ''}"
                on:click={() => (viewMode = "flex")}
                title="Lista nézet"
            >
                <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    stroke="currentColor"
                    stroke-width="2"
                    fill="none"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><line x1="8" y1="6" x2="21" y2="6"></line><line
                        x1="8"
                        y1="12"
                        x2="21"
                        y2="12"
                    ></line><line x1="8" y1="18" x2="21" y2="18"></line><line
                        x1="3"
                        y1="6"
                        x2="3.01"
                        y2="6"
                    ></line><line x1="3" y1="12" x2="3.01" y2="12"></line><line
                        x1="3"
                        y1="18"
                        x2="3.01"
                        y2="18"
                    ></line></svg
                >
                <span>Lista</span>
            </button>
        </div>
    </div>
{/if}

<div class="news-page-layout">
    <!-- Main news grid -->
    <section class="news-list">
        {#if loading}
            <div class="list grid">
                {#each Array(9) as _}
                    <div class="card news--skeleton">
                        <div class="news-img-wrap skeleton"></div>
                        <div class="news-body">
                            <div class="skeleton skeleton-text"></div>
                            <div
                                class="skeleton skeleton-text news-skeleton-short"
                            ></div>
                        </div>
                    </div>
                {/each}
            </div>
        {:else if error}
            <div class="note error">
                Hírek jelenleg nem elérhetők. Próbáld újra hamarosan.
            </div>
        {:else}
            <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
                {#each displayItems as item}
                    <article class="card news">
                        <a
                            href={item.link}
                            target="_blank"
                            rel="nofollow noopener"
                            class="news-link"
                        >
                            <div class="news-img-wrap">
                                <img
                                    src={item.image || DEFAULT_IMAGE}
                                    alt={item.title}
                                    class="news-img"
                                    loading="lazy"
                                    on:error={handleImgError}
                                />
                                <span
                                    class="news-source badge"
                                    style:background={item.bgColor}
                                    >{item.source}</span
                                >
                            </div>
                            <div class="news-body">
                                <h3 class="news-title">{item.title}</h3>
                                <time class="news-date"
                                    >{formatDate(item.pubDate)}</time
                                >
                            </div>
                        </a>
                    </article>
                {/each}
            </div>

            {#if visibleCount < totalCount}
                <div class="load-more">
                    <button class="nav-btn" on:click={showMore}>
                        Mutasd a következő híreket ↓
                    </button>
                </div>
            {/if}
        {/if}
    </section>

    <!-- Sidebar -->
    <aside class="news-sidebar">
        <div class="news-sidebar-box">
            <div class="news-sidebar-header">
                <h4 class="news-sidebar-heading">Erdélyi hírforrások</h4>
                <button
                    class="sidebar-toggle-btn"
                    class:open={sourcesOpen}
                    on:click={() => (sourcesOpen = !sourcesOpen)}
                    aria-label={sourcesOpen ? "Bezárás" : "Megnyitás"}
                >
                    ▾
                </button>
            </div>
            {#if sourcesOpen}
                {#if loading}
                    <ul class="news-sidebar-sources">
                        {#each Array(5) as _}
                            <div
                                class="news-sidebar-source-item sidebar-loader-item"
                            >
                                <span class="news-source-dot skeleton"></span>
                                <span
                                    class="skeleton skeleton-text sidebar-loader-text"
                                ></span>
                            </div>
                        {/each}
                    </ul>
                    <small
                        class="news-cache-timestamp sidebar-cache-timestamp-loading"
                    >
                        &#128336; Utoljára frissítve: <span
                            class="skeleton skeleton-text sidebar-loader-ts-skeleton"
                        ></span>
                    </small>
                {:else if sources.length > 0}
                    <ul class="news-sidebar-sources">
                        <button
                            class="news-sidebar-source-item news-sidebar-source-all"
                            class:active={selectedSource === null}
                            on:click={() => {
                                selectedSource = null;
                                scrollToTop();
                            }}
                        >
                            <span class="news-source-dot dot-all-sources"
                            ></span>
                            Minden forrás
                        </button>
                        {#each sources as src}
                            <button
                                class="news-sidebar-source-item"
                                class:active={selectedSource === src.title}
                                on:click={() => {
                                    selectedSource =
                                        selectedSource === src.title
                                            ? null
                                            : src.title;
                                    scrollToTop();
                                }}
                            >
                                <span
                                    class="news-source-dot"
                                    style:background={src.bg_color}
                                ></span>
                                {src.title}
                            </button>
                        {/each}
                    </ul>
                {:else}
                    <div class="note warn">
                        Nincsenek elérhető hírforrások. Szerver lekérési hiba.
                    </div>
                {/if}
                {#if !loading && cacheTimestamp}
                    <small class="news-cache-timestamp">
                        &#128336; Utoljára frissítve: {new Date(
                            cacheTimestamp,
                        ).toLocaleString("hu-HU", {
                            month: "short",
                            day: "numeric",
                            hour: "2-digit",
                            minute: "2-digit",
                        })}
                    </small>
                {/if}
            {/if}
        </div>

        <!-- Topics moved to header -->
    </aside>
</div>

{#if selectedSource || selectedWord}
    <div class="filter-actions">
        <div class="info-box">
            <p>
                {#if selectedSource && selectedWord}
                    💡 Leszűrve:<span class="active"
                        >{selectedSource} és {selectedWord}</span
                    >
                {:else if selectedSource}
                    💡 Leszűrve:<span class="active">{selectedSource}</span>
                {:else if selectedWord}
                    💡 Leszűrve:<span class="active">{selectedWord}</span>
                {/if}

                <button
                    class="clear-filters btn btn-sm"
                    aria-label="Szűrők törlése"
                    title="Szűrők törlése"
                    on:click={() => {
                        selectedSource = null;
                        selectedWord = null;
                        scrollToTop();
                    }}
                >
                    Szűrő törlése
                </button>
            </p>
            <p>
                <span>({displayItems.length}/{totalCount})</span>
            </p>
        </div>

        <div class="view-mode-toggle">
            <button
                class="btn btn-sm {viewMode === 'grid' ? 'active' : ''}"
                on:click={() => (viewMode = "grid")}
                title="Rács nézet"
            >
                <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    stroke="currentColor"
                    stroke-width="2"
                    fill="none"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><rect x="3" y="3" width="7" height="7"></rect><rect
                        x="14"
                        y="3"
                        width="7"
                        height="7"
                    ></rect><rect x="14" y="14" width="7" height="7"
                    ></rect><rect x="3" y="14" width="7" height="7"></rect></svg
                >
                <span>Rács</span>
            </button>
            <button
                class="btn btn-sm {viewMode === 'flex' ? 'active' : ''}"
                on:click={() => (viewMode = "flex")}
                title="Lista nézet"
            >
                <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    stroke="currentColor"
                    stroke-width="2"
                    fill="none"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><line x1="8" y1="6" x2="21" y2="6"></line><line
                        x1="8"
                        y1="12"
                        x2="21"
                        y2="12"
                    ></line><line x1="8" y1="18" x2="21" y2="18"></line><line
                        x1="3"
                        y1="6"
                        x2="3.01"
                        y2="6"
                    ></line><line x1="3" y1="12" x2="3.01" y2="12"></line><line
                        x1="3"
                        y1="18"
                        x2="3.01"
                        y2="18"
                    ></line></svg
                >
                <span>Lista</span>
            </button>
        </div>
    </div>
{:else}
    <div class="filter-actions">
        <div class="info-box">
            <p>
                💡 Leszűrve: <span class="active">Összes</span>
            </p>
            <p>
                <span>({displayItems.length}/{totalCount})</span>
            </p>
        </div>

        <div class="view-mode-toggle">
            <button
                class="btn btn-sm {viewMode === 'grid' ? 'active' : ''}"
                on:click={() => (viewMode = "grid")}
                title="Rács nézet"
            >
                <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    stroke="currentColor"
                    stroke-width="2"
                    fill="none"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><rect x="3" y="3" width="7" height="7"></rect><rect
                        x="14"
                        y="3"
                        width="7"
                        height="7"
                    ></rect><rect x="14" y="14" width="7" height="7"
                    ></rect><rect x="3" y="14" width="7" height="7"></rect></svg
                >
                <span>Rács</span>
            </button>
            <button
                class="btn btn-sm {viewMode === 'flex' ? 'active' : ''}"
                on:click={() => (viewMode = "flex")}
                title="Lista nézet"
            >
                <svg
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                    stroke="currentColor"
                    stroke-width="2"
                    fill="none"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><line x1="8" y1="6" x2="21" y2="6"></line><line
                        x1="8"
                        y1="12"
                        x2="21"
                        y2="12"
                    ></line><line x1="8" y1="18" x2="21" y2="18"></line><line
                        x1="3"
                        y1="6"
                        x2="3.01"
                        y2="6"
                    ></line><line x1="3" y1="12" x2="3.01" y2="12"></line><line
                        x1="3"
                        y1="18"
                        x2="3.01"
                        y2="18"
                    ></line></svg
                >
                <span>Lista</span>
            </button>
        </div>
    </div>
{/if}

<section class="faq">
    <h2 class="faq-title">Hogyan működik ez az oldal?</h2>
    <div class="faq-list">
        <details class="faq-item" open>
            <summary>Honnan származnak a hírek?</summary>
            <p>
                Az oldal erdélyi hírforrások RSS-feedjeit olvassa be és jeleníti
                meg időrendi sorrendben. {#if sources.length > 0}A jelenlegi
                    források: {sources
                        .slice(0, -1)
                        .map((s) => s.title)
                        .join(", ")}{sources.length > 1
                        ? ` és ${sources[sources.length - 1].title}`
                        : sources[0].title}.{/if}
            </p>
        </details>
        <details class="faq-item" open>
            <summary>Milyen sűrűn frissülnek a hírek?</summary>
            <p>
                A hírek 30 percenként töltődnek be újra a szerverről. Az első
                látogatás után a böngésző gyorsítótár (localStorage) tárolja az
                adatokat, így a következő 23:30 percen belül történő látogatás
                azonnali.
            </p>
        </details>
        <details class="faq-item" open>
            <summary>Hogyan szűrjük a híreket forrás szerint?</summary>
            <p>
                A jobb oldali „Erdélyi hírforrások” panelen kattints bármelyik
                forrásra, hogy csak az adott oldal híreit lásd. A „Minden
                forrás” gombra kattintva visszatérhetsz a teljes listához.
                Mobilon a panel a lap tetején jelenik meg.
            </p>
        </details>
        <details class="faq-item" open>
            <summary>Mit mutat a „Leggyakoribb témák” panel?</summary>
            <p>
                Ez a panel automatikusan megszámolja a címekben legtöbbször
                előforduló szavakat, kizárva a közönséges ragokat és névelőket.
                Ha egy szó legalább kétszer szerepel, megjelenik a listában.
                Kattintva az adott szavat tartalmazó cikkekre szűr. A számlálás
                megváltozik, ha forrás-szűrőt használsz.
            </p>
        </details>
        <details class="faq-item" open>
            <summary>Elérhetők-e a hírek internet nélkül is?</summary>
            <p>
                Ha a szerver nem érhető el, az oldal az utolsó gyorsítótárban
                tárolt cikkeket mutatja - akár ha régebbiek is. Az utolsó
                frissítés ideje látható a forrás panel alatt.
            </p>
        </details>
        <details class="faq-item" open>
            <summary>Hogyan szűrjük a híreket kulcsszó szerint?</summary>
            <p>
                A jobb oldali „Leggyakoribb témák” panelen kattints bármelyik
                szóra, hogy csak az adott szót tartalmazó cikkeket lásd. A
                „Minden téma” gombra kattintva visszatérhetsz a teljes listához.
                Mobilon a panel a lap tetején jelenik meg.
            </p>
        </details>
    </div>
</section>

<div class="note info">
    A lamsza.com hírlvasó egy ingyenes hírgyűjtő és szűrő szolgáltatás. A hírek
    tartalma és a hozzájuk tartozó képek az eredeti hírforrások szerzői jogi
    védelme alatt állnak. A lamsza.com nem vállal felelősséget ezen források
    tartalmáért. Az oldalon megjelenő időpont vagy dátum azt az időpillanatot
    jelöli, amikor a hírt a rendszerünk indexelte.
</div>

<style>
    .news-word-count {
        opacity: 0.6;
        font-size: 0.85em;
        margin-left: 0.3rem;
    }
    .sidebar-loader-item {
        pointer-events: none;
    }
    .sidebar-loader-text {
        width: 70%;
        height: 0.75rem;
        margin: 0;
    }
    .sidebar-cache-timestamp-loading {
        opacity: 0.45;
    }
    .sidebar-loader-ts-skeleton {
        display: inline-block;
        width: 4rem;
        height: 0.65rem;
        vertical-align: middle;
        margin: 0;
    }
    .dot-all-sources {
        background: var(--border-color);
    }
</style>
