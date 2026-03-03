<script>
    import { page } from "$app/stores";
</script>

<svelte:head>
    <title>Hiba {$page.status} - Székely Gugel</title>
</svelte:head>

<div
    class="layout-bg"
    style="display: flex; flex-direction: column; min-height: 100vh;"
>
    <!-- Simple header specifically for the root error page when it doesn't get the public layout -->
    <header class="toolbar" style="padding: 1rem;">
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
        </div>
    </header>

    <main class="container home-main" style="flex: 1;">
        <h1 class="page-title" style="color: var(--szekely-red);">
            Hoppácska! (Hiba: {$page.status})
        </h1>
        <p
            class="greeting"
            style="text-align: center; margin-top: 1rem; color: var(--text-primary);"
        >
            {$page.error?.message || "Valami hiba történt a rendszerben."}
        </p>

        <div style="text-align: center; margin-top: 2rem;">
            <button class="btn" onclick={() => navigate("/")}>
                Vissza a főoldalra
            </button>
        </div>
    </main>
</div>
