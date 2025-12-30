/**
 * @generated SignedSource<<caf1a99300d26a5163d5b46577734f90>>
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
export type MeasureGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteMeasureInput;
};
export type MeasureGraphDeleteMutation$data = {
  readonly deleteMeasure: {
    readonly deletedMeasureId: string;
  };
};
export type MeasureGraphDeleteMutation = {
  response: MeasureGraphDeleteMutation$data;
  variables: MeasureGraphDeleteMutation$variables;
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
    "name": "MeasureGraphDeleteMutation",
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
    "name": "MeasureGraphDeleteMutation",
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
    "cacheID": "415f19e622fe431330deca8d3d10f121",
    "id": null,
    "metadata": {},
    "name": "MeasureGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureGraphDeleteMutation(\n  $input: DeleteMeasureInput!\n) {\n  deleteMeasure(input: $input) {\n    deletedMeasureId\n  }\n}\n"
  }
};
})();

(node as any).hash = "867f5fe6142e2900147cb2f7492b22d3";

export default node;
