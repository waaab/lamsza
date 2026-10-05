<script>
    /**
     * One-line filter row: label on the left, chips that drag and scroll,
     * arrows when the row overflows. Same behaviour as the news topics.
     */
    /** @type {{ label?: string, children?: import('svelte').Snippet }} */
    let { label = "", children } = $props();

    let canScrollLeft = $state(false);
    let canScrollRight = $state(false);
    /** @type {HTMLElement | null} */
    let scroller = null;

    /**
     * @param {HTMLElement} node
     */
    function checkScroll(node) {
        if (!node) return;
        canScrollLeft = node.scrollLeft > 5;
        canScrollRight = node.scrollLeft < node.scrollWidth - node.clientWidth - 5;
    }

    /**
     * @param {HTMLElement} node
     */
    function dragScroll(node) {
        let isDown = false;
        let dragged = false;
        let startX = 0;
        let scrollLeft = 0;

        const onMouseDown = (/** @type {MouseEvent} */ e) => {
            isDown = true;
            dragged = false;
            node.classList.add("active");
            startX = e.pageX - node.offsetLeft;
            scrollLeft = node.scrollLeft;
        };

        const onMouseLeave = () => {
            isDown = false;
            node.classList.remove("active");
        };

        const onMouseUp = () => {
            isDown = false;
            node.classList.remove("active");
        };

        const onMouseMove = (/** @type {MouseEvent} */ e) => {
            if (!isDown) return;
            e.preventDefault();
            const x = e.pageX - node.offsetLeft;
            const walk = (x - startX) * 2;
            if (Math.abs(walk) > 4) dragged = true;
            node.scrollLeft = scrollLeft - walk;
            checkScroll(node);
        };

        const onClick = (/** @type {MouseEvent} */ e) => {
            if (!dragged) return;
            e.preventDefault();
            e.stopPropagation();
            dragged = false;
        };

        const onScroll = () => checkScroll(node);
        const onResize = () => checkScroll(node);
        const observer = new MutationObserver(() => checkScroll(node));

        scroller = node;
        node.addEventListener("mousedown", onMouseDown);
        node.addEventListener("mouseleave", onMouseLeave);
        node.addEventListener("mouseup", onMouseUp);
        node.addEventListener("mousemove", onMouseMove);
        node.addEventListener("click", onClick, true);
        node.addEventListener("scroll", onScroll);
        window.addEventListener("resize", onResize);
        observer.observe(node, { childList: true });
        setTimeout(() => checkScroll(node), 100);

        return () => {
            if (scroller === node) scroller = null;
            node.removeEventListener("mousedown", onMouseDown);
            node.removeEventListener("mouseleave", onMouseLeave);
            node.removeEventListener("mouseup", onMouseUp);
            node.removeEventListener("mousemove", onMouseMove);
            node.removeEventListener("click", onClick, true);
            node.removeEventListener("scroll", onScroll);
            window.removeEventListener("resize", onResize);
            observer.disconnect();
        };
    }

    function scrollChips(amount) {
        scroller?.scrollBy({ left: amount, behavior: "smooth" });
    }
</script>

<div class="header-tabs chips">
    {#if label}
        <span class="header-tabs-label" aria-label={label.replace(/:$/, "")}>{label}</span>
    {/if}
    <div
        class="header-tabs-filters-row"
        class:can-left={canScrollLeft}
        class:can-right={canScrollRight}
    >
        <button
            type="button"
            class="btn btn-xs scroll-arrow left"
            aria-label="Görgetés balra"
            onclick={() => scrollChips(-200)}>‹</button
        >
        <div class="chips-list" {@attach dragScroll}>
            {#if children}
                {@render children()}
            {/if}
        </div>
        <button
            type="button"
            class="btn btn-xs scroll-arrow right"
            aria-label="Görgetés jobbra"
            onclick={() => scrollChips(200)}>›</button
        >
    </div>
</div>
