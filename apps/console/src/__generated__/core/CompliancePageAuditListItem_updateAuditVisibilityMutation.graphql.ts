/**
 * @generated SignedSource<<4286fb5d45f1d4d8dba035c29219ab34>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type AuditState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
export type UpdateAuditInput = {
  id: string;
  name?: string | null | undefined;
  state?: AuditState | null | undefined;
  trustCenterVisibility?: TrustCenterVisibility | null | undefined;
  validFrom?: string | null | undefined;
  validUntil?: string | null | undefined;
};
export type CompliancePageAuditListItem_updateAuditVisibilityMutation$variables = {
  input: UpdateAuditInput;
};
export type CompliancePageAuditListItem_updateAuditVisibilityMutation$data = {
  readonly updateAudit: {
    readonly audit: {
      readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAuditListItem_auditFragment">;
    };
  };
};
export type CompliancePageAuditListItem_updateAuditVisibilityMutation = {
  response: CompliancePageAuditListItem_updateAuditVisibilityMutation$data;
  variables: CompliancePageAuditListItem_updateAuditVisibilityMutation$variables;
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
    "name": "CompliancePageAuditListItem_updateAuditVisibilityMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateAuditPayload",
        "kind": "LinkedField",
        "name": "updateAudit",
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
                "args": null,
                "kind": "FragmentSpread",
                "name": "CompliancePageAuditListItem_auditFragment"
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
    "name": "CompliancePageAuditListItem_updateAuditVisibilityMutation",
    "selections": [
      {
        "alias": null,
        "args": (v1/*: any*/),
        "concreteType": "UpdateAuditPayload",
        "kind": "LinkedField",
        "name": "updateAudit",
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
              (v2/*: any*/),
              (v3/*: any*/),
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
    "cacheID": "4363d4bbf56cf55359dc0a1ff19e1dbf",
    "id": null,
    "metadata": {},
    "name": "CompliancePageAuditListItem_updateAuditVisibilityMutation",
    "operationKind": "mutation",
    "text": "mutation CompliancePageAuditListItem_updateAuditVisibilityMutation(\n  $input: UpdateAuditInput!\n) {\n  updateAudit(input: $input) {\n    audit {\n      ...CompliancePageAuditListItem_auditFragment\n      id\n    }\n  }\n}\n\nfragment CompliancePageAuditListItem_auditFragment on Audit {\n  id\n  name\n  framework {\n    name\n    id\n  }\n  validUntil\n  state\n  trustCenterVisibility\n}\n"
  }
};
})();

(node as any).hash = "6868ffa86e630bac03c87aa363dc59e2";

export default node;
