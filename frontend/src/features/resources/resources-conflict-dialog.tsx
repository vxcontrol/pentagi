import { Replace } from 'lucide-react';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';

interface ConflictItem {
    destination: string;
    destinationName: string;
}

interface ResourcesConflictDialogProps {
    /**
     * Conflicts collected from a batch operation. Empty array keeps the dialog hidden.
     * For a single conflict the message names the conflicting item; for many it falls
     * back to a count-based summary (Finder-style "Apply to all").
     */
    conflicts: ConflictItem[];
    onCancel: () => void;
    onReplaceAll: () => Promise<unknown> | unknown;
}

const buildDescription = (conflicts: ConflictItem[]): string | undefined => {
    const single = conflicts.length === 1 ? conflicts[0] : undefined;

    if (single) {
        return `An item named "${single.destinationName}" already exists at /${single.destination}. Do you want to replace it?`;
    }

    if (conflicts.length > 1) {
        return `${conflicts.length} items already exist at the destination. Do you want to replace all of them?`;
    }

    return undefined;
};

const buildConfirmText = (count: number): string => (count > 1 ? 'Replace all' : 'Replace');

/**
 * Shared "Replace or cancel" dialog for `409` conflicts coming from move / copy hooks.
 * The hook owns the conflict state; this component only renders the prompt and forwards
 * the user's decision back through the callbacks. A batch decision (Replace all) is
 * applied to every pending conflict in one shot — this matches the OS file-manager UX
 * and keeps the user from being prompted N times for the same destination directory.
 */
export const ResourcesConflictDialog = ({ conflicts, onCancel, onReplaceAll }: ResourcesConflictDialogProps) => (
    <ConfirmationDialog
        cancelText="Cancel"
        confirmIcon={<Replace />}
        confirmText={buildConfirmText(conflicts.length)}
        confirmVariant="destructive"
        description={buildDescription(conflicts)}
        handleConfirm={async () => {
            await onReplaceAll();
        }}
        handleOpenChange={(nextOpen) => {
            if (!nextOpen) {
                onCancel();
            }
        }}
        isOpen={conflicts.length > 0}
        title="Replace existing item?"
    />
);
