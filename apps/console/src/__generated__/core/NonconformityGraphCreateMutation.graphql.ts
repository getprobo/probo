/**
 * @generated SignedSource<<1c0b47ff7df79ff30d8f4c3a6138d972>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type NonconformityStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";
export type CreateNonconformityInput = {
  auditId?: string | null | undefined;
  correctiveAction?: string | null | undefined;
  dateIdentified?: string | null | undefined;
  description?: string | null | undefined;
  dueDate?: string | null | undefined;
  effectivenessCheck?: string | null | undefined;
  organizationId: string;
  ownerId: string;
  referenceId: string;
  rootCause: string;
  status: NonconformityStatus;
};
export type NonconformityGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateNonconformityInput;
};
export type NonconformityGraphCreateMutation$data = {
  readonly createNonconformity: {
    readonly nonconformityEdge: {
      readonly node: {
        readonly audit: {
          readonly framework: {
            readonly name: string;
          };
          readonly id: string;
        } | null | undefined;
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly createdAt: string;
        readonly dateIdentified: string | null | undefined;
        readonly description: string | null | undefined;
        readonly dueDate: string | null | undefined;
        readonly id: string;
        readonly owner: {
          readonly fullName: string;
          readonly id: string;
        };
        readonly referenceId: string;
        readonly rootCause: string;
        readonly status: NonconformityStatus;
      };
    };
  };
};
export type NonconformityGraphCreateMutation = {
  response: NonconformityGraphCreateMutation$data;
  variables: NonconformityGraphCreateMutation$variables;
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
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "referenceId",
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
  "name": "status",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "dateIdentified",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "dueDate",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "rootCause",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "concreteType": "Profile",
  "kind": "LinkedField",
  "name": "owner",
  "plural": false,
  "selections": [
    (v3/*: any*/),
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
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v13 = {
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:nonconformity:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:nonconformity:update\")"
},
v14 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:nonconformity:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:nonconformity:delete\")"
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "NonconformityGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateNonconformityPayload",
        "kind": "LinkedField",
        "name": "createNonconformity",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "NonconformityEdge",
            "kind": "LinkedField",
            "name": "nonconformityEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Nonconformity",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Audit",
                    "kind": "LinkedField",
                    "name": "audit",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Framework",
                        "kind": "LinkedField",
                        "name": "framework",
                        "plural": false,
                        "selections": [
                          (v10/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  (v11/*: any*/),
                  (v12/*: any*/),
                  (v13/*: any*/),
                  (v14/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
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
    "name": "NonconformityGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateNonconformityPayload",
        "kind": "LinkedField",
        "name": "createNonconformity",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "NonconformityEdge",
            "kind": "LinkedField",
            "name": "nonconformityEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Nonconformity",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  (v7/*: any*/),
                  (v8/*: any*/),
                  (v9/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Audit",
                    "kind": "LinkedField",
                    "name": "audit",
                    "plural": false,
                    "selections": [
                      (v3/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Framework",
                        "kind": "LinkedField",
                        "name": "framework",
                        "plural": false,
                        "selections": [
                          (v10/*: any*/),
                          (v3/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  (v11/*: any*/),
                  (v12/*: any*/),
                  (v13/*: any*/),
                  (v14/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "nonconformityEdge",
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
    "cacheID": "a6643ead6aba7ef0e9e25dcf889c51e5",
    "id": null,
    "metadata": {},
    "name": "NonconformityGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation NonconformityGraphCreateMutation(\n  $input: CreateNonconformityInput!\n) {\n  createNonconformity(input: $input) {\n    nonconformityEdge {\n      node {\n        id\n        referenceId\n        description\n        status\n        dateIdentified\n        dueDate\n        rootCause\n        audit {\n          id\n          framework {\n            name\n            id\n          }\n        }\n        owner {\n          id\n          fullName\n        }\n        createdAt\n        canUpdate: permission(action: \"core:nonconformity:update\")\n        canDelete: permission(action: \"core:nonconformity:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "198593806928185e34490d5c656e70ca";

export default node;
