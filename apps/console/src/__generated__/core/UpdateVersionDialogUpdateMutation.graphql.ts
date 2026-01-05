/**
 * @generated SignedSource<<2b90fd4d14e86e5547d7b790a20242c9>>
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
export type UpdateVersionDialogUpdateMutation$variables = {
  input: UpdateDocumentVersionInput;
};
export type UpdateVersionDialogUpdateMutation$data = {
  readonly updateDocumentVersion: {
    readonly documentVersion: {
      readonly content: string;
      readonly id: string;
    };
  };
};
export type UpdateVersionDialogUpdateMutation = {
  response: UpdateVersionDialogUpdateMutation$data;
  variables: UpdateVersionDialogUpdateMutation$variables;
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
    "name": "UpdateVersionDialogUpdateMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "UpdateVersionDialogUpdateMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "527e45a341f2f991e30e8689ac24403c",
    "id": null,
    "metadata": {},
    "name": "UpdateVersionDialogUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation UpdateVersionDialogUpdateMutation(\n  $input: UpdateDocumentVersionInput!\n) {\n  updateDocumentVersion(input: $input) {\n    documentVersion {\n      id\n      content\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "ffd885c021e9fd91522a494e16e6830d";

export default node;
