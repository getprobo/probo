import {
  Badge,
  Button,
  Checkbox,
  Dialog,
  DialogContent,
  DialogFooter,
  Field,
  Spinner,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
} from "@probo/ui";
import { usePreloadedQuery, type PreloadedQuery, useQueryLoader } from "react-relay";
import type { TrustCenterAccessGraphLoadDocumentAccessesQuery } from "/hooks/graph/__generated__/TrustCenterAccessGraphLoadDocumentAccessesQuery.graphql";
import type { TrustCenterDocumentAccess } from "/coredata/TrustCenterDocumentAccess";
import { loadTrustCenterAccessDocumentAccessesQuery, updateTrustCenterAccessMutation } from "/hooks/graph/TrustCenterAccessGraph";
import { useTranslate } from "@probo/i18n";
import z from "zod";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import type { TrustCenterAccess } from "/coredata/TrustCenterAccess";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { Suspense, useEffect } from "react";

function getDocumentAccessInfo(
  docAccess: TrustCenterDocumentAccess,
  __: (key: string) => string
) {
  if (docAccess.document) {
    return {
      variant: "info" as const,
      name: docAccess.document?.title,
      type: __("Document"),
      category: docAccess.document?.documentType,
      id: docAccess.document?.id,
      requested: docAccess.requested,
      active: docAccess.active,
      status: docAccess.status,
    };
  }
  if (docAccess.report) {
    return {
      variant: "success" as const,
      name: docAccess.report?.filename,
      type: __("Report"),
      category: docAccess.report?.audit?.framework?.name,
      id: docAccess.report?.id,
      requested: docAccess.requested,
      active: docAccess.active,
      status: docAccess.status,
    };
  }
  if (docAccess.trustCenterFile) {
    return {
      variant: "highlight" as const,
      name: docAccess.trustCenterFile?.name,
      type: __("File"),
      category: docAccess.trustCenterFile?.category,
      id: docAccess.trustCenterFile?.id,
      requested: docAccess.requested,
      active: docAccess.active,
      status: docAccess.status,
    };
  }

  throw new Error("Unknown trust center access document type");
}

interface TrustCenterAccessEditDialogProps {
  access: TrustCenterAccess;
  onClose: () => void;
}

export function TrustCenterAccessEditDialog(props: TrustCenterAccessEditDialogProps) {
  const { access, onClose } = props;

  const { __ } = useTranslate();

  const [queryRef, loadDocumentAccessesQuery] =
    useQueryLoader<TrustCenterAccessGraphLoadDocumentAccessesQuery>(loadTrustCenterAccessDocumentAccessesQuery);

  useEffect(() => {
    loadDocumentAccessesQuery({
      accessId: access.id
    });
  }, [access.id, loadDocumentAccessesQuery])

  return (
    <Dialog
      defaultOpen={true}
      title={__("Edit Access")}
      onClose={onClose}
    >
      {queryRef &&
        <Suspense>
          <TrustCenterAccessEditForm
            access={access}
            queryRef={queryRef}
            onSubmit={onClose}
          />
        </Suspense>
      }
    </Dialog>
  );
}

interface TrustCenterAccessEditFormProps {
  access: TrustCenterAccess;
  onSubmit: () => void;
  queryRef: PreloadedQuery<TrustCenterAccessGraphLoadDocumentAccessesQuery>;
}

