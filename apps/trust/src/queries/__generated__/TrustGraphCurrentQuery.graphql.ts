/**
 * @generated SignedSource<<27a8919eab545c008504b2de2a7cdbd0>>
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
    readonly darkLogoFileUrl: string | null | undefined;
    readonly id: string;
    readonly isViewerMember: boolean;
    readonly logoFileUrl: string | null | undefined;
    readonly nonDisclosureAgreement: {
      readonly fileName: string;
      readonly fileUrl: string;
      readonly viewerSignature: {
        readonly status: ElectronicSignatureStatus;
      } | null | undefined;
    } | null | undefined;
    readonly organization: {
      readonly name: string;
    };
    readonly slug: string;
    readonly vendorInfo: {
      readonly totalCount: number;
    };
    readonly " $fragmentSpreads": FragmentRefs<"OrganizationSidebarFragment" | "OverviewPageFragment">;
  } | null | undefined;
  readonly viewer: {
    readonly email: any;
    readonly fullName: string;
  } | null | undefined;
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
  "name": "email",
  "storageKey": null
},
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
},
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
  "name": "slug",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "isViewerMember",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "logoFileUrl",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "darkLogoFileUrl",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileName",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileUrl",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v11 = {
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
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "websiteUrl",
  "storageKey": null
},
v14 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  }
],
v15 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "isUserAuthorized",
  "storageKey": null
},
v16 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasUserRequestedAccess",
  "storageKey": null
},
v17 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 5
  }
];
return {
  "fragment": {
    "argumentDefinitions": [],
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustGraphCurrentQuery",
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v0/*: any*/),
          (v1/*: any*/)
        ],
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "currentTrustCenter",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v3/*: any*/),
          (v4/*: any*/),
          (v5/*: any*/),
          (v6/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "NonDisclosureAgreement",
            "kind": "LinkedField",
            "name": "nonDisclosureAgreement",
            "plural": false,
            "selections": [
              (v7/*: any*/),
              (v8/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "ElectronicSignature",
                "kind": "LinkedField",
                "name": "viewerSignature",
                "plural": false,
                "selections": [
                  (v9/*: any*/)
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
              (v10/*: any*/)
            ],
            "storageKey": null
          },
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "OrganizationSidebarFragment"
          },
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "OverviewPageFragment"
          },
          (v11/*: any*/)
        ],
        "storageKey": null
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
        "concreteType": "Identity",
        "kind": "LinkedField",
        "name": "viewer",
        "plural": false,
        "selections": [
          (v0/*: any*/),
          (v1/*: any*/),
          (v2/*: any*/)
        ],
        "storageKey": null
      },
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "currentTrustCenter",
        "plural": false,
        "selections": [
          (v2/*: any*/),
          (v3/*: any*/),
          (v4/*: any*/),
          (v5/*: any*/),
          (v6/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "NonDisclosureAgreement",
            "kind": "LinkedField",
            "name": "nonDisclosureAgreement",
            "plural": false,
            "selections": [
              (v7/*: any*/),
              (v8/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "ElectronicSignature",
                "kind": "LinkedField",
                "name": "viewerSignature",
                "plural": false,
                "selections": [
                  (v9/*: any*/),
                  (v2/*: any*/)
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
              (v10/*: any*/),
              (v2/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              (v0/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "headquarterAddress",
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": (v14/*: any*/),
            "concreteType": "ComplianceBadgeConnection",
            "kind": "LinkedField",
            "name": "complianceBadges",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "ComplianceBadgeEdge",
                "kind": "LinkedField",
                "name": "edges",
                "plural": true,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ComplianceBadge",
                    "kind": "LinkedField",
                    "name": "node",
                    "plural": false,
                    "selections": [
                      (v2/*: any*/),
                      (v10/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "iconUrl",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "complianceBadges(first:50)"
          },
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
                      (v2/*: any*/),
                      {
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
                          (v15/*: any*/),
                          (v16/*: any*/)
                        ],
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "Framework",
                        "kind": "LinkedField",
                        "name": "framework",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v10/*: any*/),
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
                      (v2/*: any*/),
                      (v10/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "logoUrl",
                        "storageKey": null
                      },
                      (v13/*: any*/)
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
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "countries",
                        "storageKey": null
                      },
                      (v10/*: any*/),
                      (v12/*: any*/),
                      (v13/*: any*/)
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
            "args": (v17/*: any*/),
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
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "title",
                        "storageKey": null
                      },
                      (v15/*: any*/),
                      (v16/*: any*/),
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
            "args": (v17/*: any*/),
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
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "category",
                        "storageKey": null
                      },
                      (v10/*: any*/),
                      (v15/*: any*/),
                      (v16/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "trustCenterFiles(first:5)"
          },
          (v11/*: any*/)
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "a25c632bfd8562a8778394691d7bbebd",
    "id": null,
    "metadata": {},
    "name": "TrustGraphCurrentQuery",
    "operationKind": "query",
    "text": "query TrustGraphCurrentQuery {\n  viewer {\n    email\n    fullName\n    id\n  }\n  currentTrustCenter {\n    id\n    slug\n    isViewerMember\n    logoFileUrl\n    darkLogoFileUrl\n    nonDisclosureAgreement {\n      fileName\n      fileUrl\n      viewerSignature {\n        status\n        id\n      }\n    }\n    organization {\n      name\n      id\n    }\n    ...OrganizationSidebarFragment\n    ...OverviewPageFragment\n    vendorInfo: vendors(first: 0) {\n      totalCount\n    }\n  }\n}\n\nfragment AuditRowFragment on Audit {\n  report {\n    id\n    filename\n    isUserAuthorized\n    hasUserRequestedAccess\n  }\n  framework {\n    id\n    name\n    lightLogoURL\n    darkLogoURL\n  }\n}\n\nfragment DocumentRowFragment on Document {\n  id\n  title\n  isUserAuthorized\n  hasUserRequestedAccess\n}\n\nfragment OrganizationSidebarFragment on TrustCenter {\n  logoFileUrl\n  darkLogoFileUrl\n  organization {\n    name\n    description\n    websiteUrl\n    email\n    headquarterAddress\n    id\n  }\n  complianceBadges(first: 50) {\n    edges {\n      node {\n        id\n        name\n        iconUrl\n      }\n    }\n  }\n  audits(first: 50) {\n    edges {\n      node {\n        id\n        ...AuditRowFragment\n      }\n    }\n  }\n}\n\nfragment OverviewPageFragment on TrustCenter {\n  organization {\n    name\n    id\n  }\n  references(first: 14) {\n    edges {\n      node {\n        id\n        name\n        logoUrl\n        websiteUrl\n      }\n    }\n  }\n  audits(first: 50) {\n    edges {\n      node {\n        id\n        ...AuditRowFragment\n      }\n    }\n  }\n  vendors(first: 3) {\n    edges {\n      node {\n        id\n        countries\n        ...VendorRowFragment\n      }\n    }\n  }\n  documents(first: 5) {\n    edges {\n      node {\n        id\n        ...DocumentRowFragment\n        documentType\n      }\n    }\n  }\n  trustCenterFiles(first: 5) {\n    edges {\n      node {\n        id\n        category\n        ...TrustCenterFileRowFragment\n      }\n    }\n  }\n}\n\nfragment TrustCenterFileRowFragment on TrustCenterFile {\n  id\n  name\n  isUserAuthorized\n  hasUserRequestedAccess\n}\n\nfragment VendorRowFragment on Vendor {\n  name\n  description\n  websiteUrl\n  countries\n}\n"
  }
};
})();

(node as any).hash = "24fe81d565e77cedd2a9825bc0ba30bc";

export default node;
