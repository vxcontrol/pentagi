import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty';
import { Spinner } from '@/components/ui/spinner';

interface LoadingStateProps {
    description?: string;
    title: string;
}

export function LoadingState({ description, title }: LoadingStateProps) {
    return (
        <Empty>
            <EmptyHeader>
                <EmptyMedia>
                    <Spinner
                        className="text-muted-foreground size-10"
                        variant="circle"
                    />
                </EmptyMedia>
                <EmptyTitle>{title}</EmptyTitle>
                {description ? <EmptyDescription>{description}</EmptyDescription> : null}
            </EmptyHeader>
        </Empty>
    );
}
