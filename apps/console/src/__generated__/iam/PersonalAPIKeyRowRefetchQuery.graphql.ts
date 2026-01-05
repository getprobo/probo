/**
 * @generated SignedSource<<2d413d4e4cf2528a45ff243c0fa733bb>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type PersonalAPIKeyRowRefetchQuery$variables = {
  id: string;
  includeToken?: boolean | null | undefined;
};
export type PersonalAPIKeyRowRefetchQuery$data = {
  readonly node: {
    readonly " $fragmentSpreads": FragmentRefs<"PersonalAPIKeyRowFragment">;
  } | null | undefined;
};
export type PersonalAPIKeyRowRefetchQuery = {
  response: PersonalAPIKeyRowRefetchQuery$data;
  variables: PersonalAPIKeyRowRefetchQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "id"
},
v1 = {
  "defaultValue": false,
  "kind": "LocalArgument",
  "name": "includeToken"
},
v2 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "id"
  }
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "PersonalAPIKeyRowRefetchQuery",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "args": [
              {
                "kind": "Variable",
                "name": "includeToken",
                "variableName": "includeToken"
              }
            ],
            "kind": "FragmentSpread",
            "name": "PersonalAPIKeyRowFragment"
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "PersonalAPIKeyRowRefetchQuery",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "kind": "InlineFragment",
            "selections": [
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
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "b4732244e28d5847baaa5520a95a4cfb",
    "id": null,
    "metadata": {},
    "name": "PersonalAPIKeyRowRefetchQuery",
    "operationKind": "query",
    "text": "query PersonalAPIKeyRowRefetchQuery(\n  $includeToken: Boolean = false\n  $id: ID!\n) {\n  node(id: $id) {\n    __typename\n    ...PersonalAPIKeyRowFragment_2T7Twf\n    id\n  }\n}\n\nfragment PersonalAPIKeyRowFragment_2T7Twf on PersonalAPIKey {\n  id\n  name\n  createdAt\n  expiresAt\n  token @include(if: $includeToken)\n}\n"
  }
};
})();

(node as any).hash = "d17db443fa203ee5f9c7d0f4576295f0";

export default node;
