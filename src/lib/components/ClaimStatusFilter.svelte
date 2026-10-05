<script>
    /** @type {{ value?: "claimed" | "unclaimed" | null, claimedCount?: number, unclaimedCount?: number }} */
    let { value = $bindable(null), claimedCount = 0, unclaimedCount = 0, loading = false } = $props();

    /** @param {"claimed" | "unclaimed"} next */
    function toggle(next) {
        value = value === next ? null : next;
    }
</script>

<div class="index-tags-aside__section">
    <h5 class="index-tags-aside__heading">Állapot</h5>
    {#if loading}
        <div class="index-tags-aside-skeleton__row" aria-hidden="true">
            <span class="skeleton index-tags-aside-skeleton__chip"></span>
            <span class="skeleton index-tags-aside-skeleton__chip"></span>
        </div>
    {:else}
    <ul class="index-tag-cloud">
        <li class="index-tag-cloud__li">
            <button
                type="button"
                class="btn btn-md"
                class:active={value === "claimed"}
                aria-pressed={value === "claimed"}
                onclick={() => toggle("claimed")}
            >
                <span class="btn-label">Átvéve</span>
                <span class="btn-label-count">{claimedCount}</span>
            </button>
        </li>
        <li class="index-tag-cloud__li">
            <button
                type="button"
                class="btn btn-md"
                class:active={value === "unclaimed"}
                aria-pressed={value === "unclaimed"}
                onclick={() => toggle("unclaimed")}
            >
                <span class="btn-label">Gazdátlan</span>
                <span class="btn-label-count">{unclaimedCount}</span>
            </button>
        </li>
    </ul>
    {/if}
</div>

<style>
    .index-tags-aside-skeleton__row {
        display: flex;
        flex-wrap: wrap;
        gap: 0.4rem;
    }

    .index-tags-aside-skeleton__chip {
        display: inline-block;
        height: 1.85rem;
        border-radius: 999px;
        min-width: 5.5rem;
    }
</style>
