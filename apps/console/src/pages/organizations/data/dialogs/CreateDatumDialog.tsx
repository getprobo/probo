import { useTranslate } from "@probo/i18n";
import {
  Breadcrumb,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Option,
  useDialogRef,
} from "@probo/ui";
import { z } from "zod";

import { ControlledField } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { VendorsMultiSelectField } from "#/components/form/VendorsMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import { useCreateDatum } from "../../../../hooks/graph/DatumGraph";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  dataClassification: z.enum(["PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"]),
  ownerId: z.string().min(1, "Owner is required"),
  vendorIds: z.array(z.string()).optional(),
});

type Props = {
  children: React.ReactNode;
  connection: string;
  organizationId: string;
  onCreated?: () => void;
};

export function CreateDatumDialog({
  children,
  connection,
  organizationId,
  onCreated,
}: Props) {
  const { __ } = useTranslate();
  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(schema, {
      defaultValues: {
        name: "",
        dataClassification: "PUBLIC",
        ownerId: "",
        vendorIds: [],
      },
    });
  const ref = useDialogRef();
  const createDatum = useCreateDatum(connection);

  const onSubmit = async (data: z.infer<typeof schema>) => {
    try {
      await createDatum({
        ...data,
        organizationId,
      });
      ref.current?.close();
      reset();
      onCreated?.();
    } catch (error) {
      console.error("Failed to create datum:", error);
    }
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={<Breadcrumb items={[__("Data"), __("New Data")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <DialogContent padded className="space-y-4">
          <Field label={__("Name")} {...register("name")} type="text" />
          <ControlledField
            control={control}
            name="dataClassification"
            type="select"
            label={__("Classification")}
          >
            <Option value="PUBLIC">{__("Public")}</Option>
            <Option value="INTERNAL">{__("Internal")}</Option>
            <Option value="CONFIDENTIAL">{__("Confidential")}</Option>
            <Option value="SECRET">{__("Secret")}</Option>
          </ControlledField>
          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="ownerId"
            label={__("Owner")}
          />
          <VendorsMultiSelectField
            organizationId={organizationId}
            control={control}
            name="vendorIds"
            label={__("Vendors")}
          />
        </DialogContent>
        <DialogFooter>
          <Button disabled={formState.isSubmitting} type="submit">
            {__("Create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
