/**
 * A backend "no rows in result set" / "not found" GraphQL error, as opposed to a real load
 * failure (network, 5xx, cold-cache backend error). Detail pages redirect to the list on the
 * former and render an in-page ErrorState + Retry on the latter — collapsing the two silently
 * bounces the user off a page that a retry would have loaded.
 */
export const isNotFoundError = (error: { message: string }) => {
    // The backend's authz failure "requested permission '<perm>' not found" (graph/context.go) also
    // contains "not found"; classing it as a missing record would silently redirect a user who only
    // lacks access, instead of surfacing the denial.
    if (/\b(permission|privilege|unauthori[sz]ed|not authori[sz]ed|forbidden|access denied)\b/i.test(error.message)) {
        return false;
    }

    return /no rows in result set|not found/i.test(error.message);
};
