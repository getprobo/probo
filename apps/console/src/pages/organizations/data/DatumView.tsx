"use client";

import { Card } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useToast } from "@/hooks/use-toast";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
} from "react-relay";
import { Suspense, useEffect, useState, useCallback } from "react";
import type { DatumViewQuery as DatumViewQueryType } from "./__generated__/DatumViewQuery.graphql";
import { useParams } from "react-router";
import { PageTemplate } from "@/components/PageTemplate";
import { DatumViewSkeleton } from "./DatumPage";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import PeopleSelector from "@/components/PeopleSelector";
import { Badge } from "@/components/ui/badge";
import { X } from "lucide-react";

const datumViewQuery = graphql`
  query DatumViewQuery($datumId: ID!, $organizationId: ID!) {
    node(id: $datumId) {
      ... on Datum {
        id
        name
        dataSensitivity
        vendors {
          edges {
            node {
              id
              name
            }
          }
        }
        owner {
          id
          fullName
        }
        createdAt
        updatedAt
      }
    }
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        ...PeopleSelector_organization
        vendors(first: 100, orderBy: { direction: ASC, field: NAME }) @connection(key: "DatumView_vendors") {
          edges {
            node {
              id
              name
            }
          }
        }
      }
    }
  }
`;

const updateDatumMutation = graphql`
  mutation DatumViewUpdateDatumMutation($input: UpdateDatumInput!) {
    updateDatum(input: $input) {
      datum {
        id
        name
        dataSensitivity
        vendors {
          edges {
            node {
              id
            }
          }
        }
        owner {
          id
          fullName
        }
        updatedAt
      }
    }
  }
`;

type DataSensitivity = "NONE" | "LOW" | "MEDIUM" | "HIGH" | "CRITICAL";

interface Datum {
  readonly id?: string;
  readonly name?: string;
  readonly description?: string;
  readonly dataSensitivity?: DataSensitivity;
  readonly owner?: {
    readonly id: string;
    readonly fullName: string;
  } | null;
  readonly createdAt?: string;
  readonly updatedAt?: string;
  readonly vendors?: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly id: string;
        readonly name: string;
      } | null;
    } | null> | null;
  } | null;
}

interface Vendor {
  readonly id: string;
  readonly name: string;
}

interface Organization {
  readonly id?: string;
  readonly vendors?: {
    readonly edges: ReadonlyArray<{
      readonly node: Vendor;
    } | null> | null;
  } | null;
}

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

function DatumViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<DatumViewQueryType>;
}) {
  const data = usePreloadedQuery(datumViewQuery, queryRef);
  const [editedFields, setEditedFields] = useState<Set<string>>(new Set());
  const datum = data.node as Datum | null;
  const organization = data.organization as Organization;
  const [formData, setFormData] = useState({
    name: datum?.name || "",
    dataSensitivity: datum?.dataSensitivity || "NONE",
    ownerId: datum?.owner?.id || "",
    selectedVendorIds: datum?.vendors?.edges?.map(edge => edge?.node?.id).filter((id): id is string => id != null) || [],
  });
  const [commit] = useMutation(updateDatumMutation);
  const { toast } = useToast();
  const hasChanges = editedFields.size > 0;

  const handleSave = useCallback(() => {
    const nodeId = data.node?.id;
    if (!nodeId) return;

    commit({
      variables: {
        input: {
          id: nodeId,
          name: formData.name,
          dataSensitivity: formData.dataSensitivity,
          ownerId: formData.ownerId,
          vendorIds: formData.selectedVendorIds.length > 0 ? formData.selectedVendorIds : undefined,
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
        toast({
          title: "Error",
          description: error.message || "Failed to save changes",
          variant: "destructive",
        });
      },
    });
  }, [commit, data.node?.id, formData, toast]);

  const handleFieldChange = (field: keyof typeof formData, value: unknown) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
    setEditedFields((prev) => new Set(prev).add(field));
  };

  const handleVendorSelect = (vendorId: string) => {
    setFormData((prev) => ({
      ...prev,
      selectedVendorIds: prev.selectedVendorIds.includes(vendorId)
        ? prev.selectedVendorIds.filter((id) => id !== vendorId)
        : [...prev.selectedVendorIds, vendorId],
    }));
    setEditedFields((prev) => new Set(prev).add("selectedVendorIds"));
  };

  const handleCancel = () => {
    const datum = data.node as Datum | null;
    setFormData({
      name: datum?.name || "",
      dataSensitivity: datum?.dataSensitivity || "NONE",
      ownerId: datum?.owner?.id || "",
      selectedVendorIds: datum?.vendors?.edges?.map(edge => edge?.node?.id).filter((id): id is string => id != null) || [],
    });
    setEditedFields(new Set());
  };

  const vendors = (organization.vendors?.edges || [])
    .map((edge) => edge?.node)
    .filter((node): node is Vendor => node != null);

  return (
    <PageTemplate title={formData.name || "Data Details"}>
      <div className="space-y-6">
        <div className="max-w-4xl space-y-6">
          <EditableField
            label="Name"
            value={formData.name}
            onChange={(value) => handleFieldChange("name", value)}
          />

          <Card className="p-6">
            <div className="space-y-4">
              <div className="space-y-2">
                <h2 className="text-lg font-medium">Data Details</h2>
              </div>

              <div className="space-y-4">
                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Label className="text-sm">Owner</Label>
                  </div>
                  <PeopleSelector
                    organizationRef={data.organization}
                    selectedPersonId={formData.ownerId}
                    onSelect={(value) => handleFieldChange("ownerId", value)}
                    placeholder="Select data owner"
                  />
                </div>

                <div className="space-y-2">
                  <Label className="text-sm">Vendors</Label>
                  <Select
                    value=""
                    onValueChange={handleVendorSelect}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="Select vendors" />
                    </SelectTrigger>
                    <SelectContent>
                      {vendors.map((vendor) => (
                        <SelectItem key={vendor.id} value={vendor.id}>
                          {vendor.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <div className="flex flex-wrap gap-2 mt-2">
                    {formData.selectedVendorIds.map((vendorId) => {
                      const vendor = vendors.find((v) => v.id === vendorId);
                      if (!vendor) return null;
                      return (
                        <Badge key={vendorId} variant="secondary" className="flex items-center gap-1">
                          {vendor.name}
                          <button
                            type="button"
                            onClick={() => handleVendorSelect(vendorId)}
                            className="hover:bg-secondary-hover rounded-full p-0.5"
                          >
                            <X className="h-3 w-3" />
                          </button>
                        </Badge>
                      );
                    })}
                  </div>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center gap-2">
                    <Label className="text-sm">Data Sensitivity</Label>
                  </div>
                  <Select
                    value={formData.dataSensitivity}
                    onValueChange={(value) => handleFieldChange("dataSensitivity", value)}
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select data sensitivity" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="NONE">No sensitive data</SelectItem>
                      <SelectItem value="LOW">Public or non-sensitive data</SelectItem>
                      <SelectItem value="MEDIUM">Internal/restricted data</SelectItem>
                      <SelectItem value="HIGH">Confidential data</SelectItem>
                      <SelectItem value="CRITICAL">Regulated/PII/financial data</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
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
      </div>
    </PageTemplate>
  );
}

export default function DatumView() {
  const { datumId, organizationId } = useParams();
  const [queryRef, loadQuery] =
    useQueryLoader<DatumViewQueryType>(datumViewQuery);

  useEffect(() => {
    loadQuery({ datumId: datumId!, organizationId: organizationId! });
  }, [loadQuery, datumId, organizationId]);

  if (!queryRef || !datumId || !organizationId) {
    return <DatumViewSkeleton />;
  }

  return (
    <Suspense fallback={<DatumViewSkeleton />}>
      <DatumViewContent queryRef={queryRef} />
    </Suspense>
  );
}
