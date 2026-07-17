import { AlertCircle } from 'lucide-react';

import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';

interface DataLoadErrorProps {
    message?: string;
    title: string;
}

export function DataLoadError({ message, title }: DataLoadErrorProps) {
    return (
        <Alert variant="destructive">
            <AlertCircle className="size-4" />
            <AlertTitle>{title}</AlertTitle>
            {message ? <AlertDescription>{message}</AlertDescription> : null}
        </Alert>
    );
}
