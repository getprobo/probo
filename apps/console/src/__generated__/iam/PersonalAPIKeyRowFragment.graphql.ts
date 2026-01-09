/**
 * @generated SignedSource<<5899ebfb2c67ad34f132f53990b54136>>
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
  readonly lastUsedAt: string | null | undefined;
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
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "lastUsedAt",
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

(node as any).hash = "919daef49b4889be6801516d93f803cf";

export default node;
