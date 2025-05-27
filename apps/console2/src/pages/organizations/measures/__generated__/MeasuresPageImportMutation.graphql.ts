/**
 * @generated SignedSource<<c0e424b8b0ec7becf9ac2930c6e81b56>>
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
export type MeasuresPageImportMutation$variables = {
  connections: ReadonlyArray<string>;
  input: ImportMeasureInput;
};
export type MeasuresPageImportMutation$data = {
  readonly importMeasure: {
    readonly measureEdges: ReadonlyArray<{
      readonly node: {
        readonly category: string;
        readonly id: string;
        readonly name: string;
        readonly state: MeasureState;
      };
    }>;
  };
};
export type MeasuresPageImportMutation = {
  response: MeasuresPageImportMutation$data;
  variables: MeasuresPageImportMutation$variables;
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
          "name": "category",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "state",
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
    "name": "MeasuresPageImportMutation",
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
    "name": "MeasuresPageImportMutation",
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
    "cacheID": "13a5799ca6febcfb85294f75456deffb",
    "id": null,
    "metadata": {},
    "name": "MeasuresPageImportMutation",
    "operationKind": "mutation",
    "text": "mutation MeasuresPageImportMutation(\n  $input: ImportMeasureInput!\n) {\n  importMeasure(input: $input) {\n    measureEdges {\n      node {\n        id\n        name\n        category\n        state\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6419f51d00f273c824038e7f2d4c65ba";

export default node;
