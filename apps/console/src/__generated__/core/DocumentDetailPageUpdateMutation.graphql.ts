/**
 * @generated SignedSource<<7c1bfb7a8cc7cc0f426348c57a17f7c9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentClassification = "CONFIDENTIAL" | "INTERNAL" | "PUBLIC" | "SECRET";
export type DocumentType = "ISMS" | "OTHER" | "POLICY" | "PROCEDURE";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type UpdateDocumentInput = {
  classification?: DocumentClassification | null | undefined;
  content?: string | null | undefined;
  documentType?: DocumentType | null | undefined;
  id: string;
  ownerId?: string | null | undefined;
  title?: string | null | undefined;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
};
export type DocumentDetailPageUpdateMutation$variables = {
  input: UpdateDocumentInput;
};
export type DocumentDetailPageUpdateMutation$data = {
  readonly updateDocument: {
    readonly document: {
      readonly classification: DocumentClassification;
      readonly documentType: DocumentType;
      readonly id: string;
      readonly owner: {
        readonly fullName: string;
        readonly id: string;
      };
      readonly title: string;
    };
  };
};
export type DocumentDetailPageUpdateMutation = {
  response: DocumentDetailPageUpdateMutation$data;
  variables: DocumentDetailPageUpdateMutation$variables;
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
            "args": null,
            "kind": "ScalarField",
            "name": "classification",
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
    "name": "DocumentDetailPageUpdateMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentDetailPageUpdateMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "865056a65394e413822f798f31351341",
    "id": null,
    "metadata": {},
    "name": "DocumentDetailPageUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentDetailPageUpdateMutation(\n  $input: UpdateDocumentInput!\n) {\n  updateDocument(input: $input) {\n    document {\n      id\n      title\n      documentType\n      classification\n      owner {\n        id\n        fullName\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "bd6acddfc137a6bdce42628d05dc9ff6";

export default node;
