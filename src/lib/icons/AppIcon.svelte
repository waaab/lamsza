<script>
    /**
     * Shared icons for admin (sidebar + welcome). Public toolbar markup lives in
     * (public)/+layout.svelte — copy paths here when an entity should match the site.
     */
    export let name = "entries";
    /** Rendered box size (px). Toolbar uses 24; admin sidebar 28; welcome grid 56. */
    export let size = 24;
    let className = "";
    export { className as class };
</script>

<span
    class="app-icon-wrap {className}"
    style:width="{size}px"
    style:height="{size}px"
    aria-hidden="true"
>
    {#if name === "home"}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
            ><path d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path><polyline points="9 22 9 12 15 12 15 22"></polyline></svg
        >
    {:else if name === "entries"}
        <!-- Same as homepage “Indexelünk” -->
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="10" x2="21" y1="6" y2="6"></line>
            <line x1="10" x2="21" y1="12" y2="12"></line>
            <line x1="10" x2="21" y1="18" y2="18"></line>
            <path d="M4 6h1v4"></path>
            <path d="M4 10h2"></path>
            <path d="M6 18H4c0-1 2-2 2-3s-1-1.5-2-1"></path>
        </svg>
    {:else if name === "mondasok"}
        <!-- Single Hungarian „ — large weight to match stroke icons (see admin.css .app-icon-quote) -->
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" overflow="visible" class="app-icon-quote">
            <text
                x="12"
                y="18"
                text-anchor="middle"
                font-size="24"
                font-weight="700"
                font-family="Georgia, 'Times New Roman', Times, serif">„</text>
        </svg>
    {:else if name === "quicklinks"}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
            ><path
                d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"
            ></path><path
                d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"
            ></path></svg
        >
    {:else if name === "newsfeeds"}
        <!-- Admin: newspaper (matches public /hirek toolbar art in +layout.svelte) -->
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a4 4 0 0 1-4-4V6"></path>
            <path d="M18 14h-8"></path>
            <path d="M15 18h-5"></path>
            <path d="M10 6h8v4h-8V6Z"></path>
        </svg>
    {:else if name === "counties"}
        <!-- Same shield as homepage “A megyétől” -->
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
        </svg>
    {:else if name === "locations"}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"
            ><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"></path><circle cx="12" cy="10" r="3"></circle></svg
        >
    {:else if name === "venues"}
        <!-- Stadium: rim, bowl, entrance notch, pennants (.app-icon-solid) -->
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon class="app-icon-solid" points="6.5,2.5 8,5 5,5" />
            <polygon class="app-icon-solid" points="12,1.5 13.5,4 10.5,4" />
            <polygon class="app-icon-solid" points="17.5,2.5 19,5 16,5" />
            <path d="M 5 9 Q 12 5.6 19 9"></path>
            <path d="M 5 9 L 5 15.5 Q 5 18.6 8.6 18.6"></path>
            <path d="M 19 9 L 19 15.5 Q 19 18.6 15.4 18.6"></path>
        </svg>
    {:else if name === "attractions"}
        <!-- View Apps–style triptych: taller outer edges, visible gaps -->
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M 2 3.5 L 8.5 5.2 L 8.5 18.8 L 2 20.5 Z"></path>
            <path d="M 9.5 5.2 H 14.5 V 18.8 H 9.5 Z"></path>
            <path d="M 15.5 5.2 L 22 3.5 L 22 20.5 L 15.5 18.8 Z"></path>
        </svg>
    {:else if name === "events"}
        <!-- Same calendar as public toolbar (+layout.svelte) -->
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M8 2v4"></path>
            <path d="M16 2v4"></path>
            <rect width="18" height="18" x="3" y="4" rx="2"></rect>
            <path d="M3 10h18"></path>
        </svg>
    {:else if name === "entry_categories"}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="8" y1="6" x2="21" y2="6"></line>
            <line x1="8" y1="12" x2="21" y2="12"></line>
            <line x1="8" y1="18" x2="21" y2="18"></line>
            <line x1="3" y1="6" x2="3.01" y2="6"></line>
            <line x1="3" y1="12" x2="3.01" y2="12"></line>
            <line x1="3" y1="18" x2="3.01" y2="18"></line>
        </svg>
    {:else if name === "entry_types"}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polygon points="12 2 2 7 12 12 22 7 12 2"></polygon>
            <polyline points="2 17 12 22 22 17"></polyline>
            <polyline points="2 12 12 17 22 12"></polyline>
        </svg>
    {:else if name === "szekek"}
        <!-- Side-view chair (matches public toolbar Székek) -->
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="6" y1="5" x2="6" y2="21"></line>
            <line x1="6" y1="15" x2="19" y2="15"></line>
            <line x1="19" y1="15" x2="19" y2="21"></line>
        </svg>
    {:else if name === "varosok"}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <rect width="8" height="18" x="3" y="3" rx="2"></rect>
            <path d="M7 7h0"></path>
            <path d="M7 11h0"></path>
            <path d="M7 15h0"></path>
            <rect width="8" height="12" x="13" y="9" rx="2"></rect>
            <path d="M17 13h0"></path>
            <path d="M17 17h0"></path>
        </svg>
    {:else if name === "falvak"}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 20v-8l7-5 7 5v8"></path>
            <path d="M7 20v-4h6v4"></path>
            <path d="M17 20h4v-7l-4-3"></path>
            <path d="M3 20h18"></path>
        </svg>
    {:else if name === "settings"}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            ><circle cx="12" cy="12" r="3"></circle><path
                d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"
            ></path></svg
        >
    {:else if name === "weather_translations"}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            ><path d="M5 8l6 6"></path><path d="M4 14l6-6 2-3"></path><path d="M2 5h12"></path><path d="M7 2h3l2 2 2 4-3 1"></path><path d="M22 22l-5-10-5 10"></path><path d="M14 18h6"></path></svg
        >
    {:else if name === "pages"}
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
            ><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line><polyline points="10 9 9 9 8 9"></polyline></svg
        >
    {/if}
</span>

<style>
    .app-icon-wrap {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        flex-shrink: 0;
    }
    .app-icon-wrap :global(svg) {
        display: block;
        width: 100%;
        height: 100%;
    }
</style>
