import { UploadCloudIcon } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

export interface FileDropZoneProps {
    /** Label of the call-to-action button. */
    actionLabel?: string;
    className?: string;
    /** Sub-text shown under the title. */
    description?: string;
    /** Tiny extra hint rendered at the bottom (e.g. limits). */
    hint?: string;
    /** Highlight state — pass the `isDragging` flag from `useFilesDragAndDrop`. */
    isDragging?: boolean;
    /** Disables the action while an upload is already in flight. */
    isUploading?: boolean;
    /** Called when the user clicks the action button. */
    onBrowse: () => void;
    /** Heading of the drop zone. */
    title?: string;
}

/**
 * Stateless drop zone surface. The component itself doesn't bind any drag/drop
 * handlers — the parent attaches them on the actual drop container via
 * `useFilesDragAndDrop` and forwards the `isDragging` flag here for visual feedback.
 *
 * The button uses `onBrowse` to open the file picker (typically wired to
 * `useResourcesUpload().openFilePicker`).
 */
export const FileDropZone = ({
    actionLabel = 'Browse files',
    className,
    description = 'or click the button to choose files from your computer',
    hint,
    isDragging = false,
    isUploading = false,
    onBrowse,
    title = 'Drag and drop your files',
}: FileDropZoneProps) => (
    <div
        className={cn(
            'bg-card flex flex-col items-center justify-center rounded-md border-2 border-dashed px-6 py-12 text-center transition-all duration-200',
            isDragging
                ? 'border-foreground bg-muted/50'
                : 'border-muted-foreground/25 hover:border-muted-foreground/40',
            className,
        )}
    >
        <UploadCloudIcon
            className={cn(
                'text-muted-foreground mb-3 size-10 transition-transform duration-200',
                isDragging && 'text-foreground scale-110',
            )}
        />
        <span className="text-sm font-medium">{isDragging ? 'Drop files here' : title}</span>
        <span className="text-muted-foreground mt-1 max-w-md text-xs">{description}</span>
        <Button
            className="mt-4 h-7 text-xs"
            disabled={isUploading}
            onClick={onBrowse}
            size="sm"
            variant="outline"
        >
            {actionLabel}
        </Button>
        {hint && <span className="text-muted-foreground mt-3 text-xs">{hint}</span>}
    </div>
);
