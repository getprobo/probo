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
  Badge,
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Option,
  Select,
  Textarea,
  useDialogRef,
} from "@probo/ui";
import { forwardRef, useImperativeHandle, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql } from "react-relay";
import { z } from "zod";

import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useMutationWithToasts } from "#/hooks/useMutationWithToasts";

const updateApplicabilityStatementMutation = graphql`
    mutation EditControlDialogUpdateMutation($input: UpdateApplicabilityStatementInput!) {
        updateApplicabilityStatement(input: $input) {
            applicabilityStatement {
                id
                applicability
                justification
            }
        }
    }
`;

export type EditControlDialogRef = {
  open: (control: {
    applicabilityStatementId: string;
    sectionTitle: string;
    name: string;
    frameworkName: string;
    applicability: boolean;
    justification: string | null;
  }) => void;
};

const schema = z.object({
  applicability: z.boolean(),
  justification: z.string().optional(),
});

export const EditControlDialog = forwardRef<EditControlDialogRef>((_props, ref) => {
  const { t } = useTranslation();
  const dialogRef = useDialogRef();
  const [control, setControl] = useState<{
    applicabilityStatementId: string;
    sectionTitle: string;
    name: string;
    frameworkName: string;
    applicability: boolean;
    justification: string | null;
  } | null>(null);

  const [updateApplicabilityStatement, isUpdating] = useMutationWithToasts(
    updateApplicabilityStatementMutation,
    {
      successMessage: t("editApplicabilityStatementDialog.messages.updated"),
      errorMessage: t("editApplicabilityStatementDialog.errors.update"),
    },
  );

  const { register, handleSubmit, setValue, watch } = useFormWithSchema(schema, {
    defaultValues: {
      applicability: true,
      justification: "",
    },
  });
  const applicability = watch("applicability");

  useImperativeHandle(ref, () => ({
    open: (ctrl) => {
      setControl(ctrl);
      setValue("applicability", ctrl.applicability);
      setValue("justification", ctrl.justification || "");
      dialogRef.current?.open();
    },
  }));

  const onSubmit = async (data: z.infer<typeof schema>) => {
    if (!control) return;

    await updateApplicabilityStatement({
      variables: {
        input: {
          applicabilityStatementId: control.applicabilityStatementId,
          applicability: data.applicability,
          justification: !data.applicability ? data.justification || null : null,
        },
      },
      onSuccess: () => {
        dialogRef.current?.close();
        setControl(null);
      },
    });
  };

  return (
    <Dialog
      ref={dialogRef}
      className="max-w-lg"
      title={
        <Breadcrumb items={[t("editApplicabilityStatementDialog.breadcrumb.statementsOfApplicability"), t("editApplicabilityStatementDialog.breadcrumb.editStatement")]} />
      }
    >
      {control
        ? (
            <form onSubmit={e => void handleSubmit(onSubmit)(e)}>
              <DialogContent padded className="space-y-4">
                <div className="space-y-2">
                  <div className="text-sm font-medium text-txt-secondary">
                    {control.frameworkName}
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge size="md">{control.sectionTitle}</Badge>
                    <span className="text-base font-medium text-txt-primary">
                      {control.name}
                    </span>
                  </div>
                </div>

                <Field label={t("editApplicabilityStatementDialog.fields.applicability")}>
                  <Select
                    variant="editor"
                    value={applicability ? "yes" : "no"}
                    onValueChange={value =>
                      setValue("applicability", value === "yes")}
                  >
                    <Option value="yes">{t("editApplicabilityStatementDialog.options.yes")}</Option>
                    <Option value="no">{t("editApplicabilityStatementDialog.options.no")}</Option>
                  </Select>
                </Field>

                {!applicability && (
                  <Field label={t("editApplicabilityStatementDialog.fields.justification")}>
                    <Textarea
                      {...register("justification")}
                      placeholder={t("editApplicabilityStatementDialog.placeholders.justification")}
                      autogrow
                    />
                  </Field>
                )}
              </DialogContent>
              <DialogFooter>
                <Button type="submit" disabled={isUpdating}>
                  {t("editApplicabilityStatementDialog.actions.save")}
                </Button>
              </DialogFooter>
            </form>
          )
        : null}
    </Dialog>
  );
});

EditControlDialog.displayName = "EditControlDialog";
