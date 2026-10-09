<script>
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import LandscapeBackdrop from "$lib/components/LandscapeBackdrop.svelte";

    /** Főcím (admin «Oldalak» táblából, vagy betöltésig üres). */
    export let title = "";
    /** Bevezető a cím alatt (mindig megjelenik; betöltéskor: „…”). */
    export let greeting = "";
    /** Ha igaz, a `pages` meta (cím / bevezető) még töltődik - a bevezető helyén „…”. */
    export let loading = false;

    export let showBreadcrumbs = true;
    export let showTitle = true;
    export let showGreeting = true;
    export let breadcrumbLabel = "";
    /** Alap: Index; üres string = nincs szülő morzsa (pl. /index). */
    export let breadcrumbParentLabel = "Index";
    export let breadcrumbParentUrl = "/index";
    export let breadcrumbExtraLabel = "";
    export let breadcrumbExtraUrl = "";
    export let documentTitleSuffix = " - Lámsza";
    /**
     * The Székelyföld pages' header (UI_BASELINE "szf-pages"): a red sun and
     * three mountain ridges behind the breadcrumb, title and line. Without
     * it the wrapper is `display: contents`, so other pages lay out as before.
     */
    export let landscape = false;

    $: displayTitle = title;
    $: crumb = breadcrumbLabel.trim() || displayTitle;
    $: displayGreeting = loading ? "…" : (greeting != null ? String(greeting) : "");
</script>

<svelte:head>
    <title>{displayTitle}{documentTitleSuffix}</title>
</svelte:head>

<div class="page-header" class:page-landscape={landscape}>
{#if landscape}
    <LandscapeBackdrop />
{/if}
{#if showBreadcrumbs}
    <Breadcrumbs
        label={crumb}
        parentLabel={breadcrumbParentLabel}
        parentUrl={breadcrumbParentUrl}
        extraLabel={breadcrumbExtraLabel}
        extraUrl={breadcrumbExtraUrl}
    />
{/if}

{#if showTitle}
    {#if $$slots.title}
        <slot name="title" />
    {:else}
        <h1 class="page-title">{displayTitle || (loading ? "…" : "")}</h1>
    {/if}
{/if}

{#if showGreeting && displayGreeting}
    <h2 class="greeting">{displayGreeting}</h2>
{/if}
</div>
