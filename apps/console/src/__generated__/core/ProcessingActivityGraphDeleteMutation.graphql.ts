/**
 * @generated SignedSource<<a6cfe48929d6439d3dee377f49f161e0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteProcessingActivityInput = {
  processingActivityId: string;
};
export type ProcessingActivityGraphDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteProcessingActivityInput;
};
export type ProcessingActivityGraphDeleteMutation$data = {
  readonly deleteProcessingActivity: {
    readonly deletedProcessingActivityId: string;
  };
};
export type ProcessingActivityGraphDeleteMutation = {
  response: ProcessingActivityGraphDeleteMutation$data;
  variables: ProcessingActivityGraphDeleteMutation$variables;
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
  "name": "deletedProcessingActivityId",
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
    "name": "ProcessingActivityGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteProcessingActivityPayload",
        "kind": "LinkedField",
        "name": "deleteProcessingActivity",
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
    "name": "ProcessingActivityGraphDeleteMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteProcessingActivityPayload",
        "kind": "LinkedField",
        "name": "deleteProcessingActivity",
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
            "name": "deletedProcessingActivityId",
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
    "cacheID": "302c79f7bfa307aa484acccab48c8591",
    "id": null,
    "metadata": {},
    "name": "ProcessingActivityGraphDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation ProcessingActivityGraphDeleteMutation(\n  $input: DeleteProcessingActivityInput!\n) {\n  deleteProcessingActivity(input: $input) {\n    deletedProcessingActivityId\n  }\n}\n"
  }
};
})();

(node as any).hash = "a8724b743deef9c8889ae937311338c3";

export default node;
