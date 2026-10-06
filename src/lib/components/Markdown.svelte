<script>
    // renderMarkdown escapes raw HTML and refuses unsafe link targets, so the
    // {@html} below cannot run a script even when `source` is user-written
    // prose from an attraction suggestion. Never call the plain marked
    // renderer here: it passes raw HTML straight through.
    import { renderMarkdown } from "$lib/markdown.js";

    export let source = "";

    $: html = renderMarkdown(source);
</script>

{#if html}
    <div class="markdown-content">{@html html}</div>
{/if}

<style>
    .markdown-content :global(p) {
        margin: 0.5rem 0;
    }
    .markdown-content :global(h1),
    .markdown-content :global(h2),
    .markdown-content :global(h3) {
        margin-top: 1rem;
        margin-bottom: 0.5rem;
    }
    .markdown-content :global(ul),
    .markdown-content :global(ol) {
        margin: 0.5rem 0;
        padding-left: 1.5rem;
    }
    .markdown-content :global(a) {
        text-decoration: underline;
    }
</style>