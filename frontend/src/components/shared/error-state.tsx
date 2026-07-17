import { AlertCircle } from 'lucide-react';

import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';

interface ErrorStateProps {
    message?: null | string;
    title: string;
}

export function ErrorState({ message, title }: ErrorStateProps) {
    return (
        <Empty>
            <EmptyHeader>
                <EmptyMedia>
                    <AlertCircle className="text-destructive size-12" />
                </EmptyMedia>
                <EmptyTitle>{title}</EmptyTitle>
                {message ? <EmptyDescription>{message}</EmptyDescription> : null}
            </EmptyHeader>
        </Empty>
    );
}
