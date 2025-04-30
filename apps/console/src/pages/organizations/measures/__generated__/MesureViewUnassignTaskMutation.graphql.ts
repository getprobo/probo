/**
 * @generated SignedSource<<1a725a74db916d1672685d90c48def81>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UnassignTaskInput = {
  taskId: string;
};
export type MeasureViewUnassignTaskMutation$variables = {
  input: UnassignTaskInput;
};
export type MeasureViewUnassignTaskMutation$data = {
  readonly unassignTask: {
    readonly task: {
      readonly assignedTo: {
        readonly fullName: string;
        readonly id: string;
        readonly primaryEmailAddress: string;
      } | null | undefined;
      readonly id: string;
    };
  };
};
export type MeasureViewUnassignTaskMutation = {
  response: MeasureViewUnassignTaskMutation$data;
  variables: MeasureViewUnassignTaskMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UnassignTaskPayload",
    "kind": "LinkedField",
    "name": "unassignTask",
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
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "People",
            "kind": "LinkedField",
            "name": "assignedTo",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "fullName",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "primaryEmailAddress",
                "storageKey": null
              }
            ],
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
    "name": "MeasureViewUnassignTaskMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeasureViewUnassignTaskMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "3f91f027e7a8a935f306386469bbac21",
    "id": null,
    "metadata": {},
    "name": "MeasureViewUnassignTaskMutation",
    "operationKind": "mutation",
    "text": "mutation MeasureViewUnassignTaskMutation(\n  $input: UnassignTaskInput!\n) {\n  unassignTask(input: $input) {\n    task {\n      id\n      assignedTo {\n        id\n        fullName\n        primaryEmailAddress\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "c5c62c4c4d16dfd6147a1d44aa35b8fe";

export default node;
