import { redirect } from "@sveltejs/kit";

/** Legacy /szek/{slug} → /szekek/{slug}. */
export function load({ params }) {
    throw redirect(301, `/szekek/${params.slug}`);
}
