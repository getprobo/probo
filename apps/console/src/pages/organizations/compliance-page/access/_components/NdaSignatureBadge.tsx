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
import { Badge } from "@probo/ui";

type ElectronicSignatureStatus = "PENDING" | "ACCEPTED" | "PROCESSING" | "COMPLETED" | "FAILED";

export function NdaSignatureBadge({ status }: { status: ElectronicSignatureStatus }) {
  const { __ } = useTranslate();

  switch (status) {
    case "COMPLETED":
      return <Badge variant="success">{__("Signed")}</Badge>;
    case "ACCEPTED":
    case "PROCESSING":
      return <Badge variant="info">{__("Processing")}</Badge>;
    case "PENDING":
      return <Badge variant="warning">{__("Pending")}</Badge>;
    case "FAILED":
      return <Badge variant="danger">{__("Failed")}</Badge>;
  }
}
