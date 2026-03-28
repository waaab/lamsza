<script>
    import Markdown from "./Markdown.svelte";
    import { getApiBase } from "$lib/api.js";

    /** Logical section key — must match `page_faq_sections.section_key` */
    export let sectionKey = "home";

    let faqTitle = "";
    /** @type {{ question: string; answer: string }[]} */
    let items = [];
    let disclaimerMd = "";
    let loading = true;
    let loadError = false;
    let loadSeq = 0;

    async function loadSection(key) {
        if (!key) return;
        const seq = ++loadSeq;
        loading = true;
        loadError = false;
        const base = getApiBase();
        try {
            const res = await fetch(
                `${base}/api/page_faq?section=${encodeURIComponent(key)}`,
            );
            if (!res.ok) {
                if (seq === loadSeq) loadError = true;
                return;
            }
            const data = await res.json();
            if (seq !== loadSeq) return;
            faqTitle = data.faq_title || "";
            items = Array.isArray(data.faq_items) ? data.faq_items : [];
            disclaimerMd = data.disclaimer_markdown || "";
        } catch (e) {
            console.error("PageFaqDisclaimer:", e);
            if (seq === loadSeq) loadError = true;
        } finally {
            if (seq === loadSeq) loading = false;
        }
    }

    $: loadSection(sectionKey);

    /** FAQ block renders only when at least one row has a non-empty question */
    $: definedFaqItems = (items || []).filter(
        (it) => (it.question || "").trim().length > 0,
    );
    $: hasFaqContent = definedFaqItems.length > 0;
    $: hasDisclaimer = (disclaimerMd || "").trim().length > 0;
</script>

{#if !loading && !loadError}
    {#if hasFaqContent}
        <section class="faq" id="gyik" aria-label="Gyakori kérdések">
            {#if faqTitle}
                <h2 class="faq-title">{faqTitle}</h2>
            {/if}
            <div class="faq-list">
                {#each definedFaqItems as it, i (sectionKey + "-" + i + "-" + it.question)}
                    <details class="faq-item" open>
                        <summary>{it.question}</summary>
                        {#if (it.answer || "").trim()}
                            <div class="faq-item-body">
                                <Markdown source={it.answer} />
                            </div>
                        {/if}
                    </details>
                {/each}
            </div>
        </section>
    {/if}

    {#if hasDisclaimer}
        <section id="disclaimer" aria-label="Felelősségkizárás">
            <div class="note info">
                <Markdown source={disclaimerMd.trim()} />
            </div>
        </section>
    {/if}
{/if}

<style>
    .faq-item-body :global(.markdown-content p) {
        font-size: 0.8rem;
        color: var(--text-secondary);
        margin: 0.65rem 0 0;
        line-height: 1.6;
    }
    .faq-item-body :global(.markdown-content ul),
    .faq-item-body :global(.markdown-content ol) {
        margin: 0.65rem 0 0 0.25rem;
        padding: 0 0 0 1rem;
        color: var(--text-secondary);
        font-size: 0.8rem;
    }
    .faq-item-body :global(.markdown-content li) {
        line-height: 1.6;
        margin-bottom: 0.3rem;
    }
</style>
