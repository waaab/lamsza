<!--
    The source line every weather view shows next to its data. Each source's
    licence asks for it: MET Norway (CC BY 4.0) a credit and the licence link,
    WeatherAPI.com a link back, OpenWeatherMap (ODbL) a visible credit.
-->
<script>
    import { sourceCredit } from "$lib/weatherSymbols.js";

    let { source = "metno", fetchedAt = null } = $props();

    const credit = $derived(sourceCredit(source));
    const time = $derived(
        fetchedAt
            ? new Intl.DateTimeFormat("hu-HU", {
                  timeZone: "Europe/Bucharest",
                  hour: "2-digit",
                  minute: "2-digit",
                  hourCycle: "h23",
              }).format(new Date(fetchedAt))
            : "",
    );
</script>

<p class="wx-credit">
    {#if time}<span>Frissítve: {time}</span> · {/if}
    <span>{credit.text} <a href={credit.href} target="_blank" rel="noopener noreferrer">{credit.linkText}</a></span>
</p>

<style>
    .wx-credit {
        margin: 0.75rem 0 0;
        font-size: 0.8rem;
        color: var(--text-faint);
    }
    .wx-credit a {
        color: inherit;
        text-decoration: underline;
    }
</style>
