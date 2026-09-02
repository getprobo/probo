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
import { CopySimpleIcon } from "@phosphor-icons/react";
import { Card } from "@probo/ui/src/v2/Card/Card";
import { IconButton } from "@probo/ui/src/v2/IconButton/IconButton";
import { useTranslation } from "react-i18next";

import { copyableCodeBlock } from "./variants";

export interface CopyableCodeBlockProps {
  code: string;
}

export function CopyableCodeBlock({ code }: CopyableCodeBlockProps) {
  const { t } = useTranslation("devices");
  const { t: tApp } = useTranslation();
  const toast = Toast.useToastManager();
  const slots = copyableCodeBlock();

  function handleCopy() {
    const onFailure = () => {
      toast.add({
        title: tApp("common.error"),
        description: t("addManually.copyFailed"),
        type: "error",
      });
    };

    if (!navigator.clipboard?.writeText) {
      onFailure();
      return;
    }

    try {
      navigator.clipboard.writeText(code).then(
        () => {
          toast.add({ title: t("addManually.copied"), type: "success" });
        },
        onFailure,
      );
    } catch {
      onFailure();
    }
  }

  return (
    <Card variant="soft" size={2} padding="none" className={slots.root()}>
      <div className={slots.toolbar()}>
        <IconButton
          size={1}
          variant="ghost"
          color="neutral"
          aria-label={t("addManually.copy")}
          onClick={handleCopy}
        >
          <CopySimpleIcon />
        </IconButton>
      </div>
      <pre className={slots.pre()}>
        <code>{code}</code>
      </pre>
    </Card>
  );
}
