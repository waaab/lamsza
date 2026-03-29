<script>
    import { onMount, onDestroy } from "svelte";
    import { formatHuDateLong } from "$lib/utils";

    let now = new Date();
    let interval;

    function formatDateTime(d) {
        const dateStr = formatHuDateLong(d);
        const timeStr = d.toLocaleTimeString("hu-HU", {
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
        });
        return { dateStr, timeStr };
    }

    // Analog clock: 12 o'clock = -90deg in CSS (top of circle)
    $: secondsAngle = now.getSeconds() * 6 + now.getMilliseconds() * 0.006;
    $: minutesAngle = now.getMinutes() * 6 + now.getSeconds() * 0.1;
    $: hoursAngle = (now.getHours() % 12) * 30 + now.getMinutes() * 0.5;

    onMount(() => {
        now = new Date();
        interval = setInterval(() => {
            now = new Date();
        }, 1000);
    });

    onDestroy(() => {
        if (interval) clearInterval(interval);
    });

    $: formatted = formatDateTime(now);
</script>

<div
    id="datetime"
    class="datetime-card widget"
    data-now-ms={now.getTime()}
>
    <div class="widget-header">
        <h3 class="widget-title">Dátum és idő</h3>
    </div>

    <div class="datetime-clocks">
        <!-- Analog clock -->
        <div class="analog-clock" role="img" aria-label="Analóg óra">
            <svg viewBox="0 0 100 100" class="clock-face">
                <circle class="clock-border" cx="50" cy="50" r="48" />
                <!-- 12, 3, 6, 9 markers -->
                <line x1="50" y1="8" x2="50" y2="14" class="clock-tick major" />
                <line x1="86" y1="50" x2="92" y2="50" class="clock-tick major" />
                <line x1="50" y1="86" x2="50" y2="92" class="clock-tick major" />
                <line x1="8" y1="50" x2="14" y2="50" class="clock-tick major" />
                <!-- Hour hand -->
                <line x1="50" y1="50" x2="50" y2="26" class="clock-hand hour" transform="rotate({hoursAngle} 50 50)" />
                <!-- Minute hand -->
                <line x1="50" y1="50" x2="50" y2="16" class="clock-hand minute" transform="rotate({minutesAngle} 50 50)" />
                <!-- Second hand -->
                <line x1="50" y1="50" x2="50" y2="12" class="clock-hand second" transform="rotate({secondsAngle} 50 50)" />
                <circle cx="50" cy="50" r="3" class="clock-center" />
            </svg>
        </div>

        <!-- Digital clock + date -->
        <div class="digital-block">
            <span class="datetime-digital" class:tick={now.getSeconds() % 2 === 0}>
                {formatted.timeStr}
            </span>
            <span class="datetime-text">{formatted.dateStr}</span>
        </div>
    </div>
</div>

<style>
    .datetime-card {
        display: flex;
        flex-direction: column;
    }

    .datetime-clocks {
        display: flex;
        align-items: center;
        gap: 1.25rem;
        flex-wrap: wrap;
    }

    .analog-clock {
        flex-shrink: 0;
    }

    .clock-face {
        width: 80px;
        height: 80px;
        display: block;
    }

    .clock-border {
        fill: none;
        stroke: var(--border-color, #d1d5db);
        stroke-width: 2;
    }

    .clock-tick.major {
        stroke: var(--text-secondary, #555);
        stroke-width: 2;
        stroke-linecap: round;
    }

    .clock-hand {
        stroke-linecap: round;
    }

    .clock-hand.hour {
        stroke: var(--text-primary, #222);
        stroke-width: 3;
    }

    .clock-hand.minute {
        stroke: var(--text-primary, #222);
        stroke-width: 2;
    }

    .clock-hand.second {
        stroke: var(--szekely-red, #c0392b);
        stroke-width: 1.5;
    }

    .clock-center {
        fill: var(--szekely-red, #c0392b);
    }

    .digital-block {
        display: flex;
        flex-direction: column;
        gap: 0.25rem;
        min-width: 0;
    }

    .datetime-digital {
        font-variant-numeric: tabular-nums;
        font-size: 1.8rem;
        font-weight: 600;
        line-height: 1.1;
        color: var(--text-primary);
        margin: 0;
        letter-spacing: 0.02em;
        transition: opacity 0.15s ease;
    }

    .datetime-digital.tick {
        opacity: 0.88;
    }

    .datetime-text {
        font-size: 0.9rem;
        color: var(--text-secondary);
        margin: 0;
        line-height: 1;
    }
</style>
