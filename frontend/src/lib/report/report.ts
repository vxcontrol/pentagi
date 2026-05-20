import GithubSlugger from 'github-slugger';

import type { FlowFragmentFragment, TaskFragmentFragment } from '@/graphql/types';

import { StatusType } from '@/graphql/types';
import { Log } from '@/lib/log';

const getStatusEmoji = (status: StatusType): string => {
    switch (status) {
        case StatusType.Created: {
            return '📝';
        }

        case StatusType.Failed: {
            return '❌';
        }

        case StatusType.Finished: {
            return '✅';
        }

        case StatusType.Running: {
            return '⚡';
        }

        case StatusType.Waiting: {
            return '⏳';
        }

        default: {
            return '📝';
        }
    }
};

const shiftMarkdownHeaders = (text: string, shiftBy: number): string => {
    return text.replaceAll(/^(#{1,6})\s+(.+)$/gm, (match, hashes, content) => {
        const currentLevel = hashes.length;
        const newLevel = Math.min(currentLevel + shiftBy, 6);
        const newHashes = '#'.repeat(newLevel);

        return `${newHashes} ${content}`;
    });
};

const createAnchor = (text: string): string => {
    const slugger = new GithubSlugger();

    return slugger.slug(text);
};

const generateTableOfContents = (tasks: TaskFragmentFragment[], flow?: FlowFragmentFragment | null): string => {
    let toc = '';

    if (flow) {
        const flowEmoji = getStatusEmoji(flow.status);
        toc = `# ${flowEmoji} ${flow.id}. ${flow.title}\n\n`;
    }

    if (!tasks || tasks.length === 0) {
        return toc;
    }

    const sortedTasks = [...tasks].sort((a, b) => +a.id - +b.id);

    sortedTasks.forEach((task) => {
        const taskEmoji = getStatusEmoji(task.status);
        const taskTitle = `${taskEmoji} ${task.id}. ${task.title}`;
        // Anchor must be generated from the exact heading text (emoji included) — rehype-slug
        // computes the same slug from the rendered <h3>, so any mismatch breaks the link.
        const taskAnchor = createAnchor(`${taskEmoji} ${task.id}. ${task.title}`);

        toc += `- [${taskTitle}](#${taskAnchor})\n`;

        if (task.subtasks && task.subtasks.length > 0) {
            const sortedSubtasks = [...task.subtasks].sort((a, b) => +a.id - +b.id);

            sortedSubtasks.forEach((subtask) => {
                const subtaskEmoji = getStatusEmoji(subtask.status);
                const subtaskTitle = `${subtaskEmoji} ${subtask.id}. ${subtask.title}`;
                const subtaskAnchor = createAnchor(`${subtaskEmoji} ${subtask.id}. ${subtask.title}`);
                toc += `  - [${subtaskTitle}](#${subtaskAnchor})\n`;
            });
        }
    });

    return `${toc}\n---\n\n`;
};

export const generateReport = (tasks: TaskFragmentFragment[], flow?: FlowFragmentFragment | null): string => {
    if (!tasks || tasks.length === 0) {
        if (flow) {
            const flowEmoji = getStatusEmoji(flow.status);

            return `# ${flowEmoji} ${flow.id}. ${flow.title}\n\nNo tasks available for this flow.`;
        }

        return 'No tasks available for this flow.';
    }

    const sortedTasks = [...tasks].sort((a, b) => +a.id - +b.id);

    let report = generateTableOfContents(tasks, flow);

    sortedTasks.forEach((task, taskIndex) => {
        const taskEmoji = getStatusEmoji(task.status);
        report += `### ${taskEmoji} ${task.id}. ${task.title}\n\n`;

        // Shift the task's own input headings down by 3 so they slot below the H3 task title
        // (H1→H4, H2→H5, etc.) — keeps the report's outline consistent across sections.
        if (task.input?.trim()) {
            const shiftedInput = shiftMarkdownHeaders(task.input, 3);
            report += `${shiftedInput}\n\n`;
        }

        if (task.result?.trim()) {
            report += `---\n\n${task.result}\n\n`;
        }

        if (task.subtasks && task.subtasks.length > 0) {
            const sortedSubtasks = [...task.subtasks].sort((a, b) => +a.id - +b.id);

            sortedSubtasks.forEach((subtask) => {
                const subtaskEmoji = getStatusEmoji(subtask.status);
                report += `#### ${subtaskEmoji} ${subtask.id}. ${subtask.title}\n\n`;

                if (subtask.description?.trim()) {
                    report += `${subtask.description}\n\n`;
                }

                if (subtask.result?.trim()) {
                    report += `---\n\n${subtask.result}\n\n`;
                }
            });
        }

        if (taskIndex < sortedTasks.length - 1) {
            report += '---\n\n';
        }
    });

    return report.trim();
};

export const generateFileName = (flow: FlowFragmentFragment): string => {
    const flowId = flow.id;
    const flowTitle = flow.title
        .replaceAll(/[^\w\s.-]/g, '_')
        .replaceAll(/[\s\u2000-\u200B]+/g, '_')
        .toLowerCase()
        .slice(0, 150)
        .replace(/_+$/, '');

    const now = new Date();
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0');
    const day = String(now.getDate()).padStart(2, '0');
    const hours = String(now.getHours()).padStart(2, '0');
    const minutes = String(now.getMinutes()).padStart(2, '0');
    const seconds = String(now.getSeconds()).padStart(2, '0');

    const datetime = `${year}${month}${day}${hours}${minutes}${seconds}`;

    return `report_flow_${flowId}_${flowTitle}_${datetime}`;
};

export const downloadTextFile = (content: string, fileName: string, mimeType = 'text/plain'): void => {
    try {
        const blob = new Blob([content], { type: mimeType });
        const url = URL.createObjectURL(blob);

        const link = document.createElement('a');
        link.href = url;
        link.download = fileName;
        link.style.display = 'none';

        document.body.append(link);
        link.click();
        link.remove();

        URL.revokeObjectURL(url);
    } catch (error) {
        Log.error('Failed to download file:', error);
        throw error;
    }
};

export const copyToClipboard = async (text: string): Promise<boolean> => {
    try {
        await navigator.clipboard.writeText(text);

        return true;
    } catch (error) {
        Log.error('Failed to copy to clipboard:', error);

        return false;
    }
};

// Lazy-load the PDF generator so @react-pdf/renderer (~1.5 MB) is fetched
// only when the user actually triggers a PDF export, not on every page that
// imports report utilities (flow.tsx, flow-report.tsx).
export const generatePDFFromMarkdown = async (content: string, fileName: string): Promise<void> => {
    const { generatePDFFromMarkdownNew } = await import('./report-pdf');

    return generatePDFFromMarkdownNew(content, fileName);
};

export const generatePDFBlob = async (content: string): Promise<Blob> => {
    const { generatePDFBlobNew } = await import('./report-pdf');

    return generatePDFBlobNew(content);
};
