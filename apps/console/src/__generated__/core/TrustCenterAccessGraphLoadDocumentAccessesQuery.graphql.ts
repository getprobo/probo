/**
 * @generated SignedSource<<88767d7e81f9900c5bd01a14ffb14bfa>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type TrustCenterDocumentAccessStatus = "GRANTED" | "REJECTED" | "REQUESTED" | "REVOKED";
export type TrustCenterAccessGraphLoadDocumentAccessesQuery$variables = {
  accessId: string;
};
export type TrustCenterAccessGraphLoadDocumentAccessesQuery$data = {
  readonly node: {
    readonly availableDocumentAccesses?: {
      readonly edges: ReadonlyArray<{
        readonly node: {
          readonly document: {
            readonly documentType: DocumentType;
            readonly id: string;
            readonly title: string;
          } | null | undefined;
          readonly id: string;
          readonly report: {
            readonly audit: {
              readonly framework: {
                readonly name: string;
              };
              readonly id: string;
            } | null | undefined;
            readonly filename: string;
            readonly id: string;
          } | null | undefined;
          readonly status: TrustCenterDocumentAccessStatus;
          readonly trustCenterFile: {
            readonly category: string;
            readonly id: string;
            readonly name: string;
          } | null | undefined;
        };
      }>;
    };
    readonly id?: string;
    readonly ndaSignature?: {
      readonly " $fragmentSpreads": FragmentRefs<"ElectronicSignatureSectionFragment">;
    } | null | undefined;
  };
};
export type TrustCenterAccessGraphLoadDocumentAccessesQuery = {
  response: TrustCenterAccessGraphLoadDocumentAccessesQuery$data;
  variables: TrustCenterAccessGraphLoadDocumentAccessesQuery$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "accessId"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "id",
    "variableName": "accessId"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = [
  {
    "kind": "Literal",
    "name": "first",
    "value": 100
  },
  {
    "kind": "Literal",
    "name": "orderBy",
    "value": {
      "direction": "DESC",
      "field": "CREATED_AT"
    }
  }
],
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "status",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "concreteType": "Document",
  "kind": "LinkedField",
  "name": "document",
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
    }
  ],
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "filename",
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
  "concreteType": "TrustCenterFile",
  "kind": "LinkedField",
  "name": "trustCenterFile",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    (v7/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "category",
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustCenterAccessGraphLoadDocumentAccessesQuery",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": null,
        "kind": "LinkedField",
        "name": "node",
        "plural": false,
        "selections": [
          {
            "kind": "InlineFragment",
            "selections": [
              (v2/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "ElectronicSignature",
                "kind": "LinkedField",
                "name": "ndaSignature",
                "plural": false,
                "selections": [
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "ElectronicSignatureSectionFragment"
                  }
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v3/*: any*/),
                "concreteType": "TrustCenterDocumentAccessConnection",
                "kind": "LinkedField",
                "name": "availableDocumentAccesses",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TrustCenterDocumentAccessEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "TrustCenterDocumentAccess",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v4/*: any*/),
                          (v5/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Report",
                            "kind": "LinkedField",
                            "name": "report",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              (v6/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "Audit",
                                "kind": "LinkedField",
                                "name": "audit",
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
                                      (v7/*: any*/)
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": null
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "availableDocumentAccesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              }
            ],
            "type": "TrustCenterAccess",
            "abstractKey": null
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
    "name": "TrustCenterAccessGraphLoadDocumentAccessesQuery",
    "selections": [
      {
        "alias": null,
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
          (v2/*: any*/),
          {
            "kind": "InlineFragment",
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "ElectronicSignature",
                "kind": "LinkedField",
                "name": "ndaSignature",
                "plural": false,
                "selections": [
                  (v4/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "signedAt",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "certificateFileUrl",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "ElectronicSignatureEvent",
                    "kind": "LinkedField",
                    "name": "events",
                    "plural": true,
                    "selections": [
                      (v2/*: any*/),
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "eventType",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "actorEmail",
                        "storageKey": null
                      },
                      {
                        "alias": null,
                        "args": null,
                        "kind": "ScalarField",
                        "name": "occurredAt",
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  },
                  (v2/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": (v3/*: any*/),
                "concreteType": "TrustCenterDocumentAccessConnection",
                "kind": "LinkedField",
                "name": "availableDocumentAccesses",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "TrustCenterDocumentAccessEdge",
                    "kind": "LinkedField",
                    "name": "edges",
                    "plural": true,
                    "selections": [
                      {
                        "alias": null,
                        "args": null,
                        "concreteType": "TrustCenterDocumentAccess",
                        "kind": "LinkedField",
                        "name": "node",
                        "plural": false,
                        "selections": [
                          (v2/*: any*/),
                          (v4/*: any*/),
                          (v5/*: any*/),
                          {
                            "alias": null,
                            "args": null,
                            "concreteType": "Report",
                            "kind": "LinkedField",
                            "name": "report",
                            "plural": false,
                            "selections": [
                              (v2/*: any*/),
                              (v6/*: any*/),
                              {
                                "alias": null,
                                "args": null,
                                "concreteType": "Audit",
                                "kind": "LinkedField",
                                "name": "audit",
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
                                      (v7/*: any*/),
                                      (v2/*: any*/)
                                    ],
                                    "storageKey": null
                                  }
                                ],
                                "storageKey": null
                              }
                            ],
                            "storageKey": null
                          },
                          (v8/*: any*/)
                        ],
                        "storageKey": null
                      }
                    ],
                    "storageKey": null
                  }
                ],
                "storageKey": "availableDocumentAccesses(first:100,orderBy:{\"direction\":\"DESC\",\"field\":\"CREATED_AT\"})"
              }
            ],
            "type": "TrustCenterAccess",
            "abstractKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "175c4de41f02ea26f9a4ebc2124d6a51",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphLoadDocumentAccessesQuery",
    "operationKind": "query",
    "text": "query TrustCenterAccessGraphLoadDocumentAccessesQuery(\n  $accessId: ID!\n) {\n  node(id: $accessId) {\n    __typename\n    ... on TrustCenterAccess {\n      id\n      ndaSignature {\n        ...ElectronicSignatureSectionFragment\n        id\n      }\n      availableDocumentAccesses(first: 100, orderBy: {field: CREATED_AT, direction: DESC}) {\n        edges {\n          node {\n            id\n            status\n            document {\n              id\n              title\n              documentType\n            }\n            report {\n              id\n              filename\n              audit {\n                id\n                framework {\n                  name\n                  id\n                }\n              }\n            }\n            trustCenterFile {\n              id\n              name\n              category\n            }\n          }\n        }\n      }\n    }\n    id\n  }\n}\n\nfragment ElectronicSignatureSectionFragment on ElectronicSignature {\n  status\n  signedAt\n  certificateFileUrl\n  events {\n    id\n    eventType\n    actorEmail\n    occurredAt\n  }\n}\n"
  }
};
})();

(node as any).hash = "d104626a4915648265550778f1ce6606";

export default node;
