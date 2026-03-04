<script>
    import { page } from "$app/stores";
    import { browser } from "$app/environment";
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";

    const locType = "megye"; // Hardcoded type for this route
    let town = "";
    let displayTown = "";
    let displayTownRo = "";
    let displayTownDe = "";

    let childSettlements = [];
    let services = [];
    let loadingServices = true;
    let servicesError = null;

    let weatherData = null;
    let weatherLoading = true;

    let newsItems = [];
    let newsLoading = true;

    $: if (browser && $page.params.countySlug) {
        town = $page.params.countySlug.toLowerCase();
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
            const locRes = await fetch(
                `${apiBase}/api/locations?type=${encodeURIComponent(locType)}`,
            );
            if (locRes.ok) {
                const locations = await locRes.json();
                const locData = locations.find((l) => l.slug === town);
                if (locData) {
                    displayTown = locData.name;
                    displayTownRo = locData.name_ro;
                    displayTownDe = locData.name_de;
                }
                childSettlements = locations
                    .filter(
                        (l) =>
                            l.county_slug === town &&
                            l.type.toLowerCase() !== "megye",
                    )
                    .sort((a, b) => a.name.localeCompare(b.name));
            }
            const res = await fetch(
                `${apiBase}/api/directory?slug=${encodeURIComponent(town)}&type=megye`,
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

        // 3. Fetch News filtering by town/county (Backend Aggregated)
        try {
            const nRes = await fetch(
                `${apiBase}/api/county_news?slug=${encodeURIComponent(town)}`,
            );
            if (nRes.ok) {
                const aggregatedNews = await nRes.json();

                // Ensure date parsing works for Svelte template
                newsItems = aggregatedNews
                    .map((item) => {
                        return {
                            ...item,
                            pubDate: new Date(item.pubDate).getTime() || 0,
                        };
                    })
                    .sort((a, b) => b.pubDate - a.pubDate)
                    .slice(0, 6);
            } else {
                newsItems = [];
            }
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
    <Breadcrumbs label={displayTown} type="Megye" />

    <h1 class="page-title">{displayTown} Megye</h1>
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

    <!-- County Settlements Aside -->
    {#if childSettlements.length > 0}
        <aside
            style="margin-bottom: 2rem; padding: 1.5rem; background: var(--card-bg); border-radius: 12px; border: 1px solid var(--border-color);"
        >
            <h2
                style="margin-top: 0; margin-bottom: 1.5rem; color: var(--text-color); font-size: 1.5rem; border-bottom: 2px solid var(--border-color); padding-bottom: 0.5rem;"
            >
                Települések {displayTown} megyében
            </h2>
            <div style="display: flex; flex-wrap: wrap; gap: 0.8rem;">
                {#each childSettlements as child (child.id)}
                    <a
                        href="/{$page.params.countySlug}-megye/{child.slug}"
                        class="badge"
                        style="text-decoration: none; color: var(--primary-color); background: var(--bg-body); font-weight: 500; padding: 0.5rem 1rem; border: 1px solid var(--border-color);"
                    >
                        {child.name}
                    </a>
                {/each}
            </div>
        </aside>
    {/if}

    <!-- Directory Section -->
    <h2
        style="margin-bottom: 1rem; border-bottom: 2px solid var(--border-color); padding-bottom: 0.5rem;"
    >
        Helyi Index
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
            Nincs megjeleníthető bejegyzés {displayTown} megye területén.
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
                        📍 {[
                            service.location,
                            service.location_ro,
                            service.location_de,
                        ]
                            .filter(Boolean)
                            .join(" | ")} - {service.address}
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
