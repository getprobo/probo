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
