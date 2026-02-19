/**
 * @generated SignedSource<<9c9f3a880879349cf5d21d03bf7e2c7a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type RequestReportAccessInput = {
  reportId: string;
};
export type ReportRow_requestAccessMutation$variables = {
  input: RequestReportAccessInput;
};
export type ReportRow_requestAccessMutation$data = {
  readonly requestReportAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type ReportRow_requestAccessMutation = {
  response: ReportRow_requestAccessMutation$data;
  variables: ReportRow_requestAccessMutation$variables;
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
    "name": "ReportRow_requestAccessMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "ReportRow_requestAccessMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "54968dc206a3d92a93838db977d7f3fd",
    "id": null,
    "metadata": {},
    "name": "ReportRow_requestAccessMutation",
    "operationKind": "mutation",
    "text": "mutation ReportRow_requestAccessMutation(\n  $input: RequestReportAccessInput!\n) {\n  requestReportAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "e067d2ff33fe43b0f94205db592c6381";

export default node;
