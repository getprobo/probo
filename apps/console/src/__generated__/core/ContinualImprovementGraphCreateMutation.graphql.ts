/**
 * @generated SignedSource<<aaeb183de2ac35b0a7e60807a0668a4b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ContinualImprovementPriority = "HIGH" | "LOW" | "MEDIUM";
export type ContinualImprovementStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";
export type CreateContinualImprovementInput = {
  description?: string | null | undefined;
  organizationId: string;
  ownerId: string;
  priority: ContinualImprovementPriority;
  referenceId: string;
  source?: string | null | undefined;
  status: ContinualImprovementStatus;
  targetDate?: string | null | undefined;
};
export type ContinualImprovementGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateContinualImprovementInput;
};
export type ContinualImprovementGraphCreateMutation$data = {
  readonly createContinualImprovement: {
    readonly continualImprovementEdge: {
      readonly node: {
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly createdAt: string;
        readonly description: string | null | undefined;
        readonly id: string;
        readonly owner: {
          readonly fullName: string;
          readonly id: string;
        };
        readonly priority: ContinualImprovementPriority;
        readonly referenceId: string;
        readonly source: string | null | undefined;
        readonly status: ContinualImprovementStatus;
        readonly targetDate: string | null | undefined;
      };
    };
  };
};
export type ContinualImprovementGraphCreateMutation = {
  response: ContinualImprovementGraphCreateMutation$data;
  variables: ContinualImprovementGraphCreateMutation$variables;
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
  "concreteType": "ContinualImprovementEdge",
  "kind": "LinkedField",
  "name": "continualImprovementEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "ContinualImprovement",
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
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "createdAt",
          "storageKey": null
        },
        {
          "alias": "canUpdate",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:continual-improvement:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:continual-improvement:update\")"
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:continual-improvement:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:continual-improvement:delete\")"
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
    "name": "ContinualImprovementGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateContinualImprovementPayload",
        "kind": "LinkedField",
        "name": "createContinualImprovement",
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
    "name": "ContinualImprovementGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateContinualImprovementPayload",
        "kind": "LinkedField",
        "name": "createContinualImprovement",
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
            "name": "continualImprovementEdge",
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
    "cacheID": "5894da3b2ddc1c4a16c030770fae53bb",
    "id": null,
    "metadata": {},
    "name": "ContinualImprovementGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ContinualImprovementGraphCreateMutation(\n  $input: CreateContinualImprovementInput!\n) {\n  createContinualImprovement(input: $input) {\n    continualImprovementEdge {\n      node {\n        id\n        referenceId\n        description\n        source\n        targetDate\n        status\n        priority\n        owner {\n          id\n          fullName\n        }\n        createdAt\n        canUpdate: permission(action: \"core:continual-improvement:update\")\n        canDelete: permission(action: \"core:continual-improvement:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bcf5a1e0b02b6530c327741d4d744917";

export default node;
