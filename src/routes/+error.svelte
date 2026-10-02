<script>
    import { page } from "$app/stores";
</script>

<svelte:head>
    <title>Hiba {$page.status} - Székely Gugel</title>
</svelte:head>

<div class="layout-bg error-layout">
    <!-- Simple header specifically for the root error page when it doesn't get the public layout -->
    <header class="toolbar error-toolbar">
        <div class="nav">
            <a href="/" class="btn nav-btn" title="Vissza a főoldalra">
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
        </div>
    </header>

    <main class="container home-main error-main">
        <h1 class="page-title error-title">
            Hoppácska! (Hiba: {$page.status})
        </h1>
        <p class="greeting error-greeting">
            {$page.error?.message || "Valami hiba történt a rendszerben."}
        </p>

        <div class="error-actions">
            <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
            <button class="btn" on:click={() => (window.location.href = "/")}>
                Vissza a főoldalra
            </button>
        </div>
    </main>
</div>

<style>
    .error-layout {
        display: flex;
        flex-direction: column;
        min-height: 100vh;
    }
    .error-toolbar {
        padding: 1rem;
    }
    .error-main {
        flex: 1;
    }
    .error-title {
        color: var(--szekely-red);
    }
    .error-greeting {
        text-align: center;
        margin-top: 1rem;
        color: var(--text-primary);
    }
    .error-actions {
        text-align: center;
        margin-top: 2rem;
    }
</style>