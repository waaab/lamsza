<script>
    import {
        absoluteUrl,
        DEFAULT_OG_IMAGE,
        DEFAULT_OG_TYPE,
        DEFAULT_SEO_DESCRIPTION,
        DEFAULT_SEO_TITLE,
        SITE_LOCALE,
        SITE_NAME,
    } from "$lib/seo.js";

    /**
     * Open Graph / Twitter / canonical head tags for one page.
     * The document `<title>` stays with `PublicPageHero` - one mechanism per tag.
     * Render this once per page (the public layout already does).
     */

    /** Social card title, site name included (see `resolveSeo`). */
    export let title = DEFAULT_SEO_TITLE;
    export let description = DEFAULT_SEO_DESCRIPTION;
    /** Absolute canonical URL on the production host. */
    export let canonical = "";
    /** Share image; a relative path is made absolute. */
    export let image = DEFAULT_OG_IMAGE;
    /** `og:type`: `website` for most pages, `article` for one piece of content. */
    export let type = DEFAULT_OG_TYPE;
    export let siteName = SITE_NAME;
    export let locale = SITE_LOCALE;

    $: imageUrl = absoluteUrl(image);
</script>

<svelte:head>
    <meta name="description" content={description} />
    {#if canonical}
        <link rel="canonical" href={canonical} />
        <meta property="og:url" content={canonical} />
    {/if}
    <meta property="og:title" content={title} />
    <meta property="og:description" content={description} />
    <meta property="og:type" content={type} />
    <meta property="og:site_name" content={siteName} />
    <meta property="og:locale" content={locale} />
    <meta property="og:image" content={imageUrl} />
    <meta property="og:image:width" content="1200" />
    <meta property="og:image:height" content="630" />
    <meta property="og:image:alt" content={title} />
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content={title} />
    <meta name="twitter:description" content={description} />
    <meta name="twitter:image" content={imageUrl} />
</svelte:head>
