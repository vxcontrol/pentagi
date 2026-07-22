import type { Page } from '@playwright/test';

// The markdown editor is a ProseMirror contenteditable: fill() mutates the DOM directly and
// races the observer flush against the next action (e.g. a submit click), so the typed content
// can fail to land in the model. Real key events apply synchronously — always type through this.
export const typeIntoEditor = async (page: Page, name: string, text: string): Promise<void> => {
    const editor = page.getByRole('textbox', { name });

    await editor.click();
    await editor.pressSequentially(text);
};
