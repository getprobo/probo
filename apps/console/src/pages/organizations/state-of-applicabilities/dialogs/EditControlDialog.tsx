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
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { z } from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";

const linkControlMutation = graphql`
  mutation EditControlDialogLinkMutation($input: LinkStateOfApplicabilityControlInput!) {
    linkStateOfApplicabilityControl(input: $input) {
      stateOfApplicabilityControl {
        stateOfApplicabilityId
        controlId
        state
        exclusionJustification
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
    state: string | null;
    exclusionJustification: string | null;
  }) => void;
};

const schema = z.object({
  state: z.enum(["IMPLEMENTED", "NOT_IMPLEMENTED", "EXCLUDED"]),
  exclusionJustification: z.string().optional(),
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
      state: string | null;
      exclusionJustification: string | null;
    } | null>(null);

    const [linkMutate, isLinking] = useMutationWithToasts(linkControlMutation, {
      successMessage: __("Control updated successfully."),
      errorMessage: __("Failed to update control"),
    });

    const { register, handleSubmit, setValue, watch } = useFormWithSchema(schema, {
      defaultValues: {
        state: "IMPLEMENTED" as const,
        exclusionJustification: "",
      },
    });
    const state = watch("state");

    useImperativeHandle(ref, () => ({
      open: (ctrl) => {
        setControl(ctrl);
        setValue("state", (ctrl.state as "IMPLEMENTED" | "NOT_IMPLEMENTED" | "EXCLUDED") || "IMPLEMENTED");
        setValue("exclusionJustification", ctrl.exclusionJustification || "");
        dialogRef.current?.open();
      },
    }));

    const onSubmit = handleSubmit((data) => {
      if (!control) return;

      linkMutate({
        variables: {
          input: {
            stateOfApplicabilityId: control.stateOfApplicabilityId,
            controlId: control.controlId,
            state: data.state,
            exclusionJustification: data.state === "EXCLUDED" ? data.exclusionJustification || null : null,
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
    });

    return (
      <Dialog
        ref={dialogRef}
        className="max-w-lg"
        title={
          <Breadcrumb
            items={[__("States of Applicability"), __("Edit Control")]}
          />
        }
      >
        {control ? (
          <form onSubmit={onSubmit}>
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

              <Field label={__("State")}>
                <Select
                  variant="editor"
                  value={state || "IMPLEMENTED"}
                  onValueChange={(value) => setValue("state", value as "EXCLUDED" | "IMPLEMENTED" | "NOT_IMPLEMENTED")}
                >
                  <Option value="IMPLEMENTED">{__("Implemented")}</Option>
                  <Option value="NOT_IMPLEMENTED">{__("Not Implemented")}</Option>
                  <Option value="EXCLUDED">{__("Excluded")}</Option>
                </Select>
              </Field>

              {state === "EXCLUDED" && (
                <Field label={__("Exclusion Justification")}>
                  <Textarea
                    {...register("exclusionJustification")}
                    placeholder={__("Reason for exclusion")}
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
        ) : null}
      </Dialog>
    );
  }
);
