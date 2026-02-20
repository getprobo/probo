/**
 * @generated SignedSource<<6b6304c9a22de890f171632f8c9dabe1>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RequestDocumentAccessInput = {
  documentId: string;
};
export type useRequestAccessCallback_documentMutation$variables = {
  input: RequestDocumentAccessInput;
};
export type useRequestAccessCallback_documentMutation$data = {
  readonly requestDocumentAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type useRequestAccessCallback_documentMutation = {
  response: useRequestAccessCallback_documentMutation$data;
  variables: useRequestAccessCallback_documentMutation$variables;
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
    "concreteType": "RequestAccessesPayload",
    "kind": "LinkedField",
    "name": "requestDocumentAccess",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenterAccess",
        "kind": "LinkedField",
        "name": "trustCenterAccess",
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
    "name": "useRequestAccessCallback_documentMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "useRequestAccessCallback_documentMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "849ea6a746660bdee8ebf805baef7a33",
    "id": null,
    "metadata": {},
    "name": "useRequestAccessCallback_documentMutation",
    "operationKind": "mutation",
    "text": "mutation useRequestAccessCallback_documentMutation(\n  $input: RequestDocumentAccessInput!\n) {\n  requestDocumentAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "1ff8b5f560d8c3024e6d5c11bb2ed889";

export default node;
