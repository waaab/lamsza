<script>
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import Markdown from "$lib/components/Markdown.svelte";
    import NewsWidget from "$lib/components/NewsWidget.svelte";

    export let data;
</script>

<svelte:head>
    <title>{data.seat?.name || "Szék"} - Lámsza Index</title>
</svelte:head>

{#if data.notFound || !data.seat}
    <Breadcrumbs
        label="Nem található"
        parentLabel="Történelmi székek"
        parentUrl="/szekek"
    />
    <h1 class="page-title">Az oldal nem található</h1>
    <p class="greeting">
        <a href="/szekek">Vissza a történelmi székek listájához</a>
    </p>
{:else}
    <Breadcrumbs
        label={data.seat.name}
        parentLabel="Történelmi székek"
        parentUrl="/szekek"
    />

    <h1 class="page-title">{data.seat.name}</h1>
    <p class="greeting">
        Történelmi szék bemutatója, helyi hírek és részletes leírás.
    </p>

    <div class="widgets-box" id="hasznos-informaciok">
        <div id="attekintes" class="widget szek-widget-span">
            <div class="widget-header">
                <h3 class="widget-title">Áttekintés</h3>
            </div>
            <div class="more-info">
                <span title="Román neve"
                    >Románul: <span>{data.seat.name_ro || "–"}</span></span
                >
                <span title="Német neve"
                    >Németül: <span>{data.seat.name_de || "–"}</span></span
                >
            </div>
        </div>
    </div>

    {#if data.seat.content && data.seat.content.trim()}
        <div class="county-markdown markdown-region">
            <Markdown source={data.seat.content} />
        </div>
    {:else}
        <p class="info-box">
            Ehhez a székhez még nincs részletes szöveg megadva az adminban.
        </p>
    {/if}

    <NewsWidget ticker={true} />
{/if}

<style>
    .more-info {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }

    .widgets-box {
        display: grid;
        grid-template-columns: repeat(3, 1fr);
        gap: 2rem;
        margin-bottom: 2rem;
    }
    .szek-widget-span {
        grid-column: 1 / -1;
    }

    @media (max-width: 992px) {
        .widgets-box {
            grid-template-columns: 1fr;
        }
        .szek-widget-span {
            grid-column: 1;
        }
    }

    .county-markdown {
        margin-bottom: 1.5rem;
        max-width: 52rem;
    }
</style>
