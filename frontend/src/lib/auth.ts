export function isAuthenticated() {
    const auth = localStorage.getItem('auth');

    if (!auth) {
        return false;
    }

    const data = JSON.parse(auth);
    const expired = data.expires_at;
    const now = new Date().toISOString();

    return expired > now;
}
