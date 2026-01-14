/**
 * @generated SignedSource<<f1100222ab7c7aeff0f7643640ecc23c>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type TaskState = "DONE" | "TODO";
export type MeasureTasksTabQuery$variables = {
  measureId: string;
};
export type MeasureTasksTabQuery$data = {
  readonly node: {
    readonly id?: string;
    readonly tasks?: {
      readonly __id: string;
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly assignedTo: {
            readonly fullName: string;
            readonly id: string;
          } | null | undefined;
          readonly deadline: any | null | undefined;
          readonly description: string | null | undefined;
          readonly id: string;
          readonly name: string;
          readonly state: TaskState;
          readonly timeEstimate: any | null | undefined;
          readonly " $fragmentSpreads": FragmentRefs<"TaskFormDialogFragment">;
        };
      }>;
    };
  };
};
export type MeasureTasksTabQuery = {
  response: MeasureTasksTabQuery$data;
  variables: MeasureTasksTabQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "measureId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "measureId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "state",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "timeEstimate",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deadline",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "concreteType": "People",
  "kind": "LinkedField",
  "name": "assignedTo",
  "plural": false,
  "selections": [
    (v2/*: any*/),
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
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "concreteType": "PageInfo",
  "kind": "LinkedField",
  "name": "pageInfo",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "endCursor",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "hasNextPage",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v12 = {
  "kind": "ClientExtension",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "__id",
      "storageKey": null
    }
  ]
},
v13 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MeasureTasksTabQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              {
                "alias": "tasks",
                "args": null,
                "concreteType": "TaskConnection",
                "kind": "LinkedField",
                "name": "__Measure__tasks_connection",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TaskEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Task",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v3/*: any*/),
                          (v4/*: any*/),
                          (v5/*: any*/),
                          (v6/*: any*/),
                          (v7/*: any*/),
                          {
                            "args": null,
                            "kind": "FragmentSpread",
                            "name": "TaskFormDialogFragment"
                          },
                          (v8/*: any*/),
                          (v9/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v10/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v11/*: any*/),
                  (v12/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "type": "Measure",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeasureTasksTabQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v9/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": (v13/*: any*/),
                "concreteType": "TaskConnection",
                "kind": "LinkedField",
                "name": "tasks",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TaskEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Task",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v3/*: any*/),
                          (v4/*: any*/),
                          (v5/*: any*/),
                          (v6/*: any*/),
                          (v7/*: any*/),
                          (v8/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Measure",
                            "kind": "LinkedField",
                            "name": "measure",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v9/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v10/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v11/*: any*/),
                  (v12/*: any*/)
                ],
                "storageKey": "tasks(first:100)"
              },
              {
                "alias": null,
                "args": (v13/*: any*/),
                "filters": null,
                "handle": "connection",
                "key": "Measure__tasks",
                "kind": "LinkedHandle",
                "name": "tasks"
              }
            ],
            "type": "Measure",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "7988e2a18d5929fe99d9cef8de244627",
    "id": null,
    "metadata": {
      "connection": [
        {
          "count": null,
          "cursor": null,
          "direction": "forward",
          "path": [
            "node",
            "tasks"
          ]
        }
      ]
    },
    "name": "MeasureTasksTabQuery",
    "operationKind": "query",
    "text": "query MeasureTasksTabQuery(\n  $measureId: ID!\n) {\n  node(id: $measureId) {\n    __typename\n    ... on Measure {\n      id\n      tasks(first: 100) {\n        edges {\n          node {\n            id\n            name\n            state\n            description\n            timeEstimate\n            deadline\n            ...TaskFormDialogFragment\n            assignedTo {\n              id\n              fullName\n            }\n            __typename\n          }\n          cursor\n        }\n        pageInfo {\n          endCursor\n          hasNextPage\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment TaskFormDialogFragment on Task {\n  id\n  description\n  name\n  state\n  timeEstimate\n  deadline\n  assignedTo {\n    id\n  }\n  measure {\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "358a07bb3e13ea9828d0703541caf894";

export default node;
