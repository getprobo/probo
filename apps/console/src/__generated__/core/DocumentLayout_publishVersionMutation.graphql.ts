/**
 * @generated SignedSource<<4617b6f203cf05562dccc80767cd0f9b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PublishDocumentVersionInput = {
  changelog?: string | null | undefined;
  documentId: string;
};
export type DocumentLayout_publishVersionMutation$variables = {
  input: PublishDocumentVersionInput;
};
export type DocumentLayout_publishVersionMutation$data = {
  readonly publishDocumentVersion: {
    readonly document: {
      readonly id: string;
    };
  };
};
export type DocumentLayout_publishVersionMutation = {
  response: DocumentLayout_publishVersionMutation$data;
  variables: DocumentLayout_publishVersionMutation$variables;
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
    "concreteType": "PublishDocumentVersionPayload",
    "kind": "LinkedField",
    "name": "publishDocumentVersion",
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
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
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
    "name": "DocumentLayout_publishVersionMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentLayout_publishVersionMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "2d64ca66f58d2c312868cb6de4d01fd4",
    "id": null,
    "metadata": {},
    "name": "DocumentLayout_publishVersionMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentLayout_publishVersionMutation(\n  $input: PublishDocumentVersionInput!\n) {\n  publishDocumentVersion(input: $input) {\n    document {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "226a81f075451d6db4d4f49e4bab70d6";

export default node;
