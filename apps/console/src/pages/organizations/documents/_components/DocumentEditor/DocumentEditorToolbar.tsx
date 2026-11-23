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

import type { RepoCollaborationHandle } from "@probo/collaboration";
import { Separator } from "@probo/ui/src/v2/Separator/Separator";
import type { PreloadedQuery } from "react-relay";

import type { DocumentVersionContentQuery } from "#/__generated__/core/DocumentVersionContentQuery.graphql";

import type { DocumentMode } from "../../_lib/useDocumentModes";
import { documentEditor } from "../../variants";
import { DocumentModeSelect } from "../DocumentModeSelect";
import { DocumentVersionLabel } from "../DocumentVersionList/DocumentVersionLabel";
import { DocumentVersionsTrigger } from "../DocumentVersionList/DocumentVersionsTrigger";

import { DocumentPresenceAvatars } from "./DocumentPresenceAvatars";

interface DocumentEditorToolbarProps {
  mode: DocumentMode;
  modes: readonly DocumentMode[];
  onModeChange: (mode: DocumentMode) => void;
  collaborationHandle: RepoCollaborationHandle | null | undefined;
  // The editor portals its formatting controls here while editing or
  // suggesting. Viewing has no slot, so the editor mounts no controls.
  onControlsElement: (element: HTMLDivElement | null) => void;
  versionsOpen: boolean;
  onVersionsOpenChange: (open: boolean) => void;
  versionQueryRef: PreloadedQuery<DocumentVersionContentQuery> | null;
  onShowLatest: () => void;
}

export function DocumentEditorToolbar({
  mode,
  modes,
  onModeChange,
  collaborationHandle,
  onControlsElement,
  versionsOpen,
  onVersionsOpenChange,
  versionQueryRef,
  onShowLatest,
}: DocumentEditorToolbarProps) {
  const viewing = mode === "viewing";
  const {
    toolbar, toolbarControls, toolbarLabel, toolbarEnd, toolbarMode, toolbarSeparator,
  } = documentEditor();

  return (
    <div className={toolbar()}>
      {viewing
        ? (
            <div className={toolbarLabel()}>
              <DocumentVersionLabel queryRef={versionQueryRef} onShowLatest={onShowLatest} />
            </div>
          )
        : <div ref={onControlsElement} className={toolbarControls()} />}
      <div className={toolbarEnd()}>
        {!viewing && collaborationHandle != null && (
          <DocumentPresenceAvatars collaborationHandle={collaborationHandle} />
        )}
        <div className={toolbarMode()}>
          <DocumentModeSelect mode={mode} modes={modes} onModeChange={onModeChange} />
        </div>
        <Separator orientation="vertical" className={toolbarSeparator()} />
        <DocumentVersionsTrigger open={versionsOpen} onOpenChange={onVersionsOpenChange} />
      </div>
    </div>
  );
}
