/**
 * @generated SignedSource<<6093b0bc54fa90962376639ec87fa149>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type UploadVendorComplianceReportInput = {
  file: any;
  reportDate: any;
  reportName: string;
  validUntil?: any | null | undefined;
  vendorId: string;
};
export type VendorComplianceTabUploadReportMutation$variables = {
  connections: ReadonlyArray<string>;
  input: UploadVendorComplianceReportInput;
};
export type VendorComplianceTabUploadReportMutation$data = {
  readonly uploadVendorComplianceReport: {
    readonly vendorComplianceReportEdge: {
      readonly node: {
        readonly id: string;
        readonly " $fragmentSpreads": FragmentRefs<"VendorComplianceTabFragment_report">;
      };
    };
  };
};
export type VendorComplianceTabUploadReportMutation = {
  response: VendorComplianceTabUploadReportMutation$data;
  variables: VendorComplianceTabUploadReportMutation$variables;
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
};
return {
  "fragment": {
    "argumentDefinitions": [
      (v0/*: any*/),
      (v1/*: any*/)
    ],
    "kind": "Fragment",
    "metadata": null,
    "name": "VendorComplianceTabUploadReportMutation",
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
                  {
                    "args": null,
                    "kind": "FragmentSpread",
                    "name": "VendorComplianceTabFragment_report"
                  }
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
    "name": "VendorComplianceTabUploadReportMutation",
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
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "reportDate",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "validUntil",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "reportName",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "fileUrl",
                    "storageKey": null
                  },
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "fileSize",
                    "storageKey": null
                  }
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
    "cacheID": "fb9be687deba8567502e207b0969b2b5",
    "id": null,
    "metadata": {},
    "name": "VendorComplianceTabUploadReportMutation",
    "operationKind": "mutation",
    "text": "mutation VendorComplianceTabUploadReportMutation(\n  $input: UploadVendorComplianceReportInput!\n) {\n  uploadVendorComplianceReport(input: $input) {\n    vendorComplianceReportEdge {\n      node {\n        id\n        ...VendorComplianceTabFragment_report\n      }\n    }\n  }\n}\n\nfragment VendorComplianceTabFragment_report on VendorComplianceReport {\n  id\n  reportDate\n  validUntil\n  reportName\n  fileUrl\n  fileSize\n}\n"
  }
};
})();

(node as any).hash = "4ed863a250b0686340056f4f77eb8ab8";

export default node;
