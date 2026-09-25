// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { RichEditor } from "@probo/ui";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useCallback, useEffect, useRef, useState } from "react";
import { tv } from "tailwind-variants/lite";

import { isRichEditorContentEmpty } from "#/pages/organizations/_lib/richEditorContent";
import { useDebouncedSerializedFieldSave } from "#/pages/organizations/_lib/useSerializedFieldSave";

const richDescriptionEditor = tv({
  slots: {
    root: "min-w-0",
    editor: "min-h-40",
  },
});

const defaultSaveDelayMs = 1000;

function normalizeContent(value: string) {
  return isRichEditorContentEmpty(value) ? "" : value;
}

export interface RichDescriptionEditorProps {
  saved: string;
  canUpdate: boolean;
  ariaLabel: string;
  emptyLabel: string;
  saveDelayMs?: number;
  save: (content: string | null) => Promise<void>;
}

export function RichDescriptionEditor({
  saved,
  canUpdate,
  ariaLabel,
  emptyLabel,
  saveDelayMs = defaultSaveDelayMs,
  save,
}: RichDescriptionEditorProps) {
  const saveRef = useRef(save);
  const savedRef = useRef(saved);
  const [draft, setDraft] = useState(saved);
  const [savedContent, setSavedContent] = useState(saved);
  const [dirty, setDirty] = useState(false);
  const [editorGeneration, setEditorGeneration] = useState(0);

  useEffect(() => {
    saveRef.current = save;
  }, [save]);

  useEffect(() => {
    savedRef.current = saved;
  }, [saved]);

  if (saved !== savedContent) {
    setSavedContent(saved);
    if (!dirty) {
      setDraft(saved);
      setEditorGeneration(generation => generation + 1);
    }
  }

  const persist = useCallback(
    async (value: string) => {
      const next = normalizeContent(value);

      try {
        await saveRef.current(next || null);
        setDraft((current) => {
          if (current === value || normalizeContent(current) === next) {
            setDirty(false);
            return next;
          }
          return current;
        });
      } catch {
        setDraft((current) => {
          if (normalizeContent(current) === next) {
            setDirty(false);
            setEditorGeneration(generation => generation + 1);
            return savedRef.current;
          }
          return current;
        });
      }
    },
    [],
  );
  const persistDebounced = useDebouncedSerializedFieldSave(persist, saveDelayMs);
  const { root, editor } = richDescriptionEditor();

  return (
    <div className={root()}>
      {canUpdate
        ? (
            <RichEditor
              key={editorGeneration}
              className={editor()}
              content={draft}
              aria-label={ariaLabel}
              onChangeContent={(next) => {
                setDirty(true);
                setDraft(next);
                persistDebounced.schedule(next);
              }}
              onBlur={() => {
                persistDebounced.flush();
              }}
            />
          )
        : isRichEditorContentEmpty(saved)
          ? (
              <Text size={2} color="faint">
                {emptyLabel}
              </Text>
            )
          : (
              <RichEditor
                key={editorGeneration}
                className={editor()}
                content={saved}
                disabled
                aria-label={ariaLabel}
              />
            )}
    </div>
  );
}
