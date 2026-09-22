<script>
    import { auth } from "$lib/stores/auth.js";
    import { openLogin } from "$lib/openLogin.js";
    import EntryStars from "$lib/components/EntryStars.svelte";

    /** @type {{ entry: Record<string, any> }} */
    let { entry: entryProp } = $props();

    let score = $state(0);
    let text = $state("");
    let error = $state("");
    let busy = $state(false);

    $effect(() => {
        if (entryProp.my_review) {
            score = entryProp.my_review.score || 0;
            text = entryProp.my_review.text || "";
        }
    });

    let loggedIn = $derived($auth.loggedIn);
    let rating = $derived.by(() => {
        const n = Number(entryProp?.rating);
        return Number.isFinite(n) ? Math.min(5, Math.max(0, n)) : 0;
    });
    let reviewCount = $derived.by(() => {
        const n = Number(entryProp?.review_count);
        return Number.isFinite(n) ? Math.max(0, Math.floor(n)) : 0;
    });
    let ratingLabel = $derived(
        rating.toLocaleString("hu-HU", {
            minimumFractionDigits: 1,
            maximumFractionDigits: 1,
        }),
    );
    let reviews = $derived(Array.isArray(entryProp?.reviews) ? entryProp.reviews : []);
    let hasMyReview = $derived(Boolean(entryProp?.my_review));

    async function submitReview() {
        if (busy || score < 1 || score > 5) return;
        error = "";
        busy = true;
        try {
            const response = await fetch("/api/entry/reviews", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ slug: entryProp.slug, score, text }),
            });
            if (!response.ok) {
                throw new Error("Failed to save review");
            }
            const data = await response.json();
            entryProp.rating = data.rating;
            entryProp.review_count = data.review_count;
            entryProp.reviews = data.reviews || [];
            entryProp.my_review = data.my_review;
        } catch {
            error = "A mentés nem sikerült";
        } finally {
            busy = false;
        }
    }

    async function deleteReview() {
        if (busy || !hasMyReview) return;
        error = "";
        busy = true;
        try {
            const response = await fetch(`/api/entry/reviews?slug=${encodeURIComponent(entryProp.slug)}`, {
                method: "DELETE",
            });
            if (!response.ok) {
                throw new Error("Failed to delete review");
            }
            const data = await response.json();
            entryProp.rating = data.rating;
            entryProp.review_count = data.review_count;
            entryProp.reviews = data.reviews || [];
            entryProp.my_review = null;
            score = 0;
            text = "";
        } catch {
            error = "A mentés nem sikerült";
        } finally {
            busy = false;
        }
    }

    function setScore(newScore) {
        if (!busy) {
            score = newScore;
        }
    }
</script>

<h2 id="entry-reviews-title" class="entry-reviews__title">Értékelések</h2>
<div class="entry-reviews__header">
    <EntryStars {rating} size={20} />
    <span class="entry-reviews__rating">{ratingLabel}</span>
    <span class="entry-reviews__count">({reviewCount} értékelés)</span>
</div>

