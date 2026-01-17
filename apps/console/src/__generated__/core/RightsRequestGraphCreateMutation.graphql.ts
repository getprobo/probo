/**
 * @generated SignedSource<<909c79c6abb5022e54ee451e7e6292be>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RightsRequestState = "DONE" | "IN_PROGRESS" | "TODO";
export type RightsRequestType = "ACCESS" | "DELETION" | "PORTABILITY";
export type CreateRightsRequestInput = {
  actionTaken?: string | null | undefined;
  contact?: string | null | undefined;
  dataSubject?: string | null | undefined;
  deadline?: string | null | undefined;
  details?: string | null | undefined;
  organizationId: string;
  requestState: RightsRequestState;
  requestType: RightsRequestType;
};
export type RightsRequestGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateRightsRequestInput;
};
export type RightsRequestGraphCreateMutation$data = {
  readonly createRightsRequest: {
    readonly rightsRequestEdge: {
      readonly node: {
        readonly actionTaken: string | null | undefined;
        readonly canDelete: boolean;
        readonly canUpdate: boolean;
        readonly contact: string | null | undefined;
        readonly createdAt: string;
        readonly dataSubject: string | null | undefined;
        readonly deadline: string | null | undefined;
        readonly details: string | null | undefined;
        readonly id: string;
        readonly requestState: RightsRequestState;
        readonly requestType: RightsRequestType;
      };
    };
  };
};
export type RightsRequestGraphCreateMutation = {
  response: RightsRequestGraphCreateMutation$data;
  variables: RightsRequestGraphCreateMutation$variables;
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
  "concreteType": "RightsRequestEdge",
  "kind": "LinkedField",
  "name": "rightsRequestEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "RightsRequest",
      "kind": "LinkedField",
      "name": "node",
      "plural": false,
      "selections": [
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "id",
          "storageKey": null
        },
        {
          "alias": "canDelete",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:rights-request:delete"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:rights-request:delete\")"
        },
        {
          "alias": "canUpdate",
          "args": [
            {
              "kind": "Literal",
              "name": "action",
              "value": "core:rights-request:update"
            }
          ],
          "kind": "ScalarField",
          "name": "permission",
          "storageKey": "permission(action:\"core:rights-request:update\")"
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "requestType",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "requestState",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "dataSubject",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "contact",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "details",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "deadline",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "actionTaken",
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
    "name": "RightsRequestGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateRightsRequestPayload",
        "kind": "LinkedField",
        "name": "createRightsRequest",
        "plural": false,
        "selections": [
          (v3/*: any*/)
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
    "name": "RightsRequestGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateRightsRequestPayload",
        "kind": "LinkedField",
        "name": "createRightsRequest",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "rightsRequestEdge",
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
    "cacheID": "be9d2829a9ca6909ab3b4d23c15eb7bc",
    "id": null,
    "metadata": {},
    "name": "RightsRequestGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation RightsRequestGraphCreateMutation(\n  $input: CreateRightsRequestInput!\n) {\n  createRightsRequest(input: $input) {\n    rightsRequestEdge {\n      node {\n        id\n        canDelete: permission(action: \"core:rights-request:delete\")\n        canUpdate: permission(action: \"core:rights-request:update\")\n        requestType\n        requestState\n        dataSubject\n        contact\n        details\n        deadline\n        actionTaken\n        createdAt\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "707197faed79239f374134145ccd1639";

export default node;
