<script>
    import { onMount } from "svelte";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { initialPageHeader, loadPageMeta } from "$lib/loadPageMeta.js";

    const fb = initialPageHeader("iranyelvek");
    let page = null;
    let loading = true;
    let error = false;

    onMount(async () => {
        try {
            page = await loadPageMeta("iranyelvek");
        } catch {
            error = true;
            page = { ...fb, content: "" };
        } finally {
            loading = false;
        }
    });
</script>

<section class="page-section">
    <PublicPageHero
        title={page?.title ?? fb.title}
        greeting={page?.greeting ?? fb.greeting}
        loading={false}
        breadcrumbLabel="Irányelvek"
        breadcrumbParentLabel=""
        breadcrumbParentUrl=""
        documentTitleSuffix=" – Lámsza"
    />

    <div class="page-inner">
        {#if loading}
            <div class="info-box"><p>Betöltés…</p></div>
        {:else if error}
            <div class="info-box"><p>Az oldal nem elérhető.</p></div>
        {:else if page?.content}
            <div class="page-content">{@html page.content}</div>
        {:else}
            <div class="info-box"><p>Az oldal tartalma még nem lett hozzáadva.</p></div>
        {/if}
    </div>
    <nav class="page-nav">
        <h4 class="page-nav-title">Oldal navigáció</h4>
        <ul>
            <li><a class="btn nav-btn" href="/iranyelvek/sutik">Sütik</a></li>
            <li><a class="btn nav-btn" href="/iranyelvek/feltetelek">Feltételek</a></li>
        </ul>
    </nav>
</section>