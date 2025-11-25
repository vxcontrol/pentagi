// @ts-ignore - html2pdf.js doesn't have types
import GithubSlugger from 'github-slugger';
import html2pdf from 'html2pdf.js';
import { createElement } from 'react';
import { renderToString } from 'react-dom/server';

import type { FlowFragmentFragment, TaskFragmentFragment } from '@/graphql/types';

import Markdown from '@/components/shared/markdown';
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

// Helper function to convert markdown to HTML using our existing Markdown component
const convertMarkdownToHTML = (markdownContent: string): string => {
    // Use our existing Markdown component with light theme
    const markdownElement = createElement(Markdown, {
        children: markdownContent,
        className: 'prose prose-sm max-w-none', // Force light theme, no dark mode
    });

    // Render to HTML string
    const renderedHTML = renderToString(markdownElement);

    return renderedHTML;
};

// Get the current application's CSS styles for PDF
const getApplicationStyles = (): string => {
    // Extract styles from the current document
    const styleSheets = Array.from(document.styleSheets);
    let allStyles = '';

    try {
        styleSheets.forEach((styleSheet) => {
            try {
                if (styleSheet.cssRules) {
                    Array.from(styleSheet.cssRules).forEach((rule) => {
                        allStyles += `${rule.cssText}\n`;
                    });
                }
            } catch (e) {
                // Handle cross-origin stylesheets or other access issues
                console.warn('Could not access stylesheet rules:', e);
            }
        });
    } catch (error) {
        console.warn('Error extracting styles:', error);
    }

    // Add base PDF-specific styles
    const pdfSpecificStyles = `
        /* PDF-specific overrides */
        body {
            background: white !important;
            color: #374151 !important;
            font-size: 14px !important;
            line-height: 1.6 !important;
            padding: 20px !important;
            max-width: none !important;
        }

        /* Ensure light theme for all elements */
        * {
            background-color: inherit !important;
            color: inherit !important;
        }

        /* Force light theme for code blocks */
        .hljs {
            background: #f8f9fa !important;
            color: #212529 !important;
            border: 1px solid #e9ecef !important;
            border-radius: 0.375rem !important;
        }

        /* Print optimizations */
        @media print {
            body {
                padding: 16px !important;
                font-size: 12px !important;
            }

            h1, h2, h3, h4, h5, h6 {
                break-after: avoid !important;
                page-break-after: avoid !important;
            }

            pre, blockquote, table {
                break-inside: avoid !important;
                page-break-inside: avoid !important;
            }

            tr {
                break-inside: avoid !important;
                page-break-inside: avoid !important;
            }
        }
    `;

    return allStyles + pdfSpecificStyles;
};

// Helper function to create complete HTML document for PDF
const createPDFDocument = (htmlContent: string): string => {
    const styles = getApplicationStyles();

    return `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PentAGI Report</title>
    <style>
        ${styles}
    </style>
</head>
<body>
    <div class="prose prose-sm max-w-none">
        ${htmlContent}
    </div>
</body>
</html>`;
};

// Base function to generate PDF with configurable output
const generatePDF = async (
    content: string,
    options: {
        filename?: string;
        outputType: 'blob' | 'save';
    },
): Promise<Blob | void> => {
    try {
        // Convert markdown to HTML using our Markdown component
        const htmlContent = convertMarkdownToHTML(content);

        // Create complete HTML document with current app styles
        const fullHTML = createPDFDocument(htmlContent);

        // Configure html2pdf options for high quality
        const pdfOptions = {
            filename: options.filename,
            html2canvas: {
                allowTaint: false,
                backgroundColor: '#ffffff',
                letterRendering: true,
                scale: 2,
                useCORS: true,
            },
            image: { quality: 0.98, type: 'jpeg' as const },
            jsPDF: {
                compress: true,
                format: 'a4',
                orientation: 'portrait' as const,
                unit: 'mm',
            },
            margin: [5, 5, 5, 5] as [number, number, number, number], // top, right, bottom, left in mm
            pagebreak: { mode: ['avoid-all', 'css', 'legacy'] },
        };

        const pdf = html2pdf().set(pdfOptions).from(fullHTML);

        if (options.outputType === 'save') {
            await pdf.save();
        } else {
            return await pdf.outputPdf('blob');
        }
    } catch (error) {
        Log.error('Failed to generate PDF:', error);
        throw error;
    }
};

// Main function to generate PDF from markdown content and download it
export const generatePDFFromMarkdown = async (content: string, fileName: string): Promise<void> => {
    await generatePDF(content, {
        filename: fileName,
        outputType: 'save',
    });
};

// Function to generate PDF as blob for viewing (not used)
export const generatePDFBlob = async (content: string): Promise<Blob> => {
    const blob = await generatePDF(content, {
        outputType: 'blob',
    });

    if (!blob) {
        throw new Error('Failed to generate PDF blob');
    }

    return blob;
};
