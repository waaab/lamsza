<script>
    /**
     * The hero (UI_BASELINE "hero"): the first block of an app's home page,
     * and only there. Every other page title is a PageHeader. Shared from
     * lamsza by scripts/sync-shared-frontend.sh, so Lámsza, Szótár and
     * Játszótér open their home pages the same way: an optional drawing
     * (each app's own animated SVG), the big title, the line under it.
     *
     *   <Hero title="Székely szótár" greeting="…">
     *       {#snippet art()}<DictionaryArt />{/snippet}
     *   </Hero>
     */

    /** @type {{ title: string, greeting?: string, art?: import("svelte").Snippet }} */
    let { title, greeting = "", art } = $props();
</script>

<header class="hero">
    {#if art}
        <div class="hero-art" aria-hidden="true">{@render art()}</div>
    {/if}
    <h1 class="page-title hero-title">{title}</h1>
    {#if greeting}
        <h2 class="greeting hero-greeting">{greeting}</h2>
    {/if}
</header>
