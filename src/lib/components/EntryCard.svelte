<script>
    import { canonicalEntryType } from "$lib/entryType.js";
    import { EMPTY_PLACEHOLDER, displayInitials } from "$lib/displayValue.js";
    import { gallerySlides } from "$lib/entryPhotos.js";
    import { showListingRatings } from "$lib/entryPublicExtras.js";
    import EntryStars from "$lib/components/EntryStars.svelte";

    /**
     * Directory listing card. `layout` is `grid` (Rács) or `list` (Lista).
     * Reviews are not wired yet — `rating`, `review_count`, and `featured_review`
     * are optional placeholders.
     * @type {{
     *   entry?: Record<string, any> | null,
     *   showBadge?: boolean,
     *   placeholder?: boolean,
     *   layout?: "grid" | "list",
     * }}
     */
    let {
        entry = null,
        showBadge = true,
        placeholder = false,
        layout = "list",
    } = $props();

    let isGrid = $derived(layout === "grid");
    let isSettlement = $derived(entry?.entity_type === "settlement");
    let href = $derived.by(() => {
        const slug = String(entry?.slug ?? "").trim();
        if (!slug) return "#";
        if (isSettlement) {
            const countySlug = String(entry?.county_slug ?? "").trim();
            return countySlug ? `/${countySlug}-megye/${slug}` : `/${slug}`;
        }
        return `/bejegyzes/${slug}`;
    });
    let typeLabel = $derived(
        isSettlement
            ? "Település"
            : canonicalEntryType(entry?.type) || EMPTY_PLACEHOLDER,
    );
    let placeLines = $derived.by(() => {
        /** @type {string[]} */
        const lines = [];
        const name = String(entry?.name ?? "").trim();
        const loc = String(entry?.location ?? "").trim();
        if (loc && loc !== name) lines.push(loc);
        const addr = String(entry?.address ?? "").trim();
        if (addr && addr !== loc) lines.push(addr);
        return lines;
    });
    let placeHref = $derived.by(() => {
        const countySlug = String(entry?.county_slug ?? "").trim();
        if (isSettlement) {
            return countySlug ? `/${countySlug}-megye` : "";
        }
        const settlementSlug = String(entry?.location_slug ?? "").trim();
        if (countySlug && settlementSlug) {
            return `/${countySlug}-megye/${settlementSlug}`;
        }
        return "";
    });
    let pills = $derived.by(() => {
        /** @type {string[]} */
        const out = [];
        const cat = String(entry?.category ?? "").trim();
        if (cat && !isSettlement) out.push(cat);
        for (const raw of entry?.tags || []) {
            const s = String(raw).trim();
            if (!s) continue;
            const label = s.startsWith("#") ? s : `#${s}`;
            if (!out.includes(label) && label !== cat && `#${cat}` !== label) {
                out.push(label);
            }
        }
        return out;
    });
    let initials = $derived(displayInitials(entry?.name));
    let thumb = $derived.by(() => {
        const slide = gallerySlides(entry)[0];
        if (slide) return slide;
        const legacy = String(entry?.image ?? entry?.photo ?? entry?.logo ?? "").trim();
        if (!legacy) return null;
        return {
            src: legacy,
            alt: String(entry?.name ?? "").trim(),
            title: String(entry?.name ?? "").trim(),
            width: 160,
            height: 120,
        };
    });
    let rating = $derived.by(() => {
        const n = Number(entry?.rating);
        return Number.isFinite(n) ? Math.min(5, Math.max(0, n)) : 0;
    });
    let reviewCount = $derived.by(() => {
        const n = Number(entry?.review_count);
        return Number.isFinite(n) ? Math.max(0, Math.floor(n)) : 0;
    });
    let featuredReview = $derived(String(entry?.featured_review ?? "").trim());
    let ratingLabel = $derived(
        rating.toLocaleString("hu-HU", {
            minimumFractionDigits: 1,
            maximumFractionDigits: 1,
        }),
    );
    let reviewCountLabel = $derived(`${reviewCount} értékelés`);
</script>

