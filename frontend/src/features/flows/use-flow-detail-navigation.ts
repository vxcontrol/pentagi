import { useDetailNavigation } from '@/components/shared/detail-navigation';
import { type Flow, useFlows } from '@/providers/flows-provider';

const getLabel = (item: Flow) => item.title || `Flow #${item.id}`;
const getSearchableText = (item: Flow) => item.title;
const getId = (item: Flow) => String(item.id);
const getHref = (item: Flow) => `/flows/${item.id}`;

/**
 * Detail-page navigation wired up for flows. Encapsulates the getter
 * callbacks and the call to `useDetailNavigation` so each detail page just
 * passes the returned controller to `<DetailNavigationToolbar controller={nav}>`
 * (or to its leaf primitives for custom chrome).
 *
 * Pass `null` instead of an id while the page is in a non-viewing state
 * (e.g. `/flows/new`) so the controller reports an unmatched current item.
 */
export function useFlowDetailNavigation(currentId: null | string | undefined) {
    const { flows } = useFlows();

    return useDetailNavigation<Flow>({
        currentId,
        getHref,
        getId,
        getLabel,
        getSearchableText,
        items: flows,
    });
}
