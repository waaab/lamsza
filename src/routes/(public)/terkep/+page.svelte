<script>
    import { page } from "$app/stores";
</script>

<svelte:head>
    <title>Szekely Gugel - Térkép</title>
</svelte:head>

<div class="container" style="min-height: calc(100vh - 120px)">
    <h1 class="page-title">Erdély Térkép</h1>
    <p class="greeting">
        Hamarosan érkezik az interaktív térképünk helyi adatokkal!
    </p>

    <div
        style="width: 100%; height: 60vh; background-color: var(--surface-bg); border-radius: var(--radius-md); display: flex; align-items: center; justify-content: center; overflow: hidden; position: relative;"
    >
        <!-- Placeholder SVG for Map -->
        <svg
            xmlns="http://www.w3.org/2000/svg"
            width="100%"
            height="100%"
            viewBox="0 0 800 600"
            preserveAspectRatio="xMidYMid slice"
            style="opacity: 0.1;"
        >
            <defs>
                <pattern
                    id="grid"
                    width="40"
                    height="40"
                    patternUnits="userSpaceOnUse"
                >
                    <path
                        d="M 40 0 L 0 0 0 40"
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1"
                    />
                </pattern>
            </defs>
            <rect width="100%" height="100%" fill="url(#grid)" />
            <path
                d="M200,100 Q400,50 600,150 T650,400 T300,500 T150,300 Z"
                fill="none"
                stroke="currentColor"
                stroke-width="4"
            />
            <circle cx="350" cy="250" r="15" fill="var(--szekely-red)" />
            <circle cx="500" cy="350" r="10" fill="var(--szekely-blue)" />
            <circle cx="250" cy="400" r="8" fill="currentColor" />
        </svg>

        <div
            style="position: absolute; text-align: center; display: flex; flex-direction: column; align-items: center; gap: 1rem;"
        >
            <svg
                xmlns="http://www.w3.org/2000/svg"
                width="48"
                height="48"
                viewBox="0 0 24 24"
                fill="none"
                stroke="var(--text-color)"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><polygon points="3 6 9 3 15 6 21 3 21 18 15 21 9 18 3 21"
                ></polygon><line x1="9" y1="3" x2="9" y2="18"></line><line
                    x1="15"
                    y1="6"
                    x2="15"
                    y2="21"
                ></line></svg
            >
            <h2 style="margin: 0;">Fejlesztés alatt</h2>
            <p style="margin: 0; color: var(--text-faint);">
                Ide kerül az OpenStreetMap (OSM) TileServer integráció.
            </p>
        </div>
    </div>
</div>
