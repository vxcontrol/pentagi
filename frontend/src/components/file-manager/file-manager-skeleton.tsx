import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';

const ROOT_NAME_WIDTHS = ['w-28', 'w-36', 'w-24', 'w-32'];
const NESTED_NAME_WIDTHS = ['w-44', 'w-36', 'w-28'];

export const FileManagerSkeleton = () => (
    <div className="bg-card overflow-hidden rounded-lg border">
        <div className="bg-muted/30 flex items-center gap-3 border-b px-3 py-2">
            <Skeleton className="size-4 shrink-0 rounded-sm" />
            <Skeleton className="h-3 w-12" />
            <div className="ml-auto flex items-center gap-3">
                <Skeleton className="h-3 w-10" />
                <Skeleton className="h-3 w-16" />
                <div className="size-7" />
            </div>
        </div>
        <div className="flex flex-col gap-px py-1">
            <div className="flex items-center gap-3 px-3 py-1.5">
                <Skeleton className="size-4 shrink-0 rounded-sm" />
                <Skeleton className="size-3.5 shrink-0" />
                <Skeleton className="size-4 shrink-0" />
                <Skeleton className="h-4 w-20" />
            </div>
            {NESTED_NAME_WIDTHS.map((width, index) => (
                <div
                    className="flex items-center gap-3 px-3 py-1.5 pl-9"
                    key={`nested-${index}`}
                >
                    <Skeleton className="size-4 shrink-0 rounded-sm" />
                    <Skeleton className="size-4 shrink-0" />
                    <Skeleton className={cn('h-4', width)} />
                    <Skeleton className="ml-auto h-3 w-10" />
                    <Skeleton className="h-3 w-16" />
                </div>
            ))}
            {ROOT_NAME_WIDTHS.map((width, index) => (
                <div
                    className="flex items-center gap-3 px-3 py-1.5"
                    key={`root-${index}`}
                >
                    <Skeleton className="size-4 shrink-0 rounded-sm" />
                    <Skeleton className="size-3.5 shrink-0" />
                    <Skeleton className="size-4 shrink-0" />
                    <Skeleton className={cn('h-4', width)} />
                    <Skeleton className="ml-auto h-3 w-10" />
                    <Skeleton className="h-3 w-16" />
                </div>
            ))}
        </div>
    </div>
);
