/**
 * @generated SignedSource<<4e95ba3d22dec59dc31a1245cf6b9239>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteAuditReportInput = {
  auditId: string;
  reportId: string;
};
export type AuditGraphDeleteReportMutation$variables = {
  input: DeleteAuditReportInput;
};
export type AuditGraphDeleteReportMutation$data = {
  readonly deleteAuditReport: {
    readonly audit: {
      readonly id: string;
      readonly reports: ReadonlyArray<{
        readonly createdAt: any;
        readonly downloadUrl: string | null | undefined;
        readonly filename: string;
        readonly id: string;
        readonly size: number;
      }>;
      readonly updatedAt: any;
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
            "name": "reports",
            "plural": true,
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
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "size",
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
    "cacheID": "42c70850249e0de8b687d966bf304ca2",
    "id": null,
    "metadata": {},
    "name": "AuditGraphDeleteReportMutation",
    "operationKind": "mutation",
    "text": "mutation AuditGraphDeleteReportMutation(\n  $input: DeleteAuditReportInput!\n) {\n  deleteAuditReport(input: $input) {\n    audit {\n      id\n      reports {\n        id\n        filename\n        downloadUrl\n        createdAt\n        size\n      }\n      updatedAt\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "dde614733e46cdf728dd0262476e48bd";

export default node;
