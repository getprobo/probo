import { faviconUrl, getCountryName } from "@probo/helpers";
import { useTranslate } from "@probo/i18n";
import { IconPin } from "@probo/ui";
import { useFragment } from "react-relay";
import { graphql } from "relay-runtime";

import type { VendorRowFragment$key } from "./__generated__/VendorRowFragment.graphql";

const vendorRowFragment = graphql`
  fragment VendorRowFragment on Vendor {
    name
    description
    websiteUrl
    countries
  }
`;

export function VendorRow(props: { vendor: VendorRowFragment$key; hasAnyCountries?: boolean }) {
  const vendor = useFragment(vendorRowFragment, props.vendor);
  const logo = faviconUrl(vendor.websiteUrl);
  const { __ } = useTranslate();

  return (
    <div className="flex text-sm leading-tight gap-6 items-center">
      {logo
        ? (
            <img
              src={logo}
              className="size-8 flex-none rounded-lg"
              alt=""
            />
          )
        : (
            <div className="size-8 flex-none rounded-lg" />
          )}
      <div className="flex flex-col gap-2 flex-1">
        <span className="text-sm">{vendor.name}</span>
        <div className="text-xs text-txt-secondary w-full">{vendor.description}</div>
        {props.hasAnyCountries
          && (
            <div className="text-xs flex gap-1 items-start text-txt-quaternary">
              {vendor.countries.length > 0 && (
                <>
                  <IconPin size={16} className="flex-none" />
                  <span>
                    {vendor.countries
                      .map(country => getCountryName(__, country))
                      .join(", ")}
                  </span>
                </>
              )}
            </div>
          )}
      </div>
    </div>
  );
}
