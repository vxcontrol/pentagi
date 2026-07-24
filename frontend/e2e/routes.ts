import type { Locator, Page } from '@playwright/test';

import { routes } from '@/lib/routes';

import type { A11yWaiver } from './helpers/a11y.ts';
import type { Cassette } from './mocks/cassette.ts';

import { apiTokensCassette } from './mocks/cassettes/api-tokens.ts';
import { dashboardCassette } from './mocks/cassettes/dashboard.ts';
import { flowsCassette, flowTabsCassette, TABS_FILE_NAME, TABS_SCREENSHOT_URL } from './mocks/cassettes/flows.ts';
import { knowledgesCassette } from './mocks/cassettes/knowledges.ts';
import { resourcesCassette } from './mocks/cassettes/resources.ts';
import { settingsPromptsCassette } from './mocks/cassettes/settings-prompts.ts';
import { settingsProvidersCassette } from './mocks/cassettes/settings-providers.ts';
import { templatesCassette } from './mocks/cassettes/templates.ts';

export interface RouteManifestEntry {
    a11yWaivers?: A11yWaiver[];
    cassette: () => Cassette;
    path: string;
    ready: (page: Page) => Locator;
    /**
     * Owning src/ areas — the substrate for changed-files → affected-routes
     * mapping. A dir rendered by several routes must be listed under every one
     * of them: ownership never falls back to the run-everything path, so a
     * single-route listing silently skips the other consumers.
     */
    sources: string[];
    /** Radix unmounts inactive tab panels, so one scan of the default view sees none of them. */
    tabs?: RouteTab[];
}

export interface RouteTab {
    name: string;
    /**
     * Must be absent while the panel loads: a marker that also renders on the
     * skeleton lets every per-tab scan pass on an empty panel.
     */
    ready: (page: Page) => Locator;
}

/** Order matters: the flow auto-opens Assistant, so tabs.spec pins that no entry is already open. */
export const FLOW_DETAIL_TABS: RouteTab[] = [
    { name: 'Dashboard', ready: (page) => page.getByText('Usage by Model & Provider') },
    { name: 'Assistant', ready: (page) => page.getByText('New assistant', { exact: true }) },
    { name: 'Tasks', ready: (page) => page.getByText('E2E Task Alpha') },
    { name: 'Agents', ready: (page) => page.getByText('E2E agent reconnaissance') },
    { name: 'Searches', ready: (page) => page.getByText('E2E search for the CVE') },
    { name: 'Vector Store', ready: (page) => page.getByText('E2E recall prior findings') },
    { name: 'Files', ready: (page) => page.getByText(TABS_FILE_NAME) },
    { name: 'Screenshots', ready: (page) => page.getByText(TABS_SCREENSHOT_URL) },
];

/**
 * The routes swept by NAV/visual/a11y specs and CI diff-scoping. This is a
 * subset of the app's routes — route-manifest.unit.test.ts pins the excluded
 * remainder so a new route cannot silently stay out of every sweep. Paths come
 * from the app's own routes module so a route rename breaks this file at
 * compile time.
 */
