/**
 * @generated SignedSource<<4cd5e6a5205e9438408c95bba2378cde>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ComplianceRegistryStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";
export type CreateComplianceRegistryInput = {
  actionsToBeImplemented?: string | null | undefined;
  area?: string | null | undefined;
  dueDate?: any | null | undefined;
  lastReviewDate?: any | null | undefined;
  organizationId: string;
  ownerId: string;
  referenceId: string;
  regulator?: string | null | undefined;
  requirement?: string | null | undefined;
  source?: string | null | undefined;
  status: ComplianceRegistryStatus;
};
export type ComplianceRegistryGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateComplianceRegistryInput;
};
export type ComplianceRegistryGraphCreateMutation$data = {
  readonly createComplianceRegistry: {
    readonly complianceRegistryEdge: {
      readonly node: {
        readonly actionsToBeImplemented: string | null | undefined;
        readonly area: string | null | undefined;
        readonly createdAt: any;
        readonly dueDate: any | null | undefined;
        readonly id: string;
        readonly lastReviewDate: any | null | undefined;
        readonly owner: {
          readonly fullName: string;
          readonly id: string;
        };
        readonly referenceId: string;
        readonly regulator: string | null | undefined;
        readonly requirement: string | null | undefined;
        readonly source: string | null | undefined;
        readonly status: ComplianceRegistryStatus;
      };
    };
  };
};
export type ComplianceRegistryGraphCreateMutation = {
  response: ComplianceRegistryGraphCreateMutation$data;
  variables: ComplianceRegistryGraphCreateMutation$variables;
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
  "concreteType": "ComplianceRegistryEdge",
  "kind": "LinkedField",
  "name": "complianceRegistryEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "ComplianceRegistry",
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
          "name": "area",
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
          "name": "requirement",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "actionsToBeImplemented",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "regulator",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "lastReviewDate",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "dueDate",
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
    "name": "ComplianceRegistryGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateComplianceRegistryPayload",
        "kind": "LinkedField",
        "name": "createComplianceRegistry",
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
    "name": "ComplianceRegistryGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateComplianceRegistryPayload",
        "kind": "LinkedField",
        "name": "createComplianceRegistry",
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
            "name": "complianceRegistryEdge",
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
    "cacheID": "01d9290d9f4bb8162a0d18ab3c72bdaf",
    "id": null,
    "metadata": {},
    "name": "ComplianceRegistryGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ComplianceRegistryGraphCreateMutation(\n  $input: CreateComplianceRegistryInput!\n) {\n  createComplianceRegistry(input: $input) {\n    complianceRegistryEdge {\n      node {\n        id\n        referenceId\n        area\n        source\n        requirement\n        actionsToBeImplemented\n        regulator\n        lastReviewDate\n        dueDate\n        status\n        owner {\n          id\n          fullName\n        }\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "c203076113948b2c59dddad5577e1c13";

export default node;
