<script>
    import { onMount } from "svelte";

    const getBase = () =>
        import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

    let authenticated = false;
    let password = "";
    let activeTab = "mondasok";

    let mondasok = [];
    let quickLinks = [];
    let newsFeeds = [];
    let locations = [];
    let entries = [];
    let entryCategories = [];
    let entryTypes = [];
    let events = [];

    // Per-feed loading state
    let loadingFeeds = new Set();
    let feedTimestamps = {};

    // Form binding objects
    let newMondas = { text: "" };
    let newLink = { title: "", url: "", bg_color: "#e6f0ff" };
    let newNews = { title: "", feed_url: "", bg_color: "#ffebd6" };
    let newLocation = {
        name: "",
        name_ro: "",
        name_de: "",
        county: "",
        type: "",
        slug: "",
        post_code: "",
        coordinates: "",
        population: "",
        area: "",
        parent_id: null,
    };
    let newEntry = {
        location_id: "",
        category_id: "",
        name: "",
        slug: "",
        url: "",
        phone: "",
        address: "",
        notes: "",
        type: "entry",
        languages: ["HU"],
        tags: "",
    };
    let newEntryCategory = { name: "" };
    let newEntryType = { name: "" };
    let newEvent = {
        location_id: "",
        title: "",
        description: "",
        start_date: "",
        start_time: "",
        end_date: "",
        end_time: "",
        event_type: "cultural",
        organizer: "",
    };

    let newOrganizerModalVisible = false;
    let newOrganizerEntry = {
        location_id: "",
        category_id: "",
        name: "",
        slug: "",
        url: "",
        phone: "",
        address: "",
        notes: "",
        type: "entry",
        languages: ["HU"],
        tags: "",
    };

    // Edit modal state
    let editingEntry = null;
    let editTagsStr = "";
    let editingLocation = null;
    let editingCategory = null;
    let editingType = null;
    let editingMondas = null;
    let editingLink = null;
    let editingNews = null;
    let editingEvent = null;

    const LANGUAGES = ["HU", "RO", "DE", "EN"];
    const COUNTIES = ["Hargita", "Kovászna", "Maros"];
    const LOCATION_TYPES = ["város", "község", "falu", "megye", "municípium"];

    // Custom dialog state
    let dialogVisible = false;
    let dialogMsg = "";
    let dialogType = "alert"; // "alert" or "confirm"
    let dialogResolve = null;

    function showAlert(msg) {
        return new Promise((resolve) => {
            dialogMsg = msg;
            dialogType = "alert";
            dialogResolve = resolve;
            dialogVisible = true;
        });
    }
    function showConfirm(msg) {
        return new Promise((resolve) => {
            dialogMsg = msg;
            dialogType = "confirm";
            dialogResolve = resolve;
            dialogVisible = true;
        });
    }
    function dialogOk() {
        dialogVisible = false;
        if (dialogResolve) dialogResolve(true);
        dialogResolve = null;
    }
    function dialogCancel() {
        dialogVisible = false;
        if (dialogResolve) dialogResolve(false);
        dialogResolve = null;
    }

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
        fetchEntries();
        fetchEntryCategories();
        fetchEntryTypes();
        fetchEvents();
    }

    // generic fetch helper
    async function loadData(endpoint, setter) {
        try {
            const res = await fetch(`${getBase()}/api/admin/${endpoint}`);
            if (res.ok) setter(await res.json());
        } catch (e) {
            console.error(e);
        }
    }

    // --- specific fetches ---
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
    function fetchEntries() {
        loadData("entries", (d) => (entries = d));
    }
    function fetchEntryCategories() {
        loadData("entry_categories", (d) => (entryCategories = d));
    }
    function fetchEntryTypes() {
        loadData("entry_types", (d) => (entryTypes = d));
    }
    function fetchEvents() {
        loadData("events", (d) => (events = d));
    }

    // generic create
    async function createRecord(endpoint, data, reloadFunc, resetFormFunc) {
        try {
            const res = await fetch(`${getBase()}/api/admin/${endpoint}`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(data),
            });
            if (res.ok) {
                reloadFunc();
                resetFormFunc();
            } else {
                showAlert("Hiba: " + (await res.text()));
            }
        } catch (e) {
            console.error(e);
        }
    }

    // generic update (PUT)
    async function updateRecord(endpoint, data, reloadFunc) {
        try {
            const res = await fetch(`${getBase()}/api/admin/${endpoint}`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(data),
            });
            if (res.ok) {
                reloadFunc();
            } else {
                showAlert("Mentési hiba: " + (await res.text()));
            }
        } catch (e) {
            console.error(e);
        }
    }

    // generic delete
    async function deleteRecord(endpoint, id, reloadFunc) {
        const ok = await showConfirm("Biztosan törölni szeretnéd?");
        if (!ok) return;
        try {
            const res = await fetch(
                `${getBase()}/api/admin/${endpoint}?id=${id}`,
                { method: "DELETE" },
            );
            if (res.ok) reloadFunc();
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
                    bg_color: "#ffebd6",
                }),
        );
    }

    async function updateSingleFeed(feed) {
        loadingFeeds.add(feed.id);
        loadingFeeds = new Set(loadingFeeds);

        try {
            const proxiedUrl =
                `${getBase()}/api/proxy?url=` +
                encodeURIComponent(feed.feed_url);
            const res = await fetch(proxiedUrl);
            if (res.ok) {
                feedTimestamps[feed.feed_url] = Date.now();
                localStorage.setItem(
                    "news_feed_timestamps",
                    JSON.stringify(feedTimestamps),
                );
                feedTimestamps = { ...feedTimestamps };
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
            () =>
                (newLocation = {
                    name: "",
                    name_ro: "",
                    name_de: "",
                    county: "",
                    type: "",
                    post_code: "",
                    coordinates: "",
                    population: "",
                    area: "",
                    crest: "",
                    parent_id: null,
                }),
        );
    }
    function submitEvent(e) {
        e.preventDefault();
        createRecord(
            "events",
            { ...newEvent, location_id: parseInt(newEvent.location_id) || 0 },
            fetchEvents,
            () =>
                (newEvent = {
                    location_id: "",
                    title: "",
                    description: "",
                    start_date: "",
                    start_time: "",
                    end_date: "",
                    end_time: "",
                    event_type: "cultural",
                    organizer: "",
                }),
        );
    }

    function submitNewOrganizer(e) {
        e.preventDefault();
        const payload = {
            ...newOrganizerEntry,
            location_id: parseInt(newOrganizerEntry.location_id) || 0,
            category_id: newOrganizerEntry.category_id
                ? parseInt(newOrganizerEntry.category_id)
                : null,
            tags: tagsFromStr(newOrganizerEntry.tags),
        };
        createRecord("entries", payload, fetchEntries, () => {
            newEvent.organizer = newOrganizerEntry.name;
            newOrganizerModalVisible = false;
            newOrganizerEntry = {
                location_id: "",
                category_id: "",
                name: "",
                slug: "",
                url: "",
                phone: "",
                address: "",
                notes: "",
                type: "entry",
                languages: ["HU"],
                tags: "",
            };
        });
    }

    // --- Location edit helpers ---
    function startEditLocation(loc) {
        editingLocation = { ...loc };
    }
    function cancelEditLocation() {
        editingLocation = null;
    }
    async function saveEditLocation() {
        if (!editingLocation) return;
        await updateRecord("locations", editingLocation, fetchLocations);
        cancelEditLocation();
    }

    // --- Submit entry category ---
    function submitEntryCategory(e) {
        e.preventDefault();
        createRecord(
            "entry_categories",
            newEntryCategory,
            fetchEntryCategories,
            () => (newEntryCategory = { name: "" }),
        );
    }

    // --- Submit entry type ---
    function submitEntryType(e) {
        e.preventDefault();
        createRecord(
            "entry_types",
            newEntryType,
            fetchEntryTypes,
            () => (newEntryType = { name: "" }),
        );
    }

    // --- Inline edit helpers for categories ---
    async function startEditCategory(cat) {
        const ok = await showConfirm("Biztosan szerkeszteni szeretné?");
        if (!ok) return;
        editingCategory = { ...cat };
    }
    function cancelEditCategory() {
        editingCategory = null;
    }
    async function saveEditCategory() {
        if (!editingCategory) return;
        const ok = await showConfirm("Biztosan menteni szeretné a módosítást?");
        if (!ok) return;
        await updateRecord(
            "entry_categories",
            editingCategory,
            fetchEntryCategories,
        );
        editingCategory = null;
    }

    // --- Inline edit helpers for types ---
    async function startEditType(et) {
        const ok = await showConfirm("Biztosan szerkeszteni szeretné?");
        if (!ok) return;
        editingType = { ...et };
    }
    function cancelEditType() {
        editingType = null;
    }
    async function saveEditType() {
        if (!editingType) return;
        const ok = await showConfirm("Biztosan menteni szeretné a módosítást?");
        if (!ok) return;
        await updateRecord("entry_types", editingType, fetchEntryTypes);
        editingType = null;
    }

    // --- Inline edit helpers for mondasok ---
    async function startEditMondas(m) {
        const ok = await showConfirm("Biztosan szerkeszteni szeretné?");
        if (!ok) return;
        editingMondas = { ...m };
    }
    function cancelEditMondas() {
        editingMondas = null;
    }
    async function saveEditMondas() {
        if (!editingMondas) return;
        const ok = await showConfirm("Biztosan menteni szeretné a módosítást?");
        if (!ok) return;
        await updateRecord("mondasok", editingMondas, fetchMondasok);
        editingMondas = null;
    }

    // --- Inline edit helpers for quick links ---
    async function startEditLink(ql) {
        const ok = await showConfirm("Biztosan szerkeszteni szeretné?");
        if (!ok) return;
        editingLink = { ...ql };
    }
    function cancelEditLink() {
        editingLink = null;
    }
    async function saveEditLink() {
        if (!editingLink) return;
        const ok = await showConfirm("Biztosan menteni szeretné a módosítást?");
        if (!ok) return;
        await updateRecord("quick_links", editingLink, fetchQuickLinks);
        editingLink = null;
    }

    // --- Inline edit helpers for news feeds ---
    async function startEditNews(nf) {
        const ok = await showConfirm("Biztosan szerkeszteni szeretné?");
        if (!ok) return;
        editingNews = { ...nf };
    }
    function cancelEditNews() {
        editingNews = null;
    }
    async function saveEditNews() {
        if (!editingNews) return;
        const ok = await showConfirm("Biztosan menteni szeretné a módosítást?");
        if (!ok) return;
        await updateRecord("news_feeds", editingNews, fetchNewsFeeds);
        editingNews = null;
    }
    async function startEditEvent(ev) {
        const ok = await showConfirm("Biztosan szerkeszteni szeretné?");
        if (!ok) return;
        editingEvent = { ...ev };
    }
    function cancelEditEvent() {
        editingEvent = null;
    }
    async function saveEditEvent() {
        if (!editingEvent) return;
        const ok = await showConfirm("Biztosan menteni szeretné a módosítást?");
        if (!ok) return;
        await updateRecord("events", editingEvent, fetchEvents);
        editingEvent = null;
    }
    // --- Tag helpers ---
    function tagsFromStr(str) {
        return (str || "")
            .split(/[\s,]+/)
            .map((t) => t.replace(/^#/, "").trim())
            .filter(Boolean);
    }
    function tagsToStr(arr) {
        return (arr || []).map((t) => "#" + t).join(" ");
    }
    function getLocationName(id) {
        const l = locations.find((loc) => loc.id === id);
        return l ? `${l.name}${l.county ? " (" + l.county + ")" : ""}` : id;
    }
    function getCategoryName(id) {
        const c = entryCategories.find((cat) => cat.id === id);
        return c ? c.name : "-";
    }

    // --- Submit entry (Create) ---
    function submitEntry(e) {
        e.preventDefault();
        const payload = {
            ...newEntry,
            location_id: parseInt(newEntry.location_id) || 0,
            category_id: newEntry.category_id
                ? parseInt(newEntry.category_id)
                : null,
            tags: tagsFromStr(newEntry.tags),
        };
        createRecord(
            "entries",
            payload,
            fetchEntries,
            () =>
                (newEntry = {
                    location_id: "",
                    category_id: null,
                    name: "",
                    url: "",
                    phone: "",
                    address: "",
                    notes: "",
                    type: "entry",
                    languages: ["HU"],
                    tags: "",
                }),
        );
    }

    // --- Edit modal ---
    async function openEdit(entry) {
        const ok = await showConfirm("Biztosan szerkeszteni szeretné?");
        if (!ok) return;
        editingEntry = {
            ...entry,
            languages: entry.languages ? [...entry.languages] : ["HU"],
        };
        editTagsStr = tagsToStr(entry.tags);
    }
    function closeEdit() {
        editingEntry = null;
        editTagsStr = "";
    }
    async function saveEdit() {
        if (!editingEntry) return;
        const ok = await showConfirm("Biztosan menteni szeretné a módosítást?");
        if (!ok) return;
        const payload = {
            ...editingEntry,
            location_id: parseInt(editingEntry.location_id) || 0,
            category_id: editingEntry.category_id
                ? parseInt(editingEntry.category_id)
                : null,
            tags: tagsFromStr(editTagsStr),
        };
        await updateRecord("entries", payload, fetchEntries);
        closeEdit();
    }
</script>

<svelte:head>
    <title>Székely Gugel - Adminisztráció</title>
</svelte:head>

{#if !authenticated}
    <div class="container">
        <div class="admin-login-wrapper">
            <div class="admin-container login-box">
                <h2>Adminisztráció Belépés</h2>
                <form class="admin-form mt-lg" on:submit={login}>
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
                class="admin-sidebar-btn {activeTab === 'events'
                    ? 'active'
                    : ''}"
                on:click={() => (activeTab = "events")}
                title="Események"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
                    ><rect x="3" y="4" width="18" height="18" rx="2" ry="2"
                    ></rect><line x1="16" y1="2" x2="16" y2="6"></line><line
                        x1="8"
                        y1="2"
                        x2="8"
                        y2="6"
                    ></line><line x1="3" y1="10" x2="21" y2="10"></line></svg
                >
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'entry_categories'
                    ? 'active'
                    : ''}"
                on:click={() => (activeTab = "entry_categories")}
                title="Bejegyzés Kategóriák"
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
                class="admin-sidebar-btn {activeTab === 'entry_types'
                    ? 'active'
                    : ''}"
                on:click={() => (activeTab = "entry_types")}
                title="Bejegyzés típusok"
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
                    <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
                    <polyline points="2 17 12 22 22 17"></polyline>
                    <polyline points="2 12 12 17 22 12"></polyline>
                </svg>
            </button>
            <button
                class="admin-sidebar-btn {activeTab === 'entries'
                    ? 'active'
                    : ''}"
                on:click={() => (activeTab = "entries")}
                title="Index"
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
                    {#if activeTab === "entry_categories"}Bejegyzés Kategóriák
                        Kezelése{/if}
                    {#if activeTab === "entries"}Bejegyzések Kezelése{/if}
                    {#if activeTab === "entry_types"}Bejegyzés Típusok Kezelése{/if}
                </h2>
                <button class="btn-logout" on:click={logout}
                    >Kijelentkezés</button
                >
            </div>

            <div class="admin-container w-full">
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
                                    <th>Szerk.</th>
                                    <th>Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each mondasok as m}
                                    <tr>
                                        <td>{m.id}</td>
                                        <td>{m.text}</td>
                                        <td>
                                            <button
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditMondas(m)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "mondasok",
                                                        m.id,
                                                        fetchMondasok,
                                                    )}>Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="4">Nincsenek idézetek.</td
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
                            placeholder="#e6f0ff"
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
                                    <th>Szerk.</th>
                                    <th>Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each quickLinks as q}
                                    <tr>
                                        <td>
                                            <span
                                                class="color-swatch"
                                                style:background={q.bg_color}
                                            ></span>
                                        </td>
                                        <td>
                                            <a
                                                href={q.url}
                                                target="_blank"
                                                rel="nofollow noopener"
                                                >{q.title}</a
                                            >
                                        </td>
                                        <td>
                                            <button
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditLink(q)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "quick_links",
                                                        q.id,
                                                        fetchQuickLinks,
                                                    )}>Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="4"
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
                            placeholder="#ffebd6"
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
                                    <th>Szerk.</th>
                                    <th>Törlés</th>
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
                                            <div class="mt-xs">
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
                                        <td>
                                            <span
                                                class="color-swatch"
                                                style:background={nf.bg_color}
                                            ></span>
                                        </td>
                                        <td>
                                            <button
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditNews(nf)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "news_feeds",
                                                        nf.id,
                                                        fetchNewsFeeds,
                                                    )}>Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="6"
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
                        <label for="loc_name">Település neve (HU)</label>
                        <input
                            id="loc_name"
                            type="text"
                            bind:value={newLocation.name}
                            required
                        />

                        <label for="loc_name_ro">Név (RO)</label>
                        <input
                            id="loc_name_ro"
                            type="text"
                            bind:value={newLocation.name_ro}
                            placeholder="opcionális"
                        />

                        <label for="loc_name_de">Név (DE)</label>
                        <input
                            id="loc_name_de"
                            type="text"
                            bind:value={newLocation.name_de}
                            placeholder="opcionális"
                        />

                        <label for="loc_county">Megye</label>
                        <select id="loc_county" bind:value={newLocation.county}>
                            <option value="">Válassz...</option>
                            {#each COUNTIES as c}<option value={c}>{c}</option
                                >{/each}
                        </select>

                        <label for="loc_type">Típus</label>
                        <select id="loc_type" bind:value={newLocation.type}>
                            <option value="">Válassz...</option>
                            {#each LOCATION_TYPES as t}<option value={t}
                                    >{t}</option
                                >{/each}
                        </select>

                        <label for="loc_post_code">Posta kód</label>
                        <input
                            id="loc_post_code"
                            type="text"
                            bind:value={newLocation.post_code}
                        />

                        <label for="loc_coords">Koordináták</label>
                        <input
                            id="loc_coords"
                            type="text"
                            bind:value={newLocation.coordinates}
                        />

                        <label for="loc_pop">Lakosság (fő)</label>
                        <input
                            id="loc_pop"
                            type="text"
                            bind:value={newLocation.population}
                        />

                        <label for="loc_area">Terület (km²)</label>
                        <input
                            id="loc_area"
                            type="text"
                            bind:value={newLocation.area}
                        />

                        <label for="loc_crest">Címer URL</label>
                        <input
                            id="loc_crest"
                            type="text"
                            bind:value={newLocation.crest}
                        />

                        <label for="loc_parent">Kapcsolt település</label>
                        <select
                            id="loc_parent"
                            bind:value={newLocation.parent_id}
                        >
                            <option value={null}
                                >Nincs (Önálló város/község)</option
                            >
                            {#each locations as loc}
                                <option value={loc.id}
                                    >{loc.name} ({loc.county})</option
                                >
                            {/each}
                        </select>

                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>

                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Név (HU)</th>
                                    <th>Név (RO)</th>
                                    <th>Név (DE)</th>
                                    <th>Megye</th>
                                    <th>Típus</th>
                                    <th title="Posta kód">Irányítószám</th>
                                    <th>Koordináták</th>
                                    <th>Lakosság (fő)</th>
                                    <th>Terület (km²)</th>
                                    <th>Címer</th>
                                    <th>Szülő település</th>
                                    <th>Szerk.</th>
                                    <th>Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each locations as l}
                                    <tr>
                                        <td>{l.id}</td>
                                        <td>{l.name}</td>
                                        <td>{l.name_ro || "-"}</td>
                                        <td>{l.name_de || "-"}</td>
                                        <td>{l.county || "-"}</td>
                                        <td>
                                            <span class="badge"
                                                >{l.type || "-"}</span
                                            >
                                        </td>
                                        <td>{l.post_code || "-"}</td>
                                        <td>{l.coordinates || "-"}</td>
                                        <td>{l.population || "-"}</td>
                                        <td>{l.area || "-"}</td>
                                        <td>{l.crest ? "Van" : "-"}</td>
                                        <td>
                                            {#if l.parent_id}
                                                {getLocationName(l.parent_id)}
                                            {:else}
                                                -
                                            {/if}
                                        </td>
                                        <td>
                                            <button
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditLocation(l)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "locations",
                                                        l.id,
                                                        fetchLocations,
                                                    )}>Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="14"
                                            >Nincsenek települések.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Events Tab -->
                {#if activeTab === "events"}
                    <h3>Új Esemény</h3>
                    <form class="admin-form" on:submit={submitEvent}>
                        <label for="event_loc">Település / Helyszín</label>
                        <select
                            id="event_loc"
                            bind:value={newEvent.location_id}
                            required
                        >
                            <option value="">Válassz...</option>
                            {#each locations as loc}
                                <option value={loc.id}
                                    >{loc.name} ({loc.county})</option
                                >
                            {/each}
                        </select>

                        <label for="event_title">Esemény neve</label>
                        <input
                            id="event_title"
                            type="text"
                            bind:value={newEvent.title}
                            required
                        />

                        <label for="event_desc">Leírás</label>
                        <textarea
                            id="event_desc"
                            bind:value={newEvent.description}
                        ></textarea>

                        <div class="flex gap-lg">
                            <div class="flex-1">
                                <label for="event_start_date">Kezdő dátum</label
                                >
                                <input
                                    id="event_start_date"
                                    type="date"
                                    bind:value={newEvent.start_date}
                                    required
                                />
                            </div>
                            <div class="flex-1">
                                <label for="event_start_time"
                                    >Kezdő időpont</label
                                >
                                <input
                                    id="event_start_time"
                                    type="time"
                                    bind:value={newEvent.start_time}
                                />
                            </div>
                        </div>

                        <div class="flex gap-lg">
                            <div class="flex-1">
                                <label for="event_end_date"
                                    >Befejező dátum</label
                                >
                                <input
                                    id="event_end_date"
                                    type="date"
                                    bind:value={newEvent.end_date}
                                    required
                                />
                            </div>
                            <div class="flex-1">
                                <label for="event_end_time"
                                    >Befejező időpont</label
                                >
                                <input
                                    id="event_end_time"
                                    type="time"
                                    bind:value={newEvent.end_time}
                                />
                            </div>
                        </div>

                        <label for="event_type">Típus</label>
                        <select
                            id="event_type"
                            bind:value={newEvent.event_type}
                        >
                            <option value="cultural">Kulturális</option>
                            <option value="sports">Sport</option>
                            <option value="festival">Fesztivál</option>
                            <option value="religious">Vallási</option>
                            <option value="other">Egyéb</option>
                        </select>

                        <label for="event_org">Szervező</label>
                        <div class="flex gap-sm items-center">
                            <input
                                id="event_org"
                                type="text"
                                bind:value={newEvent.organizer}
                                list="organizers_list"
                                autocomplete="off"
                                class="flex-1"
                            />
                            <datalist id="organizers_list">
                                {#each entries as e}
                                    <option value={e.name}></option>
                                {/each}
                            </datalist>
                            <button
                                type="button"
                                class="btn-update"
                                style="margin-bottom:0"
                                on:click={() =>
                                    (newOrganizerModalVisible = true)}
                            >
                                Új szervező
                            </button>
                        </div>

                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>

                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Cím</th>
                                    <th>Dátum és Idő</th>
                                    <th>Helyszín</th>
                                    <th>Szervező</th>
                                    <th>Szerk.</th>
                                    <th>Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each events as e}
                                    <tr>
                                        <td>{e.title}</td>
                                        <td>
                                            {new Date(
                                                e.start_date,
                                            ).toLocaleDateString("hu-HU")}
                                            {#if e.start_time}
                                                - {e.start_time.slice(
                                                    0,
                                                    5,
                                                )}{/if}
                                        </td>
                                        <td>{getLocationName(e.location_id)}</td
                                        >
                                        <td>{e.organizer || "-"}</td>
                                        <td>
                                            <button
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditEvent(e)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "events",
                                                        e.id,
                                                        fetchEvents,
                                                    )}>Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="6"
                                            >Nincsenek események.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Entry Categories Tab -->
                {#if activeTab === "entry_categories"}
                    <h3>Új Kategória</h3>
                    <form class="admin-form" on:submit={submitEntryCategory}>
                        <label for="cat_name">Kategória neve</label>
                        <input
                            id="cat_name"
                            type="text"
                            bind:value={newEntryCategory.name}
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
                                    <th>Szerk.</th>
                                    <th>Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each entryCategories as cat}
                                    <tr>
                                        <td>{cat.id}</td>
                                        <td>{cat.name}</td>
                                        <td>
                                            <button
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditCategory(cat)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "entry_categories",
                                                        cat.id,
                                                        fetchEntryCategories,
                                                    )}>Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="4"
                                            >Nincsenek kategóriák.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Entries Tab -->
                {#if activeTab === "entries"}
                    <h3>Új bejegyzés</h3>
                    <form class="admin-form" on:submit={submitEntry}>
                        <label for="serv_type">Típus</label>
                        <select id="serv_type" bind:value={newEntry.type}>
                            {#each entryTypes as t}<option value={t.name}
                                    >{t.name}</option
                                >{/each}
                        </select>

                        <label for="serv_loc">Település</label>
                        <select
                            id="serv_loc"
                            bind:value={newEntry.location_id}
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
                        <select id="serv_cat" bind:value={newEntry.category_id}>
                            <option value={null}>-</option>
                            {#each entryCategories as cat}
                                <option value={cat.id}>{cat.name}</option>
                            {/each}
                        </select>

                        <label for="serv_name">Név</label>
                        <input
                            id="serv_name"
                            type="text"
                            bind:value={newEntry.name}
                            required
                        />

                        <label for="serv_url">URL / Weblap</label>
                        <input
                            id="serv_url"
                            type="url"
                            bind:value={newEntry.url}
                        />

                        <label for="serv_phone">Telefon</label>
                        <input
                            id="serv_phone"
                            type="text"
                            bind:value={newEntry.phone}
                        />

                        <label for="serv_addr">Cím</label>
                        <input
                            id="serv_addr"
                            type="text"
                            bind:value={newEntry.address}
                        />

                        <label for="serv_notes">Megjegyzések</label>
                        <textarea id="serv_notes" bind:value={newEntry.notes}
                        ></textarea>

                        <label for="serv_tags">Címkék (#cimke1 #cimke2)</label>
                        <input
                            id="serv_tags"
                            type="text"
                            bind:value={newEntry.tags}
                            placeholder="#cimke1 #cimke2"
                        />

                        <span class="form-group-label">Nyelvek</span>
                        <div class="flex gap-lg flex-wrap mb-lg">
                            {#each LANGUAGES as lang}
                                <label
                                    class="flex items-center gap-xs font-normal"
                                >
                                    <input
                                        type="checkbox"
                                        checked={newEntry.languages.includes(
                                            lang,
                                        )}
                                        on:change={() =>
                                            (newEntry.languages =
                                                newEntry.languages.includes(
                                                    lang,
                                                )
                                                    ? newEntry.languages.filter(
                                                          (l) => l !== lang,
                                                      )
                                                    : [
                                                          ...newEntry.languages,
                                                          lang,
                                                      ])}
                                        class="w-auto"
                                    />
                                    {lang}
                                </label>
                            {/each}
                        </div>

                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>

                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Név</th>
                                    <th>Típus</th>
                                    <th>URL</th>
                                    <th>Település</th>
                                    <th>Kategória</th>
                                    <th>Nyelvek</th>
                                    <th>Címkék</th>
                                    <th>Szerk.</th>
                                    <th>Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each entries as s}
                                    <tr>
                                        <td>{s.name}</td>
                                        <td
                                            ><span class="badge"
                                                >{s.type || "entry"}</span
                                            ></td
                                        >
                                        <td>{s.url ? s.url : "-"}</td>
                                        <td>{getLocationName(s.location_id)}</td
                                        >
                                        <td>{getCategoryName(s.category_id)}</td
                                        >
                                        <td>{(s.languages || []).join(", ")}</td
                                        >
                                        <td>
                                            {#if s.tags && s.tags.length > 0}
                                                <div class="admin-table-tags">
                                                    {s.tags
                                                        .map((t) => "#" + t)
                                                        .join(" ")}
                                                </div>
                                            {:else}-{/if}
                                        </td>
                                        <td>
                                            <button
                                                class="btn-update"
                                                on:click={() => openEdit(s)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "entries",
                                                        s.id,
                                                        fetchEntries,
                                                    )}>Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="9"
                                            >Nincsenek bejegyzések.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Entry Types Tab -->
                {#if activeTab === "entry_types"}
                    <h3>Új Típus</h3>
                    <form class="admin-form" on:submit={submitEntryType}>
                        <label for="etype_name">Típus neve</label>
                        <input
                            id="etype_name"
                            type="text"
                            bind:value={newEntryType.name}
                            required
                            placeholder="pl. entry, business..."
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
                                    <th>Szerk.</th>
                                    <th>Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each entryTypes as et}
                                    <tr>
                                        <td>{et.id}</td>
                                        <td>{et.name}</td>
                                        <td>
                                            <button
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditType(et)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteRecord(
                                                        "entry_types",
                                                        et.id,
                                                        fetchEntryTypes,
                                                    )}>Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="4">Nincsenek típusok.</td
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

    <!-- Edit Mondas Modal -->
    {#if editingMondas}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={cancelEditMondas}
            on:keydown={(e) => e.key === "Escape" && cancelEditMondas()}
        >
            <div class="admin-modal">
                <h3>Mondás szerkesztése</h3>
                <form
                    class="admin-form"
                    on:submit|preventDefault={saveEditMondas}
                >
                    <label for="emondas_text">Mondás szövege</label>
                    <textarea
                        id="emondas_text"
                        bind:value={editingMondas.text}
                        required
                        rows="4"
                        class="w-full"
                    ></textarea>

                    <div class="modal-actions">
                        <button type="submit" class="admin-submit-btn"
                            >Mentés</button
                        >
                        <button
                            type="button"
                            class="btn-delete"
                            on:click={cancelEditMondas}>Mégse</button
                        >
                    </div>
                </form>
            </div>
        </div>
    {/if}

    <!-- Edit QuickLink Modal -->
    {#if editingLink}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={cancelEditLink}
            on:keydown={(e) => e.key === "Escape" && cancelEditLink()}
        >
            <div class="admin-modal">
                <h3>Gyorslink szerkesztése</h3>
                <form
                    class="admin-form"
                    on:submit|preventDefault={saveEditLink}
                >
                    <label for="elink_title">Cím</label>
                    <input
                        id="elink_title"
                        type="text"
                        bind:value={editingLink.title}
                        required
                    />

                    <label for="elink_url">URL</label>
                    <input
                        id="elink_url"
                        type="url"
                        bind:value={editingLink.url}
                        required
                    />

                    <label for="elink_color">Háttérszín</label>
                    <input
                        id="elink_color"
                        type="text"
                        bind:value={editingLink.bg_color}
                    />

                    <div class="modal-actions">
                        <button type="submit" class="admin-submit-btn"
                            >Mentés</button
                        >
                        <button
                            type="button"
                            class="btn-delete"
                            on:click={cancelEditLink}>Mégse</button
                        >
                    </div>
                </form>
            </div>
        </div>
    {/if}

    <!-- Edit News Modal -->
    {#if editingNews}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={cancelEditNews}
            on:keydown={(e) => e.key === "Escape" && cancelEditNews()}
        >
            <div class="admin-modal">
                <h3>Hírfolyam szerkesztése</h3>
                <form
                    class="admin-form"
                    on:submit|preventDefault={saveEditNews}
                >
                    <label for="enews_title">Hírportál neve</label>
                    <input
                        id="enews_title"
                        type="text"
                        bind:value={editingNews.title}
                        required
                    />

                    <label for="enews_url">RSS URL</label>
                    <input
                        id="enews_url"
                        type="url"
                        bind:value={editingNews.feed_url}
                        required
                    />

                    <label for="enews_color">Háttérszín</label>
                    <input
                        id="enews_color"
                        type="text"
                        bind:value={editingNews.bg_color}
                    />

                    <div class="modal-actions">
                        <button type="submit" class="admin-submit-btn"
                            >Mentés</button
                        >
                        <button
                            type="button"
                            class="btn-delete"
                            on:click={cancelEditNews}>Mégse</button
                        >
                    </div>
                </form>
            </div>
        </div>
    {/if}

    <!-- Edit Entry Modal -->
    {#if editingEntry}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={closeEdit}
            on:keydown={(e) => e.key === "Escape" && closeEdit()}
        >
            <div class="admin-modal">
                <h3>Bejegyzés szerkesztése</h3>
                <form class="admin-form" on:submit|preventDefault={saveEdit}>
                    <label for="edit_type">Típus</label>
                    <select id="edit_type" bind:value={editingEntry.type}>
                        {#each entryTypes as t}<option value={t.name}
                                >{t.name}</option
                            >{/each}
                    </select>

                    <label for="edit_loc">Település</label>
                    <select
                        id="edit_loc"
                        bind:value={editingEntry.location_id}
                        required
                    >
                        {#each locations as loc}
                            <option value={loc.id}
                                >{loc.name} ({loc.county})</option
                            >
                        {/each}
                    </select>

                    <label for="edit_cat">Kategória</label>
                    <select id="edit_cat" bind:value={editingEntry.category_id}>
                        <option value={null}>-</option>
                        {#each entryCategories as cat}
                            <option value={cat.id}>{cat.name}</option>
                        {/each}
                    </select>

                    <label for="edit_name">Név</label>
                    <input
                        id="edit_name"
                        type="text"
                        bind:value={editingEntry.name}
                        required
                    />

                    <label for="edit_slug">Slug (URL azonosító)</label>
                    <input
                        id="edit_slug"
                        type="text"
                        bind:value={editingEntry.slug}
                    />

                    <label for="edit_url">URL / Weblap</label>
                    <input
                        id="edit_url"
                        type="url"
                        bind:value={editingEntry.url}
                    />

                    <label for="edit_phone">Telefon</label>
                    <input
                        id="edit_phone"
                        type="text"
                        bind:value={editingEntry.phone}
                    />

                    <label for="edit_addr">Cím</label>
                    <input
                        id="edit_addr"
                        type="text"
                        bind:value={editingEntry.address}
                    />

                    <label for="edit_notes">Megjegyzések</label>
                    <textarea id="edit_notes" bind:value={editingEntry.notes}
                    ></textarea>

                    <label for="edit_tags">Címkék (#cimke1 #cimke2)</label>
                    <input
                        id="edit_tags"
                        type="text"
                        bind:value={editTagsStr}
                        placeholder="#cimke1 #cimke2"
                    />

                    <span class="form-group-label">Nyelvek</span>
                    <div class="flex gap-lg flex-wrap mb-lg">
                        {#each LANGUAGES as lang}
                            <label class="flex items-center gap-xs font-normal">
                                <input
                                    type="checkbox"
                                    checked={editingEntry.languages.includes(
                                        lang,
                                    )}
                                    on:change={() =>
                                        (editingEntry.languages =
                                            editingEntry.languages.includes(
                                                lang,
                                            )
                                                ? editingEntry.languages.filter(
                                                      (l) => l !== lang,
                                                  )
                                                : [
                                                      ...editingEntry.languages,
                                                      lang,
                                                  ])}
                                    class="w-auto"
                                />
                                {lang}
                            </label>
                        {/each}
                    </div>

                    <div class="modal-actions">
                        <button type="submit" class="admin-submit-btn"
                            >Mentés</button
                        >
                        <button
                            type="button"
                            class="btn-delete"
                            on:click={closeEdit}>Mégse</button
                        >
                    </div>
                </form>
            </div>
        </div>
    {/if}

    <!-- Edit Location Modal -->
    {#if editingLocation}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={cancelEditLocation}
            on:keydown={(e) => e.key === "Escape" && cancelEditLocation()}
        >
            <div class="admin-modal">
                <h3>Település szerkesztése</h3>
                <form
                    class="admin-form"
                    on:submit|preventDefault={saveEditLocation}
                >
                    <label for="eloc_name">Név (HU)</label>
                    <input
                        id="eloc_name"
                        type="text"
                        bind:value={editingLocation.name}
                        required
                    />

                    <label for="eloc_name_ro">Név (RO)</label>
                    <input
                        id="eloc_name_ro"
                        type="text"
                        bind:value={editingLocation.name_ro}
                    />

                    <label for="eloc_name_de">Név (DE)</label>
                    <input
                        id="eloc_name_de"
                        type="text"
                        bind:value={editingLocation.name_de}
                    />

                    <label for="eloc_county">Megye</label>
                    <select
                        id="eloc_county"
                        bind:value={editingLocation.county}
                    >
                        <option value="">-</option>
                        {#each COUNTIES as c}<option value={c}>{c}</option
                            >{/each}
                    </select>

                    <label for="eloc_type">Típus</label>
                    <select id="eloc_type" bind:value={editingLocation.type}>
                        <option value="">-</option>
                        {#each LOCATION_TYPES as t}<option value={t}>{t}</option
                            >{/each}
                    </select>

                    <label for="eloc_post_code" title="Posta kód"
                        >Irányítószám</label
                    >
                    <input
                        id="eloc_post_code"
                        type="text"
                        bind:value={editingLocation.post_code}
                    />

                    <label for="eloc_coords">Koordináták</label>
                    <input
                        id="eloc_coords"
                        type="text"
                        bind:value={editingLocation.coordinates}
                    />

                    <label for="eloc_pop">Lakosság (fő)</label>
                    <input
                        id="eloc_pop"
                        type="text"
                        bind:value={editingLocation.population}
                    />

                    <label for="eloc_area">Terület (km²)</label>
                    <input
                        id="eloc_area"
                        type="text"
                        bind:value={editingLocation.area}
                    />

                    <label for="eloc_crest">Címer URL</label>
                    <input
                        id="eloc_crest"
                        type="text"
                        bind:value={editingLocation.crest}
                    />

                    <label for="eloc_parent">Kapcsolódó település</label>
                    <select
                        id="eloc_parent"
                        bind:value={editingLocation.parent_id}
                    >
                        <option value={null}>Nincs (Önálló város/község)</option
                        >
                        {#each locations.filter((l) => l.id !== editingLocation.id) as loc}
                            <option value={loc.id}
                                >{loc.name} ({loc.county})</option
                            >
                        {/each}
                    </select>

                    <div class="modal-actions">
                        <button type="submit" class="admin-submit-btn"
                            >Mentés</button
                        >
                        <button
                            type="button"
                            class="btn-delete"
                            on:click={cancelEditLocation}>Mégse</button
                        >
                    </div>
                </form>
            </div>
        </div>
    {/if}

    <!-- Edit Category Modal -->
    {#if editingCategory}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={cancelEditCategory}
            on:keydown={(e) => e.key === "Escape" && cancelEditCategory()}
        >
            <div class="admin-modal">
                <h3>Kategória szerkesztése</h3>
                <form
                    class="admin-form"
                    on:submit|preventDefault={saveEditCategory}
                >
                    <label for="ecat_name">Kategória neve</label>
                    <input
                        id="ecat_name"
                        type="text"
                        bind:value={editingCategory.name}
                        required
                    />

                    <div class="modal-actions">
                        <button type="submit" class="admin-submit-btn"
                            >Mentés</button
                        >
                        <button
                            type="button"
                            class="btn-delete"
                            on:click={cancelEditCategory}>Mégse</button
                        >
                    </div>
                </form>
            </div>
        </div>
    {/if}

    <!-- Edit Type Modal -->
    {#if editingType}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={cancelEditType}
            on:keydown={(e) => e.key === "Escape" && cancelEditType()}
        >
            <div class="admin-modal">
                <h3>Típus szerkesztése</h3>
                <form
                    class="admin-form"
                    on:submit|preventDefault={saveEditType}
                >
                    <label for="etype_name_edit">Típus neve</label>
                    <input
                        id="etype_name_edit"
                        type="text"
                        bind:value={editingType.name}
                        required
                    />

                    <div class="modal-actions">
                        <button type="submit" class="admin-submit-btn"
                            >Mentés</button
                        >
                        <button
                            type="button"
                            class="btn-delete"
                            on:click={cancelEditType}>Mégse</button
                        >
                    </div>
                </form>
            </div>
        </div>
    {/if}

    <!-- Custom Dialog -->
    {#if dialogVisible}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-dialog-overlay"
            role="alertdialog"
            tabindex="-1"
            on:click|self={dialogCancel}
        >
            <div class="admin-dialog">
                <p>{dialogMsg}</p>
                <div class="admin-dialog-actions">
                    {#if dialogType === "confirm"}
                        <button class="btn-delete" on:click={dialogCancel}
                            >Mégse</button
                        >
                    {/if}
                    <button class="admin-submit-btn" on:click={dialogOk}
                        >OK</button
                    >
                </div>
            </div>
        </div>
    {/if}

    <!-- Edit Event Modal -->
    {#if editingEvent}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={cancelEditEvent}
            on:keydown={(e) => e.key === "Escape" && cancelEditEvent()}
        >
            <div class="admin-modal">
                <h3>Esemény szerkesztése</h3>
                <form
                    class="admin-form"
                    on:submit|preventDefault={saveEditEvent}
                >
                    <label for="edit_ev_loc">Helyszín</label>
                    <select
                        id="edit_ev_loc"
                        bind:value={editingEvent.location_id}
                        required
                    >
                        {#each locations as loc}
                            <option value={loc.id}
                                >{loc.name} ({loc.county})</option
                            >
                        {/each}
                    </select>

                    <label for="edit_ev_title">Cím</label>
                    <input
                        id="edit_ev_title"
                        type="text"
                        bind:value={editingEvent.title}
                        required
                    />

                    <label for="edit_ev_desc">Leírás</label>
                    <textarea
                        id="edit_ev_desc"
                        bind:value={editingEvent.description}
                    ></textarea>

                    <div class="flex gap-lg">
                        <div class="flex-1">
                            <label for="edit_ev_start_date">Kezdő dátum</label>
                            <input
                                id="edit_ev_start_date"
                                type="date"
                                value={editingEvent.start_date
                                    ? editingEvent.start_date.split("T")[0]
                                    : ""}
                                on:change={(e) =>
                                    (editingEvent.start_date = e.target.value)}
                                required
                            />
                        </div>
                        <div class="flex-1">
                            <label for="edit_ev_start_time">Kezdő időpont</label
                            >
                            <input
                                id="edit_ev_start_time"
                                type="time"
                                bind:value={editingEvent.start_time}
                            />
                        </div>
                    </div>
                    <div class="flex gap-lg">
                        <div class="flex-1">
                            <label for="edit_ev_end_date">Befejező dátum</label>
                            <input
                                id="edit_ev_end_date"
                                type="date"
                                value={editingEvent.end_date
                                    ? editingEvent.end_date.split("T")[0]
                                    : ""}
                                on:change={(e) =>
                                    (editingEvent.end_date = e.target.value)}
                                required
                            />
                        </div>
                        <div class="flex-1">
                            <label for="edit_ev_end_time"
                                >Befejező időpont</label
                            >
                            <input
                                id="edit_ev_end_time"
                                type="time"
                                bind:value={editingEvent.end_time}
                            />
                        </div>
                    </div>

                    <label for="edit_ev_type">Típus</label>
                    <select
                        id="edit_ev_type"
                        bind:value={editingEvent.event_type}
                    >
                        <option value="cultural">Kulturális</option>
                        <option value="sports">Sport</option>
                        <option value="festival">Fesztivál</option>
                        <option value="religious">Vallási</option>
                        <option value="other">Egyéb</option>
                    </select>

                    <label for="edit_ev_org">Szervező</label>
                    <div class="flex gap-sm items-center">
                        <input
                            id="edit_ev_org"
                            type="text"
                            bind:value={editingEvent.organizer}
                            list="organizers_list"
                            autocomplete="off"
                            class="flex-1"
                        />
                        <button
                            type="button"
                            class="btn-update"
                            style="margin-bottom:0"
                            on:click={() => (newOrganizerModalVisible = true)}
                        >
                            Új szervező
                        </button>
                    </div>

                    <div class="modal-actions">
                        <button type="submit" class="admin-submit-btn"
                            >Mentés</button
                        >
                        <button
                            type="button"
                            class="btn-delete"
                            on:click={cancelEditEvent}>Mégse</button
                        >
                    </div>
                </form>
            </div>
        </div>
    {/if}

    <!-- New Organizer Modal -->
    {#if newOrganizerModalVisible}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={() => (newOrganizerModalVisible = false)}
            on:keydown={(e) =>
                e.key === "Escape" && (newOrganizerModalVisible = false)}
        >
            <div class="admin-modal">
                <h3>Új Szervező Hozzáadása</h3>
                <form class="admin-form" on:submit={submitNewOrganizer}>
                    <label for="org_loc">Település</label>
                    <select
                        id="org_loc"
                        bind:value={newOrganizerEntry.location_id}
                        required
                    >
                        <option value="">Válassz...</option>
                        {#each locations as loc}
                            <option value={loc.id}
                                >{loc.name} ({loc.county})</option
                            >
                        {/each}
                    </select>

                    <label for="org_name">Név (Szervezet Neve)</label>
                    <input
                        id="org_name"
                        type="text"
                        bind:value={newOrganizerEntry.name}
                        required
                    />

                    <label for="org_cat">Kategória</label>
                    <select
                        id="org_cat"
                        bind:value={newOrganizerEntry.category_id}
                    >
                        <option value={null}>-</option>
                        {#each entryCategories as cat}
                            <option value={cat.id}>{cat.name}</option>
                        {/each}
                    </select>

                    <label for="org_phone">Telefon</label>
                    <input
                        id="org_phone"
                        type="text"
                        bind:value={newOrganizerEntry.phone}
                    />

                    <div class="modal-actions mt-lg">
                        <button type="submit" class="admin-submit-btn"
                            >Mentés és Kiválasztás</button
                        >
                        <button
                            type="button"
                            class="btn-delete"
                            on:click={() => (newOrganizerModalVisible = false)}
                            >Mégse</button
                        >
                    </div>
                </form>
            </div>
        </div>
    {/if}
{/if}

<style>
    @import "../../styles/admin.css";

    .admin-modal-overlay {
        position: fixed;
        inset: 0;
        background: rgba(0, 0, 0, 0.6);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
    }
    .admin-modal {
        background: var(--card-bg, #1e1e2e);
        border: 1px solid var(--border-color, #444);
        border-radius: 12px;
        padding: 2rem;
        width: min(600px, 95vw);
        max-height: 90vh;
        overflow-y: auto;
    }
    .admin-modal h3 {
        margin-top: 0;
    }
    .btn-update {
        padding: 0.35rem 0.75rem;
        font-size: 0.8rem;
        border: 1px solid var(--primary, #5c6bc0);
        color: var(--primary, #5c6bc0);
        background: transparent;
        border-radius: 6px;
        cursor: pointer;
        transition: all 0.2s;
    }
    .btn-update:hover {
        background: var(--primary, #5c6bc0);
        color: #fff;
    }
    .badge {
        display: inline-block;
        padding: 0.15rem 0.5rem;
        border-radius: 999px;
        font-size: 0.75rem;
        background: var(--accent-bg, #2a2a3e);
        color: var(--muted, #aaa);
        border: 1px solid var(--border-color, #444);
    }
    .admin-dialog-overlay {
        position: fixed;
        inset: 0;
        background: rgba(0, 0, 0, 0.55);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 2000;
    }
    .admin-dialog {
        background: var(--card-bg, #1e1e2e);
        border: 1px solid var(--border-color, #444);
        border-radius: 12px;
        padding: 1.5rem 2rem;
        width: min(420px, 90vw);
        text-align: center;
    }
    .admin-dialog p {
        margin: 0 0 1.25rem;
        font-size: 1rem;
        line-height: 1.5;
    }
    .admin-dialog-actions {
        display: flex;
        gap: 0.75rem;
        justify-content: center;
    }
    .admin-dialog-actions button {
        min-width: 80px;
    }
    .form-group-label {
        display: block;
        font-weight: 600;
        margin-bottom: 0.35rem;
    }
    .w-full {
        width: 100%;
    }
    .gap-xs {
        gap: 0.3rem;
    }
    .gap-lg {
        gap: 1rem;
    }
    .mt-xs {
        margin-top: 5px;
    }
    .color-swatch {
        display: inline-block;
        width: 20px;
        height: 20px;
        border: 1px solid var(--border-color);
    }
    .modal-actions {
        display: flex;
        gap: 0.75rem;
    }
    .login-box {
        max-width: 400px;
        text-align: center;
        width: 100%;
    }
    .mt-lg {
        margin: 2rem auto;
    }
    .flex {
        display: flex;
    }
    .flex-wrap {
        flex-wrap: wrap;
    }
    .items-center {
        align-items: center;
    }
    .font-normal {
        font-weight: normal;
    }
    .w-auto {
        width: auto;
    }
    .mb-lg {
        margin-bottom: 1rem;
    }
</style>
