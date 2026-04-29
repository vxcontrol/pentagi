import { Trash2 } from 'lucide-react';
import { useCallback, useState } from 'react';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Button } from '@/components/ui/button';

import type { FileManagerLabels, FileNode } from './file-manager-types';

import { dedupeOverlappingPaths, pluralizeItemsEnglish } from './file-manager-utils';

interface FileManagerBulkActionsBarProps {
    files: FileNode[];
    labels: FileManagerLabels;
    onBulkDelete?: (files: FileNode[]) => Promise<void> | void;
    onClearSelection: () => void;
    onSelectionChange: (paths: Set<string>) => void;
    selectedPaths: Set<string>;
}

/**
 * Footer that appears when at least one row is selected. Hosts the cancel/delete
 * controls and owns the bulk-delete confirmation dialog state, so the parent doesn't
 * have to coordinate it.
 */
export const FileManagerBulkActionsBar = ({
    files,
    labels,
    onBulkDelete,
    onClearSelection,
    onSelectionChange,
    selectedPaths,
}: FileManagerBulkActionsBarProps) => {
    const [isConfirmOpen, setIsConfirmOpen] = useState(false);

    const handleConfirm = useCallback(async () => {
        if (!onBulkDelete) {
            return;
        }

        const dedupedPaths = new Set(dedupeOverlappingPaths(selectedPaths));
        const filesToDelete = files.filter((file) => dedupedPaths.has(file.path));

        await onBulkDelete(filesToDelete);
        onSelectionChange(new Set());
    }, [files, onBulkDelete, onSelectionChange, selectedPaths]);

    const handleOpenChange = useCallback((isOpen: boolean) => {
        if (!isOpen) {
            setIsConfirmOpen(false);
        }
    }, []);

    if (selectedPaths.size === 0) {
        return null;
    }

    const pluralize = labels.pluralizeItems ?? pluralizeItemsEnglish;
    const countLabel = pluralize(selectedPaths.size);
    const selectedText = labels.selectedLabel?.(selectedPaths.size) ?? `${selectedPaths.size} selected`;
    const cancelText = labels.bulkCancel ?? 'Cancel';
    const confirmText = labels.bulkConfirm ?? 'Delete';
    const titleText = labels.bulkTitle?.(countLabel) ?? `Delete ${countLabel}`;
    const descriptionText =
        labels.bulkDescription?.(countLabel) ?? `This will delete ${countLabel}. This action cannot be undone.`;

    return (
        <>
            <div className="bg-background flex items-center justify-between gap-2 border-t px-3 py-2">
                <span className="text-muted-foreground text-sm">{selectedText}</span>
                <div className="flex items-center gap-2">
                    <Button
                        onClick={onClearSelection}
                        size="sm"
                        variant="ghost"
                    >
                        {cancelText}
                    </Button>
                    {onBulkDelete && (
                        <Button
                            onClick={() => setIsConfirmOpen(true)}
                            size="sm"
                            variant="destructive"
                        >
                            <Trash2 className="size-4" />
                            {confirmText}
                        </Button>
                    )}
                </div>
            </div>

            <ConfirmationDialog
                confirmText={confirmText}
                description={descriptionText}
                handleConfirm={handleConfirm}
                handleOpenChange={handleOpenChange}
                isOpen={isConfirmOpen}
                title={titleText}
            />
        </>
    );
};
