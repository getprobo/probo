/**
 * @generated SignedSource<<5439a269fe5c79f6d90f2421fdacb090>>
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
export type AuditRow_requestAccessMutation$variables = {
  input: RequestReportAccessInput;
};
export type AuditRow_requestAccessMutation$data = {
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
    "name": "AuditRow_requestAccessMutation",
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
    "name": "AuditRow_requestAccessMutation",
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
    "cacheID": "8d33e98c3b907ff4ddf6d08a0d9e7ff8",
    "id": null,
    "metadata": {},
    "name": "AuditRow_requestAccessMutation",
    "operationKind": "mutation",
    "text": "mutation AuditRow_requestAccessMutation(\n  $input: RequestReportAccessInput!\n) {\n  requestReportAccess(input: $input) {\n    audit {\n      report {\n        access {\n          id\n          status\n        }\n        id\n      }\n      id\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "01e09978f688128c89698e9a0877fa37";

export default node;
