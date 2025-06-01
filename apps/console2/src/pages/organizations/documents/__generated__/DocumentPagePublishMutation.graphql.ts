/**
 * @generated SignedSource<<3ef23505e9378ce7ea92943b394ba290>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type PublishDocumentVersionInput = {
  documentId: string;
};
export type DocumentPagePublishMutation$variables = {
  input: PublishDocumentVersionInput;
};
export type DocumentPagePublishMutation$data = {
  readonly publishDocumentVersion: {
    readonly document: {
      readonly id: string;
    };
  };
};
export type DocumentPagePublishMutation = {
  response: DocumentPagePublishMutation$data;
  variables: DocumentPagePublishMutation$variables;
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
    "name": "DocumentPagePublishMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "DocumentPagePublishMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "fad75c1a396f60d6d13e1605fe85f33e",
    "id": null,
    "metadata": {},
    "name": "DocumentPagePublishMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentPagePublishMutation(\n  $input: PublishDocumentVersionInput!\n) {\n  publishDocumentVersion(input: $input) {\n    document {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "488029ccd88b4c637b63cd11d1bef986";

export default node;
