import type { Editor } from '@tiptap/react';

/**
 * `onCloseAutoFocus` handler shared by every editor popover and dropdown: keep focus on the caret instead of
 * the trigger so the user can keep typing.
 *
 * The destroyed check is load-bearing, not defensive. Radix closes a still-open layer asynchronously, so
 * unmounting the editor with a menu open lands this callback on a torn-down instance and `editor.commands`
 * throws — reproduced by the toolbar tests, which took `vitest run` to exit 1 while every test read green.
 */
export const returnFocusToEditor = (editor: Editor) => (event: Event) => {
    event.preventDefault();

    if (!editor.isDestroyed) {
        editor.commands.focus();
    }
};
