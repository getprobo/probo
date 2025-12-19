/**
 * @generated SignedSource<<09fbc762db2d00eb944a0ef9afe3756d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type OrganizationLayoutQuery$variables = {
  organizationId: string;
};
export type OrganizationLayoutQuery$data = {
  readonly organization: {
    readonly " $fragmentSpreads": FragmentRefs<"OrganizationDropdown_organizationFragment">;
  };
  readonly viewer: {
    readonly pendingInvitations: {
      readonly totalCount: number;
    };
    readonly " $fragmentSpreads": FragmentRefs<"OrganizationDropdown_viewerFragment" | "SessionDropdownFragment">;
  };
};
export type OrganizationLayoutQuery = {
  response: OrganizationLayoutQuery$data;
  variables: OrganizationLayoutQuery$variables;
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
  "name": "totalCount",
  "storageKey": null
},
v3 = [
  {
    "kind": "Variable",
    "name": "organizationId",
    "variableName": "organizationId"
  }
],
v4 = {
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
    "name": "OrganizationLayoutQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
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
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "OrganizationDropdown_organizationFragment"
                }
              ],
              "type": "Organization",
              "abstractKey": null
            }
          ],
          "storageKey": null
        },
        "action": "THROW"
      },
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
            {
              "args": null,
              "kind": "FragmentSpread",
              "name": "OrganizationDropdown_viewerFragment"
            },
            {
              "kind": "RequiredField",
              "field": {
                "alias": null,
                "args": null,
                "concreteType": "InvitationConnection",
                "kind": "LinkedField",
                "name": "pendingInvitations",
                "plural": false,
                "selections": [
                  {
                    "kind": "RequiredField",
                    "field": (v2/*: any*/),
                    "action": "THROW"
                  }
                ],
                "storageKey": null
              },
              "action": "THROW"
            },
            {
              "args": (v3/*: any*/),
              "kind": "FragmentSpread",
              "name": "SessionDropdownFragment"
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
    "name": "OrganizationLayoutQuery",
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
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "name",
                "storageKey": null
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          (v4/*: any*/)
        ],
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "InvitationConnection",
            "kind": "LinkedField",
            "name": "pendingInvitations",
            "plural": false,
            "selections": [
              (v2/*: any*/)
            ],
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
            "args": (v3/*: any*/),
            "concreteType": "IdentityProfile",
            "kind": "LinkedField",
            "name": "profileFor",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "firstName",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "lastName",
                "storageKey": null
              },
              (v4/*: any*/)
            ],
            "storageKey": null
          },
          (v4/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "ca7f41668eb74024237351d0e076819e",
    "id": null,
    "metadata": {},
    "name": "OrganizationLayoutQuery",
    "operationKind": "query",
    "text": "query OrganizationLayoutQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      ...OrganizationDropdown_organizationFragment\n    }\n    id\n  }\n  viewer {\n    ...OrganizationDropdown_viewerFragment\n    pendingInvitations {\n      totalCount\n    }\n    ...SessionDropdownFragment_4xMPKw\n    id\n  }\n}\n\nfragment OrganizationDropdown_organizationFragment on Organization {\n  name\n}\n\nfragment OrganizationDropdown_viewerFragment on Identity {\n  pendingInvitations {\n    totalCount\n  }\n}\n\nfragment SessionDropdownFragment_4xMPKw on Identity {\n  email\n  profileFor(organizationId: $organizationId) {\n    firstName\n    lastName\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "81b8eb0af21d649268e30c80d16226de";

export default node;
