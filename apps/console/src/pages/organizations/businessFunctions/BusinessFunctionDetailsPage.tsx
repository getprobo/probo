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
  formatError,
  type GraphQLError,
  promisifyMutation,
  sprintf,
} from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
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
import {
  ConnectionHandler,
  graphql,
  type PreloadedQuery,
  useMutation,
  usePreloadedQuery,
} from "react-relay";
import { useNavigate } from "react-router";
import { z } from "zod";

import type { BusinessFunctionDetailsPageDeleteMutation } from "#/__generated__/core/BusinessFunctionDetailsPageDeleteMutation.graphql";
import type { BusinessFunctionDetailsPageQuery } from "#/__generated__/core/BusinessFunctionDetailsPageQuery.graphql";
import type { BusinessFunctionDetailsPageUpdateMutation } from "#/__generated__/core/BusinessFunctionDetailsPageUpdateMutation.graphql";
import { AssetsMultiSelectField } from "#/components/form/AssetsMultiSelectField";
import { PeopleSelectField } from "#/components/form/PeopleSelectField";
import { ThirdPartiesMultiSelectField } from "#/components/form/ThirdPartiesMultiSelectField";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  BusinessFunctionsConnectionKey,
  businessFunctionClassificationOptions,
  businessFunctionListConnectionFilters,
  durationMinutesHelperText,
  getClassificationLabel,
  getClassificationVariant,
} from "./_lib/businessFunctionHelpers";

