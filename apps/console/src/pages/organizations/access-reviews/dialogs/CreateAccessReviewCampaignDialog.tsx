import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
} from "@probo/ui";
import type { ReactNode } from "react";
import { graphql } from "react-relay";
import { useNavigate } from "react-router";
import { z } from "zod";

import type { CreateAccessReviewCampaignDialogMutation } from "#/__generated__/core/CreateAccessReviewCampaignDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const createMutation = graphql`
  mutation CreateAccessReviewCampaignDialogMutation(
    $input: CreateAccessReviewCampaignInput!
    $connections: [ID!]!
  ) {
    createAccessReviewCampaign(input: $input) {
      accessReviewCampaignEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          status
          createdAt
          ...AccessReviewCampaignRowFragment
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  accessReviewId: string;
  connectionId: string;
};

const schema = z.object({
  name: z.string().min(1),
});

export function CreateAccessReviewCampaignDialog({
  children,
  accessReviewId,
  connectionId,
}: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const { register, handleSubmit, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        name: "",
      },
    },
  );
  const ref = useDialogRef();

  const [createCampaign, isCreating]
    = useMutationWithToasts<CreateAccessReviewCampaignDialogMutation>(
      createMutation,
      {
        successMessage: __("Campaign created successfully."),
        errorMessage: __("Failed to create campaign"),
      },
    );

  const onSubmit = async (data: z.infer<typeof schema>) => {
    await createCampaign({
      variables: {
        input: {
          accessReviewId,
          name: data.name,
        },
        connections: [connectionId],
      },
      onCompleted: (response) => {
        reset();
        ref.current?.close();
        const campaignId
          = response.createAccessReviewCampaign.accessReviewCampaignEdge
            .node.id;
        void navigate(
          `/organizations/${organizationId}/access-reviews/campaigns/${campaignId}`,
        );
      },
    });
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={(
        <Breadcrumb
          items={[
            __("Access Reviews"),
            __("New Campaign"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Field
            label={__("Name")}
            {...register("name")}
            type="text"
            required
            placeholder={__("Q1 2026 Access Review")}
          />
        </DialogContent>
        <DialogFooter>
          <Button disabled={isCreating} type="submit">
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
