/**
 * @generated SignedSource<<40a72ad7f2eb9fc26bf12c7f4ce731d7>>
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
export type FrameworkControlPageDetachMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlMeasureMappingInput;
};
export type FrameworkControlPageDetachMutation$data = {
  readonly deleteControlMeasureMapping: {
    readonly deletedMeasureId: string;
  };
};
export type FrameworkControlPageDetachMutation = {
  response: FrameworkControlPageDetachMutation$data;
  variables: FrameworkControlPageDetachMutation$variables;
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
    "name": "FrameworkControlPageDetachMutation",
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
    "name": "FrameworkControlPageDetachMutation",
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
    "cacheID": "ce39a8bbb51cd7437dbc08612b60a214",
    "id": null,
    "metadata": {},
    "name": "FrameworkControlPageDetachMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkControlPageDetachMutation(\n  $input: DeleteControlMeasureMappingInput!\n) {\n  deleteControlMeasureMapping(input: $input) {\n    deletedMeasureId\n  }\n}\n"
  }
};
})();

(node as any).hash = "c62748aaf06f0050732be91ae341ae73";

export default node;
