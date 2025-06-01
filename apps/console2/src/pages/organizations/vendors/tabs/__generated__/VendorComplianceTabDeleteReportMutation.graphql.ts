/**
 * @generated SignedSource<<6fc7bf1e239dfeccb007ee5a9beefaff>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type DeleteVendorComplianceReportInput = {
  reportId: string;
};
export type VendorComplianceTabDeleteReportMutation$variables = {
  connections: ReadonlyArray<string>;
  input: DeleteVendorComplianceReportInput;
};
export type VendorComplianceTabDeleteReportMutation$data = {
  readonly deleteVendorComplianceReport: {
    readonly deletedVendorComplianceReportId: string;
  };
};
export type VendorComplianceTabDeleteReportMutation = {
  response: VendorComplianceTabDeleteReportMutation$data;
  variables: VendorComplianceTabDeleteReportMutation$variables;
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
  "name": "deletedVendorComplianceReportId",
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
    "name": "VendorComplianceTabDeleteReportMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteVendorComplianceReportPayload",
        "kind": "LinkedField",
        "name": "deleteVendorComplianceReport",
        "plural": false,
        "selections": [
          (v3/*: any*/)
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
    "name": "VendorComplianceTabDeleteReportMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "DeleteVendorComplianceReportPayload",
        "kind": "LinkedField",
        "name": "deleteVendorComplianceReport",
        "plural": false,
        "selections": [
          (v3/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "deleteEdge",
            "key": "",
            "kind": "ScalarHandle",
            "name": "deletedVendorComplianceReportId",
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
    "cacheID": "3bfaddbf09d2cd72a0bb72687c1e187e",
    "id": null,
    "metadata": {},
    "name": "VendorComplianceTabDeleteReportMutation",
    "operationKind": "mutation",
    "text": "mutation VendorComplianceTabDeleteReportMutation(\n  $input: DeleteVendorComplianceReportInput!\n) {\n  deleteVendorComplianceReport(input: $input) {\n    deletedVendorComplianceReportId\n  }\n}\n"
  }
};
})();

(node as any).hash = "7151aea21838384dc10781b64f0209ae";

export default node;
