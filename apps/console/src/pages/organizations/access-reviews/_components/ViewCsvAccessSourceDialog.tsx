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
import {
  Breadcrumb,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  Textarea,
} from "@probo/ui";
import { Suspense } from "react";
import { graphql, useLazyLoadQuery } from "react-relay";

import type { ViewCsvAccessSourceDialogQuery } from "#/__generated__/core/ViewCsvAccessSourceDialogQuery.graphql";

const viewCsvAccessSourceDialogQuery = graphql`
  query ViewCsvAccessSourceDialogQuery($accessSourceId: ID!) {
    node(id: $accessSourceId) @required(action: THROW) {
      ... on AccessSource {
        id
        name
        csvData
      }
    }
  }
`;

type Props = {
  accessSourceId: string;
  name: string;
  onClose: () => void;
};

export function ViewCsvAccessSourceDialog({
  accessSourceId,
  name,
  onClose,
}: Props) {
  const { __ } = useTranslate();

  return (
    <Dialog
      defaultOpen
      onClose={onClose}
      title={
        <Breadcrumb items={[{ label: __("CSV access source") }, { label: name }]} />
      }
      className="max-w-3xl"
    >
      <DialogContent padded>
        <Suspense
          fallback={
            <div className="flex items-center justify-center py-8">
              <Spinner />
            </div>
          }
        >
          <ViewCsvAccessSourceContent accessSourceId={accessSourceId} />
        </Suspense>
      </DialogContent>
      <DialogFooter exitLabel={__("Close")} />
    </Dialog>
  );
}

function ViewCsvAccessSourceContent({
  accessSourceId,
}: {
  accessSourceId: string;
}) {
  const { __ } = useTranslate();
  const data = useLazyLoadQuery<ViewCsvAccessSourceDialogQuery>(
    viewCsvAccessSourceDialogQuery,
    { accessSourceId },
    { fetchPolicy: "store-or-network" },
  );

  const csvData = data.node.csvData;

  if (!csvData) {
    return (
      <p className="text-txt-secondary text-sm">
        {__("This access source does not have any CSV content stored.")}
      </p>
    );
  }

  return (
    <Textarea
      readOnly
      defaultValue={csvData}
      rows={20}
      className="font-mono text-xs whitespace-pre overflow-auto"
    />
  );
}
