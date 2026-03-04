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

<div
    class="breadcrumbs"
    style="color: var(--text-faint); font-size: 0.9rem; margin-bottom: 1rem;"
>
    <a href="/" style="color: var(--primary-color); text-decoration: none;"
        >Főoldal</a
    >
    &rsaquo;
    <a
        href={parentUrl}
        style="color: var(--primary-color); text-decoration: none;"
        >{parentLabel}</a
    >
    &rsaquo;
    {#if countyName && countySlug}
        <a
            href="/{countySlug}-megye"
            style="color: var(--primary-color); text-decoration: none;"
            >{countyName}</a
        >
        &rsaquo;
    {/if}
    {#if settlementSlug && settlementName}
        <a
            href="/{countySlug}-megye/{settlementSlug}"
            style="color: var(--primary-color); text-decoration: none;"
            >{settlementName}</a
        >
        &rsaquo;
        <span class="active" style="color: var(--text-color); font-weight: 500;"
            >{label}</span
        >
    {:else if settlementType}
        <span class="active" style="color: var(--text-color); font-weight: 500;"
            >{settlementType}: {label}</span
        >
    {:else}
        <span class="active" style="color: var(--text-color); font-weight: 500;"
            >{label}</span
        >
    {/if}
</div>
