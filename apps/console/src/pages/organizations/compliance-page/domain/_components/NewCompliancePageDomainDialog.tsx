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
import { useTranslation } from "react-i18next";
import { graphql } from "relay-runtime";
import { z } from "zod";

import type { NewCompliancePageDomainDialogMutation } from "#/__generated__/core/NewCompliancePageDomainDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutation } from "#/lib/relay/useMutation";

const createCustomDomainMutation = graphql`
  mutation NewCompliancePageDomainDialogMutation($input: CreateCustomDomainInput!) {
    createCustomDomain(input: $input) {
      customDomain {
        id
        domain
        certificate {
          status
          expiresAt
          provisioningError
        }
        dnsRecords {
          type
          name
          value
          ttl
          purpose
        }
        createdAt
        updatedAt
        canDelete: permission(action: "compliance-portal:custom-domain:delete")
        ...CompliancePageDomainCardFragment
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

export function NewCompliancePageDomainDialog(props: PropsWithChildren<{ compliancePageId: string }>) {
  const { children, compliancePageId } = props;

  const { t } = useTranslation("organizations/compliance-page");
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
    = useMutation<NewCompliancePageDomainDialogMutation>(createCustomDomainMutation, {
      successMessage: t("newDomainDialog.messages.created"),
      errorToast: t("newDomainDialog.errors.create"),
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
          compliancePortalId: compliancePageId,
          domain: normalizedDomain,
        },
      },
      updater: (store, data) => {
        const newDomainId = data?.createCustomDomain?.customDomain?.id;
        if (!newDomainId) {
          return;
        }

        const compliancePageRecord = store.get(compliancePageId);
        const newDomainRecord = store.get(newDomainId);
        if (compliancePageRecord && newDomainRecord) {
          compliancePageRecord.setLinkedRecord(newDomainRecord, "customDomain");
        }
      },
    });

    reset();
    dialogRef.current?.close();
  };

  return (
    <Dialog
      ref={dialogRef}
      trigger={children}
      title={<Breadcrumb items={[t("domainPage.title"), t("newDomainDialog.title")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <div>
            <p className="text-sm text-txt-secondary mb-4">
              {t("newDomainDialog.description")}
            </p>
          </div>

          <Field
            {...register("domain")}
            label={t("newDomainDialog.fields.domain")}
            type="text"
            placeholder={t("newDomainDialog.fields.domainPlaceholder")}
            error={formState.errors.domain?.message}
            autoFocus
          />

          <div className="bg-subtle rounded-lg p-4">
            <p className="text-xs text-txt-secondary">
              {t("newDomainDialog.examples")}
            </p>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button type="submit" disabled={isCreating || !formState.isValid}>
            {isCreating ? t("newDomainDialog.actions.adding") : t("newDomainDialog.actions.add")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
