import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Trash2 } from 'lucide-react';

interface DeleteConfirmationDialogProps {
    isOpen: boolean;
    handleOpenChange: (isOpen: boolean) => void;
    handleConfirm: () => void;
    title?: string;
    description?: string;
    itemName?: string;
    itemType?: string;
}

const DeleteConfirmationDialog = ({
    isOpen,
    handleOpenChange,
    handleConfirm,
    title = 'Confirm Deletion',
    description,
    itemName = 'this',
    itemType = 'item',
}: DeleteConfirmationDialogProps) => {
    const defaultDescription = description || (
        <>
            Are you sure you want to delete <strong className="font-semibold text-foreground">{itemName}</strong>{' '}
            {itemType}?
        </>
    );

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
                        variant="outline"
                        onClick={() => handleOpenChange(false)}
                    >
                        Cancel
                    </Button>
                    <Button
                        variant="destructive"
                        onClick={() => {
                            handleConfirm();
                            handleOpenChange(false);
                        }}
                    >
                        <Trash2 className="h-4 w-4" />
                        Delete
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
};

export default DeleteConfirmationDialog;
