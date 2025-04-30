/**
 * @generated SignedSource<<1595fa9771d6baa14534b181f4a1384a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TaskState = "DONE" | "TODO";
export type UpdateTaskInput = {
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
    "cacheID": "f545a0b12edb8f231a1a7a7c38b6e905",
    "id": null,
    "metadata": {},
    "name": "MeasureViewUpdateTaskStateMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureViewUpdateTaskStateMutation(\n  $input: UpdateTaskInput!\n) {\n  updateTask(input: $input) {\n    task {\n      id\n      state\n      timeEstimate\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e1d67c030523e56d6bcd595c27fcb3dc";

export default node;
