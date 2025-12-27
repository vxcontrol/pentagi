import type { ReactElement } from 'react';

import { Trash2 } from 'lucide-react';
import { cloneElement, isValidElement } from 'react';

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

type ConfirmationDialogIconProps = ReactElement<React.SVGProps<SVGSVGElement>>;

interface ConfirmationDialogProps {
    cancelIcon?: ConfirmationDialogIconProps;
    cancelText?: string;
    cancelVariant?: 'default' | 'destructive' | 'ghost' | 'outline' | 'secondary';
    confirmIcon?: ConfirmationDialogIconProps;
    confirmText?: string;
    confirmVariant?: 'default' | 'destructive' | 'ghost' | 'outline' | 'secondary';
    description?: string;
    handleConfirm: () => void;
    handleOpenChange: (isOpen: boolean) => void;
    isOpen: boolean;
    itemName?: string;
    itemType?: string;
    title?: string;
}

const ConfirmationDialog = ({
    cancelIcon,
    cancelText = 'Cancel',
    cancelVariant = 'outline',
    confirmIcon = <Trash2 />,
    confirmText = 'Confirm',
    confirmVariant = 'destructive',
    description,
    handleConfirm,
    handleOpenChange,
    isOpen,
    itemName = 'this',
    itemType = 'item',
    title = 'Confirm Action',
}: ConfirmationDialogProps) => {
    const defaultDescription = description || (
        <>
            Are you sure you want to perform this action on{' '}
            <strong className="text-foreground font-semibold">{itemName}</strong> {itemType}?
        </>
    );

    // Common method to process icons with h-4 w-4 classes
    const processIcon = (icon?: ConfirmationDialogIconProps): ConfirmationDialogIconProps | null => {
        if (!icon) {
            return null;
        }

        if (isValidElement(icon)) {
            const { className = '', ...restProps } = icon.props;

            return cloneElement(icon, {
                ...restProps,
                className: cn('size-4', className),
            });
        }

        return icon;
    };

    return (
        <Dialog
            onOpenChange={handleOpenChange}
            open={isOpen}
        >
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                    <DialogDescription>{defaultDescription}</DialogDescription>
                </DialogHeader>

                <DialogFooter>
                    <Button
                        onClick={() => handleOpenChange(false)}
                        variant={cancelVariant}
                    >
                        {processIcon(cancelIcon)}
                        {cancelText}
                    </Button>
                    <Button
                        onClick={() => {
                            handleConfirm();
                            handleOpenChange(false);
                        }}
                        variant={confirmVariant}
                    >
                        {processIcon(confirmIcon)}
                        {confirmText}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default ConfirmationDialog;
