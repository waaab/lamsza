<script>
    import { browser } from "$app/environment";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import { onMount } from "svelte";

    let dynamicCategories = [{ id: "osszes", label: "Összes", url: "/index" }];
    let services = [];
    let loading = true;
    let error = null;

    let viewMode = "grid";
    let currentCategory = "";
    let visibleCount = 12;
    let sortMode = "title";
    let sortOpen = false;

    const sortLabels = { title: "Név (A→Z)", newest: "Legújabb" };

    function setSortMode(mode) {
        sortMode = mode;
        sortOpen = false;
    }

    $: filteredServices = services.filter(
        (s) =>
            currentCategory === "osszes" ||
            s.category.toLowerCase() === currentCategory.toLowerCase(),
    );
    $: sortedServices = [...filteredServices].sort((a, b) => {
        if (sortMode === "newest") return b.id - a.id;
        return a.name.localeCompare(b.name);
    });
    $: totalCount = sortedServices.length;
    $: displayItems = sortedServices.slice(0, visibleCount);

    function loadMore() {
        visibleCount += 12;
    }

    // Reset pagination when category changes
    $: if (currentCategory) {
        visibleCount = 12;
    }

    // Reactively fetch data when category changes
    $: if (browser) {
        const categoryId = $page.params.category;
        currentCategory = categoryId;
        fetchData(categoryId);
    }

    async function fetchData(categoryId) {
        loading = true;
        error = null;
        try {
            const apiBase = import.meta.env.VITE_API_BASE_URL;
            if (!apiBase)
                console.warn(
                    "VITE_API_BASE_URL is not set. Falling back to http://localhost:3000",
                );
            const baseUrl = apiBase || "http://localhost:3000";
            const res = await fetch(`${baseUrl}/api/directory`);
            if (!res.ok) throw new Error("Hálózati hiba");
            const allServices = (await res.json()) || [];

            const uniqueCats = new Set(
                allServices.map((s) => s.category).filter((c) => c),
            );
            const generatedCats = Array.from(uniqueCats).map((catName) => {
                return {
                    id: catName,
                    label: catName,
                    url: "/index/" + encodeURIComponent(catName),
                };
            });
            dynamicCategories = [
                { id: "osszes", label: "Összes", url: "/index" },
                ...generatedCats,
            ];

            services = allServices.filter((s) => {
                // Keep backward compatibility if the URL contains hardcoded old slugs (e.g. 'egeszsegugy'),
                // otherwise strictly match explicit category ID strings.
                return (
                    s.category.toLowerCase() === categoryId.toLowerCase() ||
                    (categoryId === "egeszsegugy" &&
                        s.category === "Egészségügy") ||
                    (categoryId === "oktatas" && s.category === "Oktatás") ||
                    (categoryId === "mesteremberek" &&
                        s.category === "Mesteremberek") ||
                    (categoryId === "hivatalok" &&
                        s.category === "Hivatalok") ||
                    (categoryId === "egyeb" && s.category === "Egyéb") ||
                    s.category_id === categoryId
                ); // backup fallback
            });
        } catch (err) {
            console.error(err);
            error = "Hiba történt az adatok betöltésekor.";
        } finally {
            loading = false;
        }
    }
</script>

<svelte:head>
    <title>Szekely Gugel - Index</title>
</svelte:head>

