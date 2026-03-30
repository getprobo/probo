// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

import { useTranslate } from "@probo/i18n";
import { Dialog, useDialogRef } from "@probo/ui";
import { type PropsWithChildren } from "react";
import type { DataID } from "relay-runtime";

import { PersonForm } from "./PersonForm";

export function AddPersonDialog(props: PropsWithChildren<{
  connectionId: DataID;
}>) {
  const { children, connectionId } = props;
  const dialogRef = useDialogRef();
  const { __ } = useTranslate();

  return (
    <Dialog
      title={__("Add Person")}
      trigger={children}
      className="max-w-xl"
      ref={dialogRef}
    >
      <div className="p-4">
        <PersonForm connectionId={connectionId} onSubmit={() => dialogRef.current?.close()} />
      </div>
    </Dialog>
  );
}
