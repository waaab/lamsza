<script>
    import { onMount } from "svelte";

    let entries = [];
    let loading = true;
    let error = null;

    // Filter to retain explicit Entry categories
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

            const allEntries = (await res.json()) || [];

            // Filter down exclusively to the legacy entries
            entries = allEntries.filter(
                (s) =>
                    s.category &&
                    retainedCategories.includes(s.category.toLowerCase()),
            );
        } catch (err) {
            console.error(err);
            error = "Nem sikerült betölteni a bejegyzéseket.";
        } finally {
            loading = false;
        }
    });
</script>

<svelte:head>
    <title>Bejegyzések - Index</title>
</svelte:head>

<div class="breadcrumbs">
    <a href="/">Főoldal</a> &rsaquo;
    <a href="/index">Index</a> &rsaquo;
    <span class="active">Kiemelt Bejegyzések</span>
</div>

<h1 class="page-title">Bejegyzések</h1>
<p class="greeting">
    Kiemelt helyi bejegyzések, mesteremberek és egészségügyi intézmények indexe.
</p>

{#if loading}
    <div class="list grid">
        {#each Array(6) as _}
            <article class="card entry--skeleton">
                <div class="skeleton skeleton-text skeleton-cat"></div>
                <div class="skeleton skeleton-text skeleton-title"></div>
                <div class="skeleton skeleton-text skeleton-loc"></div>
            </article>
        {/each}
    </div>
{:else if error}
    <div class="note error">{error}</div>
{:else if entries.length === 0}
    <div class="note error">
        Jelenleg nincs listázott bejegyzés ebben a kategóriában.
    </div>
{:else}
    <div class="list grid">
        {#each entries as entry}
            <article class="card entry entry-card">
                <span class="badge entry-badge">{entry.category}</span>
                <h3 class="entry-name">
                    <a href="/bejegyzes/{entry.slug}" class="entry-link"
                        >{entry.name}</a
                    >
                </h3>
                {#if entry.url}
                    <div class="entry-info entry-url-wrap">
                        <span class="entry-url-icon">🔗</span>
                        <a
                            href={entry.url}
                            target="_blank"
                            rel="nofollow noopener"
                            class="entry-url-link">{entry.url}</a
                        >
                    </div>
                {/if}
                <div class="entry-info">
                    📍 {[entry.location, entry.location_ro, entry.location_de]
                        .filter(Boolean)
                        .join(" | ")} - {entry.address}
                </div>
                {#if entry.phone}
                    <div class="entry-info entry-phone">
                        📞 {entry.phone}
                    </div>
                {/if}
            </article>
        {/each}
    </div>
{/if}

<section class="faq" id="gyik">
    <h2 class="faq-title">Gyakori kérdések</h2>
    <div class="faq-list">
        <details class="faq-item" open>
            <summary>Honnan származnak az adatok?</summary>
            <p>
                Az adatok a helyi szakemberektől és intézményektől származnak,
                akiket a rendszerünk folyamatosan indexel, hogy a legfrissebb
                elérhetőségeket biztosítsa.
            </p>
        </details>
        <details class="faq-item" open>
            <summary>Hogyan kerülhet be valaki a címtárba?</summary>
            <p>
                A beküldési folyamat hamarosan elérhető lesz az oldalon. Addig
                is, ha ismersz olyan szolgáltatót, aki még nem szerepel nálunk,
                keress minket bizalommal.
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
                A keresőnk kulcsszavak, kategóriák és települések alapján szűri
                a találatokat. A kereső prioritást ad a közvetlen név- és
                kategória-egyezéseknek, de a leírásokban is keres.
            </p>
        </details>
    </div>
</section>

<section class="note info" id="disclaimer">
    A lamsza.com indexe egy ingyenes információs szolgáltatás. Az adatok
    pontosságáért és a szolgáltatások minőségéért a lamsza.com nem vállal
    felelősséget. Kérjük, minden esetben ellenőrizze az adatokat a
    szolgáltatóval való kapcsolatfelvétel előtt.
</section>

<style>
    .entry--skeleton {
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
    .entry-card {
        cursor: pointer;
        transition:
            transform 0.2s,
            box-shadow 0.2s;
    }
    .entry-badge {
        margin-bottom: 0.5rem;
        display: inline-block;
    }
    .entry-link {
        color: inherit;
        text-decoration: none;
    }
    .entry-url-wrap {
        margin-bottom: 0.5rem;
    }
    .entry-url-icon {
        color: var(--text-faint);
        margin-right: 0.3rem;
    }
    .entry-url-link {
        color: var(--primary-color);
        text-decoration: none;
    }
    .entry-phone {
        color: var(--text-faint);
        margin-top: 0.25rem;
    }
</style>
