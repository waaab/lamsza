<script>
    import { onMount, onDestroy } from "svelte";
    import {
        getEventStatus,
        EVENT_STATUS_LABELS,
        getReferenceNow,
    } from "$lib/eventStatus";

    /** Event row from /api/events or detail */
    export let event;
    /** Refresh so status matches the live clock / time progression */
    export let live = true;
    /** Top-right of card header: no left margin, don’t shrink */
    export let corner = false;

    let now = getReferenceNow();
    let interval;

    function tick() {
        now = getReferenceNow();
    }

    onMount(() => {
        tick();
        if (live) {
            interval = setInterval(tick, 30000);
        }
    });

    onDestroy(() => {
        if (interval) clearInterval(interval);
    });

    $: status = getEventStatus(event, now);
</script>

<span
    class="event-date-badge event-date-badge--{status}"
    class:event-date-badge--corner={corner}
    title={EVENT_STATUS_LABELS[status]}
>
    {EVENT_STATUS_LABELS[status]}
</span>

<style>
    .event-date-badge {
        display: inline-block;
        font-size: 0.65rem;
        font-weight: 700;
        text-transform: uppercase;
        letter-spacing: 0.04em;
        padding: 0.15rem 0.45rem;
        border-radius: 4px;
        vertical-align: middle;
        margin-left: 0.35rem;
        line-height: 1.2;
    }
    .event-date-badge--corner {
        margin-left: 0;
        flex-shrink: 0;
    }
    .event-date-badge--scheduled {
        background: #ede9fe;
        color: #5b21b6;
        border: 1px solid #c4b5fd;
    }
    .event-date-badge--upcoming {
        background: #e0f2fe;
        color: #0369a1;
        border: 1px solid #7dd3fc;
    }
    .event-date-badge--ongoing {
        background: #dcfce7;
        color: #166534;
        border: 1px solid #86efac;
    }
    .event-date-badge--ending_soon {
        background: #fef3c7;
        color: #b45309;
        border: 1px solid #fcd34d;
    }
    .event-date-badge--ended {
        background: #f4f4f5;
        color: #52525b;
        border: 1px solid #d4d4d8;
    }
</style>
