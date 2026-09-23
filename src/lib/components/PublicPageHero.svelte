<script>
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";

    /** Főcím (admin «Oldalak» táblából, vagy betöltésig üres). */
    export let title = "";
    /** Bevezető a cím alatt (mindig megjelenik; betöltéskor: „…”). */
    export let greeting = "";
    /** Ha igaz, a `pages` meta (cím / bevezető) még töltődik — a bevezető helyén „…”. */
    export let loading = false;
    /** Pl. kezdőlap bejelentkezett üdv — felülírja a `title`-t megjelenítésben és a böngésző címben. */
    export let titleOverride = /** @type {string | null} */ (null);

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

    $: displayTitle =
        titleOverride != null && String(titleOverride).trim() !== ""
            ? String(titleOverride).trim()
            : title;
    $: crumb = breadcrumbLabel.trim() || displayTitle;
    $: displayGreeting = loading ? "…" : (greeting != null ? String(greeting) : "");
</script>

<svelte:head>
    <title>{displayTitle}{documentTitleSuffix}</title>
</svelte:head>

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