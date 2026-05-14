import { useDetailNavigation } from '@/hooks/use-detail-navigation';
import { type Knowledge, useKnowledges } from '@/providers/knowledges-provider';

const getLabel = (item: Knowledge) => item.question;
const getHref = (item: Knowledge) => `/knowledges/${item.id}`;

/**
 * Detail-page navigation wired up for knowledge documents. The list page
 * filters on `question` and the header shows the same, so `getLabel`
 * doubles as the default searchable text.
 */
export const useKnowledgeDetailNavigation = (currentId: null | string | undefined) => {
    const { knowledges } = useKnowledges();

    return useDetailNavigation<Knowledge>({
        currentId,
        getHref,
        getLabel,
        items: knowledges,
    });
};
