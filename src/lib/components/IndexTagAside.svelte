<script>
    import ClaimStatusFilter from "$lib/components/ClaimStatusFilter.svelte";
    import { buildDirectoryTagCloud } from "$lib/directoryTagCloud.js";

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

    $: cloud = buildDirectoryTagCloud(entries);
    $: hasAny =
        cloud.typeItems.length > 0 || cloud.tagItems.length > 0;

    function toggleType(/** @type {string} */ k) {
        selectedTypeKey = selectedTypeKey === k ? null : k;
    }

    function toggleTag(/** @type {string} */ k) {
        selectedTagKey = selectedTagKey === k ? null : k;
    }
</script>

<div class="index-tags-aside">
    <ClaimStatusFilter
        bind:value={claimFilter}
        {claimedCount}
        {unclaimedCount}
    />
    {#if !hasAny}
        <p class="index-tags-aside__empty">Nincs megjeleníthető címke.</p>
    {:else}
        {#if cloud.typeItems.length > 0}
            <div class="index-tags-aside__section">
                <h5 class="index-tags-aside__heading">Bejegyzés típus</h5>
                <ul class="index-tag-cloud">
                    {#each cloud.typeItems as item (item.key)}
                        <li class="index-tag-cloud__li">
                        <button
                            type="button"
                            class="btn btn-md"
                            class:active={selectedTypeKey ===
                                item.key}
                            on:click={() => toggleType(item.key)}
                        >
                            <span class="btn-label">{item.label}</span>
                            <span class="btn-label-count">{item.count}</span>
                        </button>
                        </li>
                    {/each}
                </ul>
            </div>
        {/if}

        {#if cloud.tagItems.length > 0}
            <div class="index-tags-aside__section">
                <h5 class="index-tags-aside__heading">Kulcsszavak</h5>
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
            </div>
        {/if}
    {/if}
</div>