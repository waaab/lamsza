<script>
  import "../styles/global.css";
  import { onMount } from "svelte";
  import { theme, applyTheme } from "$lib/stores/theme";

  let mediaQuery;

  onMount(() => {
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
