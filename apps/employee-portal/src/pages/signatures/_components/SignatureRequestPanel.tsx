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

import { CaretLeftIcon, CaretRightIcon, CheckIcon } from "@phosphor-icons/react";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { DocumentRequestPanel } from "#/pages/_components/DocumentRequestPanel";
import { documentRequestPanel } from "#/pages/_components/variants";

interface SignatureRequestPanelProps {
  title: string;
  signed: boolean;
  consentText: string;
  queueActive: boolean;
  hasNext: boolean;
  busy: boolean;
  advancing?: boolean;
  isCurrentVersion?: boolean;
  onSign: () => void;
  onNext: () => void;
  onFinish: () => void;
}

// Left-column signing chrome: pending CTA, queue Finish / Go to next, or
// back to the list when viewing a signed document outside the queue.
export function SignatureRequestPanel({
  title,
  signed,
  consentText,
  queueActive,
  hasNext,
  busy,
  advancing = false,
  isCurrentVersion = true,
  onSign,
  onNext,
  onFinish,
}: SignatureRequestPanelProps) {
  const { t } = useTranslation("signatures");
  const slots = documentRequestPanel({ tone: "signed" });

  const detail = signed
    ? (
        <div className={slots.status()}>
          <CheckIcon className={slots.statusIcon()} />
          <Text size={2} color="current">
            {t("document.signed")}
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
      {signed
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
              <Button
                variant="solid"
                color="indigo"
                size={3}
                className="w-full"
                iconStart={<CheckIcon />}
                loading={busy}
                disabled={!isCurrentVersion}
                onClick={onSign}
              >
                {t("document.reviewAndSign")}
              </Button>
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
