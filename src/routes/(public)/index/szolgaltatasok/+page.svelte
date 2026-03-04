<script>
    import { onMount } from "svelte";

    let services = [];
    let loading = true;
    let error = null;

    // Filter to retain explicit Service categories
    const retainedCategories = [
        "mesteremberek",
        "egészségügy",
        "oktatás",
        "hivatalok",
    ];

    onMount(async () => {
        try {
            const apiBase =
                import.meta.env.VITE_API_BASE_URL || "http://localhost:3000";
            const res = await fetch(`${apiBase}/api/directory`);
            if (!res.ok) throw new Error("Hiba a címtár betöltésekor");

            const allServices = (await res.json()) || [];

            // Filter down exclusively to the legacy services
            services = allServices.filter(
                (s) =>
                    s.category &&
                    retainedCategories.includes(s.category.toLowerCase()),
            );
        } catch (err) {
            console.error(err);
            error = "Nem sikerült betölteni a szolgáltatásokat.";
        } finally {
            loading = false;
        }
    });
</script>

<svelte:head>
    <title>Szolgáltatások - Index</title>
</svelte:head>

<div class="container" style="min-height: calc(100vh - 120px)">
    <div class="breadcrumbs">
        <a href="/">Főoldal</a> &rsaquo;
        <a href="/index">Index</a> &rsaquo;
        <span class="active">Kiemelt Szolgáltatások</span>
    </div>

    <h1 class="page-title">Szolgáltatások</h1>
    <p class="greeting">
        Kiemelt helyi szolgáltatások, mesteremberek és egészségügyi intézmények
        indexe.
    </p>

    {#if loading}
        <div class="list grid">
            {#each Array(6) as _}
                <article
                    class="card service--skeleton"
                    style="height: 150px; display: flex; flex-direction: column; padding: 1rem; gap: 0.5rem;"
                >
                    <div
                        class="skeleton skeleton-text"
                        style="width: 30%;"
                    ></div>
                    <div
                        class="skeleton skeleton-text"
                        style="width: 80%; margin-top: 0.5rem; height: 1.2rem;"
                    ></div>
                    <div
                        class="skeleton skeleton-text"
                        style="width: 60%; margin-top: auto;"
                    ></div>
                </article>
            {/each}
        </div>
    {:else if error}
        <div class="error-msg">{error}</div>
    {:else if services.length === 0}
        <div class="error-msg">
            Jelenleg nincs listázott szolgáltatás ebben a kategóriában.
        </div>
    {:else}
        <div class="list grid">
            {#each services as service}
                <article
                    class="card service"
                    style="cursor: pointer; transition: transform 0.2s, box-shadow 0.2s;"
                >
                    <span
                        class="badge"
                        style="margin-bottom: 0.5rem; display: inline-block;"
                        >{service.category}</span
                    >
                    <h3 class="service-name">
                        <a
                            href="/bejegyzes/{service.slug}"
                            style="color:inherit;text-decoration:none;"
                            >{service.name}</a
                        >
                    </h3>
                    {#if service.url}
                        <div
                            class="service-info"
                            style="margin-bottom: 0.5rem;"
                        >
                            <span
                                style="color: var(--text-faint); margin-right: 0.3rem;"
                                >🔗</span
                            >
                            <a
                                href={service.url}
                                target="_blank"
                                rel="nofollow noopener"
                                style="color: var(--primary-color); text-decoration: none;"
                                >{service.url}</a
                            >
                        </div>
                    {/if}
                    <div class="service-info">
                        📍 {[
                            service.location,
                            service.location_ro,
                            service.location_de,
                        ]
                            .filter(Boolean)
                            .join(" | ")} - {service.address}
                    </div>
                    {#if service.phone}
                        <div
                            class="service-info"
                            style="color: var(--text-faint); margin-top: 0.25rem;"
                        >
                            📞 {service.phone}
                        </div>
                    {/if}
                </article>
            {/each}
        </div>
    {/if}
</div>
