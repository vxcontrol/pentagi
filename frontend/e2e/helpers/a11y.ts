import type { Page } from '@playwright/test';

import AxeBuilder from '@axe-core/playwright';
import { expect } from '@playwright/test';

/**
 * Known pre-existing violations, keyed by route. A waiver names the axe rule
 * AND a pattern for the offending nodes' target selectors — only matching
 * nodes are waived, so the rule still fires anywhere else on the page. An
 * entry here is technical debt with a name — remove it when the violation is
 * fixed, never add one without a look at the actual finding.
 */
export type A11yAllowlist = Record<string, A11yWaiver[]>;

export interface A11yWaiver {
    rule: string;
    target: RegExp;
}

const BLOCKING_IMPACTS = new Set(['critical', 'serious']);

export const scanA11y = async (page: Page, route: string, allowlist: A11yAllowlist): Promise<void> => {
    const waivers = allowlist[route] ?? [];
    const results = await new AxeBuilder({ page })
        .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'])
        .analyze();

    const blocking = results.violations
        .filter(({ impact }) => BLOCKING_IMPACTS.has(impact ?? ''))
        .map(({ id, impact, nodes }) => ({
            id,
            impact,
            nodes: nodes
                .filter(
                    (node) => !waivers.some(({ rule, target }) => rule === id && target.test(node.target.join(' '))),
                )
                .slice(0, 5)
                .map((node) => ({
                    html: node.html.slice(0, 160),
                    target: node.target.join(' '),
                })),
        }))
        .filter(({ nodes }) => nodes.length > 0);

    expect(blocking, `axe critical/serious violations on ${route}`).toEqual([]);
};
