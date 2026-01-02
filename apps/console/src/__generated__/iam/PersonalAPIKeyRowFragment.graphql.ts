/**
 * @generated SignedSource<<d4d4692688621656de57fda5884703dd>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type PersonalAPIKeyRowFragment$data = {
  readonly createdAt: string;
  readonly expiresAt: string;
  readonly id: string;
  readonly name: string;
  readonly token?: string | null | undefined;
  readonly " $fragmentType": "PersonalAPIKeyRowFragment";
};
export type PersonalAPIKeyRowFragment$key = {
  readonly " $data"?: PersonalAPIKeyRowFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"PersonalAPIKeyRowFragment">;
};

import PersonalAPIKeyRowRefetchQuery_graphql from './PersonalAPIKeyRowRefetchQuery.graphql';

const node: ReaderFragment = {
  "argumentDefinitions": [
    {
      "defaultValue": false,
      "kind": "LocalArgument",
      "name": "includeToken"
    }
  ],
  "kind": "Fragment",
  "metadata": {
    "refetch": {
      "connection": null,
      "fragmentPathInResult": [
        "node"
      ],
      "operation": PersonalAPIKeyRowRefetchQuery_graphql,
      "identifierInfo": {
        "identifierField": "id",
        "identifierQueryVariableName": "id"
      }
    }
  },
  "name": "PersonalAPIKeyRowFragment",
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
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "createdAt",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "expiresAt",
      "storageKey": null
    },
    {
      "condition": "includeToken",
      "kind": "Condition",
      "passingValue": true,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "token",
          "storageKey": null
        }
      ]
    }
  ],
  "type": "PersonalAPIKey",
  "abstractKey": null
};

(node as any).hash = "d17db443fa203ee5f9c7d0f4576295f0";

export default node;
