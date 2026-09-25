import { useCallback, useEffect, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router-dom';

import { useSearchParamTab } from '@/hooks/use-search-param-tab';
import { getLatestSearchParams } from '@/lib/url-params';
import { useFlow } from '@/providers/flow-provider';

export const CENTRAL_TAB_VALUES = ['automation', 'assistant', 'dashboard'];

export const SIDE_TAB_VALUES = ['terminal', 'tasks', 'agents', 'tools', 'vectorStores', 'files', 'screenshots'];

export const FLOW_TAB_VALUES = [...CENTRAL_TAB_VALUES, ...SIDE_TAB_VALUES];

export const SIDE_TAB_PARAM = 'side';

const DEFAULT_SIDE_TAB = 'terminal';

const readParam = (snapshot: URLSearchParams, key: string, values: readonly string[]): null | string => {
    const value = getLatestSearchParams(snapshot).get(key);

    return value && values.includes(value) ? value : null;
};

const writeParam = (key: string, value: string) => (previous: URLSearchParams) => {
    const params = getLatestSearchParams(previous);

    params.set(key, value);

    return params;
};

export function useFlowSideTab() {
    const [searchParams] = useSearchParams();

    const linked = readParam(searchParams, 'tab', SIDE_TAB_VALUES);
    const named = readParam(searchParams, SIDE_TAB_PARAM, SIDE_TAB_VALUES);

    const [resolvedTab, handleTabChange] = useSearchParamTab({
        fallback: linked ?? DEFAULT_SIDE_TAB,
        key: SIDE_TAB_PARAM,
        values: SIDE_TAB_VALUES,
    });

    useEffect(() => {
        if (linked && !named) {
            handleTabChange(linked);
        }
    }, [handleTabChange, linked, named]);

    return { handleTabChange, resolvedTab };
}

export function useFlowTabDetection(tabValues: readonly string[] = CENTRAL_TAB_VALUES) {
    const { flowData, isLoading } = useFlow();
    const [searchParams, setSearchParams] = useSearchParams();

    const flowId = flowData?.flow?.id ?? null;

    const [settled, setSettled] = useState<null | { flowId: null | string; tab: string }>(null);

    if (!isLoading && settled?.flowId !== flowId) {
        setSettled({ flowId, tab: flowData?.messageLogs?.length ? 'automation' : 'assistant' });
    }

    const resolvedTab = useMemo(() => {
        const tabParam = readParam(searchParams, 'tab', tabValues);

        if (tabParam) {
            return tabParam;
        }

        if (settled?.flowId === flowId) {
            return settled.tab;
        }

        return 'automation';
    }, [searchParams, tabValues, settled, flowId]);

    const handleTabChange = useCallback(
        (tab: string) => {
            setSearchParams((previous) => {
                const params = writeParam('tab', tab)(previous);

                if (tabValues.includes(tab) && SIDE_TAB_VALUES.includes(tab)) {
                    params.set(SIDE_TAB_PARAM, tab);
                }

                return params;
            });
        },
        [setSearchParams, tabValues],
    );

    return { handleTabChange, resolvedTab };
}
