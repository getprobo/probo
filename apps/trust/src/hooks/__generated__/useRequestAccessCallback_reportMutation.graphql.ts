/**
 * @generated SignedSource<<b995c7259ba7df94fbdc49770ec83a92>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DocumentAccessStatus = "GRANTED" | "REJECTED" | "REQUESTED" | "REVOKED";
export type RequestReportAccessInput = {
  reportId: string;
};
export type useRequestAccessCallback_reportMutation$variables = {
  input: RequestReportAccessInput;
};
export type useRequestAccessCallback_reportMutation$data = {
  readonly requestReportAccess: {
    readonly audit: {
      readonly report: {
        readonly access: {
          readonly id: string;
          readonly status: DocumentAccessStatus;
        } | null | undefined;
      } | null | undefined;
    } | null | undefined;
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
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v2 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v3 = {
  "alias": null,
  "args": null,
  "concreteType": "DocumentAccess",
  "kind": "LinkedField",
  "name": "access",
  "plural": false,
  "selections": [
    (v2/*: any*/),
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "status",
      "storageKey": null
    }
  ],
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "useRequestAccessCallback_reportMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "RequestReportAccessPayload",
        "kind": "LinkedField",
        "name": "requestReportAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Audit",
            "kind": "LinkedField",
            "name": "audit",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Report",
                "kind": "LinkedField",
                "name": "report",
                "plural": false,
                "selections": [
                  (v3/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ],
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "useRequestAccessCallback_reportMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "RequestReportAccessPayload",
        "kind": "LinkedField",
        "name": "requestReportAccess",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "Audit",
            "kind": "LinkedField",
            "name": "audit",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "Report",
                "kind": "LinkedField",
                "name": "report",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v2/*: any*/)
                ],
                "storageKey": null
              },
              (v2/*: any*/)
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "82539e9ad64aa2bd9e5bd2cac4597f23",
    "id": null,
    "metadata": {},
    "name": "useRequestAccessCallback_reportMutation",
    "operationKind": "mutation",
    "text": "mutation useRequestAccessCallback_reportMutation(\n  $input: RequestReportAccessInput!\n) {\n  requestReportAccess(input: $input) {\n    audit {\n      report {\n        access {\n          id\n          status\n        }\n        id\n      }\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "6b07d19642a0227f76f81bb64dac82ea";

export default node;
