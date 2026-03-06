<script>
    export let title;
    export let open = false;
</script>

<details class="faq-item" {open}>
    <summary>{title}</summary>
    <div class="accordion-content">
        <slot />
    </div>
</details>

<style>
    .accordion-content {
        padding: 0.5rem 1rem 1rem;
    }
</style>
