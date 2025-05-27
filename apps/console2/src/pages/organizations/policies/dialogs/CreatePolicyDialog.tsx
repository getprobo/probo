import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Input,
  Label,
  PropertyRow,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import { useState, type ReactNode } from "react";
import { graphql } from "relay-runtime";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { usePolicyForm } from "../../../../hooks/forms/usePolicyForm";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { PeopleSelect } from "/components/form/PeopleSelect";
import type { CreatePolicyDialogMutation } from "./__generated__/CreatePolicyDialogMutation.graphql";

type Props = {
  trigger?: ReactNode;
  connection: string;
};

const createPolicyMutation = graphql`
  mutation CreatePolicyDialogMutation(
    $input: CreatePolicyInput!
    $connections: [ID!]!
  ) {
    createPolicy(input: $input) {
      policyEdge @prependEdge(connections: $connections) {
        node {
          ...PoliciesPageRowFragment
        }
      }
    }
  }
`;

/**
 * Dialog to create or update a policy
 */
export function CreatePolicyDialog({ trigger, connection }: Props) {
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

  const { control, handleSubmit, register, formState, reset } = usePolicyForm();
  const errors = formState.errors ?? {};
  const [createPolicy, isLoading] =
    useMutationWithToasts<CreatePolicyDialogMutation>(createPolicyMutation);

  const onSubmit = handleSubmit((data) => {
    createPolicy({
      variables: {
        input: {
          ...data,
          organizationId,
        },
        connections: [connection!],
      },
      successMessage: __("Policy created successfully."),
      errorMessage: __("Failed to create policy. Please try again."),
      onSuccess: () => {
        dialogRef.current?.close();
        reset();
      },
    });
  });

  const dialogRef = useDialogRef();

  return (
    <Dialog
      ref={dialogRef}
      trigger={trigger}
      title={<Breadcrumb items={[__("Policies"), __("New Policy")]} />}
    >
      <form onSubmit={onSubmit}>
        <DialogContent className="grid grid-cols-[1fr_420px]">
          <div className="py-8 px-10 space-y-4">
            <Input
              id="title"
              aria-label={__("Title")}
              required
              variant="title"
              placeholder={__("Policy title")}
              {...register("title")}
            />
            <Textarea
              id="content"
              variant="ghost"
              autogrow
              placeholder={__("Add description")}
              aria-label={__("Description")}
              {...register("content")}
            />
          </div>
          {/* Properties form */}
          <div className="py-5 px-6 bg-subtle">
            <Label>{__("Properties")}</Label>
            <PropertyRow label={__("Status")}>
              <Badge variant="neutral" size="md">
                {__("Draft")}
              </Badge>
            </PropertyRow>

            <PropertyRow
              id="ownerId"
              label={__("Owner")}
              error={errors.ownerId?.message}
            >
              <PeopleSelect
                name="ownerId"
                control={control}
                organization={organizationId}
              />
            </PropertyRow>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isLoading}>
            {__("Create policy")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
