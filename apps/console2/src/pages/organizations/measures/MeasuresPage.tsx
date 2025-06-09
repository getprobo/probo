import type { MeasureGraphListQuery } from "/hooks/graph/__generated__/MeasureGraphListQuery.graphql";
import {
  useFragment,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { useTranslate } from "@probo/i18n";
import {
  ActionDropdown,
  Button,
  Card,
  DropdownItem,
  FileButton,
  IconChevronDown,
  IconChevronUp,
  IconFolderUpload,
  IconPencil,
  IconPlusLarge,
  IconTrashCan,
  MeasureBadge,
  MeasureImplementation,
  PageHeader,
  Table,
  Tbody,
  Td,
  Th,
  Thead,
  Tr,
  useConfirm,
  useDialogRef,
} from "@probo/ui";
import {
  measuresQuery,
  useDeleteMeasureMutation,
} from "/hooks/graph/MeasureGraph";
import { graphql } from "relay-runtime";
import type {
  MeasuresPageFragment$data,
  MeasuresPageFragment$key,
} from "./__generated__/MeasuresPageFragment.graphql";
import { groupBy, objectKeys, slugify, sprintf } from "@probo/helpers";
import { useMemo, useRef, useState, type ChangeEventHandler } from "react";
import type { NodeOf } from "/types";
import type { MeasuresPageImportMutation } from "./__generated__/MeasuresPageImportMutation.graphql";
import { useMutationWithToasts } from "/hooks/useMutationWithToasts";
import { useOrganizationId } from "/hooks/useOrganizationId";
import { Link, useParams } from "react-router";
import MeasureFormDialog from "./dialog/MeasureFormDialog";
import { usePageTitle } from "@probo/hooks";

type Props = {
  queryRef: PreloadedQuery<MeasureGraphListQuery>;
};

const measuresFragment = graphql`
  fragment MeasuresPageFragment on Organization {
    measures(first: 100) @connection(key: "MeasuresGraphListQuery__measures") {
      __id
      edges {
        node {
          id
          name
          category
          state
          ...MeasureFormDialogMeasureFragment
        }
      }
    }
  }
`;

const importMeasuresMutation = graphql`
  mutation MeasuresPageImportMutation(
    $input: ImportMeasureInput!
    $connections: [ID!]!
  ) {
    importMeasure(input: $input) {
      measureEdges @appendEdge(connections: $connections) {
        node {
          id
          name
          category
          state
        }
      }
    }
  }
`;

export default function MeasuresPage(props: Props) {
  const { __ } = useTranslate();
  const organization = usePreloadedQuery(
    measuresQuery,
    props.queryRef
  ).organization;
  const data = useFragment<MeasuresPageFragment$key>(
    measuresFragment,
    organization
  );
  const connectionId = data.measures.__id;
  const measures = data.measures.edges.map((edge) => edge.node);
  const measuresPerCategory = useMemo(() => {
    return groupBy(measures, (measure) => measure.category);
  }, [measures]);
  const [importMeasures] = useMutationWithToasts<MeasuresPageImportMutation>(
    importMeasuresMutation,
    {
      successMessage: __("Measures imported successfully."),
      errorMessage: __("Failed to import measures. Please try again."),
    }
  );
  const importFileRef = useRef<HTMLInputElement>(null);
  usePageTitle(__("Measures"));

  const handleImport: ChangeEventHandler<HTMLInputElement> = (event) => {
    const file = event.target.files?.[0];
    if (!file) {
      return;
    }
    importMeasures({
      variables: {
        input: {
          organizationId: organization.id,
          file: null,
        },
        connections: [connectionId],
      },
      uploadables: {
        "input.file": file,
      },
      onCompleted() {
        importFileRef.current!.value = "";
      },
    });
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title={__("Measures")}
        description={__(
          "Measures are actions taken to reduce the risk. Add them to track their implementation status."
        )}
      >
        <FileButton
          ref={importFileRef}
          variant="secondary"
          icon={IconFolderUpload}
          onChange={handleImport}
        >
          {__("Import")}
        </FileButton>
        <MeasureFormDialog connection={connectionId}>
          <Button variant="primary" icon={IconPlusLarge}>
            {__("New measure")}
          </Button>
        </MeasureFormDialog>
      </PageHeader>
      <MeasureImplementation measures={measures} className="my-10" />
      {objectKeys(measuresPerCategory)
        .sort((a, b) => a.localeCompare(b))
        .map((category) => (
          <Category
            key={category}
            category={category}
            measures={measuresPerCategory[category]}
            connectionId={connectionId}
          />
        ))}
    </div>
  );
}

type CategoryProps = {
  category: string;
  measures: NodeOf<MeasuresPageFragment$data["measures"]>[];
  connectionId: string;
};

function Category(props: CategoryProps) {
  const params = useParams<{ categoryId?: string }>();
  const { __ } = useTranslate();
  const organizationId = useOrganizationId();
  const categoryId = slugify(props.category);
  const [limit, setLimit] = useState<number | null>(4);
  const measures = useMemo(() => {
    return limit ? props.measures.slice(0, limit) : props.measures;
  }, [props.measures, limit]);
  const showMoreButton = limit !== null && props.measures.length > limit;
  const isExpanded = categoryId === params.categoryId;
  const ExpandComponent = isExpanded ? IconChevronUp : IconChevronDown;
  const completedMeasures = props.measures.filter(
    (m) => m.state === "IMPLEMENTED"
  );

  return (
    <Card className="py-3 px-5">
      <Link
        to={`/organizations/${organizationId}/measures/category/${categoryId}`}
        className="flex items-center justify-between cursor-pointer"
      >
        <h2 className="text-base font-medium">{props.category}</h2>
        <div className="flex items-center gap-3 text-sm text-txt-secondary">
          <span>
            {__("Completion")}:{" "}
            <span className="text-txt-primary font-medium">
              {completedMeasures.length}/{props.measures.length}
            </span>
          </span>
          <span className="text-border-low">|</span>
          <ExpandComponent size={16} className="text-txt-secondary" />
        </div>
      </Link>
      {isExpanded && (
        <div className="mt-3">
          <Table className="bg-invert">
            <Thead>
              <Tr>
                <Th>{__("Measure")}</Th>
                <Th>{__("State")}</Th>
                <Th></Th>
              </Tr>
            </Thead>
            <Tbody>
              {measures.map((measure) => (
                <MeasureRow
                  key={measure.id}
                  measure={measure}
                  connectionId={props.connectionId}
                />
              ))}
            </Tbody>
          </Table>
          {showMoreButton && (
            <Button
              variant="tertiary"
              onClick={() => setLimit(null)}
              className="mt-3 mx-auto"
              icon={IconChevronDown}
            >
              {sprintf(__("Show %s more"), props.measures.length - limit)}
            </Button>
          )}
        </div>
      )}
    </Card>
  );
}

type MeasureRowProps = {
  measure: NodeOf<MeasuresPageFragment$data["measures"]>;
  connectionId: string;
};

function MeasureRow(props: MeasureRowProps) {
  const { __ } = useTranslate();
  const [deleteMeasure, isDeleting] = useDeleteMeasureMutation();
  const confirm = useConfirm();
  const organizationId = useOrganizationId();

  const onDelete = () => {
    confirm(
      () =>
        new Promise<void>((resolve) => {
          deleteMeasure({
            variables: {
              input: { measureId: props.measure.id },
              connections: [props.connectionId],
            },
            onCompleted: () => resolve(),
          });
        }),
      {
        message: sprintf(
          __(
            'This will permanently delete the measure "%s". This action cannot be undone.'
          ),
          props.measure.name
        ),
      }
    );
  };

  const dialogRef = useDialogRef();

  return (
    <>
      <MeasureFormDialog measure={props.measure} ref={dialogRef} />
      <Tr to={`/organizations/${organizationId}/measures/${props.measure.id}`}>
        <Td>{props.measure.name}</Td>
        <Td width={120}>
          <MeasureBadge state={props.measure.state} />
        </Td>
        <Td noLink width={50} className="text-end">
          <ActionDropdown>
            <DropdownItem
              icon={IconPencil}
              onClick={() => dialogRef.current?.open()}
            >
              {__("Edit")}
            </DropdownItem>
            <DropdownItem
              onClick={onDelete}
              disabled={isDeleting}
              variant="danger"
              icon={IconTrashCan}
            >
              {__("Delete")}
            </DropdownItem>
          </ActionDropdown>
        </Td>
      </Tr>
    </>
  );
}
