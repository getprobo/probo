import {
  Button,
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
  useDialogRef,
  IconChevronDown,
  IconPlusLarge,
} from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { useOutletContext } from "react-router";
import { useState, useEffect, useMemo } from "react";
import z from "zod";
import {
  useTrustCenterAccesses,
  createTrustCenterAccessMutation,
} from "/hooks/graph/TrustCenterAccessGraph";
import { useFormWithSchema } from "/hooks/useFormWithSchema";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { TrustCenterAccessItem } from "./TrustCenterAccessItem";
import type { TrustCenterGraphQuery$data } from "/__generated__/core/TrustCenterGraphQuery.graphql";
import type { NodeOf } from "/types";
import type { TrustCenterAccessGraph_accesses$data } from "/__generated__/core/TrustCenterAccessGraph_accesses.graphql";

export default function TrustCenterAccessTab() {
  const { __ } = useTranslate();
  const { organization } = useOutletContext<TrustCenterGraphQuery$data>();
  const inviteSchema = z.object({
    name: z
      .string()
      .min(1, __("Name is required"))
      .min(2, __("Name must be at least 2 characters long")),
    email: z
      .string()
      .min(1, __("Email is required"))
      .email(__("Please enter a valid email address")),
  });

  const [createInvitation, isCreating] = useMutationWithToasts(
    createTrustCenterAccessMutation,
    {
      successMessage: __("Access created successfully"),
      errorMessage: __("Failed to create access"),
    },
  );

  const dialogRef = useDialogRef();
  const [editingAccess, setEditingAccess] = useState<NodeOf<
    TrustCenterAccessGraph_accesses$data["accesses"]
  > | null>(null);
  const [pendingEditEmail, setPendingEditEmail] = useState<string | null>(null);

  const inviteForm = useFormWithSchema(inviteSchema, {
    defaultValues: { name: "", email: "" },
  });

  const {
    data: trustCenterData,
    loadMore,
    hasNext,
    isLoadingNext,
  } = useTrustCenterAccesses(organization.trustCenter?.id || "");

  const accesses = useMemo(
    () => trustCenterData?.accesses?.edges.map((edge) => edge.node) ?? [],
    [trustCenterData?.accesses?.edges],
  );

  const handleInvite = inviteForm.handleSubmit(async (data) => {
    if (!organization.trustCenter?.id) {
      return;
    }

    const connectionId = trustCenterData?.accesses?.__id;
    const email = data.email.trim();

    await createInvitation({
      variables: {
        input: {
          trustCenterId: organization.trustCenter.id,
          email: email,
          name: data.name.trim(),
          active: false,
        },
        connections: connectionId ? [connectionId] : [],
      },
      onSuccess: () => {
        setPendingEditEmail(email);
      },
    });
  });

  useEffect(() => {
    if (pendingEditEmail && accesses.length > 0) {
      const newAccess = accesses.find(
        (access) => access.email === pendingEditEmail,
      );
      if (newAccess) {
        setPendingEditEmail(null);
        setEditingAccess(newAccess);
        setTimeout(() => {
          dialogRef.current?.close();
        }, 50);
        setTimeout(() => {
          inviteForm.reset();
        }, 300);
      }
    }
  }, [accesses, pendingEditEmail, dialogRef, inviteForm]);

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-base font-medium">{__("External Access")}</h3>
          <p className="text-sm text-txt-tertiary">
            {__(
              "Manage who can access your trust center with time-limited tokens",
            )}
          </p>
        </div>
        {organization.trustCenter?.id &&
          organization.trustCenter.canCreateAccess && (
            <Button
              icon={IconPlusLarge}
              onClick={() => {
                inviteForm.reset();
                dialogRef.current?.open();
              }}
            >
              {__("Add Access")}
            </Button>
          )}
      </div>

      {!organization.trustCenter?.id ? (
        <Table>
          <Tbody>
            <Tr>
              <Td className="text-center text-txt-tertiary py-8">
                <Spinner />
              </Td>
            </Tr>
          </Tbody>
        </Table>
      ) : accesses.length === 0 ? (
        <Table>
          <Tbody>
            <Tr>
              <Td className="text-center text-txt-tertiary py-8">
                {__("No external access granted yet")}
              </Td>
            </Tr>
          </Tbody>
        </Table>
      ) : (
        <>
          <Table>
            <Thead>
              <Tr>
                <Th>{__("Name")}</Th>
                <Th>{__("Email")}</Th>
                <Th>{__("Date")}</Th>
                <Th>{__("Expires")}</Th>
                <Th className="text-center">{__("Active")}</Th>
                <Th className="text-center">{__("Access")}</Th>
                <Th className="text-center">{__("Requests")}</Th>
                <Th className="text-center">{__("NDA")}</Th>
                <Th></Th>
              </Tr>
            </Thead>
            <Tbody>
              {accesses.map((access) => (
                <TrustCenterAccessItem
                  key={`${access.id}-${editingAccess?.id === access.id}`}
                  access={access}
                  connectionId={trustCenterData?.accesses?.__id}
                  dialogOpen={editingAccess?.id === access.id}
                />
              ))}
            </Tbody>
          </Table>
          {hasNext && (
            <Button
              variant="tertiary"
              onClick={loadMore}
              disabled={isLoadingNext}
              className="mt-3 mx-auto"
              icon={IconChevronDown}
            >
              {isLoadingNext && <Spinner />}
              {__("Show More")}
            </Button>
          )}
        </>
      )}

      <Dialog ref={dialogRef} title={__("Invite External Access")}>
        <form onSubmit={handleInvite}>
          <DialogContent padded className="space-y-6">
            <div>
              <p className="text-txt-secondary text-sm mb-4">
                {__(
                  "Send a 30-day access token to an external person to view your trust center",
                )}
              </p>

              <Field
                label={__("Full Name")}
                required
                error={inviteForm.formState.errors.name?.message}
                {...inviteForm.register("name")}
                placeholder={__("John Doe")}
              />

              <div className="mt-4">
                <Field
                  label={__("Email Address")}
                  required
                  error={inviteForm.formState.errors.email?.message}
                  type="email"
                  {...inviteForm.register("email")}
                  placeholder={__("john@example.com")}
                />
              </div>
            </div>
          </DialogContent>

          <DialogFooter>
            <Button type="submit" disabled={isCreating}>
              {isCreating && <Spinner />}
              {__("Create Access")}
            </Button>
          </DialogFooter>
        </form>
      </Dialog>
    </div>
  );
}
