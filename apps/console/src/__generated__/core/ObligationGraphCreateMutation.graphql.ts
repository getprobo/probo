/**
 * @generated SignedSource<<c9838ad5438c1b9b94b83825a3d3f7bd>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ObligationStatus = "COMPLIANT" | "NON_COMPLIANT" | "PARTIALLY_COMPLIANT";
export type CreateObligationInput = {
  actionsToBeImplemented?: string | null | undefined;
  area?: string | null | undefined;
  dueDate?: any | null | undefined;
  lastReviewDate?: any | null | undefined;
  organizationId: string;
  ownerId: string;
  regulator?: string | null | undefined;
  requirement?: string | null | undefined;
  source?: string | null | undefined;
  status: ObligationStatus;
};
export type ObligationGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateObligationInput;
};
export type ObligationGraphCreateMutation$data = {
  readonly createObligation: {
    readonly obligationEdge: {
      readonly node: {
        readonly actionsToBeImplemented: string | null | undefined;
        readonly area: string | null | undefined;
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly createdAt: any;
        readonly dueDate: any | null | undefined;
        readonly id: string;
        readonly lastReviewDate: any | null | undefined;
        readonly owner: {
          readonly fullName: string;
          readonly id: string;
        };
        readonly regulator: string | null | undefined;
        readonly requirement: string | null | undefined;
        readonly source: string | null | undefined;
        readonly status: ObligationStatus;
      };
    };
  };
};
export type ObligationGraphCreateMutation = {
  response: ObligationGraphCreateMutation$data;
  variables: ObligationGraphCreateMutation$variables;
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
  "concreteType": "ObligationEdge",
  "kind": "LinkedField",
  "name": "obligationEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Obligation",
      "kind": "LinkedField",
      "name": "node",
      "plural": false,
      "selections": [
        (v3/*: any*/),
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
        },
        {
          "alias": "canUpdate",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:obligation:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:obligation:update\")"
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:obligation:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:obligation:delete\")"
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
    "name": "ObligationGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateObligationPayload",
        "kind": "LinkedField",
        "name": "createObligation",
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
    "name": "ObligationGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateObligationPayload",
        "kind": "LinkedField",
        "name": "createObligation",
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
            "name": "obligationEdge",
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
    "cacheID": "04b12a573fa6d56776005407aaf115eb",
    "id": null,
    "metadata": {},
    "name": "ObligationGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation ObligationGraphCreateMutation(\n  $input: CreateObligationInput!\n) {\n  createObligation(input: $input) {\n    obligationEdge {\n      node {\n        id\n        area\n        source\n        requirement\n        actionsToBeImplemented\n        regulator\n        lastReviewDate\n        dueDate\n        status\n        owner {\n          id\n          fullName\n        }\n        createdAt\n        canUpdate: permission(action: \"core:obligation:update\")\n        canDelete: permission(action: \"core:obligation:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "9003c95b7e9857e022107fa0c53c5e05";

export default node;
