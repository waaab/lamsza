<script>
    import ClaimMark from "$lib/components/ClaimMark.svelte";
    import FavoriteButton from "$lib/components/FavoriteButton.svelte";
    import { canonicalEntryType } from "$lib/entryType.js";
    import { displayText, displayInitials, EMPTY_PLACEHOLDER } from "$lib/displayValue.js";
    import {
        WEEKDAYS,
        formatDayHours,
        normalizeHours,
        openStatus,
        todayWeekdayKey,
    } from "$lib/entryHours.js";
    import {
        ENTRY_PHOTO_SLOTS,
    } from "$lib/entryDemoPhotos.js";
    import { gallerySlides } from "$lib/entryPhotos.js";
    import {
        listingTextExpanded,
        showListingPhotos,
        showListingHours,
        showListingDeliveryHours,
        showListingLanguages,
        showListingPhone,
        showListingRatings,
        showListingSocial,
        showListingTodayHours,
        showListingWebsite,
    } from "$lib/entryPublicExtras.js";
    import EntryStars from "$lib/components/EntryStars.svelte";
    import EntryPhotoGallery from "$lib/components/EntryPhotoGallery.svelte";
    import EntryReviews from "$lib/components/EntryReviews.svelte";
    import JsonLd from "$lib/components/JsonLd.svelte";
    import { listingJsonLd } from "$lib/structuredData.js";
    import { displayTagLabel } from "$lib/directoryTagCloud.js";

    /** @type {{
     *   entry?: Record<string, any> | null,
     *   placeholder?: boolean,
     *   loggedIn?: boolean,
     *   isFavorite?: boolean,
     *   onFavorite?: () => void,
     *   showClaim?: boolean,
     *   showClaimWaiting?: boolean,
     *   showJoin?: boolean,
     *   showJoinWaiting?: boolean,
     *   claimBusy?: boolean,
     *   onClaim?: () => void,
     *   onJoin?: () => void,
     *   showOwnerEdit?: boolean,
     *   suggestionState?: "hidden" | "open" | "waiting",
     *   onSuggest?: () => void,
     *   onEdit?: () => void,
     * }} */
    let {
        entry = null,
        placeholder = false,
        loggedIn = false,
        isFavorite = false,
        onFavorite = () => {},
        showClaim = false,
        showClaimWaiting = false,
        showJoin = false,
        showJoinWaiting = false,
        claimBusy = false,
        onClaim = () => {},
        onJoin = () => {},
        showOwnerEdit = false,
        suggestionState = "hidden",
        onSuggest = () => {},
        onEdit = () => {},
    } = $props();

    let typeLabel = $derived(displayText(canonicalEntryType(entry?.type)));
    let categoryLabels = $derived.by(() => {
        const many = Array.isArray(entry?.categories)
            ? entry.categories.map((c) => String(c).trim()).filter(Boolean)
            : [];
        if (many.length) {
            return many
                .map((c) => displayText(c))
                .filter((c) => c !== EMPTY_PLACEHOLDER);
        }
        const one = displayText(entry?.category);
        return one !== EMPTY_PLACEHOLDER ? [one] : [];
    });
    let categoryLabel = $derived(categoryLabels[0] ?? EMPTY_PLACEHOLDER);
    let locationName = $derived(displayText(entry?.location));
    let countyName = $derived(displayText(entry?.location_county));
    let countySlug = $derived(String(entry?.county_slug ?? "").trim());
    let settlementSlug = $derived(String(entry?.location_slug ?? "").trim());
    let countyHref = $derived(countySlug ? `/${countySlug}-megye` : "");
    let locationHref = $derived.by(() => {
        if (!countySlug || !settlementSlug) return "";
        return `/${countySlug}-megye/${settlementSlug}`;
    });
    let showPlaceLinks = $derived.by(() => {
        const countyOk =
            Boolean(countyHref) && countyName !== EMPTY_PLACEHOLDER;
        const settlementOk =
            Boolean(locationHref) && locationName !== EMPTY_PLACEHOLDER;
        return countyOk || settlementOk;
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
    let tagLabels = $derived(tags.map((t) => displayTagLabel(t)));
    let jsonLd = $derived(placeholder ? null : listingJsonLd(entry));
    let languages = $derived(displayText(entry?.languages));
    let initials = $derived(displayInitials(entry?.name));
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
    let coordinates = $derived.by(() => {
        const raw = String(entry?.location_coordinates ?? "").trim();
        if (!raw) return null;
        const parts = raw.split(",").map((part) => part.trim());
        if (parts.length !== 2) return null;
        const lat = Number(parts[0]);
        const lng = Number(parts[1]);
        if (!Number.isFinite(lat) || !Number.isFinite(lng)) return null;
        return { lat, lng };
    });
    let mapsHref = $derived.by(() => {
        if (coordinates) {
            return `https://www.google.com/maps/search/?api=1&query=${coordinates.lat},${coordinates.lng}`;
        }
        const parts = [];
        if (address !== EMPTY_PLACEHOLDER) parts.push(address);
        if (locationName !== EMPTY_PLACEHOLDER) parts.push(locationName);
        if (!parts.length) return "";
        return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(parts.join(", "))}`;
    });
    let showMapsPin = $derived(Boolean(coordinates));
    let textExpanded = $derived(listingTextExpanded(entry));
    let socialLinks = $derived.by(() => {
        if (!showListingSocial(entry)) return [];
        return (Array.isArray(entry?.social_links) ? entry.social_links : [])
            .map((row) => ({
                label: String(row?.label ?? "").trim(),
                url: String(row?.url ?? "").trim(),
            }))
            .filter((row) => row.url);
    });
    let hostLabel = $derived.by(() => {
        if (!url) return "";
        try {
            return new URL(url).host.replace(/^www\./, "");
        } catch {
            return url.replace(/^https?:\/\//, "").replace(/\/$/, "");
        }
    });
    let claimed = $derived(Boolean(entry?.claimed));
    let verified = $derived(Boolean(entry?.verified));
    let hours = $derived(normalizeHours(entry?.hours));
    let deliveryHours = $derived(normalizeHours(entry?.delivery_hours));
    let todayKey = $derived(todayWeekdayKey());
    let status = $derived(openStatus(hours));
    let slides = $derived(gallerySlides(entry));
    let hasContactBlock = $derived(
        showListingWebsite(entry) ||
            showListingPhone(entry) ||
            socialLinks.length > 0 ||
            Boolean(showMapsPin && mapsHref) ||
            showListingLanguages(entry) ||
            showClaim ||
            showClaimWaiting ||
            showJoin ||
            showJoinWaiting ||
            showOwnerEdit ||
            suggestionState === "open" ||
            suggestionState === "waiting",
    );
</script>

{#snippet suggestEdit()}
    {#if showOwnerEdit}
        <button type="button" class="entry-profile__suggest" onclick={() => onEdit()}>
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M12 20h9" />
                <path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4 12.5-12.5z" />
            </svg>
            Szerkesztés
        </button>
    {:else if suggestionState === "open"}
        <button type="button" class="entry-profile__suggest" onclick={() => onSuggest()}>
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M12 20h9" />
                <path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4 12.5-12.5z" />
            </svg>
            Javaslat módosításra
        </button>
    {:else if suggestionState === "waiting"}
        <button type="button" class="entry-profile__suggest" disabled>
            <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                <path d="M12 20h9" />
                <path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4 12.5-12.5z" />
            </svg>
            Javaslat elküldve
        </button>
    {/if}
{/snippet}

<JsonLd data={jsonLd} />

{#if placeholder}
    <article
        class="entry-profile entry-profile--placeholder"
        aria-busy="true"
        aria-label="Bejegyzés betöltése"
    >
        <header class="entry-profile__hero">
            <div class="skeleton entry-profile__skel-badge"></div>
            <div class="skeleton skeleton-text entry-profile__skel-title"></div>
            <div class="skeleton skeleton-text entry-profile__skel-line"></div>
            <div class="skeleton skeleton-text entry-profile__skel-line entry-profile__skel-line--short"></div>
        </header>
        <div class="entry-profile__layout">
            <div class="entry-profile__main">
                <section class="entry-profile__section">
                    <div class="skeleton skeleton-text entry-profile__skel-heading"></div>
                    <div class="entry-profile__photos-skel">
                        <div class="skeleton entry-profile__skel-stage"></div>
                        <div class="entry-profile__skel-thumbs">
                            {#each { length: ENTRY_PHOTO_SLOTS }}
                                <div class="skeleton entry-profile__skel-thumb"></div>
                            {/each}
                        </div>
                    </div>
                </section>
                <section class="entry-profile__section">
                    <div class="skeleton skeleton-text entry-profile__skel-heading"></div>
                    <div class="skeleton skeleton-text entry-profile__skel-line entry-profile__skel-line--mid"></div>
                </section>
                <section class="entry-profile__section">
                    <div class="skeleton skeleton-text entry-profile__skel-heading"></div>
                    <div class="skeleton skeleton-text entry-profile__skel-line"></div>
                    <div class="skeleton skeleton-text entry-profile__skel-line entry-profile__skel-line--mid"></div>
                    <div class="skeleton skeleton-text entry-profile__skel-line entry-profile__skel-line--short"></div>
                </section>
                <section class="entry-profile__section">
                    <div class="skeleton skeleton-text entry-profile__skel-heading"></div>
                    <div class="skeleton skeleton-text entry-profile__skel-line"></div>
                    <div class="skeleton skeleton-text entry-profile__skel-line"></div>
                </section>
                <section class="entry-profile__section">
                    <div class="skeleton skeleton-text entry-profile__skel-heading"></div>
                    {#each { length: 7 }}
                        <div class="skeleton skeleton-text entry-profile__skel-hour"></div>
                    {/each}
                </section>
            </div>
            <aside class="entry-profile__aside" aria-hidden="true">
                <div class="sidebar-box">
                    <div class="entry-profile__contact">
                        {#each { length: 4 }}
                            <div class="entry-profile__contact-row">
                                <span class="skeleton entry-profile__skel-icon"></span>
                                <span class="skeleton skeleton-text entry-profile__skel-contact"></span>
                            </div>
                        {/each}
                    </div>
                </div>
            </aside>
        </div>
    </article>
{:else}
<article class="entry-profile">
    <header class="entry-profile__hero">
        {#if loggedIn}
            <div class="entry-profile__favorite">
                <FavoriteButton
                    type="entry"
                    id={entry.id}
                    active={isFavorite}
                    ontoggle={onFavorite}
                />
            </div>
        {/if}
        {#if categoryLabel !== EMPTY_PLACEHOLDER}
            <div class="badge">Index: {categoryLabel}</div>
        {/if}
        <h1 class="entry-profile__title">{entry.name}</h1>
        <p class="entry-profile__meta">
            {#if typeLabel !== EMPTY_PLACEHOLDER}
                <span>{typeLabel}</span>
            {/if}
            {#each categoryLabels as label, i (label)}
                {#if typeLabel !== EMPTY_PLACEHOLDER || i > 0}
                    <span class="entry-profile__dot" aria-hidden="true">·</span>
                {/if}
                <span>{label}</span>
            {/each}
        </p>
        <p class="entry-profile__status-row">
            <span
                class={[
                    "entry-profile__claim",
                    verified
                        ? "entry-profile__claim--verified"
                        : "entry-profile__claim--unverified",
                ]}
            >
                <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.25" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <path d="M9 12.5 11.2 14.7 15.5 9.8" />
                    <circle cx="12" cy="12" r="9" />
                </svg>
                {verified ? "Ellenőrzött" : "Nem ellenőrzött"}
            </span>
            <span class="entry-profile__dot" aria-hidden="true">·</span>
            <ClaimMark {claimed} showLabel />
            {#if showListingTodayHours(entry)}
                <span class="entry-profile__dot" aria-hidden="true">·</span>
                <span class="entry-profile__today">Nyitvatartás ma</span>
                <span
                    class={[
                        "entry-profile__open",
                        `entry-profile__open--${status.state}`,
                    ]}
                >
                    {#if status.state === "unknown"}
                        <span class="entry-profile__open-detail">-</span>
                    {:else}
                        <span>{status.label}</span>
                        {#if status.detail}
                            <span class="entry-profile__open-detail">{status.detail}</span>
                        {/if}
                    {/if}
                </span>
            {/if}
        </p>
        {#if showListingRatings(entry)}
            <div class="entry-profile__rating">
                <EntryStars {rating} size={18} />
                <span class="entry-profile__score">{ratingLabel}</span>
                <span class="entry-profile__count">({reviewCountLabel})</span>
            </div>
        {/if}
    </header>

    <div class="entry-profile__layout">
        <div class="entry-profile__main">
            {#if showListingPhotos(entry)}
                <section class="entry-profile__section" aria-labelledby="entry-photos-title">
                    <h2 id="entry-photos-title" class="entry-profile__section-title">Fotók</h2>
                    <EntryPhotoGallery slides={slides} label={`${entry.name} fotói`} />
                </section>
            {/if}

            <section class="entry-profile__section" aria-labelledby="entry-services-title">
                <h2 id="entry-services-title" class="entry-profile__section-title">
                    Szolgáltatások
                </h2>
                {#if categoryLabels.length}
                    <ul class="entry-profile__services">
                        {#each categoryLabels as label (label)}
                            <li>{label}</li>
                        {/each}
                    </ul>
                {:else}
                    <p class="entry-profile__empty">{EMPTY_PLACEHOLDER}</p>
                {/if}
            </section>

            <section class="entry-profile__section" aria-labelledby="entry-about-title">
                <h2 id="entry-about-title" class="entry-profile__section-title">
                    Bemutatkozás
                </h2>
                <p
                    class={[
                        "entry-profile__about",
                        notes === EMPTY_PLACEHOLDER && "entry-profile__empty",
                        !textExpanded && "entry-profile__about--clamp",
                    ]}
                >
                    {notes}
                </p>
            </section>

            <section class="entry-profile__section" aria-labelledby="entry-place-title">
                <h2 id="entry-place-title" class="entry-profile__section-title">
                    Helyszín és nyitvatartás
                </h2>
                <div class="entry-profile__place">
                    <p class="entry-profile__address">
                        {#if locationHref && locationName !== EMPTY_PLACEHOLDER}
                            <a href={locationHref} class="entry-profile__link">{locationName}</a>
                        {:else}
                            <span class={{ "entry-profile__empty": locationName === EMPTY_PLACEHOLDER }}
                                >{locationName}</span
                            >
                        {/if}
                        {#if address !== EMPTY_PLACEHOLDER}
                            <span>{address}</span>
                        {/if}
                    </p>
                    {#if mapsHref}
                        <a
                            class="btn btn-sm"
                            href={mapsHref}
                            target="_blank"
                            rel="nofollow noopener"
                        >Útvonal</a>
                    {/if}
                </div>
                {#if showListingHours(entry)}
                    <div class="entry-profile__hours-block">
                        <h3 class="entry-profile__hours-title">Nyitvatartás</h3>
                        <table class="entry-profile__hours">
                            <tbody>
                                {#each WEEKDAYS as day (day.key)}
                                    <tr class:entry-profile__hours-today={day.key === todayKey}>
                                        <th scope="row">{day.label}</th>
                                        <td>{formatDayHours(hours[day.key])}</td>
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}
                {#if showListingDeliveryHours(entry)}
                    <div class="entry-profile__hours-block">
                        <h3 class="entry-profile__hours-title">Kiszállítás</h3>
                        <table class="entry-profile__hours">
                            <tbody>
                                {#each WEEKDAYS as day (day.key)}
                                    <tr class:entry-profile__hours-today={day.key === todayKey}>
                                        <th scope="row">{day.label}</th>
                                        <td>{formatDayHours(deliveryHours[day.key])}</td>
                                    </tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                {/if}
            </section>

            {#if showListingRatings(entry)}
                <section class="entry-profile__section" aria-labelledby="entry-reviews-title">
                    <EntryReviews {entry} />
                </section>
            {/if}

            {#if socialLinks.length}
                <section class="entry-profile__section" aria-labelledby="entry-social-title">
                    <h2 id="entry-social-title" class="entry-profile__section-title">Közösségi oldalak</h2>
                    <ul class="entry-profile__social">
                        {#each socialLinks as link (link.url)}
                            <li>
                                <a href={link.url} target="_blank" rel="noopener noreferrer">{link.label || link.url}</a>
                            </li>
                        {/each}
                    </ul>
                </section>
            {/if}
        </div>

        <aside
            class="entry-profile__aside"
            aria-label="Elérhetőség, helyszín és címkék"
        >
            <div class="sidebar-box">
                {#if hasContactBlock}
                    <div class="entry-profile__contact">
                        {#if showListingWebsite(entry)}
                            <a
                                class="entry-profile__contact-row"
                                href={url}
                                target="_blank"
                                rel="nofollow noopener"
                            >
                                <span class="entry-profile__contact-icon" aria-hidden="true">
                                    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                        <path d="M18 13v6a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                                        <polyline points="15 3 21 3 21 9" />
                                        <line x1="10" y1="14" x2="21" y2="3" />
                                    </svg>
                                </span>
                                <span>{hostLabel}</span>
                            </a>
                        {/if}
                        {#if showListingPhone(entry)}
                            <a class="entry-profile__contact-row" href={`tel:${phoneHref}`}>
                                <span class="entry-profile__contact-icon" aria-hidden="true">
                                    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                        <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.13.96.36 1.9.7 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.91.34 1.85.57 2.81.7A2 2 0 0 1 22 16.92z" />
                                    </svg>
                                </span>
                                <span>{phone}</span>
                            </a>
                        {/if}
                        {#each socialLinks as link (link.url)}
                            <a
                                class="entry-profile__contact-row"
                                href={link.url}
                                target="_blank"
                                rel="noopener noreferrer"
                            >
                                <span class="entry-profile__contact-icon" aria-hidden="true">
                                    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                        <path d="M18 13v6a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                                        <polyline points="15 3 21 3 21 9" />
                                        <line x1="10" y1="14" x2="21" y2="3" />
                                    </svg>
                                </span>
                                <span>{link.label || link.url}</span>
                            </a>
                        {/each}
                        {#if showMapsPin && mapsHref}
                            <a
                                class="entry-profile__contact-row"
                                href={mapsHref}
                                target="_blank"
                                rel="nofollow noopener"
                            >
                                <span class="entry-profile__contact-icon" aria-hidden="true">
                                    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                        <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
                                        <circle cx="12" cy="10" r="3" />
                                    </svg>
                                </span>
                                <span class="entry-profile__contact-stack">
                                    <span>Útvonal</span>
                                    {#if address !== EMPTY_PLACEHOLDER}
                                        <span class="entry-profile__contact-sub">{address}</span>
                                    {/if}
                                    {#if locationName !== EMPTY_PLACEHOLDER}
                                        <span class="entry-profile__contact-sub">{locationName}</span>
                                    {/if}
                                </span>
                            </a>
                        {/if}
                        {#if showListingLanguages(entry)}
                            <div class="entry-profile__contact-row entry-profile__contact-row--static">
                                <span class="entry-profile__contact-icon" aria-hidden="true">
                                    <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                        <circle cx="12" cy="12" r="10" />
                                        <line x1="2" y1="12" x2="22" y2="12" />
                                        <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
                                    </svg>
                                </span>
                                <span>{languages}</span>
                            </div>
                        {/if}
                        {#if showClaim}
                            <div class="entry-profile__contact-row entry-profile__contact-row--static">
                                <button
                                    type="button"
                                    class="entry-profile__action"
                                    disabled={claimBusy}
                                    onclick={() => onClaim()}
                                >
                                    Sajátnak jelölöm
                                </button>
                            </div>
                        {/if}
                        {#if showClaimWaiting}
                            <div class="entry-profile__contact-row entry-profile__contact-row--static">
                                <button type="button" class="entry-profile__action" disabled>
                                    Átvételre vár
                                </button>
                            </div>
                        {/if}
                        {#if showJoin}
                            <div class="entry-profile__contact-row entry-profile__contact-row--static">
                                <button
                                    type="button"
                                    class="entry-profile__action"
                                    disabled={claimBusy}
                                    onclick={() => onJoin()}
                                >
                                    Tagság kérése
                                </button>
                            </div>
                        {/if}
                        {#if showJoinWaiting}
                            <div class="entry-profile__contact-row entry-profile__contact-row--static">
                                <button type="button" class="entry-profile__action" disabled>
                                    Tagságkérés elküldve
                                </button>
                            </div>
                        {/if}
                        {#if showOwnerEdit || suggestionState === "open" || suggestionState === "waiting"}
                            <div class="entry-profile__contact-row entry-profile__contact-row--static">
                                {@render suggestEdit()}
                            </div>
                        {/if}
                    </div>
                {/if}

                {#if showPlaceLinks}
                    <div class="index-tags-aside__section">
                        <h5 class="index-tags-aside__heading">Helyszín</h5>
                        <ul class="index-tag-cloud">
                            {#if countyHref && countyName !== EMPTY_PLACEHOLDER}
                                <li class="index-tag-cloud__li">
                                    <a class="btn btn-md" href={countyHref}>
                                        <span class="btn-label">{countyName}</span>
                                    </a>
                                </li>
                            {/if}
                            {#if locationHref && locationName !== EMPTY_PLACEHOLDER}
                                <li class="index-tag-cloud__li">
                                    <a class="btn btn-md" href={locationHref}>
                                        <span class="btn-label">{locationName}</span>
                                    </a>
                                </li>
                            {/if}
                        </ul>
                    </div>
                {/if}

                <div class="index-tags-aside__section">
                    <h5 class="index-tags-aside__heading">Címkék</h5>
                    {#if tagLabels.length}
                        <ul class="index-tag-cloud">
                            {#each tagLabels as label (label)}
                                <li class="index-tag-cloud__li">
                                    <span class="btn btn-md">
                                        <span class="btn-label">{label}</span>
                                    </span>
                                </li>
                            {/each}
                        </ul>
                    {:else}
                        <p class="index-tags-aside__empty">Nincs megjeleníthető címke.</p>
                    {/if}
                </div>
            </div>
        </aside>
    </div>
</article>
{/if}

<style>
    .entry-profile {
        color: var(--text-primary);
    }
    .entry-profile__hero {
        position: relative;
        margin-bottom: 1.5rem;
    }
    .entry-profile__favorite {
        position: absolute;
        top: 0;
        right: 0;
    }
    .entry-profile__title {
        margin: 0.35rem 0 0.4rem;
        line-height: 1.2;
    }
    .entry-profile__meta {
        margin: 0 0 0.45rem;
        color: var(--text-secondary);
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.35rem;
    }
    .entry-profile__dot {
        color: var(--text-faint);
    }
    .entry-profile__status-row {
        margin: 0 0 0.7rem;
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.4rem;
    }
    .entry-profile__claim {
        display: inline-flex;
        align-items: center;
        gap: 0.3rem;
        font-weight: 700;
        line-height: 1;
    }
    .entry-profile__claim--verified {
        color: var(--szekely-blue);
    }
    .entry-profile__claim--unverified {
        color: var(--text-faint);
    }
    .entry-profile__owned {
        font-weight: 700;
        color: var(--szekely-blue);
    }
    .entry-profile__today {
        font-weight: 600;
        color: var(--text-secondary);
    }
    .entry-profile__open {
        display: inline-flex;
        flex-wrap: wrap;
        align-items: baseline;
        gap: 0.35rem;
        font-weight: 700;
    }
    .entry-profile__open--open {
        color: var(--szekely-green);
    }
    .entry-profile__open--closed {
        color: var(--szekely-red);
    }
    .entry-profile__open--unknown {
        color: var(--text-muted);
        font-weight: 600;
    }
    .entry-profile__open-detail {
        font-weight: 500;
        color: var(--text-secondary);
    }
    .entry-profile__rating {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        flex-wrap: wrap;
    }
    .entry-profile__score {
        font-weight: 700;
    }
    .entry-profile__count {
        color: var(--text-muted);
    }
    .entry-profile__layout {
        display: grid;
        gap: 1.5rem;
    }
    .entry-profile__main {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 0;
    }
    .entry-profile__section + .entry-profile__section {
        margin-top: 1.5rem;
        padding-top: 1.5rem;
        border-top: 1px solid var(--border-color);
    }
    .entry-profile__section:first-child {
        margin: 0;
    }
    .entry-profile__section-title {
        margin: 0 0 0.75rem;
        font-weight: 700;
        color: var(--text-primary);
    }
    .entry-profile__suggest {
        appearance: none;
        border: none;
        background: none;
        padding: 0;
        margin: 0;
        display: inline-flex;
        align-items: center;
        gap: 0.35rem;
        font: inherit;
        font-weight: 600;
        color: var(--szekely-blue);
        cursor: pointer;
    }
    .entry-profile__suggest:hover:not(:disabled) {
        text-decoration: underline;
    }
    .entry-profile__suggest:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }
    .entry-profile__photos-skel {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    .entry-profile__skel-stage {
        width: 100%;
        aspect-ratio: 4 / 3;
        border-radius: 10px;
    }
    .entry-profile__skel-thumbs {
        display: flex;
        gap: 0.4rem;
    }
    .entry-profile__skel-thumb {
        width: 4.5rem;
        aspect-ratio: 4 / 3;
        border-radius: 8px;
        flex: 0 0 auto;
    }
    .entry-profile__about--clamp {
        display: -webkit-box;
        -webkit-line-clamp: 4;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }
    .entry-profile__action {
        appearance: none;
        border: none;
        background: none;
        padding: 0;
        margin: 0;
        font: inherit;
        font-weight: 600;
        color: var(--szekely-blue);
        cursor: pointer;
    }
    .entry-profile__action:hover:not(:disabled) {
        text-decoration: underline;
    }
    .entry-profile__action:disabled {
        opacity: 0.6;
        cursor: not-allowed;
    }
    .entry-profile__about {
        margin: 0;
        white-space: pre-wrap;
        line-height: 1.5;
    }
    .entry-profile__empty {
        margin: 0;
        color: var(--text-faint);
    }
    .entry-profile__place {
        display: flex;
        flex-wrap: wrap;
        align-items: flex-start;
        justify-content: space-between;
        gap: 0.75rem 1rem;
    }
    .entry-profile__address {
        margin: 0;
        display: flex;
        flex-direction: column;
        gap: 0.2rem;
        min-width: 0;
    }
    .entry-profile__link {
        color: var(--szekely-blue);
        font-weight: 600;
        text-decoration: none;
    }
    .entry-profile__link:hover {
        text-decoration: underline;
    }
    .entry-profile__hours-block {
        margin-top: 1.1rem;
    }
    .entry-profile__hours-title {
        margin: 0 0 0.35rem;
        font-weight: 700;
        color: var(--text-primary);
    }
    .entry-profile__hours {
        width: 100%;
        border-collapse: collapse;
    }
    .entry-profile__hours th,
    .entry-profile__hours td {
        padding: 0.35rem 0;
        border-top: 1px solid var(--border-color);
        text-align: left;
        font-weight: 500;
    }
    .entry-profile__hours th {
        width: 40%;
        color: var(--text-secondary);
        font-weight: 500;
    }
    .entry-profile__hours-today th,
    .entry-profile__hours-today td {
        color: var(--szekely-blue);
        font-weight: 700;
    }
    .entry-profile__services,
    .entry-profile__social {
        display: flex;
        flex-wrap: wrap;
        gap: 0.4rem;
        list-style: none;
        margin: 0;
        padding: 0;
    }
    .entry-profile__services li {
        padding: 0.2rem 0.55rem;
        border: 1px solid var(--border-color);
        border-radius: 999px;
        color: var(--text-secondary);
        background: var(--card-bg);
    }
    .entry-profile__social a {
        color: var(--szekely-blue);
    }
    .entry-profile__review-card {
        display: flex;
        gap: 0.85rem;
        align-items: flex-start;
        padding: 1rem;
        border: 1px solid var(--border-color);
        border-radius: 12px;
        background: var(--card-bg);
    }
    .entry-profile__review-avatar {
        flex: 0 0 2.5rem;
        width: 2.5rem;
        height: 2.5rem;
        border-radius: 999px;
        background: var(--entry-category-bg);
        color: var(--text-secondary);
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 700;
    }
    .entry-profile__review-body {
        min-width: 0;
        display: flex;
        flex-direction: column;
        gap: 0.35rem;
    }
    .entry-profile__review-text {
        margin: 0;
        font-style: italic;
        color: var(--text-secondary);
    }
    .entry-profile__aside {
        min-width: 0;
    }
    .entry-profile__aside :global(.sidebar-box) {
        display: flex;
        flex-direction: column;
        gap: 0;
    }
    .entry-profile__contact {
        margin: -0.25rem -0.35rem 0.85rem;
        overflow: hidden;
    }
    .entry-profile__contact:last-child {
        margin-bottom: 0;
    }
    .entry-profile__contact-row {
        display: flex;
        align-items: flex-start;
        gap: 0.75rem;
        padding: 0.75rem 0.5rem;
        border-top: 1px solid var(--border-color);
        color: inherit;
        text-decoration: none;
        border-radius: 0.5rem;
    }
    .entry-profile__contact-row:first-child {
        border-top: none;
    }
    a.entry-profile__contact-row:hover {
        background: color-mix(in srgb, var(--szekely-blue) 6%, var(--card-bg));
    }
    .entry-profile__aside :global(.index-tags-aside__section) {
        margin-top: 0.15rem;
    }
    .entry-profile__contact-icon {
        color: var(--szekely-green);
        display: inline-flex;
        margin-top: 0.1rem;
        flex-shrink: 0;
    }
    .entry-profile__contact-stack {
        display: flex;
        flex-direction: column;
        gap: 0.15rem;
        min-width: 0;
        font-weight: 600;
    }
    .entry-profile__contact-sub {
        font-weight: 400;
        color: var(--text-muted);
    }
    .entry-profile__skel-badge {
        width: 7.5rem;
        height: 1.35rem;
        margin-bottom: 0.55rem;
    }
    .entry-profile__skel-title {
        width: min(22rem, 90%);
        height: 1.7rem;
        margin: 0 0 0.65rem;
    }
    .entry-profile__skel-line {
        width: min(16rem, 70%);
        height: 0.85rem;
        margin: 0 0 0.45rem;
    }
    .entry-profile__skel-line--mid {
        width: min(12rem, 55%);
    }
    .entry-profile__skel-line--short {
        width: min(8rem, 40%);
    }
    .entry-profile__skel-heading {
        width: 8rem;
        height: 1.1rem;
        margin: 0 0 0.75rem;
    }
    .entry-profile__skel-hour {
        width: 100%;
        height: 0.85rem;
        margin: 0.35rem 0;
    }
    .entry-profile__skel-icon {
        width: 1.15rem;
        height: 1.15rem;
        flex-shrink: 0;
        border-radius: 4px;
    }
    .entry-profile__skel-contact {
        width: 70%;
        height: 0.85rem;
        margin: 0.15rem 0;
    }

    @media (min-width: 900px) {
        .entry-profile__layout {
            grid-template-columns: minmax(0, 1fr) 17.5rem;
            align-items: start;
            gap: 2rem;
        }
        .entry-profile__aside {
            position: sticky;
            top: 5.5rem;
        }
    }
</style>
