<script>
    import { page } from "$app/stores";
    import { theme, cycleTheme, LABELS } from "$lib/stores/theme";
    import { fade } from "svelte/transition";

    let settingsOpen = false;
    let scrollY = 0;

    function toggleSettings() {
        settingsOpen = !settingsOpen;
    }

    function closeSettings() {
        settingsOpen = false;
    }

    function scrollToTop() {
        if (typeof window !== "undefined") {
            window.scrollTo({ top: 0, behavior: "smooth" });
        }
    }
</script>

<svelte:window bind:scrollY on:click={closeSettings} />

<header class="toolbar">
    <div class="nav">
        <a href="/" class="nav-btn" title="Vissza a főoldalra">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><path
                    d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"
                /><polyline points="9 22 9 12 15 12 15 22" /></svg
            >
            <span>Lámsza</span>
        </a>
        <a href="/hirek" class="nav-btn" title="Hírek">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><path
                    d="M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a4 4 0 0 1-4-4V6"
                /><path d="M18 14h-8" /><path d="M15 18h-5" /><path
                    d="M10 6h8v4h-8V6Z"
                /></svg
            >
            <span>Hírek</span>
        </a>
        <a href="/index" class="nav-btn" title="Index">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
            >
                <line x1="10" x2="21" y1="6" y2="6" />
                <line x1="10" x2="21" y1="12" y2="12" />
                <line x1="10" x2="21" y1="18" y2="18" />
                <path d="M4 6h1v4" />
                <path d="M4 10h2" />
                <path d="M6 18H4c0-1 2-2 2-3s-1-1.5-2-1" />
            </svg>
            <span>Index</span>
        </a>
        <a href="/megyek" class="nav-btn" title="Székelyföldi Megyék">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
            >
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span>Megyék</span>
        </a>
        <a href="/varosok" class="nav-btn" title="Székelyföldi Városok">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
            >
                <rect width="8" height="18" x="3" y="3" rx="2" />
                <path d="M7 7h0" />
                <path d="M7 11h0" />
                <path d="M7 15h0" />
                <rect width="8" height="12" x="13" y="9" rx="2" />
                <path d="M17 13h0" />
                <path d="M17 17h0" />
            </svg>
            <span>Városok</span>
        </a>
        <a href="/falvak" class="nav-btn" title="Székelyföldi Falvak">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
            >
                <path d="M3 20v-8l7-5 7 5v8"></path>
                <path d="M7 20v-4h6v4"></path>
                <path d="M17 20h4v-7l-4-3"></path>
                <path d="M3 20h18"></path>
            </svg>
            <span>Falvak</span>
        </a>
        <a href="/esemenyek" class="nav-btn" title="Események">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
            >
                <path d="M8 2v4" />
                <path d="M16 2v4" />
                <rect width="18" height="18" x="3" y="4" rx="2" />
                <path d="M3 10h18" />
            </svg>
            <span>Események</span>
        </a>
    </div>
    <div class="nav">
        <a href="/admin" class="nav-btn" title="Belépés az admin panelbe">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" /><circle
                    cx="12"
                    cy="7"
                    r="4"
                /></svg
            >
            <span>Bélépés</span>
        </a>
        <a href="/valtozasnaplo" class="nav-btn" title="Változásnapló">
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><path d="M11 15h2" /><path d="M11 9h2" /><path
                    d="M8 21h8a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2z"
                /><path d="M9 12h6" /></svg
            >
            <span>Változások</span>
        </a>

        <div class="settings-container">
            <button
                class="nav-btn"
                on:click|stopPropagation={toggleSettings}
                aria-label="Beállítások megnyitása"
                aria-expanded={settingsOpen}
            >
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="24"
                    height="24"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    ><circle cx="12" cy="12" r="3" /><path
                        d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"
                    /></svg
                >
                <span>Béállítások</span>
            </button>

            {#if settingsOpen}
                <div
                    class="settings-dropdown"
                    on:click|stopPropagation
                    role="presentation"
                >
                    <div class="dropdown-header">Béállítások</div>
                    <button
                        class="dropdown-item"
                        on:click={() => cycleTheme($theme)}
                    >
                        <span class="dropdown-item-label">Téma:</span>
                        <span class="dropdown-item-value">{LABELS[$theme]}</span
                        >
                    </button>
                </div>
            {/if}
        </div>
    </div>
</header>

<main class="container">
    <slot />
</main>

<footer>
    <div class="footer-top">
        Készítette sok ❤️-el <a
            href="https://bogozi.com"
            target="_blank"
            rel="nofollow noopener"
            title="bogozi.com - webfejlesztés, webshop készítés, keresőoptimalizálás"
            >bogozi.com</a
        >
        © {new Date().getFullYear()} &bull; Na lámsza - Erdélyi magyar startlap
        és kereső. Az internet székely kapuja &bull;
        <a href="/valtozasnaplo" title="Verzió és Változásnapló"
            >v1.0.0 - Változásnapló</a
        >
    </div>
    <div class="footer-bottom">
        <div class="logo">Székely Gugel</div>
        <div class="social-links">
            <a href="/" target="_blank" rel="noopener" title="Facebook"
                >Facebook</a
            >
            <a href="/" target="_blank" rel="noopener" title="Twitter"
                >Twitter</a
            >
            <a href="/" target="_blank" rel="noopener" title="Instagram"
                >Instagram</a
            >
        </div>
    </div>
</footer>

{#if scrollY > 500}
    <button
        class="btn back-to-top"
        on:click={scrollToTop}
        aria-label="Ugrás az oldal tetejére"
        transition:fade={{ duration: 200 }}
    >
        ↑
    </button>
{/if}
