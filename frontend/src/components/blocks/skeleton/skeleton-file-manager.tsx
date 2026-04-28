import { Skeleton } from '@/components/ui/skeleton';

const fileRows = [
    { date: 'w-24', hasAction: true, name: 'w-44', size: 'w-14' },
    { date: 'w-20', hasAction: true, name: 'w-56', size: 'w-10' },
    { date: 'w-24', hasAction: true, name: 'w-36', size: 'w-12' },
    { date: 'w-20', hasAction: true, name: 'w-48', size: 'w-16' },
    { date: 'w-24', hasAction: true, name: 'w-40', size: 'w-10' },
    { date: 'w-20', hasAction: true, name: 'w-52', size: 'w-14' },
];

export default function SkeletonFileManager() {
    return (
        <section className="mx-auto w-full max-w-2xl p-4">
            <div className="bg-card overflow-hidden rounded-lg border">
                {/* Toolbar with breadcrumb and actions */}
                <div className="flex items-center justify-between border-b px-4 py-3">
                    <div className="flex items-center gap-1.5">
                        <Skeleton className="h-4 w-12" />
                        <Skeleton className="h-3 w-1.5" />
                        <Skeleton className="h-4 w-20" />
                        <Skeleton className="h-3 w-1.5" />
                        <Skeleton className="h-4 w-24" />
                    </div>
                    <div className="flex items-center gap-2">
                        <Skeleton className="size-7 rounded-md" />
                        <Skeleton className="size-7 rounded-md" />
                        <Skeleton className="h-7 w-20 rounded-md" />
                    </div>
                </div>

                {/* Column headers */}
                <div className="bg-muted/50 flex items-center gap-4 border-b px-4 py-2.5">
                    <div className="flex min-w-0 flex-1 items-center gap-3">
                        <Skeleton className="size-4 shrink-0" />
                        <Skeleton className="h-3 w-10" />
                    </div>
                    <Skeleton className="h-3 w-8 shrink-0" />
                    <Skeleton className="h-3 w-16 shrink-0" />
                    <div className="w-7 shrink-0" />
                </div>

                {/* File list */}
                <div>
                    {fileRows.map((row, i) => (
                        <div
                            className={`flex items-center gap-4 px-4 py-3 ${i < fileRows.length - 1 ? 'border-b' : ''}`}
                            key={i}
                        >
                            <div className="flex min-w-0 flex-1 items-center gap-3">
                                <Skeleton className="size-4 shrink-0" />
                                <Skeleton className={`h-4 ${row.name}`} />
                            </div>
                            <Skeleton className={`h-3 shrink-0 ${row.size}`} />
                            <Skeleton className={`h-3 shrink-0 ${row.date}`} />
                            <Skeleton className="size-7 shrink-0 rounded-md" />
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}
