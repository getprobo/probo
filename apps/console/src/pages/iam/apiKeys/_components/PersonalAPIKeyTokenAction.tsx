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

import { formatError } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { Button, useDialogRef, useToast } from "@probo/ui";
import { useRefetchableFragment } from "react-relay";

import type { PersonalAPIKeyRowFragment$key } from "#/__generated__/iam/PersonalAPIKeyRowFragment.graphql";
import type { PersonalAPIKeyRowRefetchQuery } from "#/__generated__/iam/PersonalAPIKeyRowRefetchQuery.graphql";

import { personalAPIKeyRowFragment } from "./PersonalAPIKeyRow";
import { PersonalAPIKeyTokenDialog } from "./PersonalAPIKeyTokenDialog";

export function PersonalAPIKeyTokenAction(props: {
  fKey: PersonalAPIKeyRowFragment$key;
  disabled?: boolean;
}) {
  const { fKey, disabled } = props;
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const [data, refetch] = useRefetchableFragment<
    PersonalAPIKeyRowRefetchQuery,
    PersonalAPIKeyRowFragment$key
  >(personalAPIKeyRowFragment, fKey);

  const handleShow = () => {
    dialogRef.current?.open();

    refetch(
      { includeToken: true },
      {
        fetchPolicy: "network-only",
        onComplete: (error) => {
          if (error) {
            toast({
              title: __("Error"),
              description: formatError(
                __("Failed to load API key token."),
                error,
              ),
              variant: "error",
            });
            dialogRef.current?.close();
          }
        },
      },
    );
  };

  return (
    <>
      <Button variant="secondary" onClick={handleShow} disabled={!!disabled}>
        {__("Show")}
      </Button>

      <PersonalAPIKeyTokenDialog
        dialogRef={dialogRef}
        token={data.token ?? ""}
        onDone={() => {
          dialogRef.current?.close();
        }}
      />
    </>
  );
}
