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

import { CaretLeftIcon, SpinnerGapIcon } from "@phosphor-icons/react";
import { ButtonLink } from "@probo/ui/src/v2/Button/ButtonLink";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";
import { useLocalizedPath } from "#/lib/i18n/useLocale";

import { dataUriMimeType, downloadDataUri } from "../_lib/dataUri";

import { DocumentDownloadFallback } from "./DocumentDownloadFallback";
import { DocumentLocked } from "./DocumentLocked";
import { DocumentViewerToolbar } from "./DocumentViewerToolbar";
import type { PdfPreviewHandle } from "./PdfPreview";
import { PdfPreview } from "./PdfPreview";
import { documentViewer } from "./variants";

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

interface DocumentViewerLocked {
  // Requests access for the locked resource (prompting sign-in first when
  // needed).
  onGetAccess: () => void;
  // Whether the access request is in flight.
  isRequesting: boolean;
}

interface DocumentViewerProps {
  // The document/file/report display name.
  title: string;
  // The exported base64 data URI, or null while it is still loading. Omitted
  // when the viewer is locked.
  dataUri?: string | null;
  // File name used when downloading. Omitted when the viewer is locked.
  downloadName?: string;
  // When set, the header keeps the title (no toolbar) and the body shows the
  // locked empty state with a Get Access CTA instead of the file preview.
  locked?: DocumentViewerLocked;
}

// Full-page document viewer: a header band with the title and a toolbar
// (page navigation + zoom for PDFs, copy link, download) above the scrollable
// body. PDFs render with react-pdf, images inline, and anything else offers a
// download. Locked visitors keep the title header and see a Get Access CTA in
// place of the preview.
export function DocumentViewer({ title, dataUri = null, downloadName = title, locked }: DocumentViewerProps) {
  const { t } = useTranslation("documents");
  const localizedPath = useLocalizedPath();

  const pdfRef = useRef<PdfPreviewHandle>(null);
  const [numPages, setNumPages] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [scale, setScale] = useState(1);

  const isLocked = locked != null;
  const mimeType = !isLocked && dataUri ? dataUriMimeType(dataUri) : null;
  const isPdf = mimeType === "application/pdf";
  const isImage = mimeType?.startsWith("image/") ?? false;

  const movePage = (direction: 1 | -1) => {
    const next = clamp(currentPage + direction, 1, numPages);
    pdfRef.current?.scrollToPage(next);
    setCurrentPage(next);
  };

  const handleDownload = () => {
    if (dataUri) {
      downloadDataUri(dataUri, downloadName);
    }
  };

  const slots = documentViewer();

  return (
    <div className={slots.root()}>
      <HeaderBand flushBottomSpace={!isLocked}>
        <div className={slots.header()}>
          <ButtonLink to={localizedPath("/documents")} variant="ghost" color="neutral" size={1} iconStart={<CaretLeftIcon />} className={slots.back()}>
            {t("viewer.back")}
          </ButtonLink>
          <Heading level={1} size={7} weight="medium" highContrast className="truncate">
            {title}
          </Heading>
          {!isLocked && (
            <DocumentViewerToolbar
              isPdf={isPdf}
              currentPage={currentPage}
              numPages={numPages}
              scale={scale}
              onScaleChange={setScale}
              onMovePage={movePage}
              dataUri={dataUri}
              downloadName={downloadName}
            />
          )}
        </div>
      </HeaderBand>

      <div className={slots.body()}>
        {isLocked
          ? (
              <DocumentLocked
                onGetAccess={locked.onGetAccess}
                isRequesting={locked.isRequesting}
              />
            )
          : dataUri == null
            ? (
                <div className={slots.stage()}>
                  <SpinnerGapIcon className={slots.spinner()} />
                </div>
              )
            : isPdf
              ? (
                  <PdfPreview
                    ref={pdfRef}
                    file={dataUri}
                    scale={scale}
                    onNumPages={setNumPages}
                    onVisiblePageChange={setCurrentPage}
                  />
                )
              : isImage
                ? (
                    <div className={slots.imageStage()}>
                      <img src={dataUri} alt={title} className={slots.image()} />
                    </div>
                  )
                : <DocumentDownloadFallback onDownload={handleDownload} />}
      </div>
    </div>
  );
}
