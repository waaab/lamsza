import { redirect } from "@sveltejs/kit";

/** Legacy singular path → plural. */
export function load() {
    throw redirect(301, "/szekek");
}
