/**
 * @generated SignedSource<<eabf44008fc79a4bdd03259111bbf097>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteMeasureInput = {
  measureId: string;
};
export type MeasureViewDeleteMeasureMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteMeasureInput;
};
export type MeasureViewDeleteMeasureMutation$data = {
  readonly deleteMeasure: {
    readonly deletedMeasureId: string;
  };
};
export type MeasureViewDeleteMeasureMutation = {
  response: MeasureViewDeleteMeasureMutation$data;
  variables: MeasureViewDeleteMeasureMutation$variables;
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
  "kind": "ScalarField",
  "name": "deletedMeasureId",
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
    "name": "MeasureViewDeleteMeasureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteMeasurePayload",
        "kind": "LinkedField",
        "name": "deleteMeasure",
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
    "name": "MeasureViewDeleteMeasureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteMeasurePayload",
        "kind": "LinkedField",
        "name": "deleteMeasure",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedMeasureId",
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
    "cacheID": "8369c20d01874393d53f4bcb29b43821",
    "id": null,
    "metadata": {},
    "name": "MeasureViewDeleteMeasureMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureViewDeleteMeasureMutation(\n  $input: DeleteMeasureInput!\n) {\n  deleteMeasure(input: $input) {\n    deletedMeasureId\n  }\n}\n"
  }
};
})();

(node as any).hash = "e82e708b1d2a39ae3faf6a7be8692b7d";

export default node;
