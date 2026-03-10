/**
 * @generated SignedSource<<1a8f618e333bfd588ba06578d908bb6b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UploadVendorComplianceReportInput = {
  file: any;
  reportDate: string;
  reportName: string;
  validUntil?: string | null | undefined;
  vendorId: string;
};
export type UploadComplianceReportDialogMutation$variables = {
  connections: ReadonlyArray<string>;
  input: UploadVendorComplianceReportInput;
};
export type UploadComplianceReportDialogMutation$data = {
  readonly uploadVendorComplianceReport: {
    readonly vendorComplianceReportEdge: {
      readonly node: {
        readonly canDelete: boolean;
        readonly file: {
          readonly downloadUrl: string;
          readonly fileName: string;
          readonly mimeType: string;
          readonly size: number;
        } | null | undefined;
        readonly id: string;
        readonly reportDate: string;
        readonly reportName: string;
        readonly validUntil: string | null | undefined;
      };
    };
  };
};
export type UploadComplianceReportDialogMutation = {
  response: UploadComplianceReportDialogMutation$data;
  variables: UploadComplianceReportDialogMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "connections"
},
v1 = {
  "defaultValue": null,
  "kind": "LocalArgument",
  "name": "input"
},
v2 = [
  {
    "kind": "Variable",
    "name": "input",
    "variableName": "input"
  }
],
v3 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "reportName",
  "storageKey": null
},
v5 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "reportDate",
  "storageKey": null
},
v6 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "validUntil",
  "storageKey": null
},
v7 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "fileName",
  "storageKey": null
},
v8 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "mimeType",
  "storageKey": null
},
v9 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "size",
  "storageKey": null
},
v10 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "downloadUrl",
  "storageKey": null
},
v11 = {
  "alias": "canDelete",
  "args": [
    {
      "kind": "Literal",
      "name": "action",
      "value": "core:vendor-compliance-report:delete"
    }
  ],
  "kind": "ScalarField",
  "name": "permission",
  "storageKey": "permission(action:\"core:vendor-compliance-report:delete\")"
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "UploadComplianceReportDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "UploadVendorComplianceReportPayload",
        "kind": "LinkedField",
        "name": "uploadVendorComplianceReport",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorComplianceReportEdge",
            "kind": "LinkedField",
            "name": "vendorComplianceReportEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorComplianceReport",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "File",
                    "kind": "LinkedField",
                    "name": "file",
                    "plural": false,
                    "selections": [
                      (v7/*: any*/),
                      (v8/*: any*/),
                      (v9/*: any*/),
                      (v10/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v11/*: any*/)
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
    "argumentDefinitions": [
      (v1/*: any*/),
      (v0/*: any*/)
    ],
    "kind": "Operation",
    "name": "UploadComplianceReportDialogMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "UploadVendorComplianceReportPayload",
        "kind": "LinkedField",
        "name": "uploadVendorComplianceReport",
        "plural": false,
        "selections": [
          {
            "alias": null,
            "args": null,
            "concreteType": "VendorComplianceReportEdge",
            "kind": "LinkedField",
            "name": "vendorComplianceReportEdge",
            "plural": false,
            "selections": [
              {
                "alias": null,
                "args": null,
                "concreteType": "VendorComplianceReport",
                "kind": "LinkedField",
                "name": "node",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v4/*: any*/),
                  (v5/*: any*/),
                  (v6/*: any*/),
                  {
                    "alias": null,
                    "args": null,
                    "concreteType": "File",
                    "kind": "LinkedField",
                    "name": "file",
                    "plural": false,
                    "selections": [
                      (v7/*: any*/),
                      (v8/*: any*/),
                      (v9/*: any*/),
                      (v10/*: any*/),
                      (v3/*: any*/)
                    ],
                    "storageKey": null
                  },
                  (v11/*: any*/)
                ],
                "storageKey": null
              }
            ],
            "storageKey": null
          },
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "appendEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "vendorComplianceReportEdge",
            "handleArgs": [
              {
                "kind": "Variable",
                "name": "connections",
                "variableName": "connections"
              }
            ]
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "b48652c18d3b6acde303491746764697",
    "id": null,
    "metadata": {},
    "name": "UploadComplianceReportDialogMutation",
    "operationKind": "mutation",
    "text": "mutation UploadComplianceReportDialogMutation(\n  $input: UploadVendorComplianceReportInput!\n) {\n  uploadVendorComplianceReport(input: $input) {\n    vendorComplianceReportEdge {\n      node {\n        id\n        reportName\n        reportDate\n        validUntil\n        file {\n          fileName\n          mimeType\n          size\n          downloadUrl\n          id\n        }\n        canDelete: permission(action: \"core:vendor-compliance-report:delete\")\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "8311149c4e46ed2f01cbd19d34bc3a38";

export default node;