export const businessFunctionDetailsPageQuery = graphql`
  query BusinessFunctionDetailsPageQuery($businessFunctionId: ID!) {
    node(id: $businessFunctionId) {
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
        }
        thirdParties(first: 100) {
          edges {
            node {
              id
              name
            }
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
        }
        thirdParties(first: 100) {
          edges {
            node {
              id
              name
            }
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

const updateBusinessFunctionSchema = z.object({
  referenceId: z.string().trim().min(1, "Reference ID is required"),
  name: z.string().trim().min(1, "Name is required"),
  classification: z.enum(["CRITICAL", "IMPORTANT", "SECONDARY", "STANDARD"]),
  mtdMinutes: z.coerce.number().int().min(0, "MTD must be zero or greater"),
  rtoMinutes: z.coerce.number().int().min(0, "RTO must be zero or greater"),
  rpoMinutes: z.coerce.number().int().min(0, "RPO must be zero or greater"),
  impactTolerance: z.string().optional(),
  notes: z.string().optional(),
  ownerId: z.string().nullable().optional(),
  assetIds: z.array(z.string()).optional(),
  thirdPartyIds: z.array(z.string()).optional(),
});

type Props = {
  queryRef: PreloadedQuery<BusinessFunctionDetailsPageQuery>;
};

export default function BusinessFunctionDetailsPage(props: Props) {
  const { node: businessFunction } = usePreloadedQuery<BusinessFunctionDetailsPageQuery>(
    businessFunctionDetailsPageQuery,
    props.queryRef,
  );
  const { __ } = useTranslate();
  const { toast } = useToast();
  const organizationId = useOrganizationId();
  const navigate = useNavigate();
  const confirm = useConfirm();

  const [updateBusinessFunction]
    = useMutation<BusinessFunctionDetailsPageUpdateMutation>(updateBusinessFunctionMutation);
  const [deleteBusinessFunction]
    = useMutation<BusinessFunctionDetailsPageDeleteMutation>(deleteBusinessFunctionMutation);

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

  const classificationOptions = businessFunctionClassificationOptions(__);
  const durationHelper = durationMinutesHelperText(__);

  const handleDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteBusinessFunction({
            variables: {
              input: { businessFunctionId: businessFunction.id! },
              connections,
            },
            onCompleted(_, error) {
              if (error) {
                toast({
                  title: __("Error"),
                  description: formatError(
                    __("Failed to delete business function"),
                    error,
                  ),
                  variant: "error",
                });
              } else {
                toast({
                  title: __("Success"),
                  description: __("Business function deleted successfully"),
                  variant: "success",
                });
                void navigate(`/organizations/${organizationId}/business-functions`);
              }
              resolve();
            },
            onError(error) {
              toast({
                title: __("Error"),
                description: formatError(
                  __("Failed to delete business function"),
                  error,
                ),
                variant: "error",
              });
              resolve();
            },
          });
        }),
      {
        message: sprintf(
          __(
            "This will permanently delete the business function %s. This action cannot be undone.",
          ),
          businessFunction.referenceId!,
        ),
      },
    );
  };

  const { control, formState, handleSubmit, register, reset }
    = useFormWithSchema(updateBusinessFunctionSchema, {
      defaultValues: {
        referenceId: businessFunction.referenceId || "",
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

    try {
      await promisifyMutation(updateBusinessFunction)({
        variables: {
          input: {
            id: businessFunction.id,
            referenceId: formData.referenceId,
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
            ...(formState.dirtyFields.assetIds
              ? { assetIds: formData.assetIds }
              : {}),
            ...(formState.dirtyFields.thirdPartyIds
              ? { thirdPartyIds: formData.thirdPartyIds }
              : {}),
          },
        },
      });
      reset(formData);
      toast({
        title: __("Success"),
        description: __("Business function updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(
          __("Failed to update business function"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  const breadcrumbListUrl = `/organizations/${organizationId}/business-functions`;
  const canEdit = businessFunction.canUpdate ?? false;

  return (
    <div className="space-y-6">
      <Breadcrumb
        items={[
          {
            label: __("Business functions"),
            to: breadcrumbListUrl,
          },
          {
            label: businessFunction.referenceId || __("Unknown business function"),
          },
        ]}
      />

      <div className="flex justify-between items-start">
        <div className="flex items-center gap-4">
          <div className="text-2xl font-semibold">
            {businessFunction.referenceId}
          </div>
          <Badge variant={getClassificationVariant(businessFunction.classification || "STANDARD")}>
            {getClassificationLabel(businessFunction.classification || "STANDARD", __)}
          </Badge>
        </div>
        {businessFunction.canDelete && (
          <ActionDropdown variant="secondary">
            <DropdownItem
              variant="danger"
              icon={IconTrashCan}
              onClick={handleDelete}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <div className="max-w-4xl">
        <Card padded>
          <form onSubmit={e => void onSubmit(e)} className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Field
                label={__("Reference ID")}
                required
                error={formState.errors.referenceId?.message}
              >
                <Input
                  {...register("referenceId")}
                  placeholder={__("e.g. F-10")}
                  disabled={!canEdit}
                />
              </Field>

              <Field
                label={__("Name")}
                required
                error={formState.errors.name?.message}
              >
                <Input
                  {...register("name")}
                  placeholder={__("Enter name")}
                  disabled={!canEdit}
                />
              </Field>
            </div>

            <Controller
              control={control}
              name="classification"
              render={({ field }) => (
                <Field label={__("Classification")} required>
                  <Select
                    variant="editor"
                    placeholder={__("Select classification")}
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
                label={__("MTD (minutes)")}
                required
                error={formState.errors.mtdMinutes?.message}
                help={durationHelper}
              >
                <Input
                  type="number"
                  min={0}
                  {...register("mtdMinutes")}
                  disabled={!canEdit}
                />
              </Field>

              <Field
                label={__("RTO (minutes)")}
                required
                error={formState.errors.rtoMinutes?.message}
                help={durationHelper}
              >
                <Input
                  type="number"
                  min={0}
                  {...register("rtoMinutes")}
                  disabled={!canEdit}
                />
              </Field>

              <Field
                label={__("RPO (minutes)")}
                required
                error={formState.errors.rpoMinutes?.message}
                help={durationHelper}
              >
                <Input
                  type="number"
                  min={0}
                  {...register("rpoMinutes")}
                  disabled={!canEdit}
                />
              </Field>
            </div>

            <Field label={__("Impact tolerance")}>
              <Textarea
                {...register("impactTolerance")}
                placeholder={__("Describe acceptable impact during disruption")}
                rows={3}
                disabled={!canEdit}
              />
            </Field>

            <Field label={__("Notes")}>
              <Textarea
                {...register("notes")}
                placeholder={__("Dependencies and other notes")}
                rows={3}
                disabled={!canEdit}
              />
            </Field>

            <PeopleSelectField
              organizationId={organizationId}
              control={control}
              name="ownerId"
              label={__("Owner")}
              error={formState.errors.ownerId?.message}
              optional
              disabled={!canEdit}
            />

            <AssetsMultiSelectField
              organizationId={organizationId}
              control={control}
              name="assetIds"
              label={__("Assets")}
              selectedAssets={assets}
              disabled={!canEdit}
            />

            <ThirdPartiesMultiSelectField
              organizationId={organizationId}
              control={control}
              name="thirdPartyIds"
              selectedThirdParties={thirdParties}
              label={__("Third parties")}
              disabled={!canEdit}
            />

            <div className="flex justify-end">
              {formState.isDirty && canEdit && (
                <Button type="submit" disabled={formState.isSubmitting}>
                  {formState.isSubmitting ? __("Updating...") : __("Update")}
                </Button>
              )}
            </div>
          </form>
        </Card>
      </div>
    </div>
  );
}
