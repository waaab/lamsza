import { redirect } from "@sveltejs/kit";

/** Permanent redirect: /{slug}-szek → /szekek/{slug} (hierarchical URLs). */
export function load({ params }) {
    throw redirect(301, `/szekek/${params.szekSlug}`);
}
