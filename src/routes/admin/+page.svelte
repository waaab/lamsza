<script>
    import { onMount } from "svelte";
    import AdminNavIcon from "$lib/components/admin/AdminNavIcon.svelte";
    import AdminPlusIcon from "$lib/components/admin/AdminPlusIcon.svelte";
    import AdminPaginationBar from "$lib/components/admin/AdminPaginationBar.svelte";
    import { ADMIN_PAGE_SIZE, adminPageSlice } from "$lib/adminPageSlice.js";
    import { auth } from "$lib/stores/auth";
    import {
        SCHEDULE_ACTIVITY_TYPES,
        SCHEDULE_ACTIVITY_TYPE_LABELS,
    } from "$lib/scheduleActivityTypes.js";
    import { absoluteMediaUrl } from "$lib/eventImage.js";
    import { getApiBase, apiCall } from "$lib/api.js";
    import { ENTRY_TYPE_SERVICE } from "$lib/entryType.js";
    import { emptyWeekHours, normalizeHours } from "$lib/entryHours.js";
    import { emptyPhotos, normalizePhotos } from "$lib/entryPhotos.js";
    import EntryHoursEditor from "$lib/components/EntryHoursEditor.svelte";
    import EntryPhotosEditor from "$lib/components/EntryPhotosEditor.svelte";
    import GoogleSignIn from "$lib/components/GoogleSignIn.svelte";

    /** Local calendar date as YYYY-MM-DD (for date inputs). */
    function localISODate() {
        const d = new Date();
        const z = (n) => String(n).padStart(2, "0");
        return `${d.getFullYear()}-${z(d.getMonth() + 1)}-${z(d.getDate())}`;
    }

    let authenticated = false;
    let authReady = false;
    let authDenied = false;
    let googleClientId = "";
    let configUnreachable = false;
    let adminOffline = false;
    let activeTab = "welcome";

    /** Section shortcuts: order matches sidebar (below Dashboard / home). Footer hint = right column label. */
    /** Same id, icon, and title as the sidebar buttons, in sidebar order. */
    const ADMIN_WELCOME_ITEMS = [
        { id: "mondasok", title: "Mondások" },
        { id: "quicklinks", title: "Gyorslinkek" },
        { id: "entries", title: "Index" },
        { id: "entry_categories", title: "Bejegyzés Kategóriák" },
        { id: "entry_types", title: "Bejegyzés típusok" },
        { id: "locations", title: "Települések" },
        { id: "counties", title: "Megyék" },
        { id: "venues", title: "Helyszínek" },
        { id: "attractions", title: "Látnivalók" },
        { id: "events", title: "Események" },
        { id: "pages", title: "Oldalak" },
        { id: "page_faq", title: "GYIK" },
        { id: "weather_translations", title: "Időjárás fordítások" },
        { id: "newsfeeds", title: "Hírfolyamok" },
        { id: "settings", title: "Beállítások" },
    ];

    function goToAdminTab(/** @type {string} */ tab) {
        activeTab = tab;
        if (tab === "welcome") {
            fetchDashboardStats();
            fetchSettings();
        }
        if (tab === "counties") fetchCountyRegions();
        if (tab === "venues") {
            fetchVenuesCatalog();
            fetchVenueTypes();
        }
        if (tab === "locations") fetchSettlementLocationTypes();
        if (tab === "attractions") fetchAttractions();
        if (tab === "events") fetchEvents();
        if (tab === "settings") fetchSettings();
        if (tab === "weather_translations") fetchWeatherTranslations();
        if (tab === "pages") fetchPages();
        if (tab === "page_faq") fetchPageFaq();
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
        type: ENTRY_TYPE_SERVICE,
        languages: ["HU"],
        tags: "",
        verified: false,
        hours: emptyWeekHours(),
        delivery_hours: emptyWeekHours(),
        photos: emptyPhotos(),
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
        event_type_id: "",
        event_subtype_id: "",
        access_type: "public",
        organizer: "",
        featured_image: "",
        entry_price: "",
    };
    /** @type {{ id: number, slug: string, label_hu: string, sort_order: number }[]} */
    let catalogEventTypes = [];
    /** @type {{ id: number, event_type_id: number, slug: string, label_hu: string, sort_order: number }[]} */
    let catalogEventSubtypes = [];
    let newCatalogEventType = { slug: "", label_hu: "", sort_order: 0 };
    /** @type {{ id: number, slug: string, label_hu: string, sort_order: number } | null} */
    let editingCatalogEventType = null;
    let newCatalogEventSubtype = {
        event_type_id: "",
        slug: "",
        label_hu: "",
        sort_order: 0,
    };
    /** @type {{ id: number, event_type_id: number, slug: string, label_hu: string, sort_order: number } | null} */
    let editingCatalogEventSubtype = null;
    let searchCatalogTypes = "";
    let searchCatalogSubtypes = "";
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
    let pageMondasok = 1;
    let pageQuickLinks = 1;
    let pageNewsFeeds = 1;
    let pageLocations = 1;
    let pageVenueTypes = 1;
    let pageVenues = 1;
    let pageEvents = 1;
    let pageEntryCategories = 1;
    let pageEntries = 1;
    let pageWeatherTrans = 1;
    let pageAdminPages = 1;
    let pagePageFaqRows = 1;
    let pageEntryTypes = 1;
    let pageAttractions = 1;
    let pageCounties = 1;
    let pageHistoricalSeats = 1;
    let pageCatalogTypes = 1;
    let pageCatalogSubtypes = 1;
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
        type: ENTRY_TYPE_SERVICE,
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

    /** Row counts from DB (GET /api/admin/dashboard_stats); keys match ADMIN_WELCOME_ITEMS id. */
    let dashboardStats = /** @type {Record<string, number>} */ ({});
    let dashboardStatsFetched = false;
    let dashboardStatsError = "";
    let dashboardStatsRequest = 0;
    let settingsLoaded = false;
    let settingsLoadError = "";
    /** @type {{ name: string, detail: string }[]} */
    let browserCacheRows = [];
    /** @type {{ id: string, level: string, text: string, tab?: string, action?: string }[]} */
    let browserCacheNotices = [];
    /** @type {{ source: string, text: string }[]} */
    let apiNotices = [];

    const ADMIN_API_LABELS = {
        mondasok: "Mondások",
        quick_links: "Gyorslinkek",
        news_feeds: "Hírfolyamok",
        locations: "Települések",
        attractions: "Látnivalók",
        entries: "Index",
        entry_categories: "Bejegyzés kategóriák",
        entry_types: "Bejegyzés típusok",
        events: "Események",
        catalog_event_types: "Eseménytípusok",
        catalog_event_subtypes: "Esemény-altípusok",
        pages: "Oldalak",
        page_faq: "GYIK",
        venues: "Helyszínek",
        venue_types: "Helyszíntípusok",
        settlement_location_types: "Településtípusok",
        counties: "Megyék",
        historical_seats: "Történelmi székek",
        weather_translations: "Időjárás fordítások",
    };

    function rememberApiError(source, message) {
        const text = String(message || "Az API nem válaszolt.");
        apiNotices = [
            ...apiNotices.filter((n) => n.source !== source),
            { source, text },
        ];
    }

    function forgetApiError(source) {
        if (!apiNotices.some((n) => n.source === source)) return;
        apiNotices = apiNotices.filter((n) => n.source !== source);
    }

    /** @type {{ id: number, name: string, slug: string, owner_email: string }[]} */
    let listingQueueUnpublished = [];
    /** @type {{ entry_id: number, entry_name: string, user_id: number, email: string }[]} */
    let listingQueueMembers = [];
    /** @type {{ id: number, domain: string, title: string, description: string, submitter: string }[]} */
    let listingQueueWebsites = [];
    let listingQueueError = "";
    let listingQueueFetched = false;

    function describeApiFailure(subject, status, body) {
        const raw = String(body || "").trim();
        if (/data api stopped/i.test(raw)) {
            return `${subject} nem tölthető be: a tartalom API le van állítva, csak a Google-belépés fut.`;
        }
        if (status === 401 || status === 403) {
            return `${subject} nem tölthető be: a szerver elutasította a kérést (HTTP ${status}). A munkamenet lejárt, vagy a fiók nem admin.`;
        }
        if (status === 404) {
            return `${subject} nem tölthető be: ez az útvonal nincs bekötve a szerveren (HTTP 404).`;
        }
        if (status >= 500) {
            return `${subject} nem tölthető be: a szerver hibával válaszolt (HTTP ${status}).`;
        }
        if (status) {
            return `${subject} nem tölthető be: a kérés nem sikerült (HTTP ${status}).`;
        }
        return `${subject} nem tölthető be: a kérés nem ért el a szerverig.`;
    }

    function describeTransportError(subject, error) {
        const msg = String(error?.message || error || "");
        if (/failed to fetch|networkerror|load failed|network request failed/i.test(msg)) {
            return `${subject} nem tölthető be: a kérés nem ért el a szerverig. Az API nem fut, vagy a böngésző nem éri el.`;
        }
        if (msg.startsWith(subject)) return msg;
        return msg || `${subject} nem tölthető be.`;
    }

    async function fetchDashboardStats() {
        const requestId = ++dashboardStatsRequest;
        if (!dashboardStatsFetched) dashboardStatsError = "";
        let lastError = "";
        for (let attempt = 0; attempt < 2; attempt++) {
            if (requestId !== dashboardStatsRequest) return;
            try {
                const res = await apiCall("/api/admin/dashboard_stats");
                if (requestId !== dashboardStatsRequest) return;
                if (!res.ok) {
                    throw new Error(describeApiFailure("A táblaszámlálók", res.status, await res.text()));
                }
                const raw = await res.json();
                if (requestId !== dashboardStatsRequest) return;
                if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
                    throw new Error("Az API válasza nem tartalmazza a táblaszámlálókat.");
                }
                const next = /** @type {Record<string, number>} */ ({});
                for (const item of ADMIN_WELCOME_ITEMS) {
                    const n = Number(raw[item.id]);
                    if (!Number.isFinite(n)) {
                        throw new Error("Az API válaszából hiányoznak a táblaszámlálók.");
                    }
                    next[item.id] = n;
                }
                dashboardStats = next;
                dashboardStatsFetched = true;
                dashboardStatsError = "";
                return;
            } catch (e) {
                lastError = describeTransportError("A táblaszámlálók", e);
                if (attempt === 0) {
                    await new Promise((resolve) => setTimeout(resolve, 400));
                }
            }
        }
        if (requestId !== dashboardStatsRequest) return;
        dashboardStatsError = lastError;
        console.error(lastError);
    }

    function collectBrowserCaches() {
        /** @type {{ name: string, detail: string }[]} */
        const rows = [];
        const weatherVersion =
            siteSettings.weather_cache_version != null
                ? String(siteSettings.weather_cache_version)
                : "";
        const linksVersion =
            siteSettings.quick_links_version != null
                ? String(siteSettings.quick_links_version)
                : "";
        const ttlMin = Number(siteSettings.weather_cache_ttl_minutes);
        const weatherTtlMs = Number.isFinite(ttlMin) && ttlMin > 0 ? ttlMin * 60 * 1000 : 15 * 60 * 1000;
        const hourMs = 60 * 60 * 1000;
        const newsTtlMs = 30 * 60 * 1000;
        let weatherTotal = 0;
        let weatherFresh = 0;
        let weatherStale = 0;
        let newsTotal = 0;
        let newsFresh = 0;
        let newsStale = 0;
        let promotedDetail = "nincs mentett gyorslink-válasz ebben a böngészőben";
        let promotedState = "nincs";

        const ageState = (timestamp, ttlMs, version, expectedVersion) => {
            if (!Number.isFinite(timestamp)) return "nincs időbélyeg";
            const expired = Date.now() - timestamp >= ttlMs;
            const versionOk = !expectedVersion || String(version ?? "") === expectedVersion;
            if (!versionOk) return "régi szerververzió";
            if (expired) return "lejárt";
            return "érvényes";
        };

        try {
            for (let i = 0; i < localStorage.length; i++) {
                const key = localStorage.key(i);
                if (!key) continue;
                let parsed = null;
                try {
                    parsed = JSON.parse(localStorage.getItem(key) || "");
                } catch {
                    parsed = null;
                }
                if (key.startsWith("weather_cache_")) {
                    weatherTotal += 1;
                    const state = parsed
                        ? ageState(parsed.timestamp, weatherTtlMs, parsed.cache_version, weatherVersion)
                        : "sérült";
                    if (state === "érvényes") weatherFresh += 1;
                    else weatherStale += 1;
                } else if (key === "promoted_links_cache") {
                    promotedState = parsed
                        ? ageState(parsed.timestamp, hourMs, parsed.version, linksVersion)
                        : "sérült";
                    promotedDetail = promotedState;
                } else if (key === "hirek_cache" || key.startsWith("news_cache")) {
                    newsTotal += 1;
                    const state = parsed ? ageState(parsed.timestamp, newsTtlMs, "", "") : "sérült";
                    if (state === "érvényes") newsFresh += 1;
                    else newsStale += 1;
                }
            }
        } catch {
            /* private mode */
        }

        const weatherServer = settingsLoaded
            ? `A szerver TTL ${ttlMin || "—"} perc, verzió ${weatherVersion || "—"}.`
            : "A szerver verziója most nem ismert.";
        rows.push({
            name: "Időjárás",
            detail: `${weatherServer}. Ebben a böngészőben ${weatherTotal} mentés, ${weatherFresh} érvényes.`,
        });
        rows.push({
            name: "Gyorslinkek",
            detail: `${
                settingsLoaded ? `szerververzió ${linksVersion || "—"}, böngésző TTL 60 perc` : "szerververzió még nincs betöltve"
            }. ${promotedDetail}.`,
        });
        rows.push({
            name: "Hírek",
            detail:
                newsTotal > 0
                    ? `30 perces böngésző TTL, szerververzió nélkül. ${newsTotal} mentés, ${newsFresh} érvényes, ${newsStale} lejárt vagy sérült.`
                    : "30 perces böngésző TTL, szerververzió nélkül. Ebben a böngészőben nincs hírmásolat.",
        });
        browserCacheRows = rows;

        /** @type {{ id: string, level: string, text: string, action: string }[]} */
        const notices = [];
        if (weatherTotal === 0) {
            notices.push({
                id: "cache-weather",
                level: "info",
                text: `Időjárás: ebben a böngészőben nincs mentett előrejelzés. ${weatherServer}`,
                action: "cache-refresh",
            });
        } else if (weatherStale > 0) {
            notices.push({
                id: "cache-weather",
                level: "warning",
                text: `Időjárás: ${weatherStale} mentés lejárt, régi verziójú vagy sérült. Érvényes: ${weatherFresh}. ${weatherServer}`,
                action: "cache-refresh",
            });
        } else {
            notices.push({
                id: "cache-weather",
                level: "success",
                text: `Időjárás: ${weatherFresh} mentés érvényes. ${weatherServer}`,
                action: "cache-refresh",
            });
        }

        const linksServer = settingsLoaded
            ? `Szerververzió ${linksVersion || "—"}, böngésző TTL 60 perc.`
            : "A szerververzió most nem ismert. Böngésző TTL 60 perc.";
        if (promotedState === "nincs") {
            notices.push({
                id: "cache-links",
                level: "info",
                text: `Gyorslinkek: ebben a böngészőben nincs mentett lista. ${linksServer}`,
                action: "cache-refresh",
            });
        } else if (promotedState === "érvényes") {
            notices.push({
                id: "cache-links",
                level: "success",
                text: `Gyorslinkek: a mentett lista érvényes. ${linksServer}`,
                action: "cache-refresh",
            });
        } else {
            notices.push({
                id: "cache-links",
                level: "warning",
                text: `Gyorslinkek: a mentett lista ${promotedDetail}. ${linksServer}`,
                action: "cache-refresh",
            });
        }

        if (newsTotal === 0) {
            notices.push({
                id: "cache-news",
                level: "info",
                text: "Hírek: ebben a böngészőben nincs mentett hírlista. A másolat 30 percig él, szerververzió nélkül.",
                action: "cache-refresh",
            });
        } else if (newsStale > 0) {
            notices.push({
                id: "cache-news",
                level: "warning",
                text: `Hírek: ${newsStale} mentés lejárt vagy sérült. Érvényes: ${newsFresh}. A másolat 30 percig él, szerververzió nélkül.`,
                action: "cache-refresh",
            });
        } else {
            notices.push({
                id: "cache-news",
                level: "success",
                text: `Hírek: ${newsFresh} mentés érvényes. A másolat 30 percig él, szerververzió nélkül.`,
                action: "cache-refresh",
            });
        }
        browserCacheNotices = notices;
    }

    async function fetchListingQueue() {
        listingQueueError = "";
        try {
            const res = await apiCall("/api/admin/listing-queue");
            if (!res.ok) {
                listingQueueError = describeApiFailure(
                    "A bejegyzés-jóváhagyások",
                    res.status,
                    await res.text(),
                );
                listingQueueFetched = true;
                return;
            }
            const data = await res.json();
            listingQueueUnpublished = Array.isArray(data.unpublished) ? data.unpublished : [];
            listingQueueMembers = Array.isArray(data.members) ? data.members : [];
            listingQueueWebsites = Array.isArray(data.websites) ? data.websites : [];
            listingQueueFetched = true;
        } catch (e) {
            listingQueueError = describeTransportError("A bejegyzés-jóváhagyások", e);
            listingQueueFetched = true;
            console.error(e);
        }
    }

    async function publishListingQueueEntry(entryId) {
        const res = await apiCall("/api/admin/listing-queue/publish", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ entry_id: entryId }),
        });
        if (!res.ok) {
            listingQueueError = (await res.text()) || `HTTP ${res.status}`;
            return;
        }
        await fetchListingQueue();
        await auth.refresh();
    }

    async function approveListingQueueMember(entryId, userId) {
        const res = await apiCall("/api/admin/listing-queue/member", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ entry_id: entryId, user_id: userId, action: "approve" }),
        });
        if (!res.ok) {
            listingQueueError = (await res.text()) || `HTTP ${res.status}`;
            return;
        }
        await fetchListingQueue();
        await auth.refresh();
    }

    async function rejectListingQueueMember(entryId, userId) {
        const res = await apiCall("/api/admin/listing-queue/member", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ entry_id: entryId, user_id: userId, action: "reject" }),
        });
        if (!res.ok) {
            listingQueueError = (await res.text()) || `HTTP ${res.status}`;
            return;
        }
        await fetchListingQueue();
        await auth.refresh();
    }

    async function reviewWebsite(websiteId, action) {
        const res = await apiCall("/api/admin/websites", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ id: websiteId, action }),
        });
        if (!res.ok) {
            listingQueueError = (await res.text()) || `HTTP ${res.status}`;
            return;
        }
        await fetchListingQueue();
        await auth.refresh();
    }

    /** Cím + rövid köszöntő / leírás — minden admin-fülön egységes fejléc. */
    const ADMIN_PAGE_COPY = {
        welcome: {
            title: "Dashboard",
            greeting:
                "Üdvözöllek. A kártyákon a táblák rekordjainak száma látható, a név pedig megegyezik az oldalsáv gombjaival. Kattintva megnyílik a megfelelő kezelőfelület.",
        },
        mondasok: {
            title: "Mondások",
            greeting:
                "A kezdőlap napi idézeteinek és megjelenési napjának kezelése.",
        },
        quicklinks: {
            title: "Gyorslinkek",
            greeting: "Kezdőlap gyors hivatkozásai: cím, URL és háttérszín.",
        },
        newsfeeds: {
            title: "Hírfolyamok",
            greeting: "RSS és Atom hírcsatornák a Hírek oldalhoz.",
        },
        locations: {
            title: "Települések",
            greeting:
                "Települések, megyék és kapcsolódó metaadatok (címer, irányítószám, típus).",
        },
        counties: {
            title: "Megyék",
            greeting: "Megyei tartalom és beállítások a történelmi székekhez kapcsolódóan.",
        },
        venues: {
            title: "Helyszínek",
            greeting:
                "Rendezvényhelyszínek: típusok, településhez kötés, és eseményekhez való hozzárendelés.",
        },
        attractions: {
            title: "Látnivalók",
            greeting: "Megyei látnivalók, leírások és képgaléria.",
        },
        events: {
            title: "Események",
            greeting:
                "Közösségi és sportesemények: időpontok, helyszín, típusok és opcionális program.",
        },
        entries: {
            title: "Bejegyzések",
            greeting: "Címtár-bejegyzések: kategória, típus, elérhetőségek és címkék.",
        },
        entry_categories: {
            title: "Bejegyzés kategóriák",
            greeting: "Index kategóriák felvétele, sorrendezése és törlése.",
        },
        entry_types: {
            title: "Bejegyzés típusok",
            greeting: "A címtárban használható bejegyzés-típusok kezelése.",
        },
        pages: {
            title: "Oldalak",
            greeting: "Statikus oldalak (szabályzatok, szöveges tartalmak) szerkesztése.",
        },
        page_faq: {
            title: "GYIK",
            greeting: "Oldalankénti gyakori kérdések és felelősségkizárások.",
        },
        weather_translations: {
            title: "Időjárás fordítások",
            greeting:
                "Az időjárás API szövegeinek fordítása magyarra, románra, németre.",
        },
        settings: {
            title: "Beállítások",
            greeting: "Rendszer-, időjárás- és egyéb szolgáltatás-beállítások.",
        },
    };

    $: adminPageHead =
        ADMIN_PAGE_COPY[activeTab] || {
            title: "Admin",
            greeting: "",
        };

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

    /** @type {{ id: number, slug: string, label_hu: string, sort_order: number }[]} */
    let settlementLocationTypes = [];
    let newSettlementLocationType = { slug: "", label_hu: "", sort_order: 0 };
    /** @type {{ id: number, slug: string, label_hu: string, sort_order: number } | null} */
    let editingSettlementLocationType = null;

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

    const ACCESS_TYPE_LABELS = {
        public: "Nyitott (bárki)",
        members_only: "Zártkörű (csak tagoknak)",
        invitation_only: "Meghívóval (zárt kör)",
    };

    /** @param {unknown} v */
    function accessTypeLabel(v) {
        const k = String(v || "public");
        return ACCESS_TYPE_LABELS[k] || k;
    }

    /** @param {unknown} slug */
    function settlementTypeLabel(slug) {
        const s = String(slug || "").trim();
        if (!s) return "—";
        const t = settlementLocationTypes.find((x) => x.slug === s);
        return t ? t.label_hu : s;
    }

    /** @param {string} slug */
    function eventTypeLabelFromCatalog(slug) {
        const s = String(slug || "").trim();
        if (!s) return "—";
        const t = catalogEventTypes.find((x) => x.slug === s);
        return t ? t.label_hu : s;
    }

    /** @param {number} typeId @param {string} subSlug */
    function eventSubtypeLabelFromCatalog(typeId, subSlug) {
        const ss = String(subSlug || "").trim();
        if (!ss) return "—";
        const s = catalogEventSubtypes.find(
            (x) =>
                x.slug === ss &&
                (typeId ? x.event_type_id === typeId : true),
        );
        return s ? s.label_hu : ss;
    }

    $: rfMondasok = filterRows(mondasok, searchMondasok, (m) => [
        m.id,
        m.text,
        m.display_date,
    ]);
    $: pgMondasok = adminPageSlice(rfMondasok, pageMondasok);
    $: mondasTodayYmd = localISODate();
    $: mondasokTodayCount = mondasok.filter(
        (m) => normalizeYmdInput(m.display_date) === mondasTodayYmd,
    ).length;
    $: rfQuickLinks = filterRows(quickLinks, searchQuickLinks, (q) => [
        q.title,
        q.url,
        q.bg_color,
    ]);
    $: pgQuickLinks = adminPageSlice(rfQuickLinks, pageQuickLinks);
    $: rfNewsFeeds = filterRows(newsFeeds, searchNewsFeeds, (nf) => [
        nf.title,
        nf.feed_url,
        String(nf.id),
    ]);
    $: pgNewsFeeds = adminPageSlice(rfNewsFeeds, pageNewsFeeds);
    $: rfLocations = filterRows(locations, searchLocations, (l) => {
        const parentN =
            l.parent_id != null && l.parent_id !== ""
                ? (() => {
                      const p = locations.find((x) => x.id === l.parent_id);
                      return p
                          ? `${p.name}${p.county ? " (" + p.county + ")" : ""}`
                          : String(l.parent_id);
                  })()
                : "";
        return [
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
            parentN,
        ];
    });
    $: pgLocations = adminPageSlice(rfLocations, pageLocations);
    $: rfVenueTypes = filterRows(venueTypesList, searchVenueTypes, (t) => [
        t.id,
        t.slug,
        t.label_hu,
    ]);
    $: pgVenueTypes = adminPageSlice(rfVenueTypes, pageVenueTypes);
    $: rfVenuesCatalog = filterRows(venuesCatalog, searchVenues, (v) => [
        v.name,
        v.slug,
        v.kind,
        v.settlement_name,
    ]);
    $: pgVenuesCatalog = adminPageSlice(rfVenuesCatalog, pageVenues);
    $: rfEvents = filterRows(events, searchEvents, (e) => {
        const loc = locations.find((l) => l.id === e.location_id);
        const locN = loc
            ? `${loc.name}${loc.county ? " (" + loc.county + ")" : ""}`
            : String(e.location_id ?? "");
        return [
            e.title,
            e.description,
            e.organizer,
            locN,
            e.default_venue_name,
            e.event_type,
            e.event_subtype,
            e.access_type,
            e.start_date,
            e.end_date,
            e.featured_image,
            e.entry_price,
        ];
    });
    $: pgEvents = adminPageSlice(rfEvents, pageEvents);
    $: rfEntryCategories = filterRows(
        entryCategories,
        searchEntryCategories,
        (cat) => [cat.id, cat.name],
    );
    $: pgEntryCategories = adminPageSlice(rfEntryCategories, pageEntryCategories);
    $: rfEntries = filterRows(entries, searchEntries, (s) => {
        const locN = locations.find((l) => l.id === s.location_id);
        const locName = locN
            ? `${locN.name}${locN.county ? " (" + locN.county + ")" : ""}`
            : String(s.location_id ?? "");
        const cat = entryCategories.find((c) => c.id === s.category_id);
        const catName = cat ? cat.name : String(s.category_id ?? "");
        return [
            s.name,
            s.type,
            s.url,
            s.phone,
            s.address,
            s.notes,
            locName,
            catName,
            (s.languages || []).join(","),
            (s.tags || []).join(","),
        ];
    });
    $: pgEntries = adminPageSlice(rfEntries, pageEntries);
    $: rfWeatherTrans = filterRows(
        weatherTranslations,
        searchWeatherTrans,
        (wt) => [wt.source_text, wt.lang, wt.translated_text],
    );
    $: pgWeatherTrans = adminPageSlice(rfWeatherTrans, pageWeatherTrans);
    $: rfAdminPages = filterRows(adminPages, searchAdminPages, (pg) => [
        pg.slug,
        pg.title,
        pg.greeting,
        pg.updated_at,
    ]);
    $: pgAdminPages = adminPageSlice(rfAdminPages, pageAdminPages);
    $: rfPageFaqRows = filterRows(pageFaqSections, searchPageFaqRows, (row) => [
        row.section_key,
        row.label_hu,
        row.faq_title,
        String((row.faq_items || []).length),
        row.updated_at,
    ]);
    $: pgPageFaqRows = adminPageSlice(rfPageFaqRows, pagePageFaqRows);
    $: rfEntryTypes = filterRows(entryTypes, searchEntryTypes, (et) => [
        et.id,
        et.name,
    ]);
    $: pgEntryTypes = adminPageSlice(rfEntryTypes, pageEntryTypes);
    $: rfAttractions = filterRows(attractions, searchAttractions, (att) => [
        att.name,
        att.slug,
        att.county_slug,
    ]);
    $: pgAttractions = adminPageSlice(rfAttractions, pageAttractions);
    $: pgCounties = adminPageSlice(
        displayCounties,
        pageCounties,
        editingCounty
            ? Math.max(ADMIN_PAGE_SIZE, displayCounties.length)
            : ADMIN_PAGE_SIZE,
    );
    $: pgHistoricalSeats = adminPageSlice(
        displayHistoricalSeats,
        pageHistoricalSeats,
        editingHistoricalSeat
            ? Math.max(ADMIN_PAGE_SIZE, displayHistoricalSeats.length)
            : ADMIN_PAGE_SIZE,
    );
    $: rfCatalogTypes = filterRows(
        catalogEventTypes,
        searchCatalogTypes,
        (t) => [t.id, t.slug, t.label_hu, String(t.sort_order)],
    );
    $: pgCatalogTypes = adminPageSlice(
        rfCatalogTypes,
        pageCatalogTypes,
        editingCatalogEventType
            ? Math.max(ADMIN_PAGE_SIZE, rfCatalogTypes.length)
            : ADMIN_PAGE_SIZE,
    );
    $: rfCatalogSubtypes = filterRows(
        catalogEventSubtypes,
        searchCatalogSubtypes,
        (s) => [s.id, s.slug, s.label_hu, String(s.sort_order), s.event_type_id],
    );
    $: pgCatalogSubtypes = adminPageSlice(
        rfCatalogSubtypes,
        pageCatalogSubtypes,
        editingCatalogEventSubtype
            ? Math.max(ADMIN_PAGE_SIZE, rfCatalogSubtypes.length)
            : ADMIN_PAGE_SIZE,
    );
    $: subtypesForNewEvent = catalogEventSubtypes.filter(
        (s) => s.event_type_id === Number(newEvent.event_type_id),
    );
    $: subtypesForEditEvent = editingEvent
        ? catalogEventSubtypes.filter(
              (s) =>
                  s.event_type_id === Number(editingEvent.event_type_id),
          )
        : [];

    onMount(() => {
        (async () => {
            try {
                const res = await apiCall("/api/config/public");
                if (res.ok) {
                    const data = await res.json();
                    googleClientId = data.google_client_id || "";
                    configUnreachable = false;
                } else {
                    configUnreachable = true;
                }
            } catch (e) {
                configUnreachable = true;
                console.error(e);
            }
            const me = await auth.refresh();
            authReady = true;
            adminOffline = !!me.offline;
            if (me.isAdmin) {
                authenticated = true;
                fetchAll();
            } else if (me.loggedIn) {
                authDenied = true;
            }
        })();

        const storedTs = localStorage.getItem("news_feed_timestamps");
        if (storedTs) {
            try {
                feedTimestamps = JSON.parse(storedTs);
            } catch (e) {}
        }
    });

    async function onGoogleSignedIn() {
        const me = await auth.refresh();
        if (me.isAdmin) {
            authenticated = true;
            authDenied = false;
            fetchAll();
        } else {
            authDenied = true;
        }
    }

    async function logout() {
        authenticated = false;
        await auth.logout();
        window.location.href = "/";
    }

    async function fetchAll() {
        collectBrowserCaches();
        await fetchDashboardStats();
        await fetchListingQueue();
        fetchMondasok();
        fetchQuickLinks();
        fetchNewsFeeds();
        fetchLocations();
        fetchSettlementLocationTypes();
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
        fetchPageFaq();
        fetchCountyRegions();
    }

    async function fetchWeatherTranslations() {
        await loadData("weather_translations", (d) => (weatherTranslations = d));
    }

    async function saveWeatherTranslation(e) {
        e?.preventDefault();
        if (editingWeatherTrans) {
            const res = await apiCall(`/api/admin/weather_translations`, {
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
            const res = await apiCall(`/api/admin/weather_translations`, {
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
        const res = await apiCall(`/api/admin/weather_translations?id=${id}`, { method: "DELETE" });
        if (res.ok) fetchWeatherTranslations();
        else setAdminTabError("Hiba: " + (await res.text()));
    }

    async function fetchSettings() {
        try {
            const res = await apiCall(`/api/admin/settings`);
            if (!res.ok) {
                settingsLoadError = describeApiFailure("A beállítások", res.status, await res.text());
                collectBrowserCaches();
                return;
            }
            const data = await res.json();
            siteSettings = {
                weather_cache_ttl_minutes: data.weather_cache_ttl_minutes ?? "15",
                weather_cache_version: data.weather_cache_version ?? "1",
                quick_links_version: data.quick_links_version ?? "1",
                weather_icon_style: data.weather_icon_style ?? "emoji",
                weather_active_users_estimate: data.weather_active_users_estimate ?? "10000",
                weather_provider_default: data.weather_provider_default ?? "open_meteo",
                weather_provider_open_meteo_enabled: data.weather_provider_open_meteo_enabled ?? "true",
                weather_provider_weatherapi_enabled: data.weather_provider_weatherapi_enabled ?? "true",
                weather_provider_openweathermap_enabled: data.weather_provider_openweathermap_enabled ?? "true",
                my_location_slug: data.my_location_slug ?? "csikszereda",
                ...data,
            };
            settingsLoaded = true;
            settingsLoadError = "";
            collectBrowserCaches();
        } catch (e) {
            settingsLoadError = describeTransportError("A beállítások", e);
            collectBrowserCaches();
            console.error(e);
        }
    }

    async function saveSettings() {
        settingsSaving = true;
        try {
            const payload = Object.fromEntries(
                Object.entries(siteSettings).map(([k, v]) => [k, v != null ? String(v) : ""])
            );
            const res = await apiCall(`/api/admin/settings`, {
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
            const res = await apiCall(`/api/admin/settings/clear-weather-cache`, { method: "POST" });
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
        await loadData("pages", (d) => (adminPages = d));
    }

    async function fetchPageFaq() {
        await loadData("page_faq", (d) => (pageFaqSections = d));
    }

    function startEditPage(page) {
        editingPageFaq = null;
        editingPage = { ...page, greeting: page.greeting ?? "" };
    }

    function cancelEditPage() {
        editingPage = null;
    }

    async function savePage() {
        if (!editingPage) return;
        pageSaving = true;
        try {
            const res = await apiCall(`/api/admin/pages`, {
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
            const res = await apiCall(`/api/admin/page_faq`, {
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
                fetchPageFaq();
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
        const path = endpoint.startsWith("/") ? endpoint : `/api/admin/${endpoint}`;
        const source = endpoint.replace(/^\/api\/(?:admin\/)?/, "");
        try {
            const res = await apiCall(path);
            if (!res.ok) {
                const label = ADMIN_API_LABELS[source] || source;
                rememberApiError(source, describeApiFailure(label, res.status, await res.text()));
                return;
            }
            setter(await res.json());
            forgetApiError(source);
        } catch (e) {
            rememberApiError(source, describeTransportError(ADMIN_API_LABELS[source] || source, e));
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

    async function fetchSettlementLocationTypes() {
        await loadData("settlement_location_types", (d) => (settlementLocationTypes = d));
    }

    async function submitNewSettlementLocationType(e) {
        e.preventDefault();
        const label_hu = String(newSettlementLocationType.label_hu || "").trim();
        let slug = String(newSettlementLocationType.slug || "")
            .trim()
            .toLowerCase();
        const sort_order = Number(newSettlementLocationType.sort_order) || 0;
        if (!label_hu) {
            await showAlert("A megnevezés kötelező.");
            return;
        }
        try {
            const res = await apiCall(`/api/admin/settlement_location_types`,
                {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ slug, label_hu, sort_order }),
                },
            );
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            newSettlementLocationType = { slug: "", label_hu: "", sort_order: 0 };
            await fetchSettlementLocationTypes();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }

    /** @param {Record<string, unknown>} t */
    function startEditSettlementLocationType(t) {
        editingSettlementLocationType = {
            id: Number(t.id),
            slug: String(t.slug || ""),
            label_hu: String(t.label_hu || ""),
            sort_order: Number(t.sort_order) || 0,
        };
    }
    function cancelEditSettlementLocationType() {
        editingSettlementLocationType = null;
    }
    async function saveEditSettlementLocationType() {
        if (!editingSettlementLocationType) return;
        const id = parseInt(String(editingSettlementLocationType.id || ""), 10);
        const label_hu = String(
            editingSettlementLocationType.label_hu || "",
        ).trim();
        const sort_order =
            Number(editingSettlementLocationType.sort_order) || 0;
        if (!Number.isFinite(id) || id < 1 || !label_hu) return;
        try {
            const res = await apiCall(`/api/admin/settlement_location_types`,
                {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        id,
                        label_hu,
                        sort_order,
                    }),
                },
            );
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            editingSettlementLocationType = null;
            await fetchSettlementLocationTypes();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }
    async function deleteSettlementLocationTypeRow(id) {
        const ok = await showConfirm(
            "Biztosan törlöd ezt a településtípust? (Nem lehetséges, ha van ilyen típusú település.)",
        );
        if (!ok) return;
        try {
            const res = await apiCall(`/api/admin/settlement_location_types?id=${encodeURIComponent(id)}`,
                { method: "DELETE" },
            );
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            await fetchSettlementLocationTypes();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }

    // Entries/events use settlement_id; filter out counties (type=megye)
    $: settlementsForSelect = locations.filter((l) => l.type !== "megye");

    /** @returns {Promise<boolean>} */
    async function setCountySeat(locationId) {
        try {
            const res = await apiCall(`/api/admin/county_seat`, {
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
    async function fetchEvents() {
        await Promise.all([
            loadData("events", (d) => (events = d)),
            loadData("catalog_event_types", (d) => (catalogEventTypes = d)),
            loadData("catalog_event_subtypes", (d) => (catalogEventSubtypes = d)),
        ]);
    }

    async function submitCatalogEventType(e) {
        e.preventDefault();
        const slug = String(newCatalogEventType.slug || "")
            .trim()
            .toLowerCase()
            .replace(/\s+/g, "-");
        const label_hu = String(newCatalogEventType.label_hu || "").trim();
        const sort_order = Number(newCatalogEventType.sort_order) || 0;
        if (!slug || !label_hu) {
            await showAlert("Slug és megnevezés kötelező.");
            return;
        }
        try {
            const res = await apiCall(`/api/admin/catalog_event_types`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ slug, label_hu, sort_order }),
            });
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            newCatalogEventType = { slug: "", label_hu: "", sort_order: 0 };
            await fetchEvents();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }

    /** @param {Record<string, unknown>} t */
    function startEditCatalogEventType(t) {
        editingCatalogEventType = {
            id: Number(t.id),
            slug: String(t.slug || ""),
            label_hu: String(t.label_hu || ""),
            sort_order: Number(t.sort_order) || 0,
        };
    }
    function cancelEditCatalogEventType() {
        editingCatalogEventType = null;
    }
    async function saveEditCatalogEventType() {
        if (!editingCatalogEventType) return;
        const id = parseInt(String(editingCatalogEventType.id || ""), 10);
        const label_hu = String(editingCatalogEventType.label_hu || "").trim();
        const sort_order = Number(editingCatalogEventType.sort_order) || 0;
        if (!Number.isFinite(id) || id < 1 || !label_hu) return;
        try {
            const res = await apiCall(`/api/admin/catalog_event_types`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    id,
                    label_hu,
                    sort_order,
                }),
            });
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            editingCatalogEventType = null;
            await fetchEvents();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }
    async function deleteCatalogEventTypeRow(id) {
        const ok = await showConfirm(
            "Biztosan törlöd ezt az eseménytípust? (Csak akkor sikerül, ha nincs hozzá esemény.)",
        );
        if (!ok) return;
        try {
            const res = await apiCall(`/api/admin/catalog_event_types?id=${encodeURIComponent(id)}`,
                { method: "DELETE" },
            );
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            await fetchEvents();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }

    async function submitCatalogEventSubtype(e) {
        e.preventDefault();
        const event_type_id = parseInt(
            String(newCatalogEventSubtype.event_type_id || ""),
            10,
        );
        const slug = String(newCatalogEventSubtype.slug || "")
            .trim()
            .toLowerCase()
            .replace(/\s+/g, "-");
        const label_hu = String(newCatalogEventSubtype.label_hu || "").trim();
        const sort_order = Number(newCatalogEventSubtype.sort_order) || 0;
        if (!Number.isFinite(event_type_id) || event_type_id < 1 || !slug || !label_hu) {
            await showAlert("Típus, slug és megnevezés kötelező.");
            return;
        }
        try {
            const res = await apiCall(`/api/admin/catalog_event_subtypes`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    event_type_id,
                    slug,
                    label_hu,
                    sort_order,
                }),
            });
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            newCatalogEventSubtype = {
                event_type_id: "",
                slug: "",
                label_hu: "",
                sort_order: 0,
            };
            await fetchEvents();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }

    /** @param {Record<string, unknown>} s */
    function startEditCatalogEventSubtype(s) {
        editingCatalogEventSubtype = {
            id: Number(s.id),
            event_type_id: Number(s.event_type_id),
            slug: String(s.slug || ""),
            label_hu: String(s.label_hu || ""),
            sort_order: Number(s.sort_order) || 0,
        };
    }
    function cancelEditCatalogEventSubtype() {
        editingCatalogEventSubtype = null;
    }
    async function saveEditCatalogEventSubtype() {
        if (!editingCatalogEventSubtype) return;
        const id = parseInt(String(editingCatalogEventSubtype.id || ""), 10);
        const label_hu = String(editingCatalogEventSubtype.label_hu || "").trim();
        const sort_order = Number(editingCatalogEventSubtype.sort_order) || 0;
        if (!Number.isFinite(id) || id < 1 || !label_hu) return;
        try {
            const res = await apiCall(`/api/admin/catalog_event_subtypes`, {
                method: "PUT",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    id,
                    label_hu,
                    sort_order,
                }),
            });
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            editingCatalogEventSubtype = null;
            await fetchEvents();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }
    async function deleteCatalogEventSubtypeRow(id) {
        const ok = await showConfirm(
            "Biztosan törlöd ezt az altípust? (Csak akkor sikerül, ha nincs hozzá esemény.)",
        );
        if (!ok) return;
        try {
            const res = await apiCall(`/api/admin/catalog_event_subtypes?id=${encodeURIComponent(id)}`,
                { method: "DELETE" },
            );
            if (!res.ok) {
                await showAlert(await res.text());
                return;
            }
            await fetchEvents();
        } catch (err) {
            await showAlert(String(err.message || err));
        }
    }
    async function fetchVenuesCatalog() {
        await loadData("venues", (d) => (venuesCatalog = d));
    }
    async function fetchVenueTypes() {
        try {
            const res = await apiCall(`/api/admin/venue_types`);
            if (!res.ok) {
                rememberApiError(
                    "venue_types",
                    describeApiFailure("A helyszíntípusok", res.status, await res.text()),
                );
                return;
            }
            forgetApiError("venue_types");
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
        } catch (e) {
            rememberApiError("venue_types", describeTransportError("A helyszíntípusok", e));
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
            const res = await apiCall(`/api/admin/venue_types`, {
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
            const res = await apiCall(`/api/admin/venue_types`, {
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
            const res = await apiCall(`/api/admin/venue_types?id=${encodeURIComponent(id)}`,
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
            const res = await apiCall(`/api/venues?settlement_id=${sid}`,
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
            const res = await apiCall(`/api/venues?settlement_id=${sid}`,
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
            const res = await apiCall(`/api/admin/venues`, {
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
            const res = await apiCall(`/api/admin/venues`, {
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
            const res = await apiCall(`/api/admin/venues?id=${encodeURIComponent(id)}`,
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
        const etid =
            typeof ev.event_type_id === "number"
                ? ev.event_type_id
                : parseInt(String(ev.event_type_id ?? "").trim(), 10);
        if (!Number.isFinite(etid) || etid < 1)
            return "Válassz eseménytípust.";
        return null;
    }

    $: eventsWithIncompleteDateTime = events.filter((e) => !eventDateTimeComplete(e));

    function apiNoticeTab(source) {
        const tabs = {
            quick_links: "quicklinks",
            news_feeds: "newsfeeds",
            venues: "venues",
            venue_types: "venues",
            catalog_event_types: "events",
            catalog_event_subtypes: "events",
            counties: "counties",
            historical_seats: "counties",
            settlement_location_types: "locations",
        };
        if (tabs[source]) return tabs[source];
        return ADMIN_API_LABELS[source] ? source : "";
    }

    function buildDashboardMessages(websites) {
        /** @type {{ id: string, level: string, text: string, tab?: string, action?: string, websiteId?: number }[]} */
        const messages = [];
        if (adminOffline) {
            messages.push({
                id: "api-offline",
                level: "error",
                text: "Az API nem elérhető. A felület a legutóbbi belépés alapján nyílt meg, adatok nélkül.",
            });
        }
        if (dashboardStatsError) {
            messages.push({
                id: "stats",
                level: "error",
                text: dashboardStatsError,
                action: "retry-stats",
            });
        }
        if (settingsLoadError) {
            messages.push({
                id: "settings",
                level: "error",
                text: settingsLoadError,
                tab: "settings",
                action: "open",
            });
        }
        if (listingQueueError) {
            messages.push({
                id: "queue-error",
                level: "error",
                text: listingQueueError,
                action: "retry-queue",
            });
        }
        for (const notice of apiNotices) {
            messages.push({
                id: `api-${notice.source}`,
                level: "error",
                text: notice.text,
                tab: apiNoticeTab(notice.source),
                action: "open",
            });
        }
        if (mondasok.length > 0 && mondasokTodayCount === 0) {
            messages.push({
                id: "mondas",
                level: "warning",
                text: `Ma (${mondasTodayYmd}) nincs beütemezett mondás, ezért a kezdőlapon a mondás-blokk rejtve marad.`,
                tab: "mondasok",
                action: "open",
            });
        }
        if (eventsWithIncompleteDateTime.length > 0) {
            messages.push({
                id: "events",
                level: "warning",
                text: `${eventsWithIncompleteDateTime.length} eseménynél hiányzik a kezdő vagy a befejező dátum és időpont.`,
                tab: "events",
                action: "open",
            });
        }
        for (const notice of browserCacheNotices) messages.push(notice);
        const siteRows = Array.isArray(websites) ? websites : [];
        const waiting =
            listingQueueUnpublished.length +
            listingQueueMembers.length +
            siteRows.length;
        if (listingQueueFetched && !listingQueueError) {
            if (waiting > 0) {
                messages.push({
                    id: "queue",
                    level: "info",
                    text: `${listingQueueUnpublished.length} bejegyzés, ${listingQueueMembers.length} tag és ${siteRows.length} weboldal vár jóváhagyásra.`,
                });
            } else {
                messages.push({
                    id: "queue-ok",
                    level: "success",
                    text: "Nincs jóváhagyásra váró bejegyzés, tag vagy weboldal.",
                });
            }
        }
        for (const site of siteRows) {
            messages.push({
                id: "website-" + site.id,
                level: "info",
                text: `${site.submitter} added ${site.domain}: ${site.title}. ${site.description}`,
                action: "website",
                websiteId: site.id,
            });
        }
        return messages;
    }

    $: dashboardMessages = buildDashboardMessages(
        listingQueueWebsites,
        adminOffline,
        dashboardStatsError,
        settingsLoadError,
        listingQueueError,
        listingQueueUnpublished,
        listingQueueMembers,
        mondasok,
        mondasokTodayCount,
        mondasTodayYmd,
        eventsWithIncompleteDateTime,
        browserCacheNotices,
        apiNotices,
        listingQueueFetched,
    );
    function fetchAttractions() {
        loadData("attractions", (d) => (attractions = d));
    }

    async function fetchCountyRegions() {
        await loadData("/api/counties", (d) => (countiesFromAPI = d));
        await loadData("/api/historical_seats", (d) => (historicalSeatsFromAPI = d));
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
            const res = await apiCall(`/api/admin/counties`, {
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
            const res = await apiCall(`/api/admin/historical_seats`, {
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
            const res = await apiCall(`/api/admin/${endpoint}`, {
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
            const res = await apiCall(`/api/admin/${endpoint}`, {
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
            const res = await apiCall(`/api/admin/${endpoint}?id=${id}`,
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
            const res = await apiCall(
                `/api/proxy?url=${encodeURIComponent(feed.feed_url)}`,
            );
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
    /**
     * @param {File} file
     * @param {boolean} isEdit
     */
    async function uploadEventFeaturedImage(file, isEdit) {
        if (!file) return;
        const fd = new FormData();
        fd.append("file", file);
        if (isEdit && editingEvent?.id) {
            fd.append("event_id", String(editingEvent.id));
        }
        try {
            const res = await apiCall(`/api/admin/event-images`, {
                method: "POST",
                body: fd,
            });
            if (!res.ok) {
                await showAlert(
                    (await res.text()) || "Feltöltés sikertelen.",
                );
                return;
            }
            const data = await res.json();
            const url = data.url || "";
            if (isEdit) {
                editingEvent = { ...editingEvent, featured_image: url };
            } else {
                newEvent = { ...newEvent, featured_image: url };
            }
        } catch (err) {
            await showAlert("Feltöltés hiba: " + err.message);
        }
    }

    async function submitEvent(e) {
        e.preventDefault();
        const rawLid = newEvent.location_id;
        const lid =
            typeof rawLid === "number" && Number.isFinite(rawLid)
                ? rawLid
                : parseInt(String(rawLid ?? "").trim(), 10);
        const dv = parseInt(String(newEvent.default_venue_id || ""), 10);
        const etid = parseInt(String(newEvent.event_type_id || ""), 10);
        const stRaw = newEvent.event_subtype_id;
        let event_subtype_id = null;
        if (stRaw !== "" && stRaw != null && String(stRaw).trim() !== "") {
            const s = parseInt(String(stRaw), 10);
            if (Number.isFinite(s) && s > 0) event_subtype_id = s;
        }
        const payload = {
            location_id: Number.isFinite(lid) && lid > 0 ? lid : 0,
            default_venue_id:
                Number.isFinite(dv) && dv > 0 ? dv : null,
            title: String(newEvent.title ?? "").trim(),
            description: String(newEvent.description ?? ""),
            featured_image: String(newEvent.featured_image ?? "").trim(),
            start_date: normalizeYmdInput(newEvent.start_date),
            end_date: normalizeYmdInput(newEvent.end_date),
            start_time: String(newEvent.start_time ?? "").trim(),
            end_time: String(newEvent.end_time ?? "").trim(),
            event_type_id: etid,
            event_subtype_id,
            access_type: String(newEvent.access_type || "public"),
            organizer: String(newEvent.organizer ?? "").trim(),
            entry_price: String(newEvent.entry_price ?? "").trim(),
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
                    event_type_id:
                        catalogEventTypes.find((t) => t.slug === "cultural")
                            ?.id ?? catalogEventTypes[0]?.id ?? "",
                    event_subtype_id: "",
                    access_type: "public",
                    organizer: "",
                    featured_image: "",
                    entry_price: "",
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
                type: ENTRY_TYPE_SERVICE,
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
            entry_price: ev.entry_price != null ? String(ev.entry_price) : "",
            featured_image: ev.featured_image || "",
            default_venue_id:
                ev.default_venue_id != null && ev.default_venue_id !== ""
                    ? String(ev.default_venue_id)
                    : "",
            event_type_id:
                ev.event_type_id != null && ev.event_type_id !== ""
                    ? String(ev.event_type_id)
                    : "",
            event_subtype_id:
                ev.event_subtype_id != null && ev.event_subtype_id !== ""
                    ? String(ev.event_subtype_id)
                    : "",
            access_type: ev.access_type || "public",
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
            const res = await apiCall(`/api/admin/events/schedule?event_id=${eventId}`,
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
            const res = await apiCall(`/api/admin/events/schedule`, {
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
        const rawLid = editingEvent.location_id;
        const lid =
            typeof rawLid === "number" && Number.isFinite(rawLid)
                ? rawLid
                : parseInt(String(rawLid ?? "").trim(), 10);
        const etidE = parseInt(String(editingEvent.event_type_id || ""), 10);
        const stRawE = editingEvent.event_subtype_id;
        let event_subtype_id_e = null;
        if (
            stRawE !== "" &&
            stRawE != null &&
            String(stRawE).trim() !== ""
        ) {
            const s = parseInt(String(stRawE), 10);
            if (Number.isFinite(s) && s > 0) event_subtype_id_e = s;
        }
        const payload = {
            id: editingEvent.id,
            location_id: Number.isFinite(lid) && lid > 0 ? lid : null,
            default_venue_id:
                Number.isFinite(dv) && dv > 0 ? dv : null,
            title: String(editingEvent.title ?? "").trim(),
            description: String(editingEvent.description ?? ""),
            featured_image: String(editingEvent.featured_image ?? "").trim(),
            start_date: normalizeYmdInput(editingEvent.start_date),
            end_date: normalizeYmdInput(editingEvent.end_date),
            start_time: String(editingEvent.start_time ?? "").trim(),
            end_time: String(editingEvent.end_time ?? "").trim(),
            event_type_id: etidE,
            event_subtype_id: event_subtype_id_e,
            access_type: String(editingEvent.access_type || "public"),
            organizer: String(editingEvent.organizer ?? "").trim(),
            entry_price: String(editingEvent.entry_price ?? "").trim(),
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
            verified: Boolean(newEntry.verified),
            hours: normalizeHours(newEntry.hours),
            delivery_hours: normalizeHours(newEntry.delivery_hours),
            photos: normalizePhotos(newEntry.photos),
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
                    type: ENTRY_TYPE_SERVICE,
                    languages: ["HU"],
                    tags: "",
                    verified: false,
                    hours: emptyWeekHours(),
                    delivery_hours: emptyWeekHours(),
                    photos: emptyPhotos(),
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
            verified: Boolean(entry.verified),
            hours: normalizeHours(entry.hours),
            delivery_hours: normalizeHours(entry.delivery_hours),
            photos: normalizePhotos(entry.photos),
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
            verified: Boolean(editingEntry.verified),
            hours: normalizeHours(editingEntry.hours),
            delivery_hours: normalizeHours(editingEntry.delivery_hours),
            photos: normalizePhotos(editingEntry.photos),
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
            const res = await apiCall(`/api/admin/attractions?id=${id}`, { method: "DELETE" });
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
                {#if !authReady}
                    <h2>Adminisztráció</h2>
                    <p>Ellenőrzés…</p>
                {:else if authDenied}
                    <h2>Nincs jogosultság</h2>
                    <p>
                        Ez a Google-fiók be van jelentkezve, de nem
                        adminisztrátor.
                    </p>
                    <button
                        type="button"
                        class="admin-submit-btn"
                        on:click={logout}>Kijelentkezés</button
                    >
                {:else}
                    <h2>Adminisztráció Belépés</h2>
                    <p>
                        Jelentkezz be Google-fiókkal. Csak az admin e-mail
                        érheti el ezt a felületet.
                    </p>
                    {#if googleClientId}
                        {#key googleClientId}
                            <GoogleSignIn
                                clientId={googleClientId}
                                onSignedIn={onGoogleSignedIn}
                            />
                        {/key}
                    {:else if configUnreachable}
                        <p>Az API nem elérhető, ezért a belépés most nem lehetséges.</p>
                    {:else}
                        <p>
                            A Google belépés nincs beállítva
                            (GOOGLE_CLIENT_ID).
                        </p>
                    {/if}
                {/if}
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
                title="Dashboard"
            >
                <AdminNavIcon name="dashboard" />
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

            <hr class="admin-sidebar-sep" aria-hidden="true" />

            <button
                class="admin-sidebar-btn {activeTab === 'entries'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("entries")}
                title="Index"
            >
                <AdminNavIcon name="entries" />
            </button>

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

            <hr class="admin-sidebar-sep" aria-hidden="true" />

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
                class="admin-sidebar-btn {activeTab === 'counties'
                    ? 'active'
                    : ''}"
                on:click={() => goToAdminTab("counties")}
                title="Megyék"
            >
                <AdminNavIcon name="counties" />
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
                class="admin-sidebar-btn {activeTab === 'pages' ? 'active' : ''}"
                on:click={() => goToAdminTab('pages')}
                title="Oldalak"
            >
                <AdminNavIcon name="pages" />
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'page_faq' ? 'active' : ''}"
                on:click={() => goToAdminTab('page_faq')}
                title="GYIK"
            >
                <AdminNavIcon name="page_faq" />
            </button>

            <button
                class="admin-sidebar-btn {activeTab === 'weather_translations' ? 'active' : ''}"
                on:click={() => goToAdminTab('weather_translations')}
                title="Időjárás fordítások"
            >
                <AdminNavIcon name="weather_translations" />
            </button>

            <hr class="admin-sidebar-sep" aria-hidden="true" />

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
            <header class="admin-header">
                <div class="admin-header-text">
                    <h1 class="admin-page-title">{adminPageHead.title}</h1>
                    {#if adminPageHead.greeting}
                        <p class="admin-page-greeting">{adminPageHead.greeting}</p>
                    {/if}
                </div>
                <button class="btn-logout" on:click={logout}
                    >Kijelentkezés</button
                >
            </header>

            <div class="admin-container w-full">
                {#if activeTab === "welcome"}
                    <section aria-labelledby="admin-messages-title">
                        <h3 id="admin-messages-title">Üzenetek</h3>
                        {#each dashboardMessages as msg (msg.id)}
                            <div
                                class="admin-alert admin-alert--{msg.level}"
                                role={msg.level === "error" || msg.level === "warning" ? "alert" : "status"}
                            >
                                {msg.text}
                                {#if msg.action === "retry-stats"}
                                    <button type="button" class="admin-alert__btn" on:click={fetchDashboardStats}>Újra</button>
                                {:else if msg.action === "retry-queue"}
                                    <button type="button" class="admin-alert__btn" on:click={fetchListingQueue}>Újra</button>
                                {:else if msg.action === "open" && msg.tab}
                                    <button type="button" class="admin-alert__btn" on:click={() => goToAdminTab(msg.tab)}>Megnyitás</button>
                                {:else if msg.action === "cache-refresh"}
                                    <button
                                        type="button"
                                        class="admin-alert__btn"
                                        disabled
                                        title="A böngésző-mentés törlése és újratöltése később lesz bekötve."
                                    >Frissítés</button>
                                {:else if msg.action === "website"}
                                    <button
                                        type="button"
                                        class="admin-alert__btn"
                                        on:click={() => reviewWebsite(msg.websiteId, "approve")}
                                    >Approve</button>
                                    <button
                                        type="button"
                                        class="admin-alert__btn"
                                        on:click={() => reviewWebsite(msg.websiteId, "reject")}
                                    >Reject</button>
                                    <button
                                        type="button"
                                        class="admin-alert__btn"
                                        on:click={() => reviewWebsite(msg.websiteId, "ban")}
                                    >Ban User</button>
                                {/if}
                            </div>
                        {/each}
                        {#if listingQueueFetched && !listingQueueError && (listingQueueUnpublished.length > 0 || listingQueueMembers.length > 0)}
                            {#if listingQueueUnpublished.length > 0}
                                <h4>Közzétételre váró bejegyzések</h4>
                                <div class="admin-table-wrapper">
                                    <table class="admin-table">
                                        <thead>
                                            <tr>
                                                <th>Név</th>
                                                <th>Tulajdonos</th>
                                                <th class="admin-table-col--action">Művelet</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {#each listingQueueUnpublished as row}
                                                <tr>
                                                    <td>{row.name}</td>
                                                    <td>{row.owner_email || "—"}</td>
                                                    <td class="admin-table-col--action">
                                                        <button
                                                            type="button"
                                                            class="admin-submit-btn"
                                                            on:click={() => publishListingQueueEntry(row.id)}
                                                        >Közzététel</button>
                                                    </td>
                                                </tr>
                                            {/each}
                                        </tbody>
                                    </table>
                                </div>
                            {/if}
                            {#if listingQueueMembers.length > 0}
                                <h4>Tag-jelentkezések</h4>
                                <div class="admin-table-wrapper">
                                    <table class="admin-table">
                                        <thead>
                                            <tr>
                                                <th>Bejegyzés</th>
                                                <th>Felhasználó</th>
                                                <th class="admin-table-col--action">Művelet</th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {#each listingQueueMembers as row}
                                                <tr>
                                                    <td>{row.entry_name}</td>
                                                    <td>{row.email}</td>
                                                    <td class="admin-table-col--action">
                                                        <button
                                                            type="button"
                                                            class="admin-submit-btn"
                                                            on:click={() => approveListingQueueMember(row.entry_id, row.user_id)}
                                                        >Elfogadás</button>
                                                        <button
                                                            type="button"
                                                            class="btn-logout"
                                                            on:click={() => rejectListingQueueMember(row.entry_id, row.user_id)}
                                                        >Elutasítás</button>
                                                    </td>
                                                </tr>
                                            {/each}
                                        </tbody>
                                    </table>
                                </div>
                            {/if}
                        {/if}
                    </section>
                    <div class="admin-welcome" role="navigation" aria-label="Admin sections">
                        <div class="admin-welcome-grid">
                            {#each ADMIN_WELCOME_ITEMS as item}
                                <button
                                    type="button"
                                    class="admin-welcome-card"
                                    on:click={() => goToAdminTab(item.id)}
                                    title={item.title}
                                    aria-label={item.title}
                                >
                                    <div class="admin-welcome-card-body">
                                        <AdminNavIcon name={item.id} size={56} />
                                    </div>
                                    <div class="admin-welcome-card-footer">
                                        <span class="admin-welcome-card-footer-left"
                                            title={dashboardStatsFetched && dashboardStats[item.id] != null
                                                ? "Rekordok száma az adatbázisban"
                                                : "Betöltés…"}
                                            >{dashboardStatsFetched && dashboardStats[item.id] != null
                                                ? dashboardStats[item.id]
                                                : "…"}</span
                                        >
                                        <span class="admin-welcome-card-footer-right" title={item.title}
                                            >{item.title}</span
                                        >
                                    </div>
                                </button>
                            {/each}
                        </div>
                    </div>
                    <section class="admin-cache-panel" aria-labelledby="admin-cache-title">
                        <h3 id="admin-cache-title">Gyorsítótár</h3>
                        <p>
                            A kártyák számai nincsenek gyorsítótárazva. Minden megnyitáskor a szerver számolja a táblákat.
                            Az alábbi mentések csak ebben a böngészőben vannak, a nyilvános oldalak használják.
                        </p>
                        <ul class="admin-cache-list">
                            {#each browserCacheRows as row (row.name)}
                                <li>
                                    <span class="admin-cache-name">{row.name}</span>
                                    <span class="admin-cache-detail">{row.detail}</span>
                                </li>
                            {/each}
                        </ul>
                    </section>
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
                    {#if mondasok.length > 0 && mondasokTodayCount === 0}
                        <div class="admin-alert admin-alert--warning" role="status">
                            Ma ({mondasTodayYmd}) nincs beütemezett mondás, ezért a kezdőlapon
                            a mondás-blokk rejtve marad. Állítsd egy idézet
                            <strong>megjelenés napját</strong> a mai dátumra.
                        </div>
                    {/if}
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"
                            ><span>Új mondás hozzáadása</span><AdminPlusIcon /></summary
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
                            <div class="admin-date-field">
                                <input
                                    id="mondas_day"
                                    name="display_date"
                                    type="date"
                                    bind:value={newMondas.display_date}
                                    required
                                />
                                <button
                                    type="button"
                                    class="btn btn-sm"
                                    on:click={() =>
                                        (newMondas.display_date = localISODate())}
                                    >Mai nap</button
                                >
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
                                id="search_mondasok"
                                name="search_mondasok"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchMondasok}
                                on:input={() => (pageMondasok = 1)}
                                placeholder="Szöveg vagy ID…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgMondasok.total}
                        page={pgMondasok.page}
                        totalPages={pgMondasok.totalPages}
                        from={pgMondasok.from}
                        to={pgMondasok.to}
                        on:prev={() =>
                            (pageMondasok = Math.max(1, pageMondasok - 1))}
                        on:next={() =>
                            (pageMondasok = Math.min(
                                pgMondasok.totalPages,
                                pageMondasok + 1,
                            ))}
                    />
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
                                {#each pgMondasok.rows as m (m.id)}
                                    {@const isToday =
                                        normalizeYmdInput(m.display_date) ===
                                        mondasTodayYmd}
                                    <tr class:admin-row-today={isToday}>
                                        <td>{m.id}</td>
                                        <td>
                                            {m.display_date ?? "—"}
                                            {#if isToday}
                                                <span class="admin-date-today-badge"
                                                    >ma</span
                                                >
                                            {/if}
                                        </td>
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
                    <AdminPaginationBar
                        total={pgMondasok.total}
                        page={pgMondasok.page}
                        totalPages={pgMondasok.totalPages}
                        from={pgMondasok.from}
                        to={pgMondasok.to}
                        on:prev={() =>
                            (pageMondasok = Math.max(1, pageMondasok - 1))}
                        on:next={() =>
                            (pageMondasok = Math.min(
                                pgMondasok.totalPages,
                                pageMondasok + 1,
                            ))}
                    />
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
                            ><span>Új gyorslink hozzáadása</span><AdminPlusIcon /></summary
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

                            <label for="link_url">Weblap URL</label>
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
                                on:input={() => (pageQuickLinks = 1)}
                                placeholder="Cím, URL…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgQuickLinks.total}
                        page={pgQuickLinks.page}
                        totalPages={pgQuickLinks.totalPages}
                        from={pgQuickLinks.from}
                        to={pgQuickLinks.to}
                        on:prev={() =>
                            (pageQuickLinks = Math.max(1, pageQuickLinks - 1))}
                        on:next={() =>
                            (pageQuickLinks = Math.min(
                                pgQuickLinks.totalPages,
                                pageQuickLinks + 1,
                            ))}
                    />
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Szín</th>
                                    <th>Cím</th>
                                    <th>Weblap URL</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each pgQuickLinks.rows as q}
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
                    <AdminPaginationBar
                        total={pgQuickLinks.total}
                        page={pgQuickLinks.page}
                        totalPages={pgQuickLinks.totalPages}
                        from={pgQuickLinks.from}
                        to={pgQuickLinks.to}
                        on:prev={() =>
                            (pageQuickLinks = Math.max(1, pageQuickLinks - 1))}
                        on:next={() =>
                            (pageQuickLinks = Math.min(
                                pgQuickLinks.totalPages,
                                pageQuickLinks + 1,
                            ))}
                    />
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
                            ><span>Új RSS hírfolyam hozzáadása</span><AdminPlusIcon /></summary
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
                                on:input={() => (pageNewsFeeds = 1)}
                                placeholder="Név, URL…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgNewsFeeds.total}
                        page={pgNewsFeeds.page}
                        totalPages={pgNewsFeeds.totalPages}
                        from={pgNewsFeeds.from}
                        to={pgNewsFeeds.to}
                        on:prev={() =>
                            (pageNewsFeeds = Math.max(1, pageNewsFeeds - 1))}
                        on:next={() =>
                            (pageNewsFeeds = Math.min(
                                pgNewsFeeds.totalPages,
                                pageNewsFeeds + 1,
                            ))}
                    />
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
                                {#each pgNewsFeeds.rows as nf}
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
                    <AdminPaginationBar
                        total={pgNewsFeeds.total}
                        page={pgNewsFeeds.page}
                        totalPages={pgNewsFeeds.totalPages}
                        from={pgNewsFeeds.from}
                        to={pgNewsFeeds.to}
                        on:prev={() =>
                            (pageNewsFeeds = Math.max(1, pageNewsFeeds - 1))}
                        on:next={() =>
                            (pageNewsFeeds = Math.min(
                                pgNewsFeeds.totalPages,
                                pageNewsFeeds + 1,
                            ))}
                    />
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
                        <summary class="admin-create-summary"><span>Új település</span><AdminPlusIcon /></summary>
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
                            {#each settlementLocationTypes as t}<option value={t.slug}
                                    >{t.label_hu}</option
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
                                on:input={() => (pageLocations = 1)}
                                placeholder="Név, megye, típus, ir.sz…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgLocations.total}
                        page={pgLocations.page}
                        totalPages={pgLocations.totalPages}
                        from={pgLocations.from}
                        to={pgLocations.to}
                        on:prev={() =>
                            (pageLocations = Math.max(1, pageLocations - 1))}
                        on:next={() =>
                            (pageLocations = Math.min(
                                pgLocations.totalPages,
                                pageLocations + 1,
                            ))}
                    />
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
                                {#each pgLocations.rows as l}
                                    <tr>
                                        <td>{l.id}</td>
                                        <td>{l.name}</td>
                                        <td>{l.name_ro || "-"}</td>
                                        <td>{l.name_de || "-"}</td>
                                        <td>{l.county || "-"}</td>
                                        <td>
                                            <span class="badge"
                                                >{settlementTypeLabel(l.type)}</span
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
                    <AdminPaginationBar
                        total={pgLocations.total}
                        page={pgLocations.page}
                        totalPages={pgLocations.totalPages}
                        from={pgLocations.from}
                        to={pgLocations.to}
                        on:prev={() =>
                            (pageLocations = Math.max(1, pageLocations - 1))}
                        on:next={() =>
                            (pageLocations = Math.min(
                                pgLocations.totalPages,
                                pageLocations + 1,
                            ))}
                    />

                    <h3 class="admin-subsection-title">Településtípusok</h3>
                    <p class="admin-form-hint" style="margin: 0 0 0.75rem">
                        A típus <strong>slug</strong>ja szerepel a település rekordban; a megnevezés a listákban és űrlapokban
                        jelenik meg. Új slug: opcionálisan megadható; üresen a megnevezésből képződik.
                    </p>
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"
                            ><span>Új településtípus</span><AdminPlusIcon /></summary
                        >
                        <form
                            class="admin-form admin-create-form"
                            on:submit|preventDefault={submitNewSettlementLocationType}
                        >
                            <label for="slt-slug">Slug (opcionális)</label>
                            <input
                                id="slt-slug"
                                type="text"
                                bind:value={newSettlementLocationType.slug}
                                placeholder="pl. varos — üresen automatikus"
                            />
                            <label for="slt-label">Megnevezés (HU) *</label>
                            <input
                                id="slt-label"
                                type="text"
                                bind:value={newSettlementLocationType.label_hu}
                                required
                            />
                            <label for="slt-order">Sorrend</label>
                            <input
                                id="slt-order"
                                type="number"
                                bind:value={newSettlementLocationType.sort_order}
                            />
                            <button type="submit" class="admin-submit-btn"
                                >Típus hozzáadása</button
                            >
                        </form>
                    </details>

                    {#if editingSettlementLocationType}
                        <form
                            class="admin-form admin-venues-type-edit"
                            on:submit|preventDefault={saveEditSettlementLocationType}
                        >
                            <p class="admin-form-hint">
                                Slug (azonosító, nem módosítható):
                                <code>{editingSettlementLocationType.slug}</code>
                            </p>
                            <label for="slt-edit-label">Megnevezés (HU)</label>
                            <input
                                id="slt-edit-label"
                                type="text"
                                bind:value={editingSettlementLocationType.label_hu}
                                required
                            />
                            <label for="slt-edit-order">Sorrend</label>
                            <input
                                id="slt-edit-order"
                                type="number"
                                bind:value={editingSettlementLocationType.sort_order}
                            />
                            <div class="flex gap-md">
                                <button type="submit" class="admin-submit-btn"
                                    >Mentés</button
                                >
                                <button
                                    type="button"
                                    class="btn-update"
                                    on:click={cancelEditSettlementLocationType}
                                    >Mégse</button
                                >
                            </div>
                        </form>
                    {/if}

                    <div class="admin-table-wrapper">
                        <table class="admin-table admin-table--compact">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Slug</th>
                                    <th>Megnevezés (HU)</th>
                                    <th>Sorrend</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each settlementLocationTypes as t}
                                    <tr>
                                        <td>{t.id}</td>
                                        <td><code>{t.slug}</code></td>
                                        <td>{t.label_hu}</td>
                                        <td>{t.sort_order}</td>
                                        <td>
                                            <button
                                                type="button"
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditSettlementLocationType(t)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                type="button"
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteSettlementLocationTypeRow(t.id)}
                                                >Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="6"
                                            >Nincs típus (futtasd a backend migrációt).</td
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
                            Rendezvényhelyszínek (csarnokok, terek, pályák) településhez kötve. Előbb add meg az
                            <strong>új helyszínt</strong> (ha szükséges), alatta a <strong>helyszíntípusok</strong>
                            katalógusa (slug a megnevezésből, sorrend), majd az összes helyszín listája — az
                            <strong>Események</strong> napi programjában itt választhatók.
                        </p>
                    {/if}

                    <h3 class="admin-subsection-title">Helyszínek</h3>
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"
                            ><span>Új helyszín hozzáadása</span><AdminPlusIcon /></summary
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

                    <h3 class="admin-subsection-title">Helyszíntípusok</h3>
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"
                            ><span>Új helyszíntípus</span><AdminPlusIcon /></summary
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
                                on:input={() => (pageVenueTypes = 1)}
                                placeholder="Slug, megnevezés…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgVenueTypes.total}
                        page={pgVenueTypes.page}
                        totalPages={pgVenueTypes.totalPages}
                        from={pgVenueTypes.from}
                        to={pgVenueTypes.to}
                        on:prev={() =>
                            (pageVenueTypes = Math.max(1, pageVenueTypes - 1))}
                        on:next={() =>
                            (pageVenueTypes = Math.min(
                                pgVenueTypes.totalPages,
                                pageVenueTypes + 1,
                            ))}
                    />
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
                                {#each pgVenueTypes.rows as t}
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
                    <AdminPaginationBar
                        total={pgVenueTypes.total}
                        page={pgVenueTypes.page}
                        totalPages={pgVenueTypes.totalPages}
                        from={pgVenueTypes.from}
                        to={pgVenueTypes.to}
                        on:prev={() =>
                            (pageVenueTypes = Math.max(1, pageVenueTypes - 1))}
                        on:next={() =>
                            (pageVenueTypes = Math.min(
                                pgVenueTypes.totalPages,
                                pageVenueTypes + 1,
                            ))}
                    />

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés (helyszínek)
                            <input
                                id="search_venues"
                                name="search_venues"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchVenues}
                                on:input={() => (pageVenues = 1)}
                                placeholder="Név, település, típus…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgVenuesCatalog.total}
                        page={pgVenuesCatalog.page}
                        totalPages={pgVenuesCatalog.totalPages}
                        from={pgVenuesCatalog.from}
                        to={pgVenuesCatalog.to}
                        on:prev={() =>
                            (pageVenues = Math.max(1, pageVenues - 1))}
                        on:next={() =>
                            (pageVenues = Math.min(
                                pgVenuesCatalog.totalPages,
                                pageVenues + 1,
                            ))}
                    />
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
                                {#each pgVenuesCatalog.rows as v}
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
                    <AdminPaginationBar
                        total={pgVenuesCatalog.total}
                        page={pgVenuesCatalog.page}
                        totalPages={pgVenuesCatalog.totalPages}
                        from={pgVenuesCatalog.from}
                        to={pgVenuesCatalog.to}
                        on:prev={() =>
                            (pageVenues = Math.max(1, pageVenues - 1))}
                        on:next={() =>
                            (pageVenues = Math.min(
                                pgVenuesCatalog.totalPages,
                                pageVenues + 1,
                            ))}
                    />
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
                        <summary class="admin-create-summary"><span>Új esemény</span><AdminPlusIcon /></summary>
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

                        <div class="admin-event-image-block">
                            <label for="event_featured_upload">Kiemelt kép</label>
                            {#if newEvent.featured_image}
                                <img
                                    class="admin-event-image-preview"
                                    src={absoluteMediaUrl(
                                        newEvent.featured_image,
                                        getApiBase(),
                                    )}
                                    alt=""
                                />
                            {/if}
                            <input
                                id="event_featured_upload"
                                type="file"
                                accept="image/jpeg,image/png,image/webp,image/gif"
                                on:change={(e) => {
                                    const f = e.target.files?.[0];
                                    if (f) uploadEventFeaturedImage(f, false);
                                    e.target.value = "";
                                }}
                            />
                            <label for="event_featured_url" class="admin-sublabel"
                                >Vagy kép URL (külső)</label
                            >
                            <input
                                id="event_featured_url"
                                type="url"
                                bind:value={newEvent.featured_image}
                                placeholder="https://…"
                            />
                            {#if newEvent.featured_image}
                                <button
                                    type="button"
                                    class="btn-update"
                                    style="align-self: flex-start"
                                    on:click={() =>
                                        (newEvent.featured_image = "")}
                                    >Kép törlése</button
                                >
                            {/if}
                        </div>

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

                        <label for="event_type_id"
                            >Eseménytípus <span class="admin-req" title="Kötelező">*</span></label
                        >
                        <select
                            id="event_type_id"
                            bind:value={newEvent.event_type_id}
                            on:change={() => (newEvent.event_subtype_id = "")}
                            required
                        >
                            <option value="">— válassz —</option>
                            {#each [...catalogEventTypes].sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0) || String(a.label_hu).localeCompare(String(b.label_hu), "hu")) as t}
                                <option value={String(t.id)}
                                    >{t.label_hu} ({t.slug})</option
                                >
                            {/each}
                        </select>

                        <label for="event_subtype_id">Altípus (opcionális)</label>
                        <select
                            id="event_subtype_id"
                            bind:value={newEvent.event_subtype_id}
                        >
                            <option value="">— nincs —</option>
                            {#each subtypesForNewEvent as s}
                                <option value={String(s.id)}
                                    >{s.label_hu} ({s.slug})</option
                                >
                            {/each}
                        </select>

                        <label for="event_access_type">Hozzáférés</label>
                        <select
                            id="event_access_type"
                            bind:value={newEvent.access_type}
                        >
                            <option value="public">{ACCESS_TYPE_LABELS.public}</option>
                            <option value="members_only">{ACCESS_TYPE_LABELS.members_only}</option>
                            <option value="invitation_only">{ACCESS_TYPE_LABELS.invitation_only}</option>
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

                        <label for="event_entry_price">Belépő / jegyár (opcionális)</label>
                        <input
                            id="event_entry_price"
                            name="entry_price"
                            type="text"
                            bind:value={newEvent.entry_price}
                            placeholder="pl. 99 RON, 15 EUR, ingyenes"
                            maxlength="128"
                            autocomplete="off"
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
                                id="search_events"
                                name="search_events"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchEvents}
                                on:input={() => (pageEvents = 1)}
                                placeholder="Cím, szervező, helyszín…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgEvents.total}
                        page={pgEvents.page}
                        totalPages={pgEvents.totalPages}
                        from={pgEvents.from}
                        to={pgEvents.to}
                        on:prev={() =>
                            (pageEvents = Math.max(1, pageEvents - 1))}
                        on:next={() =>
                            (pageEvents = Math.min(
                                pgEvents.totalPages,
                                pageEvents + 1,
                            ))}
                    />
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Kép</th>
                                    <th>Cím</th>
                                    <th>Kezdés</th>
                                    <th>Befejezés</th>
                                    <th>Típus</th>
                                    <th>Altípus</th>
                                    <th>Hozzáférés</th>
                                    <th>Település</th>
                                    <th>Alapért. helyszín</th>
                                    <th>Szervező</th>
                                    <th>Belépő</th>
                                    <th>Leírás</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each pgEvents.rows as e}
                                    <tr
                                        class:admin-row-warn={!eventDateTimeComplete(
                                            e,
                                        )}
                                    >
                                        <td class="admin-event-thumb-cell">
                                            {#if e.featured_image}
                                                <img
                                                    src={absoluteMediaUrl(
                                                        e.featured_image,
                                                        getApiBase(),
                                                    )}
                                                    alt=""
                                                />
                                            {:else}
                                                <span class="admin-thumb-empty">—</span>
                                            {/if}
                                        </td>
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
                                        <td>{eventTypeLabelFromCatalog(e.event_type)}</td>
                                        <td>{eventSubtypeLabelFromCatalog(Number(e.event_type_id), e.event_subtype || "")}</td>
                                        <td>{accessTypeLabel(e.access_type)}</td>
                                        <td>{getLocationName(e.location_id)}</td>
                                        <td class="admin-table-cell-preview" title={e.default_venue_name || ""}>{contentPreview(e.default_venue_name || "")}</td>
                                        <td>{e.organizer || "—"}</td>
                                        <td>{e.entry_price && String(e.entry_price).trim() ? e.entry_price : "—"}</td>
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
                                        ><td colspan="14"
                                            >Nincsenek események.</td
                                        ></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                    <AdminPaginationBar
                        total={pgEvents.total}
                        page={pgEvents.page}
                        totalPages={pgEvents.totalPages}
                        from={pgEvents.from}
                        to={pgEvents.to}
                        on:prev={() =>
                            (pageEvents = Math.max(1, pageEvents - 1))}
                        on:next={() =>
                            (pageEvents = Math.min(
                                pgEvents.totalPages,
                                pageEvents + 1,
                            ))}
                    />

                    <h3 class="admin-subsection-title">Eseménytípusok (katalógus)</h3>
                    <p class="admin-hint">
                        A típusok és altípusok itt szerkeszthetők. Az események <code>event_type_id</code> értéke ezekre
                        mutat.
                    </p>
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"><span>Új eseménytípus</span><AdminPlusIcon /></summary>
                        <form
                            class="admin-form admin-create-form"
                            on:submit|preventDefault={submitCatalogEventType}
                        >
                            <label for="cet-slug">Slug (URL, egyedi, pl. <code>sports</code>) *</label>
                            <input
                                id="cet-slug"
                                type="text"
                                bind:value={newCatalogEventType.slug}
                                required
                                placeholder="pl. workshop"
                            />
                            <label for="cet-label">Megnevezés (HU) *</label>
                            <input
                                id="cet-label"
                                type="text"
                                bind:value={newCatalogEventType.label_hu}
                                required
                            />
                            <label for="cet-sort">Sorrend</label>
                            <input
                                id="cet-sort"
                                type="number"
                                bind:value={newCatalogEventType.sort_order}
                            />
                            <button type="submit" class="admin-submit-btn"
                                >Típus hozzáadása</button
                            >
                        </form>
                    </details>

                    {#if editingCatalogEventType}
                        <form
                            class="admin-form admin-venues-type-edit"
                            on:submit|preventDefault={saveEditCatalogEventType}
                        >
                            <p class="admin-form-hint">
                                Slug: <code>{editingCatalogEventType.slug}</code> (nem változtatható)
                            </p>
                            <label for="cet-edit-label">Megnevezés (HU)</label>
                            <input
                                id="cet-edit-label"
                                type="text"
                                bind:value={editingCatalogEventType.label_hu}
                                required
                            />
                            <label for="cet-edit-sort">Sorrend</label>
                            <input
                                id="cet-edit-sort"
                                type="number"
                                bind:value={editingCatalogEventType.sort_order}
                            />
                            <div class="flex gap-md">
                                <button type="submit" class="admin-submit-btn"
                                    >Mentés</button
                                >
                                <button
                                    type="button"
                                    class="btn-update"
                                    on:click={cancelEditCatalogEventType}>Mégse</button
                                >
                            </div>
                        </form>
                    {/if}

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés (típusok)
                            <input
                                type="search"
                                class="admin-search-input"
                                bind:value={searchCatalogTypes}
                                on:input={() => (pageCatalogTypes = 1)}
                                placeholder="Slug, név…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgCatalogTypes.total}
                        page={pgCatalogTypes.page}
                        totalPages={pgCatalogTypes.totalPages}
                        from={pgCatalogTypes.from}
                        to={pgCatalogTypes.to}
                        on:prev={() =>
                            (pageCatalogTypes = Math.max(1, pageCatalogTypes - 1))}
                        on:next={() =>
                            (pageCatalogTypes = Math.min(
                                pgCatalogTypes.totalPages,
                                pageCatalogTypes + 1,
                            ))}
                    />
                    <div class="admin-table-wrapper">
                        <table class="admin-table admin-table--compact">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Slug</th>
                                    <th>Megnevezés (HU)</th>
                                    <th>Sorrend</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each pgCatalogTypes.rows as t}
                                    <tr>
                                        <td>{t.id}</td>
                                        <td><code>{t.slug}</code></td>
                                        <td>{t.label_hu}</td>
                                        <td>{t.sort_order}</td>
                                        <td>
                                            <button
                                                type="button"
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditCatalogEventType(t)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                type="button"
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteCatalogEventTypeRow(t.id)}
                                                >Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="6">Nincs típus.</td></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                    <AdminPaginationBar
                        total={pgCatalogTypes.total}
                        page={pgCatalogTypes.page}
                        totalPages={pgCatalogTypes.totalPages}
                        from={pgCatalogTypes.from}
                        to={pgCatalogTypes.to}
                        on:prev={() =>
                            (pageCatalogTypes = Math.max(1, pageCatalogTypes - 1))}
                        on:next={() =>
                            (pageCatalogTypes = Math.min(
                                pgCatalogTypes.totalPages,
                                pageCatalogTypes + 1,
                            ))}
                    />

                    <h3 class="admin-subsection-title">Esemény altípusok</h3>
                    <details class="admin-create-panel">
                        <summary class="admin-create-summary"><span>Új altípus</span><AdminPlusIcon /></summary>
                        <form
                            class="admin-form admin-create-form"
                            on:submit|preventDefault={submitCatalogEventSubtype}
                        >
                            <label for="ces-type">Főtípus *</label>
                            <select
                                id="ces-type"
                                bind:value={newCatalogEventSubtype.event_type_id}
                                required
                            >
                                <option value="">— válassz —</option>
                                {#each [...catalogEventTypes].sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0) || String(a.label_hu).localeCompare(String(b.label_hu), "hu")) as ct}
                                    <option value={String(ct.id)}
                                        >{ct.label_hu} ({ct.slug})</option
                                    >
                                {/each}
                            </select>
                            <label for="ces-slug">Slug *</label>
                            <input
                                id="ces-slug"
                                type="text"
                                bind:value={newCatalogEventSubtype.slug}
                                required
                                placeholder="pl. hockey"
                            />
                            <label for="ces-label">Megnevezés (HU) *</label>
                            <input
                                id="ces-label"
                                type="text"
                                bind:value={newCatalogEventSubtype.label_hu}
                                required
                            />
                            <label for="ces-sort">Sorrend</label>
                            <input
                                id="ces-sort"
                                type="number"
                                bind:value={newCatalogEventSubtype.sort_order}
                            />
                            <button type="submit" class="admin-submit-btn"
                                >Altípus hozzáadása</button
                            >
                        </form>
                    </details>

                    {#if editingCatalogEventSubtype}
                        <form
                            class="admin-form admin-venues-type-edit"
                            on:submit|preventDefault={saveEditCatalogEventSubtype}
                        >
                            <p class="admin-form-hint">
                                Slug: <code>{editingCatalogEventSubtype.slug}</code> · főtípus ID:
                                {editingCatalogEventSubtype.event_type_id}
                            </p>
                            <label for="ces-edit-label">Megnevezés (HU)</label>
                            <input
                                id="ces-edit-label"
                                type="text"
                                bind:value={editingCatalogEventSubtype.label_hu}
                                required
                            />
                            <label for="ces-edit-sort">Sorrend</label>
                            <input
                                id="ces-edit-sort"
                                type="number"
                                bind:value={editingCatalogEventSubtype.sort_order}
                            />
                            <div class="flex gap-md">
                                <button type="submit" class="admin-submit-btn"
                                    >Mentés</button
                                >
                                <button
                                    type="button"
                                    class="btn-update"
                                    on:click={cancelEditCatalogEventSubtype}>Mégse</button
                                >
                            </div>
                        </form>
                    {/if}

                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés (altípusok)
                            <input
                                type="search"
                                class="admin-search-input"
                                bind:value={searchCatalogSubtypes}
                                on:input={() => (pageCatalogSubtypes = 1)}
                                placeholder="Slug, név, típus ID…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgCatalogSubtypes.total}
                        page={pgCatalogSubtypes.page}
                        totalPages={pgCatalogSubtypes.totalPages}
                        from={pgCatalogSubtypes.from}
                        to={pgCatalogSubtypes.to}
                        on:prev={() =>
                            (pageCatalogSubtypes = Math.max(
                                1,
                                pageCatalogSubtypes - 1,
                            ))}
                        on:next={() =>
                            (pageCatalogSubtypes = Math.min(
                                pgCatalogSubtypes.totalPages,
                                pageCatalogSubtypes + 1,
                            ))}
                    />
                    <div class="admin-table-wrapper">
                        <table class="admin-table admin-table--compact">
                            <thead>
                                <tr>
                                    <th>ID</th>
                                    <th>Főtípus ID</th>
                                    <th>Slug</th>
                                    <th>Megnevezés (HU)</th>
                                    <th>Sorrend</th>
                                    <th class="admin-table-col--action">Szerk.</th>
                                    <th class="admin-table-col--action">Törlés</th>
                                </tr>
                            </thead>
                            <tbody>
                                {#each pgCatalogSubtypes.rows as s}
                                    <tr>
                                        <td>{s.id}</td>
                                        <td>{s.event_type_id}</td>
                                        <td><code>{s.slug}</code></td>
                                        <td>{s.label_hu}</td>
                                        <td>{s.sort_order}</td>
                                        <td>
                                            <button
                                                type="button"
                                                class="btn-update"
                                                on:click={() =>
                                                    startEditCatalogEventSubtype(s)}
                                                >Szerk.</button
                                            >
                                        </td>
                                        <td>
                                            <button
                                                type="button"
                                                class="btn-delete"
                                                on:click={() =>
                                                    deleteCatalogEventSubtypeRow(s.id)}
                                                >Törlés</button
                                            >
                                        </td>
                                    </tr>
                                {:else}
                                    <tr
                                        ><td colspan="7">Nincs altípus.</td></tr
                                    >
                                {/each}
                            </tbody>
                        </table>
                    </div>
                    <AdminPaginationBar
                        total={pgCatalogSubtypes.total}
                        page={pgCatalogSubtypes.page}
                        totalPages={pgCatalogSubtypes.totalPages}
                        from={pgCatalogSubtypes.from}
                        to={pgCatalogSubtypes.to}
                        on:prev={() =>
                            (pageCatalogSubtypes = Math.max(
                                1,
                                pageCatalogSubtypes - 1,
                            ))}
                        on:next={() =>
                            (pageCatalogSubtypes = Math.min(
                                pgCatalogSubtypes.totalPages,
                                pageCatalogSubtypes + 1,
                            ))}
                    />
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
                        <summary class="admin-create-summary"><span>Új kategória</span><AdminPlusIcon /></summary>
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
                        <label class="admin-search-label">
                            <span class="admin-search-heading">
                                Keresés
                            </span>
                            <input
                                id="search_entry_categories"
                                name="search_entry_categories"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchEntryCategories}
                                on:input={() => (pageEntryCategories = 1)}
                                placeholder="Név, ID…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgEntryCategories.total}
                        page={pgEntryCategories.page}
                        totalPages={pgEntryCategories.totalPages}
                        from={pgEntryCategories.from}
                        to={pgEntryCategories.to}
                        on:prev={() =>
                            (pageEntryCategories = Math.max(
                                1,
                                pageEntryCategories - 1,
                            ))}
                        on:next={() =>
                            (pageEntryCategories = Math.min(
                                pgEntryCategories.totalPages,
                                pageEntryCategories + 1,
                            ))}
                    />
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
                                {#each pgEntryCategories.rows as cat}
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
                    <AdminPaginationBar
                        total={pgEntryCategories.total}
                        page={pgEntryCategories.page}
                        totalPages={pgEntryCategories.totalPages}
                        from={pgEntryCategories.from}
                        to={pgEntryCategories.to}
                        on:prev={() =>
                            (pageEntryCategories = Math.max(
                                1,
                                pageEntryCategories - 1,
                            ))}
                        on:next={() =>
                            (pageEntryCategories = Math.min(
                                pgEntryCategories.totalPages,
                                pageEntryCategories + 1,
                            ))}
                    />
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
                        <summary class="admin-create-summary"><span>Új bejegyzés</span><AdminPlusIcon /></summary>
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

                        <label for="serv_url">Weblap URL</label>
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

                        <label class="flex items-center gap-xs font-normal">
                            <input
                                type="checkbox"
                                bind:checked={newEntry.verified}
                                class="w-auto"
                            />
                            Igényelt
                        </label>
                        <p class="admin-form-hint">Alapértelmezett: Nem ellenőrzött (szürke jelvény).</p>

                        <span class="form-group-label">Nyitvatartás</span>
                        <EntryHoursEditor bind:hours={newEntry.hours} />

                        <span class="form-group-label">Kiszállítási idő</span>
                        <EntryHoursEditor bind:hours={newEntry.delivery_hours} />

                        <span class="form-group-label">Fotók</span>
                        <EntryPhotosEditor
                            bind:photos={newEntry.photos}
                            alertError={(msg) => showAlert(msg)}
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
                                id="search_entries"
                                name="search_entries"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchEntries}
                                on:input={() => (pageEntries = 1)}
                                placeholder="Név, URL, címke…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgEntries.total}
                        page={pgEntries.page}
                        totalPages={pgEntries.totalPages}
                        from={pgEntries.from}
                        to={pgEntries.to}
                        on:prev={() =>
                            (pageEntries = Math.max(1, pageEntries - 1))}
                        on:next={() =>
                            (pageEntries = Math.min(
                                pgEntries.totalPages,
                                pageEntries + 1,
                            ))}
                    />
                    <div class="admin-table-wrapper">
                        <table class="admin-table">
                            <thead>
                                <tr>
                                    <th>Név</th>
                                    <th>Típus</th>
                                    <th>Igényelt</th>
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
                                {#each pgEntries.rows as s}
                                    <tr>
                                        <td>{s.name}</td>
                                        <td
                                            ><span class="badge"
                                                >{s.type || "entry"}</span
                                            ></td
                                        >
                                        <td>{s.verified ? "Ellenőrzött" : "Nem ellenőrzött"}</td>
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
                    <AdminPaginationBar
                        total={pgEntries.total}
                        page={pgEntries.page}
                        totalPages={pgEntries.totalPages}
                        from={pgEntries.from}
                        to={pgEntries.to}
                        on:prev={() =>
                            (pageEntries = Math.max(1, pageEntries - 1))}
                        on:next={() =>
                            (pageEntries = Math.min(
                                pgEntries.totalPages,
                                pageEntries + 1,
                            ))}
                    />
                {/if}

                <!-- Beállítások (Settings) Tab -->
                {#if activeTab === "settings"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {:else}
                        <p class="admin-info">
                            Oldalszintű beállítások: alapértelmezett település vendégeknek és azoknak, akik nem választottak saját települést (kezdőlap időjárás, eseményszűrés),
                            időjárás-szolgáltatók engedélyezése, ikon stílus, cache TTL és látogató-becslés.
                            A <strong>cache törlése</strong> új verziószámot ad — a látogatók frissebb időjárást kapnak.
                        </p>
                    {/if}
                    <section class="admin-form-section">
                        <h3>Alapértelmezett település (MyLocation)</h3>
                        <p class="admin-hint">Ez a vendégek, és a saját település nélküli felhasználók alaphelye a kezdőlapon és az index közelségi rendezésénél. A saját települést mindenki a felhasználói beállításokban állítja; a kereső és az index szűrője csak a találatokat szűri.</p>
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
                            <summary class="admin-create-summary"><span>Új fordítás</span><AdminPlusIcon /></summary>
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
                    <div class="admin-table-toolbar">
                        <label class="admin-search-label"
                            >Keresés
                            <input
                                id="search_weather_trans"
                                name="search_weather_trans"
                                type="search"
                                class="admin-search-input"
                                bind:value={searchWeatherTrans}
                                on:input={() => (pageWeatherTrans = 1)}
                                placeholder="Szöveg, nyelv…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgWeatherTrans.total}
                        page={pgWeatherTrans.page}
                        totalPages={pgWeatherTrans.totalPages}
                        from={pgWeatherTrans.from}
                        to={pgWeatherTrans.to}
                        on:prev={() =>
                            (pageWeatherTrans = Math.max(
                                1,
                                pageWeatherTrans - 1,
                            ))}
                        on:next={() =>
                            (pageWeatherTrans = Math.min(
                                pgWeatherTrans.totalPages,
                                pageWeatherTrans + 1,
                            ))}
                    />
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
                                {#each pgWeatherTrans.rows as wt}
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
                    <AdminPaginationBar
                        total={pgWeatherTrans.total}
                        page={pgWeatherTrans.page}
                        totalPages={pgWeatherTrans.totalPages}
                        from={pgWeatherTrans.from}
                        to={pgWeatherTrans.to}
                        on:prev={() =>
                            (pageWeatherTrans = Math.max(
                                1,
                                pageWeatherTrans - 1,
                            ))}
                        on:next={() =>
                            (pageWeatherTrans = Math.min(
                                pgWeatherTrans.totalPages,
                                pageWeatherTrans + 1,
                            ))}
                    />
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

                            <label for="page_greeting">Bevezető (a cím alatt, nyilvános oldalakon)</label>
                            <textarea id="page_greeting" name="greeting" bind:value={editingPage.greeting} rows="3" placeholder="Rövid bevezető szöveg…"></textarea>

                            <label for="page_content">Tartalom (HTML)</label>
                            <textarea id="page_content" name="content" bind:value={editingPage.content} rows="20" class="input-mono"></textarea>

                            <div class="flex gap-md mt-md">
                                <button type="submit" class="admin-submit-btn" disabled={pageSaving}>
                                    {pageSaving ? 'Mentés…' : 'Mentés'}
                                </button>
                                <button type="button" class="btn-update" on:click={cancelEditPage}>Mégse</button>
                            </div>
                        </form>
                    {:else}
                        {#if !adminTabError}
                            <p class="admin-info">
                                Statikus oldalak: <strong>cím</strong>, <strong>bevezető</strong> (a főcím alatt) és <strong>HTML tartalom</strong> (pl. irányelvek).
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
                                    on:input={() => (pageAdminPages = 1)}
                                    placeholder="Slug, cím…"
                                /></label
                            >
                        </div>
                        <AdminPaginationBar
                            total={pgAdminPages.total}
                            page={pgAdminPages.page}
                            totalPages={pgAdminPages.totalPages}
                            from={pgAdminPages.from}
                            to={pgAdminPages.to}
                            on:prev={() =>
                                (pageAdminPages = Math.max(1, pageAdminPages - 1))}
                            on:next={() =>
                                (pageAdminPages = Math.min(
                                    pgAdminPages.totalPages,
                                    pageAdminPages + 1,
                                ))}
                        />
                        <div class="admin-table-wrapper">
                            <table class="admin-table">
                                <thead>
                                    <tr>
                                        <th>Slug</th>
                                        <th>Cím</th>
                                        <th>Bevezető</th>
                                        <th>Utolsó módosítás</th>
                                        <th class="admin-table-col--action">Szerk.</th>
                                        <th class="admin-table-col--action">Törlés</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {#each pgAdminPages.rows as pg}
                                        <tr>
                                            <td><a href={pg.slug === 'home' ? '/' : '/' + pg.slug} target="_blank">{pg.slug === 'home' ? '/' : '/' + pg.slug}</a></td>
                                            <td>{pg.title}</td>
                                            <td class="admin-table-cell-preview" title={pg.greeting || ''}>{pg.greeting ? contentPreview(pg.greeting) : '—'}</td>
                                            <td>{pg.updated_at ? pg.updated_at.slice(0, 19) : ''}</td>
                                            <td class="admin-table-col--action"><button type="button" class="btn-update" on:click={() => startEditPage(pg)}>Szerk.</button></td>
                                            <td class="admin-table-col--action admin-table-col--action--muted">—</td>
                                        </tr>
                                    {:else}
                                        <tr
                                            ><td colspan="6"
                                                >{adminPages?.length
                                                    ? "Nincs találat a keresésre."
                                                    : "Nincsenek oldalak."}</td
                                            ></tr
                                        >
                                    {/each}
                                </tbody>
                            </table>
                        </div>
                        <AdminPaginationBar
                            total={pgAdminPages.total}
                            page={pgAdminPages.page}
                            totalPages={pgAdminPages.totalPages}
                            from={pgAdminPages.from}
                            to={pgAdminPages.to}
                            on:prev={() =>
                                (pageAdminPages = Math.max(1, pageAdminPages - 1))}
                            on:next={() =>
                                (pageAdminPages = Math.min(
                                    pgAdminPages.totalPages,
                                    pageAdminPages + 1,
                                ))}
                        />

                    {/if}
                {/if}

                <!-- GYIK Tab -->
                {#if activeTab === "page_faq"}
                    {#if adminTabError && activeTab === adminTabError.tab}
                        <div class="admin-alert admin-alert--error" role="alert">
                            {adminTabError.message}
                        </div>
                    {/if}
                    {#if editingPageFaq}
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
                                            class="input-mono"
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
                            <textarea id="pfaq_disc" bind:value={editingPageFaq.disclaimer_markdown} rows="8" class="input-mono"></textarea>

                            <div class="flex gap-md mt-md">
                                <button type="submit" class="admin-submit-btn" disabled={pageFaqSaving}>
                                    {pageFaqSaving ? 'Mentés…' : 'Mentés'}
                                </button>
                                <button type="button" class="btn-update" on:click={cancelEditPageFaq}>Mégse</button>
                            </div>
                        </form>
                    {:else}
                        <h3 class="admin-subtab-heading">GYIK és felelősségkizárások</h3>
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
                                    on:input={() => (pagePageFaqRows = 1)}
                                    placeholder="Kulcs, név, cím…"
                                /></label
                            >
                        </div>
                        <AdminPaginationBar
                            total={pgPageFaqRows.total}
                            page={pgPageFaqRows.page}
                            totalPages={pgPageFaqRows.totalPages}
                            from={pgPageFaqRows.from}
                            to={pgPageFaqRows.to}
                            on:prev={() =>
                                (pagePageFaqRows = Math.max(1, pagePageFaqRows - 1))}
                            on:next={() =>
                                (pagePageFaqRows = Math.min(
                                    pgPageFaqRows.totalPages,
                                    pagePageFaqRows + 1,
                                ))}
                        />
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
                                    {#each pgPageFaqRows.rows as row}
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
                        <AdminPaginationBar
                            total={pgPageFaqRows.total}
                            page={pgPageFaqRows.page}
                            totalPages={pgPageFaqRows.totalPages}
                            from={pgPageFaqRows.from}
                            to={pgPageFaqRows.to}
                            on:prev={() =>
                                (pagePageFaqRows = Math.max(1, pagePageFaqRows - 1))}
                            on:next={() =>
                                (pagePageFaqRows = Math.min(
                                    pgPageFaqRows.totalPages,
                                    pagePageFaqRows + 1,
                                ))}
                        />
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
                        <summary class="admin-create-summary"><span>Új típus</span><AdminPlusIcon /></summary>
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
                                on:input={() => (pageEntryTypes = 1)}
                                placeholder="Név, ID…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgEntryTypes.total}
                        page={pgEntryTypes.page}
                        totalPages={pgEntryTypes.totalPages}
                        from={pgEntryTypes.from}
                        to={pgEntryTypes.to}
                        on:prev={() =>
                            (pageEntryTypes = Math.max(1, pageEntryTypes - 1))}
                        on:next={() =>
                            (pageEntryTypes = Math.min(
                                pgEntryTypes.totalPages,
                                pageEntryTypes + 1,
                            ))}
                    />
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
                                {#each pgEntryTypes.rows as et}
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
                    <AdminPaginationBar
                        total={pgEntryTypes.total}
                        page={pgEntryTypes.page}
                        totalPages={pgEntryTypes.totalPages}
                        from={pgEntryTypes.from}
                        to={pgEntryTypes.to}
                        on:prev={() =>
                            (pageEntryTypes = Math.max(1, pageEntryTypes - 1))}
                        on:next={() =>
                            (pageEntryTypes = Math.min(
                                pgEntryTypes.totalPages,
                                pageEntryTypes + 1,
                            ))}
                    />
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
                        <summary class="admin-create-summary"><span>Új látnivaló</span><AdminPlusIcon /></summary>
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
                                on:input={() => (pageAttractions = 1)}
                                placeholder="Név, slug, megye…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgAttractions.total}
                        page={pgAttractions.page}
                        totalPages={pgAttractions.totalPages}
                        from={pgAttractions.from}
                        to={pgAttractions.to}
                        on:prev={() =>
                            (pageAttractions = Math.max(1, pageAttractions - 1))}
                        on:next={() =>
                            (pageAttractions = Math.min(
                                pgAttractions.totalPages,
                                pageAttractions + 1,
                            ))}
                    />
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
                                {#each pgAttractions.rows as att}
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
                    <AdminPaginationBar
                        total={pgAttractions.total}
                        page={pgAttractions.page}
                        totalPages={pgAttractions.totalPages}
                        from={pgAttractions.from}
                        to={pgAttractions.to}
                        on:prev={() =>
                            (pageAttractions = Math.max(1, pageAttractions - 1))}
                        on:next={() =>
                            (pageAttractions = Math.min(
                                pgAttractions.totalPages,
                                pageAttractions + 1,
                            ))}
                    />
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
                                on:input={() => (pageCounties = 1)}
                                placeholder="Név, slug, székhely…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgCounties.total}
                        page={pgCounties.page}
                        totalPages={pgCounties.totalPages}
                        from={pgCounties.from}
                        to={pgCounties.to}
                        on:prev={() =>
                            (pageCounties = Math.max(1, pageCounties - 1))}
                        on:next={() =>
                            (pageCounties = Math.min(
                                pgCounties.totalPages,
                                pageCounties + 1,
                            ))}
                    />
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
                                {#each pgCounties.rows as c (c.id)}
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
                    <AdminPaginationBar
                        total={pgCounties.total}
                        page={pgCounties.page}
                        totalPages={pgCounties.totalPages}
                        from={pgCounties.from}
                        to={pgCounties.to}
                        on:prev={() =>
                            (pageCounties = Math.max(1, pageCounties - 1))}
                        on:next={() =>
                            (pageCounties = Math.min(
                                pgCounties.totalPages,
                                pageCounties + 1,
                            ))}
                    />

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
                                on:input={() => (pageHistoricalSeats = 1)}
                                placeholder="Név, slug…"
                            /></label
                        >
                    </div>
                    <AdminPaginationBar
                        total={pgHistoricalSeats.total}
                        page={pgHistoricalSeats.page}
                        totalPages={pgHistoricalSeats.totalPages}
                        from={pgHistoricalSeats.from}
                        to={pgHistoricalSeats.to}
                        on:prev={() =>
                            (pageHistoricalSeats = Math.max(
                                1,
                                pageHistoricalSeats - 1,
                            ))}
                        on:next={() =>
                            (pageHistoricalSeats = Math.min(
                                pgHistoricalSeats.totalPages,
                                pageHistoricalSeats + 1,
                            ))}
                    />
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
                                {#each pgHistoricalSeats.rows as h (h.id)}
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
                    <AdminPaginationBar
                        total={pgHistoricalSeats.total}
                        page={pgHistoricalSeats.page}
                        totalPages={pgHistoricalSeats.totalPages}
                        from={pgHistoricalSeats.from}
                        to={pgHistoricalSeats.to}
                        on:prev={() =>
                            (pageHistoricalSeats = Math.max(
                                1,
                                pageHistoricalSeats - 1,
                            ))}
                        on:next={() =>
                            (pageHistoricalSeats = Math.min(
                                pgHistoricalSeats.totalPages,
                                pageHistoricalSeats + 1,
                            ))}
                    />
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
                    <div class="admin-date-field">
                        <input
                            id="emondas_day_edit"
                            name="display_date"
                            type="date"
                            bind:value={editingMondas.display_date}
                            required
                            class="w-full"
                        />
                        <button
                            type="button"
                            class="btn btn-sm"
                            on:click={() =>
                                (editingMondas.display_date = localISODate())}
                            >Mai nap</button
                        >
                    </div>

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

                    <label for="edit_url">Weblap URL</label>
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

                    <label class="flex items-center gap-xs font-normal">
                        <input
                            type="checkbox"
                            bind:checked={editingEntry.verified}
                            class="w-auto"
                        />
                        Igényelt
                    </label>
                    <p class="admin-form-hint">Alapértelmezett: Nem ellenőrzött (szürke jelvény).</p>

                    <span class="form-group-label">Nyitvatartás</span>
                    <EntryHoursEditor bind:hours={editingEntry.hours} />

                    <span class="form-group-label">Kiszállítási idő</span>
                    <EntryHoursEditor bind:hours={editingEntry.delivery_hours} />

                    <span class="form-group-label">Fotók</span>
                    <EntryPhotosEditor
                        bind:photos={editingEntry.photos}
                        entryId={editingEntry.id}
                        alertError={(msg) => showAlert(msg)}
                    />

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
                        {#each settlementLocationTypes as t}<option value={t.slug}
                                >{t.label_hu}</option
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

                    <div class="admin-event-image-block">
                        <label for="edit_ev_featured_upload">Kiemelt kép</label>
                        {#if editingEvent.featured_image}
                            <img
                                class="admin-event-image-preview"
                                src={absoluteMediaUrl(
                                    editingEvent.featured_image,
                                    getApiBase(),
                                )}
                                alt=""
                            />
                        {/if}
                        <input
                            id="edit_ev_featured_upload"
                            type="file"
                            accept="image/jpeg,image/png,image/webp,image/gif"
                            on:change={(e) => {
                                const f = e.target.files?.[0];
                                if (f) uploadEventFeaturedImage(f, true);
                                e.target.value = "";
                            }}
                        />
                        <label for="edit_ev_featured_url" class="admin-sublabel"
                            >Vagy kép URL (külső)</label
                        >
                        <input
                            id="edit_ev_featured_url"
                            type="url"
                            bind:value={editingEvent.featured_image}
                            placeholder="https://…"
                        />
                        {#if editingEvent.featured_image}
                            <button
                                type="button"
                                class="btn-update"
                                style="align-self: flex-start"
                                on:click={() =>
                                    (editingEvent.featured_image = "")}
                                >Kép törlése</button
                            >
                        {/if}
                    </div>

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

                    <label for="edit_ev_type_id"
                        >Eseménytípus <span class="admin-req" title="Kötelező">*</span></label
                    >
                    <select
                        id="edit_ev_type_id"
                        name="event_type_id"
                        bind:value={editingEvent.event_type_id}
                        on:change={() => (editingEvent.event_subtype_id = "")}
                        required
                    >
                        <option value="">— válassz —</option>
                        {#each [...catalogEventTypes].sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0) || String(a.label_hu).localeCompare(String(b.label_hu), "hu")) as t}
                            <option value={String(t.id)}
                                >{t.label_hu} ({t.slug})</option
                            >
                        {/each}
                    </select>

                    <label for="edit_ev_subtype_id">Altípus (opcionális)</label>
                    <select
                        id="edit_ev_subtype_id"
                        name="event_subtype_id"
                        bind:value={editingEvent.event_subtype_id}
                    >
                        <option value="">— nincs —</option>
                        {#each subtypesForEditEvent as s}
                            <option value={String(s.id)}
                                >{s.label_hu} ({s.slug})</option
                            >
                        {/each}
                    </select>

                    <label for="edit_ev_access">Hozzáférés</label>
                    <select
                        id="edit_ev_access"
                        name="access_type"
                        bind:value={editingEvent.access_type}
                    >
                        <option value="public">{ACCESS_TYPE_LABELS.public}</option>
                        <option value="members_only">{ACCESS_TYPE_LABELS.members_only}</option>
                        <option value="invitation_only">{ACCESS_TYPE_LABELS.invitation_only}</option>
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

                    <label for="edit_ev_entry_price">Belépő / jegyár (opcionális)</label>
                    <input
                        id="edit_ev_entry_price"
                        name="entry_price"
                        type="text"
                        bind:value={editingEvent.entry_price}
                        placeholder="pl. 99 RON, 15 EUR, ingyenes"
                        maxlength="128"
                        autocomplete="off"
                    />

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
    .admin-date-field {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        flex-wrap: wrap;
    }
    .admin-date-field input[type="date"] {
        min-width: 12rem;
    }
    .admin-date-today-badge {
        display: inline-block;
        margin-left: 0.35rem;
        padding: 0.1rem 0.45rem;
        border-radius: 999px;
        background: var(--szekely-red, #c8102e);
        color: #fff;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.03em;
    }
    tr.admin-row-today {
        background: color-mix(
            in srgb,
            var(--szekely-green, #2f4f4f) 10%,
            transparent
        );
    }
    .admin-region-heading {
        margin-top: 2.5rem;
        margin-bottom: 0.5rem;
    }
    .admin-subtab-heading {
        margin-top: 1.75rem;
        margin-bottom: 0.5rem;
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
    }
    .admin-region-edit-full textarea {
        width: 100%;
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
    /* .admin-faq-pair-fields label {
    } */

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
        text-align: left;
    }
    .org-suggestions li button:hover {
        background: var(--accent-bg, #2a2a3e);
    }
    .org-sug-meta {
        color: var(--text-faint, #888);
    }
</style>