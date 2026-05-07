import {
    OverwriteConfirmDialog,
    type OverwriteConflict,
} from '@/components/shared/overwrite-confirm-dialog';

interface ResourcesConflictDialogProps {
    /**
     * Conflicts collected from a batch operation. Empty array keeps the dialog hidden.
     * For a single conflict the message names the conflicting item; for many it falls
     * back to a count-based summary (Finder-style "Apply to all").
     */
    conflicts: OverwriteConflict[];
    onCancel: () => void;
    onReplaceAll: () => Promise<unknown> | unknown;
}

/**
 * Backwards-compatible thin wrapper over the shared {@link OverwriteConfirmDialog}.
 * Move / copy hooks already consume `ConflictItem` shaped values that mirror
 * `OverwriteConflict` 1:1, so this is a pure rename / re-export today.
 *
 * New call sites should depend on `OverwriteConfirmDialog` directly — this file
 * only exists to keep the existing imports in `resources-move-dialog.tsx` and
 * `resources-copy-dialog.tsx` working without churn.
 */
export const ResourcesConflictDialog = ({ conflicts, onCancel, onReplaceAll }: ResourcesConflictDialogProps) => (
    <OverwriteConfirmDialog
        conflicts={conflicts}
        onCancel={onCancel}
        onReplaceAll={onReplaceAll}
    />
);
