<script>
    import { onMount } from "svelte";
    import Breadcrumbs from "$lib/components/Breadcrumbs.svelte";
    import { apiFetch } from "$lib/api";

    let page = null;
    let loading = true;
    let error = false;

    onMount(async () => {
        try {
            page = await apiFetch("/api/pages?slug=iranyelvek/sutik");
        } catch {
            error = true;
        } finally {
            loading = false;
        }
    });
</script>

<svelte:head>
    <title>{page?.title || "Sütik"} – Lámsza</title>
</svelte:head>

<Breadcrumbs label="Sütik" parentLabel="Irányelvek" parentUrl="/iranyelvek" />

<section class="page-section">
    <h1 class="page-title">{page?.title || "Sütik"}</h1>
    <div class="page-inner">
        {#if loading}
            <span class="info-box">
                <p>Betöltés...</p>
            </span>
        {:else if error}
            <span class="info-box">
                <p>Az oldal nem elérhető.</p>
            </span>
        {:else if page?.content}
            <div class="page-content">{@html page.content}</div>
        {:else}
            <span class="info-box">
                <p>Az oldal tartalma még nem lett hozzáadva.</p>
            </span>
        {/if}
        <nav class="page-nav">
            <a class="nav-btn" href="/iranyelvek">Irányelvek</a>
            <a class="nav-btn" href="/iranyelvek/feltetelek">Feltételek</a>
        </nav>
    </div>
</section>

<style>
    .page-nav {
        display: flex;
        gap: 0.5rem;
    }
</style>
