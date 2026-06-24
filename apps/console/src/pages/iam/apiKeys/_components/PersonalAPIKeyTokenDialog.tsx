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

import { useCopy } from "@probo/hooks";
import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
} from "@probo/ui";

export function PersonalAPIKeyTokenDialog(props: {
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
      title={<Breadcrumb items={[__("API Keys"), __("Token")]} />}
    >
      <DialogContent padded className="space-y-4">
        <div className="bg-gray-100 p-4 rounded-lg flex items-center gap-2">
          <code className="text-sm font-mono break-all flex-1">{token}</code>
          <Button
            variant="secondary"
            onClick={() => copy(token)}
            disabled={!token}
          >
            {isCopied ? __("Copied") : __("Copy")}
          </Button>
        </div>
      </DialogContent>
      <DialogFooter>
        <Button onClick={onDone}>{__("Done")}</Button>
      </DialogFooter>
    </Dialog>
  );
}
