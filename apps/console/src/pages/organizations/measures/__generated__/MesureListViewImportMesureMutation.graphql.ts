/**
 * @generated SignedSource<<35c287e8ba9940b8d697c48c563f41c8>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MeasureState = "IMPLEMENTED" | "IN_PROGRESS" | "NOT_APPLICABLE" | "NOT_STARTED";
export type ImportMeasureInput = {
  file: any;
  organizationId: string;
};
export type MeasureListViewImportMeasureMutation$variables = {
  connections: ReadonlyArray<string>;
  input: ImportMeasureInput;
};
export type MeasureListViewImportMeasureMutation$data = {
  readonly importMeasure: {
    readonly measureEdges: ReadonlyArray<{
      readonly node: {
        readonly category: string;
        readonly createdAt: string;
        readonly description: string;
        readonly id: string;
        readonly name: string;
        readonly state: MeasureState;
        readonly updatedAt: string;
      };
    }>;
  };
};
export type MeasureListViewImportMeasureMutation = {
  response: MeasureListViewImportMeasureMutation$data;
  variables: MeasureListViewImportMeasureMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "concreteType": "MeasureEdge",
  "kind": "LinkedField",
  "name": "measureEdges",
  "plural": true,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Measure",
      "kind": "LinkedField",
      "name": "node",
      "plural": false,
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
          "name": "description",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "category",
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
          "name": "createdAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "updatedAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "MeasureListViewImportMeasureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "ImportMeasurePayload",
        "kind": "LinkedField",
        "name": "importMeasure",
        "plural": false,
        "selections": [
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "MeasureListViewImportMeasureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "ImportMeasurePayload",
        "kind": "LinkedField",
        "name": "importMeasure",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "appendEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "measureEdges",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "8cf2772dcac48b487c892d2b68a3bb51",
    "id": null,
    "metadata": {},
    "name": "MeasureListViewImportMeasureMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureListViewImportMeasureMutation(\n  $input: ImportMeasureInput!\n) {\n  importMeasure(input: $input) {\n    measureEdges {\n      node {\n        id\n        name\n        description\n        category\n        state\n        createdAt\n        updatedAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "38c30a8751334cce8ba85a64574d9dd6";

export default node;
