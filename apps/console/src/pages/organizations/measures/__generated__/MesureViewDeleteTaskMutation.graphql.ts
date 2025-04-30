/**
 * @generated SignedSource<<c70fd910b41f49426916f3f3aa1433bd>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteTaskInput = {
  taskId: string;
};
export type MeasureViewDeleteTaskMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteTaskInput;
};
export type MeasureViewDeleteTaskMutation$data = {
  readonly deleteTask: {
    readonly deletedTaskId: string;
  };
};
export type MeasureViewDeleteTaskMutation = {
  response: MeasureViewDeleteTaskMutation$data;
  variables: MeasureViewDeleteTaskMutation$variables;
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
  "name": "deletedTaskId",
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
    "name": "MeasureViewDeleteTaskMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteTaskPayload",
        "kind": "LinkedField",
        "name": "deleteTask",
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
    "name": "MeasureViewDeleteTaskMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteTaskPayload",
        "kind": "LinkedField",
        "name": "deleteTask",
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
            "name": "deletedTaskId",
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
    "cacheID": "a5096acddc7796fbc89521cc80b5538e",
    "id": null,
    "metadata": {},
    "name": "MeasureViewDeleteTaskMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureViewDeleteTaskMutation(\n  $input: DeleteTaskInput!\n) {\n  deleteTask(input: $input) {\n    deletedTaskId\n  }\n}\n"
  }
};
})();

(node as any).hash = "42867b2a8776a0facd18b0b24499ccc9";

export default node;
