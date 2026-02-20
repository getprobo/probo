/**
 * @generated SignedSource<<6db2a4322199d638b3f1ac498ecd066b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RequestTrustCenterFileAccessInput = {
  trustCenterFileId: string;
};
export type useRequestAccessCallback_fileMutation$variables = {
  input: RequestTrustCenterFileAccessInput;
};
export type useRequestAccessCallback_fileMutation$data = {
  readonly requestTrustCenterFileAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type useRequestAccessCallback_fileMutation = {
  response: useRequestAccessCallback_fileMutation$data;
  variables: useRequestAccessCallback_fileMutation$variables;
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
    "name": "requestTrustCenterFileAccess",
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
    "name": "useRequestAccessCallback_fileMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "useRequestAccessCallback_fileMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "1ed0d9f8a3b3dc08c2af16101729d3a9",
    "id": null,
    "metadata": {},
    "name": "useRequestAccessCallback_fileMutation",
    "operationKind": "mutation",
    "text": "mutation useRequestAccessCallback_fileMutation(\n  $input: RequestTrustCenterFileAccessInput!\n) {\n  requestTrustCenterFileAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5105b60de987b135523ada9ee2777dea";

export default node;
