// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { faviconUrl } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import {
  Avatar,
  Combobox,
  ComboboxItem,
  Dialog,
  DialogContent,
  DialogFooter,
  useDialogRef,
} from "@probo/ui";
import { type ReactNode, Suspense, useCallback, useState } from "react";
import { useMutation, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";
import { useDebounceCallback } from "usehooks-ts";

import type { AddChildThirdPartyDialogCreateMappingMutation } from "#/__generated__/core/AddChildThirdPartyDialogCreateMappingMutation.graphql";
import type { AddChildThirdPartyDialogCreateMutation } from "#/__generated__/core/AddChildThirdPartyDialogCreateMutation.graphql";
import type { CommonThirdPartyComboboxQuery } from "#/__generated__/core/CommonThirdPartyComboboxQuery.graphql";
import type { CreateThirdPartyInput } from "#/__generated__/core/ThirdPartyGraphCreateMutation.graphql";
import { useThirdParties } from "#/hooks/graph/ThirdPartyGraph";

import { commonThirdPartiesQuery, CommonThirdPartyCombobox } from "./CommonThirdPartyCombobox";

const createMappingMutation = graphql`
  mutation AddChildThirdPartyDialogCreateMappingMutation(
    $input: CreateThirdPartyThirdPartyMappingInput!
    $connections: [ID!]!
  ) {
    createThirdPartyThirdPartyMapping(input: $input) {
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

const createThirdPartyMutation = graphql`
  mutation AddChildThirdPartyDialogCreateMutation(
    $input: CreateThirdPartyInput!
  ) {
    createThirdParty(input: $input) {
      thirdPartyEdge {
        node {
          id
        }
      }
    }
  }
`;

type Props = {
  children: ReactNode;
  parentThirdPartyId: string;
  organizationId: string;
  connectionId: string;
  existingChildIds: string[];
};

export function AddChildThirdPartyDialog({
  children,
  parentThirdPartyId,
  organizationId,
  connectionId,
  existingChildIds,
}: Props) {
  const { __ } = useTranslate();
  const dialogRef = useDialogRef();
  const thirdParties = useThirdParties(organizationId);
  const [createMapping] = useMutation<AddChildThirdPartyDialogCreateMappingMutation>(createMappingMutation);
  const [createThirdParty] = useMutation<AddChildThirdPartyDialogCreateMutation>(createThirdPartyMutation);
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

  const existingThirdParties = thirdParties.filter(
    tp =>
      tp.id !== parentThirdPartyId
      && !existingChildIds.includes(tp.id)
      && tp.name.toLowerCase().includes(searchQuery.toLowerCase()),
  );

  const handleSelectExisting = (childId: string) => {
    createMapping({
      variables: {
        input: {
          parentThirdPartyId,
          childThirdPartyId: childId,
        },
        connections: [connectionId],
      },
      onCompleted: () => {
        dialogRef.current?.close();
      },
    });
  };

  const handleSelectCommon = (common: Omit<CreateThirdPartyInput, "organizationId">) => {
    createThirdParty({
      variables: {
        input: {
          ...common,
          organizationId,
          firstLevel: false,
        },
      },
      onCompleted: (response) => {
        const newId = response.createThirdParty.thirdPartyEdge.node.id;
        createMapping({
          variables: {
            input: {
              parentThirdPartyId,
              childThirdPartyId: newId,
            },
            connections: [connectionId],
          },
          onCompleted: () => {
            dialogRef.current?.close();
          },
        });
      },
    });
  };

  const existingNames = new Set(thirdParties.map(tp => tp.name.toLowerCase()));

  return (
    <Dialog ref={dialogRef} trigger={children} title={__("Add a third party")}>
      <DialogContent className="p-6">
        <Combobox onSearch={handleSearch} placeholder={__("Type third party's name")}>
          {existingThirdParties.map(tp => (
            <ComboboxItem key={tp.id} onClick={() => handleSelectExisting(tp.id)}>
              <Avatar name={tp.name} src={faviconUrl(tp.websiteUrl)} size="s" />
              {tp.name}
            </ComboboxItem>
          ))}
          {searchQuery.trim().length >= 2 && queryRef && (
            <Suspense>
              <CommonThirdPartyCombobox
                queryRef={queryRef}
                excludeNames={existingNames}
                onSelect={handleSelectCommon}
              />
            </Suspense>
          )}
        </Combobox>
      </DialogContent>
      <DialogFooter />
    </Dialog>
  );
}
