/**
 * @generated SignedSource<<6501e136d1171aaca8238084a7606a9e>>
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
export type EditDocumentViewVersionMutation$variables = {
  input: UpdateDocumentVersionInput;
};
export type EditDocumentViewVersionMutation$data = {
  readonly updateDocumentVersion: {
    readonly documentVersion: {
      readonly content: string;
      readonly id: string;
    };
  };
};
export type EditDocumentViewVersionMutation = {
  response: EditDocumentViewVersionMutation$data;
  variables: EditDocumentViewVersionMutation$variables;
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
    "name": "EditDocumentViewVersionMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "EditDocumentViewVersionMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "aa2a1c45efb2771987bf777971585e92",
    "id": null,
    "metadata": {},
    "name": "EditDocumentViewVersionMutation",
    "operationKind": "mutation",
    "text": "mutation EditDocumentViewVersionMutation(\n  $input: UpdateDocumentVersionInput!\n) {\n  updateDocumentVersion(input: $input) {\n    documentVersion {\n      id\n      content\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "656a8d74e00bb9e15e42987775bb1d14";

export default node;
