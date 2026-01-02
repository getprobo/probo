/**
 * @generated SignedSource<<5020a4587a487d695572eb77c5e9a959>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type MembershipRole = "ADMIN" | "AUDITOR" | "EMPLOYEE" | "OWNER" | "VIEWER";
export type MembershipLayoutQuery$variables = {
  hideSidebar: boolean;
  organizationId: string;
};
export type MembershipLayoutQuery$data = {
  readonly organization: {
    readonly __typename: "Organization";
    readonly viewerMembership: {
      readonly role: MembershipRole;
    };
    readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdown_organizationFragment" | "SessionDropdownFragment" | "SidebarFragment">;
  } | {
    // This will never be '%other', but we need some
    // value in case none of the concrete values match.
    readonly __typename: "%other";
  };
  readonly viewer: {
    readonly pendingInvitations: {
      readonly totalCount: number;
    };
    readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdown_viewerFragment">;
  };
};
export type MembershipLayoutQuery = {
  response: MembershipLayoutQuery$data;
  variables: MembershipLayoutQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "hideSidebar"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "organizationId"
},
v2 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "organizationId"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "role",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
v6 = {
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
    "name": "MembershipLayoutQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
          "alias": "organization",
          "args": (v2/*: any*/),
          "concreteType": null,
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v3/*: any*/),
            {
              "kind": "InlineFragment",
              "selections": [
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "MembershipsDropdown_organizationFragment"
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "SessionDropdownFragment"
                },
                {
                  "condition": "hideSidebar",
                  "kind": "Condition",
                  "passingValue": false,
                  "selections": [
                    {
                      "args": null,
                      "kind": "FragmentSpread",
                      "name": "SidebarFragment"
                    }
                  ]
                },
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
                      (v4/*: any*/)
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
                    "field": (v5/*: any*/),
                    "action": "THROW"
                  }
                ],
                "storageKey": null
              },
              "action": "THROW"
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
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "MembershipLayoutQuery",
    "selections": [
      {
        "alias": "organization",
        "args": (v2/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "name",
                "storageKey": null
              },
              {
                "alias": "canDelete",
                "args": [
                  {
                    "kind": "Literal",
                    "name": "action",
                    "value": "iam:organization:delete"
                  }
                ],
                "kind": "ScalarField",
                "name": "permission",
                "storageKey": "permission(action:\"iam:organization:delete\")"
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "Membership",
                "kind": "LinkedField",
                "name": "viewerMembership",
                "plural": false,
                "selections": [
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
                      (v6/*: any*/)
                    ],
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "MembershipProfile",
                    "kind": "LinkedField",
                    "name": "profile",
                    "plural": false,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "fullName",
                        "storageKey": null
                      },
                      (v6/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v6/*: any*/),
                  (v4/*: any*/)
                ],
                "storageKey": null
              },
              {
                "condition": "hideSidebar",
                "kind": "Condition",
                "passingValue": false,
                "selections": [
                  {
                    "alias": "canListMeetings",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:meeting:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:meeting:list\")"
                  },
                  {
                    "alias": "canListTasks",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:task:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:task:list\")"
                  },
                  {
                    "alias": "canListMeasures",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:measures:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:measures:list\")"
                  },
                  {
                    "alias": "canListRisks",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:risk:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:risk:list\")"
                  },
                  {
                    "alias": "canListFrameworks",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:frameworks:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:frameworks:list\")"
                  },
                  {
                    "alias": "canListPeople",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:people:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:people:list\")"
                  },
                  {
                    "alias": "canListVendors",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:vendor:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:vendor:list\")"
                  },
                  {
                    "alias": "canListDocuments",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:document:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:document:list\")"
                  },
                  {
                    "alias": "canListAssets",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:asset:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:asset:list\")"
                  },
                  {
                    "alias": "canListData",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:datum:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:datum:list\")"
                  },
                  {
                    "alias": "canListAudits",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:audit:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:audit:list\")"
                  },
                  {
                    "alias": "canListNonconformities",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:nonconformity:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:nonconformity:list\")"
                  },
                  {
                    "alias": "canListObligations",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:obligation:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:obligation:list\")"
                  },
                  {
                    "alias": "canListContinualImprovements",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:continual-improvement:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:continual-improvement:list\")"
                  },
                  {
                    "alias": "canListProcessingActivities",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:processing-activity:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:processing-activity:list\")"
                  },
                  {
                    "alias": "canListRightsRequests",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:rights-request:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:rights-request:list\")"
                  },
                  {
                    "alias": "canListSnapshots",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:snapshot:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:snapshot:list\")"
                  },
                  {
                    "alias": "canGetTrustCenter",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:trust-center:get"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:trust-center:get\")"
                  },
                  {
                    "alias": "canUpdateOrganization",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "iam:organization:update"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"iam:organization:update\")"
                  },
                  {
                    "alias": "canListStatesOfApplicability",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:state-of-applicability:list"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:state-of-applicability:list\")"
                  }
                ]
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          (v6/*: any*/)
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
              (v5/*: any*/)
            ],
            "storageKey": null
          },
          (v6/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "6b0b3a64c02e1c0ff744e57b395d14c8",
    "id": null,
    "metadata": {},
    "name": "MembershipLayoutQuery",
    "operationKind": "query",
    "text": "query MembershipLayoutQuery(\n  $organizationId: ID!\n  $hideSidebar: Boolean!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      ...MembershipsDropdown_organizationFragment\n      ...SessionDropdownFragment\n      ...SidebarFragment @skip(if: $hideSidebar)\n      viewerMembership {\n        role\n        id\n      }\n    }\n    id\n  }\n  viewer {\n    ...MembershipsDropdown_viewerFragment\n    pendingInvitations {\n      totalCount\n    }\n    id\n  }\n}\n\nfragment MembershipsDropdown_organizationFragment on Organization {\n  name\n}\n\nfragment MembershipsDropdown_viewerFragment on Identity {\n  pendingInvitations {\n    totalCount\n  }\n}\n\nfragment SessionDropdownFragment on Organization {\n  canDelete: permission(action: \"iam:organization:delete\")\n  viewerMembership {\n    identity {\n      email\n      id\n    }\n    profile {\n      fullName\n      id\n    }\n    id\n  }\n}\n\nfragment SidebarFragment on Organization {\n  canListMeetings: permission(action: \"core:meeting:list\")\n  canListTasks: permission(action: \"core:task:list\")\n  canListMeasures: permission(action: \"core:measures:list\")\n  canListRisks: permission(action: \"core:risk:list\")\n  canListFrameworks: permission(action: \"core:frameworks:list\")\n  canListPeople: permission(action: \"core:people:list\")\n  canListVendors: permission(action: \"core:vendor:list\")\n  canListDocuments: permission(action: \"core:document:list\")\n  canListAssets: permission(action: \"core:asset:list\")\n  canListData: permission(action: \"core:datum:list\")\n  canListAudits: permission(action: \"core:audit:list\")\n  canListNonconformities: permission(action: \"core:nonconformity:list\")\n  canListObligations: permission(action: \"core:obligation:list\")\n  canListContinualImprovements: permission(action: \"core:continual-improvement:list\")\n  canListProcessingActivities: permission(action: \"core:processing-activity:list\")\n  canListRightsRequests: permission(action: \"core:rights-request:list\")\n  canListSnapshots: permission(action: \"core:snapshot:list\")\n  canGetTrustCenter: permission(action: \"core:trust-center:get\")\n  canUpdateOrganization: permission(action: \"iam:organization:update\")\n  canListStatesOfApplicability: permission(action: \"core:state-of-applicability:list\")\n}\n"
  }
};
})();

(node as any).hash = "2d75823d6ae9eee901e54b71d4c397cf";

export default node;
