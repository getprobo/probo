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

import { CaretLeftIcon, CaretRightIcon, CheckCircleIcon, CheckIcon, ProhibitIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { DocumentRequestPanel } from "#/pages/_components/DocumentRequestPanel";
import { documentRequestPanel } from "#/pages/_components/variants";

type ApprovalState = "PENDING" | "APPROVED" | "REJECTED" | "VOIDED";

interface ApprovalRequestPanelProps {
  title: string;
  state: ApprovalState | null;
  consentText: string;
  queueActive: boolean;
  hasNext: boolean;
  busy: boolean;
  advancing?: boolean;
  onApprove: () => void;
  onReject: () => void;
  onNext: () => void;
  onFinish: () => void;
}

// Left-column approval chrome: approve / reject, queue Finish / Go to next,
// or back to the list when viewing a decided document outside the queue.
export function ApprovalRequestPanel({
  title,
  state,
  consentText,
  queueActive,
  hasNext,
  busy,
  advancing = false,
  onApprove,
  onReject,
  onNext,
  onFinish,
}: ApprovalRequestPanelProps) {
  const { t } = useTranslation("approvals");
  const decided = state === "APPROVED" || state === "REJECTED";
  const tone = state === "REJECTED" ? "rejected" : "approved";
  const slots = documentRequestPanel({ tone });

  const detail = decided
    ? (
        <div className={slots.status()}>
          {state === "REJECTED"
            ? <ProhibitIcon className={slots.statusIcon()} />
            : <CheckCircleIcon className={slots.statusIcon()} />}
          <Text size={2} color="current">
            {state === "REJECTED" ? t("document.rejected") : t("document.approved")}
          </Text>
        </div>
      )
    : (
        <Text size={2} color="neutral">
          {t("document.instruction")}
        </Text>
      );

  return (
    <DocumentRequestPanel eyebrow={t("document.eyebrow")} title={title} detail={detail}>
      {decided
        ? (
            <div className={slots.actions()}>
              <Button
                variant="soft"
                color="neutral"
                size={3}
                highContrast
                className="w-full"
                iconStart={!queueActive ? <CaretLeftIcon /> : undefined}
                iconEnd={queueActive && hasNext ? <CaretRightIcon /> : undefined}
                loading={advancing}
                disabled={advancing}
                onClick={queueActive && hasNext ? onNext : onFinish}
              >
                {queueActive
                  ? (hasNext ? t("document.goToNext") : t("document.finish"))
                  : t("document.backToList")}
              </Button>
            </div>
          )
        : (
            <div className={slots.actions()}>
              <div className={slots.actionRow()}>
                <Button
                  variant="soft"
                  color="neutral"
                  size={3}
                  highContrast
                  iconStart={<ProhibitIcon />}
                  disabled={busy}
                  onClick={onReject}
                >
                  {t("document.reject")}
                </Button>
                <Button
                  variant="solid"
                  color="indigo"
                  size={3}
                  className="min-w-0 flex-1"
                  iconStart={<CheckIcon />}
                  loading={busy}
                  onClick={onApprove}
                >
                  {t("document.reviewAndApprove")}
                </Button>
              </div>
              {consentText !== "" && (
                <Text size={1} className={slots.consent()}>
                  {consentText}
                </Text>
              )}
            </div>
          )}
    </DocumentRequestPanel>
  );
}
