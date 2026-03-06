<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";
    import SearchEngine from "$lib/components/SearchEngine.svelte";
    import MondasWidget from "$lib/components/MondasWidget.svelte";
    import WeatherWidget from "$lib/components/WeatherWidget.svelte";
    import NewsWidget from "$lib/components/NewsWidget.svelte";

    let quickLinksLoading = true;
    let quickLinksData = [];
    let quickLinksError = false;

    onMount(async () => {
        // Quick Links Logic (Simplified)
        const QUICK_LINKS_CACHE_KEY = "quick_links_cache";
        const QUICK_LINKS_TTL = 60 * 60 * 1000;

        const cached = localStorage.getItem(QUICK_LINKS_CACHE_KEY);
        if (cached) {
            try {
                const { items, timestamp } = JSON.parse(cached);
                if (Date.now() - timestamp < QUICK_LINKS_TTL) {
                    quickLinksData = items;
                    quickLinksLoading = false;
                    return;
                }
            } catch (e) {}
        }

        try {
            const data = await apiFetch("/api/admin/quick_links");
            quickLinksData = data || [];
            localStorage.setItem(
                QUICK_LINKS_CACHE_KEY,
                JSON.stringify({
                    items: quickLinksData,
                    timestamp: Date.now(),
                }),
            );
        } catch (err) {
            quickLinksError = true;
        } finally {
            quickLinksLoading = false;
        }
    });
</script>

<div class="home-main">
    <h1 class="page-title">Székely Gugel</h1>
    <p class="greeting">Az internet székely kapuja.</p>

    <SearchEngine />
</div>

<MondasWidget />

{#if quickLinksLoading || quickLinksData.length > 0 || quickLinksError}
    <section id="gyorslinkek">
        <div class="quick-links-heading">
            <span class="heading-label">Gyorslinkek</span>
        </div>
        <div class="quick-links">
            {#if quickLinksLoading}
                {#each Array(6) as _}
                    <div class="link-card link-card--skeleton">
                        <div class="link-card-icon skeleton"></div>
                        <div
                            class="link-card-title skeleton skeleton-text"
                        ></div>
                    </div>
                {/each}
            {:else if quickLinksError}
                <p class="error-msg quick-links-error">
                    A gyorslinkek jelenleg nem elérhetők.
                </p>
            {:else}
                {#each quickLinksData as q}
                    <a
                        href={q.url}
                        target="_blank"
                        rel="nofollow noopener"
                        class="link-card"
                    >
                        <div
                            class="link-card-icon"
                            style:background={q.bg_color || "#2f4f4f"}
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="28"
                                height="28"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="var(--border-color)"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                ><path
                                    d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"
                                ></path><polyline points="15 3 21 3 21 9"
                                ></polyline><line x1="10" y1="14" x2="21" y2="3"
                                ></line></svg
                            >
                        </div>
                        <span class="link-card-title">{q.title}</span>
                    </a>
                {/each}
            {/if}
        </div>
    </section>
{/if}

<div class="widgets-row">
    <WeatherWidget />
    <NewsWidget limit={10} />
</div>
