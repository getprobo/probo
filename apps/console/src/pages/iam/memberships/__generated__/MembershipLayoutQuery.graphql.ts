/**
 * @generated SignedSource<<90e5e1f754a2b18caa9b6f4f208ec5a8>>
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
    readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdown_organizationFragment" | "SessionDropdownFragment">;
  };
  readonly viewer: {
    readonly pendingInvitations: {
      readonly totalCount: number;
    };
    readonly " $fragmentSpreads": FragmentRefs<"MembershipsDropdown_viewerFragment" | "SidebarFragment">;
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
v1 = {
  "kind": "Variable",
  "name": "id",
  "variableName": "organizationId"
},
v2 = [
  (v1/*: any*/)
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "totalCount",
  "storageKey": null
},
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
          "args": (v2/*: any*/),
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
                },
                {
                  "args": null,
                  "kind": "FragmentSpread",
                  "name": "SessionDropdownFragment"
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
              "args": [
                {
                  "kind": "Variable",
                  "name": "organizationId",
                  "variableName": "organizationId"
                }
              ],
              "kind": "FragmentSpread",
              "name": "SidebarFragment"
            },
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
                    "field": (v3/*: any*/),
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
    "argumentDefinitions": (v0/*: any*/),
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
                      (v4/*: any*/)
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
                      (v4/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v4/*: any*/)
                ],
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
            "alias": "canListMeetings",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:meeting:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListTasks",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:task:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListMeasures",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:measures:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListRisks",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:risk:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListFrameworks",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:frameworks:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListPeople",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:people:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListVendors",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:vendor:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListDocuments",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:document:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListAssets",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:asset:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListData",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:datum:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListAudits",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:audit:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListNonconformities",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:nonconformity:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListObligations",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:obligation:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListContinualImprovements",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:continual-improvement:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListProcessingActivities",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:processing-activity:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListStatesOfApplicability",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:state-of-applicability:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canListSnapshots",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:snapshot:list"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canGetTrustCenter",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "core:trust-center:get"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": "canUpdateOrganization",
            "args": [
              {
                "kind": "Literal",
                "name": "action",
                "value": "iam:organization:update"
              },
              (v1/*: any*/)
            ],
            "kind": "ScalarField",
            "name": "permission",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "InvitationConnection",
            "kind": "LinkedField",
            "name": "pendingInvitations",
            "plural": false,
            "selections": [
              (v3/*: any*/)
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
    "cacheID": "1fb1002661548737fdb22694a23bca35",
    "id": null,
    "metadata": {},
    "name": "MembershipLayoutQuery",
    "operationKind": "query",
    "text": "query MembershipLayoutQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      ...MembershipsDropdown_organizationFragment\n      ...SessionDropdownFragment\n    }\n    id\n  }\n  viewer {\n    ...SidebarFragment_4xMPKw\n    ...MembershipsDropdown_viewerFragment\n    pendingInvitations {\n      totalCount\n    }\n    id\n  }\n}\n\nfragment MembershipsDropdown_organizationFragment on Organization {\n  name\n}\n\nfragment MembershipsDropdown_viewerFragment on Identity {\n  pendingInvitations {\n    totalCount\n  }\n}\n\nfragment SessionDropdownFragment on Organization {\n  viewerMembership {\n    identity {\n      email\n      id\n    }\n    profile {\n      fullName\n      id\n    }\n    id\n  }\n}\n\nfragment SidebarFragment_4xMPKw on Identity {\n  canListMeetings: permission(action: \"core:meeting:list\", id: $organizationId)\n  canListTasks: permission(action: \"core:task:list\", id: $organizationId)\n  canListMeasures: permission(action: \"core:measures:list\", id: $organizationId)\n  canListRisks: permission(action: \"core:risk:list\", id: $organizationId)\n  canListFrameworks: permission(action: \"core:frameworks:list\", id: $organizationId)\n  canListPeople: permission(action: \"core:people:list\", id: $organizationId)\n  canListVendors: permission(action: \"core:vendor:list\", id: $organizationId)\n  canListDocuments: permission(action: \"core:document:list\", id: $organizationId)\n  canListAssets: permission(action: \"core:asset:list\", id: $organizationId)\n  canListData: permission(action: \"core:datum:list\", id: $organizationId)\n  canListAudits: permission(action: \"core:audit:list\", id: $organizationId)\n  canListNonconformities: permission(action: \"core:nonconformity:list\", id: $organizationId)\n  canListObligations: permission(action: \"core:obligation:list\", id: $organizationId)\n  canListContinualImprovements: permission(action: \"core:continual-improvement:list\", id: $organizationId)\n  canListProcessingActivities: permission(action: \"core:processing-activity:list\", id: $organizationId)\n  canListStatesOfApplicability: permission(action: \"core:state-of-applicability:list\", id: $organizationId)\n  canListSnapshots: permission(action: \"core:snapshot:list\", id: $organizationId)\n  canGetTrustCenter: permission(action: \"core:trust-center:get\", id: $organizationId)\n  canUpdateOrganization: permission(action: \"iam:organization:update\", id: $organizationId)\n}\n"
  }
};
})();

(node as any).hash = "bea3f973a014e38e291ecd177da5c5f5";

export default node;
