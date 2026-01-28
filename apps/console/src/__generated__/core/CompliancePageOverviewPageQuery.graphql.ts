/**
 * @generated SignedSource<<18bb7ea1b507c374c33d9f4f7a56db2d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type CompliancePageOverviewPageQuery$variables = {
  organizationId: string;
};
export type CompliancePageOverviewPageQuery$data = {
  readonly organization: {
    readonly compliancePage?: {
      readonly canGetNDA: boolean;
    } | null | undefined;
    readonly " $fragmentSpreads": FragmentRefs<"CompliancePageNDASectionFragment" | "CompliancePageSlackSectionFragment" | "CompliancePageStatusSectionFragment">;
  };
};
export type CompliancePageOverviewPageQuery = {
  response: CompliancePageOverviewPageQuery$data;
  variables: CompliancePageOverviewPageQuery$variables;
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
  "alias": "canGetNDA",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:trust-center:get-nda"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:trust-center:get-nda\")"
},
v3 = {
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
    "name": "CompliancePageOverviewPageQuery",
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
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": "compliancePage",
                "args": null,
                "concreteType": "TrustCenter",
                "kind": "LinkedField",
                "name": "trustCenter",
                "plural": false,
                "selections": [
                  (v2/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "CompliancePageStatusSectionFragment"
          },
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "CompliancePageNDASectionFragment"
          },
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "CompliancePageSlackSectionFragment"
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "CompliancePageOverviewPageQuery",
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
                "alias": "compliancePage",
                "args": null,
                "concreteType": "TrustCenter",
                "kind": "LinkedField",
                "name": "trustCenter",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
                  (v3/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "active",
                    "storageKey": null
                  },
                  {
                    "alias": "canUpdate",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:trust-center:update"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:trust-center:update\")"
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "ndaFileName",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "ndaFileUrl",
                    "storageKey": null
                  },
                  {
                    "alias": "canUploadNDA",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:trust-center:upload-nda"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:trust-center:upload-nda\")"
                  },
                  {
                    "alias": "canDeleteNDA",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:trust-center:delete-nda"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:trust-center:delete-nda\")"
                  }
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "CustomDomain",
                "kind": "LinkedField",
                "name": "customDomain",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "domain",
                    "storageKey": null
                  },
                  (v3/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": [
                  {
                    "kind": "Literal",
                    "name": "first",
                    "value": 100
                  }
                ],
                "concreteType": "SlackConnectionConnection",
                "kind": "LinkedField",
                "name": "slackConnections",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "SlackConnectionEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "SlackConnection",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v3/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "channel",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "createdAt",
                            "storageKey": null
                          }
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "slackConnections(first:100)"
              }
            ],
            "type": "Organization",
            "abstractKey": null
          },
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "0041132d3ac46a3193d9acbd6f27bdf5",
    "id": null,
    "metadata": {},
    "name": "CompliancePageOverviewPageQuery",
    "operationKind": "query",
    "text": "query CompliancePageOverviewPageQuery(\n  $organizationId: ID!\n) {\n  organization: node(id: $organizationId) {\n    __typename\n    ... on Organization {\n      compliancePage: trustCenter {\n        canGetNDA: permission(action: \"core:trust-center:get-nda\")\n        id\n      }\n    }\n    ...CompliancePageStatusSectionFragment\n    ...CompliancePageNDASectionFragment\n    ...CompliancePageSlackSectionFragment\n    id\n  }\n}\n\nfragment CompliancePageNDASectionFragment on Organization {\n  compliancePage: trustCenter {\n    id\n    ndaFileName\n    ndaFileUrl\n    canUpdate: permission(action: \"core:trust-center:update\")\n    canUploadNDA: permission(action: \"core:trust-center:upload-nda\")\n    canDeleteNDA: permission(action: \"core:trust-center:delete-nda\")\n  }\n}\n\nfragment CompliancePageSlackSectionFragment on Organization {\n  compliancePage: trustCenter {\n    canUpdate: permission(action: \"core:trust-center:update\")\n    id\n  }\n  slackConnections(first: 100) {\n    edges {\n      node {\n        id\n        channel\n        createdAt\n      }\n    }\n  }\n}\n\nfragment CompliancePageStatusSectionFragment on Organization {\n  customDomain {\n    domain\n    id\n  }\n  compliancePage: trustCenter {\n    id\n    active\n    canUpdate: permission(action: \"core:trust-center:update\")\n  }\n}\n"
  }
};
})();

(node as any).hash = "689133b7f1c380266e6f15a60c527c91";

export default node;
