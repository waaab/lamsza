<script>
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { page } from "$app/stores";
    import PageFaqDisclaimer from "$lib/components/PageFaqDisclaimer.svelte";
    import { deriveFaqSectionKey } from "$lib/pageFaqSection.js";

    $: faqSectionKey = deriveFaqSectionKey($page.url.pathname);
    import { auth } from "$lib/stores/auth";
    import { theme, cycleTheme, LABELS } from "$lib/stores/theme";
    import { fade } from "svelte/transition";

    let settingsOpen = false;
    let scrollY = 0;
    let loginDialogOpen = false;
    let loginPassword = "";
    let loginError = "";

    const ADMIN_PASSWORD = "szekely123";

    onMount(() => {
        auth.init();
    });

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

    function logout() {
        auth.logout();
    }

    function openLoginDialog() {
        loginDialogOpen = true;
        loginPassword = "";
        loginError = "";
    }

    function closeLoginDialog() {
        loginDialogOpen = false;
        loginPassword = "";
        loginError = "";
    }

    function submitLogin(e) {
        e.preventDefault();
        loginError = "";
        if (loginPassword === ADMIN_PASSWORD) {
            auth.login("Admin", true);
            closeLoginDialog();
        } else {
            loginError = "Na de kicsibarátom, ez nem a jó jelszó!";
        }
    }

</script>

<svelte:window
    bind:scrollY
    on:click={closeSettings}
    on:keydown={(e) => loginDialogOpen && e.key === "Escape" && closeLoginDialog()}
/>

<header class="toolbar">
    <div class="nav">
        <a
            href="/"
            class="nav-btn {$page.url.pathname === '/' ? 'active' : ''}"
            title="Vissza a főódalra"
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
                ><path
                    d="m3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"
                /><polyline points="9 22 9 12 15 12 15 22" /></svg
            >
            <span>Lámsza</span>
        </a>
        <a
            href="/index"
            class="nav-btn {$page.url.pathname.startsWith('/index')
                ? 'active'
                : ''}"
            title="Index"
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
            >
                <line x1="10" x2="21" y1="6" y2="6" />
                <line x1="10" x2="21" y1="12" y2="12" />
                <line x1="10" x2="21" y1="18" y2="18" />
                <path d="M4 6h1v4" />
                <path d="M4 10h2" />
                <path d="M6 18H4c0-1 2-2 2-3s-1-1.5-2-1" />
            </svg>
            <span>Indexelünk</span>
        </a>
        <a
            href="/hirek"
            class="nav-btn {$page.url.pathname.startsWith('/hirek')
                ? 'active'
                : ''}"
            title="Hírek"
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
                ><path
                    d="M4 22h16a2 2 0 0 0 2-2V4a2 2 0 0 0-2-2H8a2 2 0 0 0-2 2v16a4 4 0 0 1-4-4V6"
                /><path d="M18 14h-8" /><path d="M15 18h-5" /><path
                    d="M10 6h8v4h-8V6Z"
                /></svg
            >
            <span>Erdélyi Hírek</span>
        </a>
        <a
        href="/esemenyek"
        class="nav-btn {$page.url.pathname.startsWith('/esemenyek')
            ? 'active'
            : ''}"
        title="Események"
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
        >
            <path d="M8 2v4" />
            <path d="M16 2v4" />
            <rect width="18" height="18" x="3" y="4" rx="2" />
            <path d="M3 10h18" />
        </svg>
        <span>Kik verekettek?</span>
    </a>
    <a
    href="/szekek"
    class="nav-btn {$page.url.pathname === '/szekek' ||
    $page.url.pathname.startsWith('/szekek/')
        ? 'active'
        : ''}"
    title="Történelmi székek"
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
    >
        <path d="M12 2L2 7l10 5 10-5-10-5z" />
        <path d="M2 17l10 5 10-5" />
        <path d="M2 12l10 5 10-5" />
    </svg>
    <span>Székek</span>
