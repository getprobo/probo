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

import { useCopy } from "@probo/hooks";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  IconCheckmark1,
  IconSquareBehindSquare2,
  IconWarning,
} from "@probo/ui";
import { clsx } from "clsx";
import { useTranslation } from "react-i18next";

export function OAuthTokenCredentialsDialog(props: {
  dialogRef: React.RefObject<{ open: () => void; close: () => void } | null>;
  token: string;
  onDone: () => void;
}) {
  const { dialogRef, token, onDone } = props;
  const { t } = useTranslation();
  const [isCopied, copy] = useCopy();

  return (
    <Dialog
      ref={dialogRef}
      title={<Breadcrumb items={[t("oauthTokensPage.title"), t("oauthTokenCredentialsDialog.title")]} />}
    >
      <DialogContent padded className="space-y-4">
        <div className="flex items-start gap-2 rounded-lg border border-border-danger bg-danger px-4 py-3 text-sm text-txt-danger">
          <IconWarning size={16} className="shrink-0 mt-0.5" />
          <p>
            {t("oauthTokenCredentialsDialog.description")}
          </p>
        </div>
        <code className="flex items-start gap-2 rounded-lg bg-subtle p-4 font-mono text-sm">
          <span className="break-all flex-1">{token}</span>
          <button
            type="button"
            className={clsx(
              "shrink-0 rounded p-1 hover:bg-bg-hover transition-colors cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed",
              isCopied && "text-success",
            )}
            onClick={() => copy(token)}
            disabled={!token}
            aria-label={isCopied ? t("oauthTokenCredentialsDialog.actions.copied") : t("oauthTokenCredentialsDialog.actions.copy")}
            title={isCopied ? t("oauthTokenCredentialsDialog.actions.copied") : t("oauthTokenCredentialsDialog.actions.copy")}
          >
            {isCopied
              ? <IconCheckmark1 size={16} />
              : <IconSquareBehindSquare2 size={16} />}
          </button>
        </code>
      </DialogContent>
      <DialogFooter>
        <Button onClick={onDone}>{t("oauthTokenCredentialsDialog.actions.done")}</Button>
      </DialogFooter>
    </Dialog>
  );
}
