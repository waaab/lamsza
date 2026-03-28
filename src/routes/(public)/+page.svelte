<script>
    import { onMount } from "svelte";
    import { auth } from "$lib/stores/auth";
    import { apiFetch } from "$lib/api";
    import SearchEngine from "$lib/components/SearchEngine.svelte";
    import MondasWidget from "$lib/components/MondasWidget.svelte";
    import WeatherWidget from "$lib/components/WeatherWidget.svelte";
    import DateTimeWidget from "$lib/components/DateTimeWidget.svelte";
    import NewsWidget from "$lib/components/NewsWidget.svelte";
    import EventsWidget from "$lib/components/EventsWidget.svelte";

    const USER_LINKS_KEY = "user_quick_links";
    const PROMOTED_CACHE_KEY = "promoted_links_cache";
    const PROMOTED_TTL = 60 * 60 * 1000;

    let userLinks = [];
    let promotedLinks = [];
    let promotedLoading = true;
    let promotedError = false;

    let linkDialogOpen = false;
    let linkDialogMode = "add";
    let linkDialogData = { id: "", title: "", url: "", bg_color: "#e6f0ff" };
    let canScrollLeft = false;
    let canScrollRight = false;
    let quickLinksContainer;
    const LINKS_BEFORE_ARROWS = 7;

    let myLocationSlug = "csikszereda";
    let myLocationName = "";
    let myLocationCountySlug = "";

    $: allLinksCount = userLinks.length + promotedLinks.length;
    $: showArrows = !promotedLoading && allLinksCount > LINKS_BEFORE_ARROWS;

    function truncateTitle(title, maxLen = 8) {
        if (!title || title.length <= maxLen) return title;
        return title.slice(0, maxLen) + "...";
    }

    function checkScroll(node) {
        if (!node) return;
        canScrollLeft = node.scrollLeft > 5;
        canScrollRight = node.scrollLeft < node.scrollWidth - node.clientWidth - 5;
    }

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
            const walk = (x - startX) * 2;
            node.scrollLeft = scrollLeft - walk;
            checkScroll(node);
        };

        const onScroll = () => checkScroll(node);
        const onResize = () => checkScroll(node);

        node.addEventListener("mousedown", onMouseDown);
        node.addEventListener("mouseleave", onMouseLeave);
        node.addEventListener("mouseup", onMouseUp);
        node.addEventListener("mousemove", onMouseMove);
        node.addEventListener("scroll", onScroll);
        window.addEventListener("resize", onResize);
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

    function scrollLinks(amount) {
        if (quickLinksContainer) {
            quickLinksContainer.scrollBy({ left: amount, behavior: "smooth" });
        }
    }

    function generateId() {
        return Date.now().toString(36) + Math.random().toString(36).slice(2, 7);
    }

    function loadUserLinks() {
        try {
            const raw = localStorage.getItem(USER_LINKS_KEY);
            userLinks = raw ? JSON.parse(raw) : [];
        } catch { userLinks = []; }
    }

    function saveUserLinks() {
        localStorage.setItem(USER_LINKS_KEY, JSON.stringify(userLinks));
    }

    function openAddLink() {
        linkDialogMode = "add";
        linkDialogData = { id: "", title: "", url: "", bg_color: "#e6f0ff" };
        linkDialogOpen = true;
    }

    function openEditLink(q, e) {
        e?.preventDefault?.();
        e?.stopPropagation?.();
        linkDialogMode = "edit";
        linkDialogData = { id: q.id, title: q.title, url: q.url, bg_color: q.bg_color || "#e6f0ff" };
        linkDialogOpen = true;
    }

    function closeLinkDialog() {
        linkDialogOpen = false;
    }

    function saveLinkDialog(e) {
        e.preventDefault();
        if (linkDialogMode === "add") {
            userLinks = [...userLinks, { id: generateId(), title: linkDialogData.title, url: linkDialogData.url, bg_color: linkDialogData.bg_color }];
        } else {
            userLinks = userLinks.map(l => l.id === linkDialogData.id ? { ...l, title: linkDialogData.title, url: linkDialogData.url, bg_color: linkDialogData.bg_color } : l);
        }
        saveUserLinks();
        closeLinkDialog();
    }

    function deleteLinkFromDialog(e) {
        e.preventDefault();
        if (!linkDialogData.id) return;
        userLinks = userLinks.filter(l => l.id !== linkDialogData.id);
        saveUserLinks();
        closeLinkDialog();
    }

    onMount(async () => {
        loadUserLinks();

        let cacheVersion = null;
        try {
            const configRes = await apiFetch("/api/config/public");
            if (configRes && configRes.quick_links_version != null) {
                cacheVersion = String(configRes.quick_links_version);
            }
            if (configRes?.my_location_slug) {
                myLocationSlug = configRes.my_location_slug;
                myLocationName = configRes.my_location_name || "";
                myLocationCountySlug = configRes.my_location_county_slug || "";
            }
        } catch (_) {}

        const cached = localStorage.getItem(PROMOTED_CACHE_KEY);
        if (cached && cacheVersion) {
            try {
                const { items, timestamp, version } = JSON.parse(cached);
                if (version === cacheVersion && Date.now() - timestamp < PROMOTED_TTL) {
                    promotedLinks = items;
                    promotedLoading = false;
                    return;
                }
            } catch (e) {}
        }

        try {
            const data = await apiFetch("/api/admin/quick_links");
            promotedLinks = data || [];
            localStorage.setItem(
                PROMOTED_CACHE_KEY,
                JSON.stringify({
                    items: promotedLinks,
                    timestamp: Date.now(),
                    version: cacheVersion ?? "",
                }),
            );
        } catch (err) {
            promotedError = true;
        } finally {
            promotedLoading = false;
        }
    });
