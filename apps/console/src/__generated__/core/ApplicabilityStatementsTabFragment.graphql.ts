/**
 * @generated SignedSource<<cfb8a499e31ed5f75438beec971aee78>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ApplicabilityStatementsTabFragment$data = {
  readonly applicabilityStatementsInfo: {
    readonly totalCount: number;
  };
  readonly availableControls: ReadonlyArray<{
    readonly applicability: boolean | null | undefined;
    readonly applicabilityStatementId: string | null | undefined;
    readonly bestPractice: boolean;
    readonly contractual: boolean;
    readonly controlId: string;
    readonly frameworkId: string;
    readonly frameworkName: string;
    readonly justification: string | null | undefined;
    readonly name: string;
    readonly organizationId: string;
    readonly regulatory: boolean;
    readonly riskAssessment: boolean;
    readonly sectionTitle: string;
    readonly stateOfApplicabilityId: string | null | undefined;
  }>;
  readonly canCreateApplicabilityStatement: boolean;
  readonly canDeleteApplicabilityStatement: boolean;
  readonly id: string;
  readonly " $fragmentType": "ApplicabilityStatementsTabFragment";
};
export type ApplicabilityStatementsTabFragment$key = {
  readonly " $data"?: ApplicabilityStatementsTabFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"ApplicabilityStatementsTabFragment">;
};

import ApplicabilityStatementsTabRefetchQuery_graphql from './ApplicabilityStatementsTabRefetchQuery.graphql';

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": {
    "refetch": {
      "connection": null,
      "fragmentPathInResult": [
        "node"
      ],
      "operation": ApplicabilityStatementsTabRefetchQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "ApplicabilityStatementsTabFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
    {
      "alias": "applicabilityStatementsInfo",
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 0
        }
      ],
      "concreteType": "ApplicabilityStatementConnection",
      "kind": "LinkedField",
      "name": "applicabilityStatements",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "totalCount",
          "storageKey": null
        }
      ],
      "storageKey": "applicabilityStatements(first:0)"
    },
    {
      "alias": "canCreateApplicabilityStatement",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:state-of-applicability-control-mapping:create"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:state-of-applicability-control-mapping:create\")"
    },
    {
      "alias": "canDeleteApplicabilityStatement",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:state-of-applicability-control-mapping:delete"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:state-of-applicability-control-mapping:delete\")"
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "AvailableStateOfApplicabilityControl",
      "kind": "LinkedField",
      "name": "availableControls",
      "plural": true,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "controlId",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "sectionTitle",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "name",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "frameworkId",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "frameworkName",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "organizationId",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "applicabilityStatementId",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "stateOfApplicabilityId",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "applicability",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "justification",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "bestPractice",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "regulatory",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "contractual",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "riskAssessment",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "StateOfApplicability",
  "abstractKey": null
};

(node as any).hash = "8183adc2185e14fe2962c58058bef8ed";

export default node;
