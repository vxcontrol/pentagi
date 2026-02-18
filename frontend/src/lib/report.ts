import GithubSlugger from 'github-slugger';

import type { FlowFragmentFragment, TaskFragmentFragment } from '@/graphql/types';

import { StatusType } from '@/graphql/types';

import { Log } from './log';

// Helper function to get emoji for status
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

// Helper function to shift markdown headers by specified levels
const shiftMarkdownHeaders = (text: string, shiftBy: number): string => {
    return text.replaceAll(/^(#{1,6})\s+(.+)$/gm, (match, hashes, content) => {
        const currentLevel = hashes.length;
        const newLevel = Math.min(currentLevel + shiftBy, 6); // Max level is 6
        const newHashes = '#'.repeat(newLevel);

        return `${newHashes} ${content}`;
    });
};

// Helper function to create anchor link from text using the same algorithm as rehype-slug
const createAnchor = (text: string): string => {
    const slugger = new GithubSlugger();

    return slugger.slug(text);
};

// Helper function to generate table of contents
const generateTableOfContents = (tasks: TaskFragmentFragment[], flow?: FlowFragmentFragment | null): string => {
    let toc = '';

    // Add flow header as H1 if flow data is available
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
        // Create anchor from the same text that will be used in the heading (including emoji)
        const taskAnchor = createAnchor(`${taskEmoji} ${task.id}. ${task.title}`);

        toc += `- [${taskTitle}](#${taskAnchor})\n`;

        // Add subtasks to TOC (removed input headers from TOC)
        if (task.subtasks && task.subtasks.length > 0) {
            const sortedSubtasks = [...task.subtasks].sort((a, b) => +a.id - +b.id);

            sortedSubtasks.forEach((subtask) => {
                const subtaskEmoji = getStatusEmoji(subtask.status);
                const subtaskTitle = `${subtaskEmoji} ${subtask.id}. ${subtask.title}`;
                // Create anchor from the same text that will be used in the heading (including emoji)
                const subtaskAnchor = createAnchor(`${subtaskEmoji} ${subtask.id}. ${subtask.title}`);
                toc += `  - [${subtaskTitle}](#${subtaskAnchor})\n`;
            });
        }
    });

    return `${toc}\n---\n\n`;
};

// Helper function to generate report content
export const generateReport = (tasks: TaskFragmentFragment[], flow?: FlowFragmentFragment | null): string => {
    if (!tasks || tasks.length === 0) {
        if (flow) {
            const flowEmoji = getStatusEmoji(flow.status);

            return `# ${flowEmoji} ${flow.id}. ${flow.title}\n\nNo tasks available for this flow.`;
        }

        return 'No tasks available for this flow.';
    }

    const sortedTasks = [...tasks].sort((a, b) => +a.id - +b.id);

    // Generate table of contents with flow header
    let report = generateTableOfContents(tasks, flow);

    sortedTasks.forEach((task, taskIndex) => {
        // Add task title with status emoji and ID (now H3 since H1 is flow, H2 is TOC)
        const taskEmoji = getStatusEmoji(task.status);
        report += `### ${taskEmoji} ${task.id}. ${task.title}\n\n`;

        // Add task input with shifted headers (shift by 3 levels: H1→H4, H2→H5, etc.)
        if (task.input?.trim()) {
            const shiftedInput = shiftMarkdownHeaders(task.input, 3);
            report += `${shiftedInput}\n\n`;
        }

        // Add separator and task result if not empty
        if (task.result?.trim()) {
            report += `---\n\n${task.result}\n\n`;
        }

        // Add subtasks (now H4 since tasks are H3)
        if (task.subtasks && task.subtasks.length > 0) {
            const sortedSubtasks = [...task.subtasks].sort((a, b) => +a.id - +b.id);

            sortedSubtasks.forEach((subtask) => {
                const subtaskEmoji = getStatusEmoji(subtask.status);
                report += `#### ${subtaskEmoji} ${subtask.id}. ${subtask.title}\n\n`;

                // Add subtask description
                if (subtask.description?.trim()) {
                    report += `${subtask.description}\n\n`;
                }

                // Add subtask result with separator if not empty
                if (subtask.result?.trim()) {
                    report += `---\n\n${subtask.result}\n\n`;
                }
            });
        }

        // Add separator between tasks (except for the last one)
        if (taskIndex < sortedTasks.length - 1) {
            report += '---\n\n';
        }
    });

    return report.trim();
};

export const generateFileName = (flow: FlowFragmentFragment): string => {
    const flowId = flow.id;
    const flowTitle = flow.title
        // Replace any invalid file name characters and whitespace with underscore
        .replaceAll(/[^\w\s.-]/g, '_')
        // Replace spaces, non-breaking spaces, and line breaks with underscore
        .replaceAll(/[\s\u2000-\u200B]+/g, '_')
        // Convert to lowercase
        .toLowerCase()
        // Trim to 150 characters
        .slice(0, 150)
        // Remove trailing underscores
        .replace(/_+$/, '');

    // DATETIME in format YYYYMMDDHHMMSS
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

// Helper function to download text content as file
export const downloadTextFile = (content: string, fileName: string, mimeType = 'text/plain'): void => {
    try {
        // Create blob with content
        const blob = new Blob([content], { type: mimeType });

        // Create temporary URL
        const url = URL.createObjectURL(blob);

        // Create temporary download link
        const link = document.createElement('a');
        link.href = url;
        link.download = fileName;
        link.style.display = 'none';

        // Add to DOM, click, and remove
        document.body.append(link);
        link.click();
        link.remove();

        // Clean up URL
        URL.revokeObjectURL(url);
    } catch (error) {
        Log.error('Failed to download file:', error);
        throw error;
    }
};

// Helper function to copy text to clipboard
export const copyToClipboard = async (text: string): Promise<boolean> => {
    try {
        await navigator.clipboard.writeText(text);

        return true;
    } catch (error) {
        Log.error('Failed to copy to clipboard:', error);

        return false;
    }
};

// Export new PDF generation functions from report-pdf.tsx
export {
    generatePDFBlobNew as generatePDFBlob,
    generatePDFFromMarkdownNew as generatePDFFromMarkdown,
} from './report-pdf';
