/**
 * A backend "no rows in result set" / "not found" GraphQL error, as opposed to a real load
 * failure (network, 5xx, cold-cache backend error). Detail pages redirect to the list on the
 * former and render an in-page ErrorState + Retry on the latter — collapsing the two silently
 * bounces the user off a page that a retry would have loaded.
 */
export const isNotFoundError = (error: { message: string }) => /no rows in result set|not found/i.test(error.message);
