/**
 * @generated SignedSource<<d72032da6df08b84ce9c7ec87b950d08>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorGraphNodeQuery$variables = {
  organizationId: string;
  vendorId: string;
};
export type VendorGraphNodeQuery$data = {
  readonly node: {
    readonly id?: string;
    readonly name?: string;
    readonly snapshotId?: string | null | undefined;
    readonly websiteUrl?: string | null | undefined;
    readonly " $fragmentSpreads": FragmentRefs<"VendorComplianceTabFragment" | "VendorContactsTabFragment" | "VendorOverviewTabBusinessAssociateAgreementFragment" | "VendorOverviewTabDataPrivacyAgreementFragment" | "VendorRiskAssessmentTabFragment" | "VendorServicesTabFragment" | "useVendorFormFragment">;
  };
  readonly viewer: {
    readonly user: {
      readonly id: string;
      readonly people: {
        readonly id: string;
      } | null | undefined;
    };
  };
};
export type VendorGraphNodeQuery = {
  response: VendorGraphNodeQuery$data;
  variables: VendorGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "organizationId"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "vendorId"
},
v2 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "vendorId"
  }
],
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
  "name": "snapshotId",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "websiteUrl",
  "storageKey": null
},
v7 = [
  (v3/*: any*/)
],
v8 = {
  "alias": null,
  "args": null,
  "concreteType": "User",
  "kind": "LinkedField",
  "name": "user",
  "plural": false,
  "selections": [
    (v3/*: any*/),
    {
      "alias": null,
      "args": [
        {
          "kind": "Variable",
          "name": "organizationId",
          "variableName": "organizationId"
        }
      ],
      "concreteType": "People",
      "kind": "LinkedField",
      "name": "people",
      "plural": false,
      "selections": (v7/*: any*/),
      "storageKey": null
    }
  ],
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v11 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  }
],
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "validUntil",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileUrl",
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endCursor",
  "storageKey": null
},
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasNextPage",
  "storageKey": null
},
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasPreviousPage",
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "startCursor",
  "storageKey": null
},
v19 = {
  "alias": null,
  "args": null,
  "concreteType": "PageInfo",
  "kind": "LinkedField",
  "name": "pageInfo",
  "plural": false,
  "selections": [
    (v15/*: any*/),
    (v16/*: any*/),
    (v17/*: any*/),
    (v18/*: any*/)
  ],
  "storageKey": null
},
v20 = {
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
},
v21 = [
  "orderBy"
],
v22 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v23 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v24 = [
  (v3/*: any*/),
  {
    "alias": null,
    "args": null,
    "kind": "ScalarField",
    "name": "fileName",
    "storageKey": null
  },
  (v13/*: any*/),
  {
    "alias": null,
    "args": null,
    "kind": "ScalarField",
    "name": "validFrom",
    "storageKey": null
  },
  (v12/*: any*/),
  (v22/*: any*/)
];
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "VendorGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "useVendorFormFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "VendorComplianceTabFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "VendorContactsTabFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "VendorServicesTabFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "VendorRiskAssessmentTabFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "VendorOverviewTabBusinessAssociateAgreementFragment"
              },
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "VendorOverviewTabDataPrivacyAgreementFragment"
              }
            ],
            "type": "Vendor",
            "abstractKey": null
          }
        ],
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v8/*: any*/)
        ],
        "storageKey": null
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
    "name": "VendorGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v9/*: any*/),
          (v3/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v10/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "category",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "statusPageUrl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "termsOfServiceUrl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "privacyPolicyUrl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "serviceLevelAgreementUrl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "dataProcessingAgreementUrl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "legalName",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "headquarterAddress",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "certifications",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "countries",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "securityPageUrl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "trustPageUrl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "People",
                "kind": "LinkedField",
                "name": "businessOwner",
                "plural": false,
                "selections": (v7/*: any*/),
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "People",
                "kind": "LinkedField",
                "name": "securityOwner",
                "plural": false,
                "selections": (v7/*: any*/),
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v11/*: any*/),
                "concreteType": "VendorComplianceReportConnection",
                "kind": "LinkedField",
                "name": "complianceReports",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "VendorComplianceReportEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "VendorComplianceReport",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v3/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "reportDate",
                            "storageKey": null
                          },
                          (v12/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "reportName",
                            "storageKey": null
                          },
                          (v13/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "fileSize",
                            "storageKey": null
                          },
                          (v9/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v14/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v19/*: any*/),
                  (v20/*: any*/)
                ],
                "storageKey": "complianceReports(first:50)"
              },
              {
                "alias": null,
                "args": (v11/*: any*/),
                "filters": (v21/*: any*/),
                "handle": "connection",
                "key": "VendorComplianceTabFragment_complianceReports",
                "kind": "LinkedHandle",
                "name": "complianceReports"
              },
              {
                "alias": null,
                "args": (v11/*: any*/),
                "concreteType": "VendorContactConnection",
                "kind": "LinkedField",
                "name": "contacts",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "VendorContactEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "VendorContact",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v3/*: any*/),
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
                            "name": "email",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "phone",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "role",
                            "storageKey": null
                          },
                          (v22/*: any*/),
                          (v23/*: any*/),
                          (v9/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v14/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v19/*: any*/),
                  (v20/*: any*/)
                ],
                "storageKey": "contacts(first:50)"
              },
              {
                "alias": null,
                "args": (v11/*: any*/),
                "filters": (v21/*: any*/),
                "handle": "connection",
                "key": "VendorContactsTabFragment_contacts",
                "kind": "LinkedHandle",
                "name": "contacts"
              },
              {
                "alias": null,
                "args": (v11/*: any*/),
                "concreteType": "VendorServiceConnection",
                "kind": "LinkedField",
                "name": "services",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "VendorServiceEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "VendorService",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v3/*: any*/),
                          (v5/*: any*/),
                          (v10/*: any*/),
                          (v22/*: any*/),
                          (v23/*: any*/),
                          (v9/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v14/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v19/*: any*/),
                  (v20/*: any*/)
                ],
                "storageKey": "services(first:50)"
              },
              {
                "alias": null,
                "args": (v11/*: any*/),
                "filters": (v21/*: any*/),
                "handle": "connection",
                "key": "VendorServicesTabFragment_services",
                "kind": "LinkedHandle",
                "name": "services"
              },
              {
                "alias": null,
                "args": (v11/*: any*/),
                "concreteType": "VendorRiskAssessmentConnection",
                "kind": "LinkedField",
                "name": "riskAssessments",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "VendorRiskAssessmentEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "VendorRiskAssessment",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v3/*: any*/),
                          (v22/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "expiresAt",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "dataSensitivity",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "businessImpact",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "notes",
                            "storageKey": null
                          },
                          (v9/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v14/*: any*/)
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
                      (v16/*: any*/),
                      (v15/*: any*/),
                      (v17/*: any*/),
                      (v18/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v20/*: any*/)
                ],
                "storageKey": "riskAssessments(first:50)"
              },
              {
                "alias": null,
                "args": (v11/*: any*/),
                "filters": (v21/*: any*/),
                "handle": "connection",
                "key": "VendorRiskAssessmentTabFragment_riskAssessments",
                "kind": "LinkedHandle",
                "name": "riskAssessments"
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorBusinessAssociateAgreement",
                "kind": "LinkedField",
                "name": "businessAssociateAgreement",
                "plural": false,
                "selections": (v24/*: any*/),
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorDataPrivacyAgreement",
                "kind": "LinkedField",
                "name": "dataPrivacyAgreement",
                "plural": false,
                "selections": (v24/*: any*/),
                "storageKey": null
              }
            ],
            "type": "Vendor",
            "abstractKey": null
          }
        ],
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "concreteType": "Viewer",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v8/*: any*/),
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "4a112f861864990f5022d279ba8d43d4",
    "id": null,
    "metadata": {},
    "name": "VendorGraphNodeQuery",
    "operationKind": "query",
    "text": "query VendorGraphNodeQuery(\n  $vendorId: ID!\n  $organizationId: ID!\n) {\n  node(id: $vendorId) {\n    __typename\n    ... on Vendor {\n      id\n      snapshotId\n      name\n      websiteUrl\n      ...useVendorFormFragment\n      ...VendorComplianceTabFragment\n      ...VendorContactsTabFragment\n      ...VendorServicesTabFragment\n      ...VendorRiskAssessmentTabFragment\n      ...VendorOverviewTabBusinessAssociateAgreementFragment\n      ...VendorOverviewTabDataPrivacyAgreementFragment\n    }\n    id\n  }\n  viewer {\n    user {\n      id\n      people(organizationId: $organizationId) {\n        id\n      }\n    }\n    id\n  }\n}\n\nfragment VendorComplianceTabFragment on Vendor {\n  complianceReports(first: 50) {\n    edges {\n      node {\n        id\n        ...VendorComplianceTabFragment_report\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment VendorComplianceTabFragment_report on VendorComplianceReport {\n  id\n  reportDate\n  validUntil\n  reportName\n  fileUrl\n  fileSize\n}\n\nfragment VendorContactsTabFragment on Vendor {\n  contacts(first: 50) {\n    edges {\n      node {\n        id\n        ...VendorContactsTabFragment_contact\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment VendorContactsTabFragment_contact on VendorContact {\n  id\n  fullName\n  email\n  phone\n  role\n  createdAt\n  updatedAt\n}\n\nfragment VendorOverviewTabBusinessAssociateAgreementFragment on Vendor {\n  businessAssociateAgreement {\n    id\n    fileName\n    fileUrl\n    validFrom\n    validUntil\n    createdAt\n  }\n}\n\nfragment VendorOverviewTabDataPrivacyAgreementFragment on Vendor {\n  dataPrivacyAgreement {\n    id\n    fileName\n    fileUrl\n    validFrom\n    validUntil\n    createdAt\n  }\n}\n\nfragment VendorRiskAssessmentTabFragment on Vendor {\n  id\n  riskAssessments(first: 50) {\n    edges {\n      node {\n        id\n        ...VendorRiskAssessmentTabFragment_assessment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment VendorRiskAssessmentTabFragment_assessment on VendorRiskAssessment {\n  id\n  createdAt\n  expiresAt\n  dataSensitivity\n  businessImpact\n  notes\n}\n\nfragment VendorServicesTabFragment on Vendor {\n  services(first: 50) {\n    edges {\n      node {\n        id\n        ...VendorServicesTabFragment_service\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment VendorServicesTabFragment_service on VendorService {\n  id\n  name\n  description\n  createdAt\n  updatedAt\n}\n\nfragment useVendorFormFragment on Vendor {\n  id\n  name\n  description\n  category\n  statusPageUrl\n  termsOfServiceUrl\n  privacyPolicyUrl\n  serviceLevelAgreementUrl\n  dataProcessingAgreementUrl\n  websiteUrl\n  legalName\n  headquarterAddress\n  certifications\n  countries\n  securityPageUrl\n  trustPageUrl\n  businessOwner {\n    id\n  }\n  securityOwner {\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "5b907e5cc3446745ff2414ee06839aed";

export default node;
