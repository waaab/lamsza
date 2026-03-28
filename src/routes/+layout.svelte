<script>
  import "../styles/global.css";
  import { onMount } from "svelte";
  import { theme, applyTheme } from "$lib/stores/theme";

  let mediaQuery;

  onMount(() => {
    // Stale SW (e.g. from an old static deploy) breaks Vite dev chunk loading (.svelte-kit/generated/...).
    if (import.meta.env.DEV && "serviceWorker" in navigator) {
      navigator.serviceWorker.getRegistrations().then((regs) => {
        for (const r of regs) r.unregister();
      });
    }

    const saved = localStorage.getItem("theme") || "system";
    applyTheme(saved);

    mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleSystemChange = () => {
      if (!document.documentElement.hasAttribute("data-theme")) {
        // re-sync if needed
      }
    };
    mediaQuery.addEventListener("change", handleSystemChange);

    return () => {
      mediaQuery.removeEventListener("change", handleSystemChange);
    };
  });
</script>

<slot />
