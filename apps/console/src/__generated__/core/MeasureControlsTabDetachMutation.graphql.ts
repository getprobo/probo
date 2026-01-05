/**
 * @generated SignedSource<<63367052d20bc149e2d4f86adbf3c735>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteControlMeasureMappingInput = {
  controlId: string;
  measureId: string;
};
export type MeasureControlsTabDetachMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlMeasureMappingInput;
};
export type MeasureControlsTabDetachMutation$data = {
  readonly deleteControlMeasureMapping: {
    readonly deletedControlId: string;
  };
};
export type MeasureControlsTabDetachMutation = {
  response: MeasureControlsTabDetachMutation$data;
  variables: MeasureControlsTabDetachMutation$variables;
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
  "name": "deletedControlId",
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
    "name": "MeasureControlsTabDetachMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlMeasureMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlMeasureMapping",
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
    "name": "MeasureControlsTabDetachMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteControlMeasureMappingPayload",
        "kind": "LinkedField",
        "name": "deleteControlMeasureMapping",
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
            "name": "deletedControlId",
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
    "cacheID": "158225e40cb2e95be134e59a4647ff69",
    "id": null,
    "metadata": {},
    "name": "MeasureControlsTabDetachMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureControlsTabDetachMutation(\n  $input: DeleteControlMeasureMappingInput!\n) {\n  deleteControlMeasureMapping(input: $input) {\n    deletedControlId\n  }\n}\n"
  }
};
})();

(node as any).hash = "d800ddbcf4b4c24ae945447d710be209";

export default node;
