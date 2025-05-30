/**
 * @generated SignedSource<<9752f2c7d1c495c0c21482fb6d881c9e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentType = "ISMS" | "OTHER" | "POLICY";
export type UpdateDocumentInput = {
  content?: string | null | undefined;
  createdBy?: string | null | undefined;
  documentType?: DocumentType | null | undefined;
  id: string;
  ownerId?: string | null | undefined;
  title?: string | null | undefined;
};
export type ShowDocumentViewUpdateDocumentTypeMutation$variables = {
  input: UpdateDocumentInput;
};
export type ShowDocumentViewUpdateDocumentTypeMutation$data = {
  readonly updateDocument: {
    readonly document: {
      readonly documentType: DocumentType;
      readonly id: string;
      readonly owner: {
        readonly fullName: string;
        readonly id: string;
        readonly primaryEmailAddress: string;
      };
    };
  };
};
export type ShowDocumentViewUpdateDocumentTypeMutation = {
  response: ShowDocumentViewUpdateDocumentTypeMutation$data;
  variables: ShowDocumentViewUpdateDocumentTypeMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateDocumentPayload",
    "kind": "LinkedField",
    "name": "updateDocument",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Document",
        "kind": "LinkedField",
        "name": "document",
        "plural": false,
        "selections": [
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "documentType",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "People",
            "kind": "LinkedField",
            "name": "owner",
            "plural": false,
            "selections": [
              (v1/*: any*/),
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
                "name": "primaryEmailAddress",
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
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "ShowDocumentViewUpdateDocumentTypeMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ShowDocumentViewUpdateDocumentTypeMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "4904ae0647cf1c01126766096adf2b4b",
    "id": null,
    "metadata": {},
    "name": "ShowDocumentViewUpdateDocumentTypeMutation",
    "operationKind": "mutation",
    "text": "mutation ShowDocumentViewUpdateDocumentTypeMutation(\n  $input: UpdateDocumentInput!\n) {\n  updateDocument(input: $input) {\n    document {\n      id\n      documentType\n      owner {\n        id\n        fullName\n        primaryEmailAddress\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5d04382a57a6020b7e8c80e615e1b32a";

export default node;
