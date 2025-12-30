/**
 * @generated SignedSource<<05cc2f7b3bf513e9a1f1e192950c011e>>
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
  targetDate?: any | null | undefined;
};
export type ContinualImprovementGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateContinualImprovementInput;
};
export type ContinualImprovementGraphCreateMutation$data = {
  readonly createContinualImprovement: {
    readonly continualImprovementEdge: {
      readonly node: {
        readonly createdAt: any;
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
        readonly targetDate: any | null | undefined;
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
    "cacheID": "8789a6fab51997d85cbc98481c047b9c",
    "id": null,
    "metadata": {},
    "name": "ContinualImprovementGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ContinualImprovementGraphCreateMutation(\n  $input: CreateContinualImprovementInput!\n) {\n  createContinualImprovement(input: $input) {\n    continualImprovementEdge {\n      node {\n        id\n        referenceId\n        description\n        source\n        targetDate\n        status\n        priority\n        owner {\n          id\n          fullName\n        }\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6c6ed7efdb1ebe213fd9e43111f06c7e";

export default node;
