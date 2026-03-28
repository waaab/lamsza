<script>
    import { marked } from "marked";

    export let source = "";

    $: html = source ? marked.parse(source, { async: false }) : "";
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
        color: var(--link-color, #2563eb);
        text-decoration: underline;
    }
</style>
