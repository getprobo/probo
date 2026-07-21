// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  Option,
  useDialogRef,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import { z } from "zod";

import { ControlledField } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { ThirdPartiesMultiSelectField } from "#/components/form/ThirdPartiesMultiSelectField";
import { useCreateAsset } from "#/hooks/graph/AssetGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

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
  const { t } = useTranslation();
  const schema = z.object({
    name: z.string().min(1, t("createAssetDialog.validation.nameRequired")),
    amount: z.number().min(1, t("createAssetDialog.validation.amountRequired")),
    assetType: z.enum(["PHYSICAL", "VIRTUAL"]),
    ownerId: z.string().min(1, t("createAssetDialog.validation.ownerRequired")),
    thirdPartyIds: z.array(z.string()).optional(),
    dataTypesStored: z.string().min(
      1,
      t("createAssetDialog.validation.dataTypesStoredRequired"),
    ),
  });
  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(schema, {
      defaultValues: {
        name: "",
        amount: 0,
        assetType: "VIRTUAL",
        ownerId: "",
        thirdPartyIds: [],
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
      title={(
        <Breadcrumb items={[
          t("createAssetDialog.breadcrumb.assets"),
          t("createAssetDialog.breadcrumb.newAsset"),
        ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <DialogContent padded className="space-y-4">
          <Field
            label={t("createAssetDialog.fields.name")}
            {...register("name")}
            type="text"
          />
          <Field
            label={t("createAssetDialog.fields.amount")}
            {...register("amount", { valueAsNumber: true })}
            type="number"
          />
          <ControlledField
            control={control}
            name="assetType"
            type="select"
            label={t("createAssetDialog.fields.assetType")}
          >
            <Option value="VIRTUAL">
              {t("createAssetDialog.assetTypes.virtual")}
            </Option>
            <Option value="PHYSICAL">
              {t("createAssetDialog.assetTypes.physical")}
            </Option>
          </ControlledField>
          <Field
            label={t("createAssetDialog.fields.dataTypesStored")}
            {...register("dataTypesStored")}
            type="text"
          />
          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="ownerId"
            label={t("createAssetDialog.fields.owner")}
          />
          <ThirdPartiesMultiSelectField
            organizationId={organizationId}
            control={control}
            name="thirdPartyIds"
            label={t("createAssetDialog.fields.thirdParties")}
          />
        </DialogContent>
        <DialogFooter>
          <Button disabled={formState.isSubmitting} type="submit">
            {t("createAssetDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
