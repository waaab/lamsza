<script>
    export let entry;
    export let showBadge = true;
</script>

<article class="card entry">
    {#if showBadge}
        {#if entry.is_direct_match}
            <div class="badge">Közvetlen Találat</div>
        {:else if entry.entity_type === "settlement"}
            <div class="badge">Település</div>
        {:else}
            <div class="badge">Index: {entry.category}</div>
        {/if}
    {/if}

    <h3 class="entry-name">
        {#if entry.entity_type === "settlement"}
            <a href="/{entry.county_slug}-megye/{entry.slug}" class="entry-link"
                >{entry.name}</a
            >
        {:else}
            <a href="/bejegyzes/{entry.slug}" class="entry-link">{entry.name}</a
            >
        {/if}
    </h3>

    {#if entry.url}
        <div class="entry-info entry-url-box">
            <span class="entry-url-icon">🔗</span>
            <a
                href={entry.url}
                target="_blank"
                rel="nofollow noopener"
                class="entry-url-link">{entry.url}</a
            >
        </div>
    {/if}

    <div class="entry-info">📍 
    <a href="/{entry.county_slug}-megye/{entry.location_slug}" class="entry-link">
        {entry.location}
    </a>
        {#if entry.address}- {entry.address}{/if}
    </div>

    {#if entry.phone}
        <div class="entry-info">📞 {entry.phone}</div>
    {/if}

    {#if entry.tags && entry.tags.length > 0}
        <div class="entry-tags">
            {#each entry.tags as t}
                <span class="entry-tag">{t.startsWith("#") ? t : "#" + t}</span>
            {/each}
        </div>
    {/if}

    {#if entry.notes}
        <div class="entry-notes">{entry.notes}</div>
    {/if}
</article>

<style>
    .entry-url-box {
        display: flex;
        align-items: center;
        gap: 0.3rem;
        margin-bottom: 0.5rem;
    }
    .entry-url-icon {
        color: var(--text-faint);
    }
    .entry-url-link {
        color: var(--primary-color);
        font-weight: 500;
    }
</style>
