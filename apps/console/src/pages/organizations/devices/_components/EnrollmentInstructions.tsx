// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { CopyIcon } from "@phosphor-icons/react";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { Heading } from "@probo/ui/src/v2/typography/Heading";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";

import { enrollment } from "./variants";

interface EnrollmentInstructionsProps {
  enrollmentToken: string;
}

export function EnrollmentInstructions(
  { enrollmentToken }: EnrollmentInstructionsProps,
) {
  const { t } = useTranslation("organizations/devices");
  const { root } = enrollment();

  return (
    <div className={root()}>
      <Heading level={2} size={3} weight="medium">
        {t("deviceEnrollment.token.title")}
      </Heading>
      <Text size={2} color="faint">{t("deviceEnrollment.token.description")}</Text>
      <EnrollmentToken token={enrollmentToken} />
    </div>
  );
}

function EnrollmentToken({ token }: { token: string }) {
  const { t } = useTranslation("organizations/devices");
  const toast = useToast();
  const { copyField, copyValue, copyCode } = enrollment();

  function handleCopy() {
    const onCopyFailure = () => {
      toast.add({
        title: t("common.error"),
        description: t("deviceEnrollment.errors.copyToClipboard"),
        type: "error",
      });
    };
    if (!navigator.clipboard?.writeText) {
      onCopyFailure();
      return;
    }
    navigator.clipboard.writeText(token).then(
      () => {
        toast.add({
          title: t("common.messages.copied"),
          type: "success",
        });
      },
      onCopyFailure,
    );
  }

  return (
    <Card size={2} variant="soft">
      <div className={copyField()}>
        <Text size={1} color="faint">{t("deviceEnrollment.token.label")}</Text>
        <div className={copyValue()}>
          <Code size={2} className={copyCode()}>{token}</Code>
          <IconButton
            type="button"
            size={1}
            variant="soft"
            color="neutral"
            aria-label={t("common.actions.copy")}
            onClick={handleCopy}
          >
            <CopyIcon />
          </IconButton>
        </div>
      </div>
    </Card>
  );
}
