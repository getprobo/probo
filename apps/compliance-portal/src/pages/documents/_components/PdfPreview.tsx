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

import "react-pdf/dist/Page/AnnotationLayer.css";
import "react-pdf/dist/Page/TextLayer.css";

import { SpinnerGapIcon } from "@phosphor-icons/react";
import { times } from "@probo/helpers";
// Vite `?url` import resolves to the bundled worker URL (string); the import-x
// resolver doesn't understand the suffix, and `vite/client` types cover it.
// eslint-disable-next-line import-x/default
import workerSrc from "pdfjs-dist/build/pdf.worker.min.mjs?url";
import type { ComponentRef, Ref } from "react";
import { useImperativeHandle, useRef, useState } from "react";
import { Document, Page, pdfjs } from "react-pdf";

import { pdfPreview } from "./variants";

// Bundle the pdf.js worker with the app (via Vite's `?url`) instead of loading
// it from a CDN, so the viewer works under a strict trust-center CSP.
pdfjs.GlobalWorkerOptions.workerSrc = workerSrc;

export interface PdfPreviewHandle {
  scrollToPage: (page: number) => void;
}

interface PdfPreviewProps {
  // Base64 data URI of the PDF to render.
  file: string;
  // Zoom factor applied to every page.
  scale: number;
  // Handle exposing imperative page navigation to the toolbar.
  ref?: Ref<PdfPreviewHandle>;
  // Reports the page count once the document has loaded.
  onNumPages: (numPages: number) => void;
  // Reports the page currently centered in the viewport.
  onVisiblePageChange: (page: number) => void;
}

// Scrollable react-pdf renderer, controlled by the viewer toolbar: it takes the
// zoom `scale`, reports the page count and the visible page, and exposes an
// imperative `scrollToPage` for the page-navigation buttons.
export function PdfPreview({ file, scale, ref, onNumPages, onVisiblePageChange }: PdfPreviewProps) {
  const [numPages, setNumPages] = useState(0);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const documentRef = useRef<ComponentRef<typeof Document>>(null);

  useImperativeHandle(ref, () => ({
    scrollToPage(page) {
      const node = documentRef.current?.pages.current[page - 1];
      node?.scrollIntoView({ behavior: "smooth", block: "start" });
    },
  }), []);

  const resolveVisiblePage = () => {
    const wrapper = wrapperRef.current;
    const pages = documentRef.current?.pages.current;
    if (!wrapper || !pages?.length) {
      return;
    }
    const middle = wrapper.getBoundingClientRect().top + wrapper.clientHeight / 2;
    for (let index = 0; index < pages.length; index += 1) {
      const rect = pages[index].getBoundingClientRect();
      if (rect.top <= middle && rect.bottom >= middle) {
        onVisiblePageChange(index + 1);
        return;
      }
    }
  };

  const slots = pdfPreview();

  return (
    <div ref={wrapperRef} onScrollEnd={resolveVisiblePage} className={slots.viewport()}>
      <Document
        ref={documentRef}
        file={file}
        className={slots.list()}
        loading={(
          <div className={slots.loading()}>
            <SpinnerGapIcon className={slots.spinner()} />
          </div>
        )}
        onLoadSuccess={(document) => {
          setNumPages(document.numPages);
          onNumPages(document.numPages);
          onVisiblePageChange(1);
        }}
      >
        {times(numPages, index => (
          <Page key={index} pageNumber={index + 1} scale={scale} className={slots.page()} />
        ))}
      </Document>
    </div>
  );
}
