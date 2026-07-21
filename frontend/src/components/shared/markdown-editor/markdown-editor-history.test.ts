import { Editor, Extension } from '@tiptap/core';
import { Plugin, PluginKey } from '@tiptap/pm/state';
import { beforeAll, expect, it } from 'vitest';

import { resetUndoHistory } from './markdown-editor';
import { createMarkdownExtensions } from './markdown-editor-extensions';
import { setupEditorJsdom } from './markdown-editor-test-setup';

beforeAll(setupEditorJsdom);

// Reproduces the production-only crash class: `view.updateState` inside resetUndoHistory can synchronously
// advance `view.state` (a plugin view dispatching during reconfigure), so a poke transaction built from the
// pre-updateState `newState` has a `before` doc that no longer matches — ProseMirror throws "Applying a
// mismatched transaction". This synthetic plugin forces that advance on demand; jsdom's real editor never
// hits the browser timing that triggers it, so the fault has to be simulated to be regression-tested.
let isArmed = false;

const AdvanceStateOnUpdate = Extension.create({
    addProseMirrorPlugins() {
        return [
            new Plugin({
                key: new PluginKey('advanceStateOnUpdate'),
                // Dispatch from the plugin-view FACTORY, not `update`: swapping the history plugin changes the
                // plugin set, so `view.updateState` recreates plugin views (factory re-runs) rather than
                // calling `update`. Dispatching here fires mid-reconfigure, advancing view.state exactly as the
                // real prod crash does.
                view: (view) => {
                    if (isArmed) {
                        isArmed = false;

                        let pos = 1;

                        view.state.doc.descendants((node, nodePos) => {
                            if (node.isText) {
                                pos = nodePos;

                                return false;
                            }

                            return true;
                        });

                        view.dispatch(view.state.tr.insertText('Z', pos));
                    }

                    return {};
                },
            }),
        ];
    },
    name: 'advanceStateOnUpdate',
});

it('resetUndoHistory survives view.state advancing during updateState', () => {
    const editor = new Editor({
        content: '- first item\n- second item',
        contentType: 'markdown',
        extensions: [...createMarkdownExtensions(), AdvanceStateOnUpdate],
    });

    isArmed = true;

    expect(() => resetUndoHistory(editor)).not.toThrow();

    editor.destroy();
});

it('resetUndoHistory actually clears the undo stack', () => {
    const editor = new Editor({ content: 'seed', contentType: 'markdown', extensions: createMarkdownExtensions() });

    editor.commands.insertContent(' more');

    expect(editor.can().undo()).toBe(true);

    resetUndoHistory(editor);

    expect(editor.can().undo()).toBe(false);

    editor.destroy();
});
