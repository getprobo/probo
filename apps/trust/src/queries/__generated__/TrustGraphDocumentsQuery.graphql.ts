/**
 * @generated SignedSource<<0faaebe3baf4c750825599985210137d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type TrustGraphDocumentsQuery$variables = {
  slug: string;
};
export type TrustGraphDocumentsQuery$data = {
  readonly trustCenterBySlug: {
    readonly documents: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly documentType: DocumentType;
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"DocumentRowFragment">;
        };
      }>;
    };
    readonly id: string;
    readonly organization: {
      readonly name: string;
    };
    readonly trustCenterFiles: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly category: string;
          readonly id: string;
          readonly " $fragmentSpreads": FragmentRefs<"TrustCenterFileRowFragment">;
        };
      }>;
    };
  } | null | undefined;
};
export type TrustGraphDocumentsQuery = {
  response: TrustGraphDocumentsQuery$data;
  variables: TrustGraphDocumentsQuery$variables;
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
  "name": "name",
  "storageKey": null
},
v4 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 50
  }
],
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "documentType",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "category",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "isUserAuthorized",
  "storageKey": null
},
v8 = {
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
    "name": "TrustGraphDocumentsQuery",
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
          {
            "alias": null,
            "args": null,
            "concreteType": "Organization",
            "kind": "LinkedField",
            "name": "organization",
            "plural": false,
            "selections": [
              (v3/*: any*/)
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": (v4/*: any*/),
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
                      (v5/*: any*/),
                      {
                        "args": null,
                        "kind": "FragmentSpread",
                        "name": "DocumentRowFragment"
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "documents(first:50)"
          },
          {
            "alias": null,
            "args": (v4/*: any*/),
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
                      (v6/*: any*/),
                      {
                        "args": null,
                        "kind": "FragmentSpread",
                        "name": "TrustCenterFileRowFragment"
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "trustCenterFiles(first:50)"
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
    "name": "TrustGraphDocumentsQuery",
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
          {
            "alias": null,
            "args": null,
            "concreteType": "Organization",
            "kind": "LinkedField",
            "name": "organization",
            "plural": false,
            "selections": [
              (v3/*: any*/),
              (v2/*: any*/)
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": (v4/*: any*/),
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
                      (v5/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "title",
                        "storageKey": null
                      },
                      (v7/*: any*/),
                      (v8/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "documents(first:50)"
          },
          {
            "alias": null,
            "args": (v4/*: any*/),
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
                      (v6/*: any*/),
                      (v3/*: any*/),
                      (v7/*: any*/),
                      (v8/*: any*/)
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": null
              }
            ],
            "storageKey": "trustCenterFiles(first:50)"
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "151fc8b1cf3becb2a69d87ce978fb71b",
    "id": null,
    "metadata": {},
    "name": "TrustGraphDocumentsQuery",
    "operationKind": "query",
    "text": "query TrustGraphDocumentsQuery(\n  $slug: String!\n) {\n  trustCenterBySlug(slug: $slug) {\n    id\n    organization {\n      name\n      id\n    }\n    documents(first: 50) {\n      edges {\n        node {\n          id\n          documentType\n          ...DocumentRowFragment\n        }\n      }\n    }\n    trustCenterFiles(first: 50) {\n      edges {\n        node {\n          id\n          category\n          ...TrustCenterFileRowFragment\n        }\n      }\n    }\n  }\n}\n\nfragment DocumentRowFragment on Document {\n  id\n  title\n  isUserAuthorized\n  hasUserRequestedAccess\n}\n\nfragment TrustCenterFileRowFragment on TrustCenterFile {\n  id\n  name\n  isUserAuthorized\n  hasUserRequestedAccess\n}\n"
  }
};
})();

(node as any).hash = "79e21d586ec18a0977c37df27aa9bdea";

export default node;
