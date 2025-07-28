/**
 * @generated SignedSource<<db6f17f33029746ad576e10564eb21f7>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type AuditState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type DocumentStatus = "DRAFT" | "PUBLISHED";
export type DocumentType = "ISMS" | "OTHER" | "POLICY";
export type VendorCategory = "ANALYTICS" | "CLOUD_MONITORING" | "CLOUD_PROVIDER" | "COLLABORATION" | "CUSTOMER_SUPPORT" | "DATA_STORAGE_AND_PROCESSING" | "DOCUMENT_MANAGEMENT" | "EMPLOYEE_MANAGEMENT" | "ENGINEERING" | "FINANCE" | "IDENTITY_PROVIDER" | "IT" | "MARKETING" | "OFFICE_OPERATIONS" | "OTHER" | "PASSWORD_MANAGEMENT" | "PRODUCT_AND_DESIGN" | "PROFESSIONAL_SERVICES" | "RECRUITING" | "SALES" | "SECURITY" | "VERSION_CONTROL";
export type PublicTrustCenterGraphQuery$variables = {
  slug: string;
};
export type PublicTrustCenterGraphQuery$data = {
  readonly trustCenterBySlug: {
    readonly active: boolean;
    readonly id: string;
    readonly organization: {
      readonly id: string;
      readonly logoUrl: string | null | undefined;
      readonly name: string;
    };
    readonly publicAudits: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly createdAt: any;
          readonly framework: {
            readonly name: string;
          };
          readonly id: string;
          readonly report: {
            readonly downloadUrl: string | null | undefined;
            readonly filename: string;
            readonly id: string;
          } | null | undefined;
          readonly reportUrl: string | null | undefined;
          readonly state: AuditState;
          readonly validFrom: any | null | undefined;
          readonly validUntil: any | null | undefined;
        };
      }>;
    };
    readonly publicDocuments: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly documentType: DocumentType;
          readonly id: string;
          readonly title: string;
          readonly versions: {
            readonly edges: ReadonlyArray<{
              readonly node: {
                readonly id: string;
                readonly status: DocumentStatus;
              };
            }>;
          };
        };
      }>;
    };
    readonly publicVendors: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly category: VendorCategory;
          readonly createdAt: any;
          readonly description: string | null | undefined;
          readonly id: string;
          readonly name: string;
          readonly privacyPolicyUrl: string | null | undefined;
          readonly websiteUrl: string | null | undefined;
        };
      }>;
    };
    readonly slug: string;
  } | null | undefined;
};
export type PublicTrustCenterGraphQuery = {
  response: PublicTrustCenterGraphQuery$data;
  variables: PublicTrustCenterGraphQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "slug"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "slug",
    "variableName": "slug"
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
  "name": "active",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "slug",
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
  "concreteType": "Organization",
  "kind": "LinkedField",
  "name": "organization",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    (v5/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "logoUrl",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v7 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  }
],
v8 = {
  "alias": null,
  "args": (v7/*: any*/),
  "concreteType": "DocumentConnection",
  "kind": "LinkedField",
  "name": "publicDocuments",
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
            (v2/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "title",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "documentType",
              "storageKey": null
            },
            {
              "alias": null,
              "args": [
                {
                  "kind": "Literal",
                  "name": "first",
                  "value": 1
                }
              ],
              "concreteType": "DocumentVersionConnection",
              "kind": "LinkedField",
              "name": "versions",
              "plural": false,
              "selections": [
                {
                  "alias": null,
                  "args": null,
                  "concreteType": "DocumentVersionEdge",
                  "kind": "LinkedField",
                  "name": "edges",
                  "plural": true,
                  "selections": [
                    {
                      "alias": null,
                      "args": null,
                      "concreteType": "DocumentVersion",
                      "kind": "LinkedField",
                      "name": "node",
                      "plural": false,
                      "selections": [
                        (v2/*: any*/),
                        {
                          "alias": null,
                          "args": null,
                          "kind": "ScalarField",
                          "name": "status",
                          "storageKey": null
                        }
                      ],
                      "storageKey": null
                    }
                  ],
                  "storageKey": null
                }
              ],
              "storageKey": "versions(first:1)"
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": "publicDocuments(first:100)"
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "validFrom",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "validUntil",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "state",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "createdAt",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "concreteType": "Report",
  "kind": "LinkedField",
  "name": "report",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "filename",
      "storageKey": null
    },
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "downloadUrl",
      "storageKey": null
    }
  ],
  "storageKey": null
},
v14 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "reportUrl",
  "storageKey": null
},
v15 = {
  "alias": null,
  "args": (v7/*: any*/),
  "concreteType": "VendorConnection",
  "kind": "LinkedField",
  "name": "publicVendors",
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
            (v2/*: any*/),
            (v5/*: any*/),
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
              "name": "description",
              "storageKey": null
            },
            (v12/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "websiteUrl",
              "storageKey": null
            },
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "privacyPolicyUrl",
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
  "storageKey": "publicVendors(first:100)"
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "PublicTrustCenterGraphQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "trustCenterBySlug",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v3/*: any*/),
          (v4/*: any*/),
          (v6/*: any*/),
          (v8/*: any*/),
          {
            "alias": null,
            "args": (v7/*: any*/),
            "concreteType": "AuditConnection",
            "kind": "LinkedField",
            "name": "publicAudits",
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
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Framework",
                        "kind": "LinkedField",
                        "name": "framework",
                        "plural": false,
                        "selections": [
                          (v5/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/),
                      (v10/*: any*/),
                      (v11/*: any*/),
                      (v12/*: any*/),
                      (v13/*: any*/),
                      (v14/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "publicAudits(first:100)"
          },
          (v15/*: any*/)
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
    "name": "PublicTrustCenterGraphQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "trustCenterBySlug",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v3/*: any*/),
          (v4/*: any*/),
          (v6/*: any*/),
          (v8/*: any*/),
          {
            "alias": null,
            "args": (v7/*: any*/),
            "concreteType": "AuditConnection",
            "kind": "LinkedField",
            "name": "publicAudits",
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
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Framework",
                        "kind": "LinkedField",
                        "name": "framework",
                        "plural": false,
                        "selections": [
                          (v5/*: any*/),
                          (v2/*: any*/)
                        ],
                        "storageKey": null
                      },
                      (v9/*: any*/),
                      (v10/*: any*/),
                      (v11/*: any*/),
                      (v12/*: any*/),
                      (v13/*: any*/),
                      (v14/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "publicAudits(first:100)"
          },
          (v15/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "8b16e3d34954bc44f90d907fbc73251b",
    "id": null,
    "metadata": {},
    "name": "PublicTrustCenterGraphQuery",
    "operationKind": "query",
    "text": "query PublicTrustCenterGraphQuery(\n  $slug: String!\n) {\n  trustCenterBySlug(slug: $slug) {\n    id\n    active\n    slug\n    organization {\n      id\n      name\n      logoUrl\n    }\n    publicDocuments(first: 100) {\n      edges {\n        node {\n          id\n          title\n          documentType\n          versions(first: 1) {\n            edges {\n              node {\n                id\n                status\n              }\n            }\n          }\n        }\n      }\n    }\n    publicAudits(first: 100) {\n      edges {\n        node {\n          id\n          framework {\n            name\n            id\n          }\n          validFrom\n          validUntil\n          state\n          createdAt\n          report {\n            id\n            filename\n            downloadUrl\n          }\n          reportUrl\n        }\n      }\n    }\n    publicVendors(first: 100) {\n      edges {\n        node {\n          id\n          name\n          category\n          description\n          createdAt\n          websiteUrl\n          privacyPolicyUrl\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ef092098026ef8420e8b58ec02274ad2";

export default node;
