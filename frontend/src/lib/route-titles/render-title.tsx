export const APP_NAME = 'PentAGI';

export type RouteParams = Record<string, string | undefined>;

// React 19 hoists <title> into <head> automatically. The child must be a
// single string of text (template literals are fine — they collapse to one
// string at render time). An empty label falls back to APP_NAME alone.
export const renderTitle = (label: null | string) => <title>{label ? `${label} — ${APP_NAME}` : APP_NAME}</title>;
