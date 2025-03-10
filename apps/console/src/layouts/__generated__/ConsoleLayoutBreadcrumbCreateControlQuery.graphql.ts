/**
 * @generated SignedSource<<051137c0676ed5fd6b8807a44b9cb2a4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ConsoleLayoutBreadcrumbCreateControlQuery$variables = {
  frameworkId: string;
};
export type ConsoleLayoutBreadcrumbCreateControlQuery$data = {
  readonly framework: {
    readonly name?: string;
  };
};
export type ConsoleLayoutBreadcrumbCreateControlQuery = {
  response: ConsoleLayoutBreadcrumbCreateControlQuery$data;
  variables: ConsoleLayoutBreadcrumbCreateControlQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "frameworkId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "frameworkId"
  }
],
v2 = {
  "kind": "InlineFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "name",
      "storageKey": null
    }
  ],
  "type": "Framework",
  "abstractKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ConsoleLayoutBreadcrumbCreateControlQuery",
    "selections": [
      {
        "alias": "framework",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ConsoleLayoutBreadcrumbCreateControlQuery",
    "selections": [
      {
        "alias": "framework",
        "args": (v1/*: any*/),
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
          (v2/*: any*/),
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
    "cacheID": "a86ec305f24f6a83d549a31e1469a89c",
    "id": null,
    "metadata": {},
    "name": "ConsoleLayoutBreadcrumbCreateControlQuery",
    "operationKind": "query",
    "text": "query ConsoleLayoutBreadcrumbCreateControlQuery(\n  $frameworkId: ID!\n) {\n  framework: node(id: $frameworkId) {\n    __typename\n    ... on Framework {\n      name\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "5033069853aec310ad544974f7e64b5b";

export default node;
