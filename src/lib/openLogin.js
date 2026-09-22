const OPEN_LOGIN_EVENT = "lamsza:open-login";

/** Opens the public layout Google sign-in dialog. */
export function openLogin() {
    if (typeof window !== "undefined") {
        window.dispatchEvent(new CustomEvent(OPEN_LOGIN_EVENT));
    }
}

/** @param {() => void} openDialog */
export function listenForOpenLogin(openDialog) {
    if (typeof window === "undefined") return () => {};
    window.addEventListener(OPEN_LOGIN_EVENT, openDialog);
    return () => window.removeEventListener(OPEN_LOGIN_EVENT, openDialog);
}
