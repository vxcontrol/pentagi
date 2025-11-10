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
import { Trash2 } from 'lucide-react';
import { type ReactElement, cloneElement, isValidElement } from 'react';

type ConfirmationDialogIconProps = ReactElement<React.SVGProps<SVGSVGElement>>;

interface ConfirmationDialogProps {
    isOpen: boolean;
    handleOpenChange: (isOpen: boolean) => void;
    handleConfirm: () => void;
    title?: string;
    description?: string;
    itemName?: string;
    itemType?: string;
    cancelText?: string;
    confirmText?: string;
    cancelVariant?: 'outline' | 'ghost' | 'secondary' | 'default' | 'destructive';
    confirmVariant?: 'outline' | 'ghost' | 'secondary' | 'default' | 'destructive';
    confirmIcon?: ConfirmationDialogIconProps;
    cancelIcon?: ConfirmationDialogIconProps;
}

const ConfirmationDialog = ({
    isOpen,
    handleOpenChange,
    handleConfirm,
    title = 'Confirm Action',
    description,
    itemName = 'this',
    itemType = 'item',
    cancelText = 'Cancel',
    confirmText = 'Confirm',
    cancelVariant = 'outline',
    confirmVariant = 'destructive',
    confirmIcon = <Trash2 />,
    cancelIcon,
}: ConfirmationDialogProps) => {
    const defaultDescription = description || (
        <>
            Are you sure you want to perform this action on{' '}
            <strong className="font-semibold text-foreground">{itemName}</strong> {itemType}?
        </>
    );

    // Common method to process icons with h-4 w-4 classes
    const processIcon = (icon?: ConfirmationDialogIconProps): ConfirmationDialogIconProps | null => {
        if (!icon) return null;

        if (isValidElement(icon)) {
            const { className = '', ...restProps } = icon.props;
            return cloneElement(icon, {
                ...restProps,
                className: cn('h-4 w-4', className),
            });
        }

        return icon;
    };

    return (
        <Dialog
            open={isOpen}
            onOpenChange={handleOpenChange}
        >
            <DialogContent className="sm:max-w-md">
                <DialogHeader>
                    <DialogTitle>{title}</DialogTitle>
                    <DialogDescription>{defaultDescription}</DialogDescription>
                </DialogHeader>

                <DialogFooter>
                    <Button
                        variant={cancelVariant}
                        onClick={() => handleOpenChange(false)}
                    >
                        {processIcon(cancelIcon)}
                        {cancelText}
                    </Button>
                    <Button
                        variant={confirmVariant}
                        onClick={() => {
                            handleConfirm();
                            handleOpenChange(false);
                        }}
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
