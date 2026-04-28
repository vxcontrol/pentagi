import {
    File,
    FileArchive,
    FileAudio,
    FileCode,
    FileImage,
    FileJson,
    FileSpreadsheet,
    FileText,
    FileType,
    FileVideo,
    Folder,
    FolderOpen,
    type LucideIcon,
} from 'lucide-react';

export interface FileTypeIcon {
    icon: LucideIcon;
    tone: string;
}

interface GetFileTypeIconArgs {
    isDir: boolean;
    isOpen?: boolean;
    name: string;
}

const EXT_MAP: Record<string, FileTypeIcon> = {
    '7z': { icon: FileArchive, tone: 'text-amber-500' },
    aac: { icon: FileAudio, tone: 'text-fuchsia-500' },
    avi: { icon: FileVideo, tone: 'text-violet-500' },
    bash: { icon: FileCode, tone: 'text-green-500' },
    bz2: { icon: FileArchive, tone: 'text-amber-500' },
    c: { icon: FileCode, tone: 'text-blue-400' },
    cfg: { icon: FileText, tone: 'text-muted-foreground' },
    conf: { icon: FileText, tone: 'text-muted-foreground' },
    cpp: { icon: FileCode, tone: 'text-blue-400' },
    cs: { icon: FileCode, tone: 'text-emerald-500' },
    css: { icon: FileCode, tone: 'text-sky-500' },
    csv: { icon: FileSpreadsheet, tone: 'text-emerald-500' },
    doc: { icon: FileText, tone: 'text-blue-500' },
    docx: { icon: FileText, tone: 'text-blue-500' },
    env: { icon: FileText, tone: 'text-muted-foreground' },
    flac: { icon: FileAudio, tone: 'text-fuchsia-500' },
    flv: { icon: FileVideo, tone: 'text-violet-500' },
    gif: { icon: FileImage, tone: 'text-purple-500' },
    go: { icon: FileCode, tone: 'text-cyan-500' },
    gz: { icon: FileArchive, tone: 'text-amber-500' },
    h: { icon: FileCode, tone: 'text-blue-400' },
    hpp: { icon: FileCode, tone: 'text-blue-400' },
    htm: { icon: FileCode, tone: 'text-orange-500' },
    html: { icon: FileCode, tone: 'text-orange-500' },
    ico: { icon: FileImage, tone: 'text-purple-500' },
    ini: { icon: FileText, tone: 'text-muted-foreground' },
    java: { icon: FileCode, tone: 'text-orange-500' },
    jpeg: { icon: FileImage, tone: 'text-purple-500' },
    jpg: { icon: FileImage, tone: 'text-purple-500' },
    js: { icon: FileCode, tone: 'text-yellow-500' },
    json: { icon: FileJson, tone: 'text-yellow-500' },
    jsx: { icon: FileCode, tone: 'text-blue-500' },
    kt: { icon: FileCode, tone: 'text-purple-500' },
    log: { icon: FileText, tone: 'text-muted-foreground' },
    md: { icon: FileText, tone: 'text-slate-400' },
    mdx: { icon: FileText, tone: 'text-slate-400' },
    mkv: { icon: FileVideo, tone: 'text-violet-500' },
    mov: { icon: FileVideo, tone: 'text-violet-500' },
    mp3: { icon: FileAudio, tone: 'text-fuchsia-500' },
    mp4: { icon: FileVideo, tone: 'text-violet-500' },
    ogg: { icon: FileAudio, tone: 'text-fuchsia-500' },
    pdf: { icon: FileType, tone: 'text-red-500' },
    php: { icon: FileCode, tone: 'text-indigo-500' },
    png: { icon: FileImage, tone: 'text-purple-500' },
    ppt: { icon: FileText, tone: 'text-orange-600' },
    pptx: { icon: FileText, tone: 'text-orange-600' },
    py: { icon: FileCode, tone: 'text-yellow-500' },
    rar: { icon: FileArchive, tone: 'text-amber-500' },
    rb: { icon: FileCode, tone: 'text-red-500' },
    rs: { icon: FileCode, tone: 'text-orange-500' },
    sh: { icon: FileCode, tone: 'text-green-500' },
    sql: { icon: FileCode, tone: 'text-cyan-500' },
    svg: { icon: FileImage, tone: 'text-pink-500' },
    swift: { icon: FileCode, tone: 'text-orange-500' },
    tar: { icon: FileArchive, tone: 'text-amber-500' },
    toml: { icon: FileJson, tone: 'text-rose-500' },
    ts: { icon: FileCode, tone: 'text-blue-500' },
    tsx: { icon: FileCode, tone: 'text-blue-500' },
    txt: { icon: FileText, tone: 'text-muted-foreground' },
    wav: { icon: FileAudio, tone: 'text-fuchsia-500' },
    webm: { icon: FileVideo, tone: 'text-violet-500' },
    webp: { icon: FileImage, tone: 'text-purple-500' },
    xls: { icon: FileSpreadsheet, tone: 'text-emerald-500' },
    xlsx: { icon: FileSpreadsheet, tone: 'text-emerald-500' },
    xml: { icon: FileCode, tone: 'text-orange-400' },
    yaml: { icon: FileJson, tone: 'text-rose-500' },
    yml: { icon: FileJson, tone: 'text-rose-500' },
    zip: { icon: FileArchive, tone: 'text-amber-500' },
};

export const getFileTypeIcon = ({ isDir, isOpen, name }: GetFileTypeIconArgs): FileTypeIcon => {
    if (isDir) {
        return { icon: isOpen ? FolderOpen : Folder, tone: 'text-blue-400' };
    }

    const ext = name.split('.').pop()?.toLowerCase() ?? '';

    return EXT_MAP[ext] ?? { icon: File, tone: 'text-muted-foreground' };
};
