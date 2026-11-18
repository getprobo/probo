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

import { ScrollArea } from "@base-ui/react/scroll-area";
import { CollaborativeRichEditor } from "@probo/ui/src/RichEditor/CollaborativeRichEditor";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useRef } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useFragment } from "react-relay";

import type { DocumentEditor_document$key } from "#/__generated__/core/DocumentEditor_document.graphql";
import { useUploadRichTextFile } from "#/pages/organizations/_lib/useUploadRichTextFile";

import { useDocumentCollaboration } from "../_lib/useDocumentCollaboration";
import type { DocumentMode } from "../_lib/useDocumentModes";
import { usePreventSuggestionTextDrop } from "../_lib/usePreventSuggestionTextDrop";
import { useSuggestionController } from "../_lib/useSuggestionController";
import { documentEditor } from "../variants";

const documentEditorFragment = graphql`
  fragment DocumentEditor_document on Document {
    id
  }
`;

interface DocumentEditorProps {
  documentKey: DocumentEditor_document$key;
  mode: DocumentMode;
}

export function DocumentEditor({ documentKey, mode }: DocumentEditorProps) {
  const document = useFragment(documentEditorFragment, documentKey);
  const { t } = useTranslation("organizations/documents");
  const { handle, failed } = useDocumentCollaboration(document.id);
  const { controller, setEditor } = useSuggestionController(document.id, mode);
  const onUploadFile = useUploadRichTextFile();
  const paperRef = useRef<HTMLDivElement>(null);
  const { frame, viewport, paper, scrollbar, thumb } = documentEditor();
  usePreventSuggestionTextDrop(paperRef, mode);

  return (
    <ScrollArea.Root className={frame()}>
      <ScrollArea.Viewport className={viewport()}>
        <ScrollArea.Content>
          <div ref={paperRef} className={paper()} data-theme="document">
            {failed
              ? <Text size={2} color="faint">{t("details.editor.openError")}</Text>
              : handle == null
                ? <Text size={2} color="faint">{t("details.editor.loading")}</Text>
                : (
                    <CollaborativeRichEditor
                      key={document.id}
                      className="w-full"
                      collaborationHandle={handle}
                      content=""
                      page={false}
                      disabled={mode === "viewing"}
                      formatting={mode === "editing"}
                      suggestingMenus={mode === "suggesting"}
                      suggesting={controller}
                      onEditor={setEditor}
                      onUploadFile={onUploadFile}
                    />
                  )}
          </div>
        </ScrollArea.Content>
      </ScrollArea.Viewport>
      <ScrollArea.Scrollbar className={scrollbar()} orientation="vertical">
        <ScrollArea.Thumb className={thumb()} />
      </ScrollArea.Scrollbar>
    </ScrollArea.Root>
  );
}
