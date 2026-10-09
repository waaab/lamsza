<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";

    /**
     * The home page's category chips under the search box: the categories an
     * admin featured (entry_categories.featured_order, at most six), in order.
     * The row keeps its height while loading and when there are none, so the
     * page below never moves (UI_BASELINE ld-reserve-space).
     */
    /** @type {Array<{ id: number, name: string, slug: string, featured_order: number }>} */
    let chips = $state([]);

    onMount(async () => {
        try {
            const rows = await apiFetch("/api/entry-categories");
            chips = (Array.isArray(rows) ? rows : [])
                .filter((c) => c.featured_order != null && c.slug)
                .sort((a, b) => a.featured_order - b.featured_order);
        } catch {
            chips = [];
        }
    });
</script>

<nav class="home-chips" aria-label="Kiemelt kategóriák">
    {#each chips as c (c.id)}
        <a class="btn btn-sm home-chip" href="/index/{c.slug}">{c.name}</a>
    {/each}
</nav>

<style>
    .home-chips {
        display: flex;
        flex-wrap: wrap;
        justify-content: center;
        gap: 0.5rem;
        min-height: 2rem;
        margin-top: 1rem;
    }

    .home-chip {
        text-decoration: none;
        color: var(--text-muted);
    }

    /* One scrolling row on phones, so six chips never add rows after load. */
    @media (max-width: 600px) {
        .home-chips {
            flex-wrap: nowrap;
            justify-content: flex-start;
            overflow-x: auto;
            scrollbar-width: none;
        }

        .home-chip {
            flex: none;
        }
    }

    .home-chip:hover,
    .home-chip:focus-visible {
        color: var(--text-primary);
    }
</style>
