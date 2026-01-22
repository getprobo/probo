import { graphql } from "relay-runtime";
import { useFragment } from "react-relay";
import { IconPin, IconShield } from "@probo/ui";
import { useTranslate } from "@probo/i18n";
import { faviconUrl, getCountryName } from "@probo/helpers";

import type { VendorRowFragment$key } from "./__generated__/VendorRowFragment.graphql";

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

export function VendorRow(props: { vendor: VendorRowFragment$key; hasAnyCountries?: boolean }) {
  const vendor = useFragment(vendorRowFragment, props.vendor);
  const logo = faviconUrl(vendor.websiteUrl);
  const { __ } = useTranslate();

  const gridCols = props.hasAnyCountries ? "grid-cols-[1fr_1fr_1fr]" : "grid-cols-[1fr_1fr]";

  return (
    <div className="flex text-sm leading-tight gap-3 md:items-center">
      {logo
        ? (
            <img
              src={logo}
              className="size-8 md:size-6 flex-none rounded-lg"
              alt=""
            />
          )
        : (
            <div className="size-8 md:size-6 flex-none rounded-lg" />
          )}
      <div className={`flex flex-col md:grid ${gridCols} flex-1 gap-0.5`}>
        <div>{vendor.name}</div>
        {vendor.privacyPolicyUrl
          ? (
              <a
                href={vendor.privacyPolicyUrl}
                target="_blank"
                className={`flex gap-1 text-txt-info items-center hover:underline ${!props.hasAnyCountries ? "md:justify-end" : ""}`}
              >
                <IconShield size={16} className="flex-none" />
                <span>{__("Privacy")}</span>
              </a>
            )
          : (
              <div></div>
            )}
        {vendor.countries.length > 0 && (
          <div className={`flex gap-1 text-txt-secondary items-center ${props.hasAnyCountries ? "md:justify-end" : ""}`}>
            <IconPin size={16} className="flex-none" />
            <span>
              {vendor.countries
                .map(country => getCountryName(__, country))
                .join(", ")}
            </span>
          </div>
        )}
      </div>
    </div>
  );
}
