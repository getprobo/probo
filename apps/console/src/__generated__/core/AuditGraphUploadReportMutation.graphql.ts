/**
 * @generated SignedSource<<497a7821f6c5e32068e495ccb8489075>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UploadAuditReportInput = {
  auditId: string;
  file: any;
};
export type AuditGraphUploadReportMutation$variables = {
  input: UploadAuditReportInput;
};
export type AuditGraphUploadReportMutation$data = {
  readonly uploadAuditReport: {
    readonly audit: {
      readonly id: string;
      readonly report: {
        readonly createdAt: string;
        readonly downloadUrl: string | null | undefined;
        readonly filename: string;
        readonly id: string;
      } | null | undefined;
      readonly updatedAt: string;
    };
  };
};
export type AuditGraphUploadReportMutation = {
  response: AuditGraphUploadReportMutation$data;
  variables: AuditGraphUploadReportMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UploadAuditReportPayload",
    "kind": "LinkedField",
    "name": "uploadAuditReport",
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
          (v1/*: any*/),
          {
            "alias": null,
            "args": null,
            "concreteType": "Report",
            "kind": "LinkedField",
            "name": "report",
            "plural": false,
            "selections": [
              (v1/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "filename",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "downloadUrl",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "createdAt",
                "storageKey": null
              }
            ],
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
    "name": "AuditGraphUploadReportMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AuditGraphUploadReportMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "05229f6e6fb7e64ceaf93c4bb8f144e6",
    "id": null,
    "metadata": {},
    "name": "AuditGraphUploadReportMutation",
    "operationKind": "mutation",
    "text": "mutation AuditGraphUploadReportMutation(\n  $input: UploadAuditReportInput!\n) {\n  uploadAuditReport(input: $input) {\n    audit {\n      id\n      report {\n        id\n        filename\n        downloadUrl\n        createdAt\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "1acff19bbb4a2e0eeca934cc90040eda";

export default node;
