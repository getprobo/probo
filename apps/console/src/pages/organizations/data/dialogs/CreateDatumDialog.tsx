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
import { useFormWithSchema } from "#/hooks/useFormWithSchema";

import { useCreateDatum } from "../../../../hooks/graph/DatumGraph";

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
  const { t } = useTranslation();
  const schema = z.object({
    name: z.string().min(1, t("createDatumDialog.validation.nameRequired")),
    dataClassification: z.enum(["PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"]),
    ownerId: z.string().min(1, t("createDatumDialog.validation.ownerRequired")),
    thirdPartyIds: z.array(z.string()).optional(),
  });
  const { control, handleSubmit, register, formState, reset }
    = useFormWithSchema(schema, {
      defaultValues: {
        name: "",
        dataClassification: "PUBLIC",
        ownerId: "",
        thirdPartyIds: [],
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
      title={(
        <Breadcrumb
          items={[
            t("createDatumDialog.breadcrumbs.data"),
            t("createDatumDialog.breadcrumbs.new"),
          ]}
        />
      )}
    >
      <form onSubmit={e => void handleSubmit(onSubmit)(e)} className="space-y-4">
        <DialogContent padded className="space-y-4">
          <Field
            label={t("createDatumDialog.fields.name")}
            {...register("name")}
            type="text"
          />
          <ControlledField
            control={control}
            name="dataClassification"
            type="select"
            label={t("createDatumDialog.fields.classification")}
          >
            <Option value="PUBLIC">
              {t("createDatumDialog.classifications.public")}
            </Option>
            <Option value="INTERNAL">
              {t("createDatumDialog.classifications.internal")}
            </Option>
            <Option value="CONFIDENTIAL">
              {t("createDatumDialog.classifications.confidential")}
            </Option>
            <Option value="SECRET">
              {t("createDatumDialog.classifications.secret")}
            </Option>
          </ControlledField>
          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="ownerId"
            label={t("createDatumDialog.fields.owner")}
          />
          <ThirdPartiesMultiSelectField
            organizationId={organizationId}
            control={control}
            name="thirdPartyIds"
            label={t("createDatumDialog.fields.thirdParties")}
          />
        </DialogContent>
        <DialogFooter>
          <Button disabled={formState.isSubmitting} type="submit">
            {t("createDatumDialog.actions.create")}
          </Button>
        </DialogFooter>
      </form>
    </Dialog>
  );
}
