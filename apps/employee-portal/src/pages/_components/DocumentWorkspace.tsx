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

import { SpinnerGapIcon } from "@phosphor-icons/react";
import { Separator } from "@probo/ui/src/v2/Separator/Separator";
import type { ReactNode } from "react";
import { useRef, useState } from "react";

import { DocumentViewerToolbar } from "./DocumentViewerToolbar";
import type { PdfPreviewHandle } from "./PdfPreview";
import { PdfPreview } from "./PdfPreview";
import { documentWorkspace } from "./variants";

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

interface DocumentWorkspaceProps {
  // Request-panel slot (signature / approval chrome).
  request: ReactNode;
  // Version history under the request panel.
  history?: ReactNode;
  // Document title used as the PDF download name.
  title: string;
  // Exported PDF data URI, or null while it is still loading.
  dataUri: string | null;
}

// Split document view: request column on the left, PDF stage on the right.
// Named for view transitions so the queue top bar stays put while the pane
// swipes between documents.
export function DocumentWorkspace({
  request,
  history,
  title,
  dataUri,
}: DocumentWorkspaceProps) {
  const pdfRef = useRef<PdfPreviewHandle>(null);
  const [numPages, setNumPages] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [loadedUri, setLoadedUri] = useState(dataUri);
  const [scale, setScale] = useState(1);
  const slots = documentWorkspace();

  if (loadedUri !== dataUri) {
    setLoadedUri(dataUri);
    setCurrentPage(1);
    setNumPages(0);
  }

  const movePage = (direction: 1 | -1) => {
    const next = clamp(currentPage + direction, 1, numPages);
    pdfRef.current?.scrollToPage(next);
    setCurrentPage(next);
  };

  return (
    <div className={slots.root()}>
      <aside className={slots.request()}>
        <div className={slots.requestBody()}>
          {request}
        </div>
        {history == null
          ? null
          : (
              <div className={slots.history()}>
                <Separator />
                {history}
              </div>
            )}
      </aside>
      <section className={slots.stage()}>
        <DocumentViewerToolbar
          currentPage={currentPage}
          numPages={numPages}
          scale={scale}
          dataUri={dataUri}
          downloadName={`${title}.pdf`}
          onScaleChange={setScale}
          onMovePage={movePage}
        />
        {dataUri == null
          ? (
              <div className={slots.loading()}>
                <SpinnerGapIcon className={slots.spinner()} />
              </div>
            )
          : (
              <PdfPreview
                ref={pdfRef}
                file={dataUri}
                scale={scale}
                onNumPages={setNumPages}
                onVisiblePageChange={setCurrentPage}
              />
            )}
      </section>
    </div>
  );
}
