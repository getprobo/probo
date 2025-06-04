/**
 * @generated SignedSource<<28c68c446adeac27a47de2b754d48d5d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TaskState = "DONE" | "TODO";
export type UpdateTaskInput = {
  deadline?: string | null | undefined;
  description?: string | null | undefined;
  name?: string | null | undefined;
  state?: TaskState | null | undefined;
  taskId: string;
  timeEstimate?: any | null | undefined;
};
export type MeasureViewUpdateTaskStateMutation$variables = {
  input: UpdateTaskInput;
};
export type MeasureViewUpdateTaskStateMutation$data = {
  readonly updateTask: {
    readonly task: {
      readonly deadline: string | null | undefined;
      readonly id: string;
      readonly state: TaskState;
      readonly timeEstimate: any | null | undefined;
    };
  };
};
export type MeasureViewUpdateTaskStateMutation = {
  response: MeasureViewUpdateTaskStateMutation$data;
  variables: MeasureViewUpdateTaskStateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateTaskPayload",
    "kind": "LinkedField",
    "name": "updateTask",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Task",
        "kind": "LinkedField",
        "name": "task",
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
            "name": "state",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "timeEstimate",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "deadline",
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "storageKey": null
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MeasureViewUpdateTaskStateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeasureViewUpdateTaskStateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "a0001fad3af2ba3fc16a1f0af4f19370",
    "id": null,
    "metadata": {},
    "name": "MeasureViewUpdateTaskStateMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureViewUpdateTaskStateMutation(\n  $input: UpdateTaskInput!\n) {\n  updateTask(input: $input) {\n    task {\n      id\n      state\n      timeEstimate\n      deadline\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b3b0d47e0632b106b6313edbd2de38ce";

export default node;
