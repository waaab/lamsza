<script>
    import { jsonLdScript } from "$lib/structuredData.js";

    /** One schema.org node, or a list of them. Empty/null emits nothing. */
    export let data = /** @type {unknown} */ (null);

    $: script = jsonLdScript(data);
</script>

<svelte:head>
    {#if script}
        {@html script}
    {/if}
</svelte:head>
