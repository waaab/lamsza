<script>
    import { onMount } from "svelte";
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import { loadPageMeta, initialPageHeader } from "$lib/loadPageMeta.js";

    let pageHeader = initialPageHeader("terkep");
    let pageHeaderLoading = false;

    onMount(async () => {
        pageHeader = await loadPageMeta("terkep");
        pageHeaderLoading = false;
    });
</script>

<PublicPageHero
    title={pageHeader.title}
    greeting={pageHeader.greeting}
    loading={pageHeaderLoading}
    showBreadcrumbs={false}
    documentTitleSuffix=" - Székely Gugel"
/>

<div class="map-container">
    <!-- Placeholder SVG for Map -->
    <svg
        xmlns="http://www.w3.org/2000/svg"
        width="100%"
        height="100%"
        viewBox="0 0 800 600"
        preserveAspectRatio="xMidYMid slice"
        class="map-skeleton-svg"
    >
        <defs>
            <pattern
                id="grid"
                width="40"
                height="40"
                patternUnits="userSpaceOnUse"
            >
                <path
                    d="M 40 0 L 0 0 0 40"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="1"
                />
            </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#grid)" />
        <path
            d="M200,100 Q400,50 600,150 T650,400 T300,500 T150,300 Z"
            fill="none"
            stroke="currentColor"
            stroke-width="4"
        />
        <circle cx="350" cy="250" r="15" fill="var(--szekely-red)" />
        <circle cx="500" cy="350" r="10" fill="var(--szekely-blue)" />
        <circle cx="250" cy="400" r="8" fill="currentColor" />
    </svg>

    <div class="map-overlay">
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="48"
            height="48"
            viewBox="0 0 24 24"
            fill="none"
            stroke="var(--text-color)"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            ><polygon points="3 6 9 3 15 6 21 3 21 18 15 21 9 18 3 21"
            ></polygon><line x1="9" y1="3" x2="9" y2="18"></line><line
                x1="15"
                y1="6"
                x2="15"
                y2="21"
            ></line></svg
        >
        <h2 class="overlay-title">Fejlesztés alatt</h2>
        <p class="overlay-text">
            Ide kerül az OpenStreetMap (OSM) TileServer integráció.
        </p>
    </div>
</div>

<style>
    .map-container {
        width: 100%;
        height: 60vh;
        background-color: var(--surface-bg);
        border-radius: var(--radius-md);
        display: flex;
        align-items: center;
        justify-content: center;
        overflow: hidden;
        position: relative;
    }
    .map-skeleton-svg {
        opacity: 0.1;
    }
    .map-overlay {
        position: absolute;
        text-align: center;
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 1rem;
    }
    .overlay-title {
        margin: 0;
    }
    .overlay-text {
        margin: 0;
        color: var(--text-faint);
    }
</style>