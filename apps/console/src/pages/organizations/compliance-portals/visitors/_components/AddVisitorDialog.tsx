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

import type { AddVisitorComboboxQuery } from "#/__generated__/core/AddVisitorComboboxQuery.graphql";
import type { AddVisitorDialogCreateMutation } from "#/__generated__/core/AddVisitorDialogCreateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { useAccessListFilters } from "../_lib/useAccessListFilters";
import { addVisitorDialog } from "../variants";

import {
  type AddVisitorCandidate,
  AddVisitorCombobox,
  addVisitorComboboxQuery,
} from "./AddVisitorCombobox";

const createAccessMutation = graphql`
  mutation AddVisitorDialogCreateMutation(
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

export interface AddVisitorDialogProps {
  children: ReactElement;
  compliancePortalId: string;
}

export function AddVisitorDialog({
  children,
  compliancePortalId,
}: AddVisitorDialogProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const navigate = useNavigate();
  const { order, query, sort } = useAccessListFilters();
  const [open, setOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [queryRef, loadQuery]
    = useQueryLoader<AddVisitorComboboxQuery>(addVisitorComboboxQuery);
  const [createAccess, isCreating] = useMutation<AddVisitorDialogCreateMutation>(
    createAccessMutation,
    {
      successMessage: t("addVisitorDialog.messages.created"),
      errorToast: t("addVisitorDialog.errors.create"),
    },
  );
  const { body, item, hit, addEmail } = addVisitorDialog();
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

  async function addVisitor(input: { profileId?: string; email?: string }) {
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

  function handleSelectCandidate(candidate: AddVisitorCandidate) {
    void addVisitor({ profileId: candidate.id });
  }

  function handleAddEmail(email: string) {
    void addVisitor({ email });
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
  const showAddEmail = isLikelyEmail(trimmedQuery);

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={children} />
      <DialogPopup>
        <DialogHeader>
          <DialogTitle>{t("addVisitorDialog.title")}</DialogTitle>
        </DialogHeader>
        <DialogBody className={body()}>
          <TextField
            icon={<MagnifyingGlassIcon />}
            value={searchQuery}
            onValueChange={handleSearch}
            placeholder={t("addVisitorDialog.searchPlaceholder")}
            aria-label={t("addVisitorDialog.searchPlaceholder")}
          />
          {canSearch && queryRef != null && queryRef.variables.query === trimmedQuery && (
            <Suspense fallback={<ListSkeleton count={3} />}>
              <AddVisitorCombobox
                queryRef={queryRef}
                onSelect={handleSelectCandidate}
              />
            </Suspense>
          )}
          {showAddEmail && (
            <List>
              <ListItem className={item()}>
                <button
                  type="button"
                  className={hit()}
                  aria-label={t("addVisitorDialog.addEmail", { email: trimmedQuery })}
                  onClick={() => {
                    handleAddEmail(trimmedQuery);
                  }}
                />
                <div className={addEmail()}>
                  <PlusIcon aria-hidden />
                  <Text size={2} weight="medium" color="neutral" highContrast>
                    {t("addVisitorDialog.addEmail", { email: trimmedQuery })}
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
