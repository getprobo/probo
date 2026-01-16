/**
 * @generated SignedSource<<0e5c234b565a3031a774995bff249b97>>
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
export type AuditRow_requestAccessMutation$variables = {
  input: RequestReportAccessInput;
};
export type AuditRow_requestAccessMutation$data = {
  readonly requestReportAccess: {
    readonly trustCenterAccess: {
      readonly id: string;
    };
  };
};
export type AuditRow_requestAccessMutation = {
  response: AuditRow_requestAccessMutation$data;
  variables: AuditRow_requestAccessMutation$variables;
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
    "name": "AuditRow_requestAccessMutation",
    "selections": (v1/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AuditRow_requestAccessMutation",
    "selections": (v1/*: any*/)
  },
  "params": {
    "cacheID": "6760f1ec8a6990f62461c3fab2b2934a",
    "id": null,
    "metadata": {},
    "name": "AuditRow_requestAccessMutation",
    "operationKind": "mutation",
    "text": "mutation AuditRow_requestAccessMutation(\n  $input: RequestReportAccessInput!\n) {\n  requestReportAccess(input: $input) {\n    trustCenterAccess {\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6ab5eb701126385e717a381e21e55ae0";

export default node;
