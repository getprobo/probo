/**
 * @generated SignedSource<<8b9ec6ee6f688ae6f6d1463104c72df4>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type TrustGraphQuery$variables = {
  slug: string;
};
export type TrustGraphQuery$data = {
  readonly trustCenterBySlug: {
    readonly audits: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"AuditRowFragment">;
        };
      }>;
    };
    readonly hasAcceptedNonDisclosureAgreement: boolean;
    readonly id: string;
    readonly isUserAuthenticated: boolean;
    readonly ndaFileName: string | null | undefined;
    readonly ndaFileUrl: string | null | undefined;
    readonly organization: {
      readonly description: string | null | undefined;
      readonly email: string | null | undefined;
      readonly headquarterAddress: string | null | undefined;
      readonly logoUrl: string | null | undefined;
      readonly name: string;
      readonly websiteUrl: string | null | undefined;
    };
    readonly slug: string;
    readonly " $fragmentSpreads": FragmentRefs<"OverviewPageFragment">;
  } | null | undefined;
};
export type TrustGraphQuery = {
  response: TrustGraphQuery$data;
  variables: TrustGraphQuery$variables;
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
  "name": "slug",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "isUserAuthenticated",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasAcceptedNonDisclosureAgreement",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "ndaFileName",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "ndaFileUrl",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "description",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "websiteUrl",
  "storageKey": null
},
v11 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "logoUrl",
  "storageKey": null
},
v12 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "email",
  "storageKey": null
},
v13 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "headquarterAddress",
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
  "name": "category",
  "storageKey": null
},
v16 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 5
  }
],
v17 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "isUserAuthorized",
  "storageKey": null
},
v18 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "hasUserRequestedAccess",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustGraphQuery",
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
          (v5/*: any*/),
          (v6/*: any*/),
          (v7/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "Organization",
            "kind": "LinkedField",
            "name": "organization",
            "plural": false,
            "selections": [
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/)
            ],
            "storageKey": null
          },
          {
            "args": null,
            "kind": "FragmentSpread",
            "name": "OverviewPageFragment"
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
    "name": "TrustGraphQuery",
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
          (v5/*: any*/),
          (v6/*: any*/),
          (v7/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "Organization",
            "kind": "LinkedField",
            "name": "organization",
            "plural": false,
            "selections": [
              (v8/*: any*/),
              (v9/*: any*/),
              (v10/*: any*/),
              (v11/*: any*/),
              (v12/*: any*/),
              (v13/*: any*/),
              (v2/*: any*/)
            ],
            "storageKey": null
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
                      (v8/*: any*/),
                      (v11/*: any*/),
                      (v10/*: any*/)
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
                      (v8/*: any*/),
                      (v15/*: any*/),
                      (v10/*: any*/),
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
            "storageKey": "vendors(first:3)"
          },
          {
            "alias": null,
            "args": (v16/*: any*/),
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
                      (v17/*: any*/),
                      (v18/*: any*/),
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
            "args": (v16/*: any*/),
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
                      (v15/*: any*/),
                      (v8/*: any*/),
                      (v17/*: any*/),
                      (v18/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "trustCenterFiles(first:5)"
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
                          (v17/*: any*/),
                          (v18/*: any*/)
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
                          (v8/*: any*/)
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
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "f3ac933cf00a798d12b1a8e1fedcaee9",
    "id": null,
    "metadata": {},
    "name": "TrustGraphQuery",
    "operationKind": "query",
    "text": "query TrustGraphQuery(\n  $slug: String!\n) {\n  trustCenterBySlug(slug: $slug) {\n    id\n    slug\n    isUserAuthenticated\n    hasAcceptedNonDisclosureAgreement\n    ndaFileName\n    ndaFileUrl\n    organization {\n      name\n      description\n      websiteUrl\n      logoUrl\n      email\n      headquarterAddress\n      id\n    }\n    ...OverviewPageFragment\n    audits(first: 50) {\n      edges {\n        node {\n          id\n          ...AuditRowFragment\n        }\n      }\n    }\n  }\n}\n\nfragment AuditRowFragment on Audit {\n  report {\n    id\n    filename\n    isUserAuthorized\n    hasUserRequestedAccess\n  }\n  framework {\n    id\n    name\n  }\n}\n\nfragment DocumentRowFragment on Document {\n  id\n  title\n  isUserAuthorized\n  hasUserRequestedAccess\n}\n\nfragment OverviewPageFragment on TrustCenter {\n  references(first: 14) {\n    edges {\n      node {\n        id\n        name\n        logoUrl\n        websiteUrl\n      }\n    }\n  }\n  vendors(first: 3) {\n    edges {\n      node {\n        id\n        countries\n        ...VendorRowFragment\n      }\n    }\n  }\n  documents(first: 5) {\n    edges {\n      node {\n        id\n        ...DocumentRowFragment\n        documentType\n      }\n    }\n  }\n  trustCenterFiles(first: 5) {\n    edges {\n      node {\n        id\n        category\n        ...TrustCenterFileRowFragment\n      }\n    }\n  }\n}\n\nfragment TrustCenterFileRowFragment on TrustCenterFile {\n  id\n  name\n  isUserAuthorized\n  hasUserRequestedAccess\n}\n\nfragment VendorRowFragment on Vendor {\n  id\n  name\n  category\n  websiteUrl\n  privacyPolicyUrl\n  countries\n}\n"
  }
};
})();

(node as any).hash = "fd44a0d5fcda45aaa0fe02b3051643c0";

export default node;
