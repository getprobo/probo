/**
 * @generated SignedSource<<2a5e6cd48ca8a8fd7fe448c3a6190ccc>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type AuditState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type UpdateAuditInput = {
  id: string;
  showOnTrustCenter?: boolean | null | undefined;
  state?: AuditState | null | undefined;
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
      readonly showOnTrustCenter: boolean;
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
  "name": "showOnTrustCenter",
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
              {
                "alias": null,
                "args": null,
                "concreteType": "Framework",
                "kind": "LinkedField",
                "name": "framework",
                "plural": false,
                "selections": [
                  {
                    "alias": null,
                    "args": null,
                    "kind": "ScalarField",
                    "name": "name",
                    "storageKey": null
                  },
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
    "cacheID": "9cb78f25506e1a67b5f29d4f742f8773",
    "id": null,
    "metadata": {},
    "name": "TrustCenterAuditGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation TrustCenterAuditGraphUpdateMutation(\n  $input: UpdateAuditInput!\n) {\n  updateAudit(input: $input) {\n    audit {\n      id\n      showOnTrustCenter\n      ...TrustCenterAuditsCardFragment\n    }\n  }\n}\n\nfragment TrustCenterAuditsCardFragment on Audit {\n  id\n  framework {\n    name\n    id\n  }\n  validFrom\n  validUntil\n  state\n  showOnTrustCenter\n  createdAt\n}\n"
  }
};
})();

(node as any).hash = "1e92374989ecf8f439d069eb003e4e4e";

export default node;