export const ROUTE_MANIFEST: RouteManifestEntry[] = [
    {
        cassette: flowsCassette,
        path: routes.flows,
        ready: (page) => page.getByRole('row', { name: /E2E Alpha/ }),
        // Per file, not the whole dir: the report and create pages sit beside these, and claiming
        // the dir would scope their diffs to routes that never render them.
        sources: ['src/pages/flows/flows.tsx', 'src/features/flows', 'src/providers/flows-provider.tsx'],
    },
    {
        // Message metadata (date + ID) renders at 50% opacity by design; revisit
        // with the design pass.
        a11yWaivers: [
            { rule: 'color-contrast', target: /text-muted-foreground\\?\/50/ },
            // The tab-panel ones below are defects awaiting a fix, not accepted design.
            { rule: 'aria-progressbar-name', target: /\.bg-primary\\\/20/ },
            // Both are the Files tab's file-manager controls. Anchored on the offending nodes: matching
            // `button[aria-label` instead would waive the rule for every labelled button on the route.
            { rule: 'button-name', target: /tooltip-trigger.*size-8\[data-slot="button"\]/ },
            { rule: 'target-size', target: /text-blue-400|button\[aria-label="Select / },
            { rule: 'color-contrast', target: /\.text-primary > \.font-semibold\.truncate/ },
            // The same file-manager row metadata waived on /resources — this tab embeds it.
            { rule: 'color-contrast', target: /text-muted-foreground\\?\/80/ },
            { rule: 'scrollable-region-focusable', target: /\[data-slot="table-container"\]/ },
        ],
        // Must stay the populated cassette: the empty seed renders none of the sources below.
        cassette: flowTabsCassette,
        path: routes.flow('5'),
        // The terminal mounts after the header, and the visual spec masks it —
        // capturing before it exists compares live pixels against a masked baseline.
        ready: (page) => page.locator('.xterm').first(),
        sources: [
            'src/pages/flows/flow.tsx',
            'src/features/flows',
            'src/providers/flow-provider.tsx',
            'src/providers/flows-provider.tsx',
            'src/components/shared/file-manager',
            'src/components/dashboard',
            'src/features/resources',
        ],
        tabs: FLOW_DETAIL_TABS,
    },
    {
        cassette: templatesCassette,
        path: routes.templates,
        ready: (page) => page.getByRole('row', { name: /E2E Seed Template/ }),
        // The list file, not the dir: template.tsx sits beside it and no manifest route renders it,
        // so owning the dir scopes a detail-only diff here instead of to the run-everything path.
        sources: ['src/pages/templates/templates.tsx', 'src/providers/templates-provider.tsx'],
    },
    {
        cassette: knowledgesCassette,
        path: routes.knowledges,
        ready: (page) => page.getByRole('row', { name: /E2E Seed Question/ }),
        // `src/features/knowledges` is left unowned deliberately: knowledge.tsx is its only importer
        // and that route is not swept, so claiming it here would scope its diffs to a route that
        // never renders it rather than to the run-everything fallback.
        sources: ['src/pages/knowledges/knowledges.tsx', 'src/providers/knowledges-provider.tsx'],
    },
    {
        cassette: apiTokensCassette,
        path: routes.settings.apiTokens,
        ready: (page) => page.getByRole('row', { name: /E2E seed token/ }),
        sources: ['src/pages/settings/settings-api-tokens.tsx'],
    },
    {
        // The period switcher drives Tabs as a segmented control with no
        // TabsContent, so the active trigger's aria-controls names an element
        // that never exists.
        a11yWaivers: [{ rule: 'aria-valid-attr-value', target: /radix-.*-trigger-/ }],
        cassette: dashboardCassette,
        path: routes.dashboard,
        ready: (page) => page.getByRole('heading', { name: 'Flows Activity Over Time' }),
        sources: ['src/pages/dashboard', 'src/components/dashboard'],
    },
    {
        cassette: settingsPromptsCassette,
        path: routes.settings.prompts,
        ready: (page) => page.getByRole('heading', { name: 'Agent Prompts' }),
        sources: ['src/pages/settings/settings-prompts.tsx'],
    },
    {
        cassette: settingsProvidersCassette,
        path: routes.settings.providers,
        ready: (page) => page.getByText('No providers configured'),
        sources: ['src/pages/settings/settings-providers.tsx'],
    },
    {
        // Row size/modified metadata renders at 80% opacity and misses AA; the
        // tree's expand toggle and row checkboxes sit under the 24px pointer
        // target floor — widening them is a density decision for the file manager.
        a11yWaivers: [
            { rule: 'color-contrast', target: /text-muted-foreground\\?\/80/ },
            // `.rounded` unanchored would also match every rounded-* utility, so a
            // future under-sized control anywhere on the page would be waived too.
            { rule: 'target-size', target: /\.rounded(?![-\w])|aria-label="Select / },
        ],
        cassette: resourcesCassette,
        path: routes.resources,
        ready: (page) => page.getByRole('treeitem', { name: /reports/ }),
        sources: ['src/pages/resources', 'src/features/resources', 'src/components/shared/file-manager'],
    },
];
