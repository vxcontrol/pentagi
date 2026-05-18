import { useState } from 'react';

const APP_NAME = 'PentAGI';

interface PageTitleProps {
    /**
     * Page-specific prefix. Skipped when nullish/empty so loading states can
     * render `<PageTitle>{flow?.title}</PageTitle>` without temporarily
     * setting the tab title to "— PentAGI".
     *
     * While this component is mounted (i.e. the user stays on the same route),
     * the last non-empty value is "sticky" — a brief loading state (e.g. when
     * DetailNavigation switches between siblings and the next document is in
     * flight) keeps showing the previous title rather than flashing the
     * generic "Provider/Knowledge/Template — PentAGI" fallback. State is
     * tied to the component instance so it resets cleanly on route change.
     */
    children?: null | string;
}

/**
 * Renders a `<title>` element that React 19 hoists into <head> automatically.
 * Each page-level route component drops one of these at the top of its JSX so
 * the browser tab, history, and shareable links reflect the actual page —
 * instead of the static "PentAGI" coming from index.html.
 */
export function PageTitle({ children }: PageTitleProps) {
    const [sticky, setSticky] = useState<null | string>(children ?? null);

    // React 19 supports setState during render for derived state; it triggers
    // a synchronous re-render before commit and avoids a one-frame flicker.
    if (children && children !== sticky) {
        setSticky(children);
    }

    const effective = children || sticky;

    return <title>{effective ? `${effective} — ${APP_NAME}` : APP_NAME}</title>;
}
