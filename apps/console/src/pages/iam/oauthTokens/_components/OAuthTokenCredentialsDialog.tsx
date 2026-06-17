// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useCopy } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
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

export function OAuthTokenCredentialsDialog(props: {
  dialogRef: React.RefObject<{ open: () => void; close: () => void } | null>;
  token: string;
  onDone: () => void;
}) {
  const { dialogRef, token, onDone } = props;
  const { __ } = useTranslate();
  const [isCopied, copy] = useCopy();

  return (
    <Dialog
      ref={dialogRef}
      title={<Breadcrumb items={[__("OAuth tokens"), __("Token")]} />}
    >
      <DialogContent padded className="space-y-4">
        <div className="flex items-start gap-2 rounded-lg border border-border-danger bg-danger px-4 py-3 text-sm text-txt-danger">
          <IconWarning size={16} className="shrink-0 mt-0.5" />
          <p>
            {__(
              "Copy this bearer token now. You will not be able to see it again.",
            )}
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
            aria-label={isCopied ? __("Copied") : __("Copy")}
            title={isCopied ? __("Copied") : __("Copy")}
          >
            {isCopied
              ? <IconCheckmark1 size={16} />
              : <IconSquareBehindSquare2 size={16} />}
          </button>
        </code>
      </DialogContent>
      <DialogFooter>
        <Button onClick={onDone}>{__("Done")}</Button>
      </DialogFooter>
    </Dialog>
  );
}
