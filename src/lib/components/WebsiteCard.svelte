<script>
    import ClaimMark from "$lib/components/ClaimMark.svelte";

    /** @type {{ website: { title: string, description: string, domain: string, url: string, claimed?: boolean } }} */
    let { website } = $props();
</script>

<article class="card entry website-card">
    <a
        href={website.url}
        target="_blank"
        rel="noopener"
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
