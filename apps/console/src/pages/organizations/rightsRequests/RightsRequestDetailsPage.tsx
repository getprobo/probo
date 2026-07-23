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
  formatDatetime,
  formatError,
  getRightsRequestStateVariant,
  type GraphQLError,
  toDateInput,
} from "@probo/helpers";
import {
  ActionDropdown,
  Badge,
  Breadcrumb,
  Button,
  Card,
  DropdownItem,
  Field,
  Input,
  Label,
  Option,
  Select,
  Textarea,
  useToast,
} from "@probo/ui";
import { Controller } from "react-hook-form";
import { useTranslation } from "react-i18next";
import {
  ConnectionHandler,
  type PreloadedQuery,
  usePreloadedQuery,
} from "react-relay";
import { z } from "zod";

import type { RightsRequestGraphNodeQuery } from "#/__generated__/core/RightsRequestGraphNodeQuery.graphql";
import { useFormWithSchema } from "#/hooks/useFormWithSchema";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  rightsRequestNodeQuery,
  RightsRequestsConnectionKey,
  useDeleteRightsRequest,
  useUpdateRightsRequest,
} from "../../../hooks/graph/RightsRequestGraph";

const updateRequestSchema = z.object({
  requestType: z.enum(["ACCESS", "DELETION", "RECTIFICATION", "PORTABILITY", "OBJECTION", "COMPLAINT"]),
  requestState: z.enum(["TODO", "IN_PROGRESS", "DONE", "REJECTED"]),
  dataSubject: z.string().optional(),
  contact: z.string().optional(),
  details: z.string().optional(),
  deadline: z.string().optional(),
  actionTaken: z.string().optional(),
});

type Props = {
  queryRef: PreloadedQuery<RightsRequestGraphNodeQuery>;
};

