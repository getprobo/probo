/**
 * @generated SignedSource<<aa1188431214a8844e8ee1879bbada7e>>
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
    readonly emailAddress: string;
    readonly fullName: string;
    readonly id: string;
    readonly source: string;
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
  "name": "emailAddress",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "source",
  "storageKey": null
},
v7 = {
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
                (v5/*: any*/),
                (v6/*: any*/),
                (v7/*: any*/),
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
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
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
    "cacheID": "f78728c7deb19de68b5b996c58ebd05e",
    "id": null,
    "metadata": {},
    "name": "PersonPageQuery",
    "operationKind": "query",
    "text": "query PersonPageQuery(\n  $personId: ID!\n) {\n  person: node(id: $personId) {\n    __typename\n    ... on Profile {\n      id\n      fullName\n      emailAddress\n      source\n      canDelete: permission(action: \"iam:membership-profile:delete\")\n      ...PersonFormFragment\n    }\n    id\n  }\n}\n\nfragment PersonFormFragment on Profile {\n  id\n  fullName\n  emailAddress\n  source\n  membership {\n    role\n    id\n  }\n  kind\n  position\n  additionalEmailAddresses\n  contractStartDate\n  contractEndDate\n  canUpdate: permission(action: \"iam:membership-profile:update\")\n}\n"
  }
};
})();

(node as any).hash = "4aecfb7a9ff05a585ccc6b533fb75bec";

export default node;
