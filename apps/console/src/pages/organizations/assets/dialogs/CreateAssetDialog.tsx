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
import { useCreateAsset } from "#/hooks/graph/AssetGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

const schema = z.object({
  name: z.string().min(1, "Name is required"),
  amount: z.number().min(1, "Amount is required"),
  assetType: z.enum(["PHYSICAL", "VIRTUAL"]),
  ownerId: z.string().min(1, "Owner is required"),
  vendorIds: z.array(z.string()).optional(),
  dataTypesStored: z.string().min(1, "Data types stored is required"),
});

type Props = {
  children: React.ReactNode;
  connection: string;
  organizationId: string;
};

export function CreateAssetDialog({
  children,
  connection,
  organizationId,
}: Props) {
  const { __ } = useTranslate();
  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(schema, {
      defaultValues: {
        name: "",
        amount: 0,
        assetType: "VIRTUAL",
        ownerId: "",
        vendorIds: [],
      },
    });
  const ref = useDialogRef();
  const [createAsset] = useCreateAsset(connection);

  const onSubmit = async (data: z.infer<typeof schema>) => {
    try {
      await createAsset({
        ...data,
        organizationId,
      });
      ref.current?.close();
      reset();
    } catch (error) {
      console.error("Failed to create asset:", error);
    }
  };

  return (
    <Dialog
      ref={ref}
      trigger={children}
      title={<Breadcrumb items={[__("Assets"), __("New Asset")]} />}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <DialogContent padded className="space-y-4">
          <Field label={__("Name")} {...register("name")} type="text" />
          <Field
            label={__("Amount")}
            {...register("amount", { valueAsNumber: true })}
            type="number"
          />
          <ControlledField
            control={control}
            name="assetType"
            type="select"
            label={__("Asset Type")}
          >
            <Option value="VIRTUAL">{__("Virtual")}</Option>
            <Option value="PHYSICAL">{__("Physical")}</Option>
          </ControlledField>
          <Field
            label={__("Data Types Stored")}
            {...register("dataTypesStored")}
            type="text"
          />
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
