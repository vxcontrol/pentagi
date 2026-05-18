import type { ComponentType } from 'react';

import { useMatches } from 'react-router-dom';

const APP_NAME = 'PentAGI';

export interface RouteHandleWithTitle {
    /**
     * Static string or a pure function of route params. Use for routes
     * where the title is known synchronously — listing pages, /new routes,
     * params-derived titles like `/settings/prompts/:promptId`.
     */
    title?: TitleResolver;
    /**
     * Component that renders `<title>` and may subscribe to Apollo cache or
     * other reactive sources. Use when the title depends on resource data
     * that needs to react to cache updates after the page mounts.
     */
    titleComponent?: TitleComponent;
}

type RouteParams = Record<string, string | undefined>;

type TitleComponent = ComponentType<{ params: RouteParams }>;

type TitleResolver = ((params: RouteParams) => string) | string;

const hasTitle = (handle: unknown): handle is { title: TitleResolver } => {
    if (typeof handle !== 'object' || handle === null || !('title' in handle)) {
        return false;
    }

    const value = (handle as { title: unknown }).title;

    return typeof value === 'string' || typeof value === 'function';
};

const hasTitleComponent = (handle: unknown): handle is { titleComponent: TitleComponent } =>
    typeof handle === 'object' &&
    handle !== null &&
    'titleComponent' in handle &&
    typeof (handle as { titleComponent: unknown }).titleComponent === 'function';

const renderTitle = (label: null | string) => <title>{label ? `${label} — ${APP_NAME}` : APP_NAME}</title>;

/**
 * Renders the document `<title>` driven by react-router route handles. The
 * deepest match exposing `handle.title` or `handle.titleComponent` wins.
 *
 * - `handle.title: string` — static (e.g. "Dashboard").
 * - `handle.title: (params) => string` — derived from route params
 *   (e.g. `/settings/prompts/:promptId` formatting the id).
 * - `handle.titleComponent` — reactive component that subscribes to Apollo
 *   cache for resource-driven titles that need to react to cache updates.
 *
 * Living in the app shell, this component survives navigation between
 * sibling detail routes — fixing the previous "Provider — PentAGI" flash
 * during DetailNavigation prev/next.
 */
export function DocumentTitle() {
    const matches = useMatches();
    const match = matches.findLast((m) => hasTitleComponent(m.handle) || hasTitle(m.handle));

    if (!match) {
        return renderTitle(null);
    }

    const handle = match.handle;

    if (hasTitleComponent(handle)) {
        const TitleComponent = handle.titleComponent;

        return <TitleComponent params={match.params} />;
    }

    if (hasTitle(handle)) {
        const resolved = typeof handle.title === 'function' ? handle.title(match.params) : handle.title;

        return renderTitle(resolved);
    }

    return renderTitle(null);
}
