/**
 * @generated SignedSource<<b771321adc35ef32afc60e1b3e0f2a58>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type NonconformityStatus = "CLOSED" | "IN_PROGRESS" | "OPEN";
export type NonconformityGraphNodeQuery$variables = {
  nonconformityId: string;
};
export type NonconformityGraphNodeQuery$data = {
  readonly node: {
    readonly audit?: {
      readonly framework: {
        readonly id: string;
        readonly name: string;
      };
      readonly id: string;
    } | null | undefined;
    readonly canDelete?: boolean;
    readonly canUpdate?: boolean;
    readonly correctiveAction?: string | null | undefined;
    readonly createdAt?: string;
    readonly dateIdentified?: string | null | undefined;
    readonly description?: string | null | undefined;
    readonly dueDate?: string | null | undefined;
    readonly effectivenessCheck?: string | null | undefined;
    readonly id?: string;
    readonly organization?: {
      readonly id: string;
      readonly name: string;
    };
    readonly owner?: {
      readonly fullName: string;
      readonly id: string;
    };
    readonly referenceId?: string;
    readonly rootCause?: string;
    readonly snapshotId?: string | null | undefined;
    readonly status?: NonconformityStatus;
    readonly updatedAt?: string;
  };
};
export type NonconformityGraphNodeQuery = {
  response: NonconformityGraphNodeQuery$data;
  variables: NonconformityGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "nonconformityId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "nonconformityId"
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
  "name": "snapshotId",
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
  "name": "dateIdentified",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "rootCause",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "correctiveAction",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "dueDate",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "effectivenessCheck",
  "storageKey": null
},
v12 = [
  (v2/*: any*/),
  {
    "alias": null,
    "args": null,
    "kind": "ScalarField",
    "name": "name",
    "storageKey": null
  }
],
v13 = {
  "alias": null,
  "args": null,
  "concreteType": "Audit",
  "kind": "LinkedField",
  "name": "audit",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "concreteType": "Framework",
      "kind": "LinkedField",
      "name": "framework",
      "plural": false,
      "selections": (v12/*: any*/),
      "storageKey": null
    }
  ],
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "concreteType": "Profile",
  "kind": "LinkedField",
  "name": "owner",
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
v15 = {
  "alias": null,
  "args": null,
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": (v12/*: any*/),
  "storageKey": null
},
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v18 = {
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
v19 = {
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
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "NonconformityGraphNodeQuery",
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
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
              (v16/*: any*/),
              (v17/*: any*/),
              (v18/*: any*/),
              (v19/*: any*/)
            ],
            "type": "Nonconformity",
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
    "name": "NonconformityGraphNodeQuery",
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
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "__typename",
            "storageKey": null
          },
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              (v15/*: any*/),
              (v16/*: any*/),
              (v17/*: any*/),
              (v18/*: any*/),
              (v19/*: any*/)
            ],
            "type": "Nonconformity",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "f917aa1ad5d7628a6193743a18c9aab9",
    "id": null,
    "metadata": {},
    "name": "NonconformityGraphNodeQuery",
    "operationKind": "query",
    "text": "query NonconformityGraphNodeQuery(\n  $nonconformityId: ID!\n) {\n  node(id: $nonconformityId) {\n    __typename\n    ... on Nonconformity {\n      id\n      snapshotId\n      referenceId\n      description\n      dateIdentified\n      rootCause\n      correctiveAction\n      dueDate\n      status\n      effectivenessCheck\n      audit {\n        id\n        framework {\n          id\n          name\n        }\n      }\n      owner {\n        id\n        fullName\n      }\n      organization {\n        id\n        name\n      }\n      createdAt\n      updatedAt\n      canUpdate: permission(action: \"core:nonconformity:update\")\n      canDelete: permission(action: \"core:nonconformity:delete\")\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "6d1ecfb7df432a591c7e16461842796d";

export default node;
