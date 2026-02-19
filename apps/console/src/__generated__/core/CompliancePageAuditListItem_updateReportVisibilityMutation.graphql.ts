/**
 * @generated SignedSource<<94e22364a84498f77600d636ef0ed80d>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type ReportState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type UpdateReportInput = {
  frameworkType?: string | null | undefined;
  id: string;
  name?: string | null | undefined;
  state?: ReportState | null | undefined;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
  validFrom?: string | null | undefined;
  validUntil?: string | null | undefined;
};
export type CompliancePageAuditListItem_updateReportVisibilityMutation$variables = {
  input: UpdateReportInput;
};
export type CompliancePageAuditListItem_updateReportVisibilityMutation$data = {
  readonly updateReport: {
    readonly report: {
      readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAuditListItem_reportFragment">;
    };
  };
};
export type CompliancePageAuditListItem_updateReportVisibilityMutation = {
  response: CompliancePageAuditListItem_updateReportVisibilityMutation$data;
  variables: CompliancePageAuditListItem_updateReportVisibilityMutation$variables;
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
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
};
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "CompliancePageAuditListItem_updateReportVisibilityMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateReportPayload",
        "kind": "LinkedField",
        "name": "updateReport",
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
              {
                "args": null,
                "kind": "FragmentSpread",
                "name": "CompliancePageAuditListItem_reportFragment"
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
    "name": "CompliancePageAuditListItem_updateReportVisibilityMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateReportPayload",
        "kind": "LinkedField",
        "name": "updateReport",
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
              (v2/*: any*/),
              (v3/*: any*/),
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "frameworkType",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "concreteType": "Framework",
                "kind": "LinkedField",
                "name": "framework",
                "plural": false,
                "selections": [
                  (v3/*: any*/),
                  (v2/*: any*/)
                ],
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
                "name": "state",
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "trustCenterVisibility",
                "storageKey": null
              }
            ],
            "storageKey": null
          }
        ],
        "storageKey": null
      }
    ]
  },
  "params": {
    "cacheID": "7924d023ad78c8a7dcd2180a47d1589f",
    "id": null,
    "metadata": {},
    "name": "CompliancePageAuditListItem_updateReportVisibilityMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageAuditListItem_updateReportVisibilityMutation(\n  $input: UpdateReportInput!\n) {\n  updateReport(input: $input) {\n    report {\n      ...CompliancePageAuditListItem_reportFragment\n      id\n    }\n  }\n}\n\nfragment CompliancePageAuditListItem_reportFragment on Report {\n  id\n  name\n  frameworkType\n  framework {\n    name\n    id\n  }\n  validUntil\n  state\n  trustCenterVisibility\n}\n"
  }
};
})();

(node as any).hash = "f88337e9c833176860cb03b4c1f9cb7c";

export default node;
