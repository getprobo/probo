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
