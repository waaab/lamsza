<script>
    import { weatherIconEmoji } from "$lib/utils";

    /** @type {string} OWM-style icon code (e.g. "01d", "02n") */
    export let code = "01d";
    /** "emoji" | "svg" */
    export let style = "emoji";
    /** Optional time for day/night (SVG uses same logic as emoji) */
    export let now = new Date();

    $: hour = now.getHours();
    $: isNight = hour >= 19 || hour < 6;
    $: normalized = (code && isNight && code.endsWith("d")) ? code.slice(0, -1) + "n" : code;
    $: displayCode = normalized || code;
</script>

{#if style === "svg"}
    <span
        class="weather-icon-svg weather-icon-svg--{displayCode}"
        data-weather-code={displayCode}
        aria-hidden="true"
        role="img"
    >
        <!-- Shared: stroke round, consistent weight; clouds use smooth bezier puffs -->
        {#if displayCode === "01d"}
            <!-- Clear day: soft sun with 8 rays -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <circle cx="32" cy="32" r="9" fill="currentColor" opacity="0.95" />
                <g stroke="currentColor" stroke-width="2">
                    <line x1="32" y1="6" x2="32" y2="14" />
                    <line x1="32" y1="50" x2="32" y2="58" />
                    <line x1="6" y1="32" x2="14" y2="32" />
                    <line x1="50" y1="32" x2="58" y2="32" />
                    <line x1="11" y1="11" x2="18" y2="18" />
                    <line x1="46" y1="46" x2="53" y2="53" />
                    <line x1="11" y1="53" x2="18" y2="46" />
                    <line x1="46" y1="18" x2="53" y2="11" />
                </g>
            </svg>
        {:else if displayCode === "01n"}
            <!-- Clear night: moon -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 24 24" fill="currentColor">
                <path d="M12 11.807A9.002 9.002 0 0 1 10.049 2a9.942 9.942 0 0 0-5.12 2.735c-3.905 3.905-3.905 10.237 0 14.142 3.906 3.906 10.237 3.905 14.143 0a9.946 9.946 0 0 0 2.735-5.119A9.003 9.003 0 0 1 12 11.807z"/>
            </svg>
        {:else if displayCode === "02d"}
            <!-- Few clouds day: sun + cloud -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" stroke-miterlimit="10">
                <circle cx="50" cy="22" r="5" fill="currentColor" opacity="0.95" />
                <line x1="50" y1="14" x2="50" y2="17" />
                <line x1="50" y1="27" x2="50" y2="30" />
                <line x1="42" y1="22" x2="45" y2="22" />
                <line x1="55" y1="22" x2="58" y2="22" />
                <line x1="47" y1="17" x2="48" y2="19" />
                <line x1="53" y1="17" x2="52" y2="19" />
                <line x1="47" y1="27" x2="48" y2="25" />
                <line x1="53" y1="27" x2="52" y2="25" />
                <g transform="translate(2, 22) scale(1.05)">
                    <path fill="none" stroke="currentColor" stroke-width="2" stroke-miterlimit="10" d="M41,50h14c4.565,0,8-3.582,8-8s-3.435-8-8-8 c0-11.046-9.52-20-20.934-20C23.966,14,14.8,20.732,13,30c0,0-0.831,0-1.667,0C5.626,30,1,34.477,1,40s4.293,10,10,10H41"/>
                </g>
            </svg>
        {:else if displayCode === "02n"}
            <!-- Few clouds night: moon + cloud -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" stroke-miterlimit="10">
                <g transform="translate(40, 8)" fill="currentColor">
                    <path d="M12 11.807A9.002 9.002 0 0 1 10.049 2a9.942 9.942 0 0 0-5.12 2.735c-3.905 3.905-3.905 10.237 0 14.142 3.906 3.906 10.237 3.905 14.143 0a9.946 9.946 0 0 0 2.735-5.119A9.003 9.003 0 0 1 12 11.807z"/>
                </g>
                <g transform="translate(2, 22) scale(1.05)">
                    <path fill="none" stroke="currentColor" stroke-width="2" stroke-miterlimit="10" d="M41,50h14c4.565,0,8-3.582,8-8s-3.435-8-8-8 c0-11.046-9.52-20-20.934-20C23.966,14,14.8,20.732,13,30c0,0-0.831,0-1.667,0C5.626,30,1,34.477,1,40s4.293,10,10,10H41"/>
                </g>
            </svg>
        {:else if displayCode === "03d" || displayCode === "03n"}
            <!-- Scattered clouds: two clouds -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" stroke-miterlimit="10">
                <g transform="translate(-6, 22) scale(0.75)">
                    <path fill="none" stroke="currentColor" stroke-width="2" stroke-miterlimit="10" d="M41,50h14c4.565,0,8-3.582,8-8s-3.435-8-8-8 c0-11.046-9.52-20-20.934-20C23.966,14,14.8,20.732,13,30c0,0-0.831,0-1.667,0C5.626,30,1,34.477,1,40s4.293,10,10,10H41"/>
                </g>
                <g transform="translate(20, 6) scale(0.55)">
                    <path fill="none" stroke="currentColor" stroke-width="2" stroke-miterlimit="10" d="M41,50h14c4.565,0,8-3.582,8-8s-3.435-8-8-8 c0-11.046-9.52-20-20.934-20C23.966,14,14.8,20.732,13,30c0,0-0.831,0-1.667,0C5.626,30,1,34.477,1,40s4.293,10,10,10H41"/>
                </g>
            </svg>
        {:else if displayCode === "04d" || displayCode === "04n"}
            <!-- Overcast: one cloud -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 56 56" fill="none" stroke="currentColor" stroke-width="2" stroke-miterlimit="10">
                <path fill="none" stroke="currentColor" stroke-width="2" stroke-miterlimit="10" d="M41,50h14c4.565,0,8-3.582,8-8s-3.435-8-8-8 c0-11.046-9.52-20-20.934-20C23.966,14,14.8,20.732,13,30c0,0-0.831,0-1.667,0C5.626,30,1,34.477,1,40s4.293,10,10,10H41"/>
            </svg>
        {:else if displayCode === "09d" || displayCode === "09n"}
            <!-- Shower rain: cloud + three rain drops -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" stroke-miterlimit="10">
                <g transform="translate(2, 18) scale(1.05)">
                    <path fill="none" stroke="currentColor" stroke-width="2" stroke-miterlimit="10" d="M41,50h14c4.565,0,8-3.582,8-8s-3.435-8-8-8 c0-11.046-9.52-20-20.934-20C23.966,14,14.8,20.732,13,30c0,0-0.831,0-1.667,0C5.626,30,1,34.477,1,40s4.293,10,10,10H41"/>
                </g>
                <path d="M26 44 L24 56 M38 42 L36 54 M50 46 L48 58" />
            </svg>
        {:else if displayCode === "10d" || displayCode === "10n"}
            <!-- Rain: cloud + four rain lines -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" stroke-miterlimit="10">
                <g transform="translate(2, 18) scale(1.05)">
                    <path fill="none" stroke="currentColor" stroke-width="2" stroke-miterlimit="10" d="M41,50h14c4.565,0,8-3.582,8-8s-3.435-8-8-8 c0-11.046-9.52-20-20.934-20C23.966,14,14.8,20.732,13,30c0,0-0.831,0-1.667,0C5.626,30,1,34.477,1,40s4.293,10,10,10H41"/>
                </g>
                <path d="M22 44 L18 58 M34 42 L30 56 M46 44 L42 58 M58 42 L54 56" />
            </svg>
        {:else if displayCode === "11d" || displayCode === "11n"}
            <!-- Thunderstorm: cloud + lightning bolt -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" stroke-miterlimit="10">
                <g transform="translate(2, 18) scale(1.05)">
                    <path fill="none" stroke="currentColor" stroke-width="2" stroke-miterlimit="10" d="M41,50h14c4.565,0,8-3.582,8-8s-3.435-8-8-8 c0-11.046-9.52-20-20.934-20C23.966,14,14.8,20.732,13,30c0,0-0.831,0-1.667,0C5.626,30,1,34.477,1,40s4.293,10,10,10H41"/>
                </g>
                <path d="M38 32 L30 44 L36 44 L26 56 L42 40 L34 40 Z" fill="currentColor" stroke="none" />
            </svg>
        {:else if displayCode === "13d" || displayCode === "13n"}
            <!-- Snow: cloud + 6-point snowflake -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" stroke-miterlimit="10">
                <g transform="translate(2, 18) scale(1.05)">
                    <path fill="none" stroke="currentColor" stroke-width="2" stroke-miterlimit="10" d="M41,50h14c4.565,0,8-3.582,8-8s-3.435-8-8-8 c0-11.046-9.52-20-20.934-20C23.966,14,14.8,20.732,13,30c0,0-0.831,0-1.667,0C5.626,30,1,34.477,1,40s4.293,10,10,10H41"/>
                </g>
                <line x1="32" y1="44" x2="32" y2="58" />
                <line x1="26" y1="47" x2="38" y2="55" />
                <line x1="38" y1="47" x2="26" y2="55" />
                <line x1="28" y1="58" x2="36" y2="44" />
                <line x1="36" y1="58" x2="28" y2="44" />
                <line x1="26" y1="51" x2="38" y2="51" />
            </svg>
        {:else if displayCode === "50d" || displayCode === "50n"}
            <!-- Mist: soft horizontal bands (varying length for depth) -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                <line x1="8" y1="20" x2="56" y2="20" />
                <line x1="12" y1="30" x2="52" y2="30" />
                <line x1="6" y1="40" x2="58" y2="40" />
                <line x1="10" y1="50" x2="54" y2="50" />
            </svg>
        {:else}
            <!-- Fallback: sun -->
            <svg class="weather-icon-svg__graphic" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
                <circle cx="32" cy="32" r="9" fill="currentColor" opacity="0.95" />
                <line x1="32" y1="8" x2="32" y2="16" />
                <line x1="32" y1="48" x2="32" y2="56" />
                <line x1="8" y1="32" x2="16" y2="32" />
                <line x1="48" y1="32" x2="56" y2="32" />
            </svg>
        {/if}
    </span>
{:else}
    <span class="weather-icon-emoji" aria-hidden="true">{weatherIconEmoji(code, now)}</span>
{/if}

<style>
    .weather-icon-svg {
        display: inline-block;
        line-height: 0;
        vertical-align: middle;
    }
    .weather-icon-svg--01d,
    .weather-icon-svg--02d {
        color: #f0b429;
    }
    .weather-icon-svg svg {
        display: block;
        width: 5rem;
        height: 5rem;
    }
    .weather-icon-emoji {
        display: inline-block;
        line-height: 1;
    }
</style>
