/**
 * @generated SignedSource<<66b92b83fa5be4a0dcc2e7f11ad162b6>>
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
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "referenceId",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "description",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "source",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "targetDate",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "status",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "priority",
          "storageKey": null
        },
        {
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
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "createdAt",
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
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
          (v4/*: any*/)
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
          (v4/*: any*/),
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
    "cacheID": "7d03cdc29b6ecb916bf8bdf5ea1df737",
    "id": null,
    "metadata": {},
    "name": "ContinualImprovementRegistryGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ContinualImprovementRegistryGraphCreateMutation(\n  $input: CreateContinualImprovementRegistryInput!\n) {\n  createContinualImprovementRegistry(input: $input) {\n    continualImprovementRegistryEdge {\n      node {\n        id\n        referenceId\n        description\n        source\n        targetDate\n        status\n        priority\n        owner {\n          id\n          fullName\n        }\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "c2acef94a51f4ae101d7d115aa9761bf";

export default node;
