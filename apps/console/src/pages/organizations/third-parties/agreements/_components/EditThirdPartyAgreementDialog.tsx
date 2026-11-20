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
import {
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Input,
  Option,
  Select,
  Spinner,
  useDialogRef,
} from "@probo/ui";
import useToast from "@probo/ui/src/v2/Toaster/useToast";
import { Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import { graphql, useFragment, useMutation } from "react-relay";

import type { EditThirdPartyAgreementDialog_agreement$key } from "#/__generated__/core/EditThirdPartyAgreementDialog_agreement.graphql";
import type { EditThirdPartyAgreementDialogMutation } from "#/__generated__/core/EditThirdPartyAgreementDialogMutation.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { z } from "#/lib/zod";

import { THIRD_PARTY_AGREEMENT_TYPES } from "./agreementTypes";

const agreementFragment = graphql`
  fragment EditThirdPartyAgreementDialog_agreement on ThirdPartyAgreement {
    id
    type
    validity {
      start
      end
    }
  }
`;

const updateAgreementMutation = graphql`
  mutation EditThirdPartyAgreementDialogMutation(
    $input: UpdateThirdPartyAgreementInput!
  ) {
    updateThirdPartyAgreement(input: $input) {
      thirdPartyAgreement {
        ...ThirdPartyAgreementListItem_agreement
      }
    }
  }
`;

type Props = {
  children: React.ReactNode;
  agreementKey: EditThirdPartyAgreementDialog_agreement$key;
};

export function EditThirdPartyAgreementDialog({ children, agreementKey }: Props) {
  const { t } = useTranslation();
  const agreement = useFragment(agreementFragment, agreementKey);
  const ref = useDialogRef();
  const schema = z.object({
    type: z.enum(THIRD_PARTY_AGREEMENT_TYPES),
    validFrom: z.string().optional(),
    validUntil: z.string().optional(),
  });

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    control,
  } = useFormWithSchema(schema, {
    defaultValues: formValues(agreement),
  });

  const toast = useToast();
  const [updateAgreement, isUpdating]
    = useMutation<EditThirdPartyAgreementDialogMutation>(updateAgreementMutation);

  const onSubmit = (data: z.infer<typeof schema>) => {
    updateAgreement({
      variables: {
        input: {
          id: agreement.id,
          type: data.type,
          validity: {
            start: formatDatetime(data.validFrom),
            end: formatDatetime(data.validUntil),
          },
        },
      },
      onCompleted(_response, errors) {
        if (errors) {
          toast.add({
            title: t("editThirdPartyAgreementDialog.messages.error"),
            description: formatError(
              t("editThirdPartyAgreementDialog.errors.update"),
              errors,
            ),
            type: "error",
          });
          return;
        }
        toast.add({
          title: t("editThirdPartyAgreementDialog.messages.success"),
          description: t("editThirdPartyAgreementDialog.messages.updated"),
          type: "success",
        });
        ref.current?.close();
      },
      onError(error) {
        toast.add({
          title: t("editThirdPartyAgreementDialog.messages.error"),
          description: formatError(
            t("editThirdPartyAgreementDialog.errors.update"),
            error,
          ),
          type: "error",
        });
      },
    });
  };

  return (
    <Dialog
      title={t("editThirdPartyAgreementDialog.title")}
      ref={ref}
      trigger={children}
      className="max-w-lg"
      onOpenChange={(open) => {
        if (open) {
          reset(formValues(agreement));
        }
      }}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
        <DialogContent padded className="space-y-4">
          <Controller
            control={control}
            name="type"
            render={({ field }) => (
              <Field
                label={t("editThirdPartyAgreementDialog.fields.type")}
                required
                error={errors.type?.message}
              >
                <Select
                  variant="editor"
                  value={field.value}
                  onValueChange={field.onChange}
                  className="w-full"
                >
                  {THIRD_PARTY_AGREEMENT_TYPES.map(type => (
                    <Option key={type} value={type}>
                      {t(`thirdPartyAgreementsPage.types.${type}`)}
                    </Option>
                  ))}
                </Select>
              </Field>
            )}
          />

          <div className="grid grid-cols-2 gap-4">
            <Field label={t("editThirdPartyAgreementDialog.fields.validFrom")}>
              <Input {...register("validFrom")} type="date" />
            </Field>
            <Field label={t("editThirdPartyAgreementDialog.fields.validUntil")}>
              <Input {...register("validUntil")} type="date" />
            </Field>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button
            type="submit"
            disabled={isUpdating}
            icon={isUpdating ? Spinner : undefined}
          >
            {t("editThirdPartyAgreementDialog.actions.update")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}

function formValues(agreement: {
  type: (typeof THIRD_PARTY_AGREEMENT_TYPES)[number];
  validity: { start?: string | null; end?: string | null } | null | undefined;
}) {
  return {
    type: agreement.type,
    validFrom: dateInputValue(agreement.validity?.start),
    validUntil: dateInputValue(agreement.validity?.end),
  };
}

function dateInputValue(datetime?: string | null) {
  if (!datetime) {
    return "";
  }

  return datetime.split("T")[0] ?? "";
}

function formatDatetime(dateString?: string) {
  if (!dateString) {
    return null;
  }

  return `${dateString}T00:00:00Z`;
}