</a>
        <a
            href="/megyek"
            class="nav-btn {$page.url.pathname === '/megyek' ||
            $page.url.pathname.includes('-megye')
                ? 'active'
                : ''}"
            title="Székelyföldi Megyék"
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
            >
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
            </svg>
            <span class="nav-btn-label">A megyétől</span>
            <span class="sr-only nav-btn-label-lang szekely-hungarian">Megyék</span>
        </a>

        <a
            href="/varosok"
            class="nav-btn {$page.url.pathname === '/varosok' ||
            $page.url.pathname.startsWith('/varos/')
                ? 'active'
                : ''}"
            title="Székelyföldi Városok"
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
            >
                <rect width="8" height="18" x="3" y="3" rx="2" />
                <path d="M7 7h0" />
                <path d="M7 11h0" />
                <path d="M7 15h0" />
                <rect width="8" height="12" x="13" y="9" rx="2" />
                <path d="M17 13h0" />
                <path d="M17 17h0" />
            </svg>
            <span>Városiak</span>
        </a>
        <a
            href="/falvak"
            class="nav-btn {$page.url.pathname === '/falvak' ||
            $page.url.pathname.startsWith('/falu/')
                ? 'active'
                : ''}"
            title="Székelyföldi Falvak"
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
            >
                <path d="M3 20v-8l7-5 7 5v8"></path>
                <path d="M7 20v-4h6v4"></path>
                <path d="M17 20h4v-7l-4-3"></path>
                <path d="M3 20h18"></path>
            </svg>
            <span>Falusiak</span>
        </a>
    </div>
    <div class="nav">
        {#if $auth.loggedIn}
            <span class="nav-admin-user" title="{$auth.user} bejelentkezve">{$auth.user}</span>
            {#if $auth.isAdmin}
            <button
                type="button"
                class="nav-btn {$page.url.pathname.startsWith('/admin') ? 'active' : ''}"
                title="Admin panel"
                on:click={() => goto('/admin')}
            >
                <span class="sr-only">Admin</span>
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
            </button>
            {/if}
            <button
                type="button"
                class="nav-btn"
                on:click={logout}
                title="Kijelentkezés"
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
                    ><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><polyline points="16 17 21 12 16 7" /><line x1="21" y1="12" x2="9" y2="12" /></svg
                >
                <span class="sr-only">Kijelentkezés</span>
            </button>
        {:else}
            <button
                type="button"
                class="nav-btn"
                title="Belépés"
                on:click={openLoginDialog}
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
                    ><path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" /><polyline points="10 17 15 12 10 7" /><line x1="15" y1="12" x2="3" y2="12" /></svg
                >
                <span class="sr-only">Belépés</span>
            </button>
        {/if}

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
                <span class="sr-only">Béállítások</span>
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
    <!-- /valtozasnaplo keeps its own version changelog as FAQ-shaped content -->
    {#if $page.url.pathname !== "/valtozasnaplo"}
        <PageFaqDisclaimer sectionKey={faqSectionKey} />
    {/if}
</main>

<footer>
    <div class="copyright">
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
    <div class="brand-info">
        <div class="logo">Na Lámsza!</div>
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
    <div class="policy-links">
        <a href="/iranyelvek" title="Irányelvek">Irányelvek</a>
        <a href="/iranyelvek/feltetelek" title="Feltételek">Feltételek</a>
        <a href="/iranyelvek/sutik" title="Sütik">Sütik</a>
    </div>
</footer>

{#if loginDialogOpen}
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div
        class="link-dialog-overlay"
        role="dialog"
        aria-labelledby="login-dialog-title"
        tabindex="-1"
        on:click|self={closeLoginDialog}
    >
        <div class="link-dialog" on:click|stopPropagation>
            <h3 id="login-dialog-title">Belépés</h3>
            <form class="link-dialog-form" on:submit|preventDefault={submitLogin}>
                <label for="login_password">Jelszó</label>
                <input
                    id="login_password"
                    type="password"
                    bind:value={loginPassword}
                    placeholder="Jelszó..."
                    required
                />
                {#if loginError}
                    <p class="login-error">{loginError}</p>
                {/if}
                <div class="link-dialog-actions">
                    <button type="submit" class="link-dialog-submit">Belépés</button>
                    <button type="button" class="link-dialog-cancel" on:click={closeLoginDialog}>Mégse</button>
                </div>
            </form>
        </div>
    </div>
{/if}

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
