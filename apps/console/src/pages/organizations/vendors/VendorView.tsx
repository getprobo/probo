"use client";

import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import { HelpCircle } from "lucide-react";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
} from "react-relay";
import { Suspense, useEffect, useState, useCallback } from "react";
import type { VendorViewQuery as VendorViewQueryType } from "./__generated__/VendorViewQuery.graphql";
import { useParams } from "react-router";
import { cn } from "@/lib/utils";
import { PageTemplate } from "@/components/PageTemplate";
import { VendorViewSkeleton } from "./VendorPage";

const vendorViewQuery = graphql`
  query VendorViewQuery($vendorId: ID!) {
    node(id: $vendorId) {
      ... on Vendor {
        id
        name
        description
        serviceStartAt
        serviceTerminationAt
        serviceCriticality
        riskTier
        statusPageUrl
        termsOfServiceUrl
        privacyPolicyUrl
        createdAt
        updatedAt
      }
    }
  }
`;

const updateVendorMutation = graphql`
  mutation VendorViewUpdateVendorMutation($input: UpdateVendorInput!) {
    updateVendor(input: $input) {
      vendor {
        id
        name
        description
        serviceStartAt
        serviceTerminationAt
        serviceCriticality
        riskTier
        statusPageUrl
        termsOfServiceUrl
        privacyPolicyUrl
        updatedAt
      }
    }
  }
`;

function EditableField({
  label,
  value,
  onChange,
  type = "text",
  helpText,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  helpText?: string;
}) {
  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <HelpCircle className="h-4 w-4 text-tertiary" />
        <Label className="text-sm">{label}</Label>
      </div>
      <div className="space-y-2">
        <Input
          type={type}
          value={value}
          onChange={(e) => onChange(e.target.value)}
        />
        {helpText && <p className="text-sm text-secondary">{helpText}</p>}
      </div>
    </div>
  );
}

// Format date for input field (YYYY-MM-DDTHH:mm)
function formatDateForInput(date: string | null | undefined): string {
  if (!date) return "";
  return new Date(date).toISOString().slice(0, 16);
}

// Format date for API (2006-01-02T15:04:05.999999999Z07:00)
function formatDateForAPI(dateStr: string): string {
  if (!dateStr) return "";
  const date = new Date(dateStr);
  return date.toISOString();
}

function VendorViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<VendorViewQueryType>;
}) {
  const data = usePreloadedQuery(vendorViewQuery, queryRef);
  const [editedFields, setEditedFields] = useState<Set<string>>(new Set());
  const [formData, setFormData] = useState({
    name: data.node.name || "",
    description: data.node.description || "",
    // Format dates properly for datetime-local input
    serviceStartAt: formatDateForInput(data.node.serviceStartAt),
    serviceTerminationAt: formatDateForInput(data.node.serviceTerminationAt),
    serviceCriticality: data.node.serviceCriticality,
    riskTier: data.node.riskTier,
    statusPageUrl: data.node.statusPageUrl || "",
    termsOfServiceUrl: data.node.termsOfServiceUrl || "",
    privacyPolicyUrl: data.node.privacyPolicyUrl || "",
  });
  const [commit] = useMutation(updateVendorMutation);
  const [, loadQuery] = useQueryLoader<VendorViewQueryType>(vendorViewQuery);
  const { toast } = useToast();

  const hasChanges = editedFields.size > 0;

  const handleSave = useCallback(() => {
    const formattedData = {
      ...formData,
      serviceStartAt: formatDateForAPI(formData.serviceStartAt),
      serviceTerminationAt: formData.serviceTerminationAt
        ? formatDateForAPI(formData.serviceTerminationAt)
        : null,
    };

    commit({
      variables: {
        input: {
          id: data.node.id,
          ...formattedData,
        },
      },
      onCompleted: () => {
        toast({
          title: "Success",
          description: "Changes saved successfully",
          variant: "default",
        });
        setEditedFields(new Set());
      },
      onError: (error) => {
        if (error.message?.includes("concurrent modification")) {
          toast({
            title: "Error",
            description:
              "Someone else modified this vendor. Reloading latest data.",
            variant: "destructive",
          });

          loadQuery({ vendorId: data.node.id! });
        } else {
          toast({
            title: "Error",
            description: error.message || "Failed to save changes",
            variant: "destructive",
          });
        }
      },
    });
  }, [commit, data.node.id, formData, loadQuery, toast]);

  const handleFieldChange = (field: keyof typeof formData, value: unknown) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
    setEditedFields((prev) => new Set(prev).add(field));
  };

  // Update the cancel handler to also format dates
  const handleCancel = () => {
    setFormData({
      name: data.node.name || "",
      description: data.node.description || "",
      serviceStartAt: formatDateForInput(data.node.serviceStartAt),
      serviceTerminationAt: formatDateForInput(data.node.serviceTerminationAt),
      serviceCriticality: data.node.serviceCriticality,
      riskTier: data.node.riskTier,
      statusPageUrl: data.node.statusPageUrl || "",
      termsOfServiceUrl: data.node.termsOfServiceUrl || "",
      privacyPolicyUrl: data.node.privacyPolicyUrl || "",
    });
    setEditedFields(new Set());
  };

  return (
    <PageTemplate title={formData.name}>
      <div className="max-w-2xl space-y-6">
        <EditableField
          label="Name"
          value={formData.name}
          onChange={(value) => handleFieldChange("name", value)}
        />

        <EditableField
          label="Description"
          value={formData.description}
          onChange={(value) => handleFieldChange("description", value)}
        />

        <Card className="p-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <h2 className="text-lg font-medium">Service Information</h2>
              <p className="text-sm text-secondary">
                Basic information about the vendor service
              </p>
            </div>

            <div className="space-y-4">
              <EditableField
                label="Service Start At"
                value={formData.serviceStartAt}
                type="datetime-local"
                onChange={(value) => handleFieldChange("serviceStartAt", value)}
              />

              <EditableField
                label="Service Termination At"
                value={formData.serviceTerminationAt}
                type="datetime-local"
                onChange={(value) =>
                  handleFieldChange("serviceTerminationAt", value)
                }
              />

              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <HelpCircle className="h-4 w-4 text-tertiary" />
                  <Label className="text-sm">Service Criticality</Label>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() =>
                      handleFieldChange("serviceCriticality", "LOW")
                    }
                    className={cn(
                      "rounded-full px-4 py-1 text-sm transition-colors",
                      formData.serviceCriticality === "LOW"
                        ? "bg-success-bg text-success ring ring-success-b"
                        : "bg-invert-bg"
                    )}
                  >
                    Low
                  </button>
                  <button
                    onClick={() =>
                      handleFieldChange("serviceCriticality", "MEDIUM")
                    }
                    className={cn(
                      "rounded-full px-4 py-1 text-sm transition-colors",
                      formData.serviceCriticality === "MEDIUM"
                        ? "bg-warning-bg text-warning ring ring-warning-b"
                        : "bg-invert-bg"
                    )}
                  >
                    Medium
                  </button>
                  <button
                    onClick={() =>
                      handleFieldChange("serviceCriticality", "HIGH")
                    }
                    className={cn(
                      "rounded-full px-4 py-1 text-sm transition-colors",
                      formData.serviceCriticality === "HIGH"
                        ? "bg-danger-bg text-danger ring ring-danger-b"
                        : "bg-invert-bg"
                    )}
                  >
                    High
                  </button>
                </div>
                <p className="text-sm text-secondary">
                  {formData.serviceCriticality === "HIGH" &&
                    "Critical service - downtime severely impacts end-users"}
                  {formData.serviceCriticality === "MEDIUM" &&
                    "Important service - downtime moderately affects end-users"}
                  {formData.serviceCriticality === "LOW" &&
                    "Non-critical service - minimal end-user impact if down"}
                </p>
              </div>

              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <HelpCircle className="h-4 w-4 text-tertiary" />
                  <Label className="text-sm">Risk Tier</Label>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleFieldChange("riskTier", "CRITICAL")}
                    className={cn(
                      "rounded-full px-4 py-1 text-sm transition-colors",
                      formData.riskTier === "CRITICAL"
                        ? "bg-danger-bg text-danger ring ring-danger-b"
                        : "bg-invert-bg"
                    )}
                  >
                    Critical
                  </button>
                  <button
                    onClick={() => handleFieldChange("riskTier", "SIGNIFICANT")}
                    className={cn(
                      "rounded-full px-4 py-1 text-sm transition-colors",
                      formData.riskTier === "SIGNIFICANT"
                        ? "bg-warning-bg text-warning ring ring-warning-b"
                        : "bg-invert-bg"
                    )}
                  >
                    Significant
                  </button>
                  <button
                    onClick={() => handleFieldChange("riskTier", "GENERAL")}
                    className={cn(
                      "rounded-full px-4 py-1 text-sm transition-colors",
                      formData.riskTier === "GENERAL"
                        ? "bg-info-bg text-info ring ring-info-b"
                        : "bg-invert-bg"
                    )}
                  >
                    General
                  </button>
                </div>
                <p className="text-sm text-secondary">
                  {formData.riskTier === "CRITICAL" &&
                    "Handles sensitive data, critical for platform operation"}
                  {formData.riskTier === "SIGNIFICANT" &&
                    "No user data access, but important for platform management"}
                  {formData.riskTier === "GENERAL" &&
                    "General vendor with minimal risk"}
                </p>
              </div>

              <EditableField
                label="Status Page URL"
                value={formData.statusPageUrl || ""}
                onChange={(value) => handleFieldChange("statusPageUrl", value)}
              />

              <EditableField
                label="Terms of Service URL"
                value={formData.termsOfServiceUrl || ""}
                onChange={(value) =>
                  handleFieldChange("termsOfServiceUrl", value)
                }
              />

              <EditableField
                label="Privacy Policy URL"
                value={formData.privacyPolicyUrl || ""}
                onChange={(value) =>
                  handleFieldChange("privacyPolicyUrl", value)
                }
              />
            </div>
          </div>
        </Card>
        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" onClick={handleCancel}>
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            className="bg-primary text-invert hover:bg-primary/90"
            disabled={!hasChanges}
          >
            Save Changes
          </Button>
        </div>
      </div>
    </PageTemplate>
  );
}

export default function VendorView() {
  const { vendorId } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<VendorViewQueryType>(vendorViewQuery);

  useEffect(() => {
    loadQuery({ vendorId: vendorId! });
  }, [loadQuery, vendorId]);

  if (!queryRef) {
    return <VendorViewSkeleton />;
  }

  return (
    <Suspense fallback={<VendorViewSkeleton />}>
      <VendorViewContent queryRef={queryRef} />
    </Suspense>
  );
}
