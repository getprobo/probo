/**
 * @generated SignedSource<<74757bf9a32552f6353eb888c9f663cd>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ElectronicSignatureStatus = "ACCEPTED" | "COMPLETED" | "FAILED" | "PENDING" | "PROCESSING";
export type TrustGraphCurrentQuery$variables = Record<PropertyKey, never>;
export type TrustGraphCurrentQuery$data = {
  readonly currentTrustCenter: {
    readonly audits: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"AuditRowFragment">;
        };
      }>;
    };
    readonly complianceFrameworks: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly framework: {
            readonly " $fragmentSpreads": FragmentRefs<"FrameworkBadgeFragment">;
          };
          readonly id: string;
        };
      }>;
    };
    readonly darkLogoFileUrl: string | null | undefined;
    readonly externalUrls: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly name: string;
          readonly url: string;
        };
      }>;
    };
    readonly id: string;
    readonly logoFileUrl: string | null | undefined;
    readonly nonDisclosureAgreement: {
      readonly fileName: string;
      readonly fileUrl: string;
      readonly viewerSignature: {
        readonly status: ElectronicSignatureStatus;
      } | null | undefined;
    } | null | undefined;
    readonly organization: {
      readonly description: string | null | undefined;
      readonly email: string | null | undefined;
      readonly headquarterAddress: string | null | undefined;
      readonly name: string;
      readonly websiteUrl: string | null | undefined;
    };
    readonly slug: string;
    readonly vendorInfo: {
      readonly totalCount: number;
    };
    readonly " $fragmentSpreads": FragmentRefs<"OverviewPageFragment">;
  };
};
export type TrustGraphCurrentQuery = {
  response: TrustGraphCurrentQuery$data;
  variables: TrustGraphCurrentQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "slug",
  "storageKey": null
},
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "logoFileUrl",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "darkLogoFileUrl",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileName",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileUrl",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "websiteUrl",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "headquarterAddress",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": [
    {
      "kind": "Literal",
      "name": "first",
      "value": 20
    }
  ],
  "concreteType": "ComplianceExternalURLConnection",
  "kind": "LinkedField",
  "name": "externalUrls",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "ComplianceExternalURLEdge",
      "kind": "LinkedField",
      "name": "edges",
      "plural": true,
      "selections": [
        {
          "alias": null,
          "args": null,
          "concreteType": "ComplianceExternalURL",
          "kind": "LinkedField",
          "name": "node",
          "plural": false,
          "selections": [
            (v0/*: any*/),
            (v7/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "url",
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": "externalUrls(first:20)"
},
v13 = {
  "alias": "vendorInfo",
  "args": [
    {
      "kind": "Literal",
      "name": "first",
      "value": 0
    }
  ],
  "concreteType": "VendorConnection",
  "kind": "LinkedField",
  "name": "vendors",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "totalCount",
      "storageKey": null
    }
  ],
  "storageKey": "vendors(first:0)"
},
v14 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  }
],
v15 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 5
  }
],
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "isUserAuthorized",
  "storageKey": null
},
v17 = {
  "alias": null,
  "args": null,
  "concreteType": "DocumentAccess",
  "kind": "LinkedField",
  "name": "access",
  "plural": false,
  "selections": [
    (v0/*: any*/),
    (v6/*: any*/)
  ],
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "concreteType": "Framework",
  "kind": "LinkedField",
  "name": "framework",
  "plural": false,
  "selections": [
    (v0/*: any*/),
    (v7/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "lightLogoURL",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "darkLogoURL",
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustGraphCurrentQuery",
    "selections": [
      {
        "kind": "RequiredField",
        "field": {
          "alias": null,
          "args": null,
          "concreteType": "TrustCenter",
          "kind": "LinkedField",
          "name": "currentTrustCenter",
          "plural": false,
          "selections": [
            (v0/*: any*/),
            (v1/*: any*/),
            (v2/*: any*/),
            (v3/*: any*/),
            {
              "alias": null,
              "args": null,
              "concreteType": "NonDisclosureAgreement",
              "kind": "LinkedField",
              "name": "nonDisclosureAgreement",
              "plural": false,
              "selections": [
                (v4/*: any*/),
                (v5/*: any*/),
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "ElectronicSignature",
                  "kind": "LinkedField",
                  "name": "viewerSignature",
                  "plural": false,
                  "selections": [
                    (v6/*: any*/)
                  ],
                  "storageKey": null
                }
              ],
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "concreteType": "Organization",
              "kind": "LinkedField",
              "name": "organization",
              "plural": false,
              "selections": [
                (v7/*: any*/),
                (v8/*: any*/),
                (v9/*: any*/),
                (v10/*: any*/),
                (v11/*: any*/)
              ],
              "storageKey": null
            },
            (v12/*: any*/),
            {
              "args": null,
              "kind": "FragmentSpread",
              "name": "OverviewPageFragment"
            },
            (v13/*: any*/),
            {
              "alias": null,
              "args": (v14/*: any*/),
              "concreteType": "AuditConnection",
              "kind": "LinkedField",
              "name": "audits",
              "plural": false,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "AuditEdge",
                  "kind": "LinkedField",
                  "name": "edges",
                  "plural": true,
                  "selections": [
                    {
                      "alias": null,
                      "args": null,
                      "concreteType": "Audit",
                      "kind": "LinkedField",
                      "name": "node",
                      "plural": false,
                      "selections": [
                        (v0/*: any*/),
                        {
                          "args": null,
                          "kind": "FragmentSpread",
                          "name": "AuditRowFragment"
                        }
                      ],
                      "storageKey": null
                    }
                  ],
                  "storageKey": null
                }
              ],
              "storageKey": "audits(first:50)"
            },
            {
              "alias": null,
              "args": (v14/*: any*/),
              "concreteType": "ComplianceFrameworkConnection",
              "kind": "LinkedField",
              "name": "complianceFrameworks",
              "plural": false,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "ComplianceFrameworkEdge",
                  "kind": "LinkedField",
                  "name": "edges",
                  "plural": true,
                  "selections": [
                    {
                      "alias": null,
                      "args": null,
                      "concreteType": "ComplianceFramework",
                      "kind": "LinkedField",
                      "name": "node",
                      "plural": false,
                      "selections": [
                        (v0/*: any*/),
                        {
                          "alias": null,
                          "args": null,
                          "concreteType": "Framework",
                          "kind": "LinkedField",
                          "name": "framework",
                          "plural": false,
                          "selections": [
                            {
                              "args": null,
                              "kind": "FragmentSpread",
                              "name": "FrameworkBadgeFragment"
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
              "storageKey": "complianceFrameworks(first:50)"
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
    "argumentDefinitions": [],
    "kind": "Operation",
    "name": "TrustGraphCurrentQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "currentTrustCenter",
        "plural": false,
        "selections": [
          (v0/*: any*/),
          (v1/*: any*/),
          (v2/*: any*/),
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "NonDisclosureAgreement",
            "kind": "LinkedField",
            "name": "nonDisclosureAgreement",
            "plural": false,
            "selections": [
              (v4/*: any*/),
              (v5/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "ElectronicSignature",
                "kind": "LinkedField",
                "name": "viewerSignature",
                "plural": false,
                "selections": [
                  (v6/*: any*/),
                  (v0/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "Organization",
            "kind": "LinkedField",
            "name": "organization",
            "plural": false,
            "selections": [
              (v7/*: any*/),
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v0/*: any*/)
            ],
            "storageKey": null
          },
          (v12/*: any*/),
          {
            "alias": null,
            "args": [
              {
                "kind": "Literal",
                "name": "first",
                "value": 14
              }
            ],
            "concreteType": "TrustCenterReferenceConnection",
            "kind": "LinkedField",
            "name": "references",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "TrustCenterReferenceEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TrustCenterReference",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      (v7/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "logoUrl",
                        "storageKey": null
                      },
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "references(first:14)"
          },
          {
            "alias": null,
            "args": [
              {
                "kind": "Literal",
                "name": "first",
                "value": 3
              }
            ],
            "concreteType": "VendorConnection",
            "kind": "LinkedField",
            "name": "vendors",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Vendor",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "countries",
                        "storageKey": null
                      },
                      (v7/*: any*/),
                      (v8/*: any*/),
                      (v9/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "vendors(first:3)"
          },
          {
            "alias": null,
            "args": (v15/*: any*/),
            "concreteType": "DocumentConnection",
            "kind": "LinkedField",
            "name": "documents",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "DocumentEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Document",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "title",
                        "storageKey": null
                      },
                      (v16/*: any*/),
                      (v17/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "documentType",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "documents(first:5)"
          },
          {
            "alias": null,
            "args": (v15/*: any*/),
            "concreteType": "TrustCenterFileConnection",
            "kind": "LinkedField",
            "name": "trustCenterFiles",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "TrustCenterFileEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TrustCenterFile",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "category",
                        "storageKey": null
                      },
                      (v7/*: any*/),
                      (v16/*: any*/),
                      (v17/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "trustCenterFiles(first:5)"
          },
          (v13/*: any*/),
          {
            "alias": null,
            "args": (v14/*: any*/),
            "concreteType": "AuditConnection",
            "kind": "LinkedField",
            "name": "audits",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "AuditEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "Audit",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      (v7/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Report",
                        "kind": "LinkedField",
                        "name": "report",
                        "plural": false,
                        "selections": [
                          (v0/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "kind": "ScalarField",
                            "name": "filename",
                            "storageKey": null
                          },
                          (v16/*: any*/),
                          (v17/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v18/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "audits(first:50)"
          },
          {
            "alias": null,
            "args": (v14/*: any*/),
            "concreteType": "ComplianceFrameworkConnection",
            "kind": "LinkedField",
            "name": "complianceFrameworks",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "ComplianceFrameworkEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ComplianceFramework",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v0/*: any*/),
                      (v18/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "complianceFrameworks(first:50)"
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "96e75937f385e17b760440e001cefc47",
    "id": null,
    "metadata": {},
    "name": "TrustGraphCurrentQuery",
    "operationKind": "query",
    "text": "query TrustGraphCurrentQuery {\n  currentTrustCenter {\n    id\n    slug\n    logoFileUrl\n    darkLogoFileUrl\n    nonDisclosureAgreement {\n      fileName\n      fileUrl\n      viewerSignature {\n        status\n        id\n      }\n    }\n    organization {\n      name\n      description\n      websiteUrl\n      email\n      headquarterAddress\n      id\n    }\n    externalUrls(first: 20) {\n      edges {\n        node {\n          id\n          name\n          url\n        }\n      }\n    }\n    ...OverviewPageFragment\n    vendorInfo: vendors(first: 0) {\n      totalCount\n    }\n    audits(first: 50) {\n      edges {\n        node {\n          id\n          ...AuditRowFragment\n        }\n      }\n    }\n    complianceFrameworks(first: 50) {\n      edges {\n        node {\n          id\n          framework {\n            ...FrameworkBadgeFragment\n            id\n          }\n        }\n      }\n    }\n  }\n}\n\nfragment AuditRowFragment on Audit {\n  name\n  report {\n    id\n    filename\n    isUserAuthorized\n    access {\n      id\n      status\n    }\n  }\n  framework {\n    id\n    name\n    lightLogoURL\n    darkLogoURL\n  }\n}\n\nfragment DocumentRowFragment on Document {\n  id\n  title\n  isUserAuthorized\n  access {\n    id\n    status\n  }\n}\n\nfragment FrameworkBadgeFragment on Framework {\n  id\n  name\n  lightLogoURL\n  darkLogoURL\n}\n\nfragment OverviewPageFragment on TrustCenter {\n  references(first: 14) {\n    edges {\n      node {\n        id\n        name\n        logoUrl\n        websiteUrl\n      }\n    }\n  }\n  vendors(first: 3) {\n    edges {\n      node {\n        id\n        countries\n        ...VendorRowFragment\n      }\n    }\n  }\n  documents(first: 5) {\n    edges {\n      node {\n        id\n        ...DocumentRowFragment\n        documentType\n      }\n    }\n  }\n  trustCenterFiles(first: 5) {\n    edges {\n      node {\n        id\n        category\n        ...TrustCenterFileRowFragment\n      }\n    }\n  }\n}\n\nfragment TrustCenterFileRowFragment on TrustCenterFile {\n  id\n  name\n  isUserAuthorized\n  access {\n    id\n    status\n  }\n}\n\nfragment VendorRowFragment on Vendor {\n  name\n  description\n  websiteUrl\n  countries\n}\n"
  }
};
})();

(node as any).hash = "ca6f04fecf9bce92407f33e8c0e4bd03";

export default node;
