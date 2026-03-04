<script>
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";

    let locType = ""; // Dynamically populated from locations API
    let town = "";
    let displayTown = "";
    let displayTownRo = "";
    let displayTownDe = "";
    let countyName = "";

    let services = [];
    let loadingServices = true;
    let servicesError = null;

    let weatherData = null;
    let weatherLoading = true;

    let newsItems = [];
    let newsLoading = true;

    $: if (browser && $page.params.slug) {
        town = $page.params.slug.toLowerCase();
        displayTown = town.charAt(0).toUpperCase() + town.slice(1);
        displayTownRo = "";
        displayTownDe = "";
        fetchData();
    }

    async function fetchData() {
        loadingServices = true;
        weatherLoading = true;
        newsLoading = true;

        const apiBase =
            import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

        // 1. Fetch Directory Services
        try {
            const locRes = await fetch(`${apiBase}/api/locations`);
            if (locRes.ok) {
                const locations = await locRes.json();
                const locData = locations.find((l) => l.slug === town);
                if (locData) {
                    displayTown = locData.name;
                    displayTownRo = locData.name_ro;
                    displayTownDe = locData.name_de;
                    locType = locData.type || "";
                    countyName = locData.county || "";
                }
            }
            const res = await fetch(
                `${apiBase}/api/directory?location_slug=${encodeURIComponent(town)}`,
            );
            if (!res.ok) throw new Error("Hiba a címtár betöltésekor");
            services = (await res.json()) || [];
        } catch (err) {
            console.error(err);
            servicesError = "Nem sikerült betölteni az adatokat.";
        } finally {
            loadingServices = false;
        }

        // 2. Fetch Weather
        const API_KEY = import.meta.env.VITE_WEATHER_API_KEY;
        if (API_KEY) {
            try {
                const wRes = await fetch(
                    `${apiBase}/api/weather?slug=${encodeURIComponent(town)}&appid=${API_KEY}`,
                );
                if (wRes.ok) {
                    const wData = await wRes.json();
                    weatherData = {
                        temp: Math.round(wData.main.temp),
                        tempMin: Math.round(wData.main.temp_min),
                        desc: wData.weather[0].description,
                        icon: wData.weather[0].icon,
                    };
                }
            } catch (err) {
                console.error("Weather fetch error:", err);
            } finally {
                weatherLoading = false;
            }
        } else {
            weatherLoading = false;
        }

        // 3. Fetch News filtering by town
        try {
            const nRes = await fetch(`${apiBase}/api/admin/news_feeds`);
            const dbFeeds = await nRes.json();
            let allNews = [];

            if (dbFeeds && dbFeeds.length > 0) {
                for (const feed of dbFeeds) {
                    if (!feed.feed_url) continue;
                    try {
                        const proxiedUrl =
                            `${apiBase}/api/proxy?url=` +
                            encodeURIComponent(feed.feed_url);
                        const response = await fetch(proxiedUrl);
                        const xmlText = await response.text();
                        const parser = new DOMParser();
                        const xmlDoc = parser.parseFromString(
                            xmlText,
                            "text/xml",
                        );
                        const nodes = Array.from(
                            xmlDoc.querySelectorAll("item"),
                        );

                        nodes.forEach((node) => {
                            const title =
                                node.querySelector("title")?.textContent || "";
                            const desc =
                                node.querySelector("description")
                                    ?.textContent || "";

                            // Check if town is mentioned
                            if (
                                title.toLowerCase().includes(town) ||
                                desc.toLowerCase().includes(town)
                            ) {
                                allNews.push({
                                    title: title,
                                    link:
                                        node.querySelector("link")
                                            ?.textContent || "#",
                                    pubDate:
                                        new Date(
                                            node.querySelector(
                                                "pubDate",
                                            )?.textContent,
                                        ).getTime() || 0,
                                    source: feed.title || "Hír",
                                });
                            }
                        });
                    } catch (e) {
                        console.error("RSS feed error:", feed.feed_url, e);
                    }
                }
            }
            allNews.sort((a, b) => b.pubDate - a.pubDate);
            newsItems = allNews.slice(0, 6); // Top 6 news
        } catch (err) {
            console.error("News fetch error:", err);
        } finally {
            newsLoading = false;
        }
    }

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
</script>

<svelte:head>
    <title>{displayTown} - Index</title>
</svelte:head>

