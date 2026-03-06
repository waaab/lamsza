<script>
    import { onMount } from "svelte";

    const apiBase = import.meta.env.VITE_API_BASE_URL;
    if (!apiBase)
        console.warn(
            "VITE_API_BASE_URL is not set. Falling back to http://localhost:3000",
        );
    const baseUrl = apiBase || "http://localhost:3000";

    // --- Component State ---
    let mondasLoading = true;
    let mondasText = "";
    let mondasError = false;

    let newsLoadingTeaser = true;
    let newsFeedsTeaser = [];
    let newsFeedsError = false;

    let quickLinksLoading = true;
    let quickLinksData = [];
    let quickLinksError = false;

    let searchResultsHTML = null;
    let searchInputValue = "";

    // --- Search Logic ---
    function handleKeydown(e) {
        if (e.key === "Enter") {
            executeSearch("szekely");
        } else if (e.key === "Escape") {
            if (searchResultsHTML !== null || searchInputValue !== "") {
                clearSearch();
            }
        }
    }

    function clearSearch() {
        searchResultsHTML = null;
        searchInputValue = "";
        const searchInput = document.getElementById("searchInput");
        if (searchInput) {
            searchInput.classList.remove("search-input--active");
            searchInput.focus();
        }
    }

    function buildExternalLinks(query) {
        const q = encodeURIComponent(query);
        const externalIcon = `<svg class="ext-icon" xmlns="http://www.w3.org/2000/svg" width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path><polyline points="15 3 21 3 21 9"></polyline><line x1="10" y1="14" x2="21" y2="3"></line></svg>`;
        return `
        <div class="external-search-links">
            <span class="ext-label">Reakereshetsz máshol es:</span>
            <a href="https://www.google.com/search?q=${q}" target="_blank" rel="nofollow noopener">Google ${externalIcon}</a>
            <a href="https://www.bing.com/search?q=${q}" target="_blank" rel="nofollow noopener">Bing ${externalIcon}</a>
            <a href="https://duckduckgo.com/?q=${q}" target="_blank" rel="nofollow noopener">DuckDuckGo ${externalIcon}</a>
            <a href="https://yandex.com/search/?text=${q}" target="_blank" rel="nofollow noopener">Yandex ${externalIcon}</a>
        </div>`;
    }

    async function executeSearch(engine) {
        const query = searchInputValue.trim();
        if (!query) {
            alert(
                "Aszongyák, aki keres az talál! Te nem kerestel semmit ember!",
            );
            return;
        }

        if (engine === "szekely") {
            searchResultsHTML =
                '<div class="skeleton skeleton-text search-loading-skeleton">Keresés...</div>';

            try {
                const [resultsRes, suggestionsRes] = await Promise.all([
                    fetch(
                        `${baseUrl}/api/directory?q=` +
                            encodeURIComponent(query),
                    ),
                    fetch(
                        `${baseUrl}/api/autosuggest?q=` +
                            encodeURIComponent(query),
                    ),
                ]);

                const [data, suggestionsData] = await Promise.all([
                    resultsRes.json(),
                    suggestionsRes.json(),
                ]);

                let html = "";
                const qLower = query.toLowerCase();

                // 1. Suggestions / Related keywords
                let suggestionsHtml = "";
                const cachedNews = localStorage.getItem("news_cache");
                let newsKeywords = [];
                if (cachedNews) {
                    try {
                        const { items } = JSON.parse(cachedNews);
                        if (items) {
                            // Extract words from news titles that match query
                            const freq = {};
                            items.forEach((item) => {
                                if (item.title.toLowerCase().includes(qLower)) {
                                    item.title
                                        .toLowerCase()
                                        .split(/\s+/)
                                        .forEach((w) => {
                                            const clean = w.replace(
                                                /[^a-z\u00e1\u00e9\u00ed\u00f3\u00f6\u0151\u00fa\u00fc\u0171]/gi,
                                                "",
                                            );
                                            if (
                                                clean.length > 3 &&
                                                clean !== qLower &&
                                                clean.includes(qLower)
                                            ) {
                                                freq[clean] =
                                                    (freq[clean] || 0) + 1;
                                            }
                                        });
                                }
                            });
                            newsKeywords = Object.entries(freq)
                                .sort((a, b) => b[1] - a[1])
                                .slice(0, 3)
                                .map((e) => e[0]);
                        }
                    } catch (e) {}
                }

                if (
                    (suggestionsData && suggestionsData.length > 0) ||
                    newsKeywords.length > 0
                ) {
                    const filteredSuggestions = (suggestionsData || [])
                        .filter((s) => s.toLowerCase() !== qLower)
                        .slice(0, 5);

                    if (filteredSuggestions.length > 0) {
                        suggestionsHtml = filteredSuggestions
                            .map(
                                (s) => `
                                <button 
                                    class="suggestion-tag btn btn-sm"
                                    onclick="window.executeSearchForSuggestion('${s.replace(/'/g, "\\'")}')"
                                >#${s.replace(/^#/, "")}</button>
                            `,
                            )
                            .join("");
                    }
                }

                // --- Data preparation for both sections ---
                const directoryCount = data ? data.length : 0;
                let dirSuggestionsText = suggestionsHtml
                    ? `Kapcsolódó szavak: ${suggestionsHtml}`
                    : "";

                let newsCount = 0;
                let matchingNews = [];
                if (cachedNews) {
                    try {
                        const { items } = JSON.parse(cachedNews);
                        matchingNews = items
                            .filter(
                                (item) =>
                                    item.title.toLowerCase().includes(qLower) ||
                                    item.source.toLowerCase().includes(qLower),
                            )
                            .slice(0, 5);
                        newsCount = matchingNews.length;
                    } catch (e) {}
                }

                // --- 1. Combined Header Section ---
                html += `<div class="filter-actions">`;

                // Directory Info-Box
                html += `<div class="info-box">`;
                html += `<p>`;
                if (directoryCount === 0) {
                    html += `<span>Nincs találat az Indexben erre a keresésre.</span>`;
                    if (dirSuggestionsText)
                        html += `<span class="results-separator">|</span> ${dirSuggestionsText}`;
                } else {
                    html += `🔍 Keresés: <span class="active">${query}</span>`;
                    if (dirSuggestionsText)
                        html += `<span class="results-separator">|</span> ${dirSuggestionsText}`;
                }
                html += `</p>`;
                html += `<p><span>(${directoryCount} találat)</span></p>`;
                html += `</div>`; // End Directory Info-Box

                // News Info-Box (only if matches or suggestions exist)
                if (newsCount > 0 || newsKeywords.length > 0) {
                    let newsSuggestionsText =
                        newsKeywords.length > 0
                            ? `<span class="active news-badge-label">📰 Hírekben:</span>` +
                              newsKeywords
                                  .map(
                                      (s) =>
                                          `<button class="news-tag btn btn-sm" onclick="window.executeSearchForSuggestion('${s.replace(/'/g, "\\'")}')">#${s.replace(/^#/, "")}</button>`,
                                  )
                                  .join("")
                            : "";

                    html += `<div class="info-box brown">`;
                    html += `<p>`;
                    if (newsCount === 0) {
                        html += `<span>Nincs találat a hírekben erre a keresésre.</span>`;
                        if (newsSuggestionsText)
                            html += `<span class="results-separator">|</span> ${newsSuggestionsText}`;
                    } else {
                        html += `📰 Hírek: <span class="active">${query}</span>`;
                        if (newsSuggestionsText)
                            html += `<span class="results-separator">|</span> ${newsSuggestionsText}`;
                    }
                    html += `</p>`;
                    html += `<p><span>(${newsCount} találat)</span></p>`;
                    html += `</div>`; // End News Info-Box
                }

                html += `</div>`; // End Single Filter-Actions Parent

                // --- 2. Render Cards Section ---
                let hasCards = false;
                if (
                    directoryCount > 0 ||
                    newsCount > 0 ||
                    qLower.includes("idő") ||
                    qLower.includes("időjárás") ||
                    qLower.includes("milyen") ||
                    qLower.includes("hőmérséklet") ||
                    qLower.includes("hír") ||
                    qLower.includes("hirek") ||
                    qLower.includes("hírek")
                ) {
                    html += '<div class="list flex">';
                    hasCards = true;
                }

                // 2.1 Directory Cards (Green)
                if (directoryCount > 0) {
                    html += data
                        .map(
                            (service) => `
                        <div class="card service">
                            ${
                                service.is_direct_match
                                    ? `<div class="badge">Közvetlen Találat</div>`
                                    : `<div class="badge">Tartalom Találat</div>`
                            }
                            <div class="badge">Index: ${service.category}</div>
                            <h3 class="service-name">
                                ${service.entity_type === "settlement" ? `<a href="/${service.county_slug}-megye/${service.slug}" class="service-link">${service.name}</a>` : `<a href="/bejegyzes/${service.slug}" class="service-link">${service.name}</a>`}
                            </h3>
                            ${service.url ? `<div class="service-info service-url-box"><span class="service-url-icon">🔗</span><a href="${service.url}" target="_blank" rel="nofollow noopener" class="service-url-link">${service.url}</a></div>` : ""}
                            <div class="service-info">📍 ${service.location} - ${service.address}</div>
                            <div class="service-info"> ${service.phone}</div>
                            ${
                                service.tags && service.tags.length > 0
                                    ? `<div class="service-tags">${service.tags
                                          .map(
                                              (t) =>
                                                  `<span class="service-tag">${t.startsWith("#") ? t : "#" + t}</span>`,
                                          )
                                          .join("")}</div>`
                                    : ""
                            }
                            ${service.notes ? `<div class="service-notes">${service.notes}</div>` : ""}
                        </div>
                    `,
                        )
                        .join("");
                }

                // 2.2 Weather Cards (Blue)
                if (
                    qLower.includes("idő") ||
                    qLower.includes("időjárás") ||
                    qLower.includes("milyen") ||
                    qLower.includes("hőmérséklet")
                ) {
                    const cachedWeather = localStorage.getItem("weather_cache");
                    if (cachedWeather) {
                        try {
                            const { temp, desc } = JSON.parse(cachedWeather);
                            html += `
                                <div class="card weather">
                                    <div class="weather-badge">Időjárás</div>
                                    <h3 class="service-name">Jelenlegi idő</h3>
                                    <div class="weather-current-large">${temp}°C</div>
                                    <div class="weather-desc">${desc}</div>
                                </div>
                            `;
                        } catch (e) {}
                    }
                }

                // 2.3 News Cards (Orange)
                if (newsCount > 0) {
                    html += matchingNews
                        .map(
                            (item) => `
                        <div class="card news">
                            <div class="badge">Hírek: ${item.source}</div>
                            <h3 class="service-name">
                                <a href="${item.link}" target="_blank" rel="nofollow noopener" class="service-link">📰 ${item.title}</a>
                            </h3>
                            <div class="service-info">${new Date(item.pubDate).toLocaleDateString("hu-HU")}</div>
                        </div>
                    `,
                        )
                        .join("");
                } else if (
                    qLower.includes("hír") ||
                    qLower.includes("hirek") ||
                    qLower.includes("hírek")
                ) {
                    const cachedNews = localStorage.getItem("news_cache");
                    if (cachedNews) {
                        try {
                            const { items } = JSON.parse(cachedNews);
                            if (items && items.length > 0) {
                                html += `
                                    <div class="card news">
                                        <div class="badge">Hírek</div>
                                        <h3 class="service-name">Kapcsolódó hírek</h3>
                                        ${items
                                            .slice(0, 3)
                                            .map(
                                                (i) =>
                                                    `<div class="news-item-link-box"><a href="${i.link}" class="news-item-link" target="_blank" rel="nofollow noopener">📰 ${i.title}</a></div>`,
                                            )
                                            .join("")}
                                    </div>
                                `;
                            }
                        } catch (e) {}
                    }
                }

                if (hasCards) {
                    html += "</div>";
                }

                html += buildExternalLinks(query);
                searchResultsHTML = html;
                const sInput = document.getElementById("searchInput");
                if (sInput) sInput.classList.add("search-input--active");
            } catch (err) {
                console.error(err);
                searchResultsHTML =
                    '<div class="error-msg">Hiba történt a keresés során.</div>';
            }
        }
    }

    // Expose search function to window for @html clicks
    if (typeof window !== "undefined") {
        window.executeSearchForSuggestion = function (s) {
            searchInputValue = s;
            const sInput = document.getElementById("searchInput");
            if (sInput) {
                sInput.value = s;
                sInput.focus();
            }
            executeSearch("szekely");
        };

        window.clearSearchAndRefocus = function () {
            searchInputValue = "";
            searchResultsHTML = "";
            const sInput = document.getElementById("searchInput");
            if (sInput) {
                sInput.classList.remove("search-input--active");
                sInput.focus();
            }
        };
    }

    // --- Utilities ---
    $: teaserItems = (() => {
        const counts = {};
        return newsFeedsTeaser.filter((item) => {
            counts[item.source] = (counts[item.source] || 0) + 1;
            return counts[item.source] <= 2;
        });
    })();

    function relativeTime(ts) {
        const diff = Date.now() - ts;
        const minutes = Math.floor(diff / 60000);
        const hours = Math.floor(diff / 3600000);
        const days = Math.floor(diff / 86400000);
        if (minutes < 2) return "most";
        if (minutes < 60) return minutes + " perce";
        if (hours < 24) return hours + " órája";
        if (days === 1) return "1 napja";
        return days + " napja";
    }

    onMount(() => {
        // 1. Mondás Logic
        const MONDAS_CACHE_KEY = "mondas_cache";
        const MONDAS_TTL = 24 * 60 * 60 * 1000;
        const cachedMondas = localStorage.getItem(MONDAS_CACHE_KEY);
        let usedMondasCache = false;

        if (cachedMondas) {
            try {
                const { text, timestamp } = JSON.parse(cachedMondas);
                if (Date.now() - timestamp < MONDAS_TTL) {
                    mondasText = text;
                    mondasLoading = false;
                    usedMondasCache = true;
                }
            } catch (e) {}
        }

        if (!usedMondasCache) {
            fetch(`${baseUrl}/api/admin/mondasok`)
                .then((res) => res.json())
                .then((data) => {
                    mondasLoading = false;
                    if (data && data.length > 0) {
                        const randomIndex = Math.floor(
                            Math.random() * data.length,
                        );
                        const text = data[randomIndex].text;
                        mondasText = text;
                        localStorage.setItem(
                            MONDAS_CACHE_KEY,
                            JSON.stringify({ text, timestamp: Date.now() }),
                        );
                    }
                })
                .catch(() => {
                    mondasError = true;
                    mondasLoading = false;
                });
        }

        // 2. Weather Logic
        const API_KEY = import.meta.env.VITE_WEATHER_API_KEY;
        if (!API_KEY) console.warn("VITE_WEATHER_API_KEY is missing");
        const CITY = encodeURIComponent("Miercurea Ciuc");
        const WEATHER_CACHE_KEY_MOUNT = "weather_cache";
        const WEATHER_TTL = 30 * 60 * 1000;

        function iconEmoji(code) {
            const map = {
                "01d": "☀️",
                "01n": "🌙",
                "02d": "⛅",
                "02n": "☁️",
                "03d": "☁️",
                "03n": "☁️",
                "04d": "☁️",
                "04n": "☁️",
                "09d": "🌧️",
                "09n": "🌧️",
                "10d": "🌦️",
                "10n": "🌧️",
                "11d": "⛈️",
                "11n": "⛈️",
                "13d": "❄️",
                "13n": "❄️",
                "50d": "🌫️",
                "50n": "🌫️",
            };
            return map[code] || "🌡️";
        }

        function displayWeather(temp, tempMin, desc, icon, timestamp) {
            const weatherLoading = document.getElementById("weatherLoading");
            const weatherContent = document.getElementById("weatherContent");
            const tempEl = document.getElementById("temp");
            const descEl = document.getElementById("desc");

            if (weatherLoading) weatherLoading.style.display = "none";
            if (weatherContent) weatherContent.style.display = "flex";
            if (tempEl) tempEl.innerText = temp;
            if (descEl) descEl.innerText = desc;

            const tempMinEl = document.getElementById("weatherTempMin");
            if (tempMinEl)
                tempMinEl.innerText =
                    tempMin != null ? "/ " + tempMin + "°C" : "";

            const iconEl = document.getElementById("weatherIcon");
            if (iconEl) iconEl.innerText = iconEmoji(icon);

            const timestampEl = document.getElementById("weatherTimestamp");
            if (timestampEl && timestamp) {
                const timeStr = new Date(timestamp).toLocaleTimeString(
                    "hu-HU",
                    { hour: "2-digit", minute: "2-digit" },
                );
                timestampEl.innerText = "Utoljára frissítve: " + timeStr;
            }
        }

        const cachedWeather = localStorage.getItem(WEATHER_CACHE_KEY_MOUNT);
        let useWeatherCache = false;
        if (cachedWeather) {
            try {
                const { temp, tempMin, desc, icon, timestamp } =
                    JSON.parse(cachedWeather);
                if (Date.now() - timestamp < WEATHER_TTL) {
                    displayWeather(temp, tempMin, desc, icon, timestamp);
                    useWeatherCache = true;
                }
            } catch (e) {}
        }

        if (!useWeatherCache) {
            const full_weather_url = `https://api.openweathermap.org/data/2.5/weather?q=${CITY},RO&units=metric&appid=${API_KEY}&lang=hu`;
            fetch(
                `${baseUrl}/api/proxy?url=` +
                    encodeURIComponent(full_weather_url),
            )
                .then((res) => res.json())
                .then((data) => {
                    const temp = Math.round(data.main.temp);
                    const tempMin = Math.round(data.main.temp_min);
                    const desc = data.weather[0].description;
                    const icon = data.weather[0].icon;
                    const ts = Date.now();
                    displayWeather(temp, tempMin, desc, icon, ts);
                    localStorage.setItem(
                        WEATHER_CACHE_KEY_MOUNT,
                        JSON.stringify({
                            temp,
                            tempMin,
                            desc,
                            icon,
                            timestamp: ts,
                        }),
                    );
                })
                .catch((err) => console.error("Weather hiba:", err));
        }

        // 3. News RSS Logic
        const NEWS_CACHE_KEY_RSS = "news_cache";
        const NEWS_TTL_RSS = 30 * 60 * 1000;

        async function fetchNews() {
            try {
                const cachedNews = localStorage.getItem(NEWS_CACHE_KEY_RSS);
                if (cachedNews) {
                    const { items, timestamp } = JSON.parse(cachedNews);
                    if (Date.now() - timestamp < NEWS_TTL_RSS) {
                        newsFeedsTeaser = items;
                        newsLoadingTeaser = false;
                        return;
                    }
                }

                const feedsRes = await fetch(`${baseUrl}/api/admin/news_feeds`);
                if (!feedsRes.ok) throw new Error("News feeds API error");
                const dbFeeds = await feedsRes.json();

                if (!dbFeeds || dbFeeds.length === 0) {
                    newsLoadingTeaser = false;
                    newsFeedsError = true;
                    return;
                }

                const allItems = [];
                for (const feedObj of dbFeeds) {
                    try {
                        const proxiedUrl =
                            `${baseUrl}/api/proxy?url=` +
                            encodeURIComponent(feedObj.feed_url);
                        const response = await fetch(proxiedUrl);
                        const xmlText = await response.text();
                        const parser = new DOMParser();
                        const xmlDoc = parser.parseFromString(
                            xmlText,
                            "text/xml",
                        );
                        const nodes = Array.from(
                            xmlDoc.querySelectorAll("item"),
                        ).slice(0, 10);

                        nodes.forEach((node) => {
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
                                bgColor: feedObj.bg_color || "#ffebd6",
                            });
                        });
                    } catch (e) {
                        console.error("RSS feed error:", feedObj.feed_url, e);
                    }
                }

                if (allItems.length > 0) {
                    allItems.sort((a, b) => b.pubDate - a.pubDate);
                    newsFeedsTeaser = allItems;
                    newsLoadingTeaser = false;
                    localStorage.setItem(
                        NEWS_CACHE_KEY_RSS,
                        JSON.stringify({
                            items: allItems,
                            timestamp: Date.now(),
                        }),
                    );
                } else {
                    newsLoadingTeaser = false;
                    newsFeedsError = true;
                }
            } catch (err) {
                console.error("News fetch error:", err);
                newsLoadingTeaser = false;
                newsFeedsError = true;
            }
        }
        fetchNews();

        // 4. Quick Links Logic
        const QUICK_LINKS_CACHE_KEY_MOUNT = "quick_links_cache";
        const QUICK_LINKS_TTL_MOUNT = 60 * 60 * 1000;

        const cachedQuickLinks = localStorage.getItem(
            QUICK_LINKS_CACHE_KEY_MOUNT,
        );
        let usedQuickLinksCache = false;
        if (cachedQuickLinks) {
            try {
                const { items, timestamp } = JSON.parse(cachedQuickLinks);
                if (Date.now() - timestamp < QUICK_LINKS_TTL_MOUNT) {
                    quickLinksData = items;
                    quickLinksLoading = false;
                    usedQuickLinksCache = true;
                }
            } catch (e) {}
        }

        if (!usedQuickLinksCache) {
            fetch(`${baseUrl}/api/admin/quick_links`)
                .then((res) => res.json())
                .then((data) => {
                    quickLinksLoading = false;
                    if (data && data.length > 0) {
                        quickLinksData = data;
                        localStorage.setItem(
                            QUICK_LINKS_CACHE_KEY_MOUNT,
                            JSON.stringify({
                                items: data,
                                timestamp: Date.now(),
                            }),
                        );
                    }
                })
                .catch(() => {
                    quickLinksError = true;
                    quickLinksLoading = false;
                });
        }

        // 5. Search button wire-up
        const searchBtn = document.getElementById("searchBtn");
        if (searchBtn) {
            searchBtn.addEventListener("click", () => executeSearch("szekely"));
        }
    });
