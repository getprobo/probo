/**
 * @generated SignedSource<<cbab0933c2f9f3fad67a6ac7ffc49a67>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type TrustCenterGraphQuery$variables = {
  organizationId: string;
};
export type TrustCenterGraphQuery$data = {
  readonly organization: {
    readonly canCreateTrustCenterFile?: boolean;
    readonly customDomain?: {
      readonly domain: string;
      readonly id: string;
    } | null | undefined;
    readonly id?: string;
    readonly name?: string;
    readonly trustCenter?: {
      readonly active: boolean;
      readonly canCreateAccess: boolean;
      readonly canCreateReference: boolean;
      readonly canDeleteNDA: boolean;
      readonly canGetNDA: boolean;
      readonly canUpdate: boolean;
      readonly canUploadNDA: boolean;
      readonly createdAt: string;
      readonly id: string;
      readonly ndaFileName: string | null | undefined;
      readonly ndaFileUrl: string | null | undefined;
      readonly updatedAt: string;
    } | null | undefined;
  };
};
export type TrustCenterGraphQuery = {
  response: TrustCenterGraphQuery$data;
  variables: TrustCenterGraphQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
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
  "name": "name",
  "storageKey": null
},
v4 = {
  "alias": "canCreateTrustCenterFile",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:trust-center-file:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:trust-center-file:create\")"
},
v5 = {
  "alias": null,
  "args": null,
  "concreteType": "CustomDomain",
  "kind": "LinkedField",
  "name": "customDomain",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "domain",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "concreteType": "TrustCenter",
  "kind": "LinkedField",
  "name": "trustCenter",
  "plural": false,
  "selections": [
    (v2/*: any*/),
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
      "name": "ndaFileName",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "ndaFileUrl",
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
      "name": "updatedAt",
      "storageKey": null
    },
    {
      "alias": "canUpdate",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center:update\")"
    },
    {
      "alias": "canGetNDA",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:get-nda"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center:get-nda\")"
    },
    {
      "alias": "canUploadNDA",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:upload-nda"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center:upload-nda\")"
    },
    {
      "alias": "canDeleteNDA",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:delete-nda"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center:delete-nda\")"
    },
    {
      "alias": "canCreateReference",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center-reference:create"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center-reference:create\")"
    },
    {
      "alias": "canCreateAccess",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center-access:create"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center-access:create\")"
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustCenterGraphQuery",
    "selections": [
      {
        "alias": "organization",
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
              (v6/*: any*/)
            ],
            "type": "Organization",
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
    "name": "TrustCenterGraphQuery",
    "selections": [
      {
        "alias": "organization",
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
              (v6/*: any*/)
            ],
            "type": "Organization",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "be28b37377c063e9b63304ee7353e845",
    "id": null,
    "metadata": {},
    "name": "TrustCenterGraphQuery",
    "operationKind": "query",
    "text": "query TrustCenterGraphQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      id\n      name\n      canCreateTrustCenterFile: permission(action: \"core:trust-center-file:create\")\n      customDomain {\n        id\n        domain\n      }\n      trustCenter {\n        id\n        active\n        ndaFileName\n        ndaFileUrl\n        createdAt\n        updatedAt\n        canUpdate: permission(action: \"core:trust-center:update\")\n        canGetNDA: permission(action: \"core:trust-center:get-nda\")\n        canUploadNDA: permission(action: \"core:trust-center:upload-nda\")\n        canDeleteNDA: permission(action: \"core:trust-center:delete-nda\")\n        canCreateReference: permission(action: \"core:trust-center-reference:create\")\n        canCreateAccess: permission(action: \"core:trust-center-access:create\")\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "9e6d6626e164fd2e3ea553d3f51eb34c";

export default node;
