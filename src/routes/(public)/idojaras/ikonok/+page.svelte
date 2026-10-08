<!--
    /idojaras/ikonok: every weather icon on one page, for reviewing the set
    (UI_BASELINE "wx-icons"). Not linked from the site; noindex.
-->
<script>
    import PublicPageHero from "$lib/components/PublicPageHero.svelte";
    import WeatherSymbol from "$lib/icons/weather/WeatherSymbol.svelte";
    import WeatherGlyph from "$lib/icons/weather/WeatherGlyph.svelte";
    import { SYMBOL_BASES, symbolEmoji, moonPhaseName } from "$lib/weatherSymbols.js";

    let animated = $state(true);
    let size = $state(64);

    const hasVariants = (/** @type {string} */ b) => ["clearsky", "fair", "partlycloudy"].includes(b) || b.includes("showers");
    const codes = SYMBOL_BASES.flatMap((b) => (hasVariants(b) ? [`${b}_day`, `${b}_night`] : [b]));

    const glyphs = [
        { kind: "thermometer", value: 22, label: "Hőmérséklet 22°" },
        { kind: "thermometer", value: -8, label: "Hőmérséklet -8°" },
        { kind: "feels", value: 18, label: "Hőérzet" },
        { kind: "wind", value: 300, label: "Szél ÉNy felől" },
        { kind: "gust", value: null, label: "Széllökés" },
        { kind: "humidity", value: 70, label: "Páratartalom 70%" },
        { kind: "dewpoint", value: 40, label: "Harmatpont" },
        { kind: "precip", value: 6, label: "Csapadék 6 mm" },
        { kind: "precip_prob", value: null, label: "Csapadék esélye" },
        { kind: "pressure", value: 1018, label: "Légnyomás 1018 hPa" },
        { kind: "uv", value: 2, label: "UV 2" },
        { kind: "uv", value: 7, label: "UV 7" },
        { kind: "cloud", value: null, label: "Felhőzet" },
        { kind: "fog", value: null, label: "Köd" },
        { kind: "sunrise", value: null, label: "Napkelte" },
        { kind: "sunset", value: null, label: "Napnyugta" },
        { kind: "daylength", value: null, label: "Nappal hossza" },
        { kind: "moonrise", value: null, label: "Holdkelte" },
        { kind: "moonset", value: null, label: "Holdnyugta" },
    ];
    const phases = [0, 45, 90, 135, 180, 225, 270, 315];
</script>

<svelte:head>
    <meta name="robots" content="noindex" />
</svelte:head>

<PublicPageHero
    title="Időjárás-ikonok"
    greeting="A teljes készlet: a MET Norway 41 időjárás-kódja nappal és éjjel, és a mért értékek rajzai."
    breadcrumbLabel="Ikonok"
    breadcrumbParentLabel="Időjárás"
    breadcrumbParentUrl="/idojaras"
/>

<div class="wx-controls">
    <label><input type="checkbox" bind:checked={animated} /> Animáció</label>
    <label>
        Méret
        <select bind:value={size}>
            <option value={32}>32 px</option>
            <option value={48}>48 px</option>
            <option value={64}>64 px</option>
            <option value={128}>128 px</option>
        </select>
    </label>
</div>

<h3 class="widget-title">Időjárás ({codes.length})</h3>
<div class="wx-gallery">
    {#each codes as code (code)}
        <figure class="card sm">
            <WeatherSymbol symbol={code} {size} {animated} />
            <figcaption><code>{code}</code> <span aria-hidden="true">{symbolEmoji(code)}</span></figcaption>
        </figure>
    {/each}
</div>

<h3 class="widget-title">Mért értékek</h3>
<div class="wx-gallery">
    {#each glyphs as g, i (i)}
        <figure class="card sm">
            <WeatherGlyph kind={g.kind} value={g.value} size={Math.min(size, 64)} {animated} />
            <figcaption>{g.label}</figcaption>
        </figure>
    {/each}
</div>

<h3 class="widget-title">Holdfázisok</h3>
<div class="wx-gallery">
    {#each phases as p (p)}
        <figure class="card sm">
            <WeatherGlyph kind="moonphase" value={p} size={Math.min(size, 64)} {animated} />
            <figcaption>{p}° · {moonPhaseName(p)}</figcaption>
        </figure>
    {/each}
</div>

<style>
    .wx-controls {
        display: flex;
        gap: 1.5rem;
        align-items: center;
        margin: 1rem 0;
    }
    .wx-gallery {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(10rem, 1fr));
        gap: 0.6rem;
        margin: 0.6rem 0 1.5rem;
    }
    .wx-gallery figure {
        margin: 0;
        align-items: center;
        gap: 0.5rem;
    }
    .wx-gallery figcaption {
        font-size: 0.75rem;
        color: var(--text-faint);
        text-align: center;
        word-break: break-word;
    }
</style>
