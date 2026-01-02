/**
 * @generated SignedSource<<bf8435134aedfa0071f3ee75c67afad3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type StateOfApplicabilityControlState = "EXCLUDED" | "IMPLEMENTED" | "NOT_IMPLEMENTED";
import { FragmentRefs } from "relay-runtime";
export type StateOfApplicabilityControlsTabFragment$data = {
  readonly availableControls: ReadonlyArray<{
    readonly controlId: string;
    readonly exclusionJustification: string | null | undefined;
    readonly frameworkId: string;
    readonly frameworkName: string;
    readonly name: string;
    readonly organizationId: string;
    readonly sectionTitle: string;
    readonly state: StateOfApplicabilityControlState | null | undefined;
    readonly stateOfApplicabilityId: string | null | undefined;
  }>;
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
      "alias": null,
      "args": null,
      "concreteType": "AvailableControlForStateOfApplicability",
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
          "name": "state",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "exclusionJustification",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "type": "StateOfApplicability",
  "abstractKey": null
};

(node as any).hash = "3a26a89593c6b1d809e23362a7768bf9";

export default node;
