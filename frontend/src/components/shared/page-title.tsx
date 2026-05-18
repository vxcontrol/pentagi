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
 *
 * Detail routes (templates/:id, knowledges/:id, providers/:id, flows/:id)
 * use route handles via `<DocumentTitle/>` instead — that lets the shell
 * own the title and survives navigation between sibling documents. This
 * helper is for listing/static pages where in-route flicker isn't a
 * concern.
 */
export function PageTitle({ children }: PageTitleProps) {
    return <title>{children ? `${children} — ${APP_NAME}` : APP_NAME}</title>;
}
