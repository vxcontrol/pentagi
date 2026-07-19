import type { Page } from '@playwright/test';

import AxeBuilder from '@axe-core/playwright';
import { expect } from '@playwright/test';

/**
 * Known pre-existing violations, keyed by route, listing axe rule ids. An entry
 * here is technical debt with a name — remove the line when the violation is
 * fixed, never add one without a look at the actual finding.
 */
export type A11yAllowlist = Record<string, string[]>;

const BLOCKING_IMPACTS = new Set(['critical', 'serious']);

export const scanA11y = async (page: Page, route: string, allowlist: A11yAllowlist): Promise<void> => {
    const allowed = new Set(allowlist[route] ?? []);
    const results = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa']).analyze();

    const blocking = results.violations
        .filter(({ id, impact }) => BLOCKING_IMPACTS.has(impact ?? '') && !allowed.has(id))
        .map(({ id, impact, nodes }) => ({
            id,
            impact,
            nodes: nodes.slice(0, 5).map((node) => ({
                html: node.html.slice(0, 160),
                target: node.target.join(' '),
            })),
        }));

    expect(blocking, `axe critical/serious violations on ${route}`).toEqual([]);
};
