const APP_NAME = 'PentAGI';

interface PageTitleProps {
    /**
     * Page-specific prefix. Skipped when nullish/empty so loading states can
     * render `<PageTitle>{flow?.title}</PageTitle>` without temporarily
     * setting the tab title to "— PentAGI".
     */
    children?: null | string;
}

/**
 * Renders a `<title>` element that React 19 hoists into <head> automatically.
 * Each page-level route component drops one of these at the top of its JSX so
 * the browser tab, history, and shareable links reflect the actual page —
 * instead of the static "PentAGI" coming from index.html.
 */
export const PageTitle = ({ children }: PageTitleProps) => (
    <title>{children ? `${children} — ${APP_NAME}` : APP_NAME}</title>
);
