import { useTranslate } from "@probo/i18n";
import { Button, Dialog, DialogContent, DialogFooter, type DialogRef, Field, Spinner } from "@probo/ui";
import { type DataID, graphql } from "relay-runtime";
import { z } from "zod";

import type { NewCompliancePageAccessDialogMutation } from "#/__generated__/core/NewCompliancePageAccessDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const createCompliancePageAccessMutation = graphql`
  mutation NewCompliancePageAccessDialogMutation(
    $input: CreateTrustCenterAccessInput!
    $connections: [ID!]!
  ) {
    createTrustCenterAccess(input: $input) {
      trustCenterAccessEdge @prependEdge(connections: $connections) {
        cursor
        node {
          id
          ...CompliancePageAccessListItemFragment
        }
      }
    }
  }
`;

export function NewCompliancePageAccessDialog(props: {
  connectionId: DataID;
  compliancePageId: string;
  ref: DialogRef;
  onCreate: (accessId: string) => void;
}) {
  const { connectionId, compliancePageId, ref, onCreate } = props;

  const { __ } = useTranslate();

  const inviteSchema = z.object({
    name: z
      .string()
      .min(1, __("Name is required"))
      .min(2, __("Name must be at least 2 characters long")),
    email: z
      .string()
      .min(1, __("Email is required"))
      .email(__("Please enter a valid email address")),
  });
  const inviteForm = useFormWithSchema(inviteSchema, {
    defaultValues: { name: "", email: "" },
  });

  const [createInvitation, isCreating] = useMutationWithToasts<NewCompliancePageAccessDialogMutation>(
    createCompliancePageAccessMutation,
    {
      successMessage: __("Access created successfully"),
      errorMessage: __("Failed to create access"),
    },
  );
  const handleInvite = async (data: z.infer<typeof inviteSchema>) => {
    const email = data.email.trim();

    await createInvitation({
      variables: {
        input: {
          trustCenterId: compliancePageId,
          email: email,
          name: data.name.trim(),
        },
        connections: connectionId ? [connectionId] : [],
      },
      onCompleted: (response, errors) => {
        if (errors?.length) {
          return;
        }

        const newAccess = response.createTrustCenterAccess.trustCenterAccessEdge.node;
        onCreate(newAccess.id);
        setTimeout(() => {
          inviteForm.reset();
          ref.current?.close();
        }, 50);
        setTimeout(() => {
          inviteForm.reset();
        }, 300);
      },
    });
  };

  return (
    <Dialog ref={ref} title={__("Invite External Access")}>
      <form onSubmit={e => void inviteForm.handleSubmit(handleInvite)(e)}>
        <DialogContent padded className="space-y-6">
          <div>
            <p className="text-txt-secondary text-sm mb-4">
              {__("Give a person access to your compliance page")}
            </p>

            <Field
              label={__("Full Name")}
              required
              error={inviteForm.formState.errors.name?.message}
              {...inviteForm.register("name")}
              placeholder={__("John Doe")}
            />

            <div className="mt-4">
              <Field
                label={__("Email Address")}
                required
                error={inviteForm.formState.errors.email?.message}
                type="email"
                {...inviteForm.register("email")}
                placeholder={__("john@example.com")}
              />
            </div>
          </div>
        </DialogContent>

        <DialogFooter>
          <Button type="submit" disabled={isCreating}>
            {isCreating && <Spinner />}
            {__("Create Access")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
