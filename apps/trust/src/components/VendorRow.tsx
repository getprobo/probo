import { graphql } from "relay-runtime";
import type { VendorRowFragment$key } from "./__generated__/VendorRowFragment.graphql";
import { useFragment } from "react-relay";
import { IconPin, Td, Tr } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { faviconUrl, getCountryName } from "@probo/helpers";

const vendorRowFragment = graphql`
  fragment VendorRowFragment on Vendor {
    id
    name
    category
    websiteUrl
    privacyPolicyUrl
    countries
  }
`;

export function VendorRow(props: { vendor: VendorRowFragment$key }) {
  const vendor = useFragment(vendorRowFragment, props.vendor);
  const logo = faviconUrl(vendor.websiteUrl);
  const { __ } = useTranslate();

  return (
    <Tr className="text-sm *:border-border-solid *:border-b-1">
      <Td className="">
        <div className="flex items-center gap-3 py-1">
          {logo ? <img src={logo} className="size-6" alt="" /> : null}
          {vendor.name}
        </div>
      </Td>
      <Td className=" text-txt-info">
        {vendor.privacyPolicyUrl && (
          <a href={vendor.privacyPolicyUrl} target="_blank">
            {vendor.privacyPolicyUrl.split("//").at(-1)}
          </a>
        )}
      </Td>
      <Td className="text-end">
        <div className="flex gap-2 text-txt-secondary items-center text-sm justify-end">
          <IconPin size={16} />
          <span>
            {vendor.countries
              .map((country) => getCountryName(__, country))
              .join(", ")}
          </span>
        </div>
      </Td>
    </Tr>
  );
}
