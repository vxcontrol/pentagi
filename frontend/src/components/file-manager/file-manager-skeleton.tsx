import { type CSSProperties } from 'react';

import { Skeleton } from '@/components/ui/skeleton';
import { cn } from '@/lib/utils';

import type { FileManagerColumnsConfig } from './file-manager-types';

import { buildFileManagerGridTemplate } from './file-manager-utils';

interface FileManagerSkeletonRow {
    depth: number;
    isDir: boolean;
    nameWidth: string;
    sizeWidth: string;
}

const SKELETON_ROWS: readonly FileManagerSkeletonRow[] = [
    { depth: 0, isDir: true, nameWidth: 'w-28', sizeWidth: 'w-10' },
    { depth: 1, isDir: false, nameWidth: 'w-44', sizeWidth: 'w-12' },
    { depth: 1, isDir: false, nameWidth: 'w-36', sizeWidth: 'w-10' },
    { depth: 1, isDir: false, nameWidth: 'w-28', sizeWidth: 'w-14' },
    { depth: 0, isDir: true, nameWidth: 'w-36', sizeWidth: 'w-10' },
    { depth: 0, isDir: false, nameWidth: 'w-24', sizeWidth: 'w-14' },
    { depth: 0, isDir: false, nameWidth: 'w-32', sizeWidth: 'w-12' },
];

interface FileManagerSkeletonProps {
    columns?: FileManagerColumnsConfig;
    hasActions?: boolean;
    isCheckboxVisible?: boolean;
}

export const FileManagerSkeleton = ({
    columns,
    hasActions = false,
    isCheckboxVisible = false,
}: FileManagerSkeletonProps) => {
    const isSizeVisible = columns?.isSizeVisible ?? true;
    const isModifiedVisible = columns?.isModifiedVisible ?? true;
    const gridTemplate = buildFileManagerGridTemplate(isSizeVisible, isModifiedVisible, hasActions);

    return (
        <div className="bg-card flex flex-col overflow-hidden rounded-lg border">
            <div
                className="bg-muted/30 grid items-center gap-3 border-b px-3 py-2"
                style={{ gridTemplateColumns: gridTemplate }}
            >
                {isCheckboxVisible ? (
                    <Skeleton className="size-4 shrink-0 rounded-sm" />
                ) : (
                    <span
                        aria-hidden="true"
                        className="size-4"
                    />
                )}
                <Skeleton className="h-3 w-12" />
                {isSizeVisible && <Skeleton className="h-3 w-10" />}
                {isModifiedVisible && <Skeleton className="h-3 w-16" />}
                {hasActions && (
                    <span
                        aria-hidden="true"
                        className="size-7"
                    />
                )}
            </div>

            <div className="flex flex-col py-1">
                {SKELETON_ROWS.map((row, index) => {
                    const rowStyle = {
                        '--fm-depth': row.depth,
                        gridTemplateColumns: gridTemplate,
                    } as CSSProperties & Record<'--fm-depth', number>;

                    return (
                        <div
                            className="grid items-center gap-3 border border-transparent px-3 py-1.5"
                            key={index}
                            style={rowStyle}
                        >
                            {isCheckboxVisible ? (
                                <Skeleton className="size-4 shrink-0 rounded-sm" />
                            ) : (
                                <span
                                    aria-hidden="true"
                                    className="size-4"
                                />
                            )}

                            <div className="flex min-w-0 items-center gap-1.5 pl-[calc(var(--fm-depth)*16px)]">
                                {row.isDir ? (
                                    <Skeleton className="-mx-0.5 size-4 shrink-0 rounded-sm" />
                                ) : (
                                    <span
                                        aria-hidden="true"
                                        className="-mx-0.5 size-4 shrink-0"
                                    />
                                )}
                                <Skeleton className="size-4 shrink-0 rounded-sm" />
                                <Skeleton className={cn('h-4', row.nameWidth)} />
                            </div>

                            {isSizeVisible && (
                                <Skeleton
                                    className={cn('h-3 shrink-0', row.isDir ? 'w-0 opacity-0' : row.sizeWidth)}
                                />
                            )}
                            {isModifiedVisible && <Skeleton className="h-3 w-16 shrink-0" />}
                            {hasActions && (
                                <span
                                    aria-hidden="true"
                                    className="size-7"
                                />
                            )}
                        </div>
                    );
                })}
            </div>
        </div>
    );
};
