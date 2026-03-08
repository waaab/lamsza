<script>
    import { onMount } from "svelte";
    import { apiFetch } from "$lib/api";

    let page = null;
    let loading = true;
    let error = false;

    onMount(async () => {
        try {
            page = await apiFetch("/api/pages?slug=iranyelvek/feltetelek");
        } catch {
            error = true;
        } finally {
            loading = false;
        }
    });
</script>

<svelte:head>
    <title>{page?.title || "Feltételek"} – Lámsza</title>
</svelte:head>

<section class="page-section">
    <h1 class="page-title">{page?.title || "Feltételek"}</h1>

    {#if loading}
        <p class="info-box">Betöltés...</p>
    {:else if error}
        <p class="info-box">Az oldal nem elérhető.</p>
    {:else if page?.content}
        <div class="page-content">{@html page.content}</div>
    {:else}
        <p class="info-box">Az oldal tartalma még nem lett hozzáadva.</p>
    {/if}

    <nav class="policy-nav">
        <a href="/iranyelvek">Irányelvek</a>
        <a href="/iranyelvek/sutik">Sütik</a>
    </nav>
</section>

<style>
    .policy-nav {
        display: flex;
        gap: 1rem;
        margin-top: 2rem;
        padding-top: 1rem;
        border-top: 1px solid var(--border-color, #ddd);
    }
    .policy-nav a {
        color: var(--link-color, #2563eb);
        text-decoration: none;
        font-weight: 500;
    }
    .policy-nav a:hover {
        text-decoration: underline;
    }
</style>
