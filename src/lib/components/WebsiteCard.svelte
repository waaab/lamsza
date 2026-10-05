<script>
    import ClaimMark from "$lib/components/ClaimMark.svelte";

    /**
     * @type {{
     *   website: {
     *     title: string,
     *     description: string,
     *     domain: string,
     *     url: string,
     *     claimed?: boolean,
     *     entry_id?: number,
     *     entry_slug?: string,
     *   }
     * }}
     */
    let { website } = $props();

    let listingHref = $derived.by(() => {
        const slug = String(website?.entry_slug ?? "").trim();
        return slug ? `/bejegyzes/${slug}` : "";
    });
    let externalHref = $derived(String(website?.url ?? "").trim());
    let href = $derived(listingHref || externalHref);
    let isExternal = $derived(!listingHref);
</script>

<article class="card entry website-card">
    <a
        {href}
        target={isExternal ? "_blank" : undefined}
        rel={isExternal ? "noopener" : undefined}
        class="website-card__link"
    >
        <h3 class="website-card__title">
            {website.title}
            <ClaimMark claimed={Boolean(website.claimed)} />
        </h3>
        {#if website.description}
            <p class="website-card__description">{website.description}</p>
        {/if}
        <p class="website-card__domain">{website.domain}</p>
    </a>
</article>
