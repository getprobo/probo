/**
 * @generated SignedSource<<f1ba7d8550a4585118c8b9ed260607c4>>
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
export type DocumentSignatureListItem_cancelSignatureMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CancelSignatureRequestInput;
};
export type DocumentSignatureListItem_cancelSignatureMutation$data = {
  readonly cancelSignatureRequest: {
    readonly deletedDocumentVersionSignatureId: string;
  };
};
export type DocumentSignatureListItem_cancelSignatureMutation = {
  response: DocumentSignatureListItem_cancelSignatureMutation$data;
  variables: DocumentSignatureListItem_cancelSignatureMutation$variables;
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
    "name": "DocumentSignatureListItem_cancelSignatureMutation",
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
    "name": "DocumentSignatureListItem_cancelSignatureMutation",
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
    "cacheID": "494afd972f44f534cf6f728196717a82",
    "id": null,
    "metadata": {},
    "name": "DocumentSignatureListItem_cancelSignatureMutation",
    "operationKind": "mutation",
    "text": "mutation DocumentSignatureListItem_cancelSignatureMutation(\n  $input: CancelSignatureRequestInput!\n) {\n  cancelSignatureRequest(input: $input) {\n    deletedDocumentVersionSignatureId\n  }\n}\n"
  }
};
})();

(node as any).hash = "1a98d7aee9dc6758f186e915e1abe073";

export default node;
