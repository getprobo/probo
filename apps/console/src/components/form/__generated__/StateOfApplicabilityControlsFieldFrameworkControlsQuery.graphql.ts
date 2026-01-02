/**
 * @generated SignedSource<<760a4c9ef58d487ab996d36a6cf1aa21>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type StateOfApplicabilityControlsFieldFrameworkControlsQuery$variables = {
  frameworkId: string;
};
export type StateOfApplicabilityControlsFieldFrameworkControlsQuery$data = {
  readonly framework: {
    readonly controls?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly name: string;
          readonly sectionTitle: string;
        };
      }>;
    };
    readonly id?: string;
  };
};
export type StateOfApplicabilityControlsFieldFrameworkControlsQuery = {
  response: StateOfApplicabilityControlsFieldFrameworkControlsQuery$data;
  variables: StateOfApplicabilityControlsFieldFrameworkControlsQuery$variables;
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
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": [
    {
      "kind": "Literal",
      "name": "first",
      "value": 500
    },
    {
      "kind": "Literal",
      "name": "orderBy",
      "value": {
        "direction": "ASC",
        "field": "SECTION_TITLE"
      }
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
      "concreteType": "ControlEdge",
      "kind": "LinkedField",
      "name": "edges",
      "plural": true,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "Control",
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v2/*: any*/),
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
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": "controls(first:500,orderBy:{\"direction\":\"ASC\",\"field\":\"SECTION_TITLE\"})"
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "StateOfApplicabilityControlsFieldFrameworkControlsQuery",
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
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/)
            ],
            "type": "Framework",
            "abstractKey": null
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "StateOfApplicabilityControlsFieldFrameworkControlsQuery",
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
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/)
            ],
            "type": "Framework",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "4e57853f9452f6167ef9dd2521f0b998",
    "id": null,
    "metadata": {},
    "name": "StateOfApplicabilityControlsFieldFrameworkControlsQuery",
    "operationKind": "query",
    "text": "query StateOfApplicabilityControlsFieldFrameworkControlsQuery(\n  $frameworkId: ID!\n) {\n  framework: node(id: $frameworkId) {\n    __typename\n    ... on Framework {\n      id\n      controls(first: 500, orderBy: {field: SECTION_TITLE, direction: ASC}) {\n        edges {\n          node {\n            id\n            sectionTitle\n            name\n          }\n        }\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "2817a6951ce5ad1efdee9951636916ff";

export default node;
