import { useCallback, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';

import { useFlow } from '@/providers/flow-provider';

const CENTRAL_TAB_VALUES = ['automation', 'assistant', 'dashboard'];

/**
 * Detects the appropriate central tab based on priority:
 * 1. Manual user selection
 * 2. URL search parameter (?tab=...)
 * 3. Auto-detection: 'assistant' if flow is loaded and has no message logs
 * 4. Default: 'automation'
 */
export function useFlowTabDetection() {
    const { flowData, isLoading } = useFlow();
    const [searchParams, setSearchParams] = useSearchParams();
    const [manualTab, setManualTab] = useState<null | string>(null);

    const resolvedTab = useMemo(() => {
        // If user manually selected a tab, use it
        if (manualTab) {
            return manualTab;
        }

        // Check URL parameter
        const tabParam = searchParams.get('tab');

        if (tabParam && CENTRAL_TAB_VALUES.includes(tabParam)) {
            return tabParam;
        }

        // Auto-detect: switch to assistant tab if flow is loaded and messageLogs are empty
        if (!isLoading && !flowData?.messageLogs?.length) {
            return 'assistant';
        }

        return 'automation';
    }, [manualTab, searchParams, isLoading, flowData?.messageLogs]);

    const handleTabChange = useCallback(
        (tab: string) => {
            setManualTab(tab);
            setSearchParams({ tab });
        },
        [setSearchParams],
    );

    return { handleTabChange, resolvedTab };
}
