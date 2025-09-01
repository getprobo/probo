import { useTranslate } from "@probo/i18n";
import { useVendorForm } from "/hooks/forms/useVendorForm";
import { useOutletContext, useParams } from "react-router";
import { Button, Card, Field, Input, IconPlusLarge, IconTrashCan, IconPencil } from "@probo/ui";
import { PeopleSelectField } from "/components/form/PeopleSelectField";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { useMemo } from "react";
import { usePageTitle } from "@probo/hooks";
import { downloadFile } from "@probo/helpers";
import { useFragment, graphql } from "react-relay";
import { UploadBusinessAssociateAgreementDialog } from "../dialogs/UploadBusinessAssociateAgreementDialog";
import { DeleteBusinessAssociateAgreementDialog } from "../dialogs/DeleteBusinessAssociateAgreementDialog";
import { EditBusinessAssociateAgreementDialog } from "../dialogs/EditBusinessAssociateAgreementDialog";
import { UploadDataPrivacyAgreementDialog } from "../dialogs/UploadDataPrivacyAgreementDialog";
import { DeleteDataPrivacyAgreementDialog } from "../dialogs/DeleteDataPrivacyAgreementDialog";
import { EditDataPrivacyAgreementDialog } from "../dialogs/EditDataPrivacyAgreementDialog";
import type { useVendorFormFragment$key } from "/hooks/forms/__generated__/useVendorFormFragment.graphql";
import type { VendorOverviewTabBusinessAssociateAgreementFragment$key } from "./__generated__/VendorOverviewTabBusinessAssociateAgreementFragment.graphql";
import type { VendorOverviewTabDataPrivacyAgreementFragment$key } from "./__generated__/VendorOverviewTabDataPrivacyAgreementFragment.graphql";

const vendorBusinessAssociateAgreementFragment = graphql`
  fragment VendorOverviewTabBusinessAssociateAgreementFragment on Vendor {
    businessAssociateAgreement {
      id
      fileName
      fileUrl
      validFrom
      validUntil
      createdAt
    }
  }
`;

const vendorDataPrivacyAgreementFragment = graphql`
  fragment VendorOverviewTabDataPrivacyAgreementFragment on Vendor {
    dataPrivacyAgreement {
      id
      fileName
      fileUrl
      validFrom
      validUntil
      createdAt
    }
  }
`;

