<script>
    import { onMount } from "svelte";
    import AdminNavIcon from "$lib/components/admin/AdminNavIcon.svelte";
    import { auth } from "$lib/stores/auth";
    import {
        SCHEDULE_ACTIVITY_TYPES,
        SCHEDULE_ACTIVITY_TYPE_LABELS,
    } from "$lib/scheduleActivityTypes.js";

    const getBase = () =>
        import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

    /** Local calendar date as YYYY-MM-DD (for date inputs). */
    function localISODate() {
        const d = new Date();
        const z = (n) => String(n).padStart(2, "0");
        return `${d.getFullYear()}-${z(d.getMonth() + 1)}-${z(d.getDate())}`;
    }

    let authenticated = false;
    let password = "";
    let activeTab = "welcome";

    /** Section shortcuts on the dashboard (order matches sidebar). */
    const ADMIN_WELCOME_ITEMS = [
        { id: "mondasok", label: "Mondások" },
        { id: "quicklinks", label: "Gyorslinkek" },
        { id: "newsfeeds", label: "Hírfolyamok" },
        { id: "counties", label: "Megyék" },
        { id: "locations", label: "Települések" },
        { id: "venues", label: "Helyszínek" },
        { id: "attractions", label: "Látnivalók" },
        { id: "events", label: "Események" },
        { id: "entry_categories", label: "Bejegyzés kategóriák" },
        { id: "entry_types", label: "Bejegyzés típusok" },
        { id: "entries", label: "Bejegyzések" },
        { id: "weather_translations", label: "Időjárás fordítások" },
        { id: "pages", label: "Oldalak" },
        { id: "settings", label: "Beállítások" },
    ];

    function goToAdminTab(/** @type {string} */ tab) {
        activeTab = tab;
        if (tab === "counties") fetchCountyRegions();
        if (tab === "venues") {
            fetchVenuesCatalog();
            fetchVenueTypes();
        }
        if (tab === "attractions") fetchAttractions();
        if (tab === "events") fetchEvents();
        if (tab === "settings") fetchSettings();
        if (tab === "weather_translations") fetchWeatherTranslations();
        if (tab === "pages") fetchPages();
    }

    /** @type {{ tab: string, message: string } | null} */
    let adminTabError = null;
    $: if (adminTabError && activeTab !== adminTabError.tab) {
        adminTabError = null;
    }

    function setAdminTabError(msg) {
        adminTabError = { tab: activeTab, message: msg };
    }

    function clearAdminTabError() {
        adminTabError = null;
    }

    let mondasok = [];
    let quickLinks = [];
    let newsFeeds = [];
    let locations = [];
    let attractions = [];
    let countiesFromAPI = [];
    let historicalSeatsFromAPI = [];
    /** Inline table edit state for Megyék / Történelmi székek tabs */
    let editingCounty = null;
    let editingHistoricalSeat = null;
    let entries = [];
    let entryCategories = [];
    let entryTypes = [];
    let events = [];

    // Per-feed loading state
    let loadingFeeds = new Set();
    let feedTimestamps = {};

    // Form binding objects
    let newMondas = { text: "", display_date: localISODate() };
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
        default_venue_id: "",
        title: "",
        description: "",
        start_date: "",
        start_time: "",
        end_date: "",
        end_time: "",
        event_type: "cultural",
        organizer: "",
    };
    /** @type {Record<string, unknown>[]} */
    let venuesCatalog = [];
    /** @type {typeof venuesCatalog} */
    let venueOptionsNew = [];
    /** @type {typeof venuesCatalog} */
    let venueOptionsEdit = [];
    let newVenue = {
        settlement_id: "",
        name: "",
        name_ro: "",
        name_de: "",
        slug: "",
        kind: "sports_arena",
        address: "",
        latitude: "",
        longitude: "",
        seating_capacity: "",
        description: "",
        notes: "",
    };
    /** @type {{ id: number, slug: string, label_hu: string }[]} */
    let venueTypesList = [];
    let newVenueType = { label_hu: "" };
    /** @type {Record<string, unknown> | null} */
    let editingVenueType = null;

    let searchMondasok = "";
    let searchQuickLinks = "";
    let searchNewsFeeds = "";
    let searchLocations = "";
    let searchVenueTypes = "";
    let searchVenues = "";
    let searchEvents = "";
    let searchEntryCategories = "";
    let searchEntries = "";
    let searchEntryTypes = "";
    let searchAttractions = "";
    let searchCounties = "";
    let searchHistoricalSeats = "";
    let searchWeatherTrans = "";
    let searchAdminPages = "";
    let searchPageFaqRows = "";
    /** Counties / historical seats: keep inline edit row visible when search would hide it */
    $: displayCounties = (countiesFromAPI || []).filter(
        (c) =>
            editingCounty?.id === c.id ||
            countyMatchesSearch(c, searchCounties),
    );
    $: displayHistoricalSeats = (historicalSeatsFromAPI || []).filter(
        (h) =>
            editingHistoricalSeat?.id === h.id ||
            historicalSeatMatchesSearch(h, searchHistoricalSeats),
    );
    let newAttraction = {
        county_slug: "hargita",
        name: "",
        name_ro: "",
        name_de: "",
        slug: "",
        description: "",
        latitude: "",
        longitude: "",
        featured_image: "",
        content: "",
        images: "",
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
    /** @type {Array<{ schedule_date: string, notes: string, activities: Array<{ activity_type: string, starts_at: string, ends_at: string, title: string, description: string }> }>} */
    let scheduleDraftDays = [];
    let editingAttraction = null;
    /** @type {Record<string, unknown> | null} */
    let editingVenue = null;

    let orgQuery = "";
    let orgEditQuery = "";
    let orgSuggestions = [];
    let orgEditSuggestions = [];
    let orgDropdownOpen = false;
    let orgEditDropdownOpen = false;

    // Site settings (weather providers, cache)
    let siteSettings = {};
    let settingsSaving = false;
    let settingsCacheClearing = false;

    // Weather description translations (multi-language)
    let weatherTranslations = [];
    let newWeatherTrans = { source_text: "", lang: "hu", translated_text: "" };
    let editingWeatherTrans = null;
    const WEATHER_TRANS_LANGS = [
        { value: "hu", label: "Magyar" },
        { value: "ro", label: "Română" },
        { value: "de", label: "Deutsch" },
    ];

    // Pages (policy pages editor)
    let adminPages = [];
    let pageFaqSections = [];
    let editingPage = null;
    let pageSaving = false;
    let editingPageFaq = null;
    let pageFaqSaving = false;

    function filterOrganizers(query, target) {
        if (!query || query.length < 2) return [];
        const q = query.toLowerCase();
        return entries
            .filter((e) => e.name.toLowerCase().includes(q))
            .slice(0, 8);
    }

    function onOrgInput(isEdit = false) {
        const q = isEdit ? orgEditQuery : orgQuery;
        const results = filterOrganizers(q);
        if (isEdit) {
            orgEditSuggestions = results;
            orgEditDropdownOpen = results.length > 0;
        } else {
            orgSuggestions = results;
            orgDropdownOpen = results.length > 0;
        }
    }

    function selectOrganizer(name, isEdit = false) {
        if (isEdit) {
            editingEvent.organizer = name;
            orgEditQuery = name;
            orgEditDropdownOpen = false;
        } else {
            newEvent.organizer = name;
            orgQuery = name;
            orgDropdownOpen = false;
        }
    }

    function handleOrgBlur(isEdit = false) {
        setTimeout(() => {
            if (isEdit) orgEditDropdownOpen = false;
            else orgDropdownOpen = false;
        }, 200);
    }

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

    /** @param {unknown[]} fields */
    function matchesSearch(q, fields) {
        const s = String(q || "").trim().toLowerCase();
        if (!s) return true;
        return fields.some((f) => String(f ?? "").toLowerCase().includes(s));
    }
    /** @template T
     * @param {T[]} rows
     * @param {string} q
     * @param {(row: T) => unknown[]} fieldFn
     */
    function filterRows(rows, q, fieldFn) {
        if (!rows || !rows.length) return rows || [];
        const s = String(q || "").trim().toLowerCase();
        if (!s) return rows;
        return rows.filter((row) =>
            fieldFn(row).some((f) =>
                String(f ?? "")
                    .toLowerCase()
                    .includes(s),
            ),
        );
    }

    onMount(() => {
        auth.init();
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
            auth.login("Admin", true);
            fetchAll();
        } else {
            alert("Na de kicsibarátom, ez nem a jó jelszó!");
        }
    }

    function logout() {
        authenticated = false;
        password = "";
        auth.logout();
        window.location.href = "/";
    }

    async function fetchAll() {
        fetchMondasok();
        fetchQuickLinks();
        fetchNewsFeeds();
        fetchLocations();
        fetchAttractions();
        fetchEntries();
        fetchEntryCategories();
        fetchEntryTypes();
        fetchEvents();
        fetchVenuesCatalog();
        fetchVenueTypes();
        fetchSettings();
        fetchWeatherTranslations();
        fetchPages();
        fetchCountyRegions();
    }

    async function fetchWeatherTranslations() {
        try {
            const res = await fetch(`${getBase()}/api/admin/weather_translations`);
            if (res.ok) weatherTranslations = await res.json();
        } catch (e) {
            console.error(e);
        }
    }

    async function saveWeatherTranslation(e) {
        e?.preventDefault();
        if (editingWeatherTrans) {
            const res = await fetch(`${getBase()}/api/admin/weather_translations`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(editingWeatherTrans),
            });
            if (res.ok) {
                clearAdminTabError();
                fetchWeatherTranslations();
                editingWeatherTrans = null;
            } else {
                setAdminTabError("Hiba: " + (await res.text()));
            }
        } else {
            const res = await fetch(`${getBase()}/api/admin/weather_translations`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(newWeatherTrans),
            });
            if (res.ok) {
                clearAdminTabError();
                fetchWeatherTranslations();
                newWeatherTrans = { source_text: "", lang: "hu", translated_text: "" };
            } else {
                setAdminTabError("Hiba: " + (await res.text()));
            }
        }
    }

    function startEditWeatherTrans(t) {
        editingWeatherTrans = { ...t };
    }

    function cancelEditWeatherTrans() {
        editingWeatherTrans = null;
    }

    async function deleteWeatherTranslation(id) {
        const ok = await showConfirm("Biztosan törölni szeretnéd ezt a fordítást?");
        if (!ok) return;
        const res = await fetch(`${getBase()}/api/admin/weather_translations?id=${id}`, { method: "DELETE" });
        if (res.ok) fetchWeatherTranslations();
        else setAdminTabError("Hiba: " + (await res.text()));
    }

    async function fetchSettings() {
        try {
            const res = await fetch(`${getBase()}/api/admin/settings`);
            if (res.ok) {
                const data = await res.json();
                siteSettings = {
                    weather_cache_ttl_minutes: data.weather_cache_ttl_minutes ?? "15",
                    weather_cache_version: data.weather_cache_version ?? "1",
                    weather_icon_style: data.weather_icon_style ?? "emoji",
                    weather_active_users_estimate: data.weather_active_users_estimate ?? "10000",
                    weather_provider_default: data.weather_provider_default ?? "open_meteo",
                    weather_provider_open_meteo_enabled: data.weather_provider_open_meteo_enabled ?? "true",
                    weather_provider_weatherapi_enabled: data.weather_provider_weatherapi_enabled ?? "true",
                    weather_provider_openweathermap_enabled: data.weather_provider_openweathermap_enabled ?? "true",
                    my_location_slug: data.my_location_slug ?? "csikszereda",
                    ...data,
                };
            }
        } catch (e) {
            console.error(e);
        }
    }

    async function saveSettings() {
        settingsSaving = true;
        try {
            const payload = Object.fromEntries(
                Object.entries(siteSettings).map(([k, v]) => [k, v != null ? String(v) : ""])
            );
            const res = await fetch(`${getBase()}/api/admin/settings`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(payload),
            });
            if (res.ok) {
                clearAdminTabError();
                await showAlert("Beállítások mentve.");
            } else setAdminTabError("Hiba: " + (await res.text()));
        } catch (e) {
            setAdminTabError("Hiba: " + e.message);
        } finally {
            settingsSaving = false;
        }
    }

    async function clearWeatherCache() {
        settingsCacheClearing = true;
        try {
            const res = await fetch(`${getBase()}/api/admin/settings/clear-weather-cache`, { method: "POST" });
            if (res.ok) {
                clearAdminTabError();
                await showAlert("Időjárás cache verzió növelve – látogatók friss adatot fognak kapni.");
                fetchSettings();
            } else setAdminTabError("Hiba: " + (await res.text()));
        } catch (e) {
            setAdminTabError("Hiba: " + e.message);
        } finally {
            settingsCacheClearing = false;
        }
    }

    async function fetchPages() {
        try {
            const res = await fetch(`${getBase()}/api/admin/pages`);
            if (res.ok) adminPages = await res.json();
            const r2 = await fetch(`${getBase()}/api/admin/page_faq`);
            if (r2.ok) pageFaqSections = await r2.json();
        } catch (e) {
            console.error(e);
        }
    }

    function startEditPage(page) {
        editingPageFaq = null;
        editingPage = { ...page };
    }

    function cancelEditPage() {
        editingPage = null;
    }

    async function savePage() {
        if (!editingPage) return;
        pageSaving = true;
        try {
            const res = await fetch(`${getBase()}/api/admin/pages`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(editingPage),
            });
            if (res.ok) {
                clearAdminTabError();
                await showAlert("Oldal mentve.");
                editingPage = null;
                fetchPages();
            } else {
                setAdminTabError("Hiba: " + (await res.text()));
            }
        } catch (e) {
            setAdminTabError("Hiba: " + e.message);
        } finally {
            pageSaving = false;
        }
    }

    function startEditPageFaq(row) {
        editingPage = null;
        const items = Array.isArray(row.faq_items)
            ? row.faq_items.map((x) => ({
                  question: x.question ?? "",
                  answer: x.answer ?? "",
              }))
            : [];
        editingPageFaq = { ...row, faq_items: items };
    }

    function cancelEditPageFaq() {
        editingPageFaq = null;
    }

    function addFaqItem() {
        if (!editingPageFaq) return;
        editingPageFaq.faq_items = [
            ...(editingPageFaq.faq_items || []),
            { question: "", answer: "" },
        ];
    }

    function removeFaqItem(index) {
        if (!editingPageFaq?.faq_items) return;
        editingPageFaq.faq_items = editingPageFaq.faq_items.filter(
            (_, i) => i !== index,
        );
    }

    async function savePageFaq() {
        if (!editingPageFaq) return;
        pageFaqSaving = true;
        try {
            const res = await fetch(`${getBase()}/api/admin/page_faq`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    id: editingPageFaq.id,
                    label_hu: editingPageFaq.label_hu ?? "",
                    faq_title: editingPageFaq.faq_title ?? "",
                    faq_items: editingPageFaq.faq_items ?? [],
                    disclaimer_markdown: editingPageFaq.disclaimer_markdown ?? "",
                }),
            });
            if (res.ok) {
                clearAdminTabError();
                await showAlert("GYIK / disclaimer mentve.");
                editingPageFaq = null;
                fetchPages();
            } else {
                setAdminTabError("Hiba: " + (await res.text()));
            }
        } catch (e) {
            setAdminTabError("Hiba: " + e.message);
        } finally {
            pageFaqSaving = false;
        }
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
    // Entries/events use settlement_id; filter out counties (type=megye)
    $: settlementsForSelect = locations.filter((l) => l.type !== "megye");

    /** @returns {Promise<boolean>} */
    async function setCountySeat(locationId) {
        try {
            const res = await fetch(`${getBase()}/api/admin/county_seat`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ location_id: locationId }),
            });
            if (res.ok) {
                fetchLocations();
                return true;
            }
            console.error("Failed to set county seat:", await res.text());
            return false;
        } catch (e) {
            console.error("Error setting county seat:", e);
            return false;
        }
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
    async function fetchVenuesCatalog() {
        try {
            const res = await fetch(`${getBase()}/api/admin/venues`);
            if (res.ok) venuesCatalog = await res.json();
        } catch (e) {
            console.error(e);
        }
    }
    async function fetchVenueTypes() {
        try {
            const res = await fetch(`${getBase()}/api/admin/venue_types`);
            if (res.ok) {
                venueTypesList = await res.json();
                if (venueTypesList.length) {
                    const slugs = new Set(venueTypesList.map((t) => t.slug));
                    if (!slugs.has(String(newVenue.kind))) {
                        newVenue = {
                            ...newVenue,
                            kind: venueTypesList[0].slug,
                        };
                    }
                    if (
                        editingVenue &&
                        !slugs.has(String(editingVenue.kind))
                    ) {
                        editingVenue = {
                            ...editingVenue,
                            kind: venueTypesList[0].slug,
                        };
                    }
                }
            }
        } catch (e) {
            console.error(e);
        }
    }
    async function submitNewVenueType(e) {
        e.preventDefault();
        if (!String(newVenueType.label_hu || "").trim()) {
            await showAlert("A megnevezés kötelező.");
            return;
        }
        try {
            const res = await fetch(`${getBase()}/api/admin/venue_types`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    label_hu: String(newVenueType.label_hu).trim(),
                }),
            });
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            newVenueType = { label_hu: "" };
            await fetchVenueTypes();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }
    /** @param {Record<string, unknown>} t */
    function startEditVenueType(t) {
        editingVenueType = {
            id: t.id,
            slug: t.slug,
            label_hu: t.label_hu ?? "",
        };
    }
    function cancelEditVenueType() {
        editingVenueType = null;
    }
    async function saveEditVenueType() {
        if (!editingVenueType) return;
        const id = parseInt(String(editingVenueType.id || ""), 10);
        if (!Number.isFinite(id) || id < 1) return;
        try {
            const res = await fetch(`${getBase()}/api/admin/venue_types`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    id,
                    label_hu: String(editingVenueType.label_hu || "").trim(),
                }),
            });
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            editingVenueType = null;
            await fetchVenueTypes();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }
    async function deleteVenueTypeRow(id) {
        const ok = await showConfirm(
            "Biztosan törlöd ezt a helyszíntípust? (Nem lehetséges, ha van hozzárendelt helyszín.)",
        );
        if (!ok) return;
        try {
            const res = await fetch(
                `${getBase()}/api/admin/venue_types?id=${encodeURIComponent(id)}`,
                { method: "DELETE" },
            );
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            await fetchVenueTypes();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }
    async function loadVenuesForNewEvent() {
        const sid = parseInt(String(newEvent.location_id || ""), 10);
        if (!Number.isFinite(sid) || sid < 1) {
            venueOptionsNew = [];
            return;
        }
        try {
            const res = await fetch(
                `${getBase()}/api/venues?settlement_id=${sid}`,
            );
            venueOptionsNew = res.ok ? await res.json() : [];
        } catch (e) {
            console.error(e);
            venueOptionsNew = [];
        }
    }
    async function loadVenuesForEditSettlement(sidRaw) {
        const sid = parseInt(String(sidRaw || ""), 10);
        if (!Number.isFinite(sid) || sid < 1) {
            venueOptionsEdit = [];
            return;
        }
        try {
            const res = await fetch(
                `${getBase()}/api/venues?settlement_id=${sid}`,
            );
            venueOptionsEdit = res.ok ? await res.json() : [];
        } catch (e) {
            console.error(e);
            venueOptionsEdit = [];
        }
    }
    function parseOptFloatVenue(s) {
        const t = String(s ?? "")
            .trim()
            .replace(",", ".");
        if (!t) return null;
        const n = parseFloat(t);
        return Number.isFinite(n) ? n : null;
    }
    function parseOptIntVenue(s) {
        const t = String(s ?? "").trim();
        if (!t) return null;
        const n = parseInt(t, 10);
        return Number.isFinite(n) ? n : null;
    }

    function emptyNewVenue() {
        return {
            settlement_id: "",
            name: "",
            name_ro: "",
            name_de: "",
            slug: "",
            kind: "sports_arena",
            address: "",
            latitude: "",
            longitude: "",
            seating_capacity: "",
            description: "",
            notes: "",
        };
    }

    async function submitNewVenue(e) {
        e.preventDefault();
        const sid = parseInt(String(newVenue.settlement_id || ""), 10);
        if (!Number.isFinite(sid) || sid < 1 || !String(newVenue.name || "").trim()) {
            await showAlert("Válassz települést és adj meg nevet.");
            return;
        }
        try {
            const res = await fetch(`${getBase()}/api/admin/venues`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    settlement_id: sid,
                    name: newVenue.name.trim(),
                    name_ro: String(newVenue.name_ro || "").trim(),
                    name_de: String(newVenue.name_de || "").trim(),
                    slug: newVenue.slug?.trim() || "",
                    kind: newVenue.kind || "other",
                    address: newVenue.address || "",
                    notes: newVenue.notes || "",
                    latitude: parseOptFloatVenue(newVenue.latitude),
                    longitude: parseOptFloatVenue(newVenue.longitude),
                    seating_capacity: parseOptIntVenue(newVenue.seating_capacity),
                    description: String(newVenue.description || "").trim(),
                }),
            });
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            await fetchVenuesCatalog();
            await loadVenuesForNewEvent();
            if (editingEvent)
                await loadVenuesForEditSettlement(editingEvent.location_id);
            newVenue = emptyNewVenue();
            await showAlert("Helyszín elmentve.");
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }

    /** @param {Record<string, unknown>} v */
    function startEditVenue(v) {
        editingVenue = {
            id: v.id,
            settlement_id: String(v.settlement_id ?? ""),
            name: v.name ?? "",
            name_ro: v.name_ro ?? "",
            name_de: v.name_de ?? "",
            slug: v.slug ?? "",
            kind: v.kind ?? "other",
            address: v.address ?? "",
            notes: v.notes ?? "",
            latitude:
                v.latitude != null && v.latitude !== ""
                    ? String(v.latitude)
                    : "",
            longitude:
                v.longitude != null && v.longitude !== ""
                    ? String(v.longitude)
                    : "",
            seating_capacity:
                v.seating_capacity != null && v.seating_capacity !== ""
                    ? String(v.seating_capacity)
                    : "",
            description: v.description ?? "",
        };
    }
    function cancelEditVenue() {
        editingVenue = null;
    }
    async function saveEditVenue() {
        if (!editingVenue) return;
        const sid = parseInt(String(editingVenue.settlement_id || ""), 10);
        const id = parseInt(String(editingVenue.id || ""), 10);
        if (!Number.isFinite(sid) || sid < 1 || !Number.isFinite(id) || id < 1) {
            await showAlert("Érvénytelen azonosító.");
            return;
        }
        if (!String(editingVenue.name || "").trim()) {
            await showAlert("A magyar név kötelező.");
            return;
        }
        try {
            const res = await fetch(`${getBase()}/api/admin/venues`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    id,
                    settlement_id: sid,
                    name: String(editingVenue.name).trim(),
                    name_ro: String(editingVenue.name_ro || "").trim(),
                    name_de: String(editingVenue.name_de || "").trim(),
                    slug: String(editingVenue.slug || "").trim(),
                    kind: editingVenue.kind || "other",
                    address: String(editingVenue.address || ""),
                    notes: String(editingVenue.notes || ""),
                    latitude: parseOptFloatVenue(editingVenue.latitude),
                    longitude: parseOptFloatVenue(editingVenue.longitude),
                    seating_capacity: parseOptIntVenue(editingVenue.seating_capacity),
                    description: String(editingVenue.description || "").trim(),
                }),
            });
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            editingVenue = null;
            await fetchVenuesCatalog();
            await loadVenuesForNewEvent();
            if (editingEvent)
                await loadVenuesForEditSettlement(editingEvent.location_id);
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }
    async function deleteVenueRow(id) {
        const ok = await showConfirm("Biztosan törlöd ezt a helyszínt?");
        if (!ok) return;
        try {
            const res = await fetch(
                `${getBase()}/api/admin/venues?id=${encodeURIComponent(id)}`,
                { method: "DELETE" },
            );
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            await fetchVenuesCatalog();
            await loadVenuesForNewEvent();
            if (editingEvent)
                await loadVenuesForEditSettlement(editingEvent.location_id);
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }

    /** @param {Record<string, unknown>} ev */
    function eventDateTimeComplete(ev) {
        const sd = String(ev.start_date ?? "").trim();
        const ed = String(ev.end_date ?? "").trim();
        const st = String(ev.start_time ?? "").trim();
        const et = String(ev.end_time ?? "").trim();
        return !!(sd && ed && st && et);
    }

    /** @param {Record<string, unknown>} ev */
    function validateEventFields(ev) {
        const loc = ev.location_id;
        const locNum =
            typeof loc === "number"
                ? loc
                : parseInt(String(loc ?? ""), 10);
        if (
            loc === "" ||
            loc === null ||
            loc === undefined ||
            !Number.isFinite(locNum) ||
            locNum <= 0
        ) {
            return "Válassz települést / helyszínt.";
        }
        if (!String(ev.title ?? "").trim()) return "Az esemény címe kötelező.";
        if (!String(ev.start_date ?? "").trim()) return "A kezdő dátum kötelező.";
        if (!String(ev.end_date ?? "").trim()) return "A befejező dátum kötelező.";
        if (!String(ev.start_time ?? "").trim())
            return "A kezdő időpont (óra:perc) kötelező.";
        if (!String(ev.end_time ?? "").trim())
            return "A befejező időpont (óra:perc) kötelező.";
        return null;
    }

    $: eventsWithIncompleteDateTime = events.filter((e) => !eventDateTimeComplete(e));
    function fetchAttractions() {
        loadData("attractions", (d) => (attractions = d));
    }

    async function fetchCountyRegions() {
        try {
            const r1 = await fetch(`${getBase()}/api/counties`);
            if (r1.ok) countiesFromAPI = await r1.json();
            const r2 = await fetch(`${getBase()}/api/historical_seats`);
            if (r2.ok) historicalSeatsFromAPI = await r2.json();
        } catch (e) {
            console.error(e);
        }
    }

    /** One limit for every truncated label in admin tables (full value in title/tooltip). */
    const ADMIN_TABLE_PREVIEW_MAX = 20;

    function contentPreview(text, maxLen = ADMIN_TABLE_PREVIEW_MAX) {
        if (!text || !String(text).trim()) return "—";
        const t = String(text).replace(/\s+/g, " ").trim();
        return t.length > maxLen ? t.slice(0, maxLen) + "…" : t;
    }

    /** URLs and long strings in table cells use the same max length as contentPreview. */
    function urlPreview(url, maxLen = ADMIN_TABLE_PREVIEW_MAX) {
        if (!url || !String(url).trim()) return "—";
        const s = String(url).trim();
        return s.length > maxLen ? s.slice(0, maxLen) + "…" : s;
    }

    function formatLatLon(lat, lon) {
        const la = lat != null && lat !== "" ? Number(lat) : NaN;
        const lo = lon != null && lon !== "" ? Number(lon) : NaN;
        if (!Number.isFinite(la) && !Number.isFinite(lo)) return "—";
        if (la === 0 && lo === 0) return "—";
        const a = Number.isFinite(la) ? la.toFixed(4) : "—";
        const o = Number.isFinite(lo) ? lo.toFixed(4) : "—";
        return `${a}, ${o}`;
    }

    function settlementsForCountyName(countyName) {
        return locations
            .filter((l) => l.county === countyName && l.type !== "megye")
            .sort((a, b) => {
                if (a.is_county_seat && !b.is_county_seat) return -1;
                if (!a.is_county_seat && b.is_county_seat) return 1;
                const typeOrder = { municípium: 0, város: 1, község: 2, falu: 3 };
                const ta = typeOrder[a.type] ?? 9;
                const tb = typeOrder[b.type] ?? 9;
                if (ta !== tb) return ta - tb;
                return a.name.localeCompare(b.name);
            });
    }

    function countySeatDisplayName(c) {
        const seat = locations.find(
            (l) => l.county === c.name && l.type !== "megye" && l.is_county_seat,
        );
        return seat ? `${seat.name} (${seat.type})` : "—";
    }

    function countyMatchesSearch(c, q) {
        return matchesSearch(q, [
            c.name,
            c.name_ro,
            c.name_de,
            c.slug,
            countySeatDisplayName(c),
            c.content || "",
        ]);
    }

    function historicalSeatMatchesSearch(h, q) {
        return matchesSearch(q, [
            h.name,
            h.name_ro,
            h.name_de,
            h.slug,
            h.content || "",
        ]);
    }

    function startEditCounty(c) {
        editingHistoricalSeat = null;
        const seat = locations.find((l) => l.county === c.name && l.is_county_seat);
        editingCounty = {
            id: c.id,
            name: c.name ?? "",
            name_ro: c.name_ro ?? "",
            name_de: c.name_de ?? "",
            slug: c.slug ?? "",
            content: c.content ?? "",
            seat_location_id: seat ? String(seat.id) : "",
        };
    }

    function cancelEditCounty() {
        editingCounty = null;
    }

    async function saveEditingCounty() {
        const ec = editingCounty;
        if (!ec) return;
        try {
            const res = await fetch(`${getBase()}/api/admin/counties`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    id: ec.id,
                    name: ec.name ?? "",
                    name_ro: ec.name_ro ?? "",
                    name_de: ec.name_de ?? "",
                    slug: ec.slug ?? "",
                    content: ec.content ?? "",
                }),
            });
            if (!res.ok) {
                setAdminTabError("Hiba (megye): " + (await res.text()));
                return;
            }
            if (ec.seat_location_id) {
                const ok = await setCountySeat(Number(ec.seat_location_id));
                if (!ok) {
                    setAdminTabError(
                        "A megye szövege mentve, de a megyeszékhely beállítása nem sikerült.",
                    );
                    editingCounty = null;
                    fetchCountyRegions();
                    fetchLocations();
                    return;
                }
            }
            editingCounty = null;
            clearAdminTabError();
            await showAlert("Megye mentve: " + ec.name);
            fetchCountyRegions();
            fetchLocations();
        } catch (e) {
            setAdminTabError("Hiba: " + e.message);
        }
    }

    function startEditHistoricalSeat(h) {
        editingCounty = null;
        editingHistoricalSeat = {
            id: h.id,
            name: h.name ?? "",
            name_ro: h.name_ro ?? "",
            name_de: h.name_de ?? "",
            slug: h.slug ?? "",
            content: h.content ?? "",
        };
    }

    function cancelEditHistoricalSeat() {
        editingHistoricalSeat = null;
    }

    async function saveEditingHistoricalSeat() {
        const h = editingHistoricalSeat;
        if (!h) return;
        try {
            const res = await fetch(`${getBase()}/api/admin/historical_seats`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    id: h.id,
                    name: h.name ?? "",
                    name_ro: h.name_ro ?? "",
                    name_de: h.name_de ?? "",
                    slug: h.slug ?? "",
                    content: h.content ?? "",
                }),
            });
            if (!res.ok) {
                setAdminTabError("Hiba (szék): " + (await res.text()));
                return;
            }
            editingHistoricalSeat = null;
            clearAdminTabError();
            await showAlert("Szék mentve: " + h.name);
            fetchCountyRegions();
        } catch (e) {
            setAdminTabError("Hiba: " + e.message);
        }
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
                clearAdminTabError();
                reloadFunc();
                resetFormFunc();
            } else {
                setAdminTabError("Hiba: " + (await res.text()));
            }
        } catch (e) {
            console.error(e);
            setAdminTabError("Hiba: " + (e && e.message ? e.message : String(e)));
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
                clearAdminTabError();
                reloadFunc();
            } else {
                setAdminTabError("Mentési hiba: " + (await res.text()));
            }
        } catch (e) {
            console.error(e);
            setAdminTabError("Hiba: " + (e && e.message ? e.message : String(e)));
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

    /** Normalize YYYY-MM-DD from date input or API (may include time). */
    function normalizeYmdInput(v) {
        const s = String(v ?? "").trim();
        return s.length >= 10 ? s.slice(0, 10) : s;
    }

    // specific creates
    function submitMondas(e) {
        e.preventDefault();
        const display_date =
            normalizeYmdInput(newMondas.display_date) || localISODate();
        const text = String(newMondas.text ?? "").trim();
        createRecord(
            "mondasok",
            { text, display_date },
            fetchMondasok,
            () => (newMondas = { text: "", display_date: localISODate() }),
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
    async function submitEvent(e) {
        e.preventDefault();
        const lid = parseInt(String(newEvent.location_id), 10);
        const dv = parseInt(String(newEvent.default_venue_id || ""), 10);
        const payload = {
            ...newEvent,
            location_id: Number.isFinite(lid) && lid > 0 ? lid : 0,
            default_venue_id:
                Number.isFinite(dv) && dv > 0 ? dv : null,
        };
        const err = validateEventFields(payload);
        if (err) {
            setAdminTabError(err);
            return;
        }
        createRecord(
            "events",
            payload,
            fetchEvents,
            () =>
                (newEvent = {
                    location_id: "",
                    default_venue_id: "",
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
            orgQuery = newOrganizerEntry.name;
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
        editingMondas = {
            ...m,
            display_date:
                normalizeYmdInput(m.display_date) || localISODate(),
        };
    }
    function cancelEditMondas() {
        editingMondas = null;
    }
    async function saveEditMondas() {
        if (!editingMondas) return;
        const ok = await showConfirm("Biztosan menteni szeretné a módosítást?");
        if (!ok) return;
        const id = parseInt(String(editingMondas.id ?? ""), 10);
        const display_date =
            normalizeYmdInput(editingMondas.display_date) || localISODate();
        const text = String(editingMondas.text ?? "").trim();
        await updateRecord(
            "mondasok",
            { id, text, display_date },
            fetchMondasok,
        );
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
        const st = ev.start_time ? String(ev.start_time) : "";
        const et = ev.end_time ? String(ev.end_time) : "";
        editingEvent = {
            ...ev,
            default_venue_id:
                ev.default_venue_id != null && ev.default_venue_id !== ""
                    ? String(ev.default_venue_id)
                    : "",
            start_time: st.length >= 5 ? st.slice(0, 5) : "",
            end_time: et.length >= 5 ? et.slice(0, 5) : "",
        };
        orgEditQuery = ev.organizer || "";
        await loadVenuesForEditSettlement(ev.location_id);
        await loadScheduleForEditing(ev.id);
    }
    function cancelEditEvent() {
        editingEvent = null;
        scheduleDraftDays = [];
    }

    async function loadScheduleForEditing(eventId) {
        try {
            const res = await fetch(
                `${getBase()}/api/admin/events/schedule?event_id=${eventId}`,
            );
            if (!res.ok) {
                scheduleDraftDays = [];
                return;
            }
            const data = await res.json();
            scheduleDraftDays = (data.days || []).map((d) => ({
                schedule_date: d.schedule_date,
                notes: d.notes || "",
                activities: (d.activities || []).map((a) => ({
                    activity_type: a.activity_type || "other",
                    starts_at: a.starts_at
                        ? String(a.starts_at).slice(0, 5)
                        : "",
                    ends_at: a.ends_at ? String(a.ends_at).slice(0, 5) : "",
                    venue_id:
                        a.venue_id != null && a.venue_id !== ""
                            ? String(a.venue_id)
                            : "",
                    title: a.title || "",
                    description: a.description || "",
                })),
            }));
        } catch (e) {
            console.error(e);
            scheduleDraftDays = [];
        }
    }

    async function generateScheduleDaysFromEvent() {
        if (!editingEvent) return;
        const ok = await showConfirm(
            "A jelenlegi napi program törlődik, és a kezdő–befejező dátum közötti minden nap üres programmal kerül be. Folytatja?",
        );
        if (!ok) return;
        const s = editingEvent.start_date?.split("T")[0];
        const e = editingEvent.end_date?.split("T")[0];
        if (!s || !e) {
            await showAlert("Előbb állítsa be a kezdő és befejező dátumot.");
            return;
        }
        const out = [];
        const cur = new Date(s + "T12:00:00");
        const end = new Date(e + "T12:00:00");
        while (cur <= end) {
            out.push({
                schedule_date: cur.toISOString().slice(0, 10),
                notes: "",
                activities: [],
            });
            cur.setDate(cur.getDate() + 1);
        }
        scheduleDraftDays = out;
    }

    function addScheduleDayRow() {
        scheduleDraftDays = [
            ...scheduleDraftDays,
            { schedule_date: "", notes: "", activities: [] },
        ];
    }

    function removeScheduleDayRow(i) {
        scheduleDraftDays = scheduleDraftDays.filter((_, j) => j !== i);
    }

    function addScheduleActivity(dayIndex) {
        const d = scheduleDraftDays[dayIndex];
        if (!d) return;
        d.activities = [
            ...d.activities,
            {
                activity_type: "match",
                starts_at: "",
                ends_at: "",
                venue_id: "",
                title: "",
                description: "",
            },
        ];
        scheduleDraftDays = [...scheduleDraftDays];
    }

    function removeScheduleActivity(dayIndex, actIndex) {
        const d = scheduleDraftDays[dayIndex];
        if (!d) return;
        d.activities = d.activities.filter((_, j) => j !== actIndex);
        scheduleDraftDays = [...scheduleDraftDays];
    }

    async function saveEventSchedule() {
        if (!editingEvent) return;
        try {
            const body = {
                event_id: editingEvent.id,
                days: scheduleDraftDays
                    .filter((d) => d.schedule_date && String(d.schedule_date).trim())
                    .map((d) => ({
                        schedule_date: d.schedule_date,
                        notes: "",
                        activities: (d.activities || [])
                            .filter((a) => a.title && String(a.title).trim())
                            .map((a, ai) => {
                                const vid = parseInt(
                                    String(a.venue_id || ""),
                                    10,
                                );
                                return {
                                    activity_type: a.activity_type || "other",
                                    starts_at: a.starts_at || "",
                                    ends_at: a.ends_at || "",
                                    venue_id:
                                        Number.isFinite(vid) && vid > 0
                                            ? vid
                                            : null,
                                    title: a.title.trim(),
                                    description: a.description || "",
                                    sort_order: ai,
                                };
                            }),
                    })),
            };
            const res = await fetch(`${getBase()}/api/admin/events/schedule`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(body),
            });
            if (!res.ok) {
                await showAlert(
                    "Program mentése sikertelen: " + (await res.text()),
                );
                return;
            }
            await showAlert("Napi program elmentve.");
            await loadScheduleForEditing(editingEvent.id);
        } catch (err) {
            await showAlert("Hiba: " + err.message);
        }
    }
    async function saveEditEvent() {
        if (!editingEvent) return;
        const err = validateEventFields(editingEvent);
        if (err) {
            await showAlert(err);
            return;
        }
        const ok = await showConfirm("Biztosan menteni szeretné a módosítást?");
        if (!ok) return;
        const dv = parseInt(
            String(editingEvent.default_venue_id || ""),
            10,
        );
        const payload = {
            ...editingEvent,
            default_venue_id:
                Number.isFinite(dv) && dv > 0 ? dv : null,
        };
        await updateRecord("events", payload, fetchEvents);
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
        if (id == null || id === "") return "—";
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

    // --- Attractions ---
    function submitNewAttraction(e) {
        e.preventDefault();
        const imgs = newAttraction.images
            ? newAttraction.images.split("\n").map((s) => s.trim()).filter(Boolean)
            : [];
        createRecord(
            "attractions",
            {
                county_slug: newAttraction.county_slug,
                name: newAttraction.name,
                name_ro: newAttraction.name_ro || "",
                name_de: newAttraction.name_de || "",
                slug: newAttraction.slug || "",
                description: newAttraction.description || "",
                latitude: parseFloat(newAttraction.latitude) || 0,
                longitude: parseFloat(newAttraction.longitude) || 0,
                featured_image: newAttraction.featured_image || "",
                content: newAttraction.content || "",
                images: imgs,
            },
            fetchAttractions,
            () =>
                (newAttraction = {
                    county_slug: "hargita",
                    name: "",
                    name_ro: "",
                    name_de: "",
                    slug: "",
                    description: "",
                    latitude: "",
                    longitude: "",
                    featured_image: "",
                    content: "",
                    images: "",
                }),
        );
    }
    function openEditAttraction(att) {
        editingAttraction = {
            id: att.id,
            county_slug: att.county_slug,
            name: att.name,
            name_ro: att.name_ro || "",
            name_de: att.name_de || "",
            slug: att.slug,
            description: att.description || "",
            latitude: att.latitude ? String(att.latitude) : "",
            longitude: att.longitude ? String(att.longitude) : "",
            featured_image: att.featured_image || "",
            content: att.content || "",
            images: (att.images || []).join("\n"),
        };
    }
    function cancelEditAttraction() {
        editingAttraction = null;
    }
    async function saveEditAttraction(e) {
        e.preventDefault();
        if (!editingAttraction) return;
        const imgs = editingAttraction.images
            ? editingAttraction.images.split("\n").map((s) => s.trim()).filter(Boolean)
            : [];
        await updateRecord(
            "attractions",
            {
                ...editingAttraction,
                latitude: parseFloat(editingAttraction.latitude) || 0,
                longitude: parseFloat(editingAttraction.longitude) || 0,
                images: imgs,
            },
            fetchAttractions,
        );
        cancelEditAttraction();
    }
    async function deleteAttraction(id) {
        if (!confirm("Biztosan törölni szeretnéd ezt a látnivalót?")) return;
        try {
            const res = await fetch(`${getBase()}/api/admin/attractions?id=${id}`, { method: "DELETE" });
            if (res.ok) fetchAttractions();
        } catch (e) {
            console.error(e);
        }
    }
</script>

<svelte:head>
    <title>Lámsza - Adminisztráció</title>
</svelte:head>

{#if !authenticated}
    <div class="container">
        <div class="admin-login-wrapper">
            <div class="admin-container login-box">
                <h2>Adminisztráció Belépés</h2>
                <form
                    class="admin-form mt-lg"
                    on:submit={login}
                    autocomplete="on"
                >
                    <input
                        id="admin-login-password"
                        name="password"
                        type="password"
                        bind:value={password}
                        placeholder="Jelszó..."
                        autocomplete="current-password"
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
            <div class="admin-sidebar-inner">
            <a
                href="/"
                target="_blank"
                rel="noopener noreferrer"
                class="admin-sidebar-btn admin-sidebar-btn--external"
                title="Open homepage in new tab"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" aria-hidden="true"
                    ><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline
                        points="15 3 21 3 21 9"
                    ></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg
                >
            </a>

            <button
                type="button"
                class="admin-sidebar-btn {activeTab === 'welcome' ? 'active' : ''}"
                on:click={() => goToAdminTab('welcome')}
                title="Dashboard home"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" aria-hidden="true"
                    ><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline
                        points="9 22 9 12 15 12 15 22"
                    ></polyline></svg
                >
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'mondasok'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("mondasok")}
                title="Mondások"
            >
                <AdminNavIcon name="mondasok" />
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'quicklinks'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("quicklinks")}
                title="Gyorslinkek"
            >
                <AdminNavIcon name="quicklinks" />
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'newsfeeds'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("newsfeeds")}
                title="Hírfolyamok"
            >
                <AdminNavIcon name="newsfeeds" />
            </button>

            <hr class="admin-sidebar-sep" aria-hidden="true" />

            <button
                class="admin-sidebar-btn {activeTab === 'counties'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("counties")}
                title="Megyék"
            >
                <AdminNavIcon name="counties" />
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'locations'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("locations")}
                title="Települések"
            >
                <AdminNavIcon name="locations" />
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'venues' ? 'active' : ''}"
                on:click={() => goToAdminTab("venues")}
                title="Helyszínek"
            >
                <AdminNavIcon name="venues" />
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'attractions'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("attractions")}
                title="Látnivalók"
            >
                <AdminNavIcon name="attractions" />
            </button>

            <hr class="admin-sidebar-sep" aria-hidden="true" />

            <button
                class="admin-sidebar-btn {activeTab === 'events'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("events")}
                title="Események"
            >
                <AdminNavIcon name="events" />
            </button>

            <hr class="admin-sidebar-sep" aria-hidden="true" />

            <button
                class="admin-sidebar-btn {activeTab === 'entry_categories'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("entry_categories")}
                title="Bejegyzés Kategóriák"
            >
                <AdminNavIcon name="entry_categories" />
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'entry_types'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("entry_types")}
                title="Bejegyzés típusok"
            >
                <AdminNavIcon name="entry_types" />
            </button>
            <button
                class="admin-sidebar-btn {activeTab === 'entries'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("entries")}
                title="Index"
            >
                <AdminNavIcon name="entries" />
            </button>

            <hr class="admin-sidebar-sep" aria-hidden="true" />

            <button
                class="admin-sidebar-btn {activeTab === 'weather_translations' ? 'active' : ''}"
                on:click={() => goToAdminTab('weather_translations')}
                title="Időjárás fordítások"
            >
                <AdminNavIcon name="weather_translations" />
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'pages' ? 'active' : ''}"
                on:click={() => goToAdminTab('pages')}
                title="Oldalak"
            >
                <AdminNavIcon name="pages" />
            </button>

            <hr class="admin-sidebar-sep" aria-hidden="true" />

            <button
                class="admin-sidebar-btn {activeTab === 'settings'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("settings")}
                title="Beállítások"
            >
                <AdminNavIcon name="settings" />
            </button>
            </div>
        </aside>

        <main class="admin-main">
            <div class="admin-header">
                <h2>
                    {#if activeTab === "welcome"}Dashboard{/if}
                    {#if activeTab === "mondasok"}Mondások Kezelése{/if}
                    {#if activeTab === "quicklinks"}Gyorslinkek Kezelése{/if}
                    {#if activeTab === "newsfeeds"}Hírfolyamok Kezelése{/if}
                    {#if activeTab === "locations"}Települések Kezelése{/if}
                    {#if activeTab === "venues"}Helyszínek Kezelése{/if}
                    {#if activeTab === "entry_categories"}Bejegyzés Kategóriák
                        Kezelése{/if}
                    {#if activeTab === "entries"}Bejegyzések Kezelése{/if}
                    {#if activeTab === "entry_types"}Bejegyzés Típusok Kezelése{/if}
                    {#if activeTab === "attractions"}Látnivalók Kezelése{/if}
                    {#if activeTab === "counties"}Megyék Kezelése{/if}
                    {#if activeTab === "settings"}Beállítások{/if}
                    {#if activeTab === "weather_translations"}Időjárás fordítások{/if}
                    {#if activeTab === "pages"}Oldalak{/if}
                    {#if activeTab === "events"}Események Kezelése{/if}
                </h2>
                <button class="btn-logout" on:click={logout}
                    >Kijelentkezés</button
                >
            </div>

            <div class="admin-container w-full">
                {#if activeTab === "welcome"}
                    <div class="admin-welcome" role="navigation" aria-label="Admin sections">
                        <div class="admin-welcome-grid">
                            {#each ADMIN_WELCOME_ITEMS as item}
                                <button
                                    type="button"
                                    class="admin-welcome-card"
                                    on:click={() => goToAdminTab(item.id)}
                                    aria-label={item.label}
                                >
                                    <AdminNavIcon name={item.id} size={56} />
                                </button>
                            {/each}
                        </div>
                    </div>
                {/if}

                <!-- Mondások Tab -->
                {#if activeTab === "mondasok"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            A kezdőlapon az adott <strong>naptári napra</strong> beütemezett mondások jelennek meg
                            (a látogató böngészőjének helyi dátuma, ugyanaz mint a „Dátum és idő” widget a főoldalon).
                            Ugyanarra a napra több mondás is
                            beállítható. Ha nincs egyetlen idézet sem az aktuális napra, a főoldalon nem
                            jelenik meg mondás-blokk.
                        </p>
                    {/if}
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"
                            >Új mondás hozzáadása</summary
                        >
                        <form class="admin-form admin-create-form" on:submit={submitMondas}>
                            <label for="mondas_text">Mondás szövege</label>
                            <textarea
                                id="mondas_text"
                                name="mondas_text"
                                bind:value={newMondas.text}
                                required
                                rows="3"
                            ></textarea>
                            <label for="mondas_day">Megjelenés napja</label>
                            <input
                                id="mondas_day"
                                name="display_date"
                                type="date"
                                bind:value={newMondas.display_date}
                                required
                            />
                            <button type="submit" class="admin-submit-btn"
                                >Hozzáadás</button
                            >
                        </form>
                    </details>

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_mondasok"
                                name="search_mondasok"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchMondasok}
                                placeholder="Szöveg vagy ID…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Megjelenés napja</th>
                                    <th>Szöveg</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(mondasok, searchMondasok, (m) => [m.id, m.text, m.display_date]) as m}
                                    <tr>
                                        <td>{m.id}</td>
                                        <td>{m.display_date ?? "—"}</td>
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
                                        ><td colspan="5">Nincsenek idézetek.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Quick Links Tab -->
                {#if activeTab === "quicklinks"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Gyorslinkek a kezdőlaphoz: cím, URL és opcionális háttérszín. A kártyák a főoldalon
                            jelennek meg.
                        </p>
                    {/if}
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"
                            >Új gyorslink hozzáadása</summary
                        >
                        <form class="admin-form admin-create-form" on:submit={submitLink}>
                            <label for="link_title">Cím</label>
                            <input
                                id="link_title"
                                name="title"
                                type="text"
                                bind:value={newLink.title}
                                required
                            />

                            <label for="link_url">URL</label>
                            <input
                                id="link_url"
                                name="url"
                                type="url"
                                bind:value={newLink.url}
                                required
                            />

                            <label for="link_color">Háttérszín (pl. #e6f0ff)</label>
                            <input
                                id="link_color"
                                name="bg_color"
                                type="text"
                                bind:value={newLink.bg_color}
                                placeholder="#e6f0ff"
                            />

                            <button type="submit" class="admin-submit-btn"
                                >Hozzáadás</button
                            >
                        </form>
                    </details>

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_quick_links"
                                name="search_quick_links"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchQuickLinks}
                                placeholder="Cím, URL…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Szín</th>
                                    <th>Cím</th>
                                    <th>URL</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(quickLinks, searchQuickLinks, (q) => [q.title, q.url, q.bg_color]) as q}
                                    <tr>
                                        <td>
                                            <span
                                                class="color-swatch"
                                                style:background={q.bg_color}
                                            ></span>
                                        </td>
                                        <td>{q.title}</td>
                                        <td class="admin-table-cell-preview" title={q.url || ""}>
                                            <a href={q.url} target="_blank" rel="nofollow noopener">{urlPreview(q.url)}</a>
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
                                        ><td colspan="5"
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
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            RSS / Atom hírfolyamok: a <strong>Hírek</strong> oldal ezekből gyűjti a cikkeket.
                            Utolsó frissítés időpontja és egyedi szín is beállítható.
                        </p>
                    {/if}
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"
                            >Új RSS hírfolyam hozzáadása</summary
                        >
                        <form class="admin-form admin-create-form" on:submit={submitNews}>
                            <label for="news_title">Hírportál neve</label>
                            <input
                                id="news_title"
                                name="title"
                                type="text"
                                bind:value={newNews.title}
                                required
                            />

                            <label for="news_url">RSS URL</label>
                            <input
                                id="news_url"
                                name="feed_url"
                                type="url"
                                bind:value={newNews.feed_url}
                                required
                            />

                            <label for="news_color">Háttérszín (pl. #ffebd6)</label>
                            <input
                                id="news_color"
                                name="bg_color"
                                type="text"
                                bind:value={newNews.bg_color}
                                placeholder="#ffebd6"
                            />

                            <button type="submit" class="admin-submit-btn"
                                >Hozzáadás</button
                            >
                        </form>
                    </details>

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_news_feeds"
                                name="search_news_feeds"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchNewsFeeds}
                                placeholder="Név, URL…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Név</th>
                                    <th>Forrás</th>
                                    <th>Utolsó frissítés</th>
                                    <th>Szín</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(newsFeeds, searchNewsFeeds, (nf) => [nf.title, nf.feed_url, String(nf.id)]) as nf}
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
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Települések (falu, város, község, municípium): név, megye, típus, irányítószám,
                            koordináták és kapcsolódó adatok. Az egész oldal ezekre az azonosítókra épül.
                        </p>
                    {/if}
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary">Új település</summary>
                        <form class="admin-form admin-create-form" on:submit={submitLocation}>
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
                            {#each settlementsForSelect as loc}
                                <option value={loc.id}
                                    >{loc.name} ({loc.county})</option
                                >
                            {/each}
                        </select>

                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>
                    </details>

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_locations"
                                name="search_locations"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchLocations}
                                placeholder="Név, megye, típus, ir.sz…"
                            /></label
                        >
                    </div>
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
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(locations, searchLocations, (l) => [
                                        l.id,
                                        l.name,
                                        l.name_ro,
                                        l.name_de,
                                        l.county,
                                        l.type,
                                        l.post_code,
                                        l.coordinates,
                                        l.population,
                                        l.area,
                                        getLocationName(l.parent_id),
                                    ]) as l}
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
                                        <td class="admin-table-cell-preview" title={l.crest || ""}>{l.crest ? urlPreview(l.crest) : "—"}</td>
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

                <!-- Venues Tab -->
                {#if activeTab === "venues"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Rendezvényhelyszínek (csarnokok, terek, pályák) településhez kötve. Felül a
                            <strong>helyszíntípusok</strong> (a slug a megnevezésből képződik, megjelenített név, sorrend) — ezek
                            szerepelnek a listákban és a nyilvános helyszín-oldalakon. Alatta a konkrét
                            helyszínek; az <strong>Események</strong> napi programjában itt választhatók.
                        </p>
                    {/if}

                    <h3 class="admin-subsection-title">Helyszín típusok</h3>
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"
                            >Új helyszíntípus</summary
                        >
                        <form
                            class="admin-form admin-create-form"
                            on:submit|preventDefault={submitNewVenueType}
                        >
                            <label for="vt-label">Megnevezés (HU) *</label>
                            <input
                                id="vt-label"
                                type="text"
                                bind:value={newVenueType.label_hu}
                                required
                            />
                            <button type="submit" class="admin-submit-btn"
                                >Típus hozzáadása</button
                            >
                        </form>
                    </details>

                    {#if editingVenueType}
                        <form
                            class="admin-form admin-venues-type-edit"
                            on:submit|preventDefault={saveEditVenueType}
                        >
                            <p class="admin-form-hint">
                                Slug (automatikusan a megnevezésből; mentéskor frissül, és a hozzá tartozó
                                helyszínek <code>kind</code> mezője is ehhez igazodik):
                                <code>{editingVenueType.slug}</code>
                            </p>
                            <label for="vt-edit-label">Megnevezés (HU)</label>
                            <input
                                id="vt-edit-label"
                                type="text"
                                bind:value={editingVenueType.label_hu}
                                required
                            />
                            <div class="flex gap-md">
                                <button type="submit" class="admin-submit-btn"
                                    >Mentés</button
                                >
                                <button
                                    type="button"
                                    class="btn-update"
                                    on:click={cancelEditVenueType}>Mégse</button
                                >
                            </div>
                        </form>
                    {/if}

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés (típusok)
                            <input
                                id="search_venue_types"
                                name="search_venue_types"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchVenueTypes}
                                placeholder="Slug, megnevezés…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table admin-table--compact">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Slug</th>
                                    <th>Megnevezés (HU)</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(venueTypesList, searchVenueTypes, (t) => [t.id, t.slug, t.label_hu]) as t}
                                    <tr>
                                        <td>{t.id}</td>
                                        <td><code>{t.slug}</code></td>
                                        <td>{t.label_hu}</td>
                                        <td>
                                            <button
                                                type="button"
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditVenueType(t)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                type="button"
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteVenueTypeRow(t.id)}
                                                >Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="5"
                                            >Nincs típus (futtasd a migrációt).</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>

                    <h3 class="admin-subsection-title" style="margin-top:2rem"
                        >Helyszínek</h3
                    >
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"
                            >Új helyszín hozzáadása</summary
                        >
                    <form
                        class="admin-form admin-venues-form admin-create-form"
                        on:submit|preventDefault={submitNewVenue}
                    >
                        <div class="flex gap-lg flex-wrap">
                            <label class="flex-1" style="min-width:12rem"
                                >Település (város / falu)
                                <select
                                    bind:value={newVenue.settlement_id}
                                    required
                                >
                                    <option value="">Válassz...</option>
                                    {#each settlementsForSelect as loc}
                                        <option value={loc.id}
                                            >{loc.name} ({loc.county})</option
                                        >
                                    {/each}
                                </select>
                            </label>
                            <label class="flex-1" style="min-width:12rem"
                                >Név (HU) *
                                <input
                                    type="text"
                                    bind:value={newVenue.name}
                                    required
                                    placeholder="pl. Deme László Műjégpálya"
                                />
                            </label>
                        </div>
                        <div class="flex gap-lg flex-wrap">
                            <label
                                class="flex-1"
                                style="min-width:10rem"
                                for="new-venue-name-ro"
                                >Név (RO)
                                <input
                                    id="new-venue-name-ro"
                                    type="text"
                                    bind:value={newVenue.name_ro}
                                    placeholder="opcionális"
                                />
                            </label>
                            <label
                                class="flex-1"
                                style="min-width:10rem"
                                for="new-venue-name-de"
                                >Név (DE)
                                <input
                                    id="new-venue-name-de"
                                    type="text"
                                    bind:value={newVenue.name_de}
                                    placeholder="opcionális"
                                />
                            </label>
                        </div>
                        <div class="flex gap-lg flex-wrap">
                            <label class="flex-1" style="min-width:8rem"
                                >Slug (opcionális)
                                <input
                                    type="text"
                                    bind:value={newVenue.slug}
                                    placeholder="auto, ha üres"
                                />
                            </label>
                            <label class="flex-1" style="min-width:10rem"
                                >Típus
                                <select bind:value={newVenue.kind}>
                                    {#each venueTypesList as vt}
                                        <option value={vt.slug}
                                            >{vt.label_hu}</option
                                        >
                                    {/each}
                                </select>
                            </label>
                        </div>
                        <label for="new-venue-address">Cím</label>
                        <input
                            id="new-venue-address"
                            type="text"
                            bind:value={newVenue.address}
                            placeholder="Utca, házszám"
                        />
                        <div class="flex gap-lg flex-wrap">
                            <label class="flex-1" style="min-width:8rem"
                                >Szélesség (lat)
                                <input
                                    type="text"
                                    bind:value={newVenue.latitude}
                                    placeholder="pl. 46.1234"
                                />
                            </label>
                            <label class="flex-1" style="min-width:8rem"
                                >Hosszúság (lon)
                                <input
                                    type="text"
                                    bind:value={newVenue.longitude}
                                    placeholder="pl. 25.5678"
                                />
                            </label>
                            <label class="flex-1" style="min-width:8rem"
                                >Férőhely
                                <input
                                    type="text"
                                    bind:value={newVenue.seating_capacity}
                                    placeholder="ülőhely / kapacitás"
                                />
                            </label>
                        </div>
                        <label for="new-venue-description">Leírás</label>
                        <textarea
                            id="new-venue-description"
                            bind:value={newVenue.description}
                            rows="3"
                        ></textarea>
                        <label for="new-venue-notes">Belső megjegyzés</label>
                        <textarea
                            id="new-venue-notes"
                            bind:value={newVenue.notes}
                            rows="2"
                        ></textarea>
                        <button type="submit" class="admin-submit-btn"
                            >Helyszín hozzáadása</button
                        >
                    </form>
                    </details>

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés (helyszínek)
                            <input
                                id="search_venues"
                                name="search_venues"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchVenues}
                                placeholder="Név, település, típus…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table admin-table--compact">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Település</th>
                                    <th>Név (HU)</th>
                                    <th>Név (RO)</th>
                                    <th>Név (DE)</th>
                                    <th>Slug</th>
                                    <th>Típus</th>
                                    <th>Cím</th>
                                    <th>Koordináták</th>
                                    <th>Férőhely</th>
                                    <th>Leírás</th>
                                    <th>Belső megjegyzés</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(venuesCatalog, searchVenues, (v) => [
                                        v.id,
                                        v.name,
                                        v.name_ro,
                                        v.name_de,
                                        v.settlement_name,
                                        v.county_name,
                                        v.kind,
                                        v.kind_label,
                                        v.slug,
                                        v.address,
                                        v.description,
                                        v.notes,
                                    ]) as v}
                                    <tr>
                                        <td>{v.id}</td>
                                        <td
                                            >{v.settlement_name}, {v.county_name}</td
                                        >
                                        <td>{v.name}</td>
                                        <td>{v.name_ro || "—"}</td>
                                        <td>{v.name_de || "—"}</td>
                                        <td><code>{v.slug || "—"}</code></td>
                                        <td
                                            >{v.kind_label || v.kind}</td
                                        >
                                        <td class="admin-table-cell-preview" title={v.address || ""}>{contentPreview(v.address)}</td>
                                        <td
                                            >{#if v.latitude != null && v.longitude != null}{Number(
                                                    v.latitude,
                                                ).toFixed(4)}, {Number(
                                                    v.longitude,
                                                ).toFixed(4)}{:else}—{/if}</td
                                        >
                                        <td>{v.seating_capacity ?? "—"}</td>
                                        <td class="admin-table-cell-preview" title={v.description || ""}>{contentPreview(v.description)}</td>
                                        <td class="admin-table-cell-preview" title={v.notes || ""}>{contentPreview(v.notes)}</td>
                                        <td>
                                            <button
                                                type="button"
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditVenue(v)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                type="button"
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteVenueRow(v.id)}
                                                >Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="14"
                                            >Nincs még helyszín.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Events Tab -->
                {#if activeTab === "events"}
                    {#if eventsWithIncompleteDateTime.length > 0}
                        <div class="admin-alert admin-alert--warning" role="alert">
                            <strong>Hiányos esemény-időpontok.</strong>
                            {eventsWithIncompleteDateTime.length} eseménynél nincs meg minden kötelező mező
                            (kezdő/befejező dátum és óra:perc). Szerkeszd a listában a ⚠ jelű sorokat, és töltsd
                            ki a mezőket.
                        </div>
                    {:else if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Közösségi és sportesemények: település, opcionális kiválasztott helyszín, típus,
                            szervező, leírás. A kezdő és befejező <strong>dátum és időpont (óra:perc)</strong> mind
                            kötelező — a mentés és a nyilvános időjelzések ettől függnek. Opcionálisan
                            <strong>napi program</strong> (több nap, helyszínenkénti tételek) adható meg a szerkesztőben.
                        </p>
                    {/if}
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary">Új esemény</summary>
                        <p class="admin-form-hint">
                            A <strong>kezdő és befejező dátum</strong> és a hozzájuk tartozó
                            <strong>időpontok (óra:perc)</strong> mind kötelezőek — a mentés nélkülük nem lehetséges.
                        </p>
                    <form class="admin-form admin-create-form" on:submit={submitEvent}>
                        <label for="event_loc"
                            >Település / Helyszín <span class="admin-req" title="Kötelező"
                                >*</span
                            ></label
                        >
                        <select
                            id="event_loc"
                            bind:value={newEvent.location_id}
                            required
                            on:change={loadVenuesForNewEvent}
                        >
                            <option value="">Válassz...</option>
                            {#each settlementsForSelect as loc}
                                <option value={loc.id}
                                    >{loc.name} ({loc.county})</option
                                >
                            {/each}
                        </select>

                        <label for="event_default_venue"
                            >Konkrét helyszín (opcionális)</label
                        >
                        <select
                            id="event_default_venue"
                            name="default_venue_id"
                            bind:value={newEvent.default_venue_id}
                        >
                            <option value="">— nincs megadva —</option>
                            {#each venueOptionsNew as v}
                                <option value={String(v.id)}>{v.name}</option>
                            {/each}
                        </select>

                        <label for="event_title"
                            >Esemény neve <span class="admin-req" title="Kötelező">*</span
                            ></label
                        >
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
                                <label for="event_start_date"
                                    >Kezdő dátum <span class="admin-req" title="Kötelező"
                                        >*</span
                                    ></label
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
                                    >Kezdő időpont (óra:perc) <span
                                        class="admin-req"
                                        title="Kötelező">*</span
                                    ></label
                                >
                                <input
                                    id="event_start_time"
                                    type="time"
                                    bind:value={newEvent.start_time}
                                    required
                                />
                            </div>
                        </div>

                        <div class="flex gap-lg">
                            <div class="flex-1">
                                <label for="event_end_date"
                                    >Befejező dátum <span class="admin-req" title="Kötelező"
                                        >*</span
                                    ></label
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
                                    >Befejező időpont (óra:perc) <span
                                        class="admin-req"
                                        title="Kötelező">*</span
                                    ></label
                                >
                                <input
                                    id="event_end_time"
                                    type="time"
                                    bind:value={newEvent.end_time}
                                    required
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
                        <div class="org-autosuggest-wrapper">
                            <div class="org-autosuggest-row">
                                <input
                                    id="event_org"
                                    type="text"
                                    bind:value={orgQuery}
                                    on:input={() => {
                                        newEvent.organizer = orgQuery;
                                        onOrgInput(false);
                                    }}
                                    on:focus={() => onOrgInput(false)}
                                    on:blur={() => handleOrgBlur(false)}
                                    autocomplete="off"
                                    placeholder="Keresés szervező neve..."
                                    class="flex-1"
                                />
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
                            {#if orgDropdownOpen && orgSuggestions.length > 0}
                                <ul class="org-suggestions">
                                    {#each orgSuggestions as s}
                                        <li>
                                            <button type="button" on:click={() => selectOrganizer(s.name, false)}>
                                                <strong>{s.name}</strong>
                                                {#if s.location}<span class="org-sug-meta">{s.location}</span>{/if}
                                            </button>
                                        </li>
                                    {/each}
                                </ul>
                            {/if}
                        </div>

                        <button type="submit" class="admin-submit-btn"
                            >Hozzáadás</button
                        >
                    </form>
                    </details>

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_events"
                                name="search_events"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchEvents}
                                placeholder="Cím, szervező, helyszín…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Cím</th>
                                    <th>Kezdés</th>
                                    <th>Befejezés</th>
                                    <th>Típus</th>
                                    <th>Település</th>
                                    <th>Alapért. helyszín</th>
                                    <th>Szervező</th>
                                    <th>Leírás</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(events, searchEvents, (e) => [
                                        e.title,
                                        e.description,
                                        e.organizer,
                                        getLocationName(e.location_id),
                                        e.default_venue_name,
                                        e.event_type,
                                        e.start_date,
                                        e.end_date,
                                    ]) as e}
                                    <tr
                                        class:admin-row-warn={!eventDateTimeComplete(
                                            e,
                                        )}
                                    >
                                        <td>
                                            {#if !eventDateTimeComplete(e)}
                                                <span
                                                    class="admin-req"
                                                    title="Hiányos dátum vagy időpont — szerkessze és töltse ki."
                                                    >⚠</span
                                                >
                                            {/if}
                                            {e.title}
                                        </td>
                                        <td>
                                            {new Date(
                                                e.start_date,
                                            ).toLocaleDateString("hu-HU")}
                                            {#if e.start_time}
                                                {e.start_time.slice(0, 5)}{/if}
                                        </td>
                                        <td>
                                            {#if e.end_date}
                                                {new Date(
                                                    e.end_date,
                                                ).toLocaleDateString("hu-HU")}
                                                {#if e.end_time}
                                                    {e.end_time.slice(0, 5)}{/if}
                                            {:else}
                                                -
                                            {/if}
                                        </td>
                                        <td>{({cultural: "Kulturális", sports: "Sport", festival: "Fesztivál", religious: "Vallási", other: "Egyéb"})[e.event_type] || e.event_type}</td>
                                        <td>{getLocationName(e.location_id)}</td>
                                        <td class="admin-table-cell-preview" title={e.default_venue_name || ""}>{contentPreview(e.default_venue_name || "")}</td>
                                        <td>{e.organizer || "—"}</td>
                                        <td class="admin-table-cell-preview" title={e.description || ""}>{contentPreview(e.description)}</td>
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
                                        ><td colspan="10"
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
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Bejegyzés-kategóriák (pl. szolgáltatás típusok): a településoldali és index
                            bejegyzések csoportosításához.
                        </p>
                    {/if}
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary">Új kategória</summary>
                        <form class="admin-form admin-create-form" on:submit={submitEntryCategory}>
                            <label for="cat_name">Kategória neve</label>
                            <input
                                id="cat_name"
                                name="name"
                                type="text"
                                bind:value={newEntryCategory.name}
                                required
                            />

                            <button type="submit" class="admin-submit-btn"
                                >Hozzáadás</button
                            >
                        </form>
                    </details>

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_entry_categories"
                                name="search_entry_categories"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchEntryCategories}
                                placeholder="Név, ID…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Név</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(entryCategories, searchEntryCategories, (cat) => [cat.id, cat.name]) as cat}
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
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Településhez kötött bejegyzések (üzletek, szervezetek, szolgáltatások): típus,
                            kategória, elérhetőség, nyelvek és címkék. Ezek a város/falu oldalakon és indexeken
                            jelennek meg.
                        </p>
                    {/if}
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary">Új bejegyzés</summary>
                    <form class="admin-form admin-create-form" on:submit={submitEntry}>
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
                            {#each settlementsForSelect as loc}
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
                    </details>

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_entries"
                                name="search_entries"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchEntries}
                                placeholder="Név, URL, címke…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Név</th>
                                    <th>Típus</th>
                                    <th>URL</th>
                                    <th>Település</th>
                                    <th>Kategória</th>
                                    <th>Telefon</th>
                                    <th>Cím</th>
                                    <th>Megjegyzés</th>
                                    <th>Nyelvek</th>
                                    <th>Címkék</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(entries, searchEntries, (s) => [
                                        s.name,
                                        s.type,
                                        s.url,
                                        s.phone,
                                        s.address,
                                        s.notes,
                                        getLocationName(s.location_id),
                                        getCategoryName(s.category_id),
                                        (s.languages || []).join(","),
                                        (s.tags || []).join(","),
                                    ]) as s}
                                    <tr>
                                        <td>{s.name}</td>
                                        <td
                                            ><span class="badge"
                                                >{s.type || "entry"}</span
                                            ></td
                                        >
                                        <td class="admin-table-cell-preview" title={s.url || ""}>{s.url ? urlPreview(s.url) : "—"}</td>
                                        <td>{getLocationName(s.location_id)}</td
                                        >
                                        <td>{getCategoryName(s.category_id)}</td
                                        >
                                        <td>{s.phone || "—"}</td>
                                        <td class="admin-table-cell-preview" title={s.address || ""}>{contentPreview(s.address)}</td>
                                        <td class="admin-table-cell-preview" title={s.notes || ""}>{contentPreview(s.notes)}</td>
                                        <td>{(s.languages || []).join(", ")}</td
                                        >
                                        <td>
                                            {#if s.tags && s.tags.length > 0}
                                                <div class="admin-table-tags">
                                                    {s.tags
                                                        .map((t) => "#" + t)
                                                        .join(" ")}
                                                </div>
                                            {:else}—{/if}
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
                                        ><td colspan="12"
                                            >Nincsenek bejegyzések.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Beállítások (Settings) Tab -->
                {#if activeTab === "settings"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Oldalszintű beállítások: alapértelmezett település (kezdőlap időjárás, eseményszűrés),
                            időjárás-szolgáltatók engedélyezése, ikon stílus, cache TTL és látogató-becslés.
                            A <strong>cache törlése</strong> új verziószámot ad — a látogatók frissebb időjárást kapnak.
                        </p>
                    {/if}
                    <section class="admin-form-section">
                        <h3>Alapértelmezett település (MyLocation)</h3>
                        <p class="admin-hint">A kezdőlap időjárás widgetje és az események szűrése ezt a települést használja alapértelmezettként.</p>
                        <div class="admin-form" style="max-width: 32rem;">
                            <label for="my_location_slug">Település</label>
                            <select id="my_location_slug" name="my_location_slug" bind:value={siteSettings.my_location_slug}>
                                {#each settlementsForSelect as loc}
                                    <option value={loc.slug}>{loc.name}{loc.county ? ` (${loc.county})` : ''}{loc.type ? ` – ${loc.type}` : ''}</option>
                                {/each}
                            </select>
                            <div class="flex gap-md mt-md">
                                <button type="button" class="admin-submit-btn" on:click={saveSettings} disabled={settingsSaving}>
                                    {settingsSaving ? 'Mentés…' : 'Mentés'}
                                </button>
                            </div>
                        </div>
                    </section>

                    <section class="admin-form-section">
                        <h3>Időjárás (Weather)</h3>
                        <div class="admin-form" style="max-width: 32rem;">
                            <label for="weather_provider_default">Alapértelmezett szolgáltató</label>
                            <select id="weather_provider_default" name="weather_provider_default" bind:value={siteSettings.weather_provider_default}>
                                <option value="open_meteo">Open-Meteo</option>
                                <option value="weatherapi_com">WeatherAPI.com</option>
                                <option value="openweathermap">OpenWeatherMap</option>
                            </select>

                            <span class="form-group-label">Szolgáltatók engedélyezése</span>
                            <div class="flex gap-lg flex-wrap mb-lg">
                                <label class="flex items-center gap-xs font-normal">
                                    <input id="weather_provider_open_meteo_enabled" name="weather_provider_open_meteo_enabled" type="checkbox" checked={siteSettings.weather_provider_open_meteo_enabled === 'true'} on:change={(e) => siteSettings.weather_provider_open_meteo_enabled = e.target.checked ? 'true' : 'false'} class="w-auto" />
                                    Open-Meteo
                                </label>
                                <label class="flex items-center gap-xs font-normal">
                                    <input id="weather_provider_weatherapi_enabled" name="weather_provider_weatherapi_enabled" type="checkbox" checked={siteSettings.weather_provider_weatherapi_enabled === 'true'} on:change={(e) => siteSettings.weather_provider_weatherapi_enabled = e.target.checked ? 'true' : 'false'} class="w-auto" />
                                    WeatherAPI.com
                                </label>
                                <label class="flex items-center gap-xs font-normal">
                                    <input id="weather_provider_openweathermap_enabled" name="weather_provider_openweathermap_enabled" type="checkbox" checked={siteSettings.weather_provider_openweathermap_enabled === 'true'} on:change={(e) => siteSettings.weather_provider_openweathermap_enabled = e.target.checked ? 'true' : 'false'} class="w-auto" />
                                    OpenWeatherMap
                                </label>
                            </div>

                            <label for="weather_icon_style">Időjárás ikon stílus</label>
                            <select id="weather_icon_style" name="weather_icon_style" bind:value={siteSettings.weather_icon_style}>
                                <option value="emoji">Emoji</option>
                                <option value="svg">SVG (saját ikonok)</option>
                            </select>

                            <label for="weather_cache_ttl">Időjárás cache TTL (perc)</label>
                            <input id="weather_cache_ttl" name="weather_cache_ttl_minutes" type="number" min="1" max="1440" bind:value={siteSettings.weather_cache_ttl_minutes} />

                            <label for="weather_active_users">Aktív felhasználók becslése</label>
                            <input id="weather_active_users" name="weather_active_users_estimate" type="number" min="1" bind:value={siteSettings.weather_active_users_estimate} />

                            <div class="flex gap-md mt-md flex-wrap">
                                <button type="button" class="admin-submit-btn" on:click={saveSettings} disabled={settingsSaving}>
                                    {settingsSaving ? 'Mentés…' : 'Mentés'}
                                </button>
                                <button type="button" class="btn-update" on:click={clearWeatherCache} disabled={settingsCacheClearing}>
                                    {settingsCacheClearing ? '…' : 'Időjárás cache törlése'}
                                </button>
                            </div>
                        </div>
                    </section>
                {/if}

                <!-- Időjárás fordítások (Weather translations) Tab -->
                {#if activeTab === "weather_translations"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Az időjárás API angol (vagy más) szövegeinek fordítása (hu, ro, de). Ha nincs egyedi sor,
                            a rendszer az alapértelmezett magyar megnevezést használja. Új sor: eredeti szöveg =
                            pontos egyezés a bejövő API szöveggel.
                        </p>
                    {/if}
                    {#if editingWeatherTrans}
                        <details class="admin-create-panel" open>
                            <summary class="admin-create-summary">Fordítás szerkesztése</summary>
                            <form class="admin-form admin-create-form" on:submit={saveWeatherTranslation} style="max-width: 28rem;">
                                <label for="wet_src">Eredeti szöveg (pl. API angol)</label>
                                <input id="wet_src" name="source_text" type="text" bind:value={editingWeatherTrans.source_text} required />
                                <label for="wet_lang">Nyelv</label>
                                <select id="wet_lang" name="lang" bind:value={editingWeatherTrans.lang}>
                                    {#each WEATHER_TRANS_LANGS as opt}
                                        <option value={opt.value}>{opt.label}</option>
                                    {/each}
                                </select>
                                <label for="wet_txt">Lefordított szöveg</label>
                                <input id="wet_txt" name="translated_text" type="text" bind:value={editingWeatherTrans.translated_text} required />
                                <div class="flex gap-md mt-md">
                                    <button type="submit" class="admin-submit-btn">Mentés</button>
                                    <button type="button" class="btn-update" on:click={cancelEditWeatherTrans}>Mégse</button>
                                </div>
                            </form>
                        </details>
                    {:else}
                        <details class="admin-create-panel">
                            <summary class="admin-create-summary">Új fordítás</summary>
                        <form class="admin-form admin-create-form" on:submit={saveWeatherTranslation} style="max-width: 28rem;">
                            <label for="wt_src">Eredeti szöveg (pl. overcast, partly cloudy)</label>
                            <input id="wt_src" name="source_text" type="text" bind:value={newWeatherTrans.source_text} required placeholder="pl. overcast" />
                            <label for="wt_lang">Nyelv</label>
                            <select id="wt_lang" name="lang" bind:value={newWeatherTrans.lang}>
                                {#each WEATHER_TRANS_LANGS as opt}
                                    <option value={opt.value}>{opt.label}</option>
                                {/each}
                            </select>
                            <label for="wt_txt">Lefordított szöveg</label>
                            <input id="wt_txt" name="translated_text" type="text" bind:value={newWeatherTrans.translated_text} required placeholder="pl. borult" />
                            <button type="submit" class="admin-submit-btn">Hozzáadás</button>
                        </form>
                        </details>
                    {/if}
                    <div class="admin-table-toolbar mt-lg">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_weather_trans"
                                name="search_weather_trans"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchWeatherTrans}
                                placeholder="Szöveg, nyelv…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Eredeti</th>
                                    <th>Nyelv</th>
                                    <th>Fordítás</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(weatherTranslations, searchWeatherTrans, (wt) => [wt.source_text, wt.lang, wt.translated_text]) as wt}
                                    <tr>
                                        <td>{wt.source_text}</td>
                                        <td>{wt.lang}</td>
                                        <td>{wt.translated_text}</td>
                                        <td><button type="button" class="btn-update" on:click={() => startEditWeatherTrans(wt)}>Szerk.</button></td>
                                        <td><button type="button" class="btn-delete" on:click={() => deleteWeatherTranslation(wt.id)}>Törlés</button></td>
                                    </tr>
                                {:else}
                                    <tr><td colspan="5">Nincs egyéni fordítás. Az alapértelmezett magyar szavak érvényesek.</td></tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Oldalak (Pages) Tab -->
                {#if activeTab === "pages"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {/if}
                    {#if editingPage}
                        <h3>Oldal szerkesztése: {editingPage.title}</h3>
                        <form class="admin-form" on:submit|preventDefault={savePage} style="max-width: 48rem;">
                            <label for="page_title">Cím</label>
                            <input id="page_title" name="title" type="text" bind:value={editingPage.title} required />

                            <label for="page_content">Tartalom (HTML)</label>
                            <textarea id="page_content" name="content" bind:value={editingPage.content} rows="20" style="font-family: monospace; font-size: 0.85rem;"></textarea>

                            <div class="flex gap-md mt-md">
                                <button type="submit" class="admin-submit-btn" disabled={pageSaving}>
                                    {pageSaving ? 'Mentés…' : 'Mentés'}
                                </button>
                                <button type="button" class="btn-update" on:click={cancelEditPage}>Mégse</button>
                            </div>
                        </form>
                    {:else if editingPageFaq}
                        <h3>GYIK / disclaimer: {editingPageFaq.label_hu || editingPageFaq.section_key}</h3>
                        <p class="admin-info">
                            Kulcs: <code>{editingPageFaq.section_key}</code> — a nyilvános oldalon a
                            <code>PageFaqDisclaimer</code> ugyanazt a HTML-struktúrát használja (<code>.faq</code>,
                            <code>details.faq-item</code>, <code>#disclaimer</code>, <code>.note.info</code>). Minden
                            blokk egy külön kérdés / válasz pár.
                        </p>
                        <form class="admin-form" on:submit|preventDefault={savePageFaq} style="max-width: 52rem;">
                            <label for="pfaq_label">Megjelenített név (admin)</label>
                            <input id="pfaq_label" name="label_hu" type="text" bind:value={editingPageFaq.label_hu} />

                            <label for="pfaq_title">GYIK szekció címe (H2)</label>
                            <input id="pfaq_title" name="faq_title" type="text" bind:value={editingPageFaq.faq_title} placeholder="pl. Hogyan működik ez az oldal?" />

                            <div class="admin-faq-toolbar">
                                <span class="admin-faq-toolbar-label">Kérdések és válaszok</span>
                                <button type="button" class="btn-update btn-sm" on:click={addFaqItem}
                                    >+ Új kérdés</button
                                >
                            </div>

                            {#each editingPageFaq.faq_items || [] as item, i (i)}
                                <details class="admin-faq-pair" open>
                                    <summary>Kérdés {i + 1}</summary>
                                    <div class="admin-faq-pair-fields">
                                        <label for={"pfaq_q_" + i}>Kérdés (summary)</label>
                                        <input
                                            id={"pfaq_q_" + i}
                                            name={"faq_question_" + i}
                                            type="text"
                                            bind:value={editingPageFaq.faq_items[i].question}
                                            placeholder="Rövid kérdés"
                                        />
                                        <label for={"pfaq_a_" + i}>Válasz (Markdown)</label>
                                        <textarea
                                            id={"pfaq_a_" + i}
                                            name={"faq_answer_" + i}
                                            bind:value={editingPageFaq.faq_items[i].answer}
                                            rows="5"
                                            style="font-family: monospace; font-size: 0.85rem;"
                                            placeholder="Válasz szövege…"
                                        ></textarea>
                                        <button
                                            type="button"
                                            class="btn-delete btn-sm"
                                            on:click={() => removeFaqItem(i)}>Kérdés törlése</button
                                        >
                                    </div>
                                </details>
                            {:else}
                                <p class="admin-info">Még nincs kérdés — kattints az „Új kérdés” gombra.</p>
                            {/each}

                            <label for="pfaq_disc">Disclaimer (Markdown)</label>
                            <textarea id="pfaq_disc" bind:value={editingPageFaq.disclaimer_markdown} rows="8" style="font-family: monospace; font-size: 0.85rem;"></textarea>

                            <div class="flex gap-md mt-md">
                                <button type="submit" class="admin-submit-btn" disabled={pageFaqSaving}>
                                    {pageFaqSaving ? 'Mentés…' : 'Mentés'}
                                </button>
                                <button type="button" class="btn-update" on:click={cancelEditPageFaq}>Mégse</button>
                            </div>
                        </form>
                    {:else}
                        {#if !adminTabError}
                            <p class="admin-info">
                                Statikus HTML oldalak (pl. irányelvek), valamint oldalankénti <strong>GYIK és disclaimer</strong>
                                szekciók. A GYIK a nyilvános oldalon a <code>PageFaqDisclaimer</code> komponensen keresztül
                                jelenik meg (kérdés–válasz párok, disclaimer).
                            </p>
                        {/if}
                        <h3 class="admin-subtab-heading">Irányelvek és statikus oldalak</h3>
                        <div class="admin-table-toolbar">
                            <label class="admin-search-label"
                                >Keresés (oldalak)
                                <input
                                    id="search_admin_pages"
                                    name="search_admin_pages"
                                    type="search"
                                    class="admin-search-input"
                                    bind:value={searchAdminPages}
                                    placeholder="Slug, cím…"
                                /></label
                            >
                        </div>
                        <div class="admin-table-wrapper">
                            <table class="admin-table">
                                <thead>
                                    <tr>
                                        <th>Slug</th>
                                        <th>Cím</th>
                                        <th>Utolsó módosítás</th>
                                        <th class="admin-table-col--action">Szerk.</th>
                                        <th class="admin-table-col--action">Törlés</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {#each filterRows(adminPages, searchAdminPages, (pg) => [pg.slug, pg.title, pg.updated_at]) as pg}
                                        <tr>
                                            <td><a href="/{pg.slug}" target="_blank">/{pg.slug}</a></td>
                                            <td>{pg.title}</td>
                                            <td>{pg.updated_at ? pg.updated_at.slice(0, 19) : ''}</td>
                                            <td class="admin-table-col--action"><button type="button" class="btn-update" on:click={() => startEditPage(pg)}>Szerk.</button></td>
                                            <td class="admin-table-col--action admin-table-col--action--muted">—</td>
                                        </tr>
                                    {:else}
                                        <tr
                                            ><td colspan="5"
                                                >{adminPages?.length
                                                    ? "Nincs találat a keresésre."
                                                    : "Nincsenek oldalak."}</td
                                            ></tr
                                        >
                                    {/each}
                                </tbody>
                            </table>
                        </div>

                        <h3 class="admin-subtab-heading">GYIK és felelősségkizárások (oldalanként)</h3>
                        <p class="admin-info">
                            Ugyanaz a kinézet, mint a <code>/hirek</code> oldalon: <code>.faq</code>,
                            <code>.faq-title</code>, <code>.faq-list</code>, <code>.faq-item</code>, <code>#disclaimer</code>,
                            <code>.note.info</code>.
                        </p>
                        <div class="admin-table-toolbar">
                            <label class="admin-search-label"
                                >Keresés (GYIK)
                                <input
                                    id="search_page_faq"
                                    name="search_page_faq"
                                    type="search"
                                    class="admin-search-input"
                                    bind:value={searchPageFaqRows}
                                    placeholder="Kulcs, név, cím…"
                                /></label
                            >
                        </div>
                        <div class="admin-table-wrapper">
                            <table class="admin-table">
                                <thead>
                                    <tr>
                                        <th>Kulcs</th>
                                        <th>Megjelenített név</th>
                                        <th>GYIK cím</th>
                                        <th>Kérdések</th>
                                        <th>Utolsó módosítás</th>
                                        <th class="admin-table-col--action">Szerk.</th>
                                        <th class="admin-table-col--action">Törlés</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {#each filterRows(pageFaqSections, searchPageFaqRows, (row) => [row.section_key, row.label_hu, row.faq_title, String((row.faq_items || []).length), row.updated_at]) as row}
                                        <tr>
                                            <td><code>{row.section_key}</code></td>
                                            <td>{row.label_hu}</td>
                                            <td class="admin-table-cell-preview" title={row.faq_title || ''}>{contentPreview(row.faq_title || "")}</td>
                                            <td>{(row.faq_items || []).length}</td>
                                            <td>{row.updated_at ? row.updated_at.slice(0, 19) : ''}</td>
                                            <td class="admin-table-col--action"><button type="button" class="btn-update" on:click={() => startEditPageFaq(row)}>Szerk.</button></td>
                                            <td class="admin-table-col--action admin-table-col--action--muted">—</td>
                                        </tr>
                                    {:else}
                                        <tr
                                            ><td colspan="7"
                                                >{pageFaqSections?.length
                                                    ? "Nincs találat a keresésre."
                                                    : "Nincs GYIK rekord (futtasd a backend migrációt)."}</td
                                            ></tr
                                        >
                                    {/each}
                                </tbody>
                            </table>
                        </div>
                    {/if}
                {/if}

                <!-- Entry Types Tab -->
                {#if activeTab === "entry_types"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Bejegyzés <strong>típusok</strong> (pl. entry, business): belső címkék a bejegyzések
                            szerkezetéhez és szűréséhez — nem ugyanaz, mint a kategória.
                        </p>
                    {/if}
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary">Új típus</summary>
                        <form class="admin-form admin-create-form" on:submit={submitEntryType}>
                            <label for="etype_name">Típus neve</label>
                            <input
                                id="etype_name"
                                name="name"
                                type="text"
                                bind:value={newEntryType.name}
                                required
                                placeholder="pl. entry, business..."
                            />
                            <button type="submit" class="admin-submit-btn"
                                >Hozzáadás</button
                            >
                        </form>
                    </details>

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_entry_types"
                                name="search_entry_types"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchEntryTypes}
                                placeholder="Név, ID…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Név</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(entryTypes, searchEntryTypes, (et) => [et.id, et.name]) as et}
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

                <!-- Attractions Tab -->
                {#if activeTab === "attractions"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Megyéhez kötött látnivalók (természet, kultúra): név, slug, rövid leírás, koordináták,
                            kiemelt kép és bővebb tartalom (Markdown). A megye és település oldalakon jelennek meg.
                        </p>
                    {/if}
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary">Új látnivaló</summary>
                    <form class="admin-form admin-create-form mb-lg" on:submit|preventDefault={submitNewAttraction}>
                        <div class="form-row">
                            <label for="att_county">Megye</label>
                            <select id="att_county" name="county_slug" bind:value={newAttraction.county_slug}>
                                <option value="hargita">Hargita</option>
                                <option value="kovaszna">Kovászna</option>
                                <option value="maros">Maros</option>
                            </select>
                        </div>
                        <div class="form-row">
                            <label for="att_name">Név</label>
                            <input id="att_name" name="name" type="text" bind:value={newAttraction.name} required placeholder="pl. Szent Anna-tó" />
                        </div>
                        <div class="form-row">
                            <label for="att_desc">Rövid leírás</label>
                            <input id="att_desc" name="description" type="text" bind:value={newAttraction.description} placeholder="Közép-Európa egyetlen vulkanikus tava..." />
                        </div>
                        <div class="form-row">
                            <label for="att_coords">Koordináták (lat, lon)</label>
                            <input id="att_coords" name="latitude" type="text" bind:value={newAttraction.latitude} placeholder="46.1265" style="width:6rem" />
                            <input id="att_lon" name="longitude" type="text" bind:value={newAttraction.longitude} placeholder="25.8876" style="width:6rem" />
                        </div>
                        <div class="form-row">
                            <label for="att_featured">Kiemelt kép URL</label>
                            <input id="att_featured" name="featured_image" type="url" bind:value={newAttraction.featured_image} placeholder="https://..." />
                        </div>
                        <div class="form-row">
                            <label for="att_content">Tartalom (Markdown)</label>
                            <textarea id="att_content" name="content" bind:value={newAttraction.content} rows="6" placeholder="## Cím&#10;Szöveg..."></textarea>
                        </div>
                        <div class="form-row">
                            <label for="att_images">Galéria URL-ek (soronként egy)</label>
                            <textarea id="att_images" name="images" bind:value={newAttraction.images} rows="3" placeholder="https://kep1.jpg&#10;https://kep2.jpg"></textarea>
                        </div>
                        <button type="submit" class="admin-submit-btn">Hozzáadás</button>
                    </form>
                    </details>
                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_attractions"
                                name="search_attractions"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchAttractions}
                                placeholder="Név, slug, megye…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Név</th>
                                    <th>Megye</th>
                                    <th>Slug</th>
                                    <th>Rövid leírás</th>
                                    <th>Koordináták</th>
                                    <th>Kiemelt kép</th>
                                    <th>Tartalom (előnézet)</th>
                                    <th>Galéria</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each filterRows(attractions, searchAttractions, (att) => [
                                        att.name,
                                        att.slug,
                                        att.county_name,
                                        att.description,
                                        att.featured_image,
                                        att.content,
                                        (att.images || []).join(" "),
                                        String(att.latitude ?? ""),
                                        String(att.longitude ?? ""),
                                    ]) as att}
                                    <tr>
                                        <td>{att.name}</td>
                                        <td>{att.county_name}</td>
                                        <td><code>{att.slug}</code></td>
                                        <td class="admin-table-cell-preview" title={att.description || ""}>{contentPreview(att.description)}</td>
                                        <td class="admin-table__mono">{formatLatLon(att.latitude, att.longitude)}</td>
                                        <td class="admin-table-cell-preview" title={att.featured_image || ""}>{urlPreview(att.featured_image)}</td>
                                        <td class="admin-table-cell-preview" title={att.content || ""}>{contentPreview(att.content)}</td>
                                        <td>{(att.images && att.images.length) || 0} kép</td>
                                        <td class="admin-table-col--action">
                                            <button type="button" class="btn-update" on:click={() => openEditAttraction(att)}>Szerk.</button>
                                        </td>
                                        <td class="admin-table-col--action">
                                            <button type="button" class="btn-delete" on:click={() => deleteAttraction(att.id)}>Törlés</button>
                                        </td>
                                    </tr>
                                {:else}
                                    <tr><td colspan="10">Nincsenek látnivalók.</td></tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Counties Tab -->
                {#if activeTab === "counties"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            <strong>Megyék:</strong> magyar / román / név név, URL-slug, bemutatkozó szöveg (Markdown),
                            és a <strong>megyeszékhely</strong> település kiválasztása a listából.
                            <strong>Történelmi székek</strong> (pl. Csíkszék): külön név, slug és tartalom —
                            a <code>/szekek</code> oldalakon jelennek meg.
                        </p>
                    {/if}

                    <h3 class="admin-region-heading">Megyék</h3>
                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés (megyék)
                            <input
                                id="search_counties"
                                name="search_counties"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchCounties}
                                placeholder="Név, slug, székhely…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Megye</th>
                                    <th>Név (RO)</th>
                                    <th>Név (DE)</th>
                                    <th>Slug</th>
                                    <th>Megyeszékhely</th>
                                    <th>Tartalom (előnézet)</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each displayCounties as c (c.id)}
                                    {#if editingCounty?.id === c.id}
                                        <tr class="admin-table-edit-row">
                                            <td colspan="8">
                                                <div class="admin-region-edit-panel">
                                                    <div class="admin-region-edit-grid">
                                                        <label for="county_edit_name">
                                                            Megye (HU)
                                                            <input id="county_edit_name" name="name" type="text" bind:value={editingCounty.name} />
                                                        </label>
                                                        <label for="county_edit_name_ro">
                                                            Név (RO)
                                                            <input id="county_edit_name_ro" name="name_ro" type="text" bind:value={editingCounty.name_ro} />
                                                        </label>
                                                        <label for="county_edit_name_de">
                                                            Név (DE)
                                                            <input id="county_edit_name_de" name="name_de" type="text" bind:value={editingCounty.name_de} />
                                                        </label>
                                                        <label for="county_edit_slug">
                                                            Slug (URL)
                                                            <input id="county_edit_slug" name="slug" type="text" bind:value={editingCounty.slug} placeholder="pl. hargita" />
                                                        </label>
                                                        <label class="admin-region-edit-span2" for="county_edit_seat_location_id">
                                                            Megyeszékhely
                                                            <select id="county_edit_seat_location_id" name="seat_location_id" bind:value={editingCounty.seat_location_id}>
                                                                <option value="">— válassz települést —</option>
                                                                {#each settlementsForCountyName(c.name) as loc (loc.id)}
                                                                    <option value={String(loc.id)}
                                                                        >{loc.name} ({loc.type}){loc.name_ro ? " — " + loc.name_ro : ""}</option
                                                                    >
                                                                {/each}
                                                            </select>
                                                        </label>
                                                    </div>
                                                    <label class="admin-region-edit-full" for="county_edit_content">
                                                        Bemutatkozás (Markdown)
                                                        <textarea
                                                            id="county_edit_content"
                                                            name="content"
                                                            rows="10"
                                                            bind:value={editingCounty.content}
                                                            placeholder="## Bevezető&#10;..."
                                                        ></textarea>
                                                    </label>
                                                    <div class="admin-region-edit-actions">
                                                        <button
                                                            type="button"
                                                            class="admin-submit-btn"
                                                            on:click={saveEditingCounty}>Mentés</button
                                                        >
                                                        <button
                                                            type="button"
                                                            class="btn-update"
                                                            on:click={cancelEditCounty}>Mégse</button
                                                        >
                                                    </div>
                                                </div>
                                            </td>
                                        </tr>
                                    {:else}
                                        <tr>
                                            <td><strong>{c.name}</strong></td>
                                            <td>{c.name_ro || "—"}</td>
                                            <td>{c.name_de || "—"}</td>
                                            <td><code>{c.slug}</code></td>
                                            <td>{countySeatDisplayName(c)}</td>
                                            <td class="admin-table-cell-preview" title={c.content || ""}
                                                >{contentPreview(c.content)}</td
                                            >
                                            <td class="admin-table-col--action">
                                                <button
                                                    type="button"
                                                    class="btn-update"
                                                    on:click={() => startEditCounty(c)}>Szerk.</button
                                                >
                                            </td>
                                            <td class="admin-table-col--action admin-table-col--action--muted">—</td>
                                        </tr>
                                    {/if}
                                {:else}
                                    <tr
                                        ><td colspan="8"
                                            >{countiesFromAPI?.length
                                                ? "Nincs találat a keresésre."
                                                : "Nincs megye-adat (API / migráció)."}</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>

                    <h3 class="admin-region-heading">Történelmi székek</h3>
                    <p class="admin-hint">
                        Megjelenés: <a href="/szekek" target="_blank" rel="noopener">/szekek</a> és
                        <code>/szekek/…</code> oldalak.
                    </p>
                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés (székek)
                            <input
                                id="search_historical_seats"
                                name="search_historical_seats"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchHistoricalSeats}
                                placeholder="Név, slug…"
                            /></label
                        >
                    </div>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Név (HU)</th>
                                    <th>Név (RO)</th>
                                    <th>Név (DE)</th>
                                    <th>Slug</th>
                                    <th>Tartalom (előnézet)</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each displayHistoricalSeats as h (h.id)}
                                    {#if editingHistoricalSeat?.id === h.id}
                                        <tr class="admin-table-edit-row">
                                            <td colspan="7">
                                                <div class="admin-region-edit-panel">
                                                    <div class="admin-region-edit-grid">
                                                        <label for="hseat_edit_name">
                                                            Név (HU)
                                                            <input id="hseat_edit_name" name="name" type="text" bind:value={editingHistoricalSeat.name} />
                                                        </label>
                                                        <label for="hseat_edit_name_ro">
                                                            Név (RO)
                                                            <input id="hseat_edit_name_ro" name="name_ro" type="text" bind:value={editingHistoricalSeat.name_ro} />
                                                        </label>
                                                        <label for="hseat_edit_name_de">
                                                            Név (DE)
                                                            <input id="hseat_edit_name_de" name="name_de" type="text" bind:value={editingHistoricalSeat.name_de} />
                                                        </label>
                                                        <label for="hseat_edit_slug">
                                                            Slug (URL)
                                                            <input id="hseat_edit_slug" name="slug" type="text" bind:value={editingHistoricalSeat.slug} placeholder="pl. csikszek" />
                                                        </label>
                                                    </div>
                                                    <label class="admin-region-edit-full" for="hseat_edit_content">
                                                        Tartalom (Markdown)
                                                        <textarea
                                                            id="hseat_edit_content"
                                                            name="content"
                                                            rows="10"
                                                            bind:value={editingHistoricalSeat.content}
                                                            placeholder="## …&#10;..."
                                                        ></textarea>
                                                    </label>
                                                    <div class="admin-region-edit-actions">
                                                        <button
                                                            type="button"
                                                            class="admin-submit-btn"
                                                            on:click={saveEditingHistoricalSeat}>Mentés</button
                                                        >
                                                        <button
                                                            type="button"
                                                            class="btn-update"
                                                            on:click={cancelEditHistoricalSeat}>Mégse</button
                                                        >
                                                    </div>
                                                </div>
                                            </td>
                                        </tr>
                                    {:else}
                                        <tr>
                                            <td><strong>{h.name}</strong></td>
                                            <td>{h.name_ro || "—"}</td>
                                            <td>{h.name_de || "—"}</td>
                                            <td><code>{h.slug}</code></td>
                                            <td class="admin-table-cell-preview" title={h.content || ""}
                                                >{contentPreview(h.content)}</td
                                            >
                                            <td class="admin-table-col--action">
                                                <button
                                                    type="button"
                                                    class="btn-update"
                                                    on:click={() => startEditHistoricalSeat(h)}>Szerk.</button
                                                >
                                            </td>
                                            <td class="admin-table-col--action admin-table-col--action--muted">—</td>
                                        </tr>
                                    {/if}
                                {:else}
                                    <tr
                                        ><td colspan="7"
                                            >{historicalSeatsFromAPI?.length
                                                ? "Nincs találat a keresésre."
                                                : "Nincs szék-adat (API / migráció)."}</td
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
                        name="mondas_text"
                        bind:value={editingMondas.text}
                        required
                        rows="4"
                        class="w-full"
                    ></textarea>
                    <label for="emondas_day_edit">Megjelenés napja</label>
                    <input
                        id="emondas_day_edit"
                        name="display_date"
                        type="date"
                        bind:value={editingMondas.display_date}
                        required
                        class="w-full"
                    />

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
                        {#each settlementsForSelect as loc}
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
                        {#each settlementsForSelect.filter((l) => l.id !== editingLocation.id) as loc}
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

    <!-- Edit Venue Modal -->
    {#if editingVenue}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={cancelEditVenue}
            on:keydown={(e) => e.key === "Escape" && cancelEditVenue()}
        >
            <div class="admin-modal">
                <h3>Helyszín szerkesztése</h3>
                <form
                    class="admin-form"
                    on:submit|preventDefault={saveEditVenue}
                >
                    <label for="ev-venue-settlement">Település</label>
                    <select
                        id="ev-venue-settlement"
                        bind:value={editingVenue.settlement_id}
                        required
                    >
                        {#each settlementsForSelect as loc}
                            <option value={String(loc.id)}
                                >{loc.name} ({loc.county})</option
                            >
                        {/each}
                    </select>

                    <label for="ev-venue-name">Név (HU)</label>
                    <input
                        id="ev-venue-name"
                        type="text"
                        bind:value={editingVenue.name}
                        required
                    />

                    <label for="ev-venue-name-ro">Név (RO)</label>
                    <input
                        id="ev-venue-name-ro"
                        type="text"
                        bind:value={editingVenue.name_ro}
                    />

                    <label for="ev-venue-name-de">Név (DE)</label>
                    <input
                        id="ev-venue-name-de"
                        type="text"
                        bind:value={editingVenue.name_de}
                    />

                    <div class="flex gap-lg flex-wrap">
                        <label class="flex-1" style="min-width:8rem"
                            >Slug
                            <input
                                type="text"
                                bind:value={editingVenue.slug}
                            />
                        </label>
                        <label class="flex-1" style="min-width:10rem"
                            >Típus
                            <select bind:value={editingVenue.kind}>
                                {#each venueTypesList as vt}
                                    <option value={vt.slug}>{vt.label_hu}</option>
                                {/each}
                            </select>
                        </label>
                    </div>

                    <label for="ev-venue-address">Cím</label>
                    <input
                        id="ev-venue-address"
                        type="text"
                        bind:value={editingVenue.address}
                    />

                    <div class="flex gap-lg flex-wrap">
                        <label class="flex-1" style="min-width:8rem"
                            >Szélesség (lat)
                            <input
                                type="text"
                                bind:value={editingVenue.latitude}
                            />
                        </label>
                        <label class="flex-1" style="min-width:8rem"
                            >Hosszúság (lon)
                            <input
                                type="text"
                                bind:value={editingVenue.longitude}
                            />
                        </label>
                        <label class="flex-1" style="min-width:8rem"
                            >Férőhely
                            <input
                                type="text"
                                bind:value={editingVenue.seating_capacity}
                            />
                        </label>
                    </div>

                    <label for="ev-venue-description">Leírás</label>
                    <textarea
                        id="ev-venue-description"
                        bind:value={editingVenue.description}
                        rows="4"
                    ></textarea>

                    <label for="ev-venue-notes">Belső megjegyzés</label>
                    <textarea
                        id="ev-venue-notes"
                        bind:value={editingVenue.notes}
                        rows="2"
                    ></textarea>

                    <div class="modal-actions">
                        <button type="submit" class="admin-submit-btn"
                            >Mentés</button
                        >
                        <button
                            type="button"
                            class="btn-delete"
                            on:click={cancelEditVenue}>Mégse</button
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

    <!-- Edit Attraction Modal -->
    {#if editingAttraction}
        <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
        <div
            class="admin-modal-overlay"
            role="dialog"
            tabindex="-1"
            on:click|self={cancelEditAttraction}
            on:keydown={(e) => e.key === "Escape" && cancelEditAttraction()}
        >
            <div class="admin-modal">
                <h3>Látnivaló szerkesztése</h3>
                <form
                    class="admin-form"
                    on:submit|preventDefault={saveEditAttraction}
                >
                    <label for="eatt_county">Megye</label>
                    <select id="eatt_county" name="county_slug" bind:value={editingAttraction.county_slug}>
                        <option value="hargita">Hargita</option>
                        <option value="kovaszna">Kovászna</option>
                        <option value="maros">Maros</option>
                    </select>
                    <label for="eatt_name">Név</label>
                    <input id="eatt_name" name="name" type="text" bind:value={editingAttraction.name} required />
                    <label for="eatt_desc">Rövid leírás</label>
                    <input id="eatt_desc" name="description" type="text" bind:value={editingAttraction.description} />
                    <label for="eatt_coords">Koordináták (lat, lon)</label>
                    <input id="eatt_coords" name="latitude" type="text" bind:value={editingAttraction.latitude} placeholder="46.1265" style="width:6rem" />
                    <input id="eatt_lon" name="longitude" type="text" bind:value={editingAttraction.longitude} placeholder="25.8876" style="width:6rem" />
                    <label for="eatt_featured">Kiemelt kép URL</label>
                    <input id="eatt_featured" name="featured_image" type="url" bind:value={editingAttraction.featured_image} />
                    <label for="eatt_content">Tartalom (Markdown)</label>
                    <textarea id="eatt_content" name="content" bind:value={editingAttraction.content} rows="6"></textarea>
                    <label for="eatt_images">Galéria URL-ek (soronként egy)</label>
                    <textarea id="eatt_images" name="images" bind:value={editingAttraction.images} rows="3"></textarea>
                    <div class="modal-actions">
                        <button type="submit" class="admin-submit-btn">Mentés</button>
                        <button type="button" class="btn-delete" on:click={cancelEditAttraction}>Mégse</button>
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
                <p class="admin-form-hint">
                    A dátumok és időpontok (óra:perc) kötelezőek.
                </p>
                <form
                    class="admin-form"
                    autocomplete="off"
                    on:submit|preventDefault={saveEditEvent}
                >
                    <label for="edit_ev_loc"
                        >Helyszín <span class="admin-req" title="Kötelező">*</span></label
                    >
                    <select
                        id="edit_ev_loc"
                        name="location_id"
                        bind:value={editingEvent.location_id}
                        required
                        on:change={() => {
                            editingEvent.default_venue_id = "";
                            loadVenuesForEditSettlement(editingEvent.location_id);
                        }}
                    >
                        {#each settlementsForSelect as loc}
                            <option value={loc.id}
                                >{loc.name} ({loc.county})</option
                            >
                        {/each}
                    </select>

                    <label for="edit_ev_default_venue"
                        >Konkrét helyszín (opcionális)</label
                    >
                    <select
                        id="edit_ev_default_venue"
                        name="default_venue_id"
                        bind:value={editingEvent.default_venue_id}
                    >
                        <option value="">— nincs megadva —</option>
                        {#each venueOptionsEdit as v}
                            <option value={String(v.id)}>{v.name}</option>
                        {/each}
                    </select>

                    <label for="edit_ev_title"
                        >Cím <span class="admin-req" title="Kötelező">*</span></label
                    >
                    <input
                        id="edit_ev_title"
                        name="title"
                        type="text"
                        bind:value={editingEvent.title}
                        required
                    />

                    <label for="edit_ev_desc">Leírás</label>
                    <textarea
                        id="edit_ev_desc"
                        name="description"
                        bind:value={editingEvent.description}
                    ></textarea>

                    <div class="flex gap-lg">
                        <div class="flex-1">
                            <label for="edit_ev_start_date"
                                >Kezdő dátum <span class="admin-req" title="Kötelező"
                                    >*</span
                                ></label
                            >
                            <input
                                id="edit_ev_start_date"
                                name="start_date"
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
                            <label for="edit_ev_start_time"
                                >Kezdő időpont (óra:perc) <span
                                    class="admin-req"
                                    title="Kötelező">*</span
                                ></label
                            >
                            <input
                                id="edit_ev_start_time"
                                name="start_time"
                                type="time"
                                bind:value={editingEvent.start_time}
                                required
                            />
                        </div>
                    </div>
                    <div class="flex gap-lg">
                        <div class="flex-1">
                            <label for="edit_ev_end_date"
                                >Befejező dátum <span class="admin-req" title="Kötelező"
                                    >*</span
                                ></label
                            >
                            <input
                                id="edit_ev_end_date"
                                name="end_date"
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
                                >Befejező időpont (óra:perc) <span
                                    class="admin-req"
                                    title="Kötelező">*</span
                                ></label
                            >
                            <input
                                id="edit_ev_end_time"
                                name="end_time"
                                type="time"
                                bind:value={editingEvent.end_time}
                                required
                            />
                        </div>
                    </div>

                    <label for="edit_ev_type">Típus</label>
                    <select
                        id="edit_ev_type"
                        name="event_type"
                        bind:value={editingEvent.event_type}
                    >
                        <option value="cultural">Kulturális</option>
                        <option value="sports">Sport</option>
                        <option value="festival">Fesztivál</option>
                        <option value="religious">Vallási</option>
                        <option value="other">Egyéb</option>
                    </select>

                    <label for="edit_ev_org">Szervező</label>
                    <div class="org-autosuggest-wrapper">
                        <div class="org-autosuggest-row">
                            <input
                                id="edit_ev_org"
                                name="organizer"
                                type="text"
                                bind:value={orgEditQuery}
                                on:input={() => {
                                    editingEvent.organizer = orgEditQuery;
                                    onOrgInput(true);
                                }}
                                on:focus={() => onOrgInput(true)}
                                on:blur={() => handleOrgBlur(true)}
                                autocomplete="off"
                                placeholder="Keresés szervező neve..."
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
                        {#if orgEditDropdownOpen && orgEditSuggestions.length > 0}
                            <ul class="org-suggestions">
                                {#each orgEditSuggestions as s}
                                    <li>
                                        <button type="button" on:click={() => selectOrganizer(s.name, true)}>
                                            <strong>{s.name}</strong>
                                            {#if s.location}<span class="org-sug-meta">{s.location}</span>{/if}
                                        </button>
                                    </li>
                                {/each}
                            </ul>
                        {/if}
                    </div>

                    <details class="admin-schedule-details">
                        <summary>Napi program (opcionális)</summary>
                        <p class="admin-form-hint admin-schedule-hint">
                            A fenti kezdő–befejező dátum és idő továbbra is az alap; ide
                            naponkénti tételeket írhat (megnyitó, mérkőzések, záró stb.).
                            A <strong>vége</strong> idő opcionális (pl. ismeretlen mérkőzés-hossz).
                            Üresen hagyható. A helyszínt soronként a <strong>Helyszín</strong> oszlopban
                            állíthatod (üres = esemény alaphelyszíne).
                        </p>
                        <div class="schedule-toolbar">
                            <button
                                type="button"
                                class="btn-update"
                                on:click|preventDefault={generateScheduleDaysFromEvent}
                                >Napok generálása a dátumokból</button
                            >
                            <button
                                type="button"
                                class="btn-update"
                                on:click|preventDefault={addScheduleDayRow}
                                >Új nap</button
                            >
                        </div>

                        {#each scheduleDraftDays as day, di}
                            <div class="schedule-day-block">
                                <div class="schedule-day-head">
                                    <label class="schedule-inline"
                                        >Dátum
                                        <input
                                            id={`schedule-day-${di}-date`}
                                            name={`schedule_day_${di}_date`}
                                            type="date"
                                            bind:value={day.schedule_date}
                                        /></label
                                    >
                                    <button
                                        type="button"
                                        class="btn-delete btn-xs"
                                        on:click={() =>
                                            removeScheduleDayRow(di)}
                                        >Nap törlése</button
                                    >
                                </div>
                                <table class="admin-table schedule-act-table">
                                    <thead>
                                        <tr>
                                            <th>Típus</th>
                                            <th title="Opcionális">Kezdés</th>
                                            <th title="Opcionális; mérkőzésnél gyakran üres"
                                                >Vége</th
                                            >
                                            <th title="Opcionális; üres = esemény alaphelyszíne"
                                                >Helyszín</th
                                            >
                                            <th>Cím / program</th>
                                            <th>Leírás</th>
                                            <th></th>
                                        </tr>
                                    </thead>
                                    <tbody>
                                        {#each day.activities || [] as act, ai}
                                            <tr>
                                                <td>
                                                    <select
                                                        id={`schedule-${di}-act-${ai}-type`}
                                                        name={`schedule_${di}_act_${ai}_type`}
                                                        class="schedule-act-type"
                                                        bind:value={act.activity_type}
                                                    >
                                                        {#each SCHEDULE_ACTIVITY_TYPES as t}
                                                            <option value={t}
                                                                >{SCHEDULE_ACTIVITY_TYPE_LABELS[
                                                                    t
                                                                ]}</option
                                                            >
                                                        {/each}
                                                    </select>
                                                </td>
                                                <td
                                                    ><input
                                                        id={`schedule-${di}-act-${ai}-start`}
                                                        name={`schedule_${di}_act_${ai}_starts_at`}
                                                        type="time"
                                                        bind:value={act.starts_at}
                                                /></td>
                                                <td
                                                    ><input
                                                        id={`schedule-${di}-act-${ai}-end`}
                                                        name={`schedule_${di}_act_${ai}_ends_at`}
                                                        type="time"
                                                        bind:value={act.ends_at}
                                                /></td>
                                                <td>
                                                    <select
                                                        id={`schedule-${di}-act-${ai}-venue`}
                                                        name={`schedule_${di}_act_${ai}_venue`}
                                                        class="schedule-act-venue"
                                                        bind:value={act.venue_id}
                                                    >
                                                        <option value=""
                                                            >— alapértelmezett —</option
                                                        >
                                                        {#each venueOptionsEdit as v}
                                                            <option
                                                                value={String(
                                                                    v.id,
                                                                )}
                                                                >{v.name}</option
                                                            >
                                                        {/each}
                                                    </select>
                                                </td>
                                                <td
                                                    ><input
                                                        id={`schedule-${di}-act-${ai}-title`}
                                                        name={`schedule_${di}_act_${ai}_title`}
                                                        type="text"
                                                        placeholder="Kötelező cím"
                                                        bind:value={act.title}
                                                /></td>
                                                <td
                                                    ><input
                                                        id={`schedule-${di}-act-${ai}-desc`}
                                                        name={`schedule_${di}_act_${ai}_description`}
                                                        type="text"
                                                        bind:value={act.description}
                                                /></td>
                                                <td
                                                    ><button
                                                        type="button"
                                                        class="btn-delete btn-xs"
                                                        on:click={() =>
                                                            removeScheduleActivity(
                                                                di,
                                                                ai,
                                                            )}>×</button
                                                    ></td
                                                >
                                            </tr>
                                        {/each}
                                    </tbody>
                                </table>
                                <button
                                    type="button"
                                    class="btn-update btn-xs"
                                    on:click={() => addScheduleActivity(di)}
                                    >+ Tevékenység</button
                                >
                            </div>
                        {/each}

                        <button
                            type="button"
                            class="admin-submit-btn schedule-save-btn"
                            on:click|preventDefault={saveEventSchedule}
                            >Napi program mentése</button
                        >
                    </details>

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
                        {#each settlementsForSelect as loc}
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
        width: min(1200px, 95vw);
        max-height: 90vh;
        overflow-y: auto;
    }
    .admin-modal h3 {
        margin-top: 0;
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
    .admin-info {
        color: var(--text-faint, #666);
        margin-bottom: 1rem;
    }
    .admin-region-heading {
        margin-top: 2.5rem;
        margin-bottom: 0.5rem;
        font-size: 1.1rem;
    }
    .admin-subtab-heading {
        margin-top: 1.75rem;
        margin-bottom: 0.5rem;
        font-size: 1rem;
        font-weight: 600;
    }
    .admin-subtab-heading:first-of-type {
        margin-top: 0;
    }

    .admin-table-edit-row td {
        vertical-align: top;
        background: var(--hover-bg, #f9fafb);
    }
    .admin-region-edit-panel {
        padding: 0.35rem 0 0.25rem;
    }
    .admin-region-edit-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
        gap: 0.75rem 1rem;
        margin-bottom: 0.75rem;
    }
    .admin-region-edit-grid label {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        font-size: 0.85rem;
    }
    .admin-region-edit-grid input,
    .admin-region-edit-grid select {
        width: 100%;
        padding: 0.35rem 0.5rem;
    }
    .admin-region-edit-span2 {
        grid-column: span 2;
    }
    @media (max-width: 720px) {
        .admin-region-edit-span2 {
            grid-column: span 1;
        }
    }
    .admin-region-edit-full {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        margin-bottom: 0.75rem;
        font-size: 0.85rem;
    }
    .admin-region-edit-full textarea {
        width: 100%;
        font-family: ui-monospace, monospace;
        font-size: 0.9rem;
    }
    .admin-region-edit-actions {
        display: flex;
        gap: 0.5rem;
        flex-wrap: wrap;
    }

    .admin-faq-toolbar {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 0.75rem;
        margin: 1rem 0 0.5rem;
        flex-wrap: wrap;
    }
    .admin-faq-toolbar-label {
        font-weight: 600;
        font-size: 0.95rem;
    }
    .admin-faq-pair {
        margin-bottom: 0.75rem;
        border: 1px solid var(--border-color, #e5e7eb);
        border-radius: 8px;
        padding: 0.35rem 0.75rem 0.75rem;
        background: var(--card-bg, #fff);
    }
    .admin-faq-pair summary {
        cursor: pointer;
        font-weight: 600;
        padding: 0.35rem 0;
    }
    .admin-faq-pair-fields {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        margin-top: 0.35rem;
    }
    .admin-faq-pair-fields label {
        font-size: 0.85rem;
    }

    .org-autosuggest-wrapper {
        position: relative;
    }
    .org-autosuggest-row {
        display: flex;
        gap: 0.5rem;
        align-items: center;
    }
    .org-suggestions {
        position: absolute;
        top: 100%;
        left: 0;
        right: 0;
        z-index: 100;
        list-style: none;
        margin: 0;
        padding: 0;
        background: var(--card-bg, #1e1e2e);
        border: 1px solid var(--border-color, #444);
        border-top: none;
        border-radius: 0 0 6px 6px;
        max-height: 240px;
        overflow-y: auto;
    }
    .org-suggestions li button {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        width: 100%;
        padding: 0.5rem 0.75rem;
        border: none;
        background: none;
        color: var(--text-color, #ccc);
        cursor: pointer;
        font-size: 0.9rem;
        text-align: left;
    }
    .org-suggestions li button:hover {
        background: var(--accent-bg, #2a2a3e);
    }
    .org-sug-meta {
        font-size: 0.8rem;
        color: var(--text-faint, #888);
    }
</style>
