/**
 * @generated SignedSource<<5cb0be6e5ab775bdf7d509d9789e9821>>
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
  approverId?: string | null | undefined;
  classification?: DocumentClassification | null | undefined;
  content?: string | null | undefined;
  documentType?: DocumentType | null | undefined;
  id: string;
  title?: string | null | undefined;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
};
export type DocumentLayoutDrawerMutation$variables = {
  input: UpdateDocumentInput;
};
export type DocumentLayoutDrawerMutation$data = {
  readonly updateDocument: {
    readonly document: {
      readonly approver: {
        readonly fullName: string;
        readonly id: string;
      };
      readonly classification: DocumentClassification;
      readonly documentType: DocumentType;
      readonly id: string;
    };
  };
};
export type DocumentLayoutDrawerMutation = {
  response: DocumentLayoutDrawerMutation$data;
  variables: DocumentLayoutDrawerMutation$variables;
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
            "kind": "ScalarField",
            "name": "classification",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "concreteType": "Profile",
            "kind": "LinkedField",
            "name": "approver",
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
    "name": "DocumentLayoutDrawerMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentLayoutDrawerMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "717ce480e61acba39409ff42ee5344cb",
    "id": null,
    "metadata": {},
    "name": "DocumentLayoutDrawerMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentLayoutDrawerMutation(\n  $input: UpdateDocumentInput!\n) {\n  updateDocument(input: $input) {\n    document {\n      id\n      documentType\n      classification\n      approver {\n        id\n        fullName\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "815cb617b0685f018afe95e97c910771";

export default node;
