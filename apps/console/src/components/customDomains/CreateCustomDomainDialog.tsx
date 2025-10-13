import { useTranslate } from "@probo/i18n";
import { graphql } from "react-relay";
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  useDialogRef,
  Breadcrumb,
} from "@probo/ui";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import type { ReactNode } from "react";
import type { CreateCustomDomainDialogMutation } from "./__generated__/CreateCustomDomainDialogMutation.graphql";

const createCustomDomainMutation = graphql`
  mutation CreateCustomDomainDialogMutation($input: CreateCustomDomainInput!) {
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
      "Please enter a valid domain (e.g., compliance.example.com)"
    ),
});

type FormData = z.infer<typeof schema>;

interface CreateCustomDomainDialogProps {
  children: ReactNode;
  organizationId: string;
}

export function CreateCustomDomainDialog({
  children,
  organizationId,
}: CreateCustomDomainDialogProps) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();

  const { register, handleSubmit, formState, reset } = useFormWithSchema(
    schema,
    {
      defaultValues: {
        domain: "",
      },
    }
  );

  const [createCustomDomain, isCreating] =
    useMutationWithToasts<CreateCustomDomainDialogMutation>(
      createCustomDomainMutation,
      {
        successMessage: __(
          "Domain added successfully. Configure the DNS records to verify and activate your domain."
        ),
        errorMessage: __("Failed to add domain"),
      }
    );

  const onSubmit = handleSubmit(async (data: FormData) => {
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
            data.createCustomDomain.customDomain.id
          );
          if (customDomainRecord) {
            organizationRecord.setLinkedRecord(
              customDomainRecord,
              "customDomain"
            );
          }
        }
      },
      onSuccess: () => {
        reset();
        dialogRef.current?.close();
      },
    });
  });

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[__("Custom Domain"), __("Add Domain")]} />}
    >
      <form onSubmit={onSubmit}>
        <DialogContent padded className="space-y-4">
          <div>
            <p className="text-sm text-txt-secondary mb-4">
              {__(
                "Enter your domain and we'll generate the DNS records you need to add"
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
              <strong>{__("Examples:")}</strong> compliance.example.com,
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
