<script>
    import { onMount, tick } from "svelte";

    let authenticated = false;
    let password = "";

    let activeTab = "mondasok";

    let mondasok = [];
    let quickLinks = [];
    let newsFeeds = [];
    let locations = [];
    let services = [];
    let serviceCategories = [];

    // Per-feed loading state
    let loadingFeeds = new Set();
    let feedTimestamps = {};

    // Form binding objects
    let newMondas = { text: "" };
    let newLink = { title: "", url: "", bg_color: "var(--card-bg)" };
    let newNews = { title: "", feed_url: "", bg_color: "var(--warning-bg)" };
    let newLocation = { name: "", county: "", type: "" };
    let newService = {
        location_id: "",
        category_id: "",
        name: "",
        url: "",
        phone: "",
        address: "",
        notes: "",
        is_magyar_language: true,
        tags: "",
    };
    let newServiceCategory = { name: "" };

    onMount(() => {
        if (localStorage.getItem("admin_auth") === "true") {
            authenticated = true;
            fetchAll();
        }

        const storedTs = localStorage.getItem("news_feed_timestamps");
        if (storedTs) {
            try {
                feedTimestamps = JSON.parse(storedTs);
            } catch (e) {}
        }
    });

    function login(e) {
        e.preventDefault();
        if (password === "szekely123") {
            authenticated = true;
            localStorage.setItem("admin_auth", "true");
            fetchAll();
        } else {
            alert("Na de kicsibarátom, ez nem a jó jelszó!");
        }
    }

    function logout() {
        authenticated = false;
        password = "";
        localStorage.removeItem("admin_auth");
    }

    async function fetchAll() {
        fetchMondasok();
        fetchQuickLinks();
        fetchNewsFeeds();
        fetchLocations();
        fetchServices();
        fetchServiceCategories();
    }

    // generic fetch
    async function loadData(endpoint, setter) {
        try {
            const apiBase = import.meta.env.VITE_API_BASE_URL;
            if (!apiBase)
                console.warn(
                    "VITE_API_BASE_URL is not set. Falling back to http://localhost:3000",
                );
            const baseUrl = apiBase || "http://localhost:3000";

            const res = await fetch(`${baseUrl}/api/admin/${endpoint}`);
            if (res.ok) setter(await res.json());
        } catch (e) {
            console.error(e);
        }
    }

    // specific fetches
    function fetchMondasok() {
        loadData("mondasok", (d) => (mondasok = d));
    }
    function fetchQuickLinks() {
        loadData("quick_links", (d) => (quickLinks = d));
    }
    function fetchNewsFeeds() {
        loadData("news_feeds", (d) => (newsFeeds = d));
    }
    function fetchLocations() {
        loadData("locations", (d) => (locations = d));
    }
    function fetchServices() {
        loadData("services", (d) => (services = d));
    }
    function fetchServiceCategories() {
        loadData("service_categories", (d) => (serviceCategories = d));
    }

    // generic create
    async function createRecord(endpoint, data, reloadFunc, resetFormFunc) {
        try {
            const apiBase = import.meta.env.VITE_API_BASE_URL;
            if (!apiBase)
                console.warn(
                    "VITE_API_BASE_URL is not set. Falling back to http://localhost:3000",
                );
            const baseUrl = apiBase || "http://localhost:3000";

            const res = await fetch(`${baseUrl}/api/admin/${endpoint}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(data),
            });
            if (res.ok) {
                reloadFunc();
                resetFormFunc();
            } else {
                alert("Hiba történt a mentés során.");
            }
        } catch (e) {
            console.error(e);
        }
    }

    // generic delete
    async function deleteRecord(endpoint, id, reloadFunc) {
        if (!confirm("Biztosan törölni szeretnéd?")) return;
        try {
            const apiBase = import.meta.env.VITE_API_BASE_URL;
            if (!apiBase)
                console.warn(
                    "VITE_API_BASE_URL is not set. Falling back to http://localhost:3000",
                );
            const baseUrl = apiBase || "http://localhost:3000";

            const res = await fetch(
                `${baseUrl}/api/admin/${endpoint}?id=${id}`,
                {
                    method: "DELETE",
                },
            );
            if (res.ok) {
                reloadFunc();
            }
        } catch (e) {
            console.error(e);
        }
    }

    // specific creates
    function submitMondas(e) {
        e.preventDefault();
        createRecord(
            "mondasok",
            newMondas,
            fetchMondasok,
            () => (newMondas = { text: "" }),
        );
    }
    function submitLink(e) {
        e.preventDefault();
        createRecord(
            "quick_links",
            newLink,
            fetchQuickLinks,
            () =>
                (newLink = {
                    title: "",
                    url: "",
                    bg_color: "var(--card-bg)",
                }),
        );
    }
    function submitNews(e) {
        e.preventDefault();
        createRecord(
            "news_feeds",
            newNews,
            fetchNewsFeeds,
            () =>
                (newNews = {
                    title: "",
                    feed_url: "",
                    bg_color: "var(--warning-bg)",
                }),
        );
    }

    async function updateSingleFeed(feed) {
        // Mark this feed as loading
        loadingFeeds.add(feed.id);
        loadingFeeds = new Set(loadingFeeds);

        try {
            const apiBase = import.meta.env.VITE_API_BASE_URL;
            if (!apiBase)
                console.warn(
                    "VITE_API_BASE_URL is not set. Falling back to http://localhost:3000",
                );
            const baseUrl = apiBase || "http://localhost:3000";

            const proxiedUrl =
                `${baseUrl}/api/proxy?url=` + encodeURIComponent(feed.feed_url);
            const res = await fetch(proxiedUrl);
            if (res.ok) {
                feedTimestamps[feed.feed_url] = Date.now();
                localStorage.setItem(
                    "news_feed_timestamps",
                    JSON.stringify(feedTimestamps),
                );
                feedTimestamps = { ...feedTimestamps };
                // Invalidate combined cache so frontend re-fetches on next load
                localStorage.removeItem("news_cache");
            }
        } catch (e) {
            console.error("Feed frissítési hiba:", e);
        } finally {
            loadingFeeds.delete(feed.id);
            loadingFeeds = new Set(loadingFeeds);
        }
    }
    function submitLocation(e) {
        e.preventDefault();
        createRecord(
            "locations",
            newLocation,
            fetchLocations,
            () => (newLocation = { name: "", county: "", type: "" }),
        );
    }
    function submitService(e) {
        e.preventDefault();
        newService.location_id = parseInt(newService.location_id);
        if (newService.category_id) {
            newService.category_id = parseInt(newService.category_id);
        }
        createRecord(
            "services",
            newService,
            fetchServices,
            () =>
                (newService = {
                    location_id: "",
                    category_id: "",
                    name: "",
                    url: "",
                    phone: "",
                    address: "",
                    notes: "",
                    is_magyar_language: true,
                    tags: "",
                }),
        );
    }

    function submitServiceCategory(e) {
        e.preventDefault();
        createRecord(
            "service_categories",
            newServiceCategory,
            fetchServiceCategories,
            () => (newServiceCategory = { name: "" }),
        );
    }

    function getLocationName(id) {
        const l = locations.find((loc) => loc.id === id);
        return l ? l.name : id;
    }

    function getCategoryName(id) {
        const c = serviceCategories.find((cat) => cat.id === id);
        return c ? c.name : "N/A";
    }
</script>

<svelte:head>
    <title>Székely Gugel - Adminisztráció</title>
</svelte:head>

{#if !authenticated}
    <div class="container">
        <div class="admin-login-wrapper">
            <div
                class="admin-container"
                style="max-width: 400px; text-align:center; width: 100%;"
            >
                <h2>Adminisztráció Belépés</h2>
                <form
                    class="admin-form"
                    style="margin: 2rem auto;"
                    on:submit={login}
                >
                    <input
                        type="password"
                        bind:value={password}
                        placeholder="Jelszó..."
                        required
                    />
                    <button type="submit" class="admin-submit-btn"
                        >Belépés</button
                    >
                </form>
            </div>
        </div>
    </div>
{:else}
    <div class="admin-layout">
        <aside class="admin-sidebar">
            <a href="/" class="admin-sidebar-btn" title="Vissza a Gugelbe">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
                    ><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"
                    ></path><polyline points="9 22 9 12 15 12 15 22"
                    ></polyline></svg
                >
            </a>

            <button
                class="admin-sidebar-btn {activeTab === 'mondasok'
                    ? 'active'
                    : ''}"
                on:click={() => (activeTab = "mondasok")}
                title="Mondások"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
                    ><path
                        d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"
                    ></path></svg
                >
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'quicklinks'
                    ? 'active'
                    : ''}"
                on:click={() => (activeTab = "quicklinks")}
                title="Gyorslinkek"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
                    ><path
                        d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"
                    ></path><path
                        d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"
                    ></path></svg
                >
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'newsfeeds'
                    ? 'active'
                    : ''}"
                on:click={() => (activeTab = "newsfeeds")}
                title="Hírfolyamok"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
                    ><path d="M4 11a9 9 0 0 1 9 9"></path><path
                        d="M4 4a16 16 0 0 1 16 16"
                    ></path><circle cx="5" cy="19" r="1"></circle></svg
                >
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'locations'
                    ? 'active'
                    : ''}"
                on:click={() => (activeTab = "locations")}
                title="Települések"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
                    ><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"
                    ></path><circle cx="12" cy="10" r="3"></circle></svg
                >
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'service_categories'
                    ? 'active'
                    : ''}"
                on:click={() => (activeTab = "service_categories")}
                title="Szolgáltatás Kategóriák"
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                >
                    <line x1="8" y1="6" x2="21" y2="6"></line>
                    <line x1="8" y1="12" x2="21" y2="12"></line>
                    <line x1="8" y1="18" x2="21" y2="18"></line>
                    <line x1="3" y1="6" x2="3.01" y2="6"></line>
                    <line x1="3" y1="12" x2="3.01" y2="12"></line>
                    <line x1="3" y1="18" x2="3.01" y2="18"></line>
                </svg>
            </button>
            <button
                class="admin-sidebar-btn {activeTab === 'services'
                    ? 'active'
                    : ''}"
                on:click={() => (activeTab = "services")}
                title="Szolgáltatások"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
                    ><rect x="3" y="3" width="7" height="7"></rect><rect
                        x="14"
                        y="3"
                        width="7"
                        height="7"
                    ></rect><rect x="14" y="14" width="7" height="7"
                    ></rect><rect x="3" y="14" width="7" height="7"></rect></svg
                >
            </button>
        </aside>

        <main class="admin-main">
            <div class="admin-header">
                <h2>
                    {#if activeTab === "mondasok"}Mondások Kezelése{/if}
                    {#if activeTab === "quicklinks"}Gyorslinkek Kezelése{/if}
                    {#if activeTab === "newsfeeds"}Hírfolyamok Kezelése{/if}
                    {#if activeTab === "locations"}Települések Kezelése{/if}
                    {#if activeTab === "service_categories"}Szolgáltatás
                        Kategóriák Kezelése{/if}
                    {#if activeTab === "services"}Szolgáltatások Kezelése{/if}
                </h2>
                <button class="btn-logout" on:click={logout}
                    >Kijelentkezés</button
                >
            </div>

            <div class="admin-container" style="margin:0; max-width:100%;">
                <!-- Mondások Tab -->
                {#if activeTab === "mondasok"}
                    <h3>Új mondás hozzáadása</h3>
                    <form class="admin-form" on:submit={submitMondas}>
                        <label for="mondas_text">Mondás szövege</label>
                        <textarea
                            id="mondas_text"
                            bind:value={newMondas.text}
                            required
                            rows="3"
                        ></textarea>
                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>

                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Szöveg</th>
                                    <th>Művelet</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each mondasok as m}
                                    <tr>
                                        <td>{m.id}</td>
                                        <td>{m.text}</td>
                                        <td
                                            ><button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "mondasok",
                                                        m.id,
                                                        fetchMondasok,
                                                    )}>Törlés</button
                                            ></td
                                        >
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="3">Nincsenek idézetek.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Quick Links Tab -->
                {#if activeTab === "quicklinks"}
                    <h3>Új gyorslink hozzáadása</h3>
                    <form class="admin-form" on:submit={submitLink}>
                        <label for="link_title">Cím</label>
                        <input
                            id="link_title"
                            type="text"
                            bind:value={newLink.title}
                            required
                        />

                        <label for="link_url">URL</label>
                        <input
                            id="link_url"
                            type="url"
                            bind:value={newLink.url}
                            required
                        />

                        <label for="link_color">Háttérszín (pl. #e6f0ff)</label>
                        <input
                            id="link_color"
                            type="text"
                            bind:value={newLink.bg_color}
                            placeholder="var(--card-bg)"
                        />

                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>

                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Szín</th>
                                    <th>Cím</th>
                                    <th>Művelet</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each quickLinks as q}
                                    <tr>
                                        <td>
                                            <span
                                                style="display:inline-block; width:20px; height:20px; background:{q.bg_color}; border:1px solid var(--border-color);"
                                            ></span>
                                        </td>
                                        <td
                                            ><a
                                                href={q.url}
                                                target="_blank"
                                                rel="nofollow noopener"
                                                >{q.title}</a
                                            ></td
                                        >
                                        <td
                                            ><button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "quick_links",
                                                        q.id,
                                                        fetchQuickLinks,
                                                    )}>Törlés</button
                                            ></td
                                        >
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="3"
                                            >Nincsenek gyorslinkek.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- News Feeds Tab -->
                {#if activeTab === "newsfeeds"}
                    <h3>Új RSS hírfolyam hozzáadása</h3>
                    <form class="admin-form" on:submit={submitNews}>
                        <label for="news_title">Hírportál neve</label>
                        <input
                            id="news_title"
                            type="text"
                            bind:value={newNews.title}
                            required
                        />

                        <label for="news_url">RSS URL</label>
                        <input
                            id="news_url"
                            type="url"
                            bind:value={newNews.feed_url}
                            required
                        />

                        <label for="news_color">Háttérszín (pl. #ffebd6)</label>
                        <input
                            id="news_color"
                            type="text"
                            bind:value={newNews.bg_color}
                            placeholder="var(--warning-bg)"
                        />

                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>

                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Név</th>
                                    <th>Forrás</th>
                                    <th>Utolsó frissítés</th>
                                    <th>Szín</th>
                                    <th>Művelet</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each newsFeeds as nf}
                                    <tr>
                                        <td>{nf.title}</td>
                                        <td>{nf.feed_url}</td>
                                        <td>
                                            {#if feedTimestamps[nf.feed_url]}
                                                {new Date(
                                                    feedTimestamps[nf.feed_url],
                                                ).toLocaleString("hu-HU")}
                                            {:else}
                                                Soha
                                            {/if}
                                            <div style="margin-top: 5px;">
                                                <button
                                                    type="button"
                                                    class="btn-update"
                                                    disabled={loadingFeeds.has(
                                                        nf.id,
                                                    )}
                                                    on:click={() =>
                                                        updateSingleFeed(nf)}
                                                    >{loadingFeeds.has(nf.id)
                                                        ? "Folyamatban..."
                                                        : "Frissítés"}</button
                                                >
                                            </div>
                                        </td>
                                        <td
                                            ><span
                                                style="display:inline-block; width:20px; height:20px; background:{nf.bg_color}; border:1px solid var(--border-color);"
                                            ></span></td
                                        >
                                        <td
                                            ><button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "news_feeds",
                                                        nf.id,
                                                        fetchNewsFeeds,
                                                    )}>Törlés</button
                                            ></td
                                        >
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="5"
                                            >Nincsenek hírfolyamok.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Locations Tab -->
                {#if activeTab === "locations"}
                    <h3>Új Település</h3>
                    <form class="admin-form" on:submit={submitLocation}>
                        <label for="loc_name">Település neve</label>
                        <input
                            id="loc_name"
                            type="text"
                            bind:value={newLocation.name}
                            required
                        />

                        <label for="loc_county">Megye</label>
                        <input
                            id="loc_county"
                            type="text"
                            bind:value={newLocation.county}
                        />

                        <label for="loc_type">Típus (város, község, falu)</label
                        >
                        <input
                            id="loc_type"
                            type="text"
                            bind:value={newLocation.type}
                        />

                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>

                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Név</th>
                                    <th>Megye</th>
                                    <th>Típus</th>
                                    <th>Művelet</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each locations as l}
                                    <tr>
                                        <td>{l.id}</td>
                                        <td>{l.name}</td>
                                        <td>{l.county}</td>
                                        <td>{l.type}</td>
                                        <td
                                            ><button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "locations",
                                                        l.id,
                                                        fetchLocations,
                                                    )}>Törlés</button
                                            ></td
                                        >
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="5"
                                            >Nincsenek települések.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Service Categories Tab -->
                {#if activeTab === "service_categories"}
                    <h3>Új Kategória</h3>
                    <form class="admin-form" on:submit={submitServiceCategory}>
                        <label for="cat_name">Kategória neve</label>
                        <input
                            id="cat_name"
                            type="text"
                            bind:value={newServiceCategory.name}
                            required
                        />

                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>

                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Név</th>
                                    <th>Művelet</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each serviceCategories as cat}
                                    <tr>
                                        <td>{cat.id}</td>
                                        <td>{cat.name}</td>
                                        <td>
                                            <button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "service_categories",
                                                        cat.id,
                                                        fetchServiceCategories,
                                                    )}>Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="3"
                                            >Nincsenek kategóriák.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Services Tab -->
                {#if activeTab === "services"}
                    <h3>Új Szolgáltatás</h3>
                    <form class="admin-form" on:submit={submitService}>
                        <label for="serv_loc">Település</label>
                        <select
                            id="serv_loc"
                            bind:value={newService.location_id}
                            required
                        >
                            <option value="">Válassz...</option>
                            {#each locations as loc}
                                <option value={loc.id}
                                    >{loc.name} ({loc.county})</option
                                >
                            {/each}
                        </select>

                        <label for="serv_cat">Kategória</label>
                        <select
                            id="serv_cat"
                            bind:value={newService.category_id}
                            required
                        >
                            <option value="">Válassz...</option>
                            {#each serviceCategories as cat}
                                <option value={cat.id}>{cat.name}</option>
                            {/each}
                        </select>

                        <label for="serv_name">Név</label>
                        <input
                            id="serv_name"
                            type="text"
                            bind:value={newService.name}
                            required
                        />

                        <label for="serv_url">URL / Weblap</label>
                        <input
                            id="serv_url"
                            type="url"
                            bind:value={newService.url}
                        />

                        <label for="serv_phone">Telefon</label>
                        <input
                            id="serv_phone"
                            type="text"
                            bind:value={newService.phone}
                        />

                        <label for="serv_addr">Cím</label>
                        <input
                            id="serv_addr"
                            type="text"
                            bind:value={newService.address}
                        />

                        <textarea id="serv_notes" bind:value={newService.notes}
                        ></textarea>

                        <label for="serv_tags">Címkék (#cimke1 #cimke2)</label>
                        <input
                            id="serv_tags"
                            type="text"
                            bind:value={newService.tags}
                            placeholder="#cimke1 #cimke2"
                        />

                        <label
                            for="serv_lang"
                            style="display:flex; align-items:center; gap:0.5rem; font-weight:normal; margin-top:0.5rem;"
                        >
                            <input
                                id="serv_lang"
                                type="checkbox"
                                bind:checked={newService.is_magyar_language}
                                style="width:auto;"
                            />
                            Magyar nyelvű kiszolgálás
                        </label>

                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>

                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Név</th>
                                    <th>URL</th>
                                    <th>Település</th>
                                    <th>Kategória</th>
                                    <th>Címkék</th>
                                    <th>Művelet</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each services as s}
                                    <tr>
                                        <td>{s.name}</td>
                                        <td>{s.url ? s.url : "-"}</td>
                                        <td>{getLocationName(s.location_id)}</td
                                        >
                                        <td>{getCategoryName(s.category_id)}</td
                                        >
                                        <td>
                                            {#if s.tags}
                                                <div class="admin-table-tags">
                                                    {s.tags}
                                                </div>
                                            {:else}
                                                -
                                            {/if}
                                        </td>
                                        <td
                                            ><button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "services",
                                                        s.id,
                                                        fetchServices,
                                                    )}>Törlés</button
                                            ></td
                                        >
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="4"
                                            >Nincsenek szolgáltatások.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}
            </div>
        </main>
    </div>
{/if}

<style>
    @import "../../styles/admin.css";
</style>