export default function RightsRequestDetailsPage(props: Props) {
  const data = usePreloadedQuery<RightsRequestGraphNodeQuery>(
    rightsRequestNodeQuery,
    props.queryRef,
  );
  const request = data.node;
  const { t } = useTranslation();
  const { toast } = useToast();
  const organizationId = useOrganizationId();

  const updateRequest = useUpdateRightsRequest();

  const connectionId = ConnectionHandler.getConnectionID(
    organizationId,
    RightsRequestsConnectionKey,
  );

  const deleteRequest = useDeleteRightsRequest(
    { id: request.id! },
    connectionId,
  );

  const { register, handleSubmit, formState, control } = useFormWithSchema(
    updateRequestSchema,
    {
      defaultValues: {
        requestType: request.requestType || "ACCESS",
        requestState: request.requestState || "TODO",
        dataSubject: request.dataSubject || "",
        contact: request.contact || "",
        details: request.details || "",
        deadline: toDateInput(request.deadline),
        actionTaken: request.actionTaken || "",
      },
    },
  );

  const onSubmit = handleSubmit(async (formData: z.infer<typeof updateRequestSchema>) => {
    try {
      await updateRequest({
        id: request.id!,
        requestType: formData.requestType,
        requestState: formData.requestState,
        dataSubject: formData.dataSubject || undefined,
        contact: formData.contact || undefined,
        details: formData.details || undefined,
        deadline: formatDatetime(formData.deadline) ?? null,
        actionTaken: formData.actionTaken || undefined,
      });

      toast({
        title: t("rightsRequestDetailsPage.messages.success"),
        description: t("rightsRequestDetailsPage.messages.updated"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: t("rightsRequestDetailsPage.messages.error"),
        description: formatError(
          t("rightsRequestDetailsPage.errors.update"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  const typeOptions = ["ACCESS", "DELETION", "RECTIFICATION", "PORTABILITY", "OBJECTION", "COMPLAINT"] as const;
  const stateOptions = ["TODO", "IN_PROGRESS", "DONE", "REJECTED"] as const;

  const breadcrumbRequestsUrl = `/organizations/${organizationId}/rights-requests`;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Breadcrumb
          items={[
            {
              label: t("rightsRequestDetailsPage.breadcrumb.requests"),
              to: breadcrumbRequestsUrl,
            },
            { label: request.dataSubject || request.id! },
          ]}
        />
        {request.canDelete && (
          <ActionDropdown>
            <DropdownItem onClick={deleteRequest} variant="danger">
              {t("rightsRequestDetailsPage.actions.delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <Card>
        <div className="p-6">
          <div className="mb-6">
            <div className="flex items-center gap-4">
              <h1 className="text-2xl font-bold">
                {t(`rightsRequestDetailsPage.types.${(request.requestType || "ACCESS").toLowerCase()}`)}
              </h1>
              <Badge variant="neutral">
                {t(`rightsRequestDetailsPage.types.${(request.requestType || "ACCESS").toLowerCase()}`)}
              </Badge>
              <Badge
                variant={getRightsRequestStateVariant(
                  request.requestState || "TODO",
                )}
              >
                {t(`rightsRequestDetailsPage.states.${(request.requestState || "TODO").toLowerCase()}`)}
              </Badge>
            </div>
          </div>

          <form onSubmit={e => void onSubmit(e)} className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <Controller
                control={control}
                name="requestType"
                render={({ field }) => (
                  <div>
                    <Label>
                      {t("rightsRequestDetailsPage.fields.requestType")}
                      {" "}
                      *
                    </Label>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                    >
                      {typeOptions.map(option => (
                        <Option
                          key={option}
                          value={option}
                        >
                          {t(`rightsRequestDetailsPage.types.${option.toLowerCase()}`)}
                        </Option>
                      ))}
                    </Select>
                    {formState.errors.requestType
                      ?.message && (
                      <div className="text-red-500 text-sm mt-1">
                        {
                          formState.errors.requestType
                            .message
                        }
                      </div>
                    )}
                  </div>
                )}
              />

              <Controller
                control={control}
                name="requestState"
                render={({ field }) => (
                  <div>
                    <Label>
                      {t("rightsRequestDetailsPage.fields.state")}
                      {" "}
                      *
                    </Label>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                    >
                      {stateOptions.map(option => (
                        <Option
                          key={option}
                          value={option}
                        >
                          {t(`rightsRequestDetailsPage.states.${option.toLowerCase()}`)}
                        </Option>
                      ))}
                    </Select>
                    {formState.errors.requestState
                      ?.message && (
                      <div className="text-red-500 text-sm mt-1">
                        {
                          formState.errors
                            .requestState.message
                        }
                      </div>
                    )}
                  </div>
                )}
              />
            </div>

            <Field
              label={t("rightsRequestDetailsPage.fields.dataSubject")}
              {...register("dataSubject")}
              error={formState.errors.dataSubject?.message}
            />

            <Field
              label={t("rightsRequestDetailsPage.fields.contact")}
              {...register("contact")}
              error={formState.errors.contact?.message}
            />

            <div>
              <Label>{t("rightsRequestDetailsPage.fields.details")}</Label>
              <Textarea
                {...register("details")}
                placeholder={t("rightsRequestDetailsPage.placeholders.details")}
                rows={3}
              />
              {formState.errors.details?.message && (
                <div className="text-red-500 text-sm mt-1">
                  {formState.errors.details.message}
                </div>
              )}
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <Label>{t("rightsRequestDetailsPage.fields.deadline")}</Label>
                <Input type="date" {...register("deadline")} />
                {formState.errors.deadline?.message && (
                  <div className="text-red-500 text-sm mt-1">
                    {formState.errors.deadline.message}
                  </div>
                )}
              </div>
            </div>

            <div>
              <Label>{t("rightsRequestDetailsPage.fields.actionTaken")}</Label>
              <Textarea
                {...register("actionTaken")}
                placeholder={t("rightsRequestDetailsPage.placeholders.actionTaken")}
                rows={3}
              />
              {formState.errors.actionTaken?.message && (
                <div className="text-red-500 text-sm mt-1">
                  {formState.errors.actionTaken.message}
                </div>
              )}
            </div>

            <div className="flex justify-end pt-4">
              {request.canUpdate && (
                <Button
                  type="submit"
                  variant="primary"
                  disabled={formState.isSubmitting}
                >
                  {formState.isSubmitting
                    ? t("rightsRequestDetailsPage.actions.saving")
                    : t("rightsRequestDetailsPage.actions.save")}
                </Button>
              )}
            </div>
          </form>
        </div>
      </Card>
    </div>
  );
}
