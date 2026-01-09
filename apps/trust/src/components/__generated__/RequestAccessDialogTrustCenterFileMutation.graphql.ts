/**
 * @generated SignedSource<<1e744bd11b2d5dd4d982bf1ce8b2a721>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RequestTrustCenterFileAccessInput = {
  email: any;
  fullName: string;
  trustCenterFileId: string;
  trustCenterId: string;
};
export type RequestAccessDialogTrustCenterFileMutation$variables = {
  input: RequestTrustCenterFileAccessInput;
};
export type RequestAccessDialogTrustCenterFileMutation$data = {
  readonly requestTrustCenterFileAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type RequestAccessDialogTrustCenterFileMutation = {
  response: RequestAccessDialogTrustCenterFileMutation$data;
  variables: RequestAccessDialogTrustCenterFileMutation$variables;
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
    "name": "RequestAccessDialogTrustCenterFileMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "RequestAccessDialogTrustCenterFileMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "73b3007ef1a45f6dc6d31be0c627e482",
    "id": null,
    "metadata": {},
    "name": "RequestAccessDialogTrustCenterFileMutation",
    "operationKind": "mutation",
    "text": "mutation RequestAccessDialogTrustCenterFileMutation(\n  $input: RequestTrustCenterFileAccessInput!\n) {\n  requestTrustCenterFileAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "a97eb6c4ec94c79a293d1cc51bf18e66";

export default node;