<div class="container main-container">
    <h1 class="page-title">Index</h1>
    <p class="greeting">Keresd meg a helyi szakembereket és intézményeket!</p>

    <div class="header-tabs">
        <span class="header-tabs-label">Kiemelt Kategóriák:</span>
        {#if loading}
            {#each Array(6) as _}
                <div class="btn-skeleton"></div>
            {/each}
        {:else}
            {#each dynamicCategories as cat}
                <button
                    class="btn btn-md {cat.id === currentCategory ||
                    cat.id.toLowerCase() === currentCategory.toLowerCase()
                        ? 'active'
                        : ''}"
                    on:click={() => goto(cat.url)}>{cat.label}</button
                >
            {/each}
        {/if}
    </div>

    <div class="filter-actions">
        <div class="info-box">
            <p>
                💡 Leszűrve: <span class="active">{currentCategory}</span>
                <button
                    on:click={() => goto("/index")}
                    class="clear-filters btn btn-sm">Szűrő törlése</button
                >
            </p>
            <p>({displayItems.length}/{totalCount})</p>
        </div>

        <div class="view-mode-toggle">
            <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
            <div class="sort-toggle">
                <button
                    class="btn btn-sm"
                    on:click={() => (sortOpen = !sortOpen)}
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        ><line x1="4" y1="6" x2="16" y2="6"></line><line
                            x1="4"
                            y1="12"
                            x2="12"
                            y2="12"
                        ></line><line x1="4" y1="18" x2="8" y2="18"
                        ></line><polyline points="15 15 18 18 21 15"
                        ></polyline><line x1="18" y1="10" x2="18" y2="18"
                        ></line></svg
                    >
                    <span>{sortLabels[sortMode]}</span>
                </button>
                {#if sortOpen}
                    <div class="sort-toggle-menu" on:click|stopPropagation>
                        <button
                            class:active={sortMode === "title"}
                            on:click={() => setSortMode("title")}
                            >Név (A→Z)</button
                        >
                        <button
                            class:active={sortMode === "newest"}
                            on:click={() => setSortMode("newest")}
                            >Legújabb</button
                        >
                    </div>
                {/if}
            </div>

            <button
                class="btn btn-sm {viewMode === 'grid' ? 'active' : ''}"
                on:click={() => (viewMode = "grid")}
                title="Rács nézet"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
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
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
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

    {#if loading}
        <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
            {#each Array(6) as _}
                <article class="card service--skeleton">
                    <div class="skeleton skeleton-text skeleton-cat"></div>
                    <div class="skeleton skeleton-text skeleton-title"></div>
                    <div class="skeleton skeleton-text skeleton-loc"></div>
                </article>
            {/each}
        </div>
    {:else if error}
        <div class="note error">{error}</div>
    {:else if services.length === 0}
        <div class="note error">
            Nincs megjeleníthető bejegyzés ebben a kategóriában.
        </div>
    {:else}
        <div class="list {viewMode === 'grid' ? 'grid' : 'flex'}">
            {#each displayItems as service}
                <article class="card service">
                    <span class="badge">
                        {service.category}
                    </span>
                    <h3 class="service-name">
                        {#if service.entity_type === "settlement"}
                            <a
                                href="/{service.county_slug}-megye/{service.slug}"
                                class="service-link">{service.name}</a
                            >
                        {:else}
                            <a
                                href="/bejegyzes/{service.slug}"
                                class="service-link">{service.name}</a
                            >
                        {/if}
                    </h3>
                    {#if service.url}
                        <div class="service-info service-url-wrap">
                            <span class="service-url-icon">🔗</span>
                            <a
                                href={service.url}
                                target="_blank"
                                rel="nofollow noopener"
                                class="service-url-link">{service.url}</a
                            >
                        </div>
                    {/if}
                    <div class="service-info">
                        📍 {service.location} - {service.address}
                    </div>
                    <div class="service-info">📞 {service.phone}</div>
                    {#if service.notes}
                        <div class="service-notes">{service.notes}</div>
                    {/if}
                </article>
            {/each}
        </div>

        {#if visibleCount < totalCount}
            <div class="load-more">
                <button class="nav-btn" on:click={loadMore}>
                    Több betöltése ↓
                </button>
            </div>
        {/if}
    {/if}

    <div class="filter-actions">
        <div class="info-box">
            <p>
                💡 Leszűrve: <span class="active">{currentCategory}</span>
                <button
                    on:click={() => goto("/index")}
                    class="clear-filters btn btn-sm">Szűrő törlése</button
                >
            </p>
            <p>({displayItems.length}/{totalCount})</p>
        </div>

        <div class="view-mode-toggle">
            <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
            <div class="sort-toggle">
                <button
                    class="btn btn-sm"
                    on:click={() => (sortOpen = !sortOpen)}
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="16"
                        height="16"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        ><line x1="4" y1="6" x2="16" y2="6"></line><line
                            x1="4"
                            y1="12"
                            x2="12"
                            y2="12"
                        ></line><line x1="4" y1="18" x2="8" y2="18"
                        ></line><polyline points="15 15 18 18 21 15"
                        ></polyline><line x1="18" y1="10" x2="18" y2="18"
                        ></line></svg
                    >
                    <span>{sortLabels[sortMode]}</span>
                </button>
                {#if sortOpen}
                    <div class="sort-toggle-menu" on:click|stopPropagation>
                        <button
                            class:active={sortMode === "title"}
                            on:click={() => setSortMode("title")}
                            >Név (A→Z)</button
                        >
                        <button
                            class:active={sortMode === "newest"}
                            on:click={() => setSortMode("newest")}
                            >Legújabb</button
                        >
                    </div>
                {/if}
            </div>

            <button
                class="btn btn-sm {viewMode === 'grid' ? 'active' : ''}"
                on:click={() => (viewMode = "grid")}
                title="Rács nézet"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
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
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
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

    <section class="faq">
        <h2 class="faq-title">Gyakori kérdések</h2>
        <div class="faq-list">
            <details class="faq-item" open>
                <summary>Honnan származnak az adatok?</summary>
                <p>
                    Az adatok a helyi szakemberektől és intézményektől
                    származnak, akiket a rendszerünk folyamatosan indexel, hogy
                    a legfrissebb elérhetőségeket biztosítsa.
                </p>
            </details>
            <details class="faq-item" open>
                <summary>Hogyan kerülhet be valaki a címtárba?</summary>
                <p>
                    A beküldési folyamat hamarosan elérhető lesz az oldalon.
                    Addig is, ha ismersz olyan szolgáltatót, aki még nem
                    szerepel nálunk, keress minket bizalommal.
                </p>
            </details>
            <details class="faq-item" open>
                <summary>Ingyenes-e a megjelenés?</summary>
                <p>
                    Igen, az alapvető megjelenés és az adatok listázása teljesen
                    ingyenes minden helyi szolgáltató, mesterember és intézmény
                    számára.
                </p>
            </details>
            <details class="faq-item" open>
                <summary>Hogyan működik a keresés?</summary>
                <p>
                    A keresőnk kulcsszavak, kategóriák és települések alapján
                    szűri a találatokat. A kereső prioritást ad a közvetlen név-
                    és kategória-egyezéseknek, de a leírásokban is keres.
                </p>
            </details>
        </div>
    </section>

    <div class="note info">
        A lamsza.com indexe egy ingyenes információs szolgáltatás. Az adatok
        pontosságáért és a szolgáltatások minőségéért a lamsza.com nem vállal
        felelősséget. Kérjük, minden esetben ellenőrizze az adatokat a
        szolgáltatóval való kapcsolatfelvétel előtt.
    </div>
</div>

<style>
    .main-container {
        min-height: calc(100vh - 120px);
    }
    .service--skeleton {
        height: 150px;
        display: flex;
        flex-direction: column;
        padding: 1rem;
        gap: 0.5rem;
    }
    .skeleton-cat {
        width: 30%;
    }
    .skeleton-title {
        width: 80%;
        margin-top: 0.5rem;
        height: 1.2rem;
    }
    .skeleton-loc {
        width: 60%;
        margin-top: auto;
    }
    .service-link {
        color: inherit;
        text-decoration: none;
    }
    .service-url-wrap {
        margin-bottom: 0.5rem;
    }
    .service-url-icon {
        color: var(--text-faint);
        margin-right: 0.3rem;
    }
    .service-url-link {
        color: var(--primary-color);
        text-decoration: none;
    }
</style>
