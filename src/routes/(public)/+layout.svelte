<script>
    import { onMount } from "svelte";
    import { fade } from "svelte/transition";
    import { page } from "$app/stores";
    import PageFaqDisclaimer from "$lib/components/PageFaqDisclaimer.svelte";
    import { deriveFaqSectionKey } from "$lib/pageFaqSection.js";

    $: faqSectionKey = deriveFaqSectionKey($page.url.pathname);
    import { auth } from "$lib/stores/auth";
    import { apiFetch } from "$lib/api.js";
    import SignInDialog from "$lib/components/SignInDialog.svelte";
    import AppIcon from "$lib/icons/AppIcon.svelte";
    import { openLogin, listenForOpenLogin } from "$lib/openLogin.js";
    import { APP_VERSION } from "$lib/publicChangelog.js";

    const NETWORK_LAUNCH_YEAR = 2009;
    const year = new Date().getFullYear();

    let scrollY = 0;
    let loginDialogOpen = false;
    let googleClientId = "";
    let configLoaded = false;
    let socialLinks = [];

    function httpSocialLinks(raw) {
        if (!Array.isArray(raw)) return [];
        return raw.filter(
            (link) =>
                link &&
                String(link.label || "").trim() &&
                /^https?:\/\//i.test(String(link.url || "").trim()),
        );
    }

    onMount(async () => {
        const stopLoginListener = listenForOpenLogin(openLoginDialog);
        await auth.init();
        try {
            const cfg = await apiFetch("/api/config/public");
            googleClientId = cfg.google_client_id || "";
            socialLinks = httpSocialLinks(cfg.social_links);
        } catch (e) {
            console.error(e);
        }
        configLoaded = true;
        return stopLoginListener;
    });

    function scrollToTop() {
        if (typeof window !== "undefined") {
            window.scrollTo({ top: 0, behavior: "smooth" });
        }
    }

    async function logout() {
        await auth.logout();
    }

    function openLoginDialog() {
        loginDialogOpen = true;
    }

    function closeLoginDialog() {
        loginDialogOpen = false;
    }

    async function onGoogleSignedIn() {
        await auth.refresh();
        closeLoginDialog();
    }

</script>

<svelte:window
    bind:scrollY
    on:keydown={(e) => loginDialogOpen && e.key === "Escape" && closeLoginDialog()}
/>

<header class="toolbar">
    <div class="nav">
        <a
            href="/"
            class="btn nav-btn {$page.url.pathname === '/' ? 'active' : ''}"
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
                aria-hidden="true"
            >
                <circle cx="11" cy="11" r="8" />
                <path d="m21 21-4.3-4.3" />
            </svg>
            <span>Lámsza</span>
        </a>
        <a
            href="/index"
            class="btn nav-btn {$page.url.pathname.startsWith('/index')
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
            class="btn nav-btn {$page.url.pathname.startsWith('/hirek')
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
        class="btn nav-btn {$page.url.pathname.startsWith('/esemenyek')
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
    class="btn nav-btn {$page.url.pathname === '/szekek' ||
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
        aria-hidden="true"
    >
        <!-- Side-view chair: back + rear leg, seat, front leg -->
        <line x1="6" y1="5" x2="6" y2="21" />
        <line x1="6" y1="15" x2="19" y2="15" />
        <line x1="19" y1="15" x2="19" y2="21" />
    </svg>
    <span>Székek</span>
</a>
        <a
            href="/megyek"
            class="btn nav-btn {$page.url.pathname === '/megyek' ||
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
            class="btn nav-btn {$page.url.pathname === '/varosok' ||
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
            class="btn nav-btn {$page.url.pathname === '/falvak' ||
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
            <a
                href="/fiok"
                class="btn nav-btn {$page.url.pathname === '/fiok' ? 'active' : ''}"
                title="Fiók"
            >
                <AppIcon name="profile" size={16} />
                <span class="sr-only">Fiók</span>
            </a>
            <button
                type="button"
                class="btn nav-btn"
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
                class="btn nav-btn"
                title="Belépés"
                on:click={openLogin}
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

        <a
            href="/beallitasok"
            class="btn nav-btn {$page.url.pathname === '/beallitasok' ? 'active' : ''}"
            title="Felhasználói beállítások"
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
            <span class="sr-only">Felhasználói beállítások</span>
        </a>
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
        Sok ❤️-el Székelyföldről.
        © {NETWORK_LAUNCH_YEAR} - {year} &bull; Na lámsza &bull; Erdélyi magyar startlap
        és kereső. Az internet székely kapuja.
    </div>
    <div class="footer-bottom">
        {#if socialLinks.length}
            <div class="brand-info">
                <div class="social-links">
                    {#each socialLinks as link (link.url)}
                        <a href={link.url} target="_blank" rel="noopener" title={link.label}>{link.label}</a>
                    {/each}
                </div>
            </div>
        {/if}
        <div class="policy-links">
            <a href="/iranyelvek" title="Irányelvek">Irányelvek</a>
            <a href="/iranyelvek/feltetelek" title="Feltételek">Feltételek</a>
            <a href="/iranyelvek/sutik" title="Sütik">Sütik</a>
            &bull;
            <a href="/valtozasnaplo" title="Verzió és Változásnapló"
                >v{APP_VERSION} - Változásnapló</a
            >
        </div>
    </div>
</footer>

<SignInDialog
    open={loginDialogOpen}
    appName="Lámsza"
    clientId={googleClientId}
    configReady={configLoaded}
    onClose={closeLoginDialog}
    onSignedIn={onGoogleSignedIn}
/>

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

<style>
    .nav-btn--admin {
        position: relative;
    }

    .nav-admin-badge {
        position: absolute;
        top: -0.35rem;
        right: -0.35rem;
        min-width: 1.1rem;
        height: 1.1rem;
        padding: 0 0.25rem;
        border-radius: 999px;
        background: var(--szekely-red, #c8102e);
        color: #fff;
        font-size: 0.65rem;
        font-weight: 700;
        line-height: 1.1rem;
        text-align: center;
        pointer-events: none;
    }
</style>