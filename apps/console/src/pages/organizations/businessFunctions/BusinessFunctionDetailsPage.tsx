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
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  Card,
  DropdownItem,
  Field,
  IconTrashCan,
  Input,
  Option,
  Select,
  Textarea,
  useConfirm,
  useToast,
} from "@probo/ui";
import { Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";

import type { BusinessFunctionDetailsPageDeleteMutation } from "#/__generated__/core/BusinessFunctionDetailsPageDeleteMutation.graphql";
import type { BusinessFunctionDetailsPageQuery } from "#/__generated__/core/BusinessFunctionDetailsPageQuery.graphql";
import type { BusinessFunctionDetailsPageUpdateMutation } from "#/__generated__/core/BusinessFunctionDetailsPageUpdateMutation.graphql";
import { AssetsMultiSelectField } from "#/components/form/AssetsMultiSelectField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { ThirdPartiesMultiSelectField } from "#/components/form/ThirdPartiesMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";
import { useMutation } from "#/lib/relay/useMutation";
import { z } from "#/lib/zod";

import {
  businessFunctionClassificationOptions,
  businessFunctionListConnectionFilters,
  BusinessFunctionsConnectionKey,
  durationMinutesHelperText,
  getClassificationLabel,
  getClassificationVariant,
} from "./_lib/businessFunctionHelpers";

export const businessFunctionDetailsPageQuery = graphql`
  query BusinessFunctionDetailsPageQuery($businessFunctionId: ID!) {
    node(id: $businessFunctionId) @required(action: THROW) {
      __typename
      ... on BusinessFunction {
        id
        referenceId
        name
        classification
        mtdMinutes
        rtoMinutes
        rpoMinutes
        impactTolerance
        notes
        owner {
          id
        }
        assets(first: 100) {
          edges {
            node {
              id
              name
            }
          }
          pageInfo {
            hasNextPage
          }
        }
        thirdParties(first: 100) {
          edges {
            node {
              id
              name
            }
          }
          pageInfo {
            hasNextPage
          }
        }
        canUpdate: permission(action: "core:business-function:update")
        canDelete: permission(action: "core:business-function:delete")
      }
    }
  }
`;

const updateBusinessFunctionMutation = graphql`
  mutation BusinessFunctionDetailsPageUpdateMutation(
    $input: UpdateBusinessFunctionInput!
  ) {
    updateBusinessFunction(input: $input) {
      businessFunction {
        id
        referenceId
        name
        classification
        mtdMinutes
        rtoMinutes
        rpoMinutes
        impactTolerance
        notes
        owner {
          id
          fullName
        }
        assets(first: 100) {
          edges {
            node {
              id
              name
            }
          }
          pageInfo {
            hasNextPage
          }
        }
        thirdParties(first: 100) {
          edges {
            node {
              id
              name
            }
          }
          pageInfo {
            hasNextPage
          }
        }
        updatedAt
      }
    }
  }
`;

const deleteBusinessFunctionMutation = graphql`
  mutation BusinessFunctionDetailsPageDeleteMutation(
    $input: DeleteBusinessFunctionInput!
    $connections: [ID!]!
  ) {
    deleteBusinessFunction(input: $input) {
      deletedBusinessFunctionId @deleteEdge(connections: $connections)
    }
  }
`;

interface BusinessFunctionDetailsPageProps {
  queryRef: PreloadedQuery<BusinessFunctionDetailsPageQuery>;
}

export function BusinessFunctionDetailsPage({
  queryRef,
}: BusinessFunctionDetailsPageProps) {
  const { node: businessFunction } = usePreloadedQuery<BusinessFunctionDetailsPageQuery>(
    businessFunctionDetailsPageQuery,
    queryRef,
  );
  if (businessFunction.__typename !== "BusinessFunction") {
    throw new NotFoundError("Business function not found");
  }
  const { t } = useTranslation();

  const updateBusinessFunctionSchema = z.object({
    name: z.string().trim().min(
      1,
      t("businessFunctionDetailsPage.validation.nameRequired"),
    ),
    classification: z.enum(["CRITICAL", "IMPORTANT", "SECONDARY", "STANDARD"]),
    mtdMinutes: z.coerce.number().int().min(0),
    rtoMinutes: z.coerce.number().int().min(0),
    rpoMinutes: z.coerce.number().int().min(0),
    impactTolerance: z.string().optional(),
    notes: z.string().optional(),
    ownerId: z.string().nullable().optional(),
    assetIds: z.array(z.string()).optional(),
    thirdPartyIds: z.array(z.string()).optional(),
  });
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const confirm = useConfirm();
  const { toast } = useToast();

  const [updateBusinessFunction]
    = useMutation<BusinessFunctionDetailsPageUpdateMutation>(
      updateBusinessFunctionMutation,
      {
        errorToast: t("businessFunctionDetailsPage.errors.update"),
      },
    );
  const [deleteBusinessFunction]
    = useMutation<BusinessFunctionDetailsPageDeleteMutation>(
      deleteBusinessFunctionMutation,
      {
        errorToast: t("businessFunctionDetailsPage.errors.delete"),
      },
    );

  const connections = businessFunctionListConnectionFilters(businessFunction).map(filter =>
    ConnectionHandler.getConnectionID(
      organizationId,
      BusinessFunctionsConnectionKey,
      { filter },
    ),
  );

  const assets = businessFunction.assets?.edges.map(edge => edge.node) ?? [];
  const assetIds = assets.map(asset => asset.id);
  const thirdParties = businessFunction.thirdParties?.edges.map(edge => edge.node) ?? [];
  const thirdPartyIds = thirdParties.map(thirdParty => thirdParty.id);
  const canEdit = businessFunction.canUpdate ?? false;
  const dependenciesTruncated
    = Boolean(businessFunction.assets?.pageInfo.hasNextPage)
      || Boolean(businessFunction.thirdParties?.pageInfo.hasNextPage);
  const canEditDependencies = canEdit && !dependenciesTruncated;

  const classificationOptions = businessFunctionClassificationOptions(
    t,
    "businessFunctionDetailsPage",
  );
  const durationHelper = durationMinutesHelperText(t, "businessFunctionDetailsPage");

  const handleDelete = () => {
    confirm(
      async () => {
        await deleteBusinessFunction({
          variables: {
            input: { businessFunctionId: businessFunction.id },
            connections,
          },
        });
        toast({
          title: t("businessFunctionDetailsPage.messages.success"),
          description: t("businessFunctionDetailsPage.messages.deleted"),
          variant: "success",
        });
        void navigate(`/organizations/${organizationId}/settings/business-functions`);
      },
      {
        message: t("businessFunctionDetailsPage.deleteConfirmation", {
          referenceId: businessFunction.referenceId,
        }),
      },
    );
  };

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateBusinessFunctionSchema, {
      defaultValues: {
        name: businessFunction.name || "",
        classification: businessFunction.classification || "STANDARD",
        mtdMinutes: businessFunction.mtdMinutes ?? 0,
        rtoMinutes: businessFunction.rtoMinutes ?? 0,
        rpoMinutes: businessFunction.rpoMinutes ?? 0,
        impactTolerance: businessFunction.impactTolerance || "",
        notes: businessFunction.notes || "",
        ownerId: businessFunction.owner?.id ?? null,
        assetIds,
        thirdPartyIds,
      },
    });

  const onSubmit = handleSubmit(async (formData) => {
    if (!businessFunction.id) return;

    await updateBusinessFunction({
      variables: {
        input: {
          id: businessFunction.id,
          name: formData.name,
          classification: formData.classification,
          mtdMinutes: formData.mtdMinutes,
          rtoMinutes: formData.rtoMinutes,
          rpoMinutes: formData.rpoMinutes,
          impactTolerance: formData.impactTolerance || null,
          notes: formData.notes || null,
          ownerId: formData.ownerId || null,
          // Only send dependency IDs when edited so a capped (first: 100) load
          // cannot wipe relationships beyond the form's initial selection.
          ...(canEditDependencies && formState.dirtyFields.assetIds
            ? { assetIds: formData.assetIds }
            : {}),
          ...(canEditDependencies && formState.dirtyFields.thirdPartyIds
            ? { thirdPartyIds: formData.thirdPartyIds }
            : {}),
        },
      },
    });
    reset(formData);
    toast({
      title: t("businessFunctionDetailsPage.messages.success"),
      description: t("businessFunctionDetailsPage.messages.updated"),
      variant: "success",
    });
  });

  const breadcrumbListUrl = `/organizations/${organizationId}/settings/business-functions`;

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: t("businessFunctionDetailsPage.breadcrumb.businessFunctions"),
            to: breadcrumbListUrl,
          },
          {
            label: businessFunction.referenceId
              || t("businessFunctionDetailsPage.breadcrumb.unknown"),
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="text-2xl font-semibold">
            {businessFunction.referenceId}
          </div>
          <Badge variant={getClassificationVariant(businessFunction.classification || "STANDARD")}>
            {getClassificationLabel(
              businessFunction.classification || "STANDARD",
              t,
              "businessFunctionDetailsPage",
            )}
          </Badge>
        </div>
        {businessFunction.canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={handleDelete}
            >
              {t("businessFunctionDetailsPage.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <div className="max-w-4xl">
        <Card padded>
          <form onSubmit={e => void onSubmit(e)} className="space-y-6">
            <Field
              label={t("businessFunctionDetailsPage.fields.name")}
              required
              error={formState.errors.name?.message}
            >
              <Input
                {...register("name")}
                placeholder={t("businessFunctionDetailsPage.fields.namePlaceholder")}
                disabled={!canEdit}
              />
            </Field>

            <Controller
              control={control}
              name="classification"
              render={({ field }) => (
                <Field label={t("businessFunctionDetailsPage.fields.classification")} required>
                  <Select
                    variant="editor"
                    placeholder={t("businessFunctionDetailsPage.fields.classificationPlaceholder")}
                    onValueChange={field.onChange}
                    value={field.value}
                    className="w-full"
                    disabled={!canEdit}
                  >
                    {classificationOptions.map(option => (
                      <Option key={option.value} value={option.value}>
                        {option.label}
                      </Option>
                    ))}
                  </Select>
                  {formState.errors.classification && (
                    <p className="text-sm text-red-500 mt-1">
                      {formState.errors.classification.message}
                    </p>
                  )}
                </Field>
              )}
            />

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <Field
                label={t("businessFunctionDetailsPage.fields.mtd")}
                required
                error={formState.errors.mtdMinutes?.message}
              >
                <Input
                  type="number"
                  min={0}
                  title={durationHelper}
                  {...register("mtdMinutes")}
                  disabled={!canEdit}
                />
              </Field>

              <Field
                label={t("businessFunctionDetailsPage.fields.rto")}
                required
                error={formState.errors.rtoMinutes?.message}
              >
                <Input
                  type="number"
                  min={0}
                  title={durationHelper}
                  {...register("rtoMinutes")}
                  disabled={!canEdit}
                />
              </Field>

              <Field
                label={t("businessFunctionDetailsPage.fields.rpo")}
                required
                error={formState.errors.rpoMinutes?.message}
              >
                <Input
                  type="number"
                  min={0}
                  title={durationHelper}
                  {...register("rpoMinutes")}
                  disabled={!canEdit}
                />
              </Field>
            </div>

            <Field label={t("businessFunctionDetailsPage.fields.impactTolerance")}>
              <Textarea
                {...register("impactTolerance")}
                placeholder={t("businessFunctionDetailsPage.fields.impactTolerancePlaceholder")}
                rows={3}
                disabled={!canEdit}
              />
            </Field>

            <Field label={t("businessFunctionDetailsPage.fields.notes")}>
              <Textarea
                {...register("notes")}
                placeholder={t("businessFunctionDetailsPage.fields.notesPlaceholder")}
                rows={3}
                disabled={!canEdit}
              />
            </Field>

            <PeopleSelectField
              organizationId={organizationId}
              control={control}
              name="ownerId"
              label={t("businessFunctionDetailsPage.fields.owner")}
              error={formState.errors.ownerId?.message}
              optional
              disabled={!canEdit}
            />

            {dependenciesTruncated && (
              <p className="text-sm text-txt-tertiary">
                {t("businessFunctionDetailsPage.dependenciesTruncated")}
              </p>
            )}

            <AssetsMultiSelectField
              organizationId={organizationId}
              control={control}
              name="assetIds"
              label={t("businessFunctionDetailsPage.fields.assets")}
              selectedAssets={assets}
              disabled={!canEditDependencies}
            />

            <ThirdPartiesMultiSelectField
              organizationId={organizationId}
              control={control}
              name="thirdPartyIds"
              selectedThirdParties={thirdParties}
              label={t("businessFunctionDetailsPage.fields.thirdParties")}
              disabled={!canEditDependencies}
            />

            <div className="flex justify-end">
              {formState.isDirty && canEdit && (
                <Button type="submit" disabled={formState.isSubmitting}>
                  {formState.isSubmitting
                    ? t("businessFunctionDetailsPage.actions.updating")
                    : t("businessFunctionDetailsPage.actions.update")}
                </Button>
              )}
            </div>
          </form>
        </Card>
      </div>
    </div>
  );
}
