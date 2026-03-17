/**
 * @generated SignedSource<<1f7c6a92ab7920b470a03799e04f4ea3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateDocumentVersionContentInput = {
  content: string;
  id: string;
};
export type DocumentDescriptionPage_updateContentMutation$variables = {
  input: UpdateDocumentVersionContentInput;
};
export type DocumentDescriptionPage_updateContentMutation$data = {
  readonly updateDocumentVersionContent: {
    readonly content: string;
  };
};
export type DocumentDescriptionPage_updateContentMutation = {
  response: DocumentDescriptionPage_updateContentMutation$data;
  variables: DocumentDescriptionPage_updateContentMutation$variables;
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
    "concreteType": "UpdateDocumentVersionContentPayload",
    "kind": "LinkedField",
    "name": "updateDocumentVersionContent",
    "plural": false,
    "selections": [
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
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentDescriptionPage_updateContentMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentDescriptionPage_updateContentMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "2a43a5cf619ca1d475b3e480a132878b",
    "id": null,
    "metadata": {},
    "name": "DocumentDescriptionPage_updateContentMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentDescriptionPage_updateContentMutation(\n  $input: UpdateDocumentVersionContentInput!\n) {\n  updateDocumentVersionContent(input: $input) {\n    content\n  }\n}\n"
  }
};
})();

(node as any).hash = "2bfee42c8fe67762416dec052c5f051b";

export default node;
