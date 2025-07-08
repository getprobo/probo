/**
 * @generated SignedSource<<9f55e528d2b6366114a1852c9beaf1b0>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CancelSignatureRequestInput = {
  documentVersionSignatureId: string;
};
export type DocumentSignaturesDialog_cancelSignatureMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CancelSignatureRequestInput;
};
export type DocumentSignaturesDialog_cancelSignatureMutation$data = {
  readonly cancelSignatureRequest: {
    readonly deletedDocumentVersionSignatureId: string;
  };
};
export type DocumentSignaturesDialog_cancelSignatureMutation = {
  response: DocumentSignaturesDialog_cancelSignatureMutation$data;
  variables: DocumentSignaturesDialog_cancelSignatureMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "deletedDocumentVersionSignatureId",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "DocumentSignaturesDialog_cancelSignatureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CancelSignatureRequestPayload",
        "kind": "LinkedField",
        "name": "cancelSignatureRequest",
        "plural": false,
        "selections": [
          (v3/*: any*/)
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "DocumentSignaturesDialog_cancelSignatureMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CancelSignatureRequestPayload",
        "kind": "LinkedField",
        "name": "cancelSignatureRequest",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedDocumentVersionSignatureId",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "81d10822e563ceb7ebcd4fdb5d19e8f5",
    "id": null,
    "metadata": {},
    "name": "DocumentSignaturesDialog_cancelSignatureMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentSignaturesDialog_cancelSignatureMutation(\n  $input: CancelSignatureRequestInput!\n) {\n  cancelSignatureRequest(input: $input) {\n    deletedDocumentVersionSignatureId\n  }\n}\n"
  }
};
})();

(node as any).hash = "0cbec91e7ba7cc09d2bbf7169e98f30c";

export default node;
