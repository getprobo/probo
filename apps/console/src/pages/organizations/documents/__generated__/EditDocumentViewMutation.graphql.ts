/**
 * @generated SignedSource<<544173fd0ab2731cb9d3a976212ceca5>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateDocumentVersionInput = {
  content: string;
  documentVersionId: string;
};
export type EditDocumentViewMutation$variables = {
  input: UpdateDocumentVersionInput;
};
export type EditDocumentViewMutation$data = {
  readonly updateDocumentVersion: {
    readonly documentVersion: {
      readonly content: string;
      readonly id: string;
    };
  };
};
export type EditDocumentViewMutation = {
  response: EditDocumentViewMutation$data;
  variables: EditDocumentViewMutation$variables;
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
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateDocumentVersionPayload",
    "kind": "LinkedField",
    "name": "updateDocumentVersion",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "DocumentVersion",
        "kind": "LinkedField",
        "name": "documentVersion",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "content",
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
    "name": "EditDocumentViewMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditDocumentViewMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "92e3e1e6d3bf1a2066b2e71c437c3b4e",
    "id": null,
    "metadata": {},
    "name": "EditDocumentViewMutation",
    "operationKind": "mutation",
    "text": "mutation EditDocumentViewMutation(\n  $input: UpdateDocumentVersionInput!\n) {\n  updateDocumentVersion(input: $input) {\n    documentVersion {\n      id\n      content\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "14fc2fdfcdb2ccb827e237d5fc1f0a6f";

export default node;
