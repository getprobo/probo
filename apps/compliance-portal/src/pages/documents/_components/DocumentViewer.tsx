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

import { Toast } from "@base-ui/react/toast";
import {
  CaretLeftIcon,
  CaretRightIcon,
  DownloadSimpleIcon,
  MagnifyingGlassMinusIcon,
  MagnifyingGlassPlusIcon,
  ShareNetworkIcon,
  SpinnerGapIcon,
} from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Link } from "@probo/ui/src/v2/Button/Link";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Separator } from "@probo/ui/src/v2/Separator/Separator";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useRef, useState } from "react";
import { useTranslation } from "react-i18next";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";

import { dataUriMimeType, downloadDataUri } from "../_lib/dataUri";

import { DocumentDownloadFallback } from "./DocumentDownloadFallback";
import type { PdfPreviewHandle } from "./PdfPreview";
import { PdfPreview } from "./PdfPreview";
import { documentViewer } from "./variants";

const MIN_SCALE = 0.5;
const MAX_SCALE = 3;

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

interface DocumentViewerProps {
  // The document/file/report display name.
  title: string;
  // The exported base64 data URI, or null while it is still loading.
  dataUri: string | null;
  // File name used when downloading.
  downloadName: string;
}

// Full-page document viewer: a header band with the title and a toolbar
// (page navigation + zoom for PDFs, share, download) above the scrollable body.
// PDFs render with react-pdf, images inline, and anything else offers a
// download.
export function DocumentViewer({ title, dataUri, downloadName }: DocumentViewerProps) {
  const { t } = useTranslation("documents");
  const toast = Toast.useToastManager();

  const pdfRef = useRef<PdfPreviewHandle>(null);
  const [numPages, setNumPages] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [scale, setScale] = useState(1);

  const mimeType = dataUri ? dataUriMimeType(dataUri) : null;
  const isPdf = mimeType === "application/pdf";
  const isImage = mimeType?.startsWith("image/") ?? false;

  const movePage = (direction: 1 | -1) => {
    const next = clamp(currentPage + direction, 1, numPages);
    pdfRef.current?.scrollToPage(next);
    setCurrentPage(next);
  };

  const handleShare = () => {
    navigator.clipboard.writeText(window.location.href).then(
      () => toast.add({ title: t("viewer.linkCopied"), type: "success" }),
      () => {},
    );
  };

  const handleDownload = () => {
    if (dataUri) {
      downloadDataUri(dataUri, downloadName);
    }
  };

  const slots = documentViewer();

  return (
    <div className={slots.root()}>
      <HeaderBand flushBottomSpace>
        <div className={slots.header()}>
          <Link to="/documents" variant="ghost" color="neutral" size={1} iconStart={<CaretLeftIcon />} className={slots.back()}>
            {t("viewer.back")}
          </Link>
          <Heading level={1} size={7} weight="medium" highContrast className="truncate">
            {title}
          </Heading>
          <div className={slots.toolbar()}>
            <div className={slots.toolbarStart()}>
              {isPdf && (
                <>
                  <div className={slots.controls()}>
                    <IconButton
                      variant="ghost"
                      color="neutral"
                      aria-label={t("viewer.previousPage")}
                      disabled={currentPage <= 1}
                      onClick={() => movePage(-1)}
                    >
                      <CaretLeftIcon />
                    </IconButton>
                    <Text size={2} color="neutral">
                      {t("viewer.pageOf", { current: currentPage, total: numPages })}
                    </Text>
                    <IconButton
                      variant="ghost"
                      color="neutral"
                      aria-label={t("viewer.nextPage")}
                      disabled={currentPage >= numPages}
                      onClick={() => movePage(1)}
                    >
                      <CaretRightIcon />
                    </IconButton>
                  </div>
                  <Separator orientation="vertical" className={slots.separator()} />
                  <div className={slots.controls()}>
                    <IconButton
                      variant="ghost"
                      color="neutral"
                      aria-label={t("viewer.zoomOut")}
                      onClick={() => setScale(value => clamp(value * 0.8, MIN_SCALE, MAX_SCALE))}
                    >
                      <MagnifyingGlassMinusIcon />
                    </IconButton>
                    <Text size={2} color="neutral">
                      {`${Math.round(scale * 100)}%`}
                    </Text>
                    <IconButton
                      variant="ghost"
                      color="neutral"
                      aria-label={t("viewer.zoomIn")}
                      onClick={() => setScale(value => clamp(value * 1.25, MIN_SCALE, MAX_SCALE))}
                    >
                      <MagnifyingGlassPlusIcon />
                    </IconButton>
                  </div>
                </>
              )}
            </div>
            <div className={slots.actions()}>
              <Button variant="ghost" color="neutral" iconStart={<ShareNetworkIcon />} onClick={handleShare}>
                {t("viewer.share")}
              </Button>
              <Separator orientation="vertical" className={slots.separator()} />
              <Button
                variant="ghost"
                color="neutral"
                iconStart={<DownloadSimpleIcon />}
                disabled={dataUri == null}
                onClick={handleDownload}
              >
                {t("viewer.download")}
              </Button>
            </div>
          </div>
        </div>
      </HeaderBand>

      <div className={slots.body()}>
        {dataUri == null
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