export default function VendorOverviewTab() {
  const { vendor } = useOutletContext<{
    vendor: useVendorFormFragment$key & { id: string; name: string };
  }>();

  const { vendor: vendorForBAA } = useOutletContext<{
    vendor: VendorOverviewTabBusinessAssociateAgreementFragment$key & VendorOverviewTabDataPrivacyAgreementFragment$key;
  }>();

  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const { snapshotId } = useParams<{ snapshotId?: string }>();
  const isSnapshotMode = Boolean(snapshotId);

  const {
    control,
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useVendorForm(vendor);

  const vendorWithBAA = useFragment<VendorOverviewTabBusinessAssociateAgreementFragment$key>(
    vendorBusinessAssociateAgreementFragment,
    vendorForBAA
  );
  const businessAssociateAgreement = vendorWithBAA.businessAssociateAgreement;

  const vendorWithDPA = useFragment<VendorOverviewTabDataPrivacyAgreementFragment$key>(
    vendorDataPrivacyAgreementFragment,
    vendorForBAA
  );
  const dataPrivacyAgreement = vendorWithDPA.dataPrivacyAgreement;

  const urls = useMemo(
    () =>
      [
        { name: "statusPageUrl", label: __("Status page URL") },
        { name: "termsOfServiceUrl", label: __("Terms of service URL") },
        { name: "privacyPolicyUrl", label: __("Privacy document URL") },
        {
          name: "serviceLevelAgreementUrl",
          label: __("Service level agreement URL"),
        },
        {
          name: "dataProcessingAgreementUrl",
          label: __("Data processing agreement URL"),
        },
        { name: "securityPageUrl", label: __("Security page URL") },
        { name: "trustPageUrl", label: __("Trust page URL") },
      ] as const,
    [],
  );

  usePageTitle(vendor.name + " - " + __("Overview"));

  return (
    <form onSubmit={isSnapshotMode ? undefined : handleSubmit} className="space-y-12">
      {/* Vendor Details */}
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Vendor details")}</h2>
        <Card className="space-y-4" padded>
          <Field
            {...register("name")}
            label={__("Name")}
            type="text"
            error={errors.name?.message}
            disabled={isSubmitting || isSnapshotMode}
          />
          <Field
            {...register("description")}
            label={__("Description")}
            type="textarea"
            error={errors.description?.message}
            disabled={isSubmitting || isSnapshotMode}
          />
          <Field
            {...register("legalName")}
            label={__("Legal name")}
            type="text"
            error={errors.legalName?.message}
            disabled={isSubmitting || isSnapshotMode}
          />
          <Field
            {...register("headquarterAddress")}
            label={__("Headquarter address")}
            type="textarea"
            error={errors.headquarterAddress?.message}
            disabled={isSubmitting || isSnapshotMode}
          />
          <Field
            {...register("websiteUrl")}
            label={__("Website URL")}
            type="text"
            error={errors.websiteUrl?.message}
            disabled={isSubmitting || isSnapshotMode}
          />
        </Card>
      </div>

      {/* Ownership */}
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Ownership details")}</h2>
        <Card className="space-y-4" padded>
          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="businessOwnerId"
            label={__("Business owner")}
            error={errors.businessOwnerId?.message}
            disabled={isSubmitting || isSnapshotMode}
            optional={true}
          />
          <PeopleSelectField
            organizationId={organizationId}
            control={control}
            name="securityOwnerId"
            label={__("Security owner")}
            error={errors.securityOwnerId?.message}
            disabled={isSubmitting || isSnapshotMode}
            optional={true}
          />
        </Card>
      </div>

      {/* Links */}
      <div className="space-y-4 mb-4">
        <h2 className="text-base font-medium">{__("Links")}</h2>
        <Card className="divide-y divide-border-low">
          {urls.map((url) => (
            <div
              key={url.name}
              className="grid grid-cols-2 items-center divide-x divide-border-low"
            >
              <label
                className="p-4 text-sm font-medium text-txt-secondary"
                htmlFor={url.name}
              >
                {url.label}
              </label>
              <Input
                className="p-4 focus:bg-tertiary-pressed outline-none"
                id={url.name}
                key={url.name}
                {...register(url.name)}
                type="text"
                placeholder="https://..."
                variant="ghost"
                disabled={isSubmitting || isSnapshotMode}
              />
            </div>
          ))}
        </Card>
      </div>

      {/* Data agreements */}
      <div className="space-y-4">
        <h2 className="text-base font-medium">{__("Data agreements")}</h2>
        <Card className="space-y-4" padded>
          <div className="flex items-center justify-between p-4 border border-border-low rounded-lg">
            <div className="flex-1">
              <h3 className="font-medium text-txt-primary">
                {__("Business Associate Agreement")}
              </h3>
              <p className="text-sm text-txt-secondary mt-1">
                {businessAssociateAgreement ? businessAssociateAgreement.fileName : __("No business associate agreement available")}
              </p>
              {(businessAssociateAgreement?.validFrom || businessAssociateAgreement?.validUntil) && (
                <p className="text-xs text-txt-secondary mt-1">
                  {__("Valid")}
                  {businessAssociateAgreement.validFrom &&
                    ` ${__("from")} ${new Date(businessAssociateAgreement.validFrom).toLocaleDateString()}`
                  }
                  {businessAssociateAgreement.validUntil &&
                    ` ${__("until")} ${new Date(businessAssociateAgreement.validUntil).toLocaleDateString()}`
                  }
                </p>
              )}
            </div>
            <div className="flex items-center gap-2">
              {businessAssociateAgreement ? (
                <>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => downloadFile(businessAssociateAgreement.fileUrl, businessAssociateAgreement.fileName)}
                  >
                    {__("Download PDF")}
                  </Button>
                  {!isSnapshotMode && (
                    <>
                      <EditBusinessAssociateAgreementDialog
                        vendorId={vendor.id}
                        agreement={{
                          validFrom: businessAssociateAgreement.validFrom,
                          validUntil: businessAssociateAgreement.validUntil,
                        }}
                        onSuccess={() => window.location.reload()}
                      >
                        <Button variant="quaternary" icon={IconPencil} />
                      </EditBusinessAssociateAgreementDialog>
                      <DeleteBusinessAssociateAgreementDialog
                        vendorId={vendor.id}
                        fileName={businessAssociateAgreement.fileName}
                        onSuccess={() => window.location.reload()}
                      >
                        <Button variant="quaternary" icon={IconTrashCan} />
                      </DeleteBusinessAssociateAgreementDialog>
                    </>
                  )}
                </>
              ) : (
                !isSnapshotMode && (
                  <UploadBusinessAssociateAgreementDialog
                    vendorId={vendor.id}
                    onSuccess={() => window.location.reload()}
                  >
                    <Button variant="secondary" icon={IconPlusLarge}>
                      {__("Upload")}
                    </Button>
                  </UploadBusinessAssociateAgreementDialog>
                )
              )}
            </div>
          </div>

          <div className="flex items-center justify-between p-4 border border-border-low rounded-lg">
            <div className="flex-1">
              <h3 className="font-medium text-txt-primary">
                {__("Data Privacy Agreement")}
              </h3>
              <p className="text-sm text-txt-secondary mt-1">
                {dataPrivacyAgreement ? dataPrivacyAgreement.fileName : __("No data privacy agreement available")}
              </p>
              {(dataPrivacyAgreement?.validFrom || dataPrivacyAgreement?.validUntil) && (
                <p className="text-xs text-txt-secondary mt-1">
                  {__("Valid")}
                  {dataPrivacyAgreement.validFrom &&
                    ` ${__("from")} ${new Date(dataPrivacyAgreement.validFrom).toLocaleDateString()}`
                  }
                  {dataPrivacyAgreement.validUntil &&
                    ` ${__("until")} ${new Date(dataPrivacyAgreement.validUntil).toLocaleDateString()}`
                  }
                </p>
              )}
            </div>
            <div className="flex items-center gap-2">
              {dataPrivacyAgreement ? (
                <>
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={() => downloadFile(dataPrivacyAgreement.fileUrl, dataPrivacyAgreement.fileName)}
                  >
                    {__("Download PDF")}
                  </Button>
                  {!isSnapshotMode && (
                    <>
                      <EditDataPrivacyAgreementDialog
                        vendorId={vendor.id}
                        agreement={{
                          validFrom: dataPrivacyAgreement.validFrom,
                          validUntil: dataPrivacyAgreement.validUntil,
                        }}
                        onSuccess={() => window.location.reload()}
                      >
                        <Button variant="quaternary" icon={IconPencil} />
                      </EditDataPrivacyAgreementDialog>
                      <DeleteDataPrivacyAgreementDialog
                        vendorId={vendor.id}
                        fileName={dataPrivacyAgreement.fileName}
                        onSuccess={() => window.location.reload()}
                      >
                        <Button variant="quaternary" icon={IconTrashCan} />
                      </DeleteDataPrivacyAgreementDialog>
                    </>
                  )}
                </>
              ) : (
                !isSnapshotMode && (
                  <UploadDataPrivacyAgreementDialog
                    vendorId={vendor.id}
                    onSuccess={() => window.location.reload()}
                  >
                    <Button variant="secondary" icon={IconPlusLarge}>
                      {__("Upload")}
                    </Button>
                  </UploadDataPrivacyAgreementDialog>
                )
              )}
            </div>
          </div>
        </Card>
      </div>

      {/* Submit */}
      {!isSnapshotMode && (
        <div className="flex justify-end">
          <Button type="submit" disabled={isSubmitting}>
            {__("Update vendor")}
          </Button>
        </div>
      )}
    </form>
  );
}
