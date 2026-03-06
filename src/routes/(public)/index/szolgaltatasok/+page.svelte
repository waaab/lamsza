<script>
    import { onMount } from "svelte";

    let services = [];
    let loading = true;
    let error = null;

    // Filter to retain explicit Service categories
    const retainedCategories = [
        "mesteremberek",
        "egészségügy",
        "oktatás",
        "hivatalok",
    ];

    onMount(async () => {
        try {
            const apiBase =
                import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";
            const res = await fetch(`${apiBase}/api/directory`);
            if (!res.ok) throw new Error("Hiba a címtár betöltésekor");

            const allServices = (await res.json()) || [];

            // Filter down exclusively to the legacy services
            services = allServices.filter(
                (s) =>
                    s.category &&
                    retainedCategories.includes(s.category.toLowerCase()),
            );
        } catch (err) {
            console.error(err);
            error = "Nem sikerült betölteni a szolgáltatásokat.";
        } finally {
            loading = false;
        }
    });
</script>

<svelte:head>
    <title>Szolgáltatások - Index</title>
</svelte:head>

<div class="container main-container">
    <div class="breadcrumbs">
        <a href="/">Főoldal</a> &rsaquo;
        <a href="/index">Index</a> &rsaquo;
        <span class="active">Kiemelt Szolgáltatások</span>
    </div>

    <h1 class="page-title">Szolgáltatások</h1>
    <p class="greeting">
        Kiemelt helyi szolgáltatások, mesteremberek és egészségügyi intézmények
        indexe.
    </p>

    {#if loading}
        <div class="list grid">
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
            Jelenleg nincs listázott szolgáltatás ebben a kategóriában.
        </div>
    {:else}
        <div class="list grid">
            {#each services as service}
                <article class="card service service-card">
                    <span class="badge service-badge">{service.category}</span>
                    <h3 class="service-name">
                        <a href="/bejegyzes/{service.slug}" class="service-link"
                            >{service.name}</a
                        >
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
                        📍 {[
                            service.location,
                            service.location_ro,
                            service.location_de,
                        ]
                            .filter(Boolean)
                            .join(" | ")} - {service.address}
                    </div>
                    {#if service.phone}
                        <div class="service-info service-phone">
                            📞 {service.phone}
                        </div>
                    {/if}
                </article>
            {/each}
        </div>
    {/if}

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
    .service-card {
        cursor: pointer;
        transition:
            transform 0.2s,
            box-shadow 0.2s;
    }
    .service-badge {
        margin-bottom: 0.5rem;
        display: inline-block;
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
    .service-phone {
        color: var(--text-faint);
        margin-top: 0.25rem;
    }
</style>