</script>

<section id="home main" class="home-main">
    <h1 class="page-title">{$auth.loggedIn ? `Szerussz, ${$auth.user}!` : "Na Lámsza!"}</h1>
    <p class="greeting">Erdélyi magyar startlap és kereső. Az internet székely kapuja.</p>

    <SearchEngine />
</section>

<section id="home widgets" class="widgets-columns">
    <div class="widgets-box--three-col">
        <div id="gyorslinkek" class="widget">
            <div class="widget-header">
                <h3 class="widget-title">Gyorslinkek</h3>
            </div>
                <div
                    class="quick-links-wrapper"
                    class:can-left={showArrows && canScrollLeft}
                    class:can-right={showArrows && canScrollRight}
                >
                    {#if showArrows}
                        <button class="scroll-arrow left" aria-label="Görgetés balra" on:click={() => scrollLinks(-200)}>&#8249;</button>
                    {/if}
                    <div class="quick-links widget-content" use:dragScroll bind:this={quickLinksContainer}>
                        <button
                            type="button"
                            class="link-card link-card-add"
                            on:click={openAddLink}
                            title="Új gyorslink hozzáadása"
                            aria-label="Új gyorslink hozzáadása"
                        >
                            <span class="link-card-icon link-card-icon-add">
                                <span class="link-card-add-plus">+</span>
                            </span>
                            <span class="link-card-title">Új</span>
                        </button>

                        {#each userLinks as q}
                            <div class="link-card">
                                <a
                                    href={q.url}
                                    target="_blank"
                                    rel="nofollow noopener"
                                    class="link-card-link"
                                    title={q.title}>
                                    <span class="link-card-icon" style:background={q.bg_color || "#2f4f4f"}>
                                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--border-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
                                    </span>
                                    <span class="link-card-title">{truncateTitle(q.title)}</span>
                                </a>
                                <button
                                    type="button"
                                    class="link-card-edit"
                                    on:click={(e) => openEditLink(q, e)}
                                    title="Szerkesztés"
                                    aria-label="Szerkesztés"
                                >
                                <svg xmlns="http://www.w3.org/2000/svg"
                                width="10"
                                height="10" 
                                viewBox="0 0 24 24" 
                                fill="none" 
                                stroke="currentColor" 
                                stroke-width="2" 
                                stroke-linecap="round" 
                                stroke-linejoin="round" >
                                    <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path>
                                </svg>
                            </button>
                            </div>
                        {/each}

                        {#if !promotedError}
                            {#each promotedLinks as q}
                                <div class="link-card link-card--promoted">
                                    <a
                                        href={q.url}
                                        target="_blank"
                                        rel="nofollow noopener"
                                        class="link-card-link"
                                        title={q.title}>
                                        <span class="link-card-icon" style:background={q.bg_color || "#2f4f4f"}>
                                            <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="var(--border-color)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>
                                        </span>
                                        <span class="link-card-title">{truncateTitle(q.title)}</span>
                                    </a>
                                    <span class="link-card-promoted-badge" title="Promóciós link" aria-label="Promóciós link">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"></polygon></svg>
                                    </span>
                                </div>
                            {/each}
                        {/if}
                    </div>
                    {#if showArrows}
                        <button class="scroll-arrow right" aria-label="Görgetés jobbra" on:click={() => scrollLinks(200)}>&#8250;</button>
                    {/if}
                </div>
            </div>
        <DateTimeWidget />
        <WeatherWidget settlementSlug={myLocationSlug} />
    </div>
</section>

<EventsWidget ticker={true} settlementSlug={myLocationSlug} locationName={myLocationName} />

<MondasWidget />

<section id="hirek">
    <NewsWidget limit={10} />
</section>

<!-- Quick Link Add/Edit Dialog -->
{#if linkDialogOpen}
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div
        class="link-dialog-overlay"
        role="dialog"
        aria-labelledby="link-dialog-title"
        tabindex="-1"
        on:click|self={closeLinkDialog}
        on:keydown={(e) => e.key === "Escape" && closeLinkDialog()}
    >
        <div class="link-dialog" on:click|stopPropagation>
            <h3 id="link-dialog-title">{linkDialogMode === "add" ? "Új gyorslink hozzáadása" : "Gyorslink szerkesztése"}</h3>
            <form class="link-dialog-form" on:submit|preventDefault={saveLinkDialog}>
                <label for="link_dialog_title">Cím</label>
                <input
                    id="link_dialog_title"
                    type="text"
                    bind:value={linkDialogData.title}
                    required
                />

                <label for="link_dialog_url">URL</label>
                <input
                    id="link_dialog_url"
                    type="url"
                    bind:value={linkDialogData.url}
                    required
                />

                <label for="link_dialog_color">Háttérszín (pl. #e6f0ff)</label>
                <input
                    id="link_dialog_color"
                    type="text"
                    bind:value={linkDialogData.bg_color}
                    placeholder="#e6f0ff"
                />

                <div class="link-dialog-actions">
                    <button type="submit" class="link-dialog-submit">Mentés</button>
                    {#if linkDialogMode === "edit"}
                        <button type="button" class="link-dialog-delete" on:click={deleteLinkFromDialog}>Törlés</button>
                    {/if}
                    <button type="button" class="link-dialog-cancel" on:click={closeLinkDialog}>Mégse</button>
                </div>
            </form>
        </div>
    </div>
{/if}

<style>
    .widgets-box--three-col {
        grid-column: span 3;
    }
</style>