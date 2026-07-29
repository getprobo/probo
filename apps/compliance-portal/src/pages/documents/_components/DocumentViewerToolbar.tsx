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
  LinkSimpleIcon,
  MagnifyingGlassMinusIcon,
  MagnifyingGlassPlusIcon,
} from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Separator } from "@probo/ui/src/v2/Separator/Separator";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { downloadDataUri } from "../_lib/dataUri";

import { documentViewerToolbar } from "./variants";

const MIN_SCALE = 0.5;
const MAX_SCALE = 3;

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

interface DocumentViewerToolbarProps {
  // Whether PDF page/zoom controls are shown (only for PDF previews).
  isPdf: boolean;
  currentPage: number;
  numPages: number;
  scale: number;
  onScaleChange: (scale: number) => void;
  // Moves the preview by one page in the given direction.
  onMovePage: (direction: 1 | -1) => void;
  // The exported base64 data URI, or null while it is still loading.
  dataUri: string | null;
  // File name used when downloading.
  downloadName: string;
}

// Viewer chrome under the title: page navigation + zoom for PDFs, plus copy
// link and download actions shared by every previewable file type.
export function DocumentViewerToolbar({
  isPdf,
  currentPage,
  numPages,
  scale,
  onScaleChange,
  onMovePage,
  dataUri,
  downloadName,
}: DocumentViewerToolbarProps) {
  const { t } = useTranslation("documents");
  const toast = Toast.useToastManager();
  const slots = documentViewerToolbar();

  const handleCopyLink = () => {
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

  return (
    <div className={slots.root()}>
      <div className={slots.start()}>
        {isPdf && (
          <>
            <div className={slots.controls()}>
              <IconButton
                variant="ghost"
                color="neutral"
                aria-label={t("common.previousPage")}
                disabled={currentPage <= 1}
                onClick={() => onMovePage(-1)}
              >
                <CaretLeftIcon />
              </IconButton>
              <Text size={2} color="neutral">
                {t("common.pageOf", { current: currentPage, total: numPages })}
              </Text>
              <IconButton
                variant="ghost"
                color="neutral"
                aria-label={t("common.nextPage")}
                disabled={currentPage >= numPages}
                onClick={() => onMovePage(1)}
              >
                <CaretRightIcon />
              </IconButton>
            </div>
            <Separator orientation="vertical" className={slots.separator()} />
            <div className={slots.controls()}>
              <IconButton
                variant="ghost"
                color="neutral"
                aria-label={t("common.zoomOut")}
                onClick={() => onScaleChange(clamp(scale * 0.8, MIN_SCALE, MAX_SCALE))}
              >
                <MagnifyingGlassMinusIcon />
              </IconButton>
              <Text size={2} color="neutral">
                {`${Math.round(scale * 100)}%`}
              </Text>
              <IconButton
                variant="ghost"
                color="neutral"
                aria-label={t("common.zoomIn")}
                onClick={() => onScaleChange(clamp(scale * 1.25, MIN_SCALE, MAX_SCALE))}
              >
                <MagnifyingGlassPlusIcon />
              </IconButton>
            </div>
          </>
        )}
      </div>
      <div className={slots.actions()}>
        <Button
          variant="ghost"
          color="neutral"
          iconStart={<LinkSimpleIcon />}
          onClick={handleCopyLink}
          aria-label={t("viewer.copyLink")}
        >
          <span className={slots.actionLabel()}>{t("viewer.copyLink")}</span>
        </Button>
        <Separator orientation="vertical" className={slots.separator()} />
        <Button
          variant="ghost"
          color="neutral"
          iconStart={<DownloadSimpleIcon />}
          disabled={dataUri == null}
          onClick={handleDownload}
          aria-label={t("viewer.download")}
        >
          <span className={slots.actionLabel()}>{t("viewer.download")}</span>
        </Button>
      </div>
    </div>
  );
}
