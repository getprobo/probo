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
import type { PropsWithChildren } from "react";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { NewCompliancePageDomainDialogMutation } from "#/__generated__/core/NewCompliancePageDomainDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";
import { useOrganizationId } from "#/hooks/useOrganizationId";

const createCustomDomainMutation = graphql`
  mutation NewCompliancePageDomainDialogMutation($input: CreateCustomDomainInput!) {
    createCustomDomain(input: $input) {
      customDomain {
        id
        domain
        sslStatus
        dnsRecords {
          type
          name
          value
          ttl
          purpose
        }
        createdAt
        updatedAt
        sslExpiresAt
        canDelete: permission(action: "core:custom-domain:delete")
      }
    }
  }
`;

const schema = z.object({
  domain: z
    .string()
    .min(1, "Domain is required")
    .regex(
      /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$/i,
      "Please enter a valid domain (e.g., compliance.example.com)",
    ),
});

export function NewCompliancePageDomainDialog(props: PropsWithChildren) {
  const { children } = props;

  const organizationId = useOrganizationId();
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();

  const { register, handleSubmit, formState, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        domain: "",
      },
    },
  );

  const [createCustomDomain, isCreating]
    = useMutationWithToasts<NewCompliancePageDomainDialogMutation>(createCustomDomainMutation, {
      successMessage: __(
        "Domain added successfully. Configure the DNS records to verify and activate your domain.",
      ),
      errorMessage: __("Failed to add domain"),
    });

  const onSubmit = async (data: z.infer<typeof schema>) => {
    const normalizedDomain = data.domain
      .trim()
      .toLowerCase()
      .replace(/^https?:\/\//, "")
      .replace(/\/$/, "");

    await createCustomDomain({
      variables: {
        input: {
          organizationId,
          domain: normalizedDomain,
        },
      },
      updater: (store, data) => {
        // Update the cache by setting the new customDomain on the organization
        const organizationRecord = store.get(organizationId);
        if (organizationRecord && data?.createCustomDomain?.customDomain) {
          const customDomainRecord = store.get(
            data.createCustomDomain.customDomain.id,
          );
          if (customDomainRecord) {
            organizationRecord.setLinkedRecord(
              customDomainRecord,
              "customDomain",
            );
          }
        }
      },
      onSuccess: () => {
        reset();
        dialogRef.current?.close();
      },
    });
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Custom Domain"), __("Add Domain")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <div>
            <p className="text-sm text-txt-secondary mb-4">
              {__(
                "Enter your domain and we'll generate the DNS records you need to add",
              )}
            </p>
          </div>

          <Field
            {...register("domain")}
            label={__("Domain")}
            type="text"
            placeholder="compliance.example.com"
            error={formState.errors.domain?.message}
            autoFocus
          />

          <div className="bg-subtle rounded-lg p-4">
            <p className="text-xs text-txt-secondary">
              <strong>{__("Examples:")}</strong>
              {" "}
              compliance.example.com,
              trust.example.com
            </p>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating || !formState.isValid}>
            {isCreating ? __("Adding...") : __("Add Domain")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
