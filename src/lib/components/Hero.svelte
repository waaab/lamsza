<script>
    /**
     * The hero (UI_BASELINE "hero"): the first block of an app's home page,
     * and only there. Every other page title is a PageHeader. Shared from
     * lamsza by scripts/sync-shared-frontend.sh, so Lámsza, Szótár and
     * Játszótér open their home pages the same way: an optional drawing
     * (each app's own animated SVG), the big title, the line under it, and
     * optionally the app's search box under them.
     *
     *   <Hero title="Székely" accent="szótár" greeting="…">
     *       {#snippet art()}<DictionaryArt />{/snippet}
     *       <LookupBar />
     *   </Hero>
     *
     * `accent`: the title's second part, in the network red (UI_BASELINE
     * "hero"); a separate word, or with `joined` the end of the same word
     * ("Játszó" + "tér"). The greeting is pulled up under the title; when the title has
     * a letter that reaches below the line (g, j, p, q, y, J, Q), it is pulled
     * up less, so it never crosses that letter.
     */

    /** @type {{ title: string, accent?: string, joined?: boolean, greeting?: string, art?: import("svelte").Snippet, children?: import("svelte").Snippet }} */
    let { title, accent = "", joined = false, greeting = "", art, children } = $props();

    const descends = $derived(/[gjpqyJQ]/.test(`${title}${accent}`));
</script>

<header class="hero" class:hero--descends={descends}>
    {#if art}
        <div class="hero-art" aria-hidden="true">{@render art()}</div>
    {/if}
    <h1 class="page-title hero-title">{accent && !joined ? `${title} ` : title}{#if accent}<span class="hero-accent" class:hero-accent--joined={joined}>{accent}</span>{/if}</h1>
    {#if greeting}
        <h2 class="greeting hero-greeting">{greeting}</h2>
    {/if}
    {#if children}
        <div class="hero-search">{@render children()}</div>
    {/if}
</header>