{#if reviews.length === 0}
    <p class="entry-reviews__empty">Még nincs értékelés.</p>
{:else}
    <ul class="entry-reviews__list">
        {#each reviews as review (review.id || review.author_name)}
            <li class="entry-reviews__item">
                <div class="entry-reviews__avatar" aria-hidden="true">
                    {#if review.author_photo}
                        <img src={review.author_photo} alt="" />
                    {:else}
                        {(review.author_name || "?").slice(0, 1).toUpperCase()}
                    {/if}
                </div>
                <div class="entry-reviews__body">
                    <div class="entry-reviews__meta">
                        <span class="entry-reviews__author">{review.author_name || "Névtelen"}</span>
                        <EntryStars rating={review.score} size={14} />
                    </div>
                    {#if review.text}
                        <p class="entry-reviews__text">{review.text}</p>
                    {/if}
                </div>
            </li>
        {/each}
    </ul>
{/if}

{#if loggedIn}
    <div class="entry-reviews__form">
        <h3 class="entry-reviews__form-title">Az ön értékelése</h3>
        <div class="entry-reviews__stars">
            {#each [1, 2, 3, 4, 5] as star (star)}
                <button
                    type="button"
                    class="entry-reviews__star-btn"
                    class:entry-reviews__star-btn--active={score >= star}
                    onclick={() => setScore(star)}
                    disabled={busy}
                    aria-label="{star} csillag"
                >
                    <svg viewBox="0 0 24 24" width="24" height="24" fill={score >= star ? "currentColor" : "none"} stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                    </svg>
                </button>
            {/each}
        </div>
        <textarea
            class="entry-reviews__textarea"
            placeholder="Ossza meg véleményét..."
            bind:value={text}
            disabled={busy}
            rows="4"
        ></textarea>
        {#if error}
            <p class="entry-reviews__error">{error}</p>
        {/if}
        <div class="entry-reviews__actions">
            <button
                type="button"
                class="btn btn-primary"
                onclick={submitReview}
                disabled={busy || score < 1}
            >
                Mentés
            </button>
            {#if hasMyReview}
                <button
                    type="button"
                    class="btn btn-secondary"
                    onclick={deleteReview}
                    disabled={busy}
                >
                    Törlés
                </button>
            {/if}
        </div>
    </div>
{:else}
    <div class="entry-reviews__login">
        <button type="button" class="btn btn-primary" onclick={openLogin}>
            Belépés
        </button>
    </div>
{/if}

<style>
    .entry-reviews__title {
        margin: 0 0 0.75rem;
        font-weight: 700;
        color: var(--text-primary);
    }
    .entry-reviews__header {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        margin-bottom: 1rem;
    }
    .entry-reviews__rating {
        font-weight: 700;
        color: var(--text-primary);
    }
    .entry-reviews__count {
        color: var(--text-muted);
    }
    .entry-reviews__empty {
        margin: 0;
        padding: 1rem;
        color: var(--text-faint);
        background: var(--card-bg);
        border: 1px solid var(--border-color);
        border-radius: 8px;
    }
    .entry-reviews__list {
        list-style: none;
        margin: 0 0 1.5rem;
        padding: 0;
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }
    .entry-reviews__item {
        display: flex;
        gap: 0.75rem;
        padding: 1rem;
        background: var(--card-bg);
        border: 1px solid var(--border-color);
        border-radius: 8px;
    }
    .entry-reviews__avatar {
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
        overflow: hidden;
    }
    .entry-reviews__avatar img {
        width: 100%;
        height: 100%;
        object-fit: cover;
    }
    .entry-reviews__body {
        min-width: 0;
        flex: 1;
    }
    .entry-reviews__meta {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        margin-bottom: 0.35rem;
        flex-wrap: wrap;
    }
    .entry-reviews__author {
        font-weight: 600;
        color: var(--text-primary);
    }
    .entry-reviews__text {
        margin: 0;
        color: var(--text-secondary);
        line-height: 1.5;
        white-space: pre-wrap;
    }
    .entry-reviews__form {
        padding: 1.25rem;
        background: var(--card-bg);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        margin-top: 1.5rem;
    }
    .entry-reviews__form-title {
        margin: 0 0 0.75rem;
        font-size: 1rem;
        font-weight: 600;
        color: var(--text-primary);
    }
    .entry-reviews__stars {
        display: flex;
        gap: 0.25rem;
        margin-bottom: 0.75rem;
    }
    .entry-reviews__star-btn {
        appearance: none;
        border: none;
        background: none;
        padding: 0;
        cursor: pointer;
        color: var(--text-faint);
        transition: color 0.15s;
    }
    .entry-reviews__star-btn:disabled {
        cursor: not-allowed;
        opacity: 0.5;
    }
    .entry-reviews__star-btn:hover:not(:disabled) {
        color: var(--szekely-yellow);
    }
    .entry-reviews__star-btn--active {
        color: var(--szekely-yellow);
    }
    .entry-reviews__textarea {
        width: 100%;
        padding: 0.65rem;
        border: 1px solid var(--border-color);
        border-radius: 6px;
        font: inherit;
        resize: vertical;
        background: var(--bg);
        color: var(--text-primary);
    }
    .entry-reviews__textarea:disabled {
        opacity: 0.5;
        cursor: not-allowed;
    }
    .entry-reviews__error {
        margin: 0.5rem 0 0;
        color: var(--szekely-red);
        font-weight: 600;
    }
    .entry-reviews__actions {
        display: flex;
        gap: 0.5rem;
        margin-top: 0.75rem;
        flex-wrap: wrap;
    }
    .entry-reviews__login {
        margin-top: 1rem;
    }
</style>
