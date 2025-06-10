import { useTranslate } from "@probo/i18n";
import {
  Badge,
  Button,
  Card,
  Combobox,
  ComboboxItem,
  IconCrossLargeX,
  IconPlusLarge,
} from "@probo/ui";
import { Controller } from "react-hook-form";
import { useVendorForm } from "/hooks/forms/useVendorForm";
import type { useVendorFormFragment$key } from "/hooks/forms/__generated__/useVendorFormFragment.graphql";
import { useOutletContext } from "react-router";
import {
  certificationCategoryLabel,
  certifications,
  objectEntries,
} from "@probo/helpers";
import { useRef, useState } from "react";
import clsx from "clsx";

/**
 * Vendor certifications tab
 */
export default function VendorCertificationsTab() {
  const { vendor } = useOutletContext<{
    vendor: useVendorFormFragment$key & { name: string };
  }>();
  const { __ } = useTranslate();
  const { control, handleSubmit } = useVendorForm(vendor);

  return (
    <form className="space-y-4" onSubmit={handleSubmit}>
      <Card padded>
        <Controller
          control={control}
          name="certifications"
          render={({ field }) => (
            <Certifications
              onValueChange={field.onChange}
              value={field.value}
            />
          )}
        />
      </Card>
      <div className="flex justify-end">
        <Button type="submit">{__("Update vendor")}</Button>
      </div>
    </form>
  );
}

type CertificationsProps = {
  value: string[];
  onValueChange: (value: string[]) => void;
};

/**
 * List all certifications badges
 */
function Certifications(props: CertificationsProps) {
  const categorizedCertifications = Object.values(certifications).flat();
  const { __ } = useTranslate();
  const animateBadge = useRef(false);
  const categories = objectEntries(certifications)
    .map(
      ([key, value]) =>
        [key, value.filter((c) => props.value.includes(c))] as const
    )
    .filter(([_, certifications]) => certifications.length > 0);
  categories.push([
    "custom",
    props.value.filter((c) => !categorizedCertifications.includes(c)),
  ]);

  const addCertificate = (name: string) => {
    animateBadge.current = true;
    props.onValueChange([...props.value, name]);
  };

  const removeCertificate = (name: string) => {
    animateBadge.current = true;
    props.onValueChange(props.value.filter((v) => v !== name));
  };

  return (
    <div className="space-y-6">
      {categories.map(([key, certifications]) => (
        <div key={key} className="space-y-2">
          <div className="text-sm font-medium text-txt-secondary">
            {certificationCategoryLabel(__, key)}
          </div>
          <div className="flex flex-wrap gap-2">
            {certifications.map((certification) => (
              <Badge asChild size="md" key={certification}>
                <button
                  onClick={() => removeCertificate(certification)}
                  type="button"
                  className={clsx(
                    "hover:bg-subtle-hover cursor-pointer",
                    animateBadge.current &&
                      "starting:opacity-0 starting:w-0 w-max transition-all duration-500 starting:bg-accent"
                  )}
                >
                  {certification}
                  <div className="w-0 overflow-hidden group-hover:w-4 duration-200">
                    <IconCrossLargeX size={12} />
                  </div>
                </button>
              </Badge>
            ))}
          </div>
        </div>
      ))}
      <CertificationInput
        certifications={categorizedCertifications.filter(
          (c) => !props.value.includes(c)
        )}
        onAdd={addCertificate}
      />
    </div>
  );
}

/**
 * Input to add a new certification
 */
function CertificationInput({
  certifications,
  onAdd,
}: {
  certifications: string[];
  onAdd: (name: string) => void;
}) {
  const { __ } = useTranslate();
  const [search, setSearch] = useState("");
  const isCustom = !certifications.includes(search.trim());
  const filteredCertifications = certifications.filter((c) =>
    c.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="flex items-center gap-2">
      <Combobox
        autoSelect
        resetValueOnHide
        onSelect={onAdd}
        onSearch={setSearch}
        placeholder={__("Add a new certification")}
      >
        {filteredCertifications.map((certification) => (
          <ComboboxItem key={certification} value={certification}>
            {certification}
          </ComboboxItem>
        ))}
        {isCustom && search.trim().length >= 2 && (
          <ComboboxItem value={search.trim()}>
            <IconPlusLarge size={20} />
            {__("Add a custom certification")} : {search}
          </ComboboxItem>
        )}
      </Combobox>
    </div>
  );
}
