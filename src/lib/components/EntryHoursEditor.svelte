<script>
    import { WEEKDAYS, emptyWeekHours, normalizeHours } from "$lib/entryHours.js";

    /** @type {{ hours?: Record<string, { open: string, close: string, closed: boolean }> }} */
    let { hours = $bindable(emptyWeekHours()) } = $props();

    let rows = $derived(
        WEEKDAYS.map(({ key, label }) => {
            const week = normalizeHours(hours);
            return { key, label, slot: week[key] };
        }),
    );

    function patchDay(key, patch) {
        const week = normalizeHours(hours);
        hours = { ...week, [key]: { ...week[key], ...patch } };
    }
</script>

<div class="entry-hours-editor">
    <table class="entry-hours-editor__table">
        <thead>
            <tr>
                <th>Nap</th>
                <th>Zárva</th>
                <th>Nyitás</th>
                <th>Zárás</th>
            </tr>
        </thead>
        <tbody>
            {#each rows as row (row.key)}
                <tr>
                    <th scope="row">{row.label}</th>
                    <td>
                        <label class="entry-hours-editor__closed">
                            <input
                                type="checkbox"
                                checked={row.slot.closed}
                                onchange={(event) =>
                                    patchDay(row.key, {
                                        closed: event.currentTarget.checked,
                                    })}
                            />
                            <span class="sr-only">Zárva</span>
                        </label>
                    </td>
                    <td>
                        <input
                            type="time"
                            value={row.slot.open}
                            disabled={row.slot.closed}
                            oninput={(event) =>
                                patchDay(row.key, { open: event.currentTarget.value })}
                        />
                    </td>
                    <td>
                        <input
                            type="time"
                            value={row.slot.close}
                            disabled={row.slot.closed}
                            oninput={(event) =>
                                patchDay(row.key, { close: event.currentTarget.value })}
                        />
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
</div>

<style>
    .entry-hours-editor {
        overflow-x: auto;
        border: 1px solid var(--border-color);
        border-radius: 8px;
        background: var(--bg-body, var(--card-bg));
    }
    .entry-hours-editor__table {
        width: 100%;
        border-collapse: collapse;
        font-size: var(--text-sm);
    }
    .entry-hours-editor__table th,
    .entry-hours-editor__table td {
        padding: 0.4rem 0.55rem;
        border-bottom: 1px solid var(--border-color);
        text-align: left;
        vertical-align: middle;
    }
    .entry-hours-editor__table thead th {
        font-weight: 600;
        color: var(--text-secondary);
    }
    .entry-hours-editor__table tbody th {
        font-weight: 500;
        white-space: nowrap;
    }
    .entry-hours-editor__table tbody tr:last-child th,
    .entry-hours-editor__table tbody tr:last-child td {
        border-bottom: none;
    }
    .entry-hours-editor__table input[type="time"] {
        width: 100%;
        min-width: 6.5rem;
        margin: 0;
    }
    .entry-hours-editor__closed {
        display: inline-flex;
        align-items: center;
        margin: 0;
        font-weight: 400;
    }
    .entry-hours-editor__closed input {
        width: auto;
        margin: 0;
    }
</style>
