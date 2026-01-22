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
import type { Vendor } from "@probo/vendors";
import { type ReactNode } from "react";

import { useCreateVendorMutation } from "#/hooks/graph/VendorGraph";
import { useVendorSearch } from "#/hooks/useVendorSearch";

type Props = {
  children: ReactNode;
  organizationId: string;
  connection: string;
};

export function CreateVendorDialog({
  children,
  organizationId,
  connection,
}: Props) {
  const { __ } = useTranslate();
  const { search, vendors, query } = useVendorSearch();
  const [createVendor] = useCreateVendorMutation();

  const onSelect = async (vendor: Vendor | string) => {
    const input
      = typeof vendor === "string"
        ? {
            organizationId,
            name: vendor,
            category: null,
          }
        : {
            organizationId,
            name: vendor.name,
            description: vendor.description || null,
            headquarterAddress: vendor.headquarterAddress || null,
            legalName: vendor.legalName || null,
            websiteUrl: vendor.websiteUrl || null,
            category: vendor.category || null,
            privacyPolicyUrl: vendor.privacyPolicyUrl || null,
            serviceLevelAgreementUrl: vendor.serviceLevelAgreementUrl || null,
            dataProcessingAgreementUrl: vendor.dataProcessingAgreementUrl || null,
            certifications: vendor.certifications,
            countries: vendor.countries,
            securityPageUrl: vendor.securityPageUrl || null,
            trustPageUrl: vendor.trustPageUrl || null,
            statusPageUrl: vendor.statusPageUrl || null,
            termsOfServiceUrl: vendor.termsOfServiceUrl || null,
          };
    await createVendor({
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
    <Dialog ref={dialogRef} trigger={children} title={__("Add a vendor")}>
      <DialogContent className="p-6">
        <Combobox onSearch={search} placeholder={__("Type vendor's name")}>
          {vendors.map(vendor => (
            <ComboboxItem key={vendor.name} onClick={() => void onSelect(vendor)}>
              <Avatar name={vendor.name} src={faviconUrl(vendor.websiteUrl)} />
              {vendor.name}
            </ComboboxItem>
          ))}
          {query.trim().length >= 2 && (
            <ComboboxItem onClick={() => void onSelect(query.trim())}>
              <IconPlusLarge size={20} />
              {__("Create a new vendor")}
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
