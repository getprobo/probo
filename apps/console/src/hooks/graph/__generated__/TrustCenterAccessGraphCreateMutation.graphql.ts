/**
 * @generated SignedSource<<271260153dc611541b234266fbdd0c74>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateTrustCenterAccessInput = {
  active: boolean;
  email: string;
  name: string;
  trustCenterId: string;
};
export type TrustCenterAccessGraphCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateTrustCenterAccessInput;
};
export type TrustCenterAccessGraphCreateMutation$data = {
  readonly createTrustCenterAccess: {
    readonly trustCenterAccessEdge: {
      readonly cursor: any;
      readonly node: {
        readonly active: boolean;
        readonly activeCount: number;
        readonly createdAt: any;
        readonly email: string;
        readonly hasAcceptedNonDisclosureAgreement: boolean;
        readonly id: string;
        readonly lastTokenExpiresAt: any | null | undefined;
        readonly name: string;
        readonly pendingRequestCount: number;
      };
    };
  };
};
export type TrustCenterAccessGraphCreateMutation = {
  response: TrustCenterAccessGraphCreateMutation$data;
  variables: TrustCenterAccessGraphCreateMutation$variables;
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
  "concreteType": "TrustCenterAccessEdge",
  "kind": "LinkedField",
  "name": "trustCenterAccessEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "cursor",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "concreteType": "TrustCenterAccess",
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
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "email",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "name",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "active",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "hasAcceptedNonDisclosureAgreement",
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
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "lastTokenExpiresAt",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "pendingRequestCount",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "activeCount",
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
    "name": "TrustCenterAccessGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "createTrustCenterAccess",
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
    "name": "TrustCenterAccessGraphCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateTrustCenterAccessPayload",
        "kind": "LinkedField",
        "name": "createTrustCenterAccess",
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
            "name": "trustCenterAccessEdge",
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
    "cacheID": "501c3c38f8d44e41a178514443803ce2",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphCreateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAccessGraphCreateMutation(\n  $input: CreateTrustCenterAccessInput!\n) {\n  createTrustCenterAccess(input: $input) {\n    trustCenterAccessEdge {\n      cursor\n      node {\n        id\n        email\n        name\n        active\n        hasAcceptedNonDisclosureAgreement\n        createdAt\n        lastTokenExpiresAt\n        pendingRequestCount\n        activeCount\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bfe1841f269ba04a77e9b3ecf73665d2";

export default node;
