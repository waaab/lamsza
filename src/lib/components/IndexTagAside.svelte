<script>
    import ClaimStatusFilter from "$lib/components/ClaimStatusFilter.svelte";
    import { buildDirectoryTagCloud } from "$lib/directoryTagCloud.js";
    import { entryTypeChoices } from "$lib/entryType.js";

    /** @type {Array<{ type?: string, tags?: string[] }>} */
    export let entries = [];

    /** @type {string | null} */
    export let selectedTypeKey = null;
    /** @type {string | null} */
    export let selectedTagKey = null;
    /** @type {"claimed" | "unclaimed" | null} */
    export let claimFilter = null;
    export let claimedCount = 0;
    export let unclaimedCount = 0;
    export let loading = false;

    $: cloud = buildDirectoryTagCloud(entries);
    $: hasAny = cloud.tagItems.length > 0;
    $: typeButtons = entryTypeChoices().map((choice) => ({
        ...choice,
        count: cloud.typeItems.find((item) => item.key === choice.key)?.count ?? 0,
    }));

    function toggleTag(/** @type {string} */ k) {
        selectedTagKey = selectedTagKey === k ? null : k;
    }

    function toggleType(/** @type {string} */ k) {
        selectedTypeKey = selectedTypeKey === k ? null : k;
    }
</script>

<div class="index-tags-aside">
    <ClaimStatusFilter
        bind:value={claimFilter}
        {claimedCount}
        {unclaimedCount}
        {loading}
    />
    <div class="index-tags-aside__section">
        <h3 class="aside_heading">Típus</h3>
        {#if loading}
            <div class="index-tags-aside-skeleton__row" aria-hidden="true">
                {#each Array(3) as _, i (i)}
                    <span class="skeleton index-tags-aside-skeleton__chip"></span>
                {/each}
            </div>
        {:else}
        <ul class="index-tag-cloud">
            {#each typeButtons as item (item.key)}
                <li class="index-tag-cloud__li">
                    <button
                        type="button"
                        class="btn btn-md"
                        class:active={selectedTypeKey === item.key}
                        aria-pressed={selectedTypeKey === item.key}
                        on:click={() => toggleType(item.key)}
                    >
                        <span class="btn-label">{item.label}</span>
                        <span class="btn-label-count">{item.count}</span>
                    </button>
                </li>
            {/each}
        </ul>
        {/if}
    </div>
    <div class="index-tags-aside__section">
        <h3 class="aside_heading">Címkék</h3>
        {#if loading}
            <div class="index-tags-aside-skeleton__row" aria-hidden="true">
                {#each Array(6) as _, i (i)}
                    <span class="skeleton index-tags-aside-skeleton__chip"></span>
                {/each}
            </div>
        {:else if !hasAny}
            <p class="index-tags-aside__empty">Nincs megjeleníthető címke.</p>
        {:else}
                <ul class="index-tag-cloud">
                    {#each cloud.tagItems as item (item.key)}
                        <li class="index-tag-cloud__li">
                        <button
                            type="button"
                            class="btn btn-md"
                            class:active={selectedTagKey ===
                                item.key}
                            on:click={() => toggleTag(item.key)}
                        >
                            <span class="btn-label">{item.label}</span>
                            <span class="btn-label-count">{item.count}</span>
                        </button>
                        </li>
                    {/each}
                </ul>
        {/if}
    </div>
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
        min-width: 4.5rem;
    }
</style>