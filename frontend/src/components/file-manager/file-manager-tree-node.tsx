import { Fragment } from 'react';

import type { FileManagerRowDisplay, FileManagerRowHandlers } from './file-manager-row';
import type { FileManagerAction, FileManagerInternalNode } from './file-manager-types';
import type { FileManagerNodeDndHandlers } from './use-file-manager-dnd';

import { FileManagerRow } from './file-manager-row';

interface FileManagerTreeNodeProps {
    actions: readonly FileManagerAction[];
    activeRowPath: null | string;
    /** Returns drag/drop handlers for a node, or `null` when DnD is disabled. */
    bindNodeDnd: (node: FileManagerInternalNode) => FileManagerNodeDndHandlers | null;
    /** Pre-computed tri-state value per directory path; missing entries default to `false`. */
    dirSelectionStates: ReadonlyMap<string, 'indeterminate' | boolean>;
    /** Pre-computed subtree paths per directory path (the dir itself + descendants). */
    dirSubtreePaths: ReadonlyMap<string, readonly string[]>;
    /** Per-tree shared layout / i18n bundle. Forwarded as-is to every row. */
    display: FileManagerRowDisplay;
    expandedPaths: Set<string>;
    /** Per-tree shared callback bundle. Forwarded as-is to every row. */
    handlers: FileManagerRowHandlers;
    node: FileManagerInternalNode;
    /** 1-based position of the node inside its parent's child list (for `aria-posinset`). */
    posInSet: number;
    selectedPaths: ReadonlySet<string>;
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
    bindNodeDnd,
    dirSelectionStates,
    dirSubtreePaths,
    display,
    expandedPaths,
    handlers,
    node,
    posInSet,
    selectedPaths,
    setSize,
}: FileManagerTreeNodeProps) => {
    const isExpanded = expandedPaths.has(node.path);
    const isSelected = selectedPaths.has(node.path);
    const renderChildren = node.isDir && isExpanded && node.children.length > 0;
    const dnd = bindNodeDnd(node);

    // Pre-resolve tri-state + subtree paths per row: keeping the lookup outside
    // the memoized `FileManagerRow` lets each row receive primitive/stable props
    // (`'indeterminate' | true | false | undefined` plus a `Map`-stored array
    // reference) so a selection change only re-renders the rows whose computed
    // state actually flipped.
    const dirCheckboxState = node.isDir ? (dirSelectionStates.get(node.path) ?? false) : undefined;
    const subtreePaths = node.isDir ? dirSubtreePaths.get(node.path) : undefined;

    return (
        <Fragment>
            <FileManagerRow
                actions={actions}
                activeRowPath={activeRowPath}
                dirCheckboxState={dirCheckboxState}
                dirSubtreePaths={subtreePaths}
                display={display}
                dnd={dnd}
                file={node}
                handlers={handlers}
                isExpanded={isExpanded}
                isSelected={isSelected}
                posInSet={posInSet}
                setSize={setSize}
            />
            {renderChildren &&
                node.children.map((child, index) => (
                    <FileManagerTreeNode
                        actions={actions}
                        activeRowPath={activeRowPath}
                        bindNodeDnd={bindNodeDnd}
                        dirSelectionStates={dirSelectionStates}
                        dirSubtreePaths={dirSubtreePaths}
                        display={display}
                        expandedPaths={expandedPaths}
                        handlers={handlers}
                        key={child.id}
                        node={child}
                        posInSet={index + 1}
                        selectedPaths={selectedPaths}
                        setSize={node.children.length}
                    />
                ))}
        </Fragment>
    );
};
