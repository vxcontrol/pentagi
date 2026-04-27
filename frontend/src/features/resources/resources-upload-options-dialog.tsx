import { useState } from 'react';

import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { cn } from '@/lib/utils';

import type { UploadOption } from './use-file-upload';

interface OptionDescriptor {
    description: string;
    label: string;
    value: UploadOption;
}

interface ResourcesUploadOptionsDialogProps {
    isOpen: boolean;
    onCancel: () => void;
    onConfirm: (option: UploadOption) => void;
}

const OPTIONS: OptionDescriptor[] = [
    {
        description: 'Uploads will overwrite files with the same names.',
        label: 'Replace existing files',
        value: 'replace',
    },
    {
        description: 'Saves both copies. New files get renamed automatically.',
        label: 'Keep both versions',
        value: 'keep-both',
    },
];

export function ResourcesUploadOptionsDialog({ isOpen, onCancel, onConfirm }: ResourcesUploadOptionsDialogProps) {
    const [selectedOption, setSelectedOption] = useState<UploadOption>('replace');

    const handleCancel = () => {
        setSelectedOption('replace');
        onCancel();
    };

    const handleConfirm = () => {
        const option = selectedOption;
        setSelectedOption('replace');
        onConfirm(option);
    };

    return (
        <Dialog
            onOpenChange={(open) => {
                if (!open) {
                    handleCancel();
                }
            }}
            open={isOpen}
        >
            <DialogContent
                className="max-w-md"
                onEscapeKeyDown={(event) => {
                    event.stopPropagation();
                    event.preventDefault();
                }}
                onInteractOutside={(event) => event.preventDefault()}
            >
                <DialogHeader>
                    <DialogTitle>Upload options</DialogTitle>
                    <DialogDescription>One or more files already exist in Resources.</DialogDescription>
                </DialogHeader>

                <div className="flex flex-col gap-2">
                    {OPTIONS.map((option) => {
                        const isSelected = selectedOption === option.value;

                        return (
                            <button
                                aria-pressed={isSelected}
                                className={cn(
                                    'flex items-start gap-3 rounded-lg border p-3 text-left transition-colors',
                                    'hover:bg-muted/40 focus-visible:ring-ring focus-visible:ring-2 focus-visible:outline-hidden',
                                    isSelected ? 'border-primary bg-muted/30' : 'border-border',
                                )}
                                key={option.value}
                                onClick={() => setSelectedOption(option.value)}
                                type="button"
                            >
                                <span
                                    aria-hidden
                                    className={cn(
                                        'mt-0.5 inline-flex size-4 shrink-0 items-center justify-center rounded-full border',
                                        isSelected ? 'border-primary' : 'border-input',
                                    )}
                                >
                                    <span
                                        className={cn(
                                            'size-2 rounded-full transition-opacity',
                                            isSelected ? 'bg-primary opacity-100' : 'opacity-0',
                                        )}
                                    />
                                </span>
                                <span className="flex flex-col gap-0.5">
                                    <span className="text-sm font-medium">{option.label}</span>
                                    <span className="text-muted-foreground text-xs">{option.description}</span>
                                </span>
                            </button>
                        );
                    })}
                </div>

                <DialogFooter className="gap-2">
                    <Button
                        onClick={handleCancel}
                        variant="outline"
                    >
                        Cancel
                    </Button>
                    <Button onClick={handleConfirm}>Upload</Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
