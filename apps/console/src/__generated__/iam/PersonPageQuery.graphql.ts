/**
 * @generated SignedSource<<679cd9698980b9fdf8b330df8df3aac5>>
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
    readonly __typename: "Profile";
    readonly canDelete: boolean;
    readonly fullName: string;
    readonly id: string;
    readonly identity: {
      readonly email: string;
    };
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
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
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
              "type": "Profile",
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
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
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
                  (v3/*: any*/)
                ],
                "storageKey": null
              },
              (v6/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "emailAddress",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "Membership",
                "kind": "LinkedField",
                "name": "membership",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "role",
                    "storageKey": null
                  },
                  (v3/*: any*/)
                ],
                "storageKey": null
              },
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
            "type": "Profile",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "cb1cec2fd7c66b42f61d40e6a22b0d67",
    "id": null,
    "metadata": {},
    "name": "PersonPageQuery",
    "operationKind": "query",
    "text": "query PersonPageQuery(\n  $personId: ID!\n) {\n  person: node(id: $personId) {\n    __typename\n    ... on Profile {\n      id\n      fullName\n      identity {\n        email\n        id\n      }\n      canDelete: permission(action: \"iam:membership-profile:delete\")\n      ...PersonFormFragment\n    }\n    id\n  }\n}\n\nfragment PersonFormFragment on Profile {\n  id\n  fullName\n  emailAddress\n  membership {\n    role\n    id\n  }\n  kind\n  position\n  additionalEmailAddresses\n  contractStartDate\n  contractEndDate\n  canUpdate: permission(action: \"iam:membership-profile:update\")\n}\n"
  }
};
})();

(node as any).hash = "c2954bc62ed4a3e3a9b1ca9596b68e1c";

export default node;
