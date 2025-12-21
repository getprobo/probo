/**
 * @generated SignedSource<<4b0d5d271f19e23939fc2b392aa79fbc>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipLayoutQuery$variables = {
  organizationId: string;
};
export type MembershipLayoutQuery$data = {
  readonly organization: {
    readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdown_organizationFragment">;
  };
  readonly viewer: {
    readonly pendingInvitations: {
      readonly totalCount: number;
    };
    readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdown_viewerFragment" | "SessionDropdownFragment">;
  };
};
export type MembershipLayoutQuery = {
  response: MembershipLayoutQuery$data;
  variables: MembershipLayoutQuery$variables;
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
    "name": "MembershipLayoutQuery",
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
                  "name": "MembershipsDropdown_organizationFragment"
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
              "name": "MembershipsDropdown_viewerFragment"
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
    "name": "MembershipLayoutQuery",
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
    "cacheID": "a849dd96b5088ddc9a3bbab50273c411",
    "id": null,
    "metadata": {},
    "name": "MembershipLayoutQuery",
    "operationKind": "query",
    "text": "query MembershipLayoutQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      ...MembershipsDropdown_organizationFragment\n    }\n    id\n  }\n  viewer {\n    ...MembershipsDropdown_viewerFragment\n    pendingInvitations {\n      totalCount\n    }\n    ...SessionDropdownFragment_4xMPKw\n    id\n  }\n}\n\nfragment MembershipsDropdown_organizationFragment on Organization {\n  name\n}\n\nfragment MembershipsDropdown_viewerFragment on Identity {\n  pendingInvitations {\n    totalCount\n  }\n}\n\nfragment SessionDropdownFragment_4xMPKw on Identity {\n  email\n  profileFor(organizationId: $organizationId) {\n    firstName\n    lastName\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "e28f1f5e510386718b39cd547d4ce59f";

export default node;
