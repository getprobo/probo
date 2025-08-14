/**
 * @generated SignedSource<<a656bb20110ce8110b251091f80f27c0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type NonconformityRegistryStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";
export type CreateNonconformityRegistryInput = {
  auditId: string;
  correctiveAction?: string | null | undefined;
  dateIdentified?: any | null | undefined;
  description?: string | null | undefined;
  dueDate?: any | null | undefined;
  effectivenessCheck?: string | null | undefined;
  organizationId: string;
  ownerId: string;
  referenceId: string;
  rootCause: string;
  status: NonconformityRegistryStatus;
};
export type NonconformityRegistryGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateNonconformityRegistryInput;
};
export type NonconformityRegistryGraphCreateMutation$data = {
  readonly createNonconformityRegistry: {
    readonly nonconformityRegistryEdge: {
      readonly node: {
        readonly audit: {
          readonly framework: {
            readonly name: string;
          };
          readonly id: string;
        };
        readonly createdAt: any;
        readonly dateIdentified: any | null | undefined;
        readonly description: string | null | undefined;
        readonly dueDate: any | null | undefined;
        readonly id: string;
        readonly owner: {
          readonly fullName: string;
          readonly id: string;
        };
        readonly referenceId: string;
        readonly rootCause: string;
        readonly status: NonconformityRegistryStatus;
      };
    };
  };
};
export type NonconformityRegistryGraphCreateMutation = {
  response: NonconformityRegistryGraphCreateMutation$data;
  variables: NonconformityRegistryGraphCreateMutation$variables;
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
  "concreteType": "People",
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
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "NonconformityRegistryGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateNonconformityRegistryPayload",
        "kind": "LinkedField",
        "name": "createNonconformityRegistry",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "NonconformityRegistryEdge",
            "kind": "LinkedField",
            "name": "nonconformityRegistryEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "NonconformityRegistry",
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
                  (v12/*: any*/)
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
    "name": "NonconformityRegistryGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateNonconformityRegistryPayload",
        "kind": "LinkedField",
        "name": "createNonconformityRegistry",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "NonconformityRegistryEdge",
            "kind": "LinkedField",
            "name": "nonconformityRegistryEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "NonconformityRegistry",
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
                  (v12/*: any*/)
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
            "name": "nonconformityRegistryEdge",
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
    "cacheID": "de251873fde439228e9a1db5928682d3",
    "id": null,
    "metadata": {},
    "name": "NonconformityRegistryGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation NonconformityRegistryGraphCreateMutation(\n  $input: CreateNonconformityRegistryInput!\n) {\n  createNonconformityRegistry(input: $input) {\n    nonconformityRegistryEdge {\n      node {\n        id\n        referenceId\n        description\n        status\n        dateIdentified\n        dueDate\n        rootCause\n        audit {\n          id\n          framework {\n            name\n            id\n          }\n        }\n        owner {\n          id\n          fullName\n        }\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "249806dbd8c7c28908461d20b3a0ec52";

export default node;
