<script>
    import { page } from "$app/stores";
    import ErrorPage from "$lib/components/ErrorPage.svelte";
    import AppIcon from "$lib/icons/AppIcon.svelte";
</script>

<!--
  The root error page does not get the public layout, so it carries Lámsza's
  minimal shell itself: the toolbar's home button, no footer. The page body is
  the network's shared ErrorPage (title, noindex, Hungarian text).
-->
<div class="layout-bg error-layout">
    <header class="toolbar error-toolbar">
        <div class="nav">
            <a href="/" class="btn nav-btn" title="Vissza a főoldalra">
                <AppIcon name="home" size={16} />
                <span>Lámsza</span>
            </a>
        </div>
    </header>

    <main class="container home-main error-main">
        <ErrorPage status={$page.status} appName="Lámsza" />
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
</style>
