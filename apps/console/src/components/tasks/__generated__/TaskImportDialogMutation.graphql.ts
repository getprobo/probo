/**
 * @generated SignedSource<<a812a3e89ba844d96c50c59d67f7013c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type TaskState = "DONE" | "TODO";
export type ImportTasksInput = {
  file: any;
  organizationId: string;
};
export type TaskImportDialogMutation$variables = {
  input: ImportTasksInput;
};
export type TaskImportDialogMutation$data = {
  readonly importTasks: {
    readonly errorCount: number;
    readonly importResults: ReadonlyArray<{
      readonly error: string | null | undefined;
      readonly rowNumber: number;
      readonly success: boolean;
      readonly task: {
        readonly assignedTo: {
          readonly fullName: string;
          readonly id: string;
        } | null | undefined;
        readonly description: string;
        readonly id: string;
        readonly measure: {
          readonly id: string;
          readonly name: string;
        } | null | undefined;
        readonly name: string;
        readonly state: TaskState;
        readonly " $fragmentSpreads": FragmentRefs<"TaskFormDialogFragment">;
      } | null | undefined;
    }>;
    readonly successCount: number;
  };
};
export type TaskImportDialogMutation = {
  response: TaskImportDialogMutation$data;
  variables: TaskImportDialogMutation$variables;
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
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "rowNumber",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "success",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "error",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "state",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "concreteType": "Measure",
  "kind": "LinkedField",
  "name": "measure",
  "plural": false,
  "selections": [
    (v5/*: any*/),
    (v6/*: any*/)
  ],
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "concreteType": "People",
  "kind": "LinkedField",
  "name": "assignedTo",
  "plural": false,
  "selections": [
    (v5/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "fullName",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "successCount",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "errorCount",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "TaskImportDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "ImportTasksPayload",
        "kind": "LinkedField",
        "name": "importTasks",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TaskImportResult",
            "kind": "LinkedField",
            "name": "importResults",
            "plural": true,
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Task",
                "kind": "LinkedField",
                "name": "task",
                "plural": false,
                "selections": [
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "TaskFormDialogFragment"
                  },
                  (v9/*: any*/),
                  (v10/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          (v11/*: any*/),
          (v12/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TaskImportDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "ImportTasksPayload",
        "kind": "LinkedField",
        "name": "importTasks",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TaskImportResult",
            "kind": "LinkedField",
            "name": "importResults",
            "plural": true,
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
              (v4/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Task",
                "kind": "LinkedField",
                "name": "task",
                "plural": false,
                "selections": [
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
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
                  },
                  (v10/*: any*/),
                  (v9/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          (v11/*: any*/),
          (v12/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "8573e02bc14a267ce5facb1d27c04dda",
    "id": null,
    "metadata": {},
    "name": "TaskImportDialogMutation",
    "operationKind": "mutation",
    "text": "mutation TaskImportDialogMutation(\n  $input: ImportTasksInput!\n) {\n  importTasks(input: $input) {\n    importResults {\n      rowNumber\n      success\n      error\n      task {\n        id\n        name\n        state\n        description\n        ...TaskFormDialogFragment\n        measure {\n          id\n          name\n        }\n        assignedTo {\n          id\n          fullName\n        }\n      }\n    }\n    successCount\n    errorCount\n  }\n}\n\nfragment TaskFormDialogFragment on Task {\n  id\n  description\n  name\n  state\n  timeEstimate\n  deadline\n  assignedTo {\n    id\n  }\n  measure {\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "488e71990b25d3077d4e0b88b0437d8c";

export default node;
