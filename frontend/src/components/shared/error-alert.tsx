import { AlertCircle } from 'lucide-react';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';

interface ErrorAlertProps {
    message?: null | string;
    title: string;
}

export function ErrorAlert({ message, title }: ErrorAlertProps) {
    return (
        <Alert variant="destructive">
            <AlertCircle className="size-4" />
            <AlertTitle>{title}</AlertTitle>
            {message ? <AlertDescription>{message}</AlertDescription> : null}
        </Alert>
    );
}
