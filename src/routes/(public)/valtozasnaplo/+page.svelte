<script>
    import { onMount } from "svelte";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";
    import { PUBLIC_CHANGELOG } from "$lib/publicChangelog.js";

    let pageHeader = initialPageHeader("valtozasnaplo");
    let pageHeaderLoading = false;

    onMount(async () => {
        pageHeader = await loadPageMeta("valtozasnaplo");
        pageHeaderLoading = false;
    });
</script>

<svelte:head>
    <meta
        name="description"
        content="Székely Gugel változásnapló - új funkciók és fejlesztések listája."
    />
</svelte:head>

<PublicPageHero
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    loading={pageHeaderLoading}
    showBreadcrumbs={false}
    documentTitleSuffix=" - Székely Gugel"
/>

<section class="faq" id="gyik">
    <h2 class="faq-title">Aplikáció verziók</h2>
    <div class="faq-list">
        {#each PUBLIC_CHANGELOG as entry, index (entry.version)}
            <details class="faq-item" open={index === 0}>
                <summary>v{entry.version} - {entry.date}</summary>
                <ul>
                    {#each entry.items as item}
                        <li>
                            {#if typeof item === "string"}
                                {item}
                            {:else}
                                <strong>{item.lead}</strong>: {item.text}
                            {/if}
                        </li>
                    {/each}
                </ul>
            </details>
        {/each}
    </div>
</section>