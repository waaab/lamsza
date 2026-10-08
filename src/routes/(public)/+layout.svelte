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
    import AppsLauncher from "$lib/components/AppsLauncher.svelte";
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
            title="Vissza a főoldalra"
        >
            <AppIcon name="home" size={16} />
            <span>Lámsza</span>
        </a>
        <a
            href="/index"
            class="btn nav-btn {$page.url.pathname.startsWith('/index')
                ? 'active'
                : ''}"
            title="Index"
        >
            <AppIcon name="entries" size={16} />
            <span>Indexelünk</span>
        </a>
        <a
            href="/hirek"
            class="btn nav-btn {$page.url.pathname.startsWith('/hirek')
                ? 'active'
                : ''}"
            title="Hírek"
        >
            <AppIcon name="newsfeeds" size={16} />
            <span>Erdélyi Hírek</span>
        </a>
        <a
            href="/esemenyek"
            class="btn nav-btn {$page.url.pathname.startsWith('/esemenyek')
                ? 'active'
                : ''}"
            title="Események"
        >
            <AppIcon name="events" size={16} />
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
            <AppIcon name="szekek" size={16} />
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
            <AppIcon name="counties" size={16} />
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
            <AppIcon name="varosok" size={16} />
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
            <AppIcon name="falvak" size={16} />
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
                <AppIcon name="logout" size={16} />
                <span class="sr-only">Kijelentkezés</span>
            </button>
        {:else}
            <button
                type="button"
                class="btn nav-btn"
                title="Belépés"
                on:click={openLogin}
            >
                <AppIcon name="login" size={16} />
                <span class="sr-only">Belépés</span>
            </button>
        {/if}

        <a
            href="/beallitasok"
            class="btn nav-btn {$page.url.pathname === '/beallitasok' ? 'active' : ''}"
            title="Felhasználói beállítások"
        >
            <AppIcon name="settings" size={16} />
            <span class="sr-only">Felhasználói beállítások</span>
        </a>
        <AppsLauncher current="lamsza" hostname={$page.url.hostname} />
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
            <a href="/iranyelvek/adatvedelem" title="Adatvédelem">Adatvédelem</a>
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

</style>