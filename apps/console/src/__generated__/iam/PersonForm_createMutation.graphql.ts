/**
 * @generated SignedSource<<5148e84724d7649be7a0ccf0fe813893>>
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
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Identity",
                    "kind": "LinkedField",
                    "name": "identity",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "email",
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
                    "name": "createdAt",
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
    "cacheID": "ff3f29c40614a067d108d55df828f7c0",
    "id": null,
    "metadata": {},
    "name": "PersonForm_createMutation",
    "operationKind": "mutation",
    "text": "mutation PersonForm_createMutation(\n  $input: CreateUserInput!\n) {\n  createUser(input: $input) {\n    profileEdge {\n      node {\n        ...PeopleListItemFragment\n        id\n      }\n    }\n  }\n}\n\nfragment PeopleListItemFragment on Profile {\n  id\n  source\n  state\n  fullName\n  kind\n  position\n  membership {\n    id\n    role\n    canUpdate: permission(action: \"iam:membership:update\")\n    canDelete: permission(action: \"iam:membership-profile:delete\")\n  }\n  identity {\n    email\n    id\n  }\n  createdAt\n  canUpdate: permission(action: \"iam:membership-profile:update\")\n}\n"
  }
};
})();

(node as any).hash = "8dcc7d2920cc0196f71ecce6e78703c4";

export default node;
