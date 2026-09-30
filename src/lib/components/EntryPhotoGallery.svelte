<script>
    import { fade } from "svelte/transition";
    import { photoCredit } from "$lib/entryPhotos.js";

    /**
     * Native Svelte slideshow (no carousel package): stacked imgs + CSS opacity,
     * `{#key}` caption fade, keyboard on the gallery itself.
     * Clicking a photo opens a higher-resolution viewer with prev/next and caption.
     * @type {{
     *   slides?: Array<{
     *     src: string,
     *     fullSrc?: string,
     *     fallback?: string,
     *     alt: string,
     *     title: string,
     *     description: string,
     *     copyright?: string,
     *     uploader?: string,
     *     width: number,
     *     height: number,
     *     loading?: string,
     *     fetchpriority?: string,
     *   }>,
     *   label?: string,
     * }}
     */
    let { slides = [], label = "Fotók" } = $props();

    let index = $state(0);
    let root = $state(/** @type {HTMLElement | null} */ (null));
    let lightbox = $state(/** @type {HTMLDialogElement | null} */ (null));
    let viewerOpen = $state(false);
    let safeIndex = $derived(
        slides.length ? Math.min(Math.max(0, index), slides.length - 1) : 0,
    );
    let current = $derived(slides[safeIndex] ?? null);
    let viewerSrc = $derived(current?.fullSrc || current?.src || "");
    let ratio = $derived.by(() => {
        const w = current?.width || slides[0]?.width || 4;
        const h = current?.height || slides[0]?.height || 3;
        return `${w} / ${h}`;
    });

    function go(next) {
        if (!slides.length) return;
        index = (next + slides.length) % slides.length;
    }

    function onKey(event) {
        if (event.key === "ArrowRight") {
            event.preventDefault();
            go(safeIndex + 1);
        } else if (event.key === "ArrowLeft") {
            event.preventDefault();
            go(safeIndex - 1);
        } else if (event.key === "Home") {
            event.preventDefault();
            index = 0;
        } else if (event.key === "End") {
            event.preventDefault();
            index = slides.length - 1;
        }
    }

    function onWindowKey(event) {
        if (viewerOpen) {
            if (event.key === "Escape") {
                event.preventDefault();
                closeViewer();
                return;
            }
            onKey(event);
            return;
        }
        if (!root?.contains(event.target)) return;
        onKey(event);
    }

    function handleError(event, fallback) {
        const img = event.currentTarget;
        if (fallback && img.src !== fallback && img.dataset.fallback !== "1") {
            img.dataset.fallback = "1";
            img.src = fallback;
        }
    }

    function visibleBody(slide) {
        if (!slide) return "";
        if (slide.description) return slide.description;
        if (slide.alt && slide.alt !== slide.title) return slide.alt;
        return slide.alt || "";
    }

    function imgLoading(i) {
        if (i === 0 || Math.abs(i - safeIndex) <= 1) return "eager";
        return "lazy";
    }

    function openViewer(i) {
        if (!slides.length) return;
        index = i;
        viewerOpen = true;
    }

    function closeViewer() {
        if (lightbox?.open) {
            lightbox.close();
            return;
        }
        viewerOpen = false;
    }

    function onDialogClose() {
        viewerOpen = false;
    }

    $effect(() => {
        if (!viewerOpen) return;
        const prev = document.body.style.overflow;
        document.body.style.overflow = "hidden";
        queueMicrotask(() => {
            if (lightbox && !lightbox.open) lightbox.showModal();
        });
        return () => {
            document.body.style.overflow = prev;
        };
    });
</script>

<svelte:window onkeydown={onWindowKey} />

