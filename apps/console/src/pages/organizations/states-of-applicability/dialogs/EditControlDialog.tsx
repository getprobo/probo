import { useTranslate } from "@probo/i18n";
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
  const { __ } = useTranslate();
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
      successMessage: __("Statement updated successfully."),
      errorMessage: __("Failed to update statement"),
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
        <Breadcrumb items={[__("States of Applicability"), __("Edit Statement")]} />
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

                <Field label={__("Applicability")}>
                  <Select
                    variant="editor"
                    value={applicability ? "yes" : "no"}
                    onValueChange={value =>
                      setValue("applicability", value === "yes")}
                  >
                    <Option value="yes">{__("Yes")}</Option>
                    <Option value="no">{__("No")}</Option>
                  </Select>
                </Field>

                {!applicability && (
                  <Field label={__("Justification")}>
                    <Textarea
                      {...register("justification")}
                      placeholder={__("Reason for non-applicability")}
                      autogrow
                    />
                  </Field>
                )}
              </DialogContent>
              <DialogFooter>
                <Button type="submit" disabled={isUpdating}>
                  {__("Save")}
                </Button>
              </DialogFooter>
            </form>
          )
        : null}
    </Dialog>
  );
});

EditControlDialog.displayName = "EditControlDialog";
