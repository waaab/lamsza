<script>
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";

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
    <span class="page-landscape__sun" aria-hidden="true"></span>
    <svg class="page-landscape__ridge" viewBox="0 0 1200 110" preserveAspectRatio="none" aria-hidden="true" focusable="false">
        <path class="page-landscape__far" d="M0 110V60L120 38L240 64L380 24L520 58L650 18L800 54L950 32L1080 60L1200 40V110Z" />
        <path class="page-landscape__mid" d="M0 110V78L150 56L300 82L470 46L620 80L780 52L930 82L1100 60L1200 74V110Z" />
        <path class="page-landscape__near" d="M0 110V94L200 80L400 98L600 82L800 100L1000 84L1200 96V110Z" />
    </svg>
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
