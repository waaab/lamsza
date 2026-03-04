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

    $: filteredServices = services.filter(
        (s) =>
            currentCategory === "osszes" ||
            s.category.toLowerCase() === currentCategory.toLowerCase(),
    );
    $: totalCount = filteredServices.length;
    $: displayItems = filteredServices.slice(0, visibleCount);

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

<div class="container" style="min-height: calc(100vh - 120px)">
    <h1 class="page-title">Index</h1>
    <p class="greeting">Keresd meg a helyi szakembereket és intézményeket!</p>

    <div class="header-tabs">
        <span class="header-tabs-label">Kategóriák:</span>
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
            <button
                class="view-btn {viewMode === 'grid' ? 'active' : ''}"
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
                class="view-btn {viewMode === 'flex' ? 'active' : ''}"
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
                <article
                    class="card service--skeleton"
                    style="height: 150px; display: flex; flex-direction: column; padding: 1rem; gap: 0.5rem;"
                >
                    <div
                        class="skeleton skeleton-text"
                        style="width: 30%;"
                    ></div>
                    <div
                        class="skeleton skeleton-text"
                        style="width: 80%; margin-top: 0.5rem; height: 1.2rem;"
                    ></div>
                    <div
                        class="skeleton skeleton-text"
                        style="width: 60%; margin-top: auto;"
                    ></div>
                </article>
            {/each}
        </div>
    {:else if error}
        <div class="error-msg">{error}</div>
    {:else if services.length === 0}
        <div class="error-msg">
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
                                style="color:inherit;text-decoration:none;"
                                >{service.name}</a
                            >
                        {:else}
                            <a
                                href="/bejegyzes/{service.slug}"
                                style="color:inherit;text-decoration:none;"
                                >{service.name}</a
                            >
                        {/if}
                    </h3>
                    {#if service.url}
                        <div
                            class="service-info"
                            style="margin-bottom: 0.5rem;"
                        >
                            <span
                                style="color: var(--text-faint); margin-right: 0.3rem;"
                                >🔗</span
                            >
                            <a
                                href={service.url}
                                target="_blank"
                                rel="nofollow noopener"
                                style="color: var(--primary-color); text-decoration: none;"
                                >{service.url}</a
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
                    Több betöltése
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
            <button
                class="view-btn {viewMode === 'grid' ? 'active' : ''}"
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
                class="view-btn {viewMode === 'flex' ? 'active' : ''}"
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
</div>
