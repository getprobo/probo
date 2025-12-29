/**
 * @generated SignedSource<<8bbf9386bee93021b7e06e5b3bd31c99>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
export type domainSettingsPermissionsQuery$variables = {
  organizationId: string;
};
export type domainSettingsPermissionsQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly viewerMembership: {
      readonly role: MembershipRole;
    };
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
  readonly viewer: {
    readonly canCreateCustomDomain: boolean;
    readonly canDeleteCustomDomain: boolean;
  };
};
export type domainSettingsPermissionsQuery = {
  response: domainSettingsPermissionsQuery$data;
  variables: domainSettingsPermissionsQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "organizationId"
  }
],
v1 = {
  "kind": "Variable",
  "name": "id",
  "variableName": "organizationId"
},
v2 = {
  "alias": "canCreateCustomDomain",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:custom-domain:create"
    },
    (v1/*: any*/)
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": null
},
v3 = {
  "alias": "canDeleteCustomDomain",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:custom-domain:delete"
    },
    (v1/*: any*/)
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": null
},
v4 = [
  (v1/*: any*/)
],
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "role",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "domainSettingsPermissionsQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
          "alias": null,
          "args": null,
          "concreteType": "Identity",
          "kind": "LinkedField",
          "name": "viewer",
          "plural": false,
          "selections": [
            (v2/*: any*/),
            (v3/*: any*/)
          ],
          "storageKey": null
        },
        "action": "THROW"
      },
      {
        "kind": "RequiredField",
        "field": {
          "alias": "organization",
          "args": (v4/*: any*/),
          "concreteType": null,
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v5/*: any*/),
            {
              "kind": "InlineFragment",
              "selections": [
                {
                  "kind": "RequiredField",
                  "field": {
                    "alias": null,
                    "args": null,
                    "concreteType": "Membership",
                    "kind": "LinkedField",
                    "name": "viewerMembership",
                    "plural": false,
                    "selections": [
                      (v6/*: any*/)
                    ],
                    "storageKey": null
                  },
                  "action": "THROW"
                }
              ],
              "type": "Organization",
              "abstractKey": null
            }
          ],
          "storageKey": null
        },
        "action": "THROW"
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "domainSettingsPermissionsQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v3/*: any*/),
          (v7/*: any*/)
        ],
        "storageKey": null
      },
      {
        "alias": "organization",
        "args": (v4/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v5/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Membership",
                "kind": "LinkedField",
                "name": "viewerMembership",
                "plural": false,
                "selections": [
                  (v6/*: any*/),
                  (v7/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          (v7/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "de4faa8385c652071c57b4ddcc58e793",
    "id": null,
    "metadata": {},
    "name": "domainSettingsPermissionsQuery",
    "operationKind": "query",
    "text": "query domainSettingsPermissionsQuery(\n  $organizationId: ID!\n) {\n  viewer {\n    canCreateCustomDomain: permission(action: \"core:custom-domain:create\", id: $organizationId)\n    canDeleteCustomDomain: permission(action: \"core:custom-domain:delete\", id: $organizationId)\n    id\n  }\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      viewerMembership {\n        role\n        id\n      }\n    }\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "75397f9db7cb6f2320ff91f30e773e22";

export default node;
