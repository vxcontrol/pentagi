import { useDetailNavigation } from '@/components/shared/detail-navigation';
import { type Knowledge, useKnowledges } from '@/providers/knowledges-provider';

const getLabel = (item: Knowledge) => item.question;
const getHref = (item: Knowledge) => `/knowledges/${item.id}`;

/**
 * Detail-page navigation wired up for knowledge documents. Returns a
 * `DetailNavigationController<Knowledge>` for `<DetailNavigationToolbar>` /
 * `<DetailNavigationButtons>` / `<DetailNavigationSheet>`. The list page
 * filters on `question` and the header shows the same, so `getLabel`
 * doubles as the default searchable text.
 */
export function useKnowledgeDetailNavigation(currentId: null | string | undefined) {
    const { knowledges } = useKnowledges();

    return useDetailNavigation<Knowledge>({
        currentId,
        getHref,
        getLabel,
        items: knowledges,
    });
}
