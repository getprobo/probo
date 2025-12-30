/**
 * @generated SignedSource<<2866738fc7aeb6abd91cf4ed4ae72727>>
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
export type TasksCardDeleteMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteTaskInput;
};
export type TasksCardDeleteMutation$data = {
  readonly deleteTask: {
    readonly deletedTaskId: string;
  };
};
export type TasksCardDeleteMutation = {
  response: TasksCardDeleteMutation$data;
  variables: TasksCardDeleteMutation$variables;
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
    "name": "TasksCardDeleteMutation",
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
    "name": "TasksCardDeleteMutation",
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
    "cacheID": "1ab085f9650841990644d2c106effac7",
    "id": null,
    "metadata": {},
    "name": "TasksCardDeleteMutation",
    "operationKind": "mutation",
    "text": "mutation TasksCardDeleteMutation(\n  $input: DeleteTaskInput!\n) {\n  deleteTask(input: $input) {\n    deletedTaskId\n  }\n}\n"
  }
};
})();

(node as any).hash = "803ffe0cb54f7fa5c840f6d3e4f2592b";

export default node;
