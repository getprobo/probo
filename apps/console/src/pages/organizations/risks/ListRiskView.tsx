"use client";

import {
  graphql,
  PreloadedQuery,
  usePaginationFragment,
  usePreloadedQuery,
  useQueryLoader,
  useMutation,
} from "react-relay";
import {
  Suspense,
  useCallback,
  useEffect,
  useState,
  useTransition,
} from "react";
import type { ListRiskViewQuery } from "./__generated__/ListRiskViewQuery.graphql";
import type { RiskTreatment } from "./__generated__/ListRiskView_risks.graphql";
import { useParams, useSearchParams } from "react-router";
import { PageTemplate } from "@/components/PageTemplate";
import { RiskViewSkeleton } from "./ListRiskPage";
import { ListRiskViewPaginationQuery } from "./__generated__/ListRiskViewPaginationQuery.graphql";
import { ListRiskView_risks$key } from "./__generated__/ListRiskView_risks.graphql";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Link } from "react-router";
import { Plus, Trash2, Edit } from "lucide-react";
import { useToast } from "@/hooks/use-toast";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { ListRiskViewDeleteMutation } from "./__generated__/ListRiskViewDeleteMutation.graphql";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";

const defaultPageSize = 25;

// Define available risk categories
const RISK_CATEGORIES = [
  "Compliance & Legal",
  "Cybersecurity",
  "Finance",
  "Human capital",
  "Operations",
  "Reputational",
  "Strategic",
  "Technology",
  "Other"
] as const;

const listRiskViewQuery = graphql`
  query ListRiskViewQuery(
    $organizationId: ID!
    $first: Int
    $after: CursorKey
    $last: Int
    $before: CursorKey
  ) {
    organization: node(id: $organizationId) {
      id

      ...ListRiskView_risks
        @arguments(first: $first, after: $after, last: $last, before: $before)
    }
  }
`;

const listRiskViewFragment = graphql`
  fragment ListRiskView_risks on Organization
  @refetchable(queryName: "ListRiskViewPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int" }
    after: { type: "CursorKey" }
    last: { type: "Int" }
    before: { type: "CursorKey" }
  ) {
    risks(first: $first, after: $after, last: $last, before: $before)
      @connection(key: "ListRiskView_risks") {
      __id
      edges {
        node {
          id
          name
          inherentLikelihood
          inherentImpact
          residualLikelihood
          residualImpact
          treatment
          description
          category
          createdAt
          updatedAt
          owner {
            id
            fullName
          }
          measures(first: 1) {
            edges {
              node {
                category
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

const deleteRiskMutation = graphql`
  mutation ListRiskViewDeleteMutation(
    $input: DeleteRiskInput!
    $connections: [ID!]!
  ) {
    deleteRisk(input: $input) {
      deletedRiskId @deleteEdge(connections: $connections)
    }
  }
