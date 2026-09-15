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

import { PlusIcon, UserIcon } from "@phosphor-icons/react";
import { TextField } from "@probo/ui/src/v2/form/TextField";
import { Popover } from "@probo/ui/src/v2/Popover/Popover";
import { PopoverPopup } from "@probo/ui/src/v2/Popover/PopoverPopup";
import { PopoverTrigger } from "@probo/ui/src/v2/Popover/PopoverTrigger";
import { type ReactElement, useCallback, useState, useTransition } from "react";
import { useTranslation } from "react-i18next";
import { graphql, useQueryLoader } from "react-relay";
import { useNavigate } from "react-router";
import { ConnectionHandler } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type { AddVisitorComboboxQuery } from "#/__generated__/core/AddVisitorComboboxQuery.graphql";
import type { AddVisitorPopoverCreateMutation } from "#/__generated__/core/AddVisitorPopoverCreateMutation.graphql";
import { useMutation } from "#/lib/relay/useMutation";

import { useAccessListFilters } from "../_lib/useAccessListFilters";
import { addVisitorPopover } from "../variants";

import {
  type AddVisitorCandidate,
  AddVisitorCombobox,
  addVisitorComboboxQuery,
} from "./AddVisitorCombobox";

const createAccessMutation = graphql`
  mutation AddVisitorPopoverCreateMutation(
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

export interface AddVisitorPopoverProps {
  children: ReactElement;
  compliancePortalId: string;
}

export function AddVisitorPopover({
  children,
  compliancePortalId,
}: AddVisitorPopoverProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const navigate = useNavigate();
  const { order, query, sort } = useAccessListFilters();
  const [open, setOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [isPending, startTransition] = useTransition();
  const [queryRef, loadQuery, disposeQuery]
    = useQueryLoader<AddVisitorComboboxQuery>(addVisitorComboboxQuery);
  const [createAccess, isCreating] = useMutation<AddVisitorPopoverCreateMutation>(
    createAccessMutation,
    {
      successMessage: t("addVisitorDialog.messages.created"),
      errorToast: t("addVisitorDialog.errors.create"),
    },
  );
  const { body, results, item, name, icon } = addVisitorPopover({ pending: isPending });
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
        startTransition(() => {
          loadQuery(
            { compliancePortalId, query },
            { fetchPolicy: "store-or-network" },
          );
        });
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
      disposeQuery();
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
      disposeQuery();
    }
  }

  const trimmedQuery = searchQuery.trim();
  const canSearch = trimmedQuery.length >= 2;
  const showAddEmail = isLikelyEmail(trimmedQuery);

  return (
    <Popover open={open} onOpenChange={handleOpenChange} modal>
      <PopoverTrigger render={children} />
      <PopoverPopup
        side="bottom"
        align="end"
        className="w-80 p-2"
        aria-label={t("addVisitorDialog.title")}
      >
        <div className={body()}>
          <TextField
            size={1}
            icon={<UserIcon />}
            value={searchQuery}
            onValueChange={handleSearch}
            placeholder={t("addVisitorDialog.searchPlaceholder")}
            aria-label={t("addVisitorDialog.searchPlaceholder")}
          />
          {canSearch && queryRef != null && (
            <div className={results()} aria-busy={isPending}>
              <AddVisitorCombobox
                queryRef={queryRef}
                onSelect={handleSelectCandidate}
                showEmpty={!showAddEmail}
              />
            </div>
          )}
          {showAddEmail && (
            <button
              type="button"
              className={item()}
              aria-label={t("addVisitorDialog.addEmail", { email: trimmedQuery })}
              onClick={() => {
                handleAddEmail(trimmedQuery);
              }}
            >
              <PlusIcon className={icon()} aria-hidden />
              <span className={name()}>
                {t("addVisitorDialog.addEmail", { email: trimmedQuery })}
              </span>
            </button>
          )}
        </div>
      </PopoverPopup>
    </Popover>
  );
}
