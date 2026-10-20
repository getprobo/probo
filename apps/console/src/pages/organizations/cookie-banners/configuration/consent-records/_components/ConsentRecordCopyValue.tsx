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

import { CopyIcon } from "@phosphor-icons/react";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Code } from "@probo/ui/src/v2/typography/Code";
import { useTranslation } from "react-i18next";

import { consentRecordPage } from "../../../variants";

interface ConsentRecordCopyValueProps {
  label: string;
  value: string | null | undefined;
}

export function ConsentRecordCopyValue({ label, value }: ConsentRecordCopyValueProps) {
  const { t } = useTranslation("organizations/cookie-banners");
  const toast = useToast();
  const { copyRow, value: valueClass } = consentRecordPage();
  const text = value == null || value === "" ? null : value;

  async function handleCopy() {
    if (text == null) {
      return;
    }
    try {
      await navigator.clipboard.writeText(text);
      toast.add({
        title: t("consentRecordPage.messages.copiedTitle"),
        description: label,
        type: "success",
      });
    } catch {
      toast.add({
        title: t("consentRecordPage.errors.copyTitle"),
        description: t("consentRecordPage.errors.copy"),
        type: "error",
      });
    }
  }

  return (
    <div className={copyRow()}>
      <Code size={1} className={valueClass()}>{text ?? "-"}</Code>
      {text != null && (
        <IconButton
          size={1}
          variant="ghost"
          color="neutral"
          aria-label={t("consentRecordPage.actions.copy", { label })}
          onClick={() => {
            void handleCopy();
          }}
        >
          <CopyIcon />
        </IconButton>
      )}
    </div>
  );
}
