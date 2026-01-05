/**
 * @generated SignedSource<<4bc817c960234139f5aeedfed5b2fe8b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UploadTrustCenterNDAInput = {
  file: any;
  fileName: string;
  trustCenterId: string;
};
export type TrustCenterGraphUploadNDAMutation$variables = {
  input: UploadTrustCenterNDAInput;
};
export type TrustCenterGraphUploadNDAMutation$data = {
  readonly uploadTrustCenterNDA: {
    readonly trustCenter: {
      readonly id: string;
      readonly ndaFileName: string | null | undefined;
      readonly updatedAt: string;
    };
  };
};
export type TrustCenterGraphUploadNDAMutation = {
  response: TrustCenterGraphUploadNDAMutation$data;
  variables: TrustCenterGraphUploadNDAMutation$variables;
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
    "concreteType": "UploadTrustCenterNDAPayload",
    "kind": "LinkedField",
    "name": "uploadTrustCenterNDA",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "TrustCenter",
        "kind": "LinkedField",
        "name": "trustCenter",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "id",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "ndaFileName",
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "kind": "ScalarField",
            "name": "updatedAt",
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
    "name": "TrustCenterGraphUploadNDAMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterGraphUploadNDAMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "09a5dbf73f3e279173691b235f9da2e2",
    "id": null,
    "metadata": {},
    "name": "TrustCenterGraphUploadNDAMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterGraphUploadNDAMutation(\n  $input: UploadTrustCenterNDAInput!\n) {\n  uploadTrustCenterNDA(input: $input) {\n    trustCenter {\n      id\n      ndaFileName\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b9877f8a5b9c2c12addeb939360719f1";

export default node;
