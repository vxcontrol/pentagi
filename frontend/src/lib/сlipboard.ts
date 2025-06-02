import { Terminal as XTerminal } from '@xterm/xterm';
import { ResultFormat } from '@/graphql/types';

/**
 * Interface for message data that can be copied to clipboard
 */
export interface CopyableMessage {
    thinking?: string | null;
    message?: string | null;
    result?: string | null;
    resultFormat?: ResultFormat;
}

/**
 * Extracts clean text from terminal content using hidden terminal instance
 * This removes ANSI escape codes and returns formatted text as it appears in UI
 */
export const getCleanTerminalText = (terminalContent: string): Promise<string> => {
    return new Promise((resolve, reject) => {
        let hiddenTerminal: XTerminal | null = null;
        let hiddenDiv: HTMLDivElement | null = null;
        let timeoutId: NodeJS.Timeout | null = null;
        let safetyTimeoutId: NodeJS.Timeout | null = null;
        let isResolved = false;

        const cleanup = () => {
            // Clear timeouts if they exist
            if (timeoutId) {
                clearTimeout(timeoutId);
                timeoutId = null;
            }

            if (safetyTimeoutId) {
                clearTimeout(safetyTimeoutId);
                safetyTimeoutId = null;
            }

            // Dispose terminal
            if (hiddenTerminal) {
                try {
                    hiddenTerminal.dispose();
                } catch {
                    // Ignore disposal errors
                }
                hiddenTerminal = null;
            }

            // Remove DOM element
            if (hiddenDiv && hiddenDiv.parentNode) {
                try {
                    document.body.removeChild(hiddenDiv);
                } catch {
                    // Ignore removal errors
                }
                hiddenDiv = null;
            }
        };

        const safeResolve = (value: string) => {
            if (!isResolved) {
                isResolved = true;
                cleanup();
                resolve(value);
            }
        };

        const safeReject = (error: any) => {
            if (!isResolved) {
                isResolved = true;
                cleanup();
                reject(error);
            }
        };

        try {
            // Create a hidden terminal instance
            hiddenTerminal = new XTerminal({
                convertEol: true,
                disableStdin: true,
                rows: 50,
                cols: 120,
            });

            // Create a hidden div to mount the terminal
            hiddenDiv = document.createElement('div');
            hiddenDiv.style.position = 'absolute';
            hiddenDiv.style.left = '-9999px';
            hiddenDiv.style.top = '-9999px';
            hiddenDiv.style.visibility = 'hidden';
            document.body.appendChild(hiddenDiv);

            // Open terminal and write content
            hiddenTerminal.open(hiddenDiv);
            
            // Write the terminal content
            hiddenTerminal.write(terminalContent);

            // Small delay to ensure content is processed
            timeoutId = setTimeout(() => {
                try {
                    if (isResolved) {
                        return; // Already resolved/rejected
                    }

                    if (!hiddenTerminal) {
                        safeResolve(terminalContent);
                        return;
                    }

                    // Extract clean text from terminal buffer
                    let cleanText = '';
                    const buffer = hiddenTerminal.buffer.active;
                    
                    for (let i = 0; i < buffer.length; i++) {
                        const line = buffer.getLine(i);
                        if (line) {
                            const lineText = line.translateToString(true).trimEnd();
                            if (lineText || cleanText) { // Include empty lines only if we have content
                                cleanText += lineText + '\n';
                            }
                        }
                    }

                    safeResolve('```bash\n' + cleanText.trimEnd() + '\n```');
                } catch (error) {
                    // Fallback to original content
                    safeResolve(terminalContent);
                }
            }, 100);

            // Add timeout safety net (10 seconds max)
            safetyTimeoutId = setTimeout(() => {
                if (!isResolved) {
                    console.warn('Terminal text extraction timed out, falling back to original content');
                    safeResolve(terminalContent);
                }
            }, 1000);

        } catch (error) {
            // Fallback to original content on any initialization error
            safeResolve(terminalContent);
        }
    });
};

/**
 * Formats message content for copying to clipboard as markdown with collapsible sections
 */
export const formatMessageForClipboard = async (messageData: CopyableMessage): Promise<string> => {
    const { thinking, message, result, resultFormat = ResultFormat.Plain } = messageData;
    let content = '';
    
    // Add thinking if present
    if (thinking && thinking.trim()) {
        content += `<details>\n<summary>Thinking</summary>\n\n${thinking.trim()}\n\n</details>\n\n`;
    }
    
    // Add main message
    if (message && message.trim()) {
        content += message.trim();
    }
    
    // Add result if present
    if (result && result.trim()) {
        if (content) {
            content += '\n\n';
        }
        
        let resultContent = result.trim();
        
        // Handle terminal format specially to get clean text
        if (resultFormat === ResultFormat.Terminal) {
            try {
                resultContent = await getCleanTerminalText(result);
            } catch {
                // Fallback to original result
                resultContent = result.trim();
            }
        }
        
        content += `<details>\n<summary>Result</summary>\n\n${resultContent}\n\n</details>`;
    }
    
    return content;
};

/**
 * Copies formatted message content to clipboard
 */
export const copyMessageToClipboard = async (messageData: CopyableMessage): Promise<void> => {
    try {
        const content = await formatMessageForClipboard(messageData);
        await navigator.clipboard.writeText(content);
    } catch {
        // Ignore clipboard errors - they're handled silently
    }
};