</script>

<div class="home-main">
    <h1 class="page-title">Székely Gugel</h1>
    <p class="greeting">Az internet székely kapuja.</p>

    <section class="search-container">
        <input
            type="text"
            id="searchInput"
            class="search-input"
            placeholder="Mit keresel máma...?"
            bind:value={searchInputValue}
            on:keydown={handleKeydown}
            autocomplete="off"
        />
        <div class="search-buttons">
            {#if searchInputValue !== ""}
                <button
                    class="clear-search-btn"
                    on:click={clearSearch}
                    aria-label="Keresés törlése"
                    title="Keresés törlése"
                >
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        width="20"
                        height="20"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        ><line x1="18" y1="6" x2="6" y2="18"></line><line
                            x1="6"
                            y1="6"
                            x2="18"
                            y2="18"
                        ></line></svg
                    >
                </button>
            {/if}
            <button
                id="searchBtn"
                class="btn btn-primary"
                on:click={() => executeSearch("szekely")}>Na lámsza!</button
            >
        </div>
    </section>

    {#if searchResultsHTML !== null}
        <section class="results-container">
            {@html searchResultsHTML}
        </section>
    {/if}
</div>

{#if mondasLoading || mondasText !== "" || mondasError}
    <section id="szekely-mondasok">
        <div class="mondas-inner">
            <div class="mondas-label-row">
                <span class="heading-label">Napi Székely Mondás</span>
            </div>
            {#if mondasLoading}
                <div class="skeleton mondas-skeleton"></div>
            {:else if mondasError}
                <p class="mondas-quote">A mondás jelenleg nem elérhető.</p>
            {:else}
                <blockquote class="mondas-quote">{mondasText}</blockquote>
            {/if}
        </div>
    </section>
{/if}

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

<section id="idojaras">
    <div id="weatherLoading" class="weather-card">
        <div class="weather-left">
            <span class="widget-title">Időjárás</span>
            <div class="weather-temp-row">
                <div class="skeleton weather-skeleton-temp"></div>
            </div>
            <div class="weather-desc">
                <div class="skeleton weather-skeleton-desc"></div>
            </div>
            <div class="weather-footer">
                <div class="skeleton skeleton-footer-1"></div>
                <div class="skeleton skeleton-footer-2"></div>
            </div>
        </div>
        <div class="weather-right">
            <span class="weather-icon">⛅</span>
        </div>
    </div>
    <div id="weatherContent" class="weather-card weather-hidden">
        <div class="weather-left">
            <span class="widget-title">Csíkszereda</span>
            <div class="weather-temp-row">
                <span class="weather-temp" id="temp"></span><span
                    class="weather-temp-unit">°C</span
                >
                <span class="weather-temp-min" id="weatherTempMin"></span>
            </div>
            <div class="weather-desc"><span id="desc"></span></div>
            <div class="weather-footer">
                <small class="weather-timestamp" id="weatherTimestamp"
                    >1 orája</small
                >
                <small class="weather-source">Forrás: OpenWeatherMap</small>
            </div>
        </div>
        <div class="weather-right">
            <span class="weather-icon" id="weatherIcon">⛅</span>
        </div>
    </div>
</section>

<section class="news-teaser">
    <div class="news-teaser-header">
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            class="news-teaser-icon"
            ><path
                d="M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a4 4 0 0 1-4-4V6"
            ></path><line x1="2" y1="6" x2="6" y2="6"></line><line
                x1="2"
                y1="10"
                x2="6"
                y2="10"
            ></line><line x1="10" y1="6" x2="18" y2="6"></line><line
                x1="10"
                y1="10"
                x2="14"
                y2="10"
            ></line></svg
        >
        <span class="widget-title">Friss hírek erdélyi forrásból</span>
    </div>

    {#if newsLoadingTeaser}
        {#each Array(4) as _}
            <div class="news-item">
                <div class="news-item-meta">
                    <div
                        class="skeleton skeleton-text news-item-meta-skeleton"
                    ></div>
                </div>
                <div class="skeleton skeleton-text"></div>
            </div>
        {/each}
    {:else if newsFeedsError}
        <div class="note error">
            Hírek jelenleg nincsenek konfigurálva vagy nem elérhetők.
        </div>
    {:else}
        {#each teaserItems as item}
            <div class="news-item">
                <div class="news-item-meta">
                    <span class="badge" style:background={item.bgColor}
                        >{item.source}</span
                    >
                    <time
                        class="news-date"
                        title={new Date(item.pubDate).toLocaleString("hu-HU")}
                        >{relativeTime(item.pubDate)}</time
                    >
                </div>
                <a
                    href={item.link}
                    target="_blank"
                    rel="nofollow noopener"
                    class="news-link">{item.title}</a
                >
            </div>
        {/each}
    {/if}
    <div class="load-more">
        <a href="/hirek" class="nav-btn">Erdélyi hírek →</a>
    </div>
</section>

<style>
    :global(.search-loading-skeleton) {
        width: 100%;
        height: 40px;
    }

    /* Dynamic Search Results Styles (Global as they are injected via @html) */
    :global(.results-separator) {
        color: var(--text-faint);
        margin: 0 0.5rem;
    }

    :global(.news-badge-label) {
        margin-right: 0.5rem;
    }

    :global(.service-link) {
        color: inherit;
        text-decoration: none;
    }

    :global(.service-url-box) {
        margin-bottom: 0.5rem;
    }

    :global(.service-url-icon) {
        color: var(--text-faint);
        margin-right: 0.3rem;
    }

    :global(.service-url-link) {
        color: var(--primary-color);
        text-decoration: none;
    }

    :global(.weather-current-large) {
        font-size: 2rem;
    }

    :global(.news-item-link-box) {
        margin-bottom: 0.8rem;
    }

    :global(.news-item-link) {
        text-decoration: none;
        font-weight: 600;
        font-size: 1.1rem;
    }

    /* Static Parts Styles */
    .quick-links-error {
        margin: 0;
        padding: 0.5rem 0;
        font-style: italic;
    }

    .skeleton-footer-1 {
        width: 120px;
        height: 0.75rem;
    }

    .skeleton-footer-2 {
        width: 90px;
        height: 0.75rem;
    }

    .weather-hidden {
        display: none;
    }
</style>
