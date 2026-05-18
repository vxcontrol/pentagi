import type { ComponentType } from 'react';

import { useMatches } from 'react-router-dom';

import { renderTitle, type RouteParams } from '@/lib/route-titles/render-title';

type TitleComponent = ComponentType<{ params: RouteParams }>;

type TitleResolver = ((params: RouteParams) => string) | string | TitleComponent;

const hasTitle = (handle: unknown): handle is { title: TitleResolver } => {
    if (typeof handle !== 'object' || handle === null || !('title' in handle)) {
        return false;
    }

    const value = (handle as { title: unknown }).title;

    return typeof value === 'string' || typeof value === 'function';
};

// Distinguishes a React component from a plain resolver by the PascalCase
// convention — components must start with an uppercase letter (Apollo title
// factories return named `ApolloTitle`, regular resolvers are anonymous
// arrow functions). This mirrors how React itself differentiates them in
// JSX, so we don't need an extra marker on the function.
const isTitleComponent = (fn: (params: RouteParams) => unknown): fn is TitleComponent => {
    const first = fn.name.charAt(0);

    return first === first.toUpperCase() && first !== first.toLowerCase();
};

/**
 * Document `<title>` driver. Walks `useMatches()` for the deepest route
 * exposing `handle.title` and renders accordingly:
 *
 *   handle: { title: 'Dashboard' }                          // static
 *   handle: { title: (p) => formatPromptId(p.promptId!) }   // params-derived
 *   handle: { title: FlowTitle }                            // reactive (Apollo)
 *
 * Living in the app shell, this component survives navigation between
 * sibling detail routes — no flash of a generic fallback during the
 * destination page's data fetch.
 *
 * IMPORTANT: do not render `<title>` anywhere else in the tree. React 19
 * hoists every `<title>` into `<head>`, and multiple simultaneous titles
 * lead to undefined browser/SEO behavior. The only safe `<title>` outside
 * this module is inside `<svg>` (icons), which React treats as SVG-title.
 * See https://react.dev/reference/react-dom/components/title.
 */
export function DocumentTitle() {
    const matches = useMatches();
    const match = matches.findLast((m) => hasTitle(m.handle));

    if (!match || !hasTitle(match.handle)) {
        return renderTitle(null);
    }

    const value = match.handle.title;

    if (typeof value === 'string') {
        return renderTitle(value);
    }

    if (isTitleComponent(value)) {
        const TitleComp = value;

        return <TitleComp params={match.params} />;
    }

    return renderTitle(value(match.params));
}
