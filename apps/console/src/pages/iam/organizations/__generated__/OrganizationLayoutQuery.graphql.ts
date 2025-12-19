/**
 * @generated SignedSource<<1d82d885a77675a330799121edb5cb69>>
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
    readonly " $fragmentSpreads": FragmentRefs<"OrganizationDropdownFragment">;
  };
  readonly viewer: {
    readonly pendingInvitations: {
      readonly totalCount: number | null | undefined;
    } | null | undefined;
    readonly " $fragmentSpreads": FragmentRefs<"SessionDropdownFragment">;
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
  "concreteType": "InvitationConnection",
  "kind": "LinkedField",
  "name": "pendingInvitations",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "totalCount",
      "storageKey": null
    }
  ],
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
                  "name": "OrganizationDropdownFragment"
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
            (v2/*: any*/),
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
          (v2/*: any*/),
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
    "cacheID": "0ebb386f89d19888f18a05c8dc573895",
    "id": null,
    "metadata": {},
    "name": "OrganizationLayoutQuery",
    "operationKind": "query",
    "text": "query OrganizationLayoutQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      ...OrganizationDropdownFragment\n    }\n    id\n  }\n  viewer {\n    pendingInvitations {\n      totalCount\n    }\n    ...SessionDropdownFragment_4xMPKw\n    id\n  }\n}\n\nfragment OrganizationDropdownFragment on Organization {\n  name\n}\n\nfragment SessionDropdownFragment_4xMPKw on Identity {\n  email\n  profileFor(organizationId: $organizationId) {\n    firstName\n    lastName\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "d3d91888eabb36cc450ca585757b21c7";

export default node;
