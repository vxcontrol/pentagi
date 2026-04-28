import {
    ChevronRight,
    Download,
    Edit2,
    File,
    FileText,
    FolderOpen,
    Image,
    MoreVertical,
    Move,
    Sheet,
    Trash2,
    Upload,
} from 'lucide-react';
import { useState } from 'react';

import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

interface FileItem {
    folder?: string;
    id: string;
    modified: string;
    name: string;
    sharedWith: number;
    size?: string;
    type: 'document' | 'file' | 'folder' | 'image' | 'pdf' | 'spreadsheet';
}

const files: FileItem[] = [
    {
        id: 'f1',
        modified: '2 days ago',
        name: 'Documents',
        sharedWith: 0,
        type: 'folder',
    },
    {
        id: 'f2',
        modified: '1 week ago',
        name: 'Images',
        sharedWith: 3,
        type: 'folder',
    },
    {
        id: 'f3',
        modified: '3 hours ago',
        name: 'Q4_Report_2025.pdf',
        sharedWith: 5,
        size: '2.4 MB',
        type: 'pdf',
    },
    {
        id: 'f4',
        modified: 'Yesterday',
        name: 'Product_Mockup.png',
        sharedWith: 2,
        size: '1.8 MB',
        type: 'image',
    },
    {
        id: 'f5',
        modified: '5 hours ago',
        name: 'Meeting_Notes.docx',
        sharedWith: 8,
        size: '145 KB',
        type: 'document',
    },
    {
        id: 'f6',
        modified: '2 days ago',
        name: 'Budget_2026.xlsx',
        sharedWith: 4,
        size: '892 KB',
        type: 'spreadsheet',
    },
    {
        id: 'f7',
        modified: '1 week ago',
        name: 'Presentation.pptx',
        sharedWith: 12,
        size: '5.2 MB',
        type: 'document',
    },
    {
        id: 'f8',
        modified: '3 weeks ago',
        name: 'Archive',
        sharedWith: 0,
        type: 'folder',
    },
];

const getFileIcon = (type: FileItem['type']) => {
    switch (type) {
        case 'document':
            return FileText;
        case 'folder':
            return FolderOpen;
        case 'image':
            return Image;
        case 'pdf':
            return FileText;
        case 'spreadsheet':
            return Sheet;
        default:
            return File;
    }
};

const getTypeBadge = (type: FileItem['type']) => {
    switch (type) {
        case 'document':
            return 'Document';
        case 'image':
            return 'Image';
        case 'pdf':
            return 'PDF';
        case 'spreadsheet':
            return 'Spreadsheet';
        default:
            return null;
    }
};

