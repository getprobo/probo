/**
 * @generated SignedSource<<94e31e9e1e26ee4a548dc90b4f13b9ca>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type VendorGraphNodeQuery$variables = {
  vendorId: string;
};
export type VendorGraphNodeQuery$data = {
  readonly node: {
    readonly canAssess?: boolean;
    readonly canCreateContact?: boolean;
    readonly canCreateRiskAssessment?: boolean;
    readonly canCreateService?: boolean;
    readonly canDelete?: boolean;
    readonly canUpdate?: boolean;
    readonly canUploadBAA?: boolean;
    readonly canUploadComplianceReport?: boolean;
    readonly canUploadDPA?: boolean;
    readonly id: string;
    readonly name?: string;
    readonly snapshotId?: string | null | undefined;
    readonly websiteUrl?: string | null | undefined;
    readonly " $fragmentSpreads": FragmentRefs<"VendorComplianceTabFragment" | "VendorContactsTabFragment" | "VendorOverviewTabBusinessAssociateAgreementFragment" | "VendorOverviewTabDataPrivacyAgreementFragment" | "VendorRiskAssessmentTabFragment" | "VendorServicesTabFragment" | "useVendorFormFragment">;
  };
  readonly viewer: {
    readonly id: string;
  };
};
export type VendorGraphNodeQuery = {
  response: VendorGraphNodeQuery$data;
  variables: VendorGraphNodeQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "vendorId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "vendorId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "snapshotId",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "websiteUrl",
  "storageKey": null
},
v6 = {
  "alias": "canAssess",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:vendor:assess"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:vendor:assess\")"
},
v7 = {
  "alias": "canUpdate",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:vendor:update"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:vendor:update\")"
},
v8 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:vendor:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:vendor:delete\")"
},
v9 = {
  "alias": "canUploadComplianceReport",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:vendor-compliance-report:upload"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:vendor-compliance-report:upload\")"
},
v10 = {
  "alias": "canCreateRiskAssessment",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:vendor-risk-assessment:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:vendor-risk-assessment:create\")"
},
v11 = {
  "alias": "canCreateContact",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:vendor-contact:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:vendor-contact:create\")"
},
v12 = {
  "alias": "canCreateService",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:vendor-service:create"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:vendor-service:create\")"
},
v13 = {
  "alias": "canUploadBAA",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:vendor-business-associate-agreement:upload"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:vendor-business-associate-agreement:upload\")"
},
v14 = {
  "alias": "canUploadDPA",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:vendor-data-privacy-agreement:upload"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:vendor-data-privacy-agreement:upload\")"
},
v15 = [
  (v2/*: any*/)
],
v16 = {
  "alias": null,
  "args": null,
  "concreteType": "Viewer",
  "kind": "LinkedField",
  "name": "viewer",
  "plural": false,
  "selections": (v15/*: any*/),
  "storageKey": null
},
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "__typename",
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v19 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  }
],
v20 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "validUntil",
  "storageKey": null
},
v21 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileName",
  "storageKey": null
},
v22 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "cursor",
  "storageKey": null
},
v23 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "endCursor",
  "storageKey": null
},
v24 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasNextPage",
  "storageKey": null
},
v25 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasPreviousPage",
  "storageKey": null
},
v26 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "startCursor",
  "storageKey": null
},
v27 = {
  "alias": null,
  "args": null,
  "concreteType": "PageInfo",
  "kind": "LinkedField",
  "name": "pageInfo",
  "plural": false,
  "selections": [
    (v23/*: any*/),
    (v24/*: any*/),
    (v25/*: any*/),
    (v26/*: any*/)
  ],
  "storageKey": null
},
v28 = {
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
v29 = [
  "orderBy"
],
v30 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v31 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "updatedAt",
  "storageKey": null
},
v32 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileUrl",
  "storageKey": null
},
v33 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "validFrom",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "VendorGraphNodeQuery",
    "selections": [
      {
        "alias": null,
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
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
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
      (v16/*: any*/)
    ],
    "type": "Query",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "VendorGraphNodeQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          (v17/*: any*/),
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              (v3/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              (v6/*: any*/),
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              (v14/*: any*/),
              (v18/*: any*/),
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
                "concreteType": "Profile",
                "kind": "LinkedField",
                "name": "businessOwner",
                "plural": false,
                "selections": (v15/*: any*/),
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "Profile",
                "kind": "LinkedField",
                "name": "securityOwner",
                "plural": false,
                "selections": (v15/*: any*/),
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v19/*: any*/),
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
                          (v2/*: any*/),
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:vendor-compliance-report:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:vendor-compliance-report:delete\")"
                          },
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "reportDate",
                            "storageKey": null
                          },
                          (v20/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "reportName",
                            "storageKey": null
                          },
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "File",
                            "kind": "LinkedField",
                            "name": "file",
                            "plural": false,
                            "selections": [
                              (v21/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "mimeType",
                                "storageKey": null
                              },
                              {
                                "alias": null,
                                "args": null,
                                "kind": "ScalarField",
                                "name": "size",
                                "storageKey": null
                              },
                              (v2/*: any*/)
                            ],
                            "storageKey": null
                          },
                          (v17/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v22/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v27/*: any*/),
                  (v28/*: any*/)
                ],
                "storageKey": "complianceReports(first:50)"
              },
              {
                "alias": null,
                "args": (v19/*: any*/),
                "filters": (v29/*: any*/),
                "handle": "connection",
                "key": "VendorComplianceTabFragment_complianceReports",
                "kind": "LinkedHandle",
                "name": "complianceReports"
              },
              {
                "alias": null,
                "args": (v19/*: any*/),
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
                          (v2/*: any*/),
                          {
                            "alias": "canUpdate",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:vendor-contact:update"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:vendor-contact:update\")"
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:vendor-contact:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:vendor-contact:delete\")"
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
                          (v30/*: any*/),
                          (v31/*: any*/),
                          (v17/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v22/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v27/*: any*/),
                  (v28/*: any*/)
                ],
                "storageKey": "contacts(first:50)"
              },
              {
                "alias": null,
                "args": (v19/*: any*/),
                "filters": (v29/*: any*/),
                "handle": "connection",
                "key": "VendorContactsTabFragment_contacts",
                "kind": "LinkedHandle",
                "name": "contacts"
              },
              {
                "alias": null,
                "args": (v19/*: any*/),
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
                          (v2/*: any*/),
                          {
                            "alias": "canUpdate",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:vendor-service:update"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:vendor-service:update\")"
                          },
                          {
                            "alias": "canDelete",
                            "args": [
                              {
                                "kind": "Literal",
                                "name": "action",
                                "value": "core:vendor-service:delete"
                              }
                            ],
                            "kind": "ScalarField",
                            "name": "permission",
                            "storageKey": "permission(action:\"core:vendor-service:delete\")"
                          },
                          (v4/*: any*/),
                          (v18/*: any*/),
                          (v30/*: any*/),
                          (v31/*: any*/),
                          (v17/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v22/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v27/*: any*/),
                  (v28/*: any*/)
                ],
                "storageKey": "services(first:50)"
              },
              {
                "alias": null,
                "args": (v19/*: any*/),
                "filters": (v29/*: any*/),
                "handle": "connection",
                "key": "VendorServicesTabFragment_services",
                "kind": "LinkedHandle",
                "name": "services"
              },
              {
                "alias": null,
                "args": (v19/*: any*/),
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
                          (v2/*: any*/),
                          (v30/*: any*/),
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
                          (v17/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v22/*: any*/)
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
                      (v24/*: any*/),
                      (v23/*: any*/),
                      (v25/*: any*/),
                      (v26/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v28/*: any*/)
                ],
                "storageKey": "riskAssessments(first:50)"
              },
              {
                "alias": null,
                "args": (v19/*: any*/),
                "filters": (v29/*: any*/),
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
                "selections": [
                  (v2/*: any*/),
                  (v21/*: any*/),
                  (v32/*: any*/),
                  (v33/*: any*/),
                  (v20/*: any*/),
                  (v30/*: any*/),
                  {
                    "alias": "canUpdate",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:vendor-business-associate-agreement:update"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:vendor-business-associate-agreement:update\")"
                  },
                  {
                    "alias": "canDelete",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:vendor-business-associate-agreement:delete"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:vendor-business-associate-agreement:delete\")"
                  }
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorDataPrivacyAgreement",
                "kind": "LinkedField",
                "name": "dataPrivacyAgreement",
                "plural": false,
                "selections": [
                  (v2/*: any*/),
                  (v21/*: any*/),
                  (v32/*: any*/),
                  (v33/*: any*/),
                  (v20/*: any*/),
                  (v30/*: any*/),
                  {
                    "alias": "canUpdate",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:vendor-data-privacy-agreement:update"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:vendor-data-privacy-agreement:update\")"
                  },
                  {
                    "alias": "canDelete",
                    "args": [
                      {
                        "kind": "Literal",
                        "name": "action",
                        "value": "core:vendor-data-privacy-agreement:delete"
                      }
                    ],
                    "kind": "ScalarField",
                    "name": "permission",
                    "storageKey": "permission(action:\"core:vendor-data-privacy-agreement:delete\")"
                  }
                ],
                "storageKey": null
              }
            ],
            "type": "Vendor",
            "abstractKey": null
          }
        ],
        "storageKey": null
      },
      (v16/*: any*/)
    ]
  },
  "params": {
    "cacheID": "2e729f8d534ac8b2c3ed86175ef62da7",
    "id": null,
    "metadata": {},
    "name": "VendorGraphNodeQuery",
    "operationKind": "query",
    "text": "query VendorGraphNodeQuery(\n  $vendorId: ID!\n) {\n  node(id: $vendorId) {\n    __typename\n    id\n    ... on Vendor {\n      snapshotId\n      name\n      websiteUrl\n      canAssess: permission(action: \"core:vendor:assess\")\n      canUpdate: permission(action: \"core:vendor:update\")\n      canDelete: permission(action: \"core:vendor:delete\")\n      canUploadComplianceReport: permission(action: \"core:vendor-compliance-report:upload\")\n      canCreateRiskAssessment: permission(action: \"core:vendor-risk-assessment:create\")\n      canCreateContact: permission(action: \"core:vendor-contact:create\")\n      canCreateService: permission(action: \"core:vendor-service:create\")\n      canUploadBAA: permission(action: \"core:vendor-business-associate-agreement:upload\")\n      canUploadDPA: permission(action: \"core:vendor-data-privacy-agreement:upload\")\n      ...useVendorFormFragment\n      ...VendorComplianceTabFragment\n      ...VendorContactsTabFragment\n      ...VendorServicesTabFragment\n      ...VendorRiskAssessmentTabFragment\n      ...VendorOverviewTabBusinessAssociateAgreementFragment\n      ...VendorOverviewTabDataPrivacyAgreementFragment\n    }\n  }\n  viewer {\n    id\n  }\n}\n\nfragment VendorComplianceTabFragment on Vendor {\n  complianceReports(first: 50) {\n    edges {\n      node {\n        id\n        canDelete: permission(action: \"core:vendor-compliance-report:delete\")\n        ...VendorComplianceTabFragment_report\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment VendorComplianceTabFragment_report on VendorComplianceReport {\n  id\n  reportDate\n  validUntil\n  reportName\n  file {\n    fileName\n    mimeType\n    size\n    id\n  }\n  canDelete: permission(action: \"core:vendor-compliance-report:delete\")\n}\n\nfragment VendorContactsTabFragment on Vendor {\n  contacts(first: 50) {\n    edges {\n      node {\n        id\n        canUpdate: permission(action: \"core:vendor-contact:update\")\n        canDelete: permission(action: \"core:vendor-contact:delete\")\n        ...VendorContactsTabFragment_contact\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment VendorContactsTabFragment_contact on VendorContact {\n  id\n  fullName\n  email\n  phone\n  role\n  createdAt\n  updatedAt\n  canUpdate: permission(action: \"core:vendor-contact:update\")\n  canDelete: permission(action: \"core:vendor-contact:delete\")\n}\n\nfragment VendorOverviewTabBusinessAssociateAgreementFragment on Vendor {\n  businessAssociateAgreement {\n    id\n    fileName\n    fileUrl\n    validFrom\n    validUntil\n    createdAt\n    canUpdate: permission(action: \"core:vendor-business-associate-agreement:update\")\n    canDelete: permission(action: \"core:vendor-business-associate-agreement:delete\")\n  }\n}\n\nfragment VendorOverviewTabDataPrivacyAgreementFragment on Vendor {\n  dataPrivacyAgreement {\n    id\n    fileName\n    fileUrl\n    validFrom\n    validUntil\n    createdAt\n    canUpdate: permission(action: \"core:vendor-data-privacy-agreement:update\")\n    canDelete: permission(action: \"core:vendor-data-privacy-agreement:delete\")\n  }\n}\n\nfragment VendorRiskAssessmentTabFragment on Vendor {\n  id\n  riskAssessments(first: 50) {\n    edges {\n      node {\n        id\n        ...VendorRiskAssessmentTabFragment_assessment\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      hasNextPage\n      endCursor\n      hasPreviousPage\n      startCursor\n    }\n  }\n}\n\nfragment VendorRiskAssessmentTabFragment_assessment on VendorRiskAssessment {\n  id\n  createdAt\n  expiresAt\n  dataSensitivity\n  businessImpact\n  notes\n}\n\nfragment VendorServicesTabFragment on Vendor {\n  services(first: 50) {\n    edges {\n      node {\n        id\n        canUpdate: permission(action: \"core:vendor-service:update\")\n        canDelete: permission(action: \"core:vendor-service:delete\")\n        ...VendorServicesTabFragment_service\n        __typename\n      }\n      cursor\n    }\n    pageInfo {\n      endCursor\n      hasNextPage\n      hasPreviousPage\n      startCursor\n    }\n  }\n  id\n}\n\nfragment VendorServicesTabFragment_service on VendorService {\n  id\n  name\n  description\n  createdAt\n  updatedAt\n  canUpdate: permission(action: \"core:vendor-service:update\")\n  canDelete: permission(action: \"core:vendor-service:delete\")\n}\n\nfragment useVendorFormFragment on Vendor {\n  id\n  name\n  description\n  category\n  statusPageUrl\n  termsOfServiceUrl\n  privacyPolicyUrl\n  serviceLevelAgreementUrl\n  dataProcessingAgreementUrl\n  websiteUrl\n  legalName\n  headquarterAddress\n  certifications\n  countries\n  securityPageUrl\n  trustPageUrl\n  businessOwner {\n    id\n  }\n  securityOwner {\n    id\n  }\n}\n"
  }
};
})();

(node as any).hash = "4de8ce39d2595e8b9866617badc36cae";

export default node;
