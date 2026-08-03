import type { Page } from '@playwright/test';

import AxeBuilder from '@axe-core/playwright';
import { expect } from '@playwright/test';

/**
 * One known pre-existing violation. A waiver names the axe rule AND a pattern
 * for the offending nodes' target selectors — only matching nodes are waived,
 * so the rule still fires anywhere else on the page. A waiver is technical debt
 * with a name — remove it when the violation is fixed, never add one without a
 * look at the actual finding.
 */
export interface A11yWaiver {
    rule: string;
    /**
     * Flow-detail tab names whose panel owns this debt. Omitted = route-wide (applies to the base
     * view and every tab). Scoped so a Files-tab waiver cannot silence the same rule on Assistant.
     */
    tabs?: string[];
    target: RegExp;
}

/** Waivers in force for one scan: route-wide ones always, tab-scoped ones only on their own tab. */
export const waiversForScan = (waivers: A11yWaiver[] = [], tab?: string): A11yWaiver[] =>
    waivers.filter((waiver) => !waiver.tabs || (tab !== undefined && waiver.tabs.includes(tab)));

const BLOCKING_IMPACTS = new Set(['critical', 'serious']);

export const scanA11y = async (page: Page, route: string, waivers: A11yWaiver[] = []): Promise<void> => {
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
