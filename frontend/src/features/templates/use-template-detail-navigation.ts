import { useDetailNavigation } from '@/components/shared/detail-navigation';
import { routes } from '@/lib/routes';
import { type Template, useTemplates } from '@/providers/templates-provider';

const getLabel = (item: Template) => item.title;
const getId = (item: Template) => String(item.id);
const getHref = (item: Template) => routes.template(item.id);

/**
 * The list page filters on `title`, so `getLabel` doubles as the default
 * searchable text (no explicit `getSearchableText` needed).
 */
export function useTemplateDetailNavigation(currentId: null | string | undefined) {
    const { templates } = useTemplates();

    return useDetailNavigation<Template>({
        currentId,
        getHref,
        getId,
        getLabel,
        items: templates,
    });
}
