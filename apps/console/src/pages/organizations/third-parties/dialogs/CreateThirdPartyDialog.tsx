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
  IconPlusLarge,
  useDialogRef,
} from "@probo/ui";
import type { ThirdParty } from "@probo/third-parties";
import { type ReactNode } from "react";

import { useCreateThirdPartyMutation } from "#/hooks/graph/ThirdPartyGraph";
import { useThirdPartySearch } from "#/hooks/useThirdPartySearch";

type Props = {
  children: ReactNode;
  organizationId: string;
  connection: string;
};

export function CreateThirdPartyDialog({
  children,
  organizationId,
  connection,
}: Props) {
  const { __ } = useTranslate();
  const { search, thirdParties, query } = useThirdPartySearch();
  const [createThirdParty] = useCreateThirdPartyMutation();

  const onSelect = async (thirdParty: ThirdParty | string) => {
    const input
      = typeof thirdParty === "string"
        ? {
            organizationId,
            name: thirdParty,
            category: null,
          }
        : {
            organizationId,
            name: thirdParty.name,
            description: thirdParty.description || null,
            headquarterAddress: thirdParty.headquarterAddress || null,
            legalName: thirdParty.legalName || null,
            websiteUrl: thirdParty.websiteUrl || null,
            category: thirdParty.category || null,
            privacyPolicyUrl: thirdParty.privacyPolicyUrl || null,
            serviceLevelAgreementUrl: thirdParty.serviceLevelAgreementUrl || null,
            dataProcessingAgreementUrl: thirdParty.dataProcessingAgreementUrl || null,
            certifications: thirdParty.certifications,
            countries: thirdParty.countries,
            securityPageUrl: thirdParty.securityPageUrl || null,
            trustPageUrl: thirdParty.trustPageUrl || null,
            statusPageUrl: thirdParty.statusPageUrl || null,
            termsOfServiceUrl: thirdParty.termsOfServiceUrl || null,
          };
    await createThirdParty({
      variables: {
        input,
        connections: [connection],
      },
      onSuccess: () => {
        dialogRef.current?.close();
      },
    });
  };

  const dialogRef = useDialogRef();

  return (
    <Dialog ref={dialogRef} trigger={children} title={__("Add a thirdParty")}>
      <DialogContent className="p-6">
        <Combobox onSearch={search} placeholder={__("Type thirdParty's name")}>
          {thirdParties.map(thirdParty => (
            <ComboboxItem key={thirdParty.name} onClick={() => void onSelect(thirdParty)}>
              <Avatar name={thirdParty.name} src={faviconUrl(thirdParty.websiteUrl)} />
              {thirdParty.name}
            </ComboboxItem>
          ))}
          {query.trim().length >= 2 && (
            <ComboboxItem onClick={() => void onSelect(query.trim())}>
              <IconPlusLarge size={20} />
              {__("Create a new thirdParty")}
              {" "}
              :
              {query}
            </ComboboxItem>
          )}
        </Combobox>
      </DialogContent>
      <DialogFooter />
    </Dialog>
  );
}
