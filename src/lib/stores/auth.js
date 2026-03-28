import { writable } from "svelte/store";

function createAuthStore() {
    const { subscribe, set } = writable({ loggedIn: false, user: "", isAdmin: false });

    return {
        subscribe,
        init() {
            if (typeof window !== "undefined") {
                set({
                    loggedIn: localStorage.getItem("admin_auth") === "true",
                    user: localStorage.getItem("admin_user") || "",
                    isAdmin: localStorage.getItem("admin_is_admin") === "true",
                });
            }
        },
        login(user, isAdmin = false) {
            if (typeof window !== "undefined") {
                localStorage.setItem("admin_auth", "true");
                localStorage.setItem("admin_user", user || "User");
                localStorage.setItem("admin_is_admin", isAdmin ? "true" : "false");
                set({ loggedIn: true, user: user || "User", isAdmin });
            }
        },
        logout() {
            if (typeof window !== "undefined") {
                localStorage.removeItem("admin_auth");
                localStorage.removeItem("admin_user");
                localStorage.removeItem("admin_is_admin");
                set({ loggedIn: false, user: "", isAdmin: false });
            }
        },
    };
}

export const auth = createAuthStore();
