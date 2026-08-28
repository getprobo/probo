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

import { CheckCircleIcon, CheckIcon, ProhibitIcon } from "@phosphor-icons/react";
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
  onApprove: () => void;
  onReject: () => void;
  onNext: () => void;
  onFinish: () => void;
}

// Left-column approval chrome: approve / reject, or the decided status plus
// Finish / Go to next when the employee is still in the queue.
export function ApprovalRequestPanel({
  title,
  state,
  consentText,
  queueActive,
  hasNext,
  busy,
  onApprove,
  onReject,
  onNext,
  onFinish,
}: ApprovalRequestPanelProps) {
  const { t } = useTranslation("approvals");
  const decided = state === "APPROVED" || state === "REJECTED";
  const tone = state === "REJECTED" ? "rejected" : "approved";
  const slots = documentRequestPanel({ tone });

  return (
    <DocumentRequestPanel eyebrow={t("document.eyebrow")} title={title}>
      {decided
        ? (
            <>
              <div className={slots.status()}>
                {state === "REJECTED"
                  ? <ProhibitIcon className={slots.statusIcon()} />
                  : <CheckCircleIcon className={slots.statusIcon()} />}
                <Text size={2} weight="medium" className="text-current">
                  {state === "REJECTED" ? t("document.rejected") : t("document.approved")}
                </Text>
              </div>
              {queueActive
                ? (
                    <div className={slots.actions()}>
                      <Button
                        variant="surface"
                        color="neutral"
                        size={2}
                        className="w-full"
                        onClick={hasNext ? onNext : onFinish}
                      >
                        {hasNext ? t("document.goToNext") : t("document.finish")}
                      </Button>
                    </div>
                  )
                : null}
            </>
          )
        : (
            <>
              <Text size={2} color="neutral">
                {t("document.instruction")}
              </Text>
              <div className={slots.actions()}>
                <div className={slots.actionRow()}>
                  <Button
                    variant="surface"
                    color="neutral"
                    size={2}
                    className="flex-1"
                    iconStart={<ProhibitIcon />}
                    disabled={busy}
                    onClick={onReject}
                  >
                    {t("document.reject")}
                  </Button>
                  <Button
                    variant="solid"
                    color="sky"
                    size={2}
                    className="flex-1"
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
            </>
          )}
    </DocumentRequestPanel>
  );
}
