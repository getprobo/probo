/**
 * @generated SignedSource<<0a17c03011fe74b75a90dbeae02ab2ef>>
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
export type DocumentSignaturesTab_cancelSignatureMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CancelSignatureRequestInput;
};
export type DocumentSignaturesTab_cancelSignatureMutation$data = {
  readonly cancelSignatureRequest: {
    readonly deletedDocumentVersionSignatureId: string;
  };
};
export type DocumentSignaturesTab_cancelSignatureMutation = {
  response: DocumentSignaturesTab_cancelSignatureMutation$data;
  variables: DocumentSignaturesTab_cancelSignatureMutation$variables;
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
    "name": "DocumentSignaturesTab_cancelSignatureMutation",
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
    "name": "DocumentSignaturesTab_cancelSignatureMutation",
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
    "cacheID": "c57722e2b57faca64eea4950864e034d",
    "id": null,
    "metadata": {},
    "name": "DocumentSignaturesTab_cancelSignatureMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentSignaturesTab_cancelSignatureMutation(\n  $input: CancelSignatureRequestInput!\n) {\n  cancelSignatureRequest(input: $input) {\n    deletedDocumentVersionSignatureId\n  }\n}\n"
  }
};
})();

(node as any).hash = "9f5d3ff869ec6f91285140285248e824";

export default node;
