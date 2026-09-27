<script>
    import ClaimMark from "$lib/components/ClaimMark.svelte";

    let {
        href,
        title,
        meta = "",
        description = "",
        accent = "none",
        external = false,
        claimed = false,
    } = $props();
</script>

<a
    class="search-result-card"
    class:search-result-card--attraction={accent === "attraction"}
    class:search-result-card--seat={accent === "seat"}
    {href}
    target={external ? "_blank" : undefined}
    rel={external ? "nofollow noopener" : undefined}
>
    <span class="search-result-card__title">
        {title}
        {#if claimed}
            <ClaimMark claimed={true} />
        {/if}
    </span>
    {#if meta}
        <span class="search-result-card__meta">{meta}</span>
    {/if}
    {#if description}
        <span class="search-result-card__desc">{description}</span>
    {/if}
</a>

<style>
    .search-result-card {
        display: block;
        padding: 0.6rem 0.8rem;
        background: var(--bg-body);
        border-radius: 8px;
        border: 1px solid var(--border-color);
        color: var(--text-primary);
        text-decoration: none;
        transition: background 0.2s;
    }

    .search-result-card:hover {
        background: var(--tab-hover-bg);
    }

    .search-result-card--attraction {
        border-color: var(--szekely-brown, #8d6e63);
    }

    .search-result-card--seat {
        border-color: var(--szekely-blue, #42a5f5);
    }

    .search-result-card__title {
        display: block;
        font-weight: 500;
    }

    .search-result-card__meta,
    .search-result-card__desc {
        display: block;
        font-size: var(--text-sm);
    }

    .search-result-card__meta {
        color: var(--text-faint);
    }

    .search-result-card__desc {
        color: var(--text-muted);
        margin-top: 0.25rem;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
</style>
