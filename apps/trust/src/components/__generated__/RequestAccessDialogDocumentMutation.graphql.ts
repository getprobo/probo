/**
 * @generated SignedSource<<9f01a480ccc5fa00e310296e9aed1020>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RequestDocumentAccessInput = {
  documentId: string;
  email: any;
  fullName: string;
};
export type RequestAccessDialogDocumentMutation$variables = {
  input: RequestDocumentAccessInput;
};
export type RequestAccessDialogDocumentMutation$data = {
  readonly requestDocumentAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type RequestAccessDialogDocumentMutation = {
  response: RequestAccessDialogDocumentMutation$data;
  variables: RequestAccessDialogDocumentMutation$variables;
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
    "name": "RequestAccessDialogDocumentMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "RequestAccessDialogDocumentMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "2ed1f0d19711a9f22acdc6251dbfca8f",
    "id": null,
    "metadata": {},
    "name": "RequestAccessDialogDocumentMutation",
    "operationKind": "mutation",
    "text": "mutation RequestAccessDialogDocumentMutation(\n  $input: RequestDocumentAccessInput!\n) {\n  requestDocumentAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "c3ebabe2f3fac32e5035cc7acfd54799";

export default node;
