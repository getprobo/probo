/**
 * @generated SignedSource<<b4f30f53809f3ed05e6830c0afade59b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ContinualImprovementRegistriesPriority = "HIGH" | "LOW" | "MEDIUM";
export type ContinualImprovementRegistriesStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";
export type CreateContinualImprovementRegistryInput = {
  auditId: string;
  description?: string | null | undefined;
  organizationId: string;
  ownerId: string;
  priority: ContinualImprovementRegistriesPriority;
  referenceId: string;
  source?: string | null | undefined;
  status: ContinualImprovementRegistriesStatus;
  targetDate?: any | null | undefined;
};
export type ContinualImprovementRegistryGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateContinualImprovementRegistryInput;
};
export type ContinualImprovementRegistryGraphCreateMutation$data = {
  readonly createContinualImprovementRegistry: {
    readonly continualImprovementRegistryEdge: {
      readonly node: {
        readonly audit: {
          readonly framework: {
            readonly name: string;
          };
          readonly id: string;
          readonly name: string | null | undefined;
        };
        readonly createdAt: any;
        readonly description: string | null | undefined;
        readonly id: string;
        readonly owner: {
          readonly fullName: string;
          readonly id: string;
        };
        readonly priority: ContinualImprovementRegistriesPriority;
        readonly referenceId: string;
        readonly source: string | null | undefined;
        readonly status: ContinualImprovementRegistriesStatus;
        readonly targetDate: any | null | undefined;
      };
    };
  };
};
export type ContinualImprovementRegistryGraphCreateMutation = {
  response: ContinualImprovementRegistryGraphCreateMutation$data;
  variables: ContinualImprovementRegistryGraphCreateMutation$variables;
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
  "name": "source",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "targetDate",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "priority",
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
    "name": "ContinualImprovementRegistryGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateContinualImprovementRegistryPayload",
        "kind": "LinkedField",
        "name": "createContinualImprovementRegistry",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ContinualImprovementRegistryEdge",
            "kind": "LinkedField",
            "name": "continualImprovementRegistryEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "ContinualImprovementRegistry",
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
                      (v10/*: any*/),
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
    "name": "ContinualImprovementRegistryGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateContinualImprovementRegistryPayload",
        "kind": "LinkedField",
        "name": "createContinualImprovementRegistry",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ContinualImprovementRegistryEdge",
            "kind": "LinkedField",
            "name": "continualImprovementRegistryEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "ContinualImprovementRegistry",
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
                      (v10/*: any*/),
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
            "name": "continualImprovementRegistryEdge",
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
    "cacheID": "a3ddc3c5ac1f55ef2db276ae41ae213b",
    "id": null,
    "metadata": {},
    "name": "ContinualImprovementRegistryGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ContinualImprovementRegistryGraphCreateMutation(\n  $input: CreateContinualImprovementRegistryInput!\n) {\n  createContinualImprovementRegistry(input: $input) {\n    continualImprovementRegistryEdge {\n      node {\n        id\n        referenceId\n        description\n        source\n        targetDate\n        status\n        priority\n        audit {\n          id\n          name\n          framework {\n            name\n            id\n          }\n        }\n        owner {\n          id\n          fullName\n        }\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5d9e0771c839c6c220dabdb00267a7ee";

export default node;
