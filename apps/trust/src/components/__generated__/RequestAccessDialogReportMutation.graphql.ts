/**
 * @generated SignedSource<<74bc40a8998c962930b3f0b63dc6acd9>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RequestReportAccessInput = {
  email?: string | null | undefined;
  name?: string | null | undefined;
  reportId: string;
  trustCenterId: string;
};
export type RequestAccessDialogReportMutation$variables = {
  input: RequestReportAccessInput;
};
export type RequestAccessDialogReportMutation$data = {
  readonly requestReportAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type RequestAccessDialogReportMutation = {
  response: RequestAccessDialogReportMutation$data;
  variables: RequestAccessDialogReportMutation$variables;
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
    "name": "requestReportAccess",
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
    "name": "RequestAccessDialogReportMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "RequestAccessDialogReportMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "da69c3e370244e4a9a7793b06b2abdd5",
    "id": null,
    "metadata": {},
    "name": "RequestAccessDialogReportMutation",
    "operationKind": "mutation",
    "text": "mutation RequestAccessDialogReportMutation(\n  $input: RequestReportAccessInput!\n) {\n  requestReportAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "3ef2b6424ba751f15f91ad7f4bd28fc5";

export default node;
