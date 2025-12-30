/**
 * @generated SignedSource<<0d25ae690e5016a27d889ed1c4a1667f>>
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
  validFrom?: any | null | undefined;
  validUntil?: any | null | undefined;
};
export type TrustCenterAuditGraphUpdateMutation$variables = {
  input: UpdateAuditInput;
};
export type TrustCenterAuditGraphUpdateMutation$data = {
  readonly updateAudit: {
    readonly audit: {
      readonly id: string;
      readonly trustCenterVisibility: TrustCenterVisibility;
      readonly " $fragmentSpreads": FragmentRefs<"TrustCenterAuditsCardFragment">;
    };
  };
};
export type TrustCenterAuditGraphUpdateMutation = {
  response: TrustCenterAuditGraphUpdateMutation$data;
  variables: TrustCenterAuditGraphUpdateMutation$variables;
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
  "name": "trustCenterVisibility",
  "storageKey": null
},
v4 = {
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
    "name": "TrustCenterAuditGraphUpdateMutation",
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
                "args": null,
                "kind": "FragmentSpread",
                "name": "TrustCenterAuditsCardFragment"
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
    "name": "TrustCenterAuditGraphUpdateMutation",
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
              (v4/*: any*/),
              {
                "alias": null,
                "args": null,
                "concreteType": "Framework",
                "kind": "LinkedField",
                "name": "framework",
                "plural": false,
                "selections": [
                  (v4/*: any*/),
                  (v2/*: any*/)
                ],
                "storageKey": null
              },
              {
                "alias": null,
                "args": null,
                "kind": "ScalarField",
                "name": "validFrom",
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
                "name": "createdAt",
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
    "cacheID": "70742901eeb45d5e00e4e61d83b2f828",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAuditGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAuditGraphUpdateMutation(\n  $input: UpdateAuditInput!\n) {\n  updateAudit(input: $input) {\n    audit {\n      id\n      trustCenterVisibility\n      ...TrustCenterAuditsCardFragment\n    }\n  }\n}\n\nfragment TrustCenterAuditsCardFragment on Audit {\n  id\n  name\n  framework {\n    name\n    id\n  }\n  validFrom\n  validUntil\n  state\n  trustCenterVisibility\n  createdAt\n}\n"
  }
};
})();

(node as any).hash = "c958ccce67f79bb75593e9fcb31ee9d8";

export default node;
