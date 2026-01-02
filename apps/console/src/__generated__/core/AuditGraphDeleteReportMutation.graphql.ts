/**
 * @generated SignedSource<<39afab5f5f832b28f00e6bd7469fa4f3>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteAuditReportInput = {
  auditId: string;
};
export type AuditGraphDeleteReportMutation$variables = {
  input: DeleteAuditReportInput;
};
export type AuditGraphDeleteReportMutation$data = {
  readonly deleteAuditReport: {
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
export type AuditGraphDeleteReportMutation = {
  response: AuditGraphDeleteReportMutation$data;
  variables: AuditGraphDeleteReportMutation$variables;
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
    "concreteType": "DeleteAuditReportPayload",
    "kind": "LinkedField",
    "name": "deleteAuditReport",
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
    "name": "AuditGraphDeleteReportMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "AuditGraphDeleteReportMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "ebceb2fd2a2ffae09bc1578e37dde7e0",
    "id": null,
    "metadata": {},
    "name": "AuditGraphDeleteReportMutation",
    "operationKind": "mutation",
    "text": "mutation AuditGraphDeleteReportMutation(\n  $input: DeleteAuditReportInput!\n) {\n  deleteAuditReport(input: $input) {\n    audit {\n      id\n      report {\n        id\n        filename\n        downloadUrl\n        createdAt\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "4b5e9537b35458eb3daa313ead9ba655";

export default node;
