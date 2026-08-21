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

import { Form } from "@base-ui/react/form";
import { Button } from "@probo/ui/src/v2/Button/Button";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogClose } from "@probo/ui/src/v2/Dialog/DialogClose";
import { DialogDescription } from "@probo/ui/src/v2/Dialog/DialogDescription";
import { DialogFooter } from "@probo/ui/src/v2/Dialog/DialogFooter";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { Field } from "@probo/ui/src/v2/form/Field";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactElement, useState } from "react";
import { useTranslation } from "react-i18next";
import { useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { NewCompliancePortalDomainDialogMutation } from "#/__generated__/core/NewCompliancePortalDomainDialogMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { domainFormDialog } from "../variants";

const createCustomDomainMutation = graphql`
  mutation NewCompliancePortalDomainDialogMutation($input: CreateCustomDomainInput!) {
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
        ...CompliancePortalDomainCardFragment
      }
    }
  }
`;

const DOMAIN_PATTERN
  = /^(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]$/;

function normalizeDomain(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/^https?:\/\//, "")
    .replace(/\/$/, "");
}

interface NewCompliancePortalDomainDialogProps {
  children: ReactElement;
}

export function NewCompliancePortalDomainDialog({
  children,
}: NewCompliancePortalDomainDialogProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const { form, fields, examples } = domainFormDialog();
  const [open, setOpen] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const [createCustomDomain, isCreating]
    = useMutation<NewCompliancePortalDomainDialogMutation>(createCustomDomainMutation, {
      successMessage: t("newDomainDialog.messages.created"),
      errorToast: t("newDomainDialog.errors.create"),
    });

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setErrors({});
    }
  }

  function handleSubmit(formValues: Record<string, string>) {
    if (compliancePortalId == null) {
      return;
    }

    const domain = normalizeDomain(formValues.domain ?? "");
    if (!DOMAIN_PATTERN.test(domain)) {
      setErrors({ domain: t("newDomainDialog.fields.domainInvalid") });
      return;
    }

    void createCustomDomain({
      variables: {
        input: {
          compliancePortalId,
          domain,
        },
      },
      updater: (store, data) => {
        const newDomainId = data?.createCustomDomain?.customDomain?.id;
        if (!newDomainId) {
          return;
        }

        const compliancePortalRecord = store.get(compliancePortalId);
        const newDomainRecord = store.get(newDomainId);
        if (compliancePortalRecord && newDomainRecord) {
          compliancePortalRecord.setLinkedRecord(newDomainRecord, "customDomain");
        }
      },
    }).then(
      () => {
        handleOpenChange(false);
      },
      () => {
        // Error toast is already shown by useMutation.
      },
    );
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={children} />
      <DialogPopup>
        <Form
          className={form()}
          errors={errors}
          onFormSubmit={handleSubmit}
        >
          <DialogHeader>
            <DialogTitle>{t("newDomainDialog.title")}</DialogTitle>
            <DialogDescription>{t("newDomainDialog.description")}</DialogDescription>
          </DialogHeader>
          <DialogBody>
            <div className={fields()}>
              <Field
                label={t("newDomainDialog.fields.domain")}
                error={errors.domain}
              >
                <TextField
                  name="domain"
                  required
                  placeholder={t("newDomainDialog.fields.domainPlaceholder")}
                  autoFocus
                  onValueChange={() => setErrors({})}
                />
              </Field>
              <div className={examples()}>
                <Text size={1} color="faint">{t("newDomainDialog.examples")}</Text>
              </div>
            </div>
          </DialogBody>
          <DialogFooter>
            <DialogClose
              render={(
                <Button variant="soft" color="neutral">
                  {t("newDomainDialog.actions.cancel")}
                </Button>
              )}
            />
            <Button
              type="submit"
              variant="solid"
              color="neutral"
              highContrast
              loading={isCreating}
            >
              {t("newDomainDialog.actions.add")}
            </Button>
          </DialogFooter>
        </Form>
      </DialogPopup>
    </Dialog>
  );
}
