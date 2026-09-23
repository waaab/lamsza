<script>
    /** @type {{ website: { title: string, description: string, domain: string, url: string } }} */
    let { website } = $props();
</script>

<article class="card entry website-card">
    <a
        href={website.url}
        target="_blank"
        rel="noopener"
        class="website-card__link"
    >
        <h3 class="website-card__title">{website.title}</h3>
        {#if website.description}
            <p class="website-card__description">{website.description}</p>
        {/if}
        <p class="website-card__domain">{website.domain}</p>
    </a>
</article>
