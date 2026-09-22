<script>
    import { canonicalEntryType } from "$lib/entryType.js";
    import { displayText, EMPTY_PLACEHOLDER } from "$lib/displayValue.js";

    /** @type {{ entry: Record<string, unknown> }} */
    let { entry } = $props();

    let typeLabel = $derived(displayText(canonicalEntryType(entry?.type)));
    let categoryLabel = $derived(displayText(entry?.category));
    let locationName = $derived(displayText(entry?.location));
    let locationHref = $derived.by(() => {
        const countySlug = String(entry?.county_slug ?? "").trim();
        const settlementSlug = String(entry?.location_slug ?? "").trim();
        if (!countySlug || !settlementSlug) return "";
        return `/${countySlug}-megye/${settlementSlug}`;
    });
    let url = $derived(String(entry?.url ?? "").trim());
    let phone = $derived(String(entry?.phone ?? "").trim());
    let phoneHref = $derived(phone.replace(/[^0-9+]/g, ""));
    let address = $derived(displayText(entry?.address));
    let notes = $derived(displayText(entry?.notes));
    let tags = $derived(
        Array.isArray(entry?.tags)
            ? entry.tags.map((t) => String(t).trim()).filter(Boolean)
            : [],
    );
    let languages = $derived(displayText(entry?.languages));
</script>

<dl class="entry-fields">
    <div class="entry-fields__row">
        <dt>Típus</dt>
        <dd class={{ "entry-fields__empty": typeLabel === EMPTY_PLACEHOLDER }}>
            {typeLabel}
        </dd>
    </div>
    <div class="entry-fields__row">
        <dt>Település</dt>
        <dd class={{ "entry-fields__empty": locationName === EMPTY_PLACEHOLDER }}>
            {#if locationHref && locationName !== EMPTY_PLACEHOLDER}
                <a href={locationHref} class="entry-fields__link">{locationName}</a>
            {:else}
                {locationName}
            {/if}
        </dd>
    </div>
    <div class="entry-fields__row">
        <dt>Kategória</dt>
        <dd class={{ "entry-fields__empty": categoryLabel === EMPTY_PLACEHOLDER }}>
            {categoryLabel}
        </dd>
    </div>
    <div class="entry-fields__row">
        <dt>Weblap URL</dt>
        <dd class={{ "entry-fields__empty": !url }}>
            {#if url}
                <a
                    href={url}
                    target="_blank"
                    rel="nofollow noopener"
                    class="entry-fields__link">{url}</a
                >
            {:else}
                {EMPTY_PLACEHOLDER}
            {/if}
        </dd>
    </div>
    <div class="entry-fields__row">
        <dt>Telefon</dt>
        <dd class={{ "entry-fields__empty": !phone }}>
            {#if phone}
                <a href={`tel:${phoneHref}`} class="entry-fields__link">{phone}</a>
            {:else}
                {EMPTY_PLACEHOLDER}
            {/if}
        </dd>
    </div>
    <div class="entry-fields__row">
        <dt>Cím</dt>
        <dd class={{ "entry-fields__empty": address === EMPTY_PLACEHOLDER }}>
            {address}
        </dd>
    </div>
    <div class="entry-fields__row entry-fields__row--block">
        <dt>Megjegyzések</dt>
        <dd
            class={[
                "entry-fields__notes",
                notes === EMPTY_PLACEHOLDER && "entry-fields__empty",
            ]}
        >
            {notes}
        </dd>
    </div>
    <div class="entry-fields__row entry-fields__row--block">
        <dt>Címkék</dt>
        <dd>
            {#if tags.length}
                <div class="entry-tags">
                    {#each tags as t (t)}
                        <span class="entry-tag">{t.startsWith("#") ? t : "#" + t}</span>
                    {/each}
                </div>
            {:else}
                <span class="entry-fields__empty">{EMPTY_PLACEHOLDER}</span>
            {/if}
        </dd>
    </div>
    <div class="entry-fields__row">
        <dt>Nyelvek</dt>
        <dd class={{ "entry-fields__empty": languages === EMPTY_PLACEHOLDER }}>
            {languages}
        </dd>
    </div>
</dl>

<style>
    .entry-fields {
        display: grid;
        gap: 0.65rem 1rem;
        margin: 0;
    }
    .entry-fields__row {
        display: grid;
        grid-template-columns: minmax(7.5rem, 11rem) minmax(0, 1fr);
        gap: 0.5rem 1rem;
        align-items: start;
    }
    .entry-fields__row dt {
        margin: 0;
        color: var(--text-faint);
        font-weight: 600;
    }
    .entry-fields__row dd {
        margin: 0;
        min-width: 0;
        overflow-wrap: anywhere;
    }
    .entry-fields__empty {
        color: var(--text-faint);
    }
    .entry-fields__notes {
        white-space: pre-wrap;
    }
    .entry-fields__link {
        color: var(--primary-color, var(--szekely-blue, #0059b3));
        font-weight: 500;
        text-decoration: none;
    }
    .entry-fields .entry-tags {
        margin-top: 0;
    }
    @media (max-width: 520px) {
        .entry-fields__row {
            grid-template-columns: 1fr;
            gap: 0.15rem;
        }
    }
</style>
