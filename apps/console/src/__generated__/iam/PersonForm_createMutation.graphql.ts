/**
 * @generated SignedSource<<7889e38e185230a2c5b41c3b0fb73cca>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
export type ProfileKind = "CONTRACTOR" | "EMPLOYEE" | "SERVICE_ACCOUNT";
export type CreateUserInput = {
  additionalEmailAddresses?: ReadonlyArray<string> | null | undefined;
  contractEndDate?: string | null | undefined;
  contractStartDate?: string | null | undefined;
  emailAddress: string;
  fullName: string;
  kind: ProfileKind;
  organizationId: string;
  position?: string | null | undefined;
  role: MembershipRole;
};
export type PersonForm_createMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateUserInput;
};
export type PersonForm_createMutation$data = {
  readonly createUser: {
    readonly profileEdge: {
      readonly node: {
        readonly " $fragmentSpreads": FragmentRefs<"PeopleListItemFragment">;
      };
    };
  } | null | undefined;
};
export type PersonForm_createMutation = {
  response: PersonForm_createMutation$data;
  variables: PersonForm_createMutation$variables;
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
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 1
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "DESC",
      "field": "CREATED_AT"
    }
  }
],
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
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
    "name": "PersonForm_createMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateUserPayload",
        "kind": "LinkedField",
        "name": "createUser",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ProfileEdge",
            "kind": "LinkedField",
            "name": "profileEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Profile",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "PeopleListItemFragment"
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
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
    "name": "PersonForm_createMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateUserPayload",
        "kind": "LinkedField",
        "name": "createUser",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "ProfileEdge",
            "kind": "LinkedField",
            "name": "profileEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Profile",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "source",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "state",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "fullName",
                    "storageKey": null
                  },
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
                      (v3/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "role",
                        "storageKey": null
                      },
                      {
                        "alias": "canUpdate",
                        "args": [
                          {
                            "kind": "Literal",
                            "name": "action",
                            "value": "iam:membership:update"
                          }
                        ],
                        "kind": "ScalarField",
                        "name": "permission",
                        "storageKey": "permission(action:\"iam:membership:update\")"
                      }
                    ],
                    "storageKey": null
                  },
                  {
                    "alias": "lastInvitation",
                    "args": (v4/*: any*/),
                    "concreteType": "InvitationConnection",
                    "kind": "LinkedField",
                    "name": "pendingInvitations",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "InvitationEdge",
                        "kind": "LinkedField",
                        "name": "edges",
                        "plural": true,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Invitation",
                            "kind": "LinkedField",
                            "name": "node",
                            "plural": false,
                            "selections": [
                              (v3/*: any*/),
                              (v5/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "__typename",
                                "storageKey": null
                              }
                            ],
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "cursor",
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "PageInfo",
                        "kind": "LinkedField",
                        "name": "pageInfo",
                        "plural": false,
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "endCursor",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "hasNextPage",
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      },
                      {
                        "kind": "ClientExtension",
                        "selections": [
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "__id",
                            "storageKey": null
                          }
                        ]
                      }
                    ],
                    "storageKey": "pendingInvitations(first:1,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
                  },
                  {
                    "alias": "lastInvitation",
                    "args": (v4/*: any*/),
                    "filters": [
                      "orderBy"
                    ],
                    "handle": "connection",
                    "key": "PeopleListItem_lastInvitation",
                    "kind": "LinkedHandle",
                    "name": "pendingInvitations"
                  },
                  (v5/*: any*/),
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
                  },
                  {
                    "alias": "canInvite",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "iam:invitation:create"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"iam:invitation:create\")"
                  },
                  {
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
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "profileEdge",
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
    "cacheID": "738254d707f8ed725c30ee87baa585fb",
    "id": null,
    "metadata": {},
    "name": "PersonForm_createMutation",
    "operationKind": "mutation",
    "text": "mutation PersonForm_createMutation(\n  $input: CreateUserInput!\n) {\n  createUser(input: $input) {\n    profileEdge {\n      node {\n        ...PeopleListItemFragment\n        id\n      }\n    }\n  }\n}\n\nfragment PeopleListItemFragment on Profile {\n  id\n  source\n  state\n  fullName\n  emailAddress\n  membership {\n    id\n    role\n    canUpdate: permission(action: \"iam:membership:update\")\n  }\n  lastInvitation: pendingInvitations(first: 1, orderBy: {field: CREATED_AT, direction: DESC}) {\n    edges {\n      node {\n        id\n        createdAt\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n    }\n  }\n  createdAt\n  canUpdate: permission(action: \"iam:membership-profile:update\")\n  canInvite: permission(action: \"iam:invitation:create\")\n  canDelete: permission(action: \"iam:membership-profile:delete\")\n}\n"
  }
};
})();

(node as any).hash = "8dcc7d2920cc0196f71ecce6e78703c4";

export default node;
