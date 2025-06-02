import {
  useFragment,
  useMutation,
  usePreloadedQuery,
  type PreloadedQuery,
} from "react-relay";
import { usePageTitle } from "@probo/hooks";
import { graphql } from "relay-runtime";
import { ControlItem, PageHeader } from "@probo/ui";
import { FrameworkLogo } from "/components/FrameworkLogo";
import { frameworkNodeQuery } from "/hooks/graph/FrameworkGraph";
import { useTranslate } from "@probo/i18n";
import { LinkedMeasuresCard } from "/components/measures/LinkedMeasuresCard";
import { useParams } from "react-router";
import { useOrganizationId } from "/hooks/useOrganizationId";
import type { FrameworkGraphNodeQuery } from "/hooks/graph/__generated__/FrameworkGraphNodeQuery.graphql";
import type { FrameworkDetailPageFragment$key } from "./__generated__/FrameworkDetailPageFragment.graphql";
import { LinkedDocumentsCard } from "/components/documents/LinkedDocumentsCard";

const frameworkDetailFragment = graphql`
  fragment FrameworkDetailPageFragment on Framework {
    id
    name
    description
    controls(first: 100) {
      edges {
        node {
          id
          referenceId
          name
          description
          measures(first: 100)
            @connection(key: "FrameworkDetailPage_measures") {
            __id
            edges {
              node {
                id
                ...LinkedMeasuresCardFragment
              }
            }
          }
          documents(first: 100)
            @connection(key: "FrameworkDetailPage_documents") {
            __id
            edges {
              node {
                id
                ...LinkedDocumentsCardFragment
              }
            }
          }
        }
      }
    }
  }
`;

const attachMeasureMutation = graphql`
  mutation FrameworkDetailPageAttachMutation(
    $input: CreateControlMeasureMappingInput!
    $connections: [ID!]!
  ) {
    createControlMeasureMapping(input: $input) {
      measureEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedMeasuresCardFragment
        }
      }
    }
  }
`;

const detachMeasureMutation = graphql`
  mutation FrameworkDetailPageDetachMutation(
    $input: DeleteControlMeasureMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlMeasureMapping(input: $input) {
      deletedMeasureId @deleteEdge(connections: $connections)
    }
  }
`;

const attachDocumentMutation = graphql`
  mutation FrameworkDetailPageAttachDocumentMutation(
    $input: CreateControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    createControlDocumentMapping(input: $input) {
      documentEdge @prependEdge(connections: $connections) {
        node {
          id
          ...LinkedDocumentsCardFragment
        }
      }
    }
  }
`;

const detachDocumentMutation = graphql`
  mutation FrameworkDetailPageDetachDocumentMutation(
    $input: DeleteControlDocumentMappingInput!
    $connections: [ID!]!
  ) {
    deleteControlDocumentMapping(input: $input) {
      deletedDocumentId @deleteEdge(connections: $connections)
    }
  }
`;

type Props = {
  queryRef: PreloadedQuery<FrameworkGraphNodeQuery>;
};

export default function FrameworkDetailPage(props: Props) {
  const { __ } = useTranslate();
  const { controlId } = useParams<{ controlId?: string }>();
  const organizationId = useOrganizationId();
  const data = usePreloadedQuery(frameworkNodeQuery, props.queryRef);
  const framework = useFragment<FrameworkDetailPageFragment$key>(
    frameworkDetailFragment,
    data.node
  );

  const [detachMeasure, isDetachingMeasure] = useMutation(
    detachMeasureMutation
  );
  const [attachMeasure, isAttachingMeasure] = useMutation(
    attachMeasureMutation
  );
  const [detachDocument, isDetachingDocument] = useMutation(
    detachDocumentMutation
  );
  const [attachDocument, isAttachingDocument] = useMutation(
    attachDocumentMutation
  );

  const controls = framework.controls.edges.map((edge) => edge.node);

  const selectedControl = controlId
    ? controls.find((control) => control.id === controlId)
    : controls[0];
  usePageTitle(`${framework.name} | ${selectedControl?.referenceId}`);

  return (
    <div className="space-y-6">
      <PageHeader
        title={
          <>
            <FrameworkLogo {...framework} />
            {framework.name}
          </>
        }
      />
      <div className="text-lg font-semibold">
        {__("Requirement categories")}
      </div>
      <div className="divide-x divide-border-low grid grid-cols-[264px_1fr]">
        <div
          className="space-y-1 overflow-y-auto pr-6 mr-6 sticky top-0"
          style={{ maxHeight: "calc(100vh - 48px)" }}
        >
          {controls.map((control) => (
            <ControlItem
              key={control.id}
              id={control.referenceId}
              description={control.name ?? control.description}
              to={`/organizations/${organizationId}/frameworks/${framework.id}/${control.id}`}
              active={selectedControl?.id === control.id}
            />
          ))}
        </div>
        {selectedControl ? (
          <div className="space-y-6">
            <div className="text-xl font-medium px-[6px] py-[2px] border border-border-low rounded-lg w-max bg-active mb-3">
              {selectedControl.referenceId}
            </div>
            <div className="text-base">{selectedControl.name}</div>
            <LinkedMeasuresCard
              measures={
                selectedControl?.measures.edges.map((edge) => edge.node) ?? []
              }
              params={{ controlId: selectedControl.id }}
              connectionId={selectedControl.measures.__id!}
              onAttach={attachMeasure}
              onDetach={detachMeasure}
              disabled={isAttachingMeasure || isDetachingMeasure}
            />
            <LinkedDocumentsCard
              documents={
                selectedControl?.documents.edges.map((edge) => edge.node) ?? []
              }
              params={{ controlId: selectedControl.id }}
              connectionId={selectedControl.documents.__id!}
              onAttach={attachDocument}
              onDetach={detachDocument}
              disabled={isAttachingDocument || isDetachingDocument}
            />
          </div>
        ) : (
          <div className="h-full flex items-center justify-center">
            <div className="text-sm text-txt-secondary">
              {__("No control selected")}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
