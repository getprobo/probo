/**
 * @generated SignedSource<<ecd129a1f62f100fe57ee9b08f9321d0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ViewerLayoutQuery$variables = Record<PropertyKey, never>;
export type ViewerLayoutQuery$data = {
  readonly viewer: {
    readonly " $fragmentSpreads": FragmentRefs<"ViewerDropdownFragment">;
  };
};
export type ViewerLayoutQuery = {
  response: ViewerLayoutQuery$data;
  variables: ViewerLayoutQuery$variables;
};

const node: ConcreteRequest = {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "ViewerLayoutQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
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
              "name": "ViewerDropdownFragment"
            }
          ],
          "storageKey": null
        },
        "action": "THROW"
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "ViewerLayoutQuery",
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
            "alias": "canListAPIKeys",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "iam:personal-api-key:list"
              }
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": "permission(action:\"iam:personal-api-key:list\")"
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "email",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "fullName",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "faf40277113685d45576f1b466123787",
    "id": null,
    "metadata": {},
    "name": "ViewerLayoutQuery",
    "operationKind": "query",
    "text": "query ViewerLayoutQuery {\n  viewer {\n    ...ViewerDropdownFragment\n    id\n  }\n}\n\nfragment ViewerDropdownFragment on Identity {\n  canListAPIKeys: permission(action: \"iam:personal-api-key:list\")\n  email\n  fullName\n}\n"
  }
};

(node as any).hash = "48e174d981dbc9198a56aac976a20b21";

export default node;
