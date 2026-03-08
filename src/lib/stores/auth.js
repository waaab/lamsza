import { writable } from "svelte/store";

function createAuthStore() {
    const { subscribe, set } = writable({ loggedIn: false, user: "" });

    return {
        subscribe,
        init() {
            if (typeof window !== "undefined") {
                set({
                    loggedIn: localStorage.getItem("admin_auth") === "true",
                    user: localStorage.getItem("admin_user") || "Admin",
                });
            }
        },
        login() {
            if (typeof window !== "undefined") {
                localStorage.setItem("admin_auth", "true");
                localStorage.setItem("admin_user", "Admin");
                set({ loggedIn: true, user: "Admin" });
            }
        },
        logout() {
            if (typeof window !== "undefined") {
                localStorage.removeItem("admin_auth");
                localStorage.removeItem("admin_user");
                set({ loggedIn: false, user: "" });
            }
        },
    };
}

export const auth = createAuthStore();
