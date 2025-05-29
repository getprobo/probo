import { Suspense, useEffect, useState, useTransition, useRef } from "react";
import {
  graphql,
  PreloadedQuery,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
  usePaginationFragment,
} from "react-relay";
import { useSearchParams, useParams } from "react-router";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Store, Plus, MoreHorizontal } from "lucide-react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Link } from "react-router";
import Fuse from "fuse.js";
import { useToast } from "@/hooks/use-toast";
import { PageTemplate } from "@/components/PageTemplate";
import { ListVendorViewSkeleton } from "./ListVendorPage";
import { ListVendorViewCreateVendorMutation, VendorCategory } from "./__generated__/ListVendorViewCreateVendorMutation.graphql";
import { ListVendorViewPaginationQuery } from "./__generated__/ListVendorViewPaginationQuery.graphql";
import { ListVendorView_vendors$key } from "./__generated__/ListVendorView_vendors.graphql";
import { ListVendorViewQuery } from "./__generated__/ListVendorViewQuery.graphql";
import { ListVendorViewDeleteVendorMutation } from "./__generated__/ListVendorViewDeleteVendorMutation.graphql";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Vendor, Vendors } from "@probo/vendors";

const ITEMS_PER_PAGE = 25;

const listVendorViewQuery = graphql`
  query ListVendorViewQuery(
    $organizationId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
  ) {
    organization: node(id: $organizationId) {
      id

      ...ListVendorView_vendors
        @arguments(first: $first, after: $after, last: $last, before: $before)
    }
  }
`;

const vendorListFragment = graphql`
  fragment ListVendorView_vendors on Organization
  @refetchable(queryName: "ListVendorViewPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int" }
    after: { type: "CursorKey" }
    last: { type: "Int" }
    before: { type: "CursorKey" }
  ) {
    vendors(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { direction: ASC, field: NAME }
    ) @connection(key: "VendorListView_vendors") {
      __id
      edges {
        node {
          id
          name
          description
          websiteUrl
          createdAt
          updatedAt
          riskAssessments(
            first: 1
            orderBy: { direction: DESC, field: ASSESSED_AT }
          ) {
            edges {
              node {
                id
                assessedAt
                expiresAt
                dataSensitivity
                businessImpact
              }
            }
          }
        }
      }
      pageInfo {
        hasNextPage
        hasPreviousPage
        startCursor
        endCursor
      }
    }
  }
`;

const createVendorMutation = graphql`
  mutation ListVendorViewCreateVendorMutation(
    $input: CreateVendorInput!
    $connections: [ID!]!
  ) {
    createVendor(input: $input) {
      vendorEdge @prependEdge(connections: $connections) {
        node {
          id
          name
          description
          websiteUrl
          createdAt
          updatedAt
        }
      }
    }
  }
`;

const deleteVendorMutation = graphql`
  mutation ListVendorViewDeleteVendorMutation(
    $input: DeleteVendorInput!
    $connections: [ID!]!
  ) {
    deleteVendor(input: $input) {
      deletedVendorId @deleteEdge(connections: $connections)
    }
  }
`;

function LoadAboveButton({
  isLoading,
  hasMore,
  onLoadMore,
}: {
  isLoading: boolean;
  hasMore: boolean;
  onLoadMore: () => void;
}) {
  if (!hasMore) {
    return null;
  }

  return (
    <div className="flex justify-center mt-4">
      <Button
        variant="outline"
        onClick={onLoadMore}
        disabled={isLoading || !hasMore}
        className="w-full"
      >
        {isLoading ? "Loading..." : "Load above"}
      </Button>
    </div>
  );
}

function LoadBelowButton({
  isLoading,
  hasMore,
  onLoadMore,
}: {
  isLoading: boolean;
  hasMore: boolean;
  onLoadMore: () => void;
}) {
  if (!hasMore) {
    return null;
  }

  return (
    <div className="flex justify-center mt-4">
      <Button
        variant="outline"
        onClick={onLoadMore}
        disabled={isLoading || !hasMore}
        className="w-full"
      >
        {isLoading ? "Loading..." : "Load below"}
      </Button>
    </div>
  );
}

function ListVendorContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<ListVendorViewQuery>;
}) {
  const { toast } = useToast();
  const inputRef = useRef<HTMLInputElement>(null);
  const data = usePreloadedQuery<ListVendorViewQuery>(
    listVendorViewQuery,
    queryRef,
  );
  const [, setSearchParams] = useSearchParams();
  const [, startTransition] = useTransition();
  const [searchTerm, setSearchTerm] = useState("");
  const [filteredVendors, setFilteredVendors] = useState<Vendors>([]);
  const [vendorsData, setVendorsData] = useState<Vendors>([]);
  const [isLoadingVendors, setIsLoadingVendors] = useState(false);
  const [createVendor] =
    useMutation<ListVendorViewCreateVendorMutation>(createVendorMutation);
  const [deleteVendor] =
    useMutation<ListVendorViewDeleteVendorMutation>(deleteVendorMutation);
  const { organizationId } = useParams();

  useEffect(() => {
    const loadVendorsData = async () => {
      try {
        setIsLoadingVendors(true);
        const { default: vendors } = await import("@probo/vendors");
        setVendorsData(vendors);
      } catch (error) {
        console.error("Error loading vendors data:", error);
        toast({
          title: "Error",
          description: "Failed to load vendors data",
          variant: "destructive",
        });
      } finally {
        setIsLoadingVendors(false);
      }
    };

    loadVendorsData();
  }, [toast]);

  const {
    data: vendorsConnection,
    loadNext,
    loadPrevious,
    hasNext,
    hasPrevious,
    isLoadingNext,
    isLoadingPrevious,
  } = usePaginationFragment<
    ListVendorViewPaginationQuery,
    ListVendorView_vendors$key
  >(vendorListFragment, data.organization);

  const vendors =
    vendorsConnection.vendors.edges.map((edge) => edge.node) ?? [];
  const pageInfo = vendorsConnection.vendors.pageInfo;

  const fuse = new Fuse<Vendor>(vendorsData, {
    keys: ["name"],
    threshold: 0.3,
  });

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setSearchTerm(value);
    if (value.trim() === "") {
      setFilteredVendors([]);
    } else {
      const results = fuse.search(value).map((result) => result.item);
      setFilteredVendors(results);
    }
  };

  // Helper to format date (similar to "Mon, 8 Mar. 2025" in the design)
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString("en-US", {
      weekday: "short",
      day: "numeric",
      month: "short",
      year: "numeric",
    });
  };

  // Get favicon URL from website URL
  const getFaviconUrl = (websiteUrl: string | null | undefined) => {
    if (!websiteUrl) return null;

    try {
      const url = new URL(websiteUrl);
      return `https://www.google.com/s2/favicons?domain=${url.hostname}&sz=64`;
    } catch (e) {
      return null;
    }
  };

  // Function to determine risk badge style
  const getRiskBadgeStyle = (riskLevel: string | undefined) => {
    switch (riskLevel) {
      case "CRITICAL":
        return "bg-[#FFEFEF] text-[#CD2B31]";
      case "HIGH":
        return "bg-[#FFF1E7] text-[#BD4B00]";
      case "MEDIUM":
        return "bg-[#FFF8E7] text-[#B54708]";
      case "LOW":
        return "bg-[#EEFADC] text-[#5D770D]";
      case "NONE":
        return "bg-[#E6F4FF] text-[#0B4F71]";
      default:
        return "bg-gray-100 text-gray-600";
    }
  };

  // Get risk assessment status
  const getRiskAssessmentStatus = (vendor: any) => {
    const latestAssessment = vendor.riskAssessments?.edges[0]?.node;

    if (!latestAssessment) {
      return {
        status: "NEEDED",
        dataSensitivity: null,
        businessImpact: null,
        isExpired: false,
        assessedAt: null,
      };
    }

    const now = new Date();
    const expiresAt = new Date(latestAssessment.expiresAt);
    const isExpired = expiresAt < now;

    return {
      status: isExpired ? "EXPIRED" : "VALID",
      dataSensitivity: latestAssessment.dataSensitivity,
      businessImpact: latestAssessment.businessImpact,
      isExpired,
      assessedAt: latestAssessment.assessedAt,
    };
  };

  return (
    <PageTemplate title="Vendors">
      <div className="mb-6 p-4 border rounded-xl bg-white relative">
        <div className="flex items-center gap-2 mb-4">
          <Store className="h-5 w-5" />
          <h3 className="font-medium">Add a vendor</h3>
        </div>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            if (searchTerm.trim()) {
              createVendor({
                variables: {
                  connections: [vendorsConnection.vendors.__id],
                  input: {
                    organizationId: data.organization.id,
                    name: searchTerm.trim(),
                    description: "",
                  },
                },
                onCompleted() {
                  setSearchTerm("");
                  setFilteredVendors([]);
                  toast({
                    title: "Vendor created",
                    description: "The new vendor has been created successfully",
                  });
                  inputRef.current?.focus();
                },
              });
            }
          }}
        >
          <div className="flex gap-2 relative">
            <Input
              type="text"
              placeholder="Type vendor's name"
              value={searchTerm}
              style={{ borderRadius: "0.3rem" }}
              onChange={handleSearchChange}
              disabled={isLoadingVendors}
              ref={inputRef}
            />
            {isLoadingVendors && (
              <div className="absolute inset-0 flex items-center justify-center bg-background/50">
                <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
              </div>
            )}

            {searchTerm.trim() !== "" && (
              <div
                style={{ borderRadius: "0.3rem" }}
                className="absolute top-full left-0 mt-1 w-[calc(100%-100px)] max-h-48 overflow-y-auto border bg-white shadow-md z-10"
              >
                {filteredVendors.map((vendor) => (
                  <button
                    key={vendor.name}
                    className="w-full px-3 py-2 text-left hover:bg-gray-50"
                    onClick={() => {
                      createVendor({
                        variables: {
                          connections: [vendorsConnection.vendors.__id],
                          input: {
                            organizationId: data.organization.id,
                            name: vendor.name,
                            description: vendor.description,
                            headquarterAddress: vendor.headquarterAddress,
                            legalName: vendor.legalName,
                            websiteUrl: vendor.websiteUrl,
                            category: vendor.category as VendorCategory || null,
                            privacyPolicyUrl: vendor.privacyPolicyUrl,
                            serviceLevelAgreementUrl: vendor.serviceLevelAgreementUrl,
                            dataProcessingAgreementUrl: vendor.dataProcessingAgreementUrl,
                            businessAssociateAgreementUrl: vendor.businessAssociateAgreementUrl,
                            subprocessorsListUrl: vendor.subprocessorsListUrl,
                            certifications: vendor.certifications,
                            securityPageUrl: vendor.securityPageUrl,
                            trustPageUrl: vendor.trustPageUrl,
                            statusPageUrl: vendor.statusPageUrl,
                            termsOfServiceUrl: vendor.termsOfServiceUrl,
                          },
                        },
                        onCompleted() {
                          setSearchTerm("");
                          setFilteredVendors([]);
                          toast({
                            title: "Vendor added",
                            description:
                              "The vendor has been added successfully",
                          });
                          inputRef.current?.focus();
                        },
                        onError: (error) => {
                          toast({
                            title: "Error",
                            description: error.message || "Failed to create vendor",
                            variant: "destructive",
                          });
                        }
                      });
                    }}
                  >
                    <div className="flex items-center gap-2">
                      {vendor.websiteUrl && (
                        <img
                          src={getFaviconUrl(vendor.websiteUrl) || ""}
                          alt=""
                          className="w-4 h-4"
                          onError={(e) => {
                            e.currentTarget.style.display = "none";
                          }}
                        />
                      )}
                      {vendor.name}
                    </div>
                  </button>
                ))}
                <button
                  className="w-full px-3 py-2 text-left hover:bg-gray-50 flex items-center gap-2 border-t"
                  onClick={(e) => {
                    e.preventDefault();
                    if (searchTerm.trim()) {
                      createVendor({
                        variables: {
                          connections: [vendorsConnection.vendors.__id],
                          input: {
                            organizationId: data.organization.id,
                            name: searchTerm.trim(),
                            description: "",
                          },
                        },
                        onCompleted() {
                          setSearchTerm("");
                          setFilteredVendors([]);
                          toast({
                            title: "Vendor created",
                            description: "The new vendor has been created successfully",
                          });
                          inputRef.current?.focus();
                        },
                      });
                    }
                  }}
                >
                  <div className="flex items-center gap-2">
                    <Plus className="h-4 w-4" />
                    <span className="font-medium">Create new vendor:</span>{" "}
                    {searchTerm}
                  </div>
                </button>
              </div>
            )}
          </div>
        </form>
      </div>

      <div className="rounded-xl border overflow-hidden">
        <table className="w-full bg-white">
          <thead className="bg-white border-b border-gray-100">
            <tr>
              <th className="py-3 px-4 text-left text-xs font-semibold text-gray-500">
                <div className="flex items-center gap-2">
                  Vendor
                  <svg
                    width="8"
                    height="8"
                    viewBox="0 0 8 8"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                    className="text-gray-400"
                  >
                    <path d="M4 6L7 3H1L4 6Z" fill="currentColor" />
                  </svg>
                </div>
              </th>
              <th className="py-3 px-4 text-left text-xs font-semibold text-gray-500">
                <div className="flex items-center gap-2">
                  Data Risk
                  <svg
                    width="8"
                    height="8"
                    viewBox="0 0 8 8"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                    className="text-gray-400"
                  >
                    <path d="M4 6L7 3H1L4 6Z" fill="currentColor" />
                  </svg>
                </div>
              </th>
              <th className="py-3 px-4 text-left text-xs font-semibold text-gray-500">
                <div className="flex items-center gap-2">
                  Business Risk
                  <svg
                    width="8"
                    height="8"
                    viewBox="0 0 8 8"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                    className="text-gray-400"
                  >
                    <path d="M4 6L7 3H1L4 6Z" fill="currentColor" />
                  </svg>
                </div>
              </th>
              <th className="py-3 px-4 text-left text-xs font-semibold text-gray-500">
                <div className="flex items-center gap-2">
                  Accessed At
                  <svg
                    width="8"
                    height="8"
                    viewBox="0 0 8 8"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                    className="text-gray-400"
                  >
                    <path d="M4 6L7 3H1L4 6Z" fill="currentColor" />
                  </svg>
                </div>
              </th>
              <th className="py-3 px-4"></th>
            </tr>
          </thead>
          <tbody>
            {vendors.map((vendor) => {
              const riskStatus = getRiskAssessmentStatus(vendor);
              return (
                <tr
                  key={vendor.id}
                  className="border-b border-gray-100 hover:bg-gray-50 transition-colors bg-white"
                >
                  <td className="py-4 px-4">
                    <div className="flex items-center gap-3">
                      <Avatar className="h-8 w-8 bg-white">
                        {(() => {
                          const faviconUrl = vendor.websiteUrl
                            ? getFaviconUrl(vendor.websiteUrl)
                            : null;
                          return faviconUrl ? (
                            <AvatarImage src={faviconUrl} alt={vendor.name} />
                          ) : (
                            <AvatarFallback>{vendor.name?.[0]}</AvatarFallback>
                          );
                        })()}
                      </Avatar>
                      <Link
                        to={`/organizations/${organizationId}/vendors/${vendor.id}`}
                        className="font-medium hover:underline"
                      >
                        {vendor.name}
                      </Link>
                    </div>
                  </td>
                  <td className="py-4 px-4">
                    {riskStatus.status === "NEEDED" ? (
                      <Badge className="bg-gray-100 text-gray-600 font-medium text-xs px-2 py-1 rounded-md">
                        Assessment needed
                      </Badge>
                    ) : riskStatus.status === "EXPIRED" ? (
                      <Badge className="bg-[#FFEFEF] text-[#CD2B31] font-medium text-xs px-2 py-1 rounded-md">
                        Assessment expired
                      </Badge>
                    ) : (
                      <Badge
                        className={`${getRiskBadgeStyle(riskStatus.dataSensitivity)} font-medium text-xs px-2 py-1 rounded-md`}
                      >
                        {riskStatus.dataSensitivity || "NONE"}
                      </Badge>
                    )}
                  </td>
                  <td className="py-4 px-4">
                    {riskStatus.status === "NEEDED" ? (
                      <Badge className="bg-gray-100 text-gray-600 font-medium text-xs px-2 py-1 rounded-md">
                        Assessment needed
                      </Badge>
                    ) : riskStatus.status === "EXPIRED" ? (
                      <Badge className="bg-[#FFEFEF] text-[#CD2B31] font-medium text-xs px-2 py-1 rounded-md">
                        Assessment expired
                      </Badge>
                    ) : (
                      <Badge
                        className={`${getRiskBadgeStyle(riskStatus.businessImpact)} font-medium text-xs px-2 py-1 rounded-md`}
                      >
                        {riskStatus.businessImpact || "LOW"}
                      </Badge>
                    )}
                  </td>
                  <td className="py-4 px-4">
                    {riskStatus.assessedAt
                      ? formatDate(riskStatus.assessedAt)
                      : "N/A"}
                  </td>
                  <td className="py-4 px-4 text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 rounded-full"
                        >
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem
                          className="text-red-600"
                          onClick={(e) => {
                            e.preventDefault();
                            if (
                              window.confirm(
                                "Are you sure you want to delete this vendor?",
                              )
                            ) {
                              deleteVendor({
                                variables: {
                                  connections: [vendorsConnection.vendors.__id],
                                  input: {
                                    vendorId: vendor.id,
                                  },
                                },
                                onCompleted() {
                                  toast({
                                    title: "Vendor deleted",
                                    description:
                                      "The vendor has been deleted successfully",
                                  });
                                },
                              });
                            }
                          }}
                        >
                          Delete vendor
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      <LoadAboveButton
        isLoading={isLoadingPrevious}
        hasMore={hasPrevious}
        onLoadMore={() => {
          startTransition(() => {
            setSearchParams((prev) => {
              prev.set("before", pageInfo?.startCursor || "");
              prev.delete("after");
              return prev;
            });
            loadPrevious(ITEMS_PER_PAGE);
          });
        }}
      />
      <LoadBelowButton
        isLoading={isLoadingNext}
        hasMore={hasNext}
        onLoadMore={() => {
          startTransition(() => {
            setSearchParams((prev) => {
              prev.set("after", pageInfo?.endCursor || "");
              prev.delete("before");
              return prev;
            });
            loadNext(ITEMS_PER_PAGE);
          });
        }}
      />
    </PageTemplate>
  );
}

export default function ListVendorView() {
  const [searchParams] = useSearchParams();
  const [queryRef, loadQuery] =
    useQueryLoader<ListVendorViewQuery>(listVendorViewQuery);

  const { organizationId } = useParams();

  useEffect(() => {
    const after = searchParams.get("after");
    const before = searchParams.get("before");

    loadQuery({
      organizationId: organizationId!,
      first: before ? undefined : ITEMS_PER_PAGE,
      after: after || undefined,
      last: before ? ITEMS_PER_PAGE : undefined,
      before: before || undefined,
    });
  }, [loadQuery, organizationId]);

  if (!queryRef) {
    return <ListVendorViewSkeleton />;
  }

  return (
    <Suspense fallback={<ListVendorViewSkeleton />}>
      <ListVendorContent queryRef={queryRef} />
    </Suspense>
  );
}
