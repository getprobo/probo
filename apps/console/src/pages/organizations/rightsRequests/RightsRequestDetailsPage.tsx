import {
  formatDatetime,
  formatError,
  getRightsRequestStateLabel,
  getRightsRequestStateOptions,
  getRightsRequestStateVariant,
  getRightsRequestTypeLabel,
  getRightsRequestTypeOptions,
  type GraphQLError,
  toDateInput,
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
  Input,
  Label,
  Option,
  Select,
  Textarea,
  useToast,
} from "@probo/ui";
import { Controller } from "react-hook-form";
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
  requestType: z.enum(["ACCESS", "DELETION", "PORTABILITY"]),
  requestState: z.enum(["TODO", "IN_PROGRESS", "DONE"]),
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
  const { __ } = useTranslate();
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
        title: __("Success"),
        description: __("Rights request updated successfully"),
        variant: "success",
      });
    } catch (error) {
      toast({
        title: __("Error"),
        description: formatError(
          __("Failed to update rights request"),
          error as GraphQLError,
        ),
        variant: "error",
      });
    }
  });

  const typeOptions = getRightsRequestTypeOptions(__);
  const stateOptions = getRightsRequestStateOptions(__);

  const breadcrumbRequestsUrl = `/organizations/${organizationId}/rights-requests`;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Breadcrumb
          items={[
            {
              label: __("Rights Requests"),
              to: breadcrumbRequestsUrl,
            },
            { label: request.dataSubject || request.id! },
          ]}
        />
        {request.canDelete && (
          <ActionDropdown>
            <DropdownItem onClick={deleteRequest} variant="danger">
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        )}
      </div>

      <Card>
        <div className="p-6">
          <div className="mb-6">
            <div className="flex items-center gap-4">
              <h1 className="text-2xl font-bold">
                {getRightsRequestTypeLabel(
                  __,
                  request.requestType || "ACCESS",
                )}
              </h1>
              <Badge variant="neutral">
                {getRightsRequestTypeLabel(
                  __,
                  request.requestType || "ACCESS",
                )}
              </Badge>
              <Badge
                variant={getRightsRequestStateVariant(
                  request.requestState || "TODO",
                )}
              >
                {getRightsRequestStateLabel(
                  __,
                  request.requestState || "TODO",
                )}
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
                      {__("Request Type")}
                      {" "}
                      *
                    </Label>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                    >
                      {typeOptions.map(option => (
                        <Option
                          key={option.value}
                          value={option.value}
                        >
                          {option.label}
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
                      {__("State")}
                      {" "}
                      *
                    </Label>
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                    >
                      {stateOptions.map(option => (
                        <Option
                          key={option.value}
                          value={option.value}
                        >
                          {option.label}
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
              label={__("Data Subject")}
              {...register("dataSubject")}
              error={formState.errors.dataSubject?.message}
            />

            <Field
              label={__("Contact")}
              {...register("contact")}
              error={formState.errors.contact?.message}
            />

            <div>
              <Label>{__("Details")}</Label>
              <Textarea
                {...register("details")}
                placeholder={__("Enter request details")}
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
                <Label>{__("Deadline")}</Label>
                <Input type="date" {...register("deadline")} />
                {formState.errors.deadline?.message && (
                  <div className="text-red-500 text-sm mt-1">
                    {formState.errors.deadline.message}
                  </div>
                )}
              </div>
            </div>

            <div>
              <Label>{__("Action Taken")}</Label>
              <Textarea
                {...register("actionTaken")}
                placeholder={__("Enter action taken")}
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
                    ? __("Saving...")
                    : __("Save Changes")}
                </Button>
              )}
            </div>
          </form>
        </div>
      </Card>
    </div>
  );
}
