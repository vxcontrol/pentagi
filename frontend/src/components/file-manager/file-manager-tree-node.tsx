import { Fragment, type MouseEvent as ReactMouseEvent } from 'react';

import type { FileManagerAction, FileManagerInternalNode } from './file-manager-types';

import { FileManagerRow } from './file-manager-row';

interface FileManagerTreeNodeProps {
    actions: readonly FileManagerAction[];
    activeRowPath: null | string;
    expandedPaths: Set<string>;
    formatModified?: (modifiedAt: Date | string | undefined) => string;
    gridTemplate: string;
    hasActions: boolean;
    isCheckboxVisible: boolean;
    isModifiedVisible: boolean;
    isSizeVisible: boolean;
    node: FileManagerInternalNode;
    onClick: (event: ReactMouseEvent, path: string) => void;
    onFocusRow: (path: string) => void;
    onToggleCheckbox: (path: string) => void;
    onToggleExpand: (path: string, wasExpanded: boolean) => void;
    /** 1-based position of the node inside its parent's child list (for `aria-posinset`). */
    posInSet: number;
    searchQuery?: string;
    selectedPaths: Set<string>;
    /** Total number of siblings the node is part of (for `aria-setsize`). */
    setSize: number;
}

/**
 * Recursive tree node renderer. The component itself is not memoized — `FileManagerRow`
 * is, which is where per-row reconciliation savings actually matter. Extracting the
 * recursion into a component (instead of an inline `renderNode` function) keeps
 * `FileManager` lean and gives React DevTools a real boundary to inspect.
 */
export const FileManagerTreeNode = ({
    actions,
    activeRowPath,
    expandedPaths,
    formatModified,
    gridTemplate,
    hasActions,
    isCheckboxVisible,
    isModifiedVisible,
    isSizeVisible,
    node,
    onClick,
    onFocusRow,
    onToggleCheckbox,
    onToggleExpand,
    posInSet,
    searchQuery,
    selectedPaths,
    setSize,
}: FileManagerTreeNodeProps) => {
    const isExpanded = expandedPaths.has(node.path);
    const isSelected = selectedPaths.has(node.path);
    const renderChildren = node.isDir && isExpanded && node.children.length > 0;

    return (
        <Fragment>
            <FileManagerRow
                actions={actions}
                activeRowPath={activeRowPath}
                file={node}
                formatModified={formatModified}
                gridTemplate={gridTemplate}
                hasActions={hasActions}
                isCheckboxVisible={isCheckboxVisible && !node.isGroupRoot}
                isExpanded={isExpanded}
                isModifiedVisible={isModifiedVisible}
                isSelected={isSelected}
                isSizeVisible={isSizeVisible}
                onClick={onClick}
                onFocusRow={onFocusRow}
                onToggleCheckbox={onToggleCheckbox}
                onToggleExpand={onToggleExpand}
                posInSet={posInSet}
                searchQuery={searchQuery}
                setSize={setSize}
            />
            {renderChildren &&
                node.children.map((child, index) => (
                    <FileManagerTreeNode
                        actions={actions}
                        activeRowPath={activeRowPath}
                        expandedPaths={expandedPaths}
                        formatModified={formatModified}
                        gridTemplate={gridTemplate}
                        hasActions={hasActions}
                        isCheckboxVisible={isCheckboxVisible}
                        isModifiedVisible={isModifiedVisible}
                        isSizeVisible={isSizeVisible}
                        key={child.id}
                        node={child}
                        onClick={onClick}
                        onFocusRow={onFocusRow}
                        onToggleCheckbox={onToggleCheckbox}
                        onToggleExpand={onToggleExpand}
                        posInSet={index + 1}
                        searchQuery={searchQuery}
                        selectedPaths={selectedPaths}
                        setSize={node.children.length}
                    />
                ))}
        </Fragment>
    );
};
