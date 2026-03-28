<script>
    export let label = "";
    export let parentLabel = "Index";
    export let parentUrl = "/index";
    export let countySlug = "";
    export let countyName = "";
    export let settlementSlug = "";
    export let settlementName = "";
    export let settlementType = ""; // e.g., 'város', 'falu'
</script>

<div class="breadcrumbs">
    <a href="/">Főoldal</a>
    {#if parentLabel && parentUrl}
        &rsaquo;
        <a href={parentUrl}>{parentLabel}</a>
    {/if}
    &rsaquo;
    {#if countyName && countySlug}
        <a href="/{countySlug}-megye">{countyName}</a>
        &rsaquo;
    {/if}
    {#if settlementSlug && settlementName}
        <a href="/{countySlug}-megye/{settlementSlug}">{settlementName}</a>
        &rsaquo;
        <span class="active">{label}</span>
    {:else if settlementType}
        <span class="active">{settlementType}: {label}</span>
    {:else}
        <span class="active">{label}</span>
    {/if}
</div>
