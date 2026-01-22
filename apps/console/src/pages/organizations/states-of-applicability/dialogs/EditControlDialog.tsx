import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Select,
  Option,
  Textarea,
  useDialogRef,
  Badge,
} from "@probo/ui";
import { forwardRef, useImperativeHandle, useState } from "react";
import { graphql } from "react-relay";
import { z } from "zod";

import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useFormWithSchema } from "/hooks/useFormWithSchema";

const linkControlMutation = graphql`
  mutation EditControlDialogLinkMutation($input: CreateStateOfApplicabilityControlMappingInput!) {
    createStateOfApplicabilityControlMapping(input: $input) {
      stateOfApplicabilityControlEdge {
        node {
          stateOfApplicabilityId
          controlId
          applicability
          justification
        }
      }
    }
  }
`;

export type EditControlDialogRef = {
  open: (control: {
    stateOfApplicabilityId: string;
    controlId: string;
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

export const EditControlDialog = forwardRef<EditControlDialogRef, { onSuccess?: () => void }>(
  ({ onSuccess }, ref) => {
    const { __ } = useTranslate();
    const dialogRef = useDialogRef();
    const [control, setControl] = useState<{
      stateOfApplicabilityId: string;
      controlId: string;
      sectionTitle: string;
      name: string;
      frameworkName: string;
      applicability: boolean;
      justification: string | null;
    } | null>(null);

    const [linkMutate, isLinking] = useMutationWithToasts(linkControlMutation, {
      successMessage: __("Control updated successfully."),
      errorMessage: __("Failed to update control"),
    });

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

      await linkMutate({
        variables: {
          input: {
            stateOfApplicabilityId: control.stateOfApplicabilityId,
            controlId: control.controlId,
            applicability: data.applicability,
            justification: !data.applicability ? data.justification || null : null,
          },
        },
        onSuccess: () => {
          dialogRef.current?.close();
          setControl(null);
          onSuccess?.();
        },
        updater: (store) => {
          const stateOfApplicability = store.get(control.stateOfApplicabilityId);
          if (stateOfApplicability) {
            stateOfApplicability.invalidateRecord();
          }
        },
      });
    };

    return (
      <Dialog
        ref={dialogRef}
        className="max-w-lg"
        title={(
          <Breadcrumb
            items={[__("States of Applicability"), __("Edit Control")]}
          />
        )}
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
                      <span className="text-base font-medium text-txt-primary">{control.name}</span>
                    </div>
                  </div>

                  <Field label={__("Applicability")}>
                    <Select
                      variant="editor"
                      value={applicability ? "yes" : "no"}
                      onValueChange={value => setValue("applicability", value === "yes")}
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
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => dialogRef.current?.close()}
                  >
                    {__("Cancel")}
                  </Button>
                  <Button type="submit" disabled={isLinking}>
                    {__("Save")}
                  </Button>
                </DialogFooter>
              </form>
            )
          : null}
      </Dialog>
    );
  },
);
