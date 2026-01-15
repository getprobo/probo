/**
 * @generated SignedSource<<a8da5aa24d8696b55231f44d861f347c>>
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
export type TrustCenterFileRow_requestAccessMutation$variables = {
  input: RequestTrustCenterFileAccessInput;
};
export type TrustCenterFileRow_requestAccessMutation$data = {
  readonly requestTrustCenterFileAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type TrustCenterFileRow_requestAccessMutation = {
  response: TrustCenterFileRow_requestAccessMutation$data;
  variables: TrustCenterFileRow_requestAccessMutation$variables;
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
    "name": "TrustCenterFileRow_requestAccessMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterFileRow_requestAccessMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "77a09810aa9c3ddf3a3baddf709f9a42",
    "id": null,
    "metadata": {},
    "name": "TrustCenterFileRow_requestAccessMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterFileRow_requestAccessMutation(\n  $input: RequestTrustCenterFileAccessInput!\n) {\n  requestTrustCenterFileAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "eba04489de9ec18ab9816328a481ef33";

export default node;
