// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
import { Button, Spinner } from "@probo/ui";
import { graphql, useFragment } from "react-relay";

import type { VersionActionsFragment$key } from "#/__generated__/core/VersionActionsFragment.graphql";

const fragment = graphql`
  fragment VersionActionsFragment on EmployeeDocumentVersion {
    id
    signed
  }
`;

export function VersionActions({
  fKey,
  isSigning,
  onSign,
  onBack,
}: {
  fKey: VersionActionsFragment$key;
  isSigning: boolean;
  onSign: (versionId: string) => void;
  onBack: () => void;
}) {
  const { __ } = useTranslate();
  const versionData = useFragment<VersionActionsFragment$key>(fragment, fKey);
  const isSigned = versionData.signed;

  if (isSigned) {
    return (
      <>
        <Button onClick={onBack} className="h-10 w-full" variant="secondary">
          {__("Back to Documents")}
        </Button>
        <p className="text-xs text-txt-tertiary mt-2 h-5" />
      </>
    );
  }

  return (
    <>
      <Button
        onClick={() => onSign(versionData.id)}
        className="h-10 w-full"
        disabled={isSigning}
        icon={isSigning ? Spinner : undefined}
      >
        {__("I acknowledge and agree")}
      </Button>
      <p className="text-xs text-txt-tertiary mt-2 h-5">
        {__(
          "By clicking 'I acknowledge and agree', your digital signature will be recorded.",
        )}
      </p>
    </>
  );
}
