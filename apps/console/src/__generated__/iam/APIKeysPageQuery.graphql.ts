/**
 * @generated SignedSource<<76057cc6f78c32b15c6e580c7a0c30da>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type APIKeysPageQuery$variables = Record<PropertyKey, never>;
export type APIKeysPageQuery$data = {
  readonly viewer: {
    readonly " $fragmentSpreads": FragmentRefs<"PersonalAPIKeyListFragment">;
  } | null | undefined;
};
export type APIKeysPageQuery = {
  response: APIKeysPageQuery$data;
  variables: APIKeysPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1000
  }
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "APIKeysPageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "PersonalAPIKeyListFragment"
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
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "APIKeysPageQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v0/*: any*/),
          {
            "alias": null,
            "args": (v1/*: any*/),
            "concreteType": "PersonalAPIKeyConnection",
            "kind": "LinkedField",
            "name": "personalAPIKeys",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "PersonalAPIKeyEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "PersonalAPIKey",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
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
                        "name": "__typename",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "cursor",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "PageInfo",
                "kind": "LinkedField",
                "name": "pageInfo",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "endCursor",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "hasNextPage",
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "personalAPIKeys(first:1000)"
          },
          {
            "alias": null,
            "args": (v1/*: any*/),
            "filters": null,
            "handle": "connection",
            "key": "PersonalAPIKeyListFragment_personalAPIKeys",
            "kind": "LinkedHandle",
            "name": "personalAPIKeys"
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "cf6f6f621bdcb1bd24a2dd94f856729f",
    "id": null,
    "metadata": {},
    "name": "APIKeysPageQuery",
    "operationKind": "query",
    "text": "query APIKeysPageQuery {\n  viewer {\n    ...PersonalAPIKeyListFragment\n    id\n  }\n}\n\nfragment PersonalAPIKeyListFragment on Identity {\n  id\n  personalAPIKeys(first: 1000) {\n    edges {\n      node {\n        id\n        ...PersonalAPIKeyRowFragment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n}\n\nfragment PersonalAPIKeyRowFragment on PersonalAPIKey {\n  id\n  name\n  createdAt\n  expiresAt\n}\n"
  }
};
})();

(node as any).hash = "e99d7224e11d4e9060518c3155f98a7d";

export default node;
