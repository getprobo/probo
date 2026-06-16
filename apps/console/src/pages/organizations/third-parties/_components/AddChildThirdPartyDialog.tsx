// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
  Combobox,
  ComboboxItem,
  Dialog,
  DialogContent,
  DialogFooter,
  IconPlusLarge,
  useDialogRef,
} from "@probo/ui";
import { type ReactNode, Suspense, useCallback, useState } from "react";
import { useMutation, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type {
  AddChildThirdPartyDialogCreateMutation,
  CreateThirdPartyInput,
} from "#/__generated__/core/AddChildThirdPartyDialogCreateMutation.graphql";
import type { CommonThirdPartyComboboxQuery } from "#/__generated__/core/CommonThirdPartyComboboxQuery.graphql";

import { commonThirdPartiesQuery, CommonThirdPartyCombobox } from "./CommonThirdPartyCombobox";

const createChildMutation = graphql`
  mutation AddChildThirdPartyDialogCreateMutation(
    $input: CreateThirdPartyInput!
    $connections: [ID!]!
  ) {
    createThirdParty(input: $input) {
      thirdPartyEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          websiteUrl
          category
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  parentThirdPartyId: string;
  parentNamePath: string[];
  organizationId: string;
  connectionId: string;
};

export function AddChildThirdPartyDialog({
  children,
  parentThirdPartyId,
  parentNamePath,
  organizationId,
  connectionId,
}: Props) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const [createChild] = useMutation<AddChildThirdPartyDialogCreateMutation>(createChildMutation);
  const [searchQuery, setSearchQuery] = useState("");
  const [queryRef, loadQuery] = useQueryLoader<CommonThirdPartyComboboxQuery>(commonThirdPartiesQuery);

  const debouncedLoadQuery = useDebounceCallback(
    useCallback(
      (name: string) => {
        loadQuery({ name });
      },
      [loadQuery],
    ),
    500,
  );

  const handleSearch = (name: string) => {
    setSearchQuery(name);
    const trimmed = name.trim();
    if (trimmed.length >= 2) {
      debouncedLoadQuery(trimmed);
    }
  };

  const handleCreate = (input: Omit<CreateThirdPartyInput, "organizationId">) => {
    const qualifiedName = parentNamePath.length > 0
      ? `${input.name} (${parentNamePath.join("/")})`
      : input.name;

    void createChild({
      variables: {
        input: {
          ...input,
          name: qualifiedName,
          organizationId,
          parentThirdPartyId,
        },
        connections: [connectionId],
      },
      onCompleted: () => {
        dialogRef.current?.close();
      },
    });
  };

  const handleCreateNew = (name: string) => {
    handleCreate({ name, category: null });
  };

  return (
    <Dialog ref={dialogRef} trigger={children} title={__("Add a third party")}>
      <DialogContent className="p-6">
        <Combobox onSearch={handleSearch} placeholder={__("Type third party's name")}>
          {searchQuery.trim().length >= 2 && queryRef && (
            <Suspense>
              <CommonThirdPartyCombobox
                queryRef={queryRef}
                excludeNames={new Set()}
                onSelect={handleCreate}
              />
            </Suspense>
          )}
          {searchQuery.trim().length >= 2 && (
            <ComboboxItem onClick={() => handleCreateNew(searchQuery.trim())}>
              <IconPlusLarge size={20} />
              {__("Create a new third party")}
              {" "}
              :
              {searchQuery}
            </ComboboxItem>
          )}
        </Combobox>
      </DialogContent>
      <DialogFooter />
    </Dialog>
  );
}
