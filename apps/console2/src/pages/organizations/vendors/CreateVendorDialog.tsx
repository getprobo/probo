import { useTranslate } from "@probo/i18n";
import {
  Combobox,
  ComboboxItem,
  Dialog,
  DialogContent,
  DialogFooter,
  Avatar,
  IconPlusLarge,
  useDialogRef,
} from "@probo/ui";
import { useVendorSearch } from "/hooks/useVendorSearch";
import { faviconUrl } from "@probo/helpers";
import type { Vendor } from "@probo/vendors";
import { useCreateVendorMutation } from "/hooks/graph/VendorGraph";
import { type ReactNode } from "react";

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

  const onSelect = (vendor: Vendor | string) => {
    const input =
      typeof vendor === "string"
        ? {
            organizationId,
            name: vendor,
            description: "",
          }
        : {
            organizationId,
            name: vendor.name,
            description: vendor.description,
            headquarterAddress: vendor.headquarterAddress,
            legalName: vendor.legalName,
            websiteUrl: vendor.websiteUrl,
            category: vendor.category,
            privacyPolicyUrl: vendor.privacyPolicyUrl,
            serviceLevelAgreementUrl: vendor.serviceLevelAgreementUrl,
            dataProcessingAgreementUrl: vendor.dataProcessingAgreementUrl,
            certifications: vendor.certifications,
            securityPageUrl: vendor.securityPageUrl,
            trustPageUrl: vendor.trustPageUrl,
            statusPageUrl: vendor.statusPageUrl,
            termsOfServiceUrl: vendor.termsOfServiceUrl,
          };
    createVendor({
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
          {vendors.map((vendor) => (
            <ComboboxItem key={vendor.name} onClick={() => onSelect(vendor)}>
              <Avatar name={vendor.name} src={faviconUrl(vendor.websiteUrl)} />
              {vendor.name}
            </ComboboxItem>
          ))}
          {query.trim().length >= 2 && (
            <ComboboxItem onClick={() => onSelect(query.trim())}>
              <IconPlusLarge size={20} />
              {__("Create a new vendor")} : {query}
            </ComboboxItem>
          )}
        </Combobox>
      </DialogContent>
      <DialogFooter />
    </Dialog>
  );
}