<div class="container" style="min-height: calc(100vh - 120px)">
    <Breadcrumbs
        label={displayTown}
        settlementType={locType}
        {countyName}
        countySlug={$page.params.countySlug}
    />

    <h1 class="page-title">{displayTown} {locType} és Környéke</h1>
    {#if displayTownRo || displayTownDe}
        <h2
            style="text-align: center; color: var(--text-faint); margin-top: -1rem; margin-bottom: 2rem; font-size: 1.2rem; font-weight: 400;"
        >
            {[displayTown, displayTownRo, displayTownDe]
                .filter(Boolean)
                .join(" | ")}
        </h2>
    {/if}
    <p class="greeting" style="margin-top: 0;">
        Helyi hírek, időjárás és címtár {displayTown} területén.
    </p>

    <!-- Widgets Row (Weather + News Preview) -->
    <div
        class="widgets-row"
        style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 1rem; margin-bottom: 2rem;"
    >
        <!-- Weather Widget -->
        <article
            class="card weather"
            style="display:flex; flex-direction:column; justify-content:center; align-items:center; text-align:center; padding:2rem;"
        >
            {#if weatherLoading}
                <div
                    class="skeleton skeleton-text"
                    style="width: 50%; height: 2rem; margin-bottom:1rem;"
                ></div>
                <div
                    class="skeleton skeleton-text"
                    style="width: 30%; height: 1rem;"
                ></div>
            {:else if weatherData}
                <div
                    style="font-size: 3rem; margin-bottom: 0.5rem;"
                    title={weatherData.desc}
                >
                    {iconEmoji(weatherData.icon)}
                </div>
                <h3 style="margin: 0; font-size: 2rem;">
                    {weatherData.temp}°C
                </h3>
                <p
                    style="margin: 0.5rem 0 0; color: var(--text-faint); text-transform: capitalize;"
                >
                    {weatherData.desc}
                </p>
                {#if weatherData.tempMin}
                    <p
                        style="margin: 0; color: var(--text-faint); font-size: 0.9em;"
                    >
                        Min: {weatherData.tempMin}°C
                    </p>
                {/if}
            {:else}
                <p style="color: var(--text-faint);">
                    Időjárás adat nem elérhető.
                </p>
            {/if}
        </article>

        <!-- Local News Widget -->
        <article class="card news" style="display:flex; flex-direction:column;">
            <div class="badge" style="align-self: flex-start;">Helyi Hírek</div>
            {#if newsLoading}
                <div
                    style="padding: 1rem 0; display:flex; flex-direction:column; gap:0.5rem;"
                >
                    <div
                        class="skeleton skeleton-text"
                        style="height: 1.2rem;"
                    ></div>
                    <div
                        class="skeleton skeleton-text"
                        style="height: 1.2rem; width: 80%;"
                    ></div>
                </div>
            {:else if newsItems.length > 0}
                <ul
                    style="list-style: none; padding: 0; margin: 1rem 0 0; display: flex; flex-direction: column; gap: 0.8rem;"
                >
                    {#each newsItems.slice(0, 4) as item}
                        <li>
                            <a
                                href={item.link}
                                target="_blank"
                                rel="nofollow noopener"
                                style="text-decoration: none; color: inherit; font-weight: 500; font-size: 1.05rem;"
                            >
                                📰 {item.title}
                            </a>
                            <div
                                style="font-size: 0.8em; color: var(--text-faint); margin-top: 0.2rem;"
                            >
                                {item.source} - {new Date(
                                    item.pubDate,
                                ).toLocaleDateString("hu-HU")}
                            </div>
                        </li>
                    {/each}
                </ul>
            {:else}
                <p style="margin-top: 1rem; color: var(--text-faint);">
                    Nincsenek friss helyi hírek.
                </p>
            {/if}
        </article>
    </div>

    <!-- Directory Section -->
    <h2
        style="margin-bottom: 1rem; border-bottom: 2px solid var(--border-color); padding-bottom: 0.5rem;"
    >
        Index
    </h2>

    {#if loadingServices}
        <div class="list grid">
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
    {:else if servicesError}
        <div class="error-msg">{servicesError}</div>
    {:else if services.length === 0}
        <div class="error-msg">
            Nincs megjeleníthető bejegyzés {displayTown} területén.
        </div>
    {:else}
        <div class="list grid">
            {#each services as service}
                <article class="card service">
                    <span class="badge">{service.category}</span>
                    <h3 class="service-name">
                        <a
                            href="/bejegyzes/{service.slug}"
                            style="color:inherit;text-decoration:none;"
                            >{service.name}</a
                        >
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
    {/if}
</div>
