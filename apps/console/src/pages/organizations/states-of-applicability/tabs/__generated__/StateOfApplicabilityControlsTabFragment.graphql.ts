/**
 * @generated SignedSource<<306449631752d4277aa96b77601e2d8f>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type StateOfApplicabilityControlsTabFragment$data = {
  readonly availableControls: ReadonlyArray<{
    readonly applicability: boolean | null | undefined;
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
  readonly controlsInfo: {
    readonly totalCount: number;
  };
  readonly id: string;
  readonly " $fragmentType": "StateOfApplicabilityControlsTabFragment";
};
export type StateOfApplicabilityControlsTabFragment$key = {
  readonly " $data"?: StateOfApplicabilityControlsTabFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"StateOfApplicabilityControlsTabFragment">;
};

import StateOfApplicabilityControlsTabRefetchQuery_graphql from './StateOfApplicabilityControlsTabRefetchQuery.graphql';

const node: ReaderFragment = {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": {
    "refetch": {
      "connection": null,
      "fragmentPathInResult": [
        "node"
      ],
      "operation": StateOfApplicabilityControlsTabRefetchQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "StateOfApplicabilityControlsTabFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
    {
      "alias": "controlsInfo",
      "args": [
        {
          "kind": "Literal",
          "name": "first",
          "value": 0
        }
      ],
      "concreteType": "ControlConnection",
      "kind": "LinkedField",
      "name": "controls",
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
      "storageKey": "controls(first:0)"
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

(node as any).hash = "e72695289b29d1a4b60e3f7c18ba1165";

export default node;
