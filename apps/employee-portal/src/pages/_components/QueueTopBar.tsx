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

import { CaretLeftIcon, CaretRightIcon, XIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { useDisplayMode } from "@probo/ui/src/v2/displayMode/useDisplayMode";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { useParams } from "react-router";

import { useDocumentQueue } from "#/pages/_lib/DocumentQueueContext";

import { queueTopBar } from "./variants";

// Inverse-theme queue chrome: prev/next through the frozen snapshot, plus Close.
export function QueueTopBar() {
  const { t } = useTranslation();
  const { documentId } = useParams();
  const { displayMode } = useDisplayMode();
  const { snapshot, advancing, goTo, goForward, close } = useDocumentQueue();
  const slots = queueTopBar();
  const island = displayMode === "dark" ? "light" : "dark";

  if (snapshot == null || documentId == null) {
    return null;
  }

  const index = snapshot.ids.indexOf(documentId);
  const current = index >= 0 ? index + 1 : 1;
  const total = snapshot.totalCount;
  const previousId = index > 0 ? snapshot.ids[index - 1] : null;
  const nextId = index >= 0 && index < snapshot.ids.length - 1
    ? snapshot.ids[index + 1]
    : null;
  const canGoForward = nextId != null || snapshot.hasNextPage;
  const progressKey = snapshot.kind === "signatures"
    ? "queue.progressSignatures"
    : "queue.progressApprovals";

  return (
    <header
      className={slots.bar({
        className: island === "dark" ? "dark scheme-dark" : "light scheme-light",
      })}
    >
      <div className={slots.start()}>
        <div className={slots.controls()}>
          <IconButton
            variant="outline"
            color="neutral"
            aria-label={t("queue.previous")}
            disabled={previousId == null || advancing}
            onClick={() => {
              if (previousId != null) {
                goTo(previousId, "back");
              }
            }}
          >
            <CaretLeftIcon />
          </IconButton>
          <IconButton
            variant="outline"
            color="neutral"
            aria-label={t("queue.next")}
            disabled={!canGoForward || advancing}
            loading={advancing}
            onClick={() => {
              goForward();
            }}
          >
            <CaretRightIcon />
          </IconButton>
        </div>
        <Text size={2} weight="medium" highContrast className={slots.progress()}>
          {t(progressKey, { current, total })}
        </Text>
      </div>
      <Button
        variant="outline"
        color="neutral"
        size={2}
        iconStart={<XIcon />}
        onClick={() => {
          close();
        }}
      >
        {t("queue.close")}
      </Button>
    </header>
  );
}
