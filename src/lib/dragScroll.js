/**
 * Drag to scroll a horizontal row with the mouse or a pen, as the chip rows
 * do (ChipScrollRow). Touch keeps the browser's own swipe, with its momentum.
 * A drag that moved never counts as a click on what is under it.
 *
 * Use as a Svelte attachment: <div {@attach dragScroll}>.
 *
 * @param {HTMLElement} node
 * @returns {() => void}
 */
export function dragScroll(node) {
    let pointerId = -1;
    let startX = 0;
    let startScroll = 0;
    let moved = false;

    const onDown = (/** @type {PointerEvent} */ e) => {
        if (e.pointerType === "touch" || e.button !== 0) return;
        if (node.scrollWidth <= node.clientWidth) return;
        pointerId = e.pointerId;
        startX = e.clientX;
        startScroll = node.scrollLeft;
        moved = false;
    };

    const onMove = (/** @type {PointerEvent} */ e) => {
        if (e.pointerId !== pointerId) return;
        const dx = e.clientX - startX;
        if (!moved && Math.abs(dx) < 4) return;
        if (!moved) {
            moved = true;
            node.setPointerCapture(pointerId);
            node.classList.add("is-dragging");
        }
        e.preventDefault();
        node.scrollLeft = startScroll - dx;
    };

    const onUp = (/** @type {PointerEvent} */ e) => {
        if (e.pointerId !== pointerId) return;
        pointerId = -1;
        node.classList.remove("is-dragging");
        if (node.hasPointerCapture(e.pointerId)) node.releasePointerCapture(e.pointerId);
    };

    const onClick = (/** @type {MouseEvent} */ e) => {
        if (!moved) return;
        e.preventDefault();
        e.stopPropagation();
        moved = false;
    };

    // Keep the browser from starting a native drag of an image or a link.
    const onDragStart = (/** @type {DragEvent} */ e) => e.preventDefault();

    node.classList.add("drag-scroll");
    node.addEventListener("pointerdown", onDown);
    node.addEventListener("pointermove", onMove);
    node.addEventListener("pointerup", onUp);
    node.addEventListener("pointercancel", onUp);
    node.addEventListener("click", onClick, true);
    node.addEventListener("dragstart", onDragStart);

    return () => {
        node.classList.remove("drag-scroll", "is-dragging");
        node.removeEventListener("pointerdown", onDown);
        node.removeEventListener("pointermove", onMove);
        node.removeEventListener("pointerup", onUp);
        node.removeEventListener("pointercancel", onUp);
        node.removeEventListener("click", onClick, true);
        node.removeEventListener("dragstart", onDragStart);
    };
}