export function TrustCenterAccessEditForm(props: TrustCenterAccessEditFormProps) {
  const { access, onSubmit, queryRef } = props;

  const { __ } = useTranslate();
  const data = usePreloadedQuery<TrustCenterAccessGraphLoadDocumentAccessesQuery>(
    loadTrustCenterAccessDocumentAccessesQuery,
    queryRef,
  )
  const documentAccesses: TrustCenterDocumentAccess[] = data.node.availableDocumentAccesses?.edges.map(edge => edge.node) ?? [];

  const editSchema = z.object({
    name: z.string().min(1, __("Name is required")).min(2, __("Name must be at least 2 characters long")),
    active: z.boolean(),
  });
  const editForm = useFormWithSchema(editSchema, {
    defaultValues: { name: access.name, active: access.active },
  });

  const [updateTrustCenterAccess, isUpdating] = useMutationWithToasts(updateTrustCenterAccessMutation, {
    successMessage: __("Access updated successfully"),
    errorMessage: __("Failed to update access"),
  });

  const handleSubmit = editForm.handleSubmit(async (data) => {
    const { documentIds, reportIds, trustCenterFileIds } = documentAccesses.reduce(
      (acc, docAccess) => {
        // TODO status update
        if (docAccess.document?.id) {
          acc.documentIds.push(docAccess.document.id);
        } else if (docAccess.report?.id) {
          acc.reportIds.push(docAccess.report.id);
        } else if (docAccess.trustCenterFile?.id) {
          acc.trustCenterFileIds.push(docAccess.trustCenterFile.id);
        }
        return acc;
      },
      { documentIds: [] as string[], reportIds: [] as string[], trustCenterFileIds: [] as string[] }
    );

    await updateTrustCenterAccess({
      variables: {
        input: {
          id: access.id,
          name: data.name.trim(),
          active: data.active,
          documentIds,
          reportIds,
          trustCenterFileIds,
        },
      },
      onSuccess: () => {
        onSubmit();
      },
    });
  });

  return (
    <form onSubmit={handleSubmit}>
      <DialogContent padded className="space-y-6">
        <div>
          <p className="text-txt-secondary text-sm mb-4">
            {__("Update access settings and document permissions")}
          </p>

          <Field
            label={__("Full Name")}
            required
            error={editForm.formState.errors.name?.message}
            {...editForm.register("name")}
            placeholder={__("John Doe")}
          />

          <div className="flex items-center justify-between mt-6">
            <div>
              <label className="font-medium text-txt-primary">
                {__("Active Status")}
              </label>
              <p className="text-sm text-txt-secondary">
                {__("Enable or disable access for this user")}
              </p>
            </div>
            <Checkbox
              checked={editForm.watch("active")}
              onChange={(checked) => editForm.setValue("active", checked)}
            />
          </div>
        </div>

        <TrustCenterDocumentAccessList documentAccesses={documentAccesses} />
      </DialogContent>

      <DialogFooter>
        <Button type="submit" disabled={isUpdating}>
          {isUpdating && <Spinner />}
          {__("Update Access")}
        </Button>
      </DialogFooter>
    </form>
  );
}

function TrustCenterDocumentAccessList(props: {
  documentAccesses: TrustCenterDocumentAccess[];
}) {
  const { documentAccesses } = props;

  const { __ } = useTranslate();
  const formattedDocumentAccesses: NonNullable<ReturnType<typeof getDocumentAccessInfo>>[] = documentAccesses
    ?.map((docAccess) => getDocumentAccessInfo(docAccess, __)) ?? [];

  const showGrantCTA = formattedDocumentAccesses.some(da => da.status !== "GRANTED");
  const showRejectCTA = formattedDocumentAccesses.some(da => da.status !== "REJECTED" && da.status !== "REVOKED");

  return (
    <div>
      <div className="flex justify-between items-center mb-4">
        <h4 className="font-medium text-txt-primary">
          {__("Document Access Permissions")}
        </h4>
        {showGrantCTA &&
          <Button
            type="button"
            variant="tertiary"
            // TODO onClick
            className="text-xs h-7 min-h-7"
          >
            {__("Grant All")}
          </Button>
        }
        {showRejectCTA &&
          <Button
            type="button"
            variant="danger"
            // TODO onCLick
            className="text-xs h-7 min-h-7"
          >
            {__("Reject All")}
          </Button>
        }
      </div>

      {formattedDocumentAccesses.length > 0 ? (
        <div className="bg-bg-secondary rounded-lg overflow-hidden">
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Name")}</Th>
                <Th>{__("Type")}</Th>
                <Th>{__("Category")}</Th>
                <Th>
                  {__("Access")}
                </Th>
                <Th></Th>
              </Tr>
            </Thead>
            <Tbody>
              {formattedDocumentAccesses.map((info) => {
                const { variant, name, type, category, id, status } = info;

                return (
                  <Tr key={id}>
                    <Td>
                      <div className="font-medium text-txt-primary">
                        {name}
                      </div>
                    </Td>
                    <Td>
                      <Badge variant={variant}>
                        {type}
                      </Badge>
                    </Td>
                    <Td>
                      <div className="text-txt-secondary">
                        {category || "-"}
                      </div>
                    </Td>
                    <Td>
                      <Badge variant="info">
                        {status}
                      </Badge>
                    </Td>
                    <Td>
                      <div className="flex justify-end">
                        {/* TODO DROPDOWN */}
                      </div>
                    </Td>
                  </Tr>
                );
              })}
            </Tbody>
          </Table>
        </div>
      ) : (
        <div className="text-center text-txt-tertiary py-8">
          {__("No documents available")}
        </div>
      )}
    </div>
  )
}
