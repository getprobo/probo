/**
 * @generated SignedSource<<ff996c0902be9e7b8c23c9dd3d0f3759>>
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
export type FrameworkDetailPageDetachMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteControlMeasureMappingInput;
};
export type FrameworkDetailPageDetachMutation$data = {
  readonly deleteControlMeasureMapping: {
    readonly deletedMeasureId: string;
  };
};
export type FrameworkDetailPageDetachMutation = {
  response: FrameworkDetailPageDetachMutation$data;
  variables: FrameworkDetailPageDetachMutation$variables;
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
    "name": "FrameworkDetailPageDetachMutation",
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
    "name": "FrameworkDetailPageDetachMutation",
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
    "cacheID": "0bede5dc29caa4e2aba9b0dd96e58014",
    "id": null,
    "metadata": {},
    "name": "FrameworkDetailPageDetachMutation",
    "operationKind": "mutation",
    "text": "mutation FrameworkDetailPageDetachMutation(\n  $input: DeleteControlMeasureMappingInput!\n) {\n  deleteControlMeasureMapping(input: $input) {\n    deletedMeasureId\n  }\n}\n"
  }
};
})();

(node as any).hash = "96945bf8e909e7425af2297f7ee6ea0a";

export default node;
