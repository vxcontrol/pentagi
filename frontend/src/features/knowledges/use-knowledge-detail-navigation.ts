import { useDetailNavigation } from '@/components/shared/detail-navigation';
import { routes } from '@/lib/routes';
import { type Knowledge, useKnowledges } from '@/providers/knowledges-provider';

const getLabel = (item: Knowledge) => item.question;
const getHref = (item: Knowledge) => routes.knowledge(item.id);

/**
 * The list page filters on `question`, so `getLabel` doubles as the default
 * searchable text (no explicit `getSearchableText` needed).
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
