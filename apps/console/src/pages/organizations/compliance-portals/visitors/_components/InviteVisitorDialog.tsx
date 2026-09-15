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

import { MagnifyingGlassIcon, PlusIcon } from "@phosphor-icons/react";
import { Dialog } from "@probo/ui/src/v2/Dialog/Dialog";
import { DialogBody } from "@probo/ui/src/v2/Dialog/DialogBody";
import { DialogHeader } from "@probo/ui/src/v2/Dialog/DialogHeader";
import { DialogPopup } from "@probo/ui/src/v2/Dialog/DialogPopup";
import { DialogTitle } from "@probo/ui/src/v2/Dialog/DialogTitle";
import { DialogTrigger } from "@probo/ui/src/v2/Dialog/DialogTrigger";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { List } from "@probo/ui/src/v2/List/List";
import { ListItem } from "@probo/ui/src/v2/List/ListItem";
import { ListSkeleton } from "@probo/ui/src/v2/List/ListSkeleton";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { type ReactElement, Suspense, useCallback, useState } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useQueryLoader } from "react-relay";
import { useNavigate } from "react-router";
import { ConnectionHandler } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type { InviteVisitorComboboxQuery } from "#/__generated__/core/InviteVisitorComboboxQuery.graphql";
import type { InviteVisitorDialogCreateMutation } from "#/__generated__/core/InviteVisitorDialogCreateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { useAccessListFilters } from "../_lib/useAccessListFilters";
import { inviteVisitorDialog } from "../variants";

import {
  type InviteVisitorCandidate,
  InviteVisitorCombobox,
  inviteVisitorComboboxQuery,
} from "./InviteVisitorCombobox";

const createAccessMutation = graphql`
  mutation InviteVisitorDialogCreateMutation(
    $input: CreateCompliancePortalAccessInput!
    $connections: [ID!]!
  ) {
    createCompliancePortalAccess(input: $input) {
      compliancePortalAccessEdge @prependEdge(connections: $connections) {
        cursor
        node {
          id
          ...CompliancePortalAccessListItemFragment
        }
      }
    }
  }
`;

function isLikelyEmail(value: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}

export interface InviteVisitorDialogProps {
  children: ReactElement;
  compliancePortalId: string;
}

export function InviteVisitorDialog({
  children,
  compliancePortalId,
}: InviteVisitorDialogProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const navigate = useNavigate();
  const { order, query, sort } = useAccessListFilters();
  const [open, setOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [queryRef, loadQuery]
    = useQueryLoader<InviteVisitorComboboxQuery>(inviteVisitorComboboxQuery);
  const [createAccess, isCreating] = useMutation<InviteVisitorDialogCreateMutation>(
    createAccessMutation,
    {
      successMessage: t("inviteVisitorDialog.messages.created"),
      errorToast: t("inviteVisitorDialog.errors.create"),
    },
  );
  const { body, item, hit, invite } = inviteVisitorDialog();
  const connectionId = ConnectionHandler.getConnectionID(
    compliancePortalId,
    "CompliancePortalAccessList_accesses",
    {
      orderBy: order,
      filter: { query },
    },
  );

  const debouncedLoadQuery = useDebounceCallback(
    useCallback(
      (query: string) => {
        loadQuery({ compliancePortalId, query });
      },
      [compliancePortalId, loadQuery],
    ),
    500,
  );

  async function inviteVisitor(input: { profileId?: string; email?: string }) {
    if (isCreating) {
      return;
    }

    try {
      const response = await createAccess({
        variables: {
          input: {
            compliancePortalId,
            ...input,
          },
          connections: query === "" && sort === "joined" ? [connectionId] : [],
        },
      });
      const accessId
        = response.createCompliancePortalAccess.compliancePortalAccessEdge.node.id;
      setOpen(false);
      setSearchQuery("");
      void navigate(accessId);
    } catch {
      // Error toast is already shown by useMutation.
    }
  }

  function handleSelectCandidate(candidate: InviteVisitorCandidate) {
    void inviteVisitor({ profileId: candidate.id });
  }

  function handleInviteEmail(email: string) {
    void inviteVisitor({ email });
  }

  function handleSearch(query: string) {
    setSearchQuery(query);
    const trimmed = query.trim();
    if (trimmed.length >= 2) {
      debouncedLoadQuery(trimmed);
    }
  }

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      setSearchQuery("");
    }
  }

  const trimmedQuery = searchQuery.trim();
  const canSearch = trimmedQuery.length >= 2;
  const showInviteEmail = isLikelyEmail(trimmedQuery);

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={children} />
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("inviteVisitorDialog.title")}</DialogTitle>
        </DialogHeader>
        <DialogBody className={body()}>
          <TextField
            icon={<MagnifyingGlassIcon />}
            value={searchQuery}
            onValueChange={handleSearch}
            placeholder={t("inviteVisitorDialog.searchPlaceholder")}
            aria-label={t("inviteVisitorDialog.searchPlaceholder")}
          />
          {canSearch && queryRef != null && queryRef.variables.query === trimmedQuery && (
            <Suspense fallback={<ListSkeleton count={3} />}>
              <InviteVisitorCombobox
                queryRef={queryRef}
                onSelect={handleSelectCandidate}
              />
            </Suspense>
          )}
          {showInviteEmail && (
            <List>
              <ListItem className={item()}>
                <button
                  type="button"
                  className={hit()}
                  aria-label={t("inviteVisitorDialog.inviteEmail", { email: trimmedQuery })}
                  onClick={() => {
                    handleInviteEmail(trimmedQuery);
                  }}
                />
                <div className={invite()}>
                  <PlusIcon aria-hidden />
                  <Text size={2} weight="medium" color="neutral" highContrast>
                    {t("inviteVisitorDialog.inviteEmail", { email: trimmedQuery })}
                  </Text>
                </div>
              </ListItem>
            </List>
          )}
        </DialogBody>
      </DialogPopup>
    </Dialog>
  );
}
