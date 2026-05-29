import { useDetailNavigation } from '@/components/shared/detail-navigation';
import { type Template, useTemplates } from '@/providers/templates-provider';

const getLabel = (item: Template) => item.title;
const getId = (item: Template) => String(item.id);
const getHref = (item: Template) => `/templates/${item.id}`;

/**
 * Detail-page navigation wired up for templates. Returns a
 * `DetailNavigationController<Template>` for `<DetailNavigationToolbar>` /
 * `<DetailNavigationButtons>` / `<DetailNavigationSheet>`. The list page
 * filters on `title` and the breadcrumb shows the same, so `getLabel`
 * doubles as the default searchable text (no explicit `getSearchableText`
 * needed).
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