export default function CrudFileManager() {
    const [selected, setSelected] = useState<string[]>([]);
    const [path] = useState(['My Files']);

    const toggleSelect = (id: string) => {
        setSelected((prev) => (prev.includes(id) ? prev.filter((i) => i !== id) : [...prev, id]));
    };

    const toggleSelectAll = () => {
        setSelected((prev) => (prev.length === files.length ? [] : files.map((f) => f.id)));
    };

    const handleBulkDelete = () => {
        setSelected([]);
    };

    return (
        <section className="mx-auto w-full max-w-4xl p-4">
            <div className="bg-card overflow-hidden rounded-lg border">
                {/* Header with breadcrumb and upload */}
                <div className="flex items-center justify-between border-b px-4 py-3">
                    <div className="flex items-center gap-1 text-sm">
                        <button
                            className="text-muted-foreground hover:text-foreground transition-colors"
                            type="button"
                        >
                            Home
                        </button>
                        {path.map((segment, i) => (
                            <div
                                className="flex items-center gap-1"
                                key={segment}
                            >
                                <ChevronRight className="text-muted-foreground size-4" />
                                <button
                                    className={
                                        i === path.length - 1
                                            ? 'text-foreground font-medium'
                                            : 'text-muted-foreground hover:text-foreground transition-colors'
                                    }
                                    type="button"
                                >
                                    {segment}
                                </button>
                            </div>
                        ))}
                    </div>
                    <Button
                        className="h-7 gap-1.5 text-xs"
                        size="sm"
                        variant="outline"
                    >
                        <Upload className="size-3.5" />
                        Upload
                    </Button>
                </div>

                {/* Bulk actions bar */}
                {selected.length > 0 && (
                    <div className="bg-muted/50 flex items-center justify-between border-b px-4 py-2">
                        <span className="text-muted-foreground text-sm">{selected.length} selected</span>
                        <Button
                            className="h-7 gap-1.5 text-xs text-red-600 hover:text-red-700"
                            onClick={handleBulkDelete}
                            size="sm"
                            variant="ghost"
                        >
                            <Trash2 className="size-3.5" />
                            Delete
                        </Button>
                    </div>
                )}

                {/* Column headers */}
                <div className="bg-muted/30 text-muted-foreground grid grid-cols-[auto_1fr_auto_auto_auto_auto] items-center gap-4 border-b px-4 py-2 text-xs font-medium">
                    <Checkbox
                        checked={selected.length === files.length}
                        onCheckedChange={toggleSelectAll}
                    />
                    <span>Name</span>
                    <span>Type</span>
                    <span>Size</span>
                    <span>Modified</span>
                    <span className="w-8" />
                </div>

                {/* File list */}
                <div>
                    {files.map((file) => {
                        const Icon = getFileIcon(file.type);
                        const badge = getTypeBadge(file.type);

                        return (
                            <div
                                className="hover:bg-muted/50 grid grid-cols-[auto_1fr_auto_auto_auto_auto] items-center gap-4 border-b px-4 py-3 transition-colors last:border-b-0"
                                key={file.id}
                            >
                                <Checkbox
                                    checked={selected.includes(file.id)}
                                    onCheckedChange={() => toggleSelect(file.id)}
                                />

                                <div className="flex min-w-0 items-center gap-2.5">
                                    <Icon
                                        className={`size-4 shrink-0 ${
                                            file.type === 'folder' ? 'text-blue-500' : 'text-muted-foreground'
                                        }`}
                                    />
                                    <div className="min-w-0">
                                        <p className="truncate text-sm font-medium">{file.name}</p>
                                        {file.sharedWith > 0 && (
                                            <p className="text-muted-foreground text-xs">
                                                Shared with {file.sharedWith}
                                            </p>
                                        )}
                                    </div>
                                </div>

                                <div>
                                    {badge && (
                                        <span className="bg-muted text-muted-foreground inline-flex rounded-md px-2 py-0.5 text-xs">
                                            {badge}
                                        </span>
                                    )}
                                </div>

                                <span className="text-muted-foreground text-sm">{file.size || '—'}</span>

                                <span className="text-muted-foreground text-sm">{file.modified}</span>

                                {file.type !== 'folder' && (
                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <Button
                                                className="size-8 p-0"
                                                size="sm"
                                                variant="ghost"
                                            >
                                                <MoreVertical className="size-4" />
                                            </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent align="end">
                                            <DropdownMenuItem>
                                                <Download className="mr-2 size-4" />
                                                Download
                                            </DropdownMenuItem>
                                            <DropdownMenuItem>
                                                <Edit2 className="mr-2 size-4" />
                                                Rename
                                            </DropdownMenuItem>
                                            <DropdownMenuItem>
                                                <Move className="mr-2 size-4" />
                                                Move
                                            </DropdownMenuItem>
                                            <DropdownMenuItem className="text-red-600">
                                                <Trash2 className="mr-2 size-4" />
                                                Delete
                                            </DropdownMenuItem>
                                        </DropdownMenuContent>
                                    </DropdownMenu>
                                )}
                                {file.type === 'folder' && (
                                    <DropdownMenu>
                                        <DropdownMenuTrigger asChild>
                                            <Button
                                                className="size-8 p-0"
                                                size="sm"
                                                variant="ghost"
                                            >
                                                <MoreVertical className="size-4" />
                                            </Button>
                                        </DropdownMenuTrigger>
                                        <DropdownMenuContent align="end">
                                            <DropdownMenuItem>
                                                <Edit2 className="mr-2 size-4" />
                                                Rename
                                            </DropdownMenuItem>
                                            <DropdownMenuItem className="text-red-600">
                                                <Trash2 className="mr-2 size-4" />
                                                Delete
                                            </DropdownMenuItem>
                                        </DropdownMenuContent>
                                    </DropdownMenu>
                                )}
                            </div>
                        );
                    })}
                </div>
            </div>
        </section>
    );
}
