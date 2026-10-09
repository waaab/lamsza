<script>
    import { onMount } from "svelte";
    import AppIcon from "$lib/icons/AppIcon.svelte";
    import { apiFetch } from "$lib/api.js";
    import {
        DIRECTORY_CATALOG,
        directoryCatalogFromApi,
        directoryGroupsForEntries,
    } from "$lib/entryCategory.js";

    const CATEGORY_ICONS = {
        etkezes: "category-etkezes",
        szallas: "category-szallas",
        "egeszseg-es-szepseg": "category-egeszseg",
        vasarlas: "category-vasarlas",
        auto: "category-auto",
        mesteremberek: "category-mester",
        oktatas: "category-oktatas",
        hivatalok: "category-hivatal",
        "sport-es-szabadido": "category-sport",
        penzugy: "category-penzugy",
        "informatika-es-tavkozles": "category-informatika",
        "szakmai-szolgaltatasok": "category-szakma",
    };

    /**
     * `tiles`: the home page's "Böngéssz kategóriák szerint", one tile per main
     * category (icon and name, no subcategories), as in the owner's design.
     * @type {{ tiles?: boolean }}
     */
    let { tiles = false } = $props();

    let catalog = $state(DIRECTORY_CATALOG);
    /** @type {Array<{ category?: string }>} */
    let entries = $state([]);

    let groups = $derived(directoryGroupsForEntries(catalog, entries));

    onMount(async () => {
        try {
            const [categoryRows, directory] = await Promise.all([
                apiFetch("/api/entry-categories"),
                apiFetch("/api/directory"),
            ]);
            if (Array.isArray(categoryRows) && categoryRows.length) {
                catalog = directoryCatalogFromApi(categoryRows);
            }
            entries = Array.isArray(directory) ? directory : [];
        } catch {
            entries = [];
        }
    });
</script>

{#if tiles}
<section class="home-section directory-tiles" aria-labelledby="directory-categories-title">
    <div class="home-section__head">
        <h2 id="directory-categories-title" class="widget-title">Böngéssz kategóriák szerint</h2>
        <a class="home-section__more" href="/index">Indexelünk ›</a>
    </div>
    <ul class="directory-tiles__grid">
        {#each groups as group (group.slug)}
            <li>
                <a class="card directory-tile" href="/index/{group.slug}">
                    <AppIcon name={CATEGORY_ICONS[group.slug] || "category-default"} size={28} />
                    <span>{group.name}</span>
                </a>
            </li>
        {/each}
    </ul>
</section>
{:else}
<section class="directory-categories" aria-labelledby="directory-categories-title">
    <h2 id="directory-categories-title" class="directory-categories__title">Index bejegyzés kategóriák</h2>
    <ul class="directory-categories__grid">
        {#each groups as group (group.slug)}
            <li class="card directory-category">
                <a class="directory-category__parent" href="/index/{group.slug}">
                    <AppIcon name={CATEGORY_ICONS[group.slug] || "category-default"} size={40} />
                    <span>{group.name}</span>
                </a>
                {#if group.children.length}
                    <ul class="directory-category__children">
                        {#each group.children as child (child.slug)}
                            <li>
                                <a href="/index/{child.slug}">{child.name}</a>
                            </li>
                        {/each}
                    </ul>
                {/if}
            </li>
        {/each}
    </ul>
</section>
{/if}

<style>
    .directory-tiles__grid {
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        gap: 1rem;
        list-style: none;
        margin: 0;
        padding: 0;
    }

    .directory-tile {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        height: 100%;
        box-sizing: border-box;
        font-weight: 600;
    }

    .directory-tile span {
        min-width: 0;
        overflow-wrap: anywhere;
    }

    .directory-tile:hover {
        border-color: var(--szekely-red);
    }

    @media (max-width: 900px) {
        .directory-tiles__grid {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (max-width: 520px) {
        .directory-tiles__grid {
            grid-template-columns: 1fr;
        }
    }

    .directory-categories {
        margin: 2rem 0 1.5rem;
    }

    .directory-categories__title {
        margin: 0 0 1rem;
        text-align: center;
        font-size: var(--text-xl, 1.5rem);
        font-weight: 700;
    }

    .directory-categories__grid {
        display: grid;
        grid-template-columns: repeat(4, minmax(0, 1fr));
        gap: 1rem;
        list-style: none;
        margin: 0;
        padding: 0;
    }

    .directory-category {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 0.75rem;
        text-align: center;
        min-width: 0;
    }

    .directory-category__parent {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 0.65rem;
        color: inherit;
        text-decoration: none;
        font-weight: 600;
    }

    .directory-category__parent:hover {
        color: var(--szekely-red);
    }

    .directory-category__children {
        display: flex;
        flex-wrap: wrap;
        justify-content: center;
        gap: 0.35rem 0.75rem;
        list-style: none;
        margin: 0;
        padding: 0;
    }

    .directory-category__children a {
        color: var(--text-muted);
        text-decoration: none;
        font-size: var(--text-sm);
    }

    .directory-category__children a:hover {
        color: var(--szekely-red);
    }

    @media (max-width: 1100px) {
        .directory-categories__grid {
            grid-template-columns: repeat(3, minmax(0, 1fr));
        }
    }

    @media (max-width: 760px) {
        .directory-categories__grid {
            grid-template-columns: repeat(2, minmax(0, 1fr));
        }
    }

    @media (max-width: 520px) {
        .directory-categories__grid {
            grid-template-columns: 1fr;
        }
    }
</style>
