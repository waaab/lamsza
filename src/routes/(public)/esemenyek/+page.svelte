<script>
    import { onMount } from "svelte";

    const getBase = () =>
        import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";

    let events = [];
    let loading = true;
    let error = null;

    onMount(async () => {
        try {
            const res = await fetch(`${getBase()}/api/events`);
            if (res.ok) {
                events = await res.json();
            } else {
                error = "Nem sikerült betölteni az eseményeket.";
            }
        } catch (e) {
            console.error(e);
            error = "Hálózati hiba történt.";
        } finally {
            loading = false;
        }
    });

    function formatDate(dateStr) {
        if (!dateStr) return "";
        const d = new Date(dateStr);
        return d.toLocaleDateString("hu-HU", {
            year: "numeric",
            month: "long",
            day: "numeric",
            weekday: "long",
        });
    }

    function formatEventDateTime(ev) {
        if (!ev) return "";
        let res = formatDate(ev.start_date);
        if (ev.start_time) res += ` ${ev.start_time.slice(0, 5)}`;

        if (ev.end_date && ev.end_date !== ev.start_date) {
            res += ` - ${formatDate(ev.end_date)}`;
            if (ev.end_time) res += ` ${ev.end_time.slice(0, 5)}`;
        } else if (ev.end_time) {
            res += ` - ${ev.end_time.slice(0, 5)}`;
        }
        return res;
    }

    const EVENT_TYPE_LABELS = {
        cultural: "Kulturális",
        sports: "Sport",
        festival: "Fesztivál",
        religious: "Vallási",
        other: "Egyéb",
    };
</script>

<svelte:head>
    <title>Esemény Naptár - Székely Gugel</title>
</svelte:head>

<main class="container">
    <h1 class="page-title">Esemény Naptár</h1>
    <p class="greeting">Válogass a legfrissebb székelyföldi események közül.</p>

    {#if loading}
        <span class="info-box"><p>Betöltés...</p></span>
    {:else if error}
        <span class="info-box"><p>{error}</p></span>
    {:else if events.length === 0}
        <span class="info-box"
            ><p>Jelenleg nincsenek meghirdetett események.</p></span
        >
    {:else}
        <div class="events-grid">
            {#each events as event}
                <article class="event-card">
                    <div class="badge event">
                        {EVENT_TYPE_LABELS[event.event_type] ||
                            event.event_type}
                    </div>
                    <h2 class="event-title">{event.title}</h2>

                    <div class="event-meta">
                        <span class="event-date">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                ><rect
                                    x="3"
                                    y="4"
                                    width="18"
                                    height="18"
                                    rx="2"
                                    ry="2"
                                ></rect><line x1="16" y1="2" x2="16" y2="6"
                                ></line><line x1="8" y1="2" x2="8" y2="6"
                                ></line><line x1="3" y1="10" x2="21" y2="10"
                                ></line></svg
                            >
                            {formatEventDateTime(event)}
                        </span>

                        <span class="event-location">
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                ><path
                                    d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"
                                ></path><circle cx="12" cy="10" r="3"
                                ></circle></svg
                            >
                            <a
                                href="/{event.county_slug}-megye/{event.location_slug}"
                                >{event.location_name}</a
                            >,
                            <a
                                href="/{event.county_slug}-megye"
                                class="county-link">{event.county} megye</a
                            >
                        </span>
                    </div>

                    {#if event.description}
                        <p class="event-description">{event.description}</p>
                    {/if}

                    {#if event.organizer}
                        <div class="event-organizer">
                            <strong>Szervező:</strong>
                            {event.organizer}
                        </div>
                    {/if}
                </article>
            {/each}
        </div>
    {/if}
</main>

<style>
    .events-grid {
        display: grid;
        grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
        gap: 2rem;
        margin-top: 2rem;
    }

    .event-card {
        background: var(--card-bg);
        border: 1px solid var(--border-color);
        border-radius: 12px;
        padding: 1.5rem;
        position: relative;
        transition:
            transform 0.2s ease,
            box-shadow 0.2s ease;
        display: flex;
        flex-direction: column;
        gap: 1rem;
    }

    .event-card:hover {
        transform: translateY(-4px);
        box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
    }

    .badge.event {
        position: absolute;
        top: 1rem;
        right: 1rem;
    }

    .event-title {
        margin: 0;
        font-size: 1.4rem;
        color: var(--text-color);
        padding-right: 4rem;
    }

    .event-meta {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        font-size: 0.9rem;
        color: var(--text-faint);
    }

    .event-meta a {
        color: var(--primary-color);
        text-decoration: none;
    }

    .event-meta a:hover {
        text-decoration: underline;
    }

    .event-description {
        margin: 0;
        line-height: 1.6;
        color: var(--text-color);
        display: -webkit-box;
        -webkit-line-clamp: 3;
        line-clamp: 3;
        -webkit-box-orient: vertical;
        overflow: hidden;
    }

    .event-organizer {
        font-size: 0.85rem;
        color: var(--text-faint);
        margin-top: auto;
        padding-top: 1rem;
        border-top: 1px solid var(--border-color);
    }

    @media (max-width: 600px) {
        .events-grid {
            grid-template-columns: 1fr;
        }
    }
</style>
