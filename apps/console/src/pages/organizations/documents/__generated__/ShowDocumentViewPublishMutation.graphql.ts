/**
 * @generated SignedSource<<91076d34daad785f7fce9f5a2cbc4018>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentStatus = "DRAFT" | "PUBLISHED";
export type PublishDocumentVersionInput = {
  documentId: string;
};
export type ShowDocumentViewPublishMutation$variables = {
  input: PublishDocumentVersionInput;
};
export type ShowDocumentViewPublishMutation$data = {
  readonly publishDocumentVersion: {
    readonly document: {
      readonly currentPublishedVersion: number | null | undefined;
      readonly id: string;
    };
    readonly documentVersion: {
      readonly id: string;
      readonly publishedAt: string | null | undefined;
      readonly publishedBy: {
        readonly fullName: string;
      } | null | undefined;
      readonly status: DocumentStatus;
    };
  };
};
export type ShowDocumentViewPublishMutation = {
  response: ShowDocumentViewPublishMutation$data;
  variables: ShowDocumentViewPublishMutation$variables;
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
      "name": "currentPublishedVersion",
      "storageKey": null
    }
  ],
  "storageKey": null
},
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
  "kind": "ScalarField",
  "name": "publishedAt",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fullName",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ShowDocumentViewPublishMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "PublishDocumentVersionPayload",
        "kind": "LinkedField",
        "name": "publishDocumentVersion",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "DocumentVersion",
            "kind": "LinkedField",
            "name": "documentVersion",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "People",
                "kind": "LinkedField",
                "name": "publishedBy",
                "plural": false,
                "selections": [
                  (v6/*: any*/)
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
    "name": "ShowDocumentViewPublishMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "PublishDocumentVersionPayload",
        "kind": "LinkedField",
        "name": "publishDocumentVersion",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "DocumentVersion",
            "kind": "LinkedField",
            "name": "documentVersion",
            "plural": false,
            "selections": [
              (v2/*: any*/),
              (v4/*: any*/),
              (v5/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "People",
                "kind": "LinkedField",
                "name": "publishedBy",
                "plural": false,
                "selections": [
                  (v6/*: any*/),
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
    ]
  },
  "params": {
    "cacheID": "794539aab3ddce5d004e72d10b980ca4",
    "id": null,
    "metadata": {},
    "name": "ShowDocumentViewPublishMutation",
    "operationKind": "mutation",
    "text": "mutation ShowDocumentViewPublishMutation(\n  $input: PublishDocumentVersionInput!\n) {\n  publishDocumentVersion(input: $input) {\n    document {\n      id\n      currentPublishedVersion\n    }\n    documentVersion {\n      id\n      status\n      publishedAt\n      publishedBy {\n        fullName\n        id\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "66e590251b58966c2ac09625759332f7";

export default node;
