import { useState, Suspense, useEffect } from "react";
import { useNavigate, useParams } from "react-router";
import { graphql, useMutation, ConnectionHandler, useQueryLoader, PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/hooks/use-toast";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { PageTemplate } from "@/components/PageTemplate";
import { NewDatumViewCreateDatumMutation, CreateDatumInput } from "./__generated__/NewDatumViewCreateDatumMutation.graphql";
import { NewDatumViewQuery } from "./__generated__/NewDatumViewQuery.graphql";
import PeopleSelector from "@/components/PeopleSelector";
import { NewDatumViewSkeleton } from "./NewDatumPage";
import { Badge } from "@/components/ui/badge";
import { X } from "lucide-react";

const newDatumViewQuery = graphql`
  query NewDatumViewQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        id
        ...PeopleSelector_organization
        vendors(first: 100, orderBy: { direction: ASC, field: NAME }) @connection(key: "NewDatumView_vendors") {
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

const createDatumMutation = graphql`
  mutation NewDatumViewCreateDatumMutation(
    $input: CreateDatumInput!
    $connections: [ID!]!
  ) {
    createDatum(input: $input) {
      datumEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          dataSensitivity
          owner {
            id
            fullName
          }
          vendors {
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
  }
`;

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
  required,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
  type?: string;
  helpText?: string;
  required?: boolean;
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
          required={required}
        />
        {helpText && <p className="text-sm text-secondary">{helpText}</p>}
      </div>
    </div>
  );
}

function NewDataViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<NewDatumViewQuery>;
}) {
  const navigate = useNavigate();
  const { organizationId } = useParams();
  const [createData] = useMutation<NewDatumViewCreateDatumMutation>(createDatumMutation);
  const { toast } = useToast();
  const data = usePreloadedQuery(newDatumViewQuery, queryRef);

  if (!data.organization) {
    return <div>Organization not found</div>;
  }

  const organization = data.organization as Organization;

  const [formData, setFormData] = useState({
    name: "",
    dataSensitivity: "NONE" as "NONE" | "LOW" | "MEDIUM" | "HIGH" | "CRITICAL",
    ownerId: "",
    selectedVendorIds: [] as string[],
  });

  type FormData = typeof formData;

  const handleFieldChange = <K extends keyof FormData>(
    field: K,
    value: FormData[K]
  ) => {
    setFormData((prev) => ({
      ...prev,
      [field]: value,
    }));
  };

  const handleVendorSelect = (vendorId: string) => {
    setFormData((prev) => ({
      ...prev,
      selectedVendorIds: prev.selectedVendorIds.includes(vendorId)
        ? prev.selectedVendorIds.filter((id) => id !== vendorId)
        : [...prev.selectedVendorIds, vendorId],
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!formData.ownerId) {
      toast({
        title: "Error",
        description: "Please select an owner",
        variant: "destructive",
      });
      return;
    }

    createData({
      variables: {
        connections: [
          ConnectionHandler.getConnectionID(
            organizationId!,
            "DataListView_data",
            {
              orderBy: {
                direction: "ASC",
                field: "NAME",
              },
            },
          ),
        ],
        input: {
          organizationId: organizationId!,
          name: formData.name,
          dataSensitivity: formData.dataSensitivity,
          ownerId: formData.ownerId,
          vendorIds: formData.selectedVendorIds.length > 0 ? formData.selectedVendorIds : undefined,
        } satisfies CreateDatumInput,
      },
      onCompleted: () => {
        toast({
          title: "Success",
          description: "Data created successfully",
          variant: "default",
        });
        navigate(`/organizations/${organizationId}/data`);
      },
      onError: (error) => {
        toast({
          title: "Error",
          description: error.message || "Failed to create data",
          variant: "destructive",
        });
      },
    });
  };

  const vendors = (organization.vendors?.edges || [])
    .map((edge) => edge?.node)
    .filter((node): node is Vendor => node != null);

  return (
    <PageTemplate
      title="Create Data"
      description="Add new data to your organization"
    >
      <form onSubmit={handleSubmit}>
        <div className="max-w-2xl space-y-6">
          <EditableField
            label="Name"
            value={formData.name}
            onChange={(value) => handleFieldChange("name", value)}
            required
          />

          <div className="space-y-2">
            <Label className="text-sm">Owner</Label>
            <PeopleSelector
              organizationRef={data.organization}
              selectedPersonId={formData.ownerId}
              onSelect={(value) => handleFieldChange("ownerId", value)}
              placeholder="Select data owner"
              required
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
            <Label className="text-sm">Data Sensitivity</Label>
            <Select
              value={formData.dataSensitivity}
              onValueChange={(value: string) =>
                handleFieldChange("dataSensitivity", value as "NONE" | "LOW" | "MEDIUM" | "HIGH" | "CRITICAL")}
            >
              <SelectTrigger>
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

          <Button type="submit">Create Data</Button>
        </div>
      </form>
    </PageTemplate>
  );
}

export default function NewDatumView() {
  const { organizationId } = useParams();
  const [queryRef, loadQuery] = useQueryLoader<NewDatumViewQuery>(newDatumViewQuery);

  useEffect(() => {
    loadQuery({ organizationId: organizationId! });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <NewDatumViewSkeleton />;
  }

  return (
    <Suspense fallback={<NewDatumViewSkeleton />}>
      <NewDataViewContent queryRef={queryRef} />
    </Suspense>
  );
}