{#if placeholder}
    <article
        class={[
            "card entry entry-listing entry-listing--placeholder",
            isGrid ? "entry-listing--grid" : "entry-listing--list",
        ]}
        aria-hidden="true"
        aria-busy="true"
    >
        <div class="entry-listing__row">
            <div class="entry-listing__thumb skeleton"></div>
            <div class="entry-listing__main">
                <div class="entry-listing__headline">
                    <div class="skeleton skeleton-text entry-listing__skel-title"></div>
                    <div class="skeleton skeleton-text entry-listing__skel-place"></div>
                </div>
                <div class="skeleton skeleton-text entry-listing__skel-rating"></div>
                <div class="entry-listing__pills">
                    <span class="skeleton entry-listing__skel-pill"></span>
                    <span class="skeleton entry-listing__skel-pill"></span>
                </div>
                <div class="skeleton skeleton-text entry-listing__skel-review"></div>
            </div>
        </div>
    </article>
{:else}
    <article
        class={[
            "card entry entry-listing",
            isGrid ? "entry-listing--grid" : "entry-listing--list",
        ]}
    >
        <div class="entry-listing__row">
            <a
                class="entry-listing__thumb"
                href={href}
                tabindex="-1"
                aria-hidden="true"
            >
                {#if thumb}
                    <img
                        src={thumb.src}
                        alt={thumb.alt || ""}
                        title={thumb.title || thumb.alt || ""}
                        width={thumb.width}
                        height={thumb.height}
                        loading="lazy"
                        decoding="async"
                    />
                {:else}
                    <span class="entry-listing__initials">{initials}</span>
                {/if}
            </a>

            <div class="entry-listing__main">
                <div class="entry-listing__headline">
                    <div class="entry-listing__heading">
                        {#if showBadge && entry?.is_direct_match}
                            <div class="badge">Közvetlen Találat</div>
                        {/if}
                        <h3 class="entry-listing__title">
                            <a href={href} class="entry-listing__name">{entry?.name}</a>
                        </h3>
                    </div>
                    {#if placeLines.length}
                        <p class="entry-listing__place">
                            {#each placeLines as line, i (line)}
                                {#if i === 0 && placeHref}
                                    <a href={placeHref} class="entry-listing__place-link"
                                        >{line}</a
                                    >
                                {:else}
                                    <span>{line}</span>
                                {/if}
                            {/each}
                        </p>
                    {/if}
                </div>

                {#if showListingRatings(entry)}
                    <div class="entry-listing__rating">
                        <EntryStars {rating} />
                        <span class="entry-listing__score">{ratingLabel}</span>
                        <span class="entry-listing__count">({reviewCountLabel})</span>
                    </div>
                {/if}

                {#if pills.length}
                    <ul class="entry-listing__pills">
                        {#each pills as pill (pill)}
                            <li>{pill}</li>
                        {/each}
                    </ul>
                {/if}

                {#if typeLabel !== EMPTY_PLACEHOLDER}
                    <p class="entry-listing__type">
                        <span class="entry-listing__check" aria-hidden="true">
                            <svg viewBox="0 0 24 24" width="14" height="14">
                                <path
                                    d="M20 6 9 17l-5-5"
                                    fill="none"
                                    stroke="currentColor"
                                    stroke-width="2.4"
                                    stroke-linecap="round"
                                    stroke-linejoin="round"
                                />
                            </svg>
                        </span>
                        {typeLabel}
                    </p>
                {/if}

                {#if showListingRatings(entry)}
                    <p class="entry-listing__review">
                        <span class="entry-listing__quote-icon" aria-hidden="true">
                            <svg
                                viewBox="0 0 24 24"
                                width="16"
                                height="16"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                            >
                                <path
                                    d="M21 15a4 4 0 0 1-4 4H8l-5 3V7a4 4 0 0 1 4-4h10a4 4 0 0 1 4 4z"
                                />
                            </svg>
                        </span>
                        <span class="entry-listing__review-body">
                            {#if featuredReview}
                                <span class="entry-listing__review-text"
                                    >„{featuredReview}”</span
                                >
                            {:else}
                                <span
                                    class="entry-listing__review-text entry-listing__review-text--empty"
                                    >Még nincs értékelés.</span
                                >
                            {/if}
                            {" "}
                            <a href={href} class="entry-listing__more">tovább</a>
                        </span>
                    </p>
                {/if}
            </div>
        </div>
    </article>
{/if}

<style>
    .entry-listing {
        container-type: inline-size;
        color: var(--text-primary);
    }
    :global(.card.entry.entry-listing:hover) {
        color: var(--text-primary);
    }
    .entry-listing__row {
        display: flex;
        align-items: flex-start;
        gap: 1rem;
    }
    .entry-listing__thumb {
        flex: 0 0 6rem;
        width: 6rem;
        height: 6rem;
        border-radius: 8px;
        overflow: hidden;
        background: color-mix(in srgb, var(--szekely-green) 12%, var(--card-bg));
        color: var(--szekely-green);
        display: flex;
        align-items: center;
        justify-content: center;
        text-decoration: none;
    }
    .entry-listing__thumb img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        display: block;
    }
    .entry-listing__initials {
        font-weight: 800;
        letter-spacing: 0.02em;
    }
    .entry-listing__main {
        flex: 1;
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
    }
    .entry-listing__headline {
        display: flex;
        justify-content: space-between;
        gap: 0.75rem;
        align-items: flex-start;
    }
    .entry-listing__heading {
        min-width: 0;
        flex: 1;
    }
    .entry-listing__title {
        margin: 0;
        font-weight: 700;
        color: var(--text-primary);
        line-height: 1.25;
    }
    .entry-listing__name {
        color: inherit;
        text-decoration: none;
    }
    .entry-listing__name:hover {
        color: var(--szekely-red);
    }
    .entry-listing__place {
        margin: 0;
        text-align: right;
        color: var(--text-muted);
        line-height: 1.3;
        flex: 0 0 auto;
        max-width: 12rem;
    }
    .entry-listing__place span,
    .entry-listing__place-link {
        display: block;
    }
    .entry-listing__place-link {
        color: inherit;
        text-decoration: none;
    }
    .entry-listing__place-link:hover {
        color: var(--szekely-red);
        text-decoration: underline;
    }
    .entry-listing__rating {
        display: flex;
        align-items: center;
        gap: 0.35rem;
        flex-wrap: wrap;
    }
    .entry-listing__score {
        font-weight: 700;
        color: var(--text-primary);
    }
    .entry-listing__count {
        color: var(--text-muted);
    }
    .entry-listing__pills {
        display: flex;
        flex-wrap: wrap;
        gap: 0.35rem;
        list-style: none;
        margin: 0.1rem 0 0;
        padding: 0;
    }
    .entry-listing__pills li {
        background: var(--entry-category-bg);
        color: var(--text-secondary);
        padding: 0.15rem 0.55rem;
        border-radius: 999px;
        font-weight: 600;
    }
    .entry-listing__type {
        display: flex;
        align-items: center;
        gap: 0.35rem;
        margin: 0.1rem 0 0;
        color: var(--text-primary);
        font-weight: 600;
    }
    .entry-listing__check {
        color: var(--google-green);
        display: inline-flex;
        flex-shrink: 0;
    }
    .entry-listing__review {
        display: flex;
        align-items: flex-start;
        gap: 0.4rem;
        margin: 0.1rem 0 0;
        color: var(--text-secondary);
        line-height: 1.4;
    }
    .entry-listing__quote-icon {
        flex-shrink: 0;
        margin-top: 0.12rem;
        color: var(--text-muted);
        display: inline-flex;
    }
    .entry-listing__review-body {
        min-width: 0;
    }
    .entry-listing__review-text {
        font-style: italic;
    }
    .entry-listing__review-text--empty {
        font-style: normal;
        color: var(--text-faint);
    }
    .entry-listing__more {
        font-weight: 700;
        font-style: normal;
        color: var(--szekely-blue);
        text-decoration: none;
        white-space: nowrap;
    }
    .entry-listing__more:hover {
        text-decoration: underline;
    }
    .entry-listing--placeholder .entry-listing__thumb {
        background: var(--skeleton-bg);
    }
    .entry-listing__skel-title {
        width: 55%;
        height: 1rem;
    }
    .entry-listing__skel-place {
        width: 5.5rem;
        height: 0.7rem;
    }
    .entry-listing__skel-rating {
        width: 9rem;
        height: 0.7rem;
        margin-top: 0.15rem;
    }
    .entry-listing__skel-pill {
        width: 4.5rem;
        height: 1.15rem;
        border-radius: 999px;
    }
    .entry-listing__skel-review {
        width: 85%;
        height: 0.7rem;
        margin-top: 0.2rem;
    }

    /* Rács: stacked tile so a 3-column grid stays readable */
    .entry-listing--grid .entry-listing__row {
        flex-direction: column;
        gap: 0.75rem;
    }
    .entry-listing--grid .entry-listing__thumb {
        width: 100%;
        height: 7.5rem;
        flex: none;
    }
    .entry-listing--grid .entry-listing__headline {
        flex-direction: column;
        gap: 0.2rem;
    }
    .entry-listing--grid .entry-listing__place {
        text-align: left;
        max-width: none;
    }
    .entry-listing--grid .entry-listing__review-text {
        display: -webkit-box;
        -webkit-line-clamp: 2;
        line-clamp: 2;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    /* Lista: keep the horizontal directory row; stack only when the card itself is narrow */
    @container (max-width: 28rem) {
        .entry-listing--list .entry-listing__headline {
            flex-direction: column;
            gap: 0.2rem;
        }
        .entry-listing--list .entry-listing__place {
            text-align: left;
            max-width: none;
        }
    }
    @container (max-width: 18rem) {
        .entry-listing--list .entry-listing__row {
            flex-direction: column;
        }
        .entry-listing--list .entry-listing__thumb {
            width: 100%;
            height: 6.5rem;
            flex-basis: auto;
        }
    }
</style>