{#snippet captionDesc(slide)}
    {@const credit = photoCredit(slide)}
    <p class="entry-gallery__caption-desc">
        {#if visibleBody(slide)}
            <span class="entry-gallery__caption-text">{visibleBody(slide)}</span>
        {/if}
        <span class="entry-gallery__credit">© {credit.copyright}</span>
        <span class="entry-gallery__credit">{credit.uploaderLine}</span>
    </p>
{/snippet}

{#if !slides.length}
    <div class="entry-gallery entry-gallery--empty">
        <div class="entry-gallery__stage entry-gallery__stage--empty">
            <span>Képek hamarosan</span>
        </div>
    </div>
{:else}
    <section
        class="entry-gallery"
        bind:this={root}
        aria-roledescription="slideshow"
        aria-label={label}
    >
        <figure class="entry-gallery__figure">
            <div class="entry-gallery__stage-wrap">
                <div class="entry-gallery__stage" style:aspect-ratio={ratio}>
                    {#each slides as slide, i (`${slide.src}-${i}`)}
                        <img
                            class={["entry-gallery__img", i === safeIndex && "is-active"]}
                            src={slide.src}
                            alt={slide.alt || ""}
                            title={slide.title || slide.alt || ""}
                            width={slide.width}
                            height={slide.height}
                            loading={imgLoading(i)}
                            decoding="async"
                            fetchpriority={i === 0 ? "high" : "low"}
                            onerror={(event) => handleError(event, slide.fallback)}
                        />
                    {/each}
                    <button
                        type="button"
                        class="entry-gallery__open"
                        aria-label="Fotó megnyitása nagyobb méretben"
                        onclick={() => openViewer(safeIndex)}
                    ></button>
                </div>
                {#if slides.length > 1}
                    <button
                        type="button"
                        class="entry-gallery__nav entry-gallery__nav--prev"
                        aria-label="Előző fotó"
                        onclick={() => go(safeIndex - 1)}
                    >
                        ‹
                    </button>
                    <button
                        type="button"
                        class="entry-gallery__nav entry-gallery__nav--next"
                        aria-label="Következő fotó"
                        onclick={() => go(safeIndex + 1)}
                    >
                        ›
                    </button>
                    <p class="entry-gallery__count" aria-live="polite">
                        {safeIndex + 1} / {slides.length}
                    </p>
                {/if}
            </div>
            {#key safeIndex}
                <figcaption class="entry-gallery__caption" in:fade={{ duration: 160 }}>
                    {#if current?.title}
                        <strong class="entry-gallery__caption-title">{current.title}</strong>
                    {/if}
                    {@render captionDesc(current)}
                </figcaption>
            {/key}
        </figure>

        {#if slides.length > 1}
            <div class="entry-gallery__thumbs" role="tablist" aria-label="Bélyegképek">
                {#each slides as slide, i (`thumb-${slide.src}-${i}`)}
                    <button
                        type="button"
                        class={["entry-gallery__thumb", i === safeIndex && "is-active"]}
                        role="tab"
                        aria-selected={i === safeIndex}
                        aria-label={slide.title || slide.alt || `Fotó ${i + 1}`}
                        onclick={() => (index = i)}
                    >
                        <img
                            src={slide.src}
                            alt={slide.alt || ""}
                            title={slide.title || slide.alt || ""}
                            width={slide.width}
                            height={slide.height}
                            loading={i === 0 ? "eager" : "lazy"}
                            decoding="async"
                            onerror={(event) => handleError(event, slide.fallback)}
                        />
                    </button>
                {/each}
            </div>
        {/if}
    </section>

    {#if viewerOpen}
    <dialog
        bind:this={lightbox}
        class="entry-gallery__lightbox"
        aria-label={`${label} — nagyobb méret`}
        onclose={onDialogClose}
    >
        <button
            type="button"
            class="entry-gallery__lightbox-backdrop"
            tabindex="-1"
            aria-label="Bezárás"
            onclick={closeViewer}
        ></button>
        <button
            type="button"
            class="entry-gallery__lightbox-close"
            aria-label="Bezárás"
            autofocus
            onclick={closeViewer}
        >
            ×
        </button>
        {#if slides.length > 1}
            <button
                type="button"
                class="entry-gallery__nav entry-gallery__nav--prev"
                aria-label="Előző fotó"
                onclick={() => go(safeIndex - 1)}
            >
                ‹
            </button>
            <button
                type="button"
                class="entry-gallery__nav entry-gallery__nav--next"
                aria-label="Következő fotó"
                onclick={() => go(safeIndex + 1)}
            >
                ›
            </button>
        {/if}
        {#if current}
            <figure class="entry-gallery__lightbox-slide">
                {#key safeIndex}
                    <img
                        class="entry-gallery__lightbox-img"
                        src={viewerSrc}
                        alt={current.alt || ""}
                        width={current.width}
                        height={current.height}
                        decoding="async"
                        in:fade={{ duration: 180 }}
                        onerror={(event) =>
                            handleError(event, current.src || current.fallback)}
                    />
                    <figcaption class="entry-gallery__lightbox-caption" in:fade={{ duration: 160 }}>
                        <div>
                            {#if current.title}
                                <strong class="entry-gallery__caption-title">{current.title}</strong>
                            {/if}
                            {@render captionDesc(current)}
                        </div>
                        {#if slides.length > 1}
                            <p class="entry-gallery__count" aria-live="polite">
                                {safeIndex + 1} / {slides.length}
                            </p>
                        {/if}
                    </figcaption>
                {/key}
            </figure>
        {/if}
    </dialog>
    {/if}
{/if}

<style>
    .entry-gallery {
        display: flex;
        flex-direction: column;
        gap: 0.65rem;
        margin: 0;
    }
    .entry-gallery:focus-visible {
        box-shadow: 0 0 0 2px var(--szekely-blue, #1d4ed8);
        border-radius: 12px;
    }
    .entry-gallery__figure {
        margin: 0;
        display: flex;
        flex-direction: column;
        gap: 0.45rem;
    }
    .entry-gallery__stage-wrap {
        position: relative;
    }
    .entry-gallery__stage {
        position: relative;
        width: 100%;
        aspect-ratio: 4 / 3;
        border-radius: 10px;
        overflow: hidden;
        background: var(--card-bg);
        border: 1px solid var(--border-color);
        contain: layout paint;
    }
    .entry-gallery__stage--empty {
        display: flex;
        align-items: center;
        justify-content: center;
        color: var(--text-muted);
        border-style: dashed;
        min-height: 12rem;
    }
    .entry-gallery__img {
        position: absolute;
        inset: 0;
        width: 100%;
        height: 100%;
        object-fit: cover;
        opacity: 0;
        visibility: hidden;
        pointer-events: none;
        transition: opacity 180ms ease;
    }
    .entry-gallery__img.is-active {
        opacity: 1;
        visibility: visible;
        pointer-events: auto;
    }
    .entry-gallery__open {
        position: absolute;
        inset: 0;
        z-index: 1;
        border: none;
        background: transparent;
        cursor: zoom-in;
    }
    .entry-gallery__nav {
        position: absolute;
        top: 50%;
        transform: translateY(-50%);
        z-index: 2;
        width: 2.25rem;
        height: 2.25rem;
        border: none;
        border-radius: 999px;
        background: color-mix(in srgb, #0f172a 55%, transparent);
        color: #fff;
        font-size: 1.4rem;
        line-height: 1;
        cursor: pointer;
    }
    .entry-gallery__nav--prev {
        left: 0.5rem;
    }
    .entry-gallery__nav--next {
        right: 0.5rem;
    }
    .entry-gallery__nav:hover {
        background: color-mix(in srgb, #0f172a 75%, transparent);
    }
    .entry-gallery__count {
        position: absolute;
        right: 0.6rem;
        bottom: 0.55rem;
        margin: 0;
        padding: 0.15rem 0.45rem;
        border-radius: 999px;
        background: color-mix(in srgb, #0f172a 55%, transparent);
        color: #fff;
        font-size: 0.8rem;
    }
    .entry-gallery__caption {
        margin: 0;
        min-height: 2.4rem;
    }
    .entry-gallery__caption-title {
        display: block;
        font-weight: 700;
        color: var(--text-primary);
    }
    .entry-gallery__caption-desc {
        margin: 0.2rem 0 0;
        color: var(--text-secondary);
        line-height: 1.4;
    }
    .entry-gallery__caption-text,
    .entry-gallery__credit {
        display: block;
    }
    .entry-gallery__caption-text + .entry-gallery__credit,
    .entry-gallery__credit + .entry-gallery__credit {
        margin-top: 0.25rem;
    }
    .entry-gallery__thumbs {
        display: flex;
        gap: 0.4rem;
        overflow-x: auto;
        padding-bottom: 0.15rem;
    }
    .entry-gallery__thumb {
        flex: 0 0 auto;
        width: 4.5rem;
        aspect-ratio: 4 / 3;
        padding: 0;
        border: 2px solid transparent;
        border-radius: 8px;
        overflow: hidden;
        background: var(--card-bg);
        cursor: pointer;
    }
    .entry-gallery__thumb.is-active {
        border-color: var(--szekely-blue, #1d4ed8);
    }
    .entry-gallery__thumb img {
        display: block;
        width: 100%;
        height: 100%;
        object-fit: cover;
        pointer-events: none;
    }
    .entry-gallery__lightbox {
        position: fixed;
        inset: 0;
        width: 100%;
        height: 100%;
        max-width: none;
        max-height: none;
        margin: 0;
        padding: 3.25rem 4.25rem 1.5rem;
        border: none;
        box-sizing: border-box;
        background: rgb(4 6 10 / 0.94);
        color: #fff;
        display: flex;
        align-items: center;
        justify-content: center;
        overflow: auto;
        animation: entry-gallery-in 180ms ease;
    }
    .entry-gallery__lightbox::backdrop {
        background: rgb(4 6 10 / 0.94);
    }
    .entry-gallery__lightbox-backdrop {
        position: absolute;
        inset: 0;
        z-index: 0;
        border: none;
        background: transparent;
        cursor: zoom-out;
    }
    .entry-gallery__lightbox-slide {
        position: relative;
        z-index: 1;
        margin: 0;
        display: flex;
        flex-direction: column;
        width: fit-content;
        max-width: min(92vw, 76rem);
    }
    .entry-gallery__lightbox-img {
        display: block;
        width: auto;
        height: auto;
        max-width: min(92vw, 76rem);
        max-height: calc(100vh - 10rem);
        margin: 0 auto;
        object-fit: contain;
    }
    .entry-gallery__lightbox-close {
        position: fixed;
        top: 0.85rem;
        right: 0.85rem;
        z-index: 3;
        width: 2.75rem;
        height: 2.75rem;
        border: none;
        border-radius: 999px;
        background: transparent;
        color: #fff;
        font-size: 2rem;
        line-height: 1;
        cursor: pointer;
        text-shadow: 0 1px 3px rgb(0 0 0 / 0.65);
    }
    .entry-gallery__lightbox-close:hover {
        background: rgb(255 255 255 / 0.14);
    }
    .entry-gallery__lightbox .entry-gallery__nav {
        position: fixed;
        z-index: 3;
        width: 2.75rem;
        height: 2.75rem;
        font-size: 1.8rem;
        background: rgb(15 23 42 / 0.55);
    }
    .entry-gallery__lightbox .entry-gallery__nav--prev {
        left: 0.85rem;
    }
    .entry-gallery__lightbox .entry-gallery__nav--next {
        right: 0.85rem;
    }
    .entry-gallery__lightbox-caption {
        display: flex;
        align-items: flex-start;
        justify-content: space-between;
        gap: 1rem;
        width: 100%;
        margin-top: 0.85rem;
    }
    .entry-gallery__lightbox-caption .entry-gallery__caption-title,
    .entry-gallery__lightbox-caption .entry-gallery__caption-desc {
        color: #fff;
    }
    .entry-gallery__lightbox-caption .entry-gallery__caption-desc {
        opacity: 0.88;
    }
    .entry-gallery__lightbox-caption .entry-gallery__count {
        position: static;
        flex: 0 0 auto;
        margin: 0.1rem 0 0;
        padding: 0;
        background: none;
        color: #fff;
        font-size: 0.9rem;
    }
    @keyframes entry-gallery-in {
        from {
            opacity: 0;
        }
        to {
            opacity: 1;
        }
    }
    @media (max-width: 640px) {
        .entry-gallery__lightbox {
            padding: 3.25rem 0.75rem 1rem;
        }
        .entry-gallery__lightbox-img {
            max-width: 94vw;
            max-height: calc(100vh - 11rem);
        }
        .entry-gallery__lightbox .entry-gallery__nav--prev {
            left: 0.35rem;
        }
        .entry-gallery__lightbox .entry-gallery__nav--next {
            right: 0.35rem;
        }
    }
</style>
