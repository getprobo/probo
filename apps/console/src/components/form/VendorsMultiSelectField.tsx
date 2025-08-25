import { Avatar, Field, Option, Select, Badge, Button, IconCrossLargeX } from "@probo/ui";
import { Suspense, useState, type ComponentProps } from "react";
import { useTranslate } from "@probo/i18n";
import { type Control, Controller, type FieldValues } from "react-hook-form";
import { useVendors } from "/hooks/graph/VendorGraph.ts";
import { faviconUrl } from "@probo/helpers";
import type { Path } from "react-hook-form";

type Vendor = {
  id: string;
  name: string;
  websiteUrl: string | null | undefined;
};

type Props<T extends FieldValues = FieldValues> = {
  organizationId: string;
  control: Control<T>;
  name: string;
  label?: string;
  error?: string;
  selectedVendors?: Vendor[];
} & ComponentProps<typeof Field>;

export function VendorsMultiSelectField<T extends FieldValues = FieldValues>({
  organizationId,
  control,
  selectedVendors = [],
  ...props
}: Props<T>) {
  return (
    <Field {...props}>
      <Suspense
        fallback={<Select variant="editor" disabled placeholder="Loading..." />}
      >
        <VendorsMultiSelectWithQuery
          organizationId={organizationId}
          control={control}
          name={props.name}
          disabled={props.disabled}
          selectedVendors={selectedVendors}
        />
      </Suspense>
    </Field>
  );
}

function VendorsMultiSelectWithQuery<T extends FieldValues = FieldValues>(
  props: Pick<Props<T>, "organizationId" | "control" | "name" | "disabled" | "selectedVendors">
) {
  const { __ } = useTranslate();
  const { name, organizationId, control, selectedVendors = [] } = props;
  const vendors = useVendors(organizationId);
  const [isOpen, setIsOpen] = useState(false);

  const allVendors = [...vendors];
  if (props.disabled) {
    selectedVendors.forEach(selectedVendor => {
      if (!allVendors.find(v => v.id === selectedVendor.id)) {
        allVendors.push(selectedVendor);
      }
    });
  }

  return (
    <>
      <Controller
        control={control}
        name={name as Path<T>}
        render={({ field }) => {
          const selectedVendorIds = (Array.isArray(field.value) ? field.value : []) as string[];

          const selectedVendors = allVendors.filter(v => selectedVendorIds.includes(v.id));
          const availableVendors = allVendors.filter(v => !selectedVendorIds.includes(v.id));

          const handleAddVendor = (vendorId: string) => {
            const newValue = [...selectedVendorIds, vendorId];
            field.onChange(newValue);
            setIsOpen(false);
          };

          const handleRemoveVendor = (vendorId: string) => {
            const newValue = selectedVendorIds.filter((id: string) => id !== vendorId);
            field.onChange(newValue);
          };

          return (
            <div className="space-y-2">
              {availableVendors.length > 0 && !props.disabled && (
                <Select
                  disabled={props.disabled}
                  id={name}
                  variant="editor"
                  placeholder={__("Add vendors...")}
                  onValueChange={handleAddVendor}
                  key={`${selectedVendorIds.length}-${vendors.length}`}
                  className="w-full"
                  value=""
                  open={isOpen}
                  onOpenChange={setIsOpen}
                >
                  {availableVendors.map((vendor) => (
                    <Option key={vendor.id} value={vendor.id} className="flex gap-2">
                      <Avatar
                        name={vendor.name}
                        src={faviconUrl(vendor.websiteUrl)}
                        size="s"
                      />
                      <div className="flex flex-col">
                        <span>{vendor.name}</span>
                        {vendor.websiteUrl && (
                          <span className="text-xs text-txt-secondary">
                            {vendor.websiteUrl}
                          </span>
                        )}
                      </div>
                    </Option>
                  ))}
                </Select>
              )}

              {selectedVendors.length > 0 && (
                <div className="flex flex-wrap gap-2">
                  {selectedVendors.map((vendor) => (
                    <Badge key={vendor.id} variant="neutral" className="flex items-center gap-2">
                      <Avatar
                        name={vendor.name}
                        src={faviconUrl(vendor.websiteUrl)}
                        size="s"
                      />
                      <span>{vendor.name}</span>
                      {!props.disabled && (
                        <Button
                          variant="tertiary"
                          icon={IconCrossLargeX}
                          onClick={() => handleRemoveVendor(vendor.id)}
                          className="h-4 w-4 p-0 hover:bg-transparent"
                        />
                      )}
                    </Badge>
                  ))}
                </div>
              )}

              {selectedVendors.length === 0 && availableVendors.length === 0 && (
                <div className="text-sm text-txt-secondary py-2">
                  {__("No vendors available")}
                </div>
              )}
            </div>
          );
        }}
      />
    </>
  );
}