`;

// Helper function to convert risk score to severity label
const riskScoreToSeverity = (score: number): string => {
  if (score >= 20) return "Catastrophic";
  if (score >= 12) return "Critical";
  if (score >= 5) return "Marginal";
  return "Negligible";
};

// Helper function to format treatment value
const formatTreatment = (treatment: string): string => {
  const treatmentMap: Record<string, string> = {
    MITIGATED: "Mitigate",
    ACCEPTED: "Accept",
    AVOIDED: "Avoid",
    TRANSFERRED: "Transfer",
  };

  return treatmentMap[treatment] || treatment;
};

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
    <div className="flex justify-center mb-6">
      <Button
        variant="outline"
        onClick={onLoadMore}
        disabled={isLoading}
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
    <div className="flex justify-center mt-6">
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

// Define colors for risk matrix cells
const riskMatrixColors = {
  low: "bg-green-500 text-white",
  medium: "bg-yellow-300 text-black",
  high: "bg-red-500 text-white",
};

// Empty cell variants (lighter colors)
const emptyRiskMatrixColors = {
  low: "bg-green-50 text-black",
  medium: "bg-yellow-50 text-black",
  high: "bg-red-50 text-black",
};

type RiskNode = {
  readonly id: string;
  readonly name: string;
  readonly inherentLikelihood: number;
  readonly inherentImpact: number;
  readonly residualLikelihood: number;
  readonly residualImpact: number;
  readonly treatment: RiskTreatment;
  readonly description: string;
  readonly category: string;
  readonly createdAt: string;
  readonly updatedAt: string;
  readonly owner: {
    readonly id: string;
    readonly fullName: string;
  } | null;
  readonly measures: {
    readonly edges: ReadonlyArray<{
      readonly node: {
        readonly category: string;
      };
    }>;
  };
};

// Risk Matrix Component
function RiskMatrix({
  risks,
  isResidual = false,
  organizationId,
}: {
  risks: Array<{
    id: string;
    name: string;
    inherentLikelihood: number;
    inherentImpact: number;
    residualLikelihood?: number;
    residualImpact?: number;
  }>;
  isResidual?: boolean;
  organizationId: string;
}): JSX.Element {
  // Define impact values for the vertical axis (rows) - from highest (5) to lowest (1)
  const impactValues: number[] = [5, 4, 3, 2, 1];

  // Define likelihood values for the horizontal axis (columns) - from lowest (1) to highest (5)
  const likelihoodValues: number[] = [1, 2, 3, 4, 5];

  const impactLabels = [
    "Catastrophic",
    "Significant",
    "Moderate",
    "Low",
    "Negligible",
  ];

  const likelihoodLabels = [
    "Improbable",
    "Remote",
    "Occasional",
    "Probable",
    "Frequent",
  ];

  // Function to get cell content with risks that fall in this cell
  const getCellContent = (impactValue: number, likelihoodValue: number) => {
    return risks.filter((risk) => {
      const likelihood = isResidual
        ? risk.residualLikelihood ?? risk.inherentLikelihood
        : risk.inherentLikelihood;
      const impact = isResidual
        ? risk.residualImpact ?? risk.inherentImpact
        : risk.inherentImpact;

      return likelihood === likelihoodValue && impact === impactValue;
    });
  };

  // Helper to determine cell color based on position in matrix
  // Matrix has rows indexed from top to bottom (0 = highest impact, 4 = lowest impact)
  // and columns indexed from left to right (0 = lowest likelihood, 4 = highest likelihood)
  const getCellColor = (row: number, col: number, isEmpty: boolean): string => {
    const colorSet = isEmpty ? emptyRiskMatrixColors : riskMatrixColors;

    // Hard-coded color matrix where each element [row][col] represents a specific cell
    // This follows the 5x5 grid with rows representing impact (5 to 1) and columns representing likelihood (1 to 5)
    const colorMatrix = [
      // Impact 5 (Catastrophic) - first row
      [
        colorSet.medium,
        colorSet.medium,
        colorSet.high,
        colorSet.high,
        colorSet.high,
      ],
      // Impact 4 (Significant) - second row
      [
        colorSet.low,
        colorSet.medium,
        colorSet.medium,
        colorSet.high,
        colorSet.high,
      ],
      // Impact 3 (Moderate) - third row
      [
        colorSet.low,
        colorSet.low,
        colorSet.medium,
        colorSet.medium,
        colorSet.high,
      ],
      // Impact 2 (Low) - fourth row
      [
        colorSet.low,
        colorSet.low,
        colorSet.low,
        colorSet.medium,
        colorSet.medium,
      ],
      // Impact 1 (Negligible) - fifth row
      [colorSet.low, colorSet.low, colorSet.low, colorSet.low, colorSet.medium],
    ];

    return colorMatrix[row][col];
  };

  // Instead of using refs and tippy, we'll use a component with popover
  const RiskCell = ({
    rowIndex,
    colIndex,
    impactValue,
    likelihoodValue,
    organizationId,
  }: {
    rowIndex: number;
    colIndex: number;
    impactValue: number;
    likelihoodValue: number;
    organizationId: string;
  }) => {
    const cellRisks = getCellContent(impactValue, likelihoodValue);
    const isEmpty = cellRisks.length === 0;
    const cellColor = getCellColor(rowIndex, colIndex, isEmpty);

    if (isEmpty) {
      return (
        <td
          className={`border aspect-square w-14 h-14 text-center ${cellColor}`}
          data-risks={0}
        >
          <div className="text-sm font-bold flex items-center justify-center h-full"></div>
        </td>
      );
    }

    return (
      <Popover>
        <PopoverTrigger asChild>
          <td
            className={`border aspect-square w-14 h-14 text-center ${cellColor} cursor-pointer hover:opacity-90`}
            data-risks={cellRisks.length}
          >
            <div className="text-sm font-bold flex items-center justify-center h-full">
              {cellRisks.length}
            </div>
          </td>
        </PopoverTrigger>
        <PopoverContent className="w-72 p-4" align="center">
          <div className="font-semibold text-lg mb-2">
            {cellRisks.length} Risk{cellRisks.length > 1 ? "s" : ""}
          </div>

          <div className="bg-gray-50 p-3 rounded-md mb-3 space-y-1">
            <div className="grid grid-cols-2 gap-2">
              <div className="text-sm">
                <span className="text-muted-foreground">Impact:</span>{" "}
                <span className="font-medium">{impactValue}</span>
              </div>
              <div className="text-sm">
                <span className="text-muted-foreground">Likelihood:</span>{" "}
                <span className="font-medium">{likelihoodValue}</span>
              </div>
            </div>

            <div className="text-sm pt-1 border-t border-gray-200 mt-1">
              <span className="text-muted-foreground">Risk Score:</span>
              <span className="font-bold ml-1">
                {impactValue * likelihoodValue}
              </span>
              <span
                className="ml-2 px-2 py-0.5 text-xs rounded-full font-medium inline-block"
                style={{
                  backgroundColor:
                    impactValue * likelihoodValue >= 20
                      ? "#ef4444"
                      : impactValue * likelihoodValue >= 12
                      ? "#f59e0b"
                      : impactValue * likelihoodValue >= 5
                      ? "#10b981"
                      : "#94a3b8",
                  color:
                    impactValue * likelihoodValue >= 12 ? "white" : "inherit",
                }}
              >
                {impactValue * likelihoodValue >= 20
                  ? "Catastrophic"
                  : impactValue * likelihoodValue >= 12
                  ? "Critical"
                  : impactValue * likelihoodValue >= 5
                  ? "Marginal"
                  : "Negligible"}
              </span>
            </div>
          </div>

          {cellRisks.length > 0 && (
            <>
              <div className="text-sm font-medium mb-1">Items:</div>
              <ul className="space-y-1.5 max-h-48 overflow-y-auto">
                {cellRisks.map((risk) => (
                  <li
                    key={risk.id}
                    className="text-sm border-l-2 pl-2"
                    style={{
                      borderColor:
                        (isResidual
                          ? (risk.residualImpact || risk.inherentImpact) *
                            (risk.residualLikelihood || risk.inherentLikelihood)
                          : risk.inherentImpact * risk.inherentLikelihood) >= 20
                          ? "#ef4444"
                          : (isResidual
                              ? (risk.residualImpact || risk.inherentImpact) *
                                (risk.residualLikelihood ||
                                  risk.inherentLikelihood)
                              : risk.inherentImpact *
                                risk.inherentLikelihood) >= 12
                          ? "#f59e0b"
                          : (isResidual
                              ? (risk.residualImpact || risk.inherentImpact) *
                                (risk.residualLikelihood ||
                                  risk.inherentLikelihood)
                              : risk.inherentImpact *
                                risk.inherentLikelihood) >= 5
                          ? "#10b981"
                          : "#94a3b8",
                    }}
                  >
                    <Link
                      to={`/organizations/${organizationId}/risks/${risk.id}`}
                      className="hover:underline hover:text-primary transition-colors flex items-center w-full py-1 rounded hover:bg-gray-100 px-1"
                    >
                      <span className="truncate">{risk.name}</span>
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        className="w-3.5 h-3.5 ml-auto flex-shrink-0 text-muted-foreground"
                      >
                        <polyline points="9 18 15 12 9 6" />
                      </svg>
                    </Link>
                  </li>
                ))}
              </ul>
            </>
          )}
        </PopoverContent>
      </Popover>
    );
  };

  return (
    <div className="space-y-2">
      <div className="overflow-x-auto">
        <div className="flex flex-col">
          <div className="flex">
            <div
              className="text-xs font-semibold flex items-center justify-center mr-2"
              style={{
                writingMode: "vertical-rl",
                transform: "rotate(180deg)",
                alignSelf: "center",
              }}
            >
              Impact
            </div>
            <table className="w-full border-collapse table-fixed">
              <tbody>
                {impactValues.map((impactValue, rowIndex) => (
                  <tr key={rowIndex}>
                    <th className="p-1 text-xs text-right border-0 font-medium w-14 h-14">
                      {impactLabels[rowIndex]} ({impactValue})
                    </th>
                    {likelihoodValues.map((likelihoodValue, colIndex) => (
                      <RiskCell
                        key={colIndex}
                        rowIndex={rowIndex}
                        colIndex={colIndex}
                        impactValue={impactValue}
                        likelihoodValue={likelihoodValue}
                        organizationId={organizationId}
                      />
                    ))}
                  </tr>
                ))}
              </tbody>
              <tfoot>
                <tr>
                  <th className="p-1 text-center border-0 w-14"></th>
                  {likelihoodLabels.map((label, index) => (
                    <th
                      key={index}
                      className="p-1 text-xs text-center border-t font-medium w-14"
                    >
                      {label} ({likelihoodValues[index]})
                    </th>
                  ))}
                </tr>
              </tfoot>
            </table>
          </div>
          <div className="flex">
            <div style={{ width: "3.5rem" }}></div>
            <div className="text-center text-xs font-semibold mt-2 flex-1">
              Likelihood
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function ListRiskViewContent({
  queryRef,
}: {
  queryRef: PreloadedQuery<ListRiskViewQuery>;
}) {
  const data = usePreloadedQuery(listRiskViewQuery, queryRef);
  const [, setSearchParams] = useSearchParams();
  const [, startTransition] = useTransition();
  const { organizationId } = useParams<{ organizationId: string }>();
  const { toast } = useToast();

  // State for delete confirmation dialog
  const [riskToDelete, setRiskToDelete] = useState<{
    id: string;
    name: string;
  } | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  // State for toggling between initial and residual risk matrix
  const [showResidualRisk, setShowResidualRisk] = useState(false);

  // State for category filters
  const [selectedCategories, setSelectedCategories] = useState<Set<string>>(new Set());

  // Setup delete mutation
  const [commitDeleteMutation] =
    useMutation<ListRiskViewDeleteMutation>(deleteRiskMutation);

  const {
    data: risksConnection,
    loadNext,
    loadPrevious,
    hasNext,
    hasPrevious,
    isLoadingNext,
    isLoadingPrevious,
  } = usePaginationFragment<
    ListRiskViewPaginationQuery,
    ListRiskView_risks$key
  >(listRiskViewFragment, data.organization);

  const risks = risksConnection?.risks?.edges?.map((edge) => edge.node) || [];
  const pageInfo = risksConnection?.risks?.pageInfo;
  const connectionId = risksConnection?.risks?.__id;

  // Get unique categories from existing risks
  const availableCategories = Array.from(new Set(risks.map(risk => risk.category))).sort();

  // Filter risks based on selected categories
  const filteredRisks = selectedCategories.size === 0 
    ? risks 
    : risks.filter(risk => selectedCategories.has(risk.category));

  // Handle category toggle
  const toggleCategory = (category: string) => {
    setSelectedCategories(prev => {
      const newSet = new Set(prev);
      if (newSet.has(category)) {
        newSet.delete(category);
      } else {
        newSet.add(category);
      }
      return newSet;
    });
  };

  // Handle delete risk
  const handleDeleteRisk = useCallback(() => {
    if (!riskToDelete || !connectionId) return;

    setIsDeleting(true);

    commitDeleteMutation({
      variables: {
        input: {
          riskId: riskToDelete.id,
        },
        connections: [connectionId],
      },
      onCompleted: (_, errors) => {
        setIsDeleting(false);
        setRiskToDelete(null);

        if (errors) {
          console.error("Error deleting risk:", errors);
          toast({
            title: "Error",
            description: "Failed to delete risk. Please try again.",
            variant: "destructive",
          });
          return;
        }

        toast({
          title: "Success",
          description: "Risk deleted successfully.",
        });
      },
      onError: (error) => {
        setIsDeleting(false);
        setRiskToDelete(null);
        console.error("Error deleting risk:", error);
        toast({
          title: "Error",
          description: "Failed to delete risk. Please try again.",
          variant: "destructive",
        });
      },
    });
  }, [riskToDelete, connectionId, commitDeleteMutation, toast]);

  return (
    <PageTemplate
      title="Risks"
      actions={
        <Button asChild>
          <Link to={`/organizations/${organizationId}/risks/new`}>
            <Plus className="mr-2 h-4 w-4" />
            New Risk
          </Link>
        </Button>
      }
    >
      <div className="space-y-6">
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
              loadPrevious(defaultPageSize);
            });
          }}
        />

        {/* Combined Risk Matrix with Toggle */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex flex-col w-full">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-base font-semibold">
                  {showResidualRisk ? "Residual risk" : "Current risk"}
                </h3>
                <div className="flex items-center space-x-4">
                  <Label
                    htmlFor="risk-toggle"
                    className={`text-sm font-semibold cursor-pointer ${
                      !showResidualRisk
                        ? "text-primary"
                        : "text-muted-foreground"
                    }`}
                  >
                    Initial
                  </Label>
                  <Switch
                    id="risk-toggle"
                    checked={showResidualRisk}
                    onCheckedChange={setShowResidualRisk}
                    className="bg-white border-2 border-gray-300 data-[state=checked]:bg-white data-[state=checked]:border-gray-300 [&_span]:bg-gray-300"
                  />
                  <Label
                    htmlFor="risk-toggle"
                    className={`text-sm font-semibold cursor-pointer ${
                      showResidualRisk
                        ? "text-primary"
                        : "text-muted-foreground"
                    }`}
                  >
                    Residual
                  </Label>
                </div>
              </div>
              <RiskMatrix
                risks={filteredRisks}
                isResidual={showResidualRisk}
                organizationId={organizationId || ""}
              />
            </div>
          </CardContent>
        </Card>

        {/* Category Filters */}
        <div className="flex flex-wrap gap-2">
          {availableCategories.map((category) => (
            <Badge
              key={category}
              variant={selectedCategories.has(category) ? "default" : "outline"}
              className="cursor-pointer"
              onClick={() => toggleCategory(category)}
            >
              {category}
            </Badge>
          ))}
        </div>

        <Card>
          <CardContent className="p-0">
            <div className="w-full overflow-auto">
              <table className="w-full caption-bottom text-sm">
                <thead className="[&_tr]:border-b">
                  <tr className="border-b transition-colors hover:bg-h-subtle-bg data-[state=selected]:bg-subtle-bg">
                    <th className="h-12 px-4 text-left align-middle font-medium text-tertiary w-1/6">
                      Category
                    </th>
                    <th className="h-12 px-4 text-left align-middle font-medium text-tertiary w-1/3">
                      Name
                    </th>
                    <th className="h-12 px-4 text-left align-middle font-medium text-tertiary w-1/6">
                      Inherent
                    </th>
                    <th className="h-12 px-4 text-left align-middle font-medium text-tertiary w-1/6">
                      Treatment
                    </th>
                    <th className="h-12 px-4 text-left align-middle font-medium text-tertiary w-1/6">
                      Residual
                    </th>
                    <th className="h-12 px-4 text-left align-middle font-medium text-tertiary w-1/6">
                      Owner
                    </th>
                    <th className="h-12 px-4 text-left align-middle font-medium text-tertiary w-[120px]">
                      Action
                    </th>
                  </tr>
                </thead>
                <tbody className="[&_tr:last-child]:border-0">
                  {filteredRisks.length === 0 ? (
                    <tr className="border-b transition-colors hover:bg-h-subtle-bg data-[state=selected]:bg-subtle-bg">
                      <td
                        colSpan={7}
                        className="text-center p-4 align-middle text-tertiary"
                      >
                        No risks found. Create a new risk to get started.
                      </td>
                    </tr>
                  ) : (
                    filteredRisks.map((risk) => (
                      <tr
                        key={risk.id}
                        className="border-b transition-colors hover:bg-h-subtle-bg data-[state=selected]:bg-subtle-bg cursor-pointer"
                      >
                        <td className="p-0 align-middle w-1/6">
                          <Link
                            to={`/organizations/${organizationId}/risks/${risk.id}`}
                            className="block p-4 h-full w-full"
                          >
                            {risk.category}
                          </Link>
                        </td>
                        <td className="p-0 align-middle font-medium w-1/3">
                          <Link
                            to={`/organizations/${organizationId}/risks/${risk.id}`}
                            className="block p-4 h-full w-full"
                          >
                            {risk.name}
                          </Link>
                        </td>
                        <td className="p-0 align-middle w-1/6 whitespace-nowrap">
                          <Link
                            to={`/organizations/${organizationId}/risks/${risk.id}`}
                            className="block p-4 h-full w-full"
                          >
                            <span
                              className="px-2 py-0.5 text-xs rounded-full font-medium inline-block"
                              style={{
                                backgroundColor:
                                  risk.inherentLikelihood * risk.inherentImpact >= 20
                                    ? "#ef4444"
                                    : risk.inherentLikelihood * risk.inherentImpact >= 12
                                    ? "#f59e0b"
                                    : risk.inherentLikelihood * risk.inherentImpact >= 5
                                    ? "#10b981"
                                    : "#94a3b8",
                                color:
                                  risk.inherentLikelihood * risk.inherentImpact >= 12
                                    ? "white"
                                    : "inherit",
                              }}
                            >
                              {riskScoreToSeverity(
                                risk.inherentLikelihood * risk.inherentImpact
                              )}{" "}
                              ({risk.inherentLikelihood * risk.inherentImpact})
                            </span>
                          </Link>
                        </td>
                        <td className="p-0 align-middle w-1/6 whitespace-nowrap">
                          <Link
                            to={`/organizations/${organizationId}/risks/${risk.id}`}
                            className="block p-4 h-full w-full"
                          >
                            {formatTreatment(risk.treatment)}
                          </Link>
                        </td>
                        <td className="p-0 align-middle w-1/6 whitespace-nowrap">
                          <Link
                            to={`/organizations/${organizationId}/risks/${risk.id}`}
                            className="block p-4 h-full w-full"
                          >
                            {risk.residualLikelihood && risk.residualImpact ? (
                              <span
                                className="px-2 py-0.5 text-xs rounded-full font-medium inline-block"
                                style={{
                                  backgroundColor:
                                    risk.residualLikelihood * risk.residualImpact >= 20
                                      ? "#ef4444"
                                      : risk.residualLikelihood * risk.residualImpact >= 12
                                      ? "#f59e0b"
                                      : risk.residualLikelihood * risk.residualImpact >= 5
                                      ? "#10b981"
                                      : "#94a3b8",
                                  color:
                                    risk.residualLikelihood * risk.residualImpact >= 12
                                      ? "white"
                                      : "inherit",
                                }}
                              >
                                {riskScoreToSeverity(
                                  risk.residualLikelihood * risk.residualImpact
                                )}{" "}
                                ({risk.residualLikelihood * risk.residualImpact})
                              </span>
                            ) : (
                              "Not set"
                            )}
                          </Link>
                        </td>
                        <td className="p-0 align-middle w-1/6">
                          <Link
                            to={`/organizations/${organizationId}/risks/${risk.id}`}
                            className="block p-4 h-full w-full"
                          >
                            {risk.owner?.fullName || "Unassigned"}
                          </Link>
                        </td>
                        <td className="p-4 align-middle w-[120px]">
                          <div className="flex">
                            <Button
                              variant="ghost"
                              size="icon"
                              asChild
                              className="mr-1"
                            >
                              <Link
                                to={`/organizations/${organizationId}/risks/${risk.id}/edit`}
                              >
                                <Edit className="h-4 w-4 text-tertiary" />
                              </Link>
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => {
                                setRiskToDelete({
                                  id: risk.id,
                                  name: risk.name,
                                });
                              }}
                            >
                              <Trash2 className="h-4 w-4 text-danger" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>

        {/* Delete Confirmation Dialog */}
        <Dialog
          open={!!riskToDelete}
          onOpenChange={(open) => !open && setRiskToDelete(null)}
        >
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Are you sure?</DialogTitle>
              <DialogDescription>
                This will permanently delete the risk &quot;{riskToDelete?.name}
                &quot;. This action cannot be undone.
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button
                variant="outline"
                onClick={() => setRiskToDelete(null)}
                disabled={isDeleting}
              >
                Cancel
              </Button>
              <Button
                onClick={handleDeleteRisk}
                disabled={isDeleting}
                variant="destructive"
              >
                {isDeleting ? "Deleting..." : "Delete"}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>

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
              loadNext(defaultPageSize);
            });
          }}
        />
      </div>
    </PageTemplate>
  );
}

export default function ListRiskView() {
  const [searchParams] = useSearchParams();
  const [queryRef, loadQuery] =
    useQueryLoader<ListRiskViewQuery>(listRiskViewQuery);

  const { organizationId } = useParams();

  useEffect(() => {
    const after = searchParams.get("after");
    const before = searchParams.get("before");

    loadQuery({
      organizationId: organizationId!,
      first: before ? undefined : defaultPageSize,
      after: after || undefined,
      last: before ? defaultPageSize : undefined,
      before: before || undefined,
    });
  }, [loadQuery, organizationId, searchParams]);

  if (!queryRef) {
    return <RiskViewSkeleton />;
  }

  return (
    <Suspense fallback={<RiskViewSkeleton />}>
      <ListRiskViewContent queryRef={queryRef} />
    </Suspense>
  );
}
