/**
 * @generated SignedSource<<d36c3f435afa928b80d415bd53786d81>>
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
export type useRequestAccessCallback_reportMutation$variables = {
  input: RequestReportAccessInput;
};
export type useRequestAccessCallback_reportMutation$data = {
  readonly requestReportAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type useRequestAccessCallback_reportMutation = {
  response: useRequestAccessCallback_reportMutation$data;
  variables: useRequestAccessCallback_reportMutation$variables;
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
    "name": "useRequestAccessCallback_reportMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "useRequestAccessCallback_reportMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "c859384ca219b8302245e0bed292e698",
    "id": null,
    "metadata": {},
    "name": "useRequestAccessCallback_reportMutation",
    "operationKind": "mutation",
    "text": "mutation useRequestAccessCallback_reportMutation(\n  $input: RequestReportAccessInput!\n) {\n  requestReportAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6ad61c2f6638f5d31c9c25b0044f080a";

export default node;
