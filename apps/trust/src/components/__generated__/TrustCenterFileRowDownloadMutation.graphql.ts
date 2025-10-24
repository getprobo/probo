/**
 * @generated SignedSource<<8d608da0bcff489887f71f6e69f1bfe3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type ExportTrustCenterFileInput = {
  trustCenterFileId: string;
};
export type TrustCenterFileRowDownloadMutation$variables = {
  input: ExportTrustCenterFileInput;
};
export type TrustCenterFileRowDownloadMutation$data = {
  readonly exportTrustCenterFile: {
    readonly data: string;
  };
};
export type TrustCenterFileRowDownloadMutation = {
  response: TrustCenterFileRowDownloadMutation$data;
  variables: TrustCenterFileRowDownloadMutation$variables;
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
    "concreteType": "ExportTrustCenterFilePayload",
    "kind": "LinkedField",
    "name": "exportTrustCenterFile",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "kind": "ScalarField",
        "name": "data",
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
    "name": "TrustCenterFileRowDownloadMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "TrustCenterFileRowDownloadMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "8a97b9e7ac44deea6da6e79f1d0c662f",
    "id": null,
    "metadata": {},
    "name": "TrustCenterFileRowDownloadMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterFileRowDownloadMutation(\n  $input: ExportTrustCenterFileInput!\n) {\n  exportTrustCenterFile(input: $input) {\n    data\n  }\n}\n"
  }
};
})();

(node as any).hash = "e6d74e96e929cf1aa79711b8315a8b85";

export default node;
