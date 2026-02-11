/**
 * @generated SignedSource<<a809c0ce9c8f1fdb68a0b2835be99b05>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type PersonPageQuery$variables = {
  personId: string;
};
export type PersonPageQuery$data = {
  readonly person: {
    readonly __typename: "MembershipProfile";
    readonly canDelete: boolean;
    readonly fullName: string;
    readonly identity: {
      readonly email: string;
    };
    readonly membershipId: string;
    readonly " $fragmentSpreads": FragmentRefs<"PersonFormFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
};
export type PersonPageQuery = {
  response: PersonPageQuery$data;
  variables: PersonPageQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "personId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "personId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "membershipId",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
v6 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "iam:membership-profile:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"iam:membership-profile:delete\")"
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
    "name": "PersonPageQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
          "alias": "person",
          "args": (v1/*: any*/),
          "concreteType": null,
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v2/*: any*/),
            {
              "kind": "InlineFragment",
              "selections": [
                (v3/*: any*/),
                (v4/*: any*/),
                {
                  "kind": "RequiredField",
                  "field": {
                    "alias": null,
                    "args": null,
                    "concreteType": "Identity",
                    "kind": "LinkedField",
                    "name": "identity",
                    "plural": false,
                    "selections": [
                      (v5/*: any*/)
                    ],
                    "storageKey": null
                  },
                  "action": "THROW"
                },
                (v6/*: any*/),
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "PersonFormFragment"
                }
              ],
              "type": "MembershipProfile",
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
    "name": "PersonPageQuery",
    "selections": [
      {
        "alias": "person",
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v7/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Identity",
                "kind": "LinkedField",
                "name": "identity",
                "plural": false,
                "selections": [
                  (v5/*: any*/),
                  (v7/*: any*/)
                ],
                "storageKey": null
              },
              (v6/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "kind",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "position",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "additionalEmailAddresses",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "contractStartDate",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "contractEndDate",
                "storageKey": null
              },
              {
                "alias": "canUpdate",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "iam:membership-profile:update"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"iam:membership-profile:update\")"
              }
            ],
            "type": "MembershipProfile",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "f754630a84c19798f2b0611dba4c6603",
    "id": null,
    "metadata": {},
    "name": "PersonPageQuery",
    "operationKind": "query",
    "text": "query PersonPageQuery(\n  $personId: ID!\n) {\n  person: node(id: $personId) {\n    __typename\n    ... on MembershipProfile {\n      fullName\n      membershipId\n      identity {\n        email\n        id\n      }\n      canDelete: permission(action: \"iam:membership-profile:delete\")\n      ...PersonFormFragment\n    }\n    id\n  }\n}\n\nfragment PersonFormFragment on MembershipProfile {\n  id\n  fullName\n  kind\n  position\n  additionalEmailAddresses\n  contractStartDate\n  contractEndDate\n  canUpdate: permission(action: \"iam:membership-profile:update\")\n}\n"
  }
};
})();

(node as any).hash = "d701ae3a51eb1fed62aec260f42b3c41";

export default node;
