<script>
    import { onMount } from "svelte";
    import { auth } from "$lib/stores/auth";

    const getBase = () =>
        import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

    let authenticated = false;
    let password = "";
    let activeTab = "mondasok";

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
    let editingAttraction = null;

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
                fetchWeatherTranslations();
                editingWeatherTrans = null;
            } else {
                showAlert("Hiba: " + (await res.text()));
            }
        } else {
            const res = await fetch(`${getBase()}/api/admin/weather_translations`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(newWeatherTrans),
            });
            if (res.ok) {
                fetchWeatherTranslations();
                newWeatherTrans = { source_text: "", lang: "hu", translated_text: "" };
            } else {
                showAlert("Hiba: " + (await res.text()));
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
        else showAlert("Hiba: " + (await res.text()));
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
            if (res.ok) await showAlert("Beállítások mentve.");
            else await showAlert("Hiba: " + (await res.text()));
        } catch (e) {
            await showAlert("Hiba: " + e.message);
        } finally {
            settingsSaving = false;
        }
    }

    async function clearWeatherCache() {
        settingsCacheClearing = true;
        try {
            const res = await fetch(`${getBase()}/api/admin/settings/clear-weather-cache`, { method: "POST" });
            if (res.ok) {
                await showAlert("Időjárás cache verzió növelve – látogatók friss adatot fognak kapni.");
                fetchSettings();
            } else await showAlert("Hiba: " + (await res.text()));
        } catch (e) {
            await showAlert("Hiba: " + e.message);
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
                await showAlert("Oldal mentve.");
                editingPage = null;
                fetchPages();
            } else {
                await showAlert("Hiba: " + (await res.text()));
            }
        } catch (e) {
            await showAlert("Hiba: " + e.message);
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
                await showAlert("GYIK / disclaimer mentve.");
                editingPageFaq = null;
                fetchPages();
            } else {
                await showAlert("Hiba: " + (await res.text()));
            }
        } catch (e) {
            await showAlert("Hiba: " + e.message);
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

    function contentPreview(text, maxLen = 96) {
        if (!text || !String(text).trim()) return "—";
        const t = String(text).replace(/\s+/g, " ").trim();
        return t.length > maxLen ? t.slice(0, maxLen) + "…" : t;
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
                await showAlert("Hiba (megye): " + (await res.text()));
                return;
            }
            if (ec.seat_location_id) {
                const ok = await setCountySeat(Number(ec.seat_location_id));
                if (!ok) {
                    await showAlert(
                        "A megye szövege mentve, de a megyeszékhely beállítása nem sikerült.",
                    );
                }
            }
            editingCounty = null;
            await showAlert("Megye mentve: " + ec.name);
            fetchCountyRegions();
            fetchLocations();
        } catch (e) {
            await showAlert("Hiba: " + e.message);
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
                await showAlert("Hiba (szék): " + (await res.text()));
                return;
            }
            editingHistoricalSeat = null;
            await showAlert("Szék mentve: " + h.name);
            fetchCountyRegions();
        } catch (e) {
            await showAlert("Hiba: " + e.message);
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
    async function submitEvent(e) {
        e.preventDefault();
        const lid = parseInt(String(newEvent.location_id), 10);
        const payload = {
            ...newEvent,
            location_id: Number.isFinite(lid) && lid > 0 ? lid : 0,
        };
        const err = validateEventFields(payload);
        if (err) {
            await showAlert(err);
            return;
        }
        createRecord(
            "events",
            payload,
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
        const st = ev.start_time ? String(ev.start_time) : "";
        const et = ev.end_time ? String(ev.end_time) : "";
        editingEvent = {
            ...ev,
            start_time: st.length >= 5 ? st.slice(0, 5) : "",
            end_time: et.length >= 5 ? et.slice(0, 5) : "",
        };
        orgEditQuery = ev.organizer || "";
    }
    function cancelEditEvent() {
        editingEvent = null;
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
            <a href="/" target="_blank" class="admin-sidebar-btn" title="Vissza a Lámsza-ra">
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
                class="admin-sidebar-btn {activeTab === 'counties'
                    ? 'active'
                    : ''}"
                on:click={() => {
                    activeTab = "counties";
                    fetchCountyRegions();
                }}
                title="Megyék"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
                    ><path d="M3 21V3h18v18H3zm2-2h14V5H5v14zm2-2h10v-2H7v2zm0-4h10v-2H7v2z"
                    ></path></svg
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
                class="admin-sidebar-btn {activeTab === 'attractions'
                    ? 'active'
                    : ''}"
                on:click={() => { activeTab = "attractions"; fetchAttractions(); }}
                title="Látnivalók"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z"></path><circle cx="12" cy="12" r="3"></circle></svg>
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'events'
                    ? 'active'
                    : ''}"
                on:click={() => {
                    activeTab = "events";
                    fetchEvents();
                }}
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

            <button
                class="admin-sidebar-btn {activeTab === 'settings'
                    ? 'active'
                    : ''}"
                on:click={() => { activeTab = "settings"; fetchSettings(); }}
                title="Beállítások"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                    ><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg
                >
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'weather_translations' ? 'active' : ''}"
                on:click={() => { activeTab = 'weather_translations'; fetchWeatherTranslations(); }}
                title="Időjárás fordítások"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 8l6 6"></path><path d="M4 14l6-6 2-3"></path><path d="M2 5h12"></path><path d="M7 2h3l2 2 2 4-3 1"></path><path d="M22 22l-5-10-5 10"></path><path d="M14 18h6"></path></svg>
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'pages' ? 'active' : ''}"
                on:click={() => { activeTab = 'pages'; fetchPages(); }}
                title="Oldalak"
            >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg>
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

            {#if eventsWithIncompleteDateTime.length > 0}
                <div class="admin-alert admin-alert--warning" role="alert">
                    <strong>Hiányos esemény-időpontok.</strong>
                    {eventsWithIncompleteDateTime.length} eseménynél nincs meg minden
                    kötelező mező (kezdő/befejező dátum és időpont). Kérjük, szerkessze
                    az Események lapon a listában szereplő tételeket, és töltse ki a dátumokat
                    és az óra:perc időpontokat.
                    <button
                        type="button"
                        class="btn-update admin-alert__btn"
                        on:click={() => {
                            activeTab = "events";
                            fetchEvents();
                        }}>Ugrás az Eseményekhez</button
                    >
                </div>
            {/if}

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
                    <p class="admin-form-hint">
                        A <strong>kezdő és befejező dátum</strong> és a hozzájuk tartozó
                        <strong>időpontok (óra:perc)</strong> mind kötelezőek — a mentés nélkülük nem lehetséges.
                    </p>
                    <form class="admin-form" on:submit={submitEvent}>
                        <label for="event_loc"
                            >Település / Helyszín <span class="admin-req" title="Kötelező"
                                >*</span
                            ></label
                        >
                        <select
                            id="event_loc"
                            bind:value={newEvent.location_id}
                            required
                        >
                            <option value="">Válassz...</option>
                            {#each settlementsForSelect as loc}
                                <option value={loc.id}
                                    >{loc.name} ({loc.county})</option
                                >
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

                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Cím</th>
                                    <th>Kezdés</th>
                                    <th>Befejezés</th>
                                    <th>Típus</th>
                                    <th>Helyszín</th>
                                    <th>Szervező</th>
                                    <th>Szerk.</th>
                                    <th>Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each events as e}
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
                                        ><td colspan="8"
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

                <!-- Beállítások (Settings) Tab -->
                {#if activeTab === "settings"}
                    <section class="admin-form-section">
                        <h3>Alapértelmezett település (MyLocation)</h3>
                        <p class="admin-hint">A kezdőlap időjárás widgetje és az események szűrése ezt a települést használja alapértelmezettként.</p>
                        <div class="admin-form" style="max-width: 32rem;">
                            <label for="my_location_slug">Település</label>
                            <select id="my_location_slug" bind:value={siteSettings.my_location_slug}>
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
                            <select id="weather_provider_default" bind:value={siteSettings.weather_provider_default}>
                                <option value="open_meteo">Open-Meteo</option>
                                <option value="weatherapi_com">WeatherAPI.com</option>
                                <option value="openweathermap">OpenWeatherMap</option>
                            </select>

                            <span class="form-group-label">Szolgáltatók engedélyezése</span>
                            <div class="flex gap-lg flex-wrap mb-lg">
                                <label class="flex items-center gap-xs font-normal">
                                    <input type="checkbox" checked={siteSettings.weather_provider_open_meteo_enabled === 'true'} on:change={(e) => siteSettings.weather_provider_open_meteo_enabled = e.target.checked ? 'true' : 'false'} class="w-auto" />
                                    Open-Meteo
                                </label>
                                <label class="flex items-center gap-xs font-normal">
                                    <input type="checkbox" checked={siteSettings.weather_provider_weatherapi_enabled === 'true'} on:change={(e) => siteSettings.weather_provider_weatherapi_enabled = e.target.checked ? 'true' : 'false'} class="w-auto" />
                                    WeatherAPI.com
                                </label>
                                <label class="flex items-center gap-xs font-normal">
                                    <input type="checkbox" checked={siteSettings.weather_provider_openweathermap_enabled === 'true'} on:change={(e) => siteSettings.weather_provider_openweathermap_enabled = e.target.checked ? 'true' : 'false'} class="w-auto" />
                                    OpenWeatherMap
                                </label>
                            </div>

                            <label for="weather_icon_style">Időjárás ikon stílus</label>
                            <select id="weather_icon_style" bind:value={siteSettings.weather_icon_style}>
                                <option value="emoji">Emoji</option>
                                <option value="svg">SVG (saját ikonok)</option>
                            </select>

                            <label for="weather_cache_ttl">Időjárás cache TTL (perc)</label>
                            <input id="weather_cache_ttl" type="number" min="1" max="1440" bind:value={siteSettings.weather_cache_ttl_minutes} />

                            <label for="weather_active_users">Aktív felhasználók becslése</label>
                            <input id="weather_active_users" type="number" min="1" bind:value={siteSettings.weather_active_users_estimate} />

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
                    <p class="admin-hint">Ha nincs admin fordítás, a rendszer az alapértelmezett magyar szöveget jeleníti meg. Több nyelv támogatása (hu, ro, de) készen áll.</p>
                    {#if editingWeatherTrans}
                        <h3>Fordítás szerkesztése</h3>
                        <form class="admin-form" on:submit={saveWeatherTranslation} style="max-width: 28rem;">
                            <label for="wet_src">Eredeti szöveg (pl. API angol)</label>
                            <input id="wet_src" type="text" bind:value={editingWeatherTrans.source_text} required />
                            <label for="wet_lang">Nyelv</label>
                            <select id="wet_lang" bind:value={editingWeatherTrans.lang}>
                                {#each WEATHER_TRANS_LANGS as opt}
                                    <option value={opt.value}>{opt.label}</option>
                                {/each}
                            </select>
                            <label for="wet_txt">Lefordított szöveg</label>
                            <input id="wet_txt" type="text" bind:value={editingWeatherTrans.translated_text} required />
                            <div class="flex gap-md mt-md">
                                <button type="submit" class="admin-submit-btn">Mentés</button>
                                <button type="button" class="btn-update" on:click={cancelEditWeatherTrans}>Mégse</button>
                            </div>
                        </form>
                    {:else}
                        <h3>Új fordítás</h3>
                        <form class="admin-form" on:submit={saveWeatherTranslation} style="max-width: 28rem;">
                            <label for="wt_src">Eredeti szöveg (pl. overcast, partly cloudy)</label>
                            <input id="wt_src" type="text" bind:value={newWeatherTrans.source_text} required placeholder="pl. overcast" />
                            <label for="wt_lang">Nyelv</label>
                            <select id="wt_lang" bind:value={newWeatherTrans.lang}>
                                {#each WEATHER_TRANS_LANGS as opt}
                                    <option value={opt.value}>{opt.label}</option>
                                {/each}
                            </select>
                            <label for="wt_txt">Lefordított szöveg</label>
                            <input id="wt_txt" type="text" bind:value={newWeatherTrans.translated_text} required placeholder="pl. borult" />
                            <button type="submit" class="admin-submit-btn">Hozzáadás</button>
                        </form>
                    {/if}
                    <div class="admin-table-wrapper mt-lg">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Eredeti</th>
                                    <th>Nyelv</th>
                                    <th>Fordítás</th>
                                    <th>Szerk.</th>
                                    <th>Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each weatherTranslations as wt}
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
                    {#if editingPage}
                        <h3>Oldal szerkesztése: {editingPage.title}</h3>
                        <form class="admin-form" on:submit|preventDefault={savePage} style="max-width: 48rem;">
                            <label for="page_title">Cím</label>
                            <input id="page_title" type="text" bind:value={editingPage.title} required />

                            <label for="page_content">Tartalom (HTML)</label>
                            <textarea id="page_content" bind:value={editingPage.content} rows="20" style="font-family: monospace; font-size: 0.85rem;"></textarea>

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
                            <input id="pfaq_label" type="text" bind:value={editingPageFaq.label_hu} />

                            <label for="pfaq_title">GYIK szekció címe (H2)</label>
                            <input id="pfaq_title" type="text" bind:value={editingPageFaq.faq_title} placeholder="pl. Hogyan működik ez az oldal?" />

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
                                            type="text"
                                            bind:value={editingPageFaq.faq_items[i].question}
                                            placeholder="Rövid kérdés"
                                        />
                                        <label for={"pfaq_a_" + i}>Válasz (Markdown)</label>
                                        <textarea
                                            id={"pfaq_a_" + i}
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
                        <h3 class="admin-subtab-heading">Irányelvek és statikus oldalak</h3>
                        <div class="admin-table-wrapper">
                            <table class="admin-table">
                                <thead>
                                    <tr>
                                        <th>Slug</th>
                                        <th>Cím</th>
                                        <th>Utolsó módosítás</th>
                                        <th>Szerk.</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {#each adminPages as pg}
                                        <tr>
                                            <td><a href="/{pg.slug}" target="_blank">/{pg.slug}</a></td>
                                            <td>{pg.title}</td>
                                            <td>{pg.updated_at ? pg.updated_at.slice(0, 19) : ''}</td>
                                            <td><button type="button" class="btn-update" on:click={() => startEditPage(pg)}>Szerk.</button></td>
                                        </tr>
                                    {:else}
                                        <tr><td colspan="4">Nincsenek oldalak.</td></tr>
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
                        <div class="admin-table-wrapper">
                            <table class="admin-table">
                                <thead>
                                    <tr>
                                        <th>Kulcs</th>
                                        <th>Megjelenített név</th>
                                        <th>GYIK cím</th>
                                        <th>Kérdések</th>
                                        <th>Utolsó módosítás</th>
                                        <th>Szerk.</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {#each pageFaqSections as row}
                                        <tr>
                                            <td><code>{row.section_key}</code></td>
                                            <td>{row.label_hu}</td>
                                            <td class="admin-table-cell-preview" title={row.faq_title || ''}>{row.faq_title || "—"}</td>
                                            <td>{(row.faq_items || []).length}</td>
                                            <td>{row.updated_at ? row.updated_at.slice(0, 19) : ''}</td>
                                            <td><button type="button" class="btn-update" on:click={() => startEditPageFaq(row)}>Szerk.</button></td>
                                        </tr>
                                    {:else}
                                        <tr><td colspan="6">Nincs GYIK rekord (futtasd a backend migrációt).</td></tr>
                                    {/each}
                                </tbody>
                            </table>
                        </div>
                    {/if}
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

                <!-- Attractions Tab -->
                {#if activeTab === "attractions"}
                    <p class="admin-info">Látnivalók (pl. Szent Anna-tó) kezelése. Megye, név, koordináták, tartalom.</p>
                    <form class="admin-form mb-lg" on:submit|preventDefault={submitNewAttraction}>
                        <h4>Új látnivaló</h4>
                        <div class="form-row">
                            <label for="att_county">Megye</label>
                            <select id="att_county" bind:value={newAttraction.county_slug}>
                                <option value="hargita">Hargita</option>
                                <option value="kovaszna">Kovászna</option>
                                <option value="maros">Maros</option>
                            </select>
                        </div>
                        <div class="form-row">
                            <label for="att_name">Név</label>
                            <input id="att_name" type="text" bind:value={newAttraction.name} required placeholder="pl. Szent Anna-tó" />
                        </div>
                        <div class="form-row">
                            <label for="att_desc">Rövid leírás</label>
                            <input id="att_desc" type="text" bind:value={newAttraction.description} placeholder="Közép-Európa egyetlen vulkanikus tava..." />
                        </div>
                        <div class="form-row">
                            <label for="att_coords">Koordináták (lat, lon)</label>
                            <input id="att_coords" type="text" bind:value={newAttraction.latitude} placeholder="46.1265" style="width:6rem" />
                            <input type="text" bind:value={newAttraction.longitude} placeholder="25.8876" style="width:6rem" />
                        </div>
                        <div class="form-row">
                            <label for="att_featured">Kiemelt kép URL</label>
                            <input id="att_featured" type="url" bind:value={newAttraction.featured_image} placeholder="https://..." />
                        </div>
                        <div class="form-row">
                            <label for="att_content">Tartalom (Markdown)</label>
                            <textarea id="att_content" bind:value={newAttraction.content} rows="6" placeholder="## Cím&#10;Szöveg..."></textarea>
                        </div>
                        <div class="form-row">
                            <label for="att_images">Galéria URL-ek (soronként egy)</label>
                            <textarea id="att_images" bind:value={newAttraction.images} rows="3" placeholder="https://kep1.jpg&#10;https://kep2.jpg"></textarea>
                        </div>
                        <button type="submit" class="admin-submit-btn">Hozzáadás</button>
                    </form>
                    <div class="admin-table-wrap">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Név</th>
                                    <th>Megye</th>
                                    <th>Slug</th>
                                    <th></th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each attractions as att}
                                    <tr>
                                        <td>{att.name}</td>
                                        <td>{att.county_name}</td>
                                        <td>{att.slug}</td>
                                        <td>
                                            <button type="button" class="btn-sm" on:click={() => openEditAttraction(att)}>Szerkesztés</button>
                                            <button type="button" class="btn-delete btn-sm" on:click={() => deleteAttraction(att.id)}>Törlés</button>
                                        </td>
                                    </tr>
                                {:else}
                                    <tr><td colspan="4">Nincsenek látnivalók.</td></tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}

                <!-- Counties Tab -->
                {#if activeTab === "counties"}
                    <p class="admin-info">
                        Megyék: név, slug, megyeszékhely (település lista), bemutatkozó szöveg (Markdown). A
                        megyeszékhely szerkesztésénél a települések legördülő listából választhatók.
                    </p>

                    <h3 class="admin-region-heading">Megyék</h3>
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
                                    <th></th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each countiesFromAPI as c (c.id)}
                                    {#if editingCounty?.id === c.id}
                                        <tr class="admin-table-edit-row">
                                            <td colspan="7">
                                                <div class="admin-region-edit-panel">
                                                    <div class="admin-region-edit-grid">
                                                        <label>
                                                            Megye (HU)
                                                            <input type="text" bind:value={editingCounty.name} />
                                                        </label>
                                                        <label>
                                                            Név (RO)
                                                            <input type="text" bind:value={editingCounty.name_ro} />
                                                        </label>
                                                        <label>
                                                            Név (DE)
                                                            <input type="text" bind:value={editingCounty.name_de} />
                                                        </label>
                                                        <label>
                                                            Slug (URL)
                                                            <input type="text" bind:value={editingCounty.slug} placeholder="pl. hargita" />
                                                        </label>
                                                        <label class="admin-region-edit-span2">
                                                            Megyeszékhely
                                                            <select bind:value={editingCounty.seat_location_id}>
                                                                <option value="">— válassz települést —</option>
                                                                {#each settlementsForCountyName(c.name) as loc (loc.id)}
                                                                    <option value={String(loc.id)}
                                                                        >{loc.name} ({loc.type}){loc.name_ro ? " — " + loc.name_ro : ""}</option
                                                                    >
                                                                {/each}
                                                            </select>
                                                        </label>
                                                    </div>
                                                    <label class="admin-region-edit-full">
                                                        Bemutatkozás (Markdown)
                                                        <textarea
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
                                            <td>
                                                <button
                                                    type="button"
                                                    class="btn-update"
                                                    on:click={() => startEditCounty(c)}>Szerkesztés</button
                                                >
                                            </td>
                                        </tr>
                                    {/if}
                                {:else}
                                    <tr><td colspan="7">Nincs megye-adat (API / migráció).</td></tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>

                    <h3 class="admin-region-heading">Történelmi székek</h3>
                    <p class="admin-info">
                        Megjelenés: <a href="/szekek" target="_blank" rel="noopener">/szekek</a> és
                        <code>/szekek/…</code> oldalak.
                    </p>
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Név (HU)</th>
                                    <th>Név (RO)</th>
                                    <th>Név (DE)</th>
                                    <th>Slug</th>
                                    <th>Tartalom (előnézet)</th>
                                    <th></th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each historicalSeatsFromAPI as h (h.id)}
                                    {#if editingHistoricalSeat?.id === h.id}
                                        <tr class="admin-table-edit-row">
                                            <td colspan="6">
                                                <div class="admin-region-edit-panel">
                                                    <div class="admin-region-edit-grid">
                                                        <label>
                                                            Név (HU)
                                                            <input type="text" bind:value={editingHistoricalSeat.name} />
                                                        </label>
                                                        <label>
                                                            Név (RO)
                                                            <input type="text" bind:value={editingHistoricalSeat.name_ro} />
                                                        </label>
                                                        <label>
                                                            Név (DE)
                                                            <input type="text" bind:value={editingHistoricalSeat.name_de} />
                                                        </label>
                                                        <label>
                                                            Slug (URL)
                                                            <input type="text" bind:value={editingHistoricalSeat.slug} placeholder="pl. csikszek" />
                                                        </label>
                                                    </div>
                                                    <label class="admin-region-edit-full">
                                                        Tartalom (Markdown)
                                                        <textarea
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
                                            <td>
                                                <button
                                                    type="button"
                                                    class="btn-update"
                                                    on:click={() => startEditHistoricalSeat(h)}>Szerkesztés</button
                                                >
                                            </td>
                                        </tr>
                                    {/if}
                                {:else}
                                    <tr><td colspan="6">Nincs szék-adat (API / migráció).</td></tr>
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
                    <select id="eatt_county" bind:value={editingAttraction.county_slug}>
                        <option value="hargita">Hargita</option>
                        <option value="kovaszna">Kovászna</option>
                        <option value="maros">Maros</option>
                    </select>
                    <label for="eatt_name">Név</label>
                    <input id="eatt_name" type="text" bind:value={editingAttraction.name} required />
                    <label for="eatt_desc">Rövid leírás</label>
                    <input id="eatt_desc" type="text" bind:value={editingAttraction.description} />
                    <label for="eatt_coords">Koordináták (lat, lon)</label>
                    <input id="eatt_coords" type="text" bind:value={editingAttraction.latitude} style="width:6rem" />
                    <input type="text" bind:value={editingAttraction.longitude} style="width:6rem" />
                    <label for="eatt_featured">Kiemelt kép URL</label>
                    <input id="eatt_featured" type="url" bind:value={editingAttraction.featured_image} />
                    <label for="eatt_content">Tartalom (Markdown)</label>
                    <textarea id="eatt_content" bind:value={editingAttraction.content} rows="6"></textarea>
                    <label for="eatt_images">Galéria URL-ek (soronként egy)</label>
                    <textarea id="eatt_images" bind:value={editingAttraction.images} rows="3"></textarea>
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
                    on:submit|preventDefault={saveEditEvent}
                >
                    <label for="edit_ev_loc"
                        >Helyszín <span class="admin-req" title="Kötelező">*</span></label
                    >
                    <select
                        id="edit_ev_loc"
                        bind:value={editingEvent.location_id}
                        required
                    >
                        {#each settlementsForSelect as loc}
                            <option value={loc.id}
                                >{loc.name} ({loc.county})</option
                            >
                        {/each}
                    </select>

                    <label for="edit_ev_title"
                        >Cím <span class="admin-req" title="Kötelező">*</span></label
                    >
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
                            <label for="edit_ev_start_date"
                                >Kezdő dátum <span class="admin-req" title="Kötelező"
                                    >*</span
                                ></label
                            >
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
                            <label for="edit_ev_start_time"
                                >Kezdő időpont (óra:perc) <span
                                    class="admin-req"
                                    title="Kötelező">*</span
                                ></label
                            >
                            <input
                                id="edit_ev_start_time"
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
                                type="time"
                                bind:value={editingEvent.end_time}
                                required
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
                    <div class="org-autosuggest-wrapper">
                        <div class="org-autosuggest-row">
                            <input
                                id="edit_ev_org"
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

    .admin-table-cell-preview {
        max-width: 16rem;
        font-size: 0.85rem;
        color: var(--text-faint, #666);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
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
