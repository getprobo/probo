import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Spinner,
  Textarea,
  useDialogRef,
  useToast,
} from "@probo/ui";
import { type RefObject } from "react";
import { graphql } from "relay-runtime";
import { useMutation } from "react-relay";
import type { PolicyPagePolicyFragment$data } from "../__generated__/PolicyPagePolicyFragment.graphql";
import type { UpdateVersionDialogCreateMutation } from "./__generated__/UpdateVersionDialogCreateMutation.graphql";
import type { UpdateVersionDialogUpdateMutation } from "./__generated__/UpdateVersionDialogUpdateMutation.graphql";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";

const createDraftPolicy = graphql`
  mutation UpdateVersionDialogCreateMutation(
    $input: CreateDraftPolicyVersionInput!
    $connections: [ID!]!
  ) {
    createDraftPolicyVersion(input: $input) {
      policyVersionEdge @prependEdge(connections: $connections) {
        node {
          id
          content
          status
          publishedAt
          version
          updatedAt
          signatures(first: 100) {
            edges {
              node {
                id
                state
              }
            }
          }
        }
      }
    }
  }
`;

const UpdatePolicyMutation = graphql`
  mutation UpdateVersionDialogUpdateMutation(
    $input: UpdatePolicyVersionInput!
  ) {
    updatePolicyVersion(input: $input) {
      policyVersion {
        id
        content
      }
    }
  }
`;

type Props = {
  policy: PolicyPagePolicyFragment$data;
  connectionId: string;
  ref: RefObject<{ open: () => void } | null>;
};

const versionSchema = z.object({
  content: z.string(),
});

export default function UpdateVersionDialog({
  policy,
  connectionId,
  ref,
}: Props) {
  const { __ } = useTranslate();
  const { toast } = useToast();
  const dialogRef = useDialogRef();

  const version = policy.versions.edges[0].node;
  const isDraft = version?.status === "DRAFT";
  const [createDraftPolicyVersion, isCreatingDraft] =
    useMutation<UpdateVersionDialogCreateMutation>(createDraftPolicy);
  const [updatePolicyVersion, isUpdating] =
    useMutationWithToasts<UpdateVersionDialogUpdateMutation>(
      UpdatePolicyMutation,
      {
        successMessage: __("Policy updated successfully."),
        errorMessage: __("Failed to update policy. Please try again."),
      }
    );
  const { handleSubmit, register } = useFormWithSchema(versionSchema, {
    defaultValues: {
      content: version.content,
    },
  });

  ref.current = {
    open: () => {
      dialogRef.current?.open();
      if (!isDraft) {
        createDraftPolicyVersion({
          variables: {
            input: {
              policyID: policy.id,
            },
            connections: [connectionId],
          },
          onCompleted: (_, errors) => {
            if (errors) {
              toast({
                variant: "error",
                title: __("Error creating draft"),
                description:
                  errors[0]?.message || __("An unknown error occurred"),
              });
              dialogRef.current?.close();
              return;
            }
          },
        });
      }
    },
  };

  if (!version) {
    return;
  }

  const onSubmit = handleSubmit((data) => {
    updatePolicyVersion({
      variables: {
        input: {
          policyVersionId: version.id,
          content: data.content,
        },
      },
      onSuccess: () => {
        dialogRef.current?.close();
      },
    });
  });

  return (
    <Dialog
      ref={dialogRef}
      title={<Breadcrumb items={[__("Policies"), __("Edit policy")]} />}
    >
      {isCreatingDraft && <Spinner centered />}
      <form onSubmit={onSubmit}>
        <DialogContent>
          <Textarea
            id="content"
            variant="ghost"
            autogrow
            required
            placeholder={__("Add description")}
            aria-label={__("Description")}
            className="p-6"
            {...register("content")}
          />
        </DialogContent>
        <DialogFooter>
          <Button disabled={isUpdating} type="submit">
            {__("Update policy")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
