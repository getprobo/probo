/**
 * @generated SignedSource<<786bab8a5d91a80d7652fa2da6f6088b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY";
export type UpdateTrustCenterDocumentAccessStatusInput = {
  active: boolean;
  id: string;
};
export type TrustCenterAccessGraphUpdateDocumentAccessStatusMutation$variables = {
  input: UpdateTrustCenterDocumentAccessStatusInput;
};
export type TrustCenterAccessGraphUpdateDocumentAccessStatusMutation$data = {
  readonly updateTrustCenterDocumentAccessStatus: {
    readonly trustCenterDocumentAccess: {
      readonly active: boolean;
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
      readonly updatedAt: any;
    };
  };
};
export type TrustCenterAccessGraphUpdateDocumentAccessStatusMutation = {
  response: TrustCenterAccessGraphUpdateDocumentAccessStatusMutation$data;
  variables: TrustCenterAccessGraphUpdateDocumentAccessStatusMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
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
  "name": "updatedAt",
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
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "TrustCenterAccessGraphUpdateDocumentAccessStatusMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateTrustCenterDocumentAccessStatusPayload",
        "kind": "LinkedField",
        "name": "updateTrustCenterDocumentAccessStatus",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterDocumentAccess",
            "kind": "LinkedField",
            "name": "trustCenterDocumentAccess",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
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
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterAccessGraphUpdateDocumentAccessStatusMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateTrustCenterDocumentAccessStatusPayload",
        "kind": "LinkedField",
        "name": "updateTrustCenterDocumentAccessStatus",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "TrustCenterDocumentAccess",
            "kind": "LinkedField",
            "name": "trustCenterDocumentAccess",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              (v3/*: any*/),
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
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "d6551e8cf21a2342783ff8224515104f",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAccessGraphUpdateDocumentAccessStatusMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAccessGraphUpdateDocumentAccessStatusMutation(\n  $input: UpdateTrustCenterDocumentAccessStatusInput!\n) {\n  updateTrustCenterDocumentAccessStatus(input: $input) {\n    trustCenterDocumentAccess {\n      id\n      active\n      updatedAt\n      document {\n        id\n        title\n        documentType\n      }\n      report {\n        id\n        filename\n        audit {\n          id\n          framework {\n            name\n            id\n          }\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bf26eeeb551b615fd6a889a9ee932ca4";

export default node;
