// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { useTranslate } from "@probo/i18n";
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

const updateDatumSchema = z.object({
  name: z.string().min(1, "Name is required"),
  dataClassification: z.enum(["PUBLIC", "INTERNAL", "CONFIDENTIAL", "SECRET"]),
  ownerId: z.string().min(1, "Owner is required"),
  thirdPartyIds: z.array(z.string()).optional(),
});

type Props = {
  queryRef: PreloadedQuery<DatumGraphNodeQuery>;
};

export default function DatumDetailsPage(props: Props) {
  const queryData = usePreloadedQuery<DatumGraphNodeQuery>(
    datumNodeQuery,
    props.queryRef,
  );

  const datumEntry = queryData.node;

  const { __ } = useTranslate();
  const organizationId = useOrganizationId();

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
      alert("id is missing from data");
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
      label: __("Data"),
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
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <form onSubmit={e => void onSubmit(e)} className="space-y-6 max-w-2xl">
        <Field
          label={__("Name")}
          {...register("name")}
          type="text"
          disabled={!datumEntry.canUpdate}
        />

        <ControlledField
          control={control}
          name="dataClassification"
          type="select"
          label={__("Classification")}
          disabled={!datumEntry.canUpdate}
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
          disabled={!datumEntry.canUpdate}
        />

        <ThirdPartiesMultiSelectField
          organizationId={organizationId}
          control={control}
          name="thirdPartyIds"
          label={__("Third Parties")}
          disabled={!datumEntry.canUpdate}
          selectedThirdParties={thirdParties}
        />

        <div className="flex justify-end">
          {formState.isDirty && datumEntry.canUpdate && (
            <Button type="submit" disabled={formState.isSubmitting}>
              {formState.isSubmitting ? __("Updating...") : __("Update")}
            </Button>
          )}
        </div>
      </form>
    </div>
  );
}
