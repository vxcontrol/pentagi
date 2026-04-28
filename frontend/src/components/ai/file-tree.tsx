import { ChevronRightIcon, FileIcon, FolderIcon, FolderOpenIcon } from 'lucide-react';
import { createContext, type HTMLAttributes, type ReactNode, useContext, useState } from 'react';

import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { cn } from '@/lib/utils';

interface FileTreeContextType {
    expandedPaths: Set<string>;
    onSelect?: (path: string) => void;
    selectedPath?: string;
    togglePath: (path: string) => void;
}

const FileTreeContext = createContext<FileTreeContextType>({
    expandedPaths: new Set(),
    togglePath: () => undefined,
});

export type FileTreeProps = HTMLAttributes<HTMLDivElement> & {
    defaultExpanded?: Set<string>;
    expanded?: Set<string>;
    onExpandedChange?: (expanded: Set<string>) => void;
    onSelect?: (path: string) => void;
    selectedPath?: string;
};

export const FileTree = ({
    children,
    className,
    defaultExpanded = new Set(),
    expanded: controlledExpanded,
    onExpandedChange,
    onSelect,
    selectedPath,
    ...props
}: FileTreeProps) => {
    const [internalExpanded, setInternalExpanded] = useState(defaultExpanded);
    const expandedPaths = controlledExpanded ?? internalExpanded;

    const togglePath = (path: string) => {
        const newExpanded = new Set(expandedPaths);

        if (newExpanded.has(path)) {
            newExpanded.delete(path);
        } else {
            newExpanded.add(path);
        }

        setInternalExpanded(newExpanded);
        onExpandedChange?.(newExpanded);
    };

    return (
        <FileTreeContext.Provider value={{ expandedPaths, onSelect, selectedPath, togglePath }}>
            <div
                className={cn('bg-background rounded-lg border font-mono text-sm', className)}
                role="tree"
                {...props}
            >
                <div className="p-2">{children}</div>
            </div>
        </FileTreeContext.Provider>
    );
};

interface FileTreeFolderContextType {
    isExpanded: boolean;
    name: string;
    path: string;
}

const FileTreeFolderContext = createContext<FileTreeFolderContextType>({
    isExpanded: false,
    name: '',
    path: '',
});

export type FileTreeFolderProps = HTMLAttributes<HTMLDivElement> & {
    name: string;
    path: string;
};

export const FileTreeFolder = ({ children, className, name, path, ...props }: FileTreeFolderProps) => {
    const { expandedPaths, onSelect, selectedPath, togglePath } = useContext(FileTreeContext);
    const isExpanded = expandedPaths.has(path);
    const isSelected = selectedPath === path;

    return (
        <FileTreeFolderContext.Provider value={{ isExpanded, name, path }}>
            <Collapsible
                onOpenChange={() => togglePath(path)}
                open={isExpanded}
            >
                <div
                    className={cn('', className)}
                    role="treeitem"
                    tabIndex={0}
                    {...props}
                >
                    <CollapsibleTrigger asChild>
                        <button
                            className={cn(
                                'hover:bg-muted/50 flex w-full items-center gap-1 rounded px-2 py-1 text-left transition-colors',
                                isSelected && 'bg-muted',
                            )}
                            onClick={() => onSelect?.(path)}
                            type="button"
                        >
                            <ChevronRightIcon
                                className={cn(
                                    'text-muted-foreground size-4 shrink-0 transition-transform',
                                    isExpanded && 'rotate-90',
                                )}
                            />
                            <FileTreeIcon>
                                {isExpanded ? (
                                    <FolderOpenIcon className="size-4 text-blue-500" />
                                ) : (
                                    <FolderIcon className="size-4 text-blue-500" />
                                )}
                            </FileTreeIcon>
                            <FileTreeName>{name}</FileTreeName>
                        </button>
                    </CollapsibleTrigger>
                    <CollapsibleContent>
                        <div className="ml-4 border-l pl-2">{children}</div>
                    </CollapsibleContent>
                </div>
            </Collapsible>
        </FileTreeFolderContext.Provider>
    );
};

interface FileTreeFileContextType {
    name: string;
    path: string;
}

const FileTreeFileContext = createContext<FileTreeFileContextType>({
    name: '',
    path: '',
});

export type FileTreeFileProps = HTMLAttributes<HTMLDivElement> & {
    icon?: ReactNode;
    name: string;
    path: string;
};

export const FileTreeFile = ({ children, className, icon, name, path, ...props }: FileTreeFileProps) => {
    const { onSelect, selectedPath } = useContext(FileTreeContext);
    const isSelected = selectedPath === path;

    return (
        <FileTreeFileContext.Provider value={{ name, path }}>
            <div
                className={cn(
                    'hover:bg-muted/50 flex cursor-pointer items-center gap-1 rounded px-2 py-1 transition-colors',
                    isSelected && 'bg-muted',
                    className,
                )}
                onClick={() => onSelect?.(path)}
                onKeyDown={(event) => {
                    if (event.key === 'Enter' || event.key === ' ') {
                        onSelect?.(path);
                    }
                }}
                role="treeitem"
                tabIndex={0}
                {...props}
            >
                {children ?? (
                    <>
                        <span className="size-4" />
                        <FileTreeIcon>{icon ?? <FileIcon className="text-muted-foreground size-4" />}</FileTreeIcon>
                        <FileTreeName>{name}</FileTreeName>
                    </>
                )}
            </div>
        </FileTreeFileContext.Provider>
    );
};

export type FileTreeIconProps = HTMLAttributes<HTMLSpanElement>;

export const FileTreeIcon = ({ children, className, ...props }: FileTreeIconProps) => (
    <span
        className={cn('shrink-0', className)}
        {...props}
    >
        {children}
    </span>
);

export type FileTreeNameProps = HTMLAttributes<HTMLSpanElement>;

export const FileTreeName = ({ children, className, ...props }: FileTreeNameProps) => (
    <span
        className={cn('truncate', className)}
        {...props}
    >
        {children}
    </span>
);

export type FileTreeActionsProps = HTMLAttributes<HTMLDivElement>;

export const FileTreeActions = ({ children, className, ...props }: FileTreeActionsProps) => (
    <div
        className={cn('ml-auto flex items-center gap-1', className)}
        onClick={(event) => event.stopPropagation()}
        onKeyDown={(event) => event.stopPropagation()}
        role="group"
        {...props}
    >
        {children}
    </div>
);
