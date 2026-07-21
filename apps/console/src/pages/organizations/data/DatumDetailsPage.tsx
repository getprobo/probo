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
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  DropdownItem,
  Field,
  IconTrashCan,
  Option,
} from "@probo/ui";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { z } from "zod";

import type { DatumGraphNodeQuery } from "#/__generated__/core/DatumGraphNodeQuery.graphql";
import { ControlledField } from "#/components/form/ControlledField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { ThirdPartiesMultiSelectField } from "#/components/form/ThirdPartiesMultiSelectField";
import {
  datumNodeQuery,
  useDeleteDatum,
  useUpdateDatum,
} from "#/hooks/graph/DatumGraph";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

type Props = {
  queryRef: PreloadedQuery<DatumGraphNodeQuery>;
};

export default function DatumDetailsPage(props: Props) {
  const queryData = usePreloadedQuery<DatumGraphNodeQuery>(
    datumNodeQuery,
    props.queryRef,
  );

  const datumEntry = queryData.node;

  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const updateDatumSchema = z.object({
    name: z.string().min(1, t("datumDetails.validation.nameRequired")),
    dataClassification: z.enum(["PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"]),
    ownerId: z.string().min(1, t("datumDetails.validation.ownerRequired")),
    thirdPartyIds: z.array(z.string()).optional(),
  });

  const deleteDatum = useDeleteDatum(
    datumEntry,
    ConnectionHandler.getConnectionID(organizationId, "DataPage_data"),
  );

  const thirdParties = datumEntry?.thirdParties?.edges.map(edge => edge.node) ?? [];
  const thirdPartyIds = thirdParties.map(thirdParty => thirdParty.id);

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateDatumSchema, {
      defaultValues: {
        name: datumEntry?.name || "",
        dataClassification: datumEntry?.dataClassification || "PUBLIC",
        ownerId: datumEntry?.owner?.id || "",
        thirdPartyIds: thirdPartyIds,
      },
    });

  const updateDatum = useUpdateDatum();

  const onSubmit = handleSubmit(async (formData) => {
    if (!datumEntry?.id) {
      alert(t("datumDetails.errors.missingId"));
      return;
    }
    try {
      await updateDatum({
        id: datumEntry.id,
        ...formData,
      });
      reset(formData);
    } catch (error) {
      console.error("Failed to update datum:", error);
    }
  });

  const breadcrumbItems = [
    {
      label: t("datumDetails.breadcrumbs.data"),
      to: `/organizations/${organizationId}/data`,
    },
    {
      label: datumEntry?.name || "",
    },
  ];

  return (
    <div className="space-y-6">
      <Breadcrumb items={breadcrumbItems} />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="text-2xl">{datumEntry?.name}</div>
          <Badge variant="info">{datumEntry?.dataClassification}</Badge>
        </div>
        {datumEntry.canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={deleteDatum}
            >
              {t("datumDetails.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <form onSubmit={e => void onSubmit(e)} className="space-y-6 max-w-2xl">
        <Field
          label={t("datumDetails.fields.name")}
          {...register("name")}
          type="text"
          disabled={!datumEntry.canUpdate}
        />

        <ControlledField
          control={control}
          name="dataClassification"
          type="select"
          label={t("datumDetails.fields.classification")}
          disabled={!datumEntry.canUpdate}
        >
          <Option value="PUBLIC">{t("datumDetails.classifications.public")}</Option>
          <Option value="INTERNAL">{t("datumDetails.classifications.internal")}</Option>
          <Option value="CONFIDENTIAL">
            {t("datumDetails.classifications.confidential")}
          </Option>
          <Option value="SECRET">{t("datumDetails.classifications.secret")}</Option>
        </ControlledField>

        <PeopleSelectField
          organizationId={organizationId}
          control={control}
          name="ownerId"
          label={t("datumDetails.fields.owner")}
          disabled={!datumEntry.canUpdate}
        />

        <ThirdPartiesMultiSelectField
          organizationId={organizationId}
          control={control}
          name="thirdPartyIds"
          label={t("datumDetails.fields.thirdParties")}
          disabled={!datumEntry.canUpdate}
          selectedThirdParties={thirdParties}
        />

        <div className="flex justify-end">
          {formState.isDirty && datumEntry.canUpdate && (
            <Button type="submit" disabled={formState.isSubmitting}>
              {formState.isSubmitting
                ? t("datumDetails.actions.updating")
                : t("datumDetails.actions.update")}
            </Button>
          )}
        </div>
      </form>
    </div>
  );
}
