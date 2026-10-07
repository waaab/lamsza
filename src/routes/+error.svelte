<script>
    import { page } from "$app/stores";
    import ErrorShell from "$lib/components/ErrorShell.svelte";
</script>

<!--
  The root error page sits outside the (public) layout, so it never gets the
  site's toolbar, footer or data. The network's shared ErrorShell draws the
  minimal shell and the page (title, noindex, Hungarian text).
-->
<ErrorShell status={$page.status} appName="Lámsza" />
