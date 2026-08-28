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
import type { ReactNode } from "react";
import { useTranslation } from "react-i18next";

import { downloadDataUri } from "#/pages/_lib/dataUri";
import { useCopyDocumentLink } from "#/pages/_lib/useCopyDocumentLink";

import { documentViewerToolbar } from "./variants";

const MIN_SCALE = 0.5;
const MAX_SCALE = 3;

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

interface ViewerToolbarActionProps {
  label: string;
  icon: ReactNode;
  disabled?: boolean;
  labeledClassName: string;
  iconClassName: string;
  onClick: () => void;
}

// Labeled ghost button when the toolbar is wide; IconButton when the stage
// container is narrower than xl so the row does not wrap under h-12.
function ViewerToolbarAction({
  label,
  icon,
  disabled,
  labeledClassName,
  iconClassName,
  onClick,
}: ViewerToolbarActionProps) {
  return (
    <>
      <span className={labeledClassName}>
        <Button
          variant="ghost"
          color="neutral"
          iconStart={icon}
          disabled={disabled}
          onClick={onClick}
        >
          {label}
        </Button>
      </span>
      <span className={iconClassName}>
        <IconButton
          variant="ghost"
          color="neutral"
          aria-label={label}
          disabled={disabled}
          onClick={onClick}
        >
          {icon}
        </IconButton>
      </span>
    </>
  );
}

interface DocumentViewerToolbarProps {
  currentPage: number;
  numPages: number;
  scale: number;
  // The exported base64 data URI, or null while it is still loading.
  dataUri: string | null;
  // File name used when downloading.
  downloadName: string;
  onScaleChange: (scale: number) => void;
  // Moves the preview by one page in the given direction.
  onMovePage: (direction: 1 | -1) => void;
}

// Page navigation, zoom, share, and download chrome above the PDF stage.
export function DocumentViewerToolbar({
  currentPage,
  numPages,
  scale,
  dataUri,
  downloadName,
  onScaleChange,
  onMovePage,
}: DocumentViewerToolbarProps) {
  const { t } = useTranslation();
  const copyDocumentLink = useCopyDocumentLink();
  const slots = documentViewerToolbar();

  const handleCopyLink = () => {
    copyDocumentLink(window.location.href);
  };

  const handleDownload = () => {
    if (dataUri) {
      downloadDataUri(dataUri, downloadName);
    }
  };

  return (
    <div className={slots.root()}>
      <div className={slots.start()}>
        <div className={slots.controls()}>
          <IconButton
            variant="ghost"
            color="neutral"
            aria-label={t("viewer.previousPage")}
            disabled={currentPage <= 1}
            onClick={() => onMovePage(-1)}
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
            disabled={currentPage >= numPages || numPages === 0}
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
            aria-label={t("viewer.zoomOut")}
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
            aria-label={t("viewer.zoomIn")}
            onClick={() => onScaleChange(clamp(scale * 1.25, MIN_SCALE, MAX_SCALE))}
          >
            <MagnifyingGlassPlusIcon />
          </IconButton>
        </div>
      </div>
      <div className={slots.actions()}>
        <ViewerToolbarAction
          label={t("viewer.share")}
          icon={<LinkSimpleIcon />}
          labeledClassName={slots.actionLabeled()}
          iconClassName={slots.actionIcon()}
          onClick={handleCopyLink}
        />
        <Separator orientation="vertical" className={slots.actionSeparator()} />
        <ViewerToolbarAction
          label={t("viewer.download")}
          icon={<DownloadSimpleIcon />}
          disabled={dataUri == null}
          labeledClassName={slots.actionLabeled()}
          iconClassName={slots.actionIcon()}
          onClick={handleDownload}
        />
      </div>
    </div>
  );
}
