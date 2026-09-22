<script>
    import { onMount, tick } from "svelte";
    import EntryCard from "$lib/components/EntryCard.svelte";

    let { items = [] } = $props();

    let scroller = $state(/** @type {HTMLDivElement | null} */ (null));
    let showNav = $state(false);

    function checkOverflow() {
        const el = scroller;
        if (!el) {
            showNav = false;
            return;
        }
        showNav = el.scrollWidth > el.clientWidth;
    }

    function scrollPrev() {
        scroller?.scrollBy({ left: -240, behavior: "smooth" });
    }

    function scrollNext() {
        scroller?.scrollBy({ left: 240, behavior: "smooth" });
    }

    /** @param {{ slug: string, name: string, category?: string, location?: string, photo?: string }} item */
    function entryFromItem(item) {
        return {
            slug: item.slug,
            name: item.name,
            category: item.category,
            location: item.location,
            photos: item.photo
                ? [
                      {
                          url: item.photo,
                          alt: item.name,
                          title: item.name,
                          width: 160,
                          height: 120,
                      },
                  ]
                : [],
        };
    }

    onMount(async () => {
        await tick();
        checkOverflow();
    });
</script>

<svelte:window onresize={checkOverflow} />

{#if items.length > 0}
    <section class="entry-history" aria-labelledby="entry-history-title">
        <div class="entry-history__header">
            <h2 id="entry-history-title" class="entry-history__title">Böngészési előzmények</h2>
            {#if showNav}
                <div class="entry-history__nav">
                    <button
                        type="button"
                        class="entry-history__btn"
                        aria-label="Előző"
                        onclick={scrollPrev}
                    >
                        ‹
                    </button>
                    <button
                        type="button"
                        class="entry-history__btn"
                        aria-label="Következő"
                        onclick={scrollNext}
                    >
                        ›
                    </button>
                </div>
            {/if}
        </div>
        <div class="entry-history__scroller" bind:this={scroller}>
            {#each items as item (item.slug)}
                <div class="entry-history__card">
                    <EntryCard entry={entryFromItem(item)} layout="grid" showBadge={false} />
                </div>
            {/each}
        </div>
    </section>
{/if}

<style>
    .entry-history {
        padding-top: 0.5rem;
        border-top: 1px solid var(--border-color);
    }
    .entry-history__header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 0.75rem;
        margin-bottom: 0.75rem;
    }
    .entry-history__title {
        margin: 0;
        font-size: 1.05rem;
    }
    .entry-history__nav {
        display: flex;
        gap: 0.35rem;
        flex-shrink: 0;
    }
    .entry-history__btn {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        width: 2rem;
        height: 2rem;
        padding: 0;
        border: 1px solid var(--border-color);
        border-radius: 6px;
        background: var(--card-bg);
        color: var(--text-primary);
        font-size: 1.25rem;
        line-height: 1;
        cursor: pointer;
    }
    .entry-history__btn:hover {
        border-color: var(--szekely-blue, #1d4ed8);
        color: var(--szekely-blue, #1d4ed8);
    }
    .entry-history__scroller {
        display: flex;
        flex-direction: row;
        gap: 1rem;
        overflow-x: auto;
        scroll-behavior: smooth;
        -webkit-overflow-scrolling: touch;
        scrollbar-width: thin;
    }
    .entry-history__card {
        flex: 0 0 auto;
        min-width: 16rem;
        max-width: 16rem;
    }
</style>
