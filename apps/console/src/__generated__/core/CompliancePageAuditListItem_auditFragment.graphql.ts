/**
 * @generated SignedSource<<6201aa75d6361fd302caba39179ebe6e>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type AuditState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageAuditListItem_auditFragment$data = {
  readonly framework: {
    readonly name: string;
  };
  readonly frameworkType: string | null | undefined;
  readonly id: string;
  readonly name: string | null | undefined;
  readonly state: AuditState;
  readonly trustCenterVisibility: TrustCenterVisibility;
  readonly validUntil: string | null | undefined;
  readonly " $fragmentType": "CompliancePageAuditListItem_auditFragment";
};
export type CompliancePageAuditListItem_auditFragment$key = {
  readonly " $data"?: CompliancePageAuditListItem_auditFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAuditListItem_auditFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "name",
  "storageKey": null
};
return {
  "argumentDefinitions": [],
  "kind": "Fragment",
  "metadata": null,
  "name": "CompliancePageAuditListItem_auditFragment",
  "selections": [
    {
      "alias": null,
      "args": null,
      "kind": "ScalarField",
      "name": "id",
      "storageKey": null
    },
    (v0/*: any*/),
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
        (v0/*: any*/)
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
  "type": "Audit",
  "abstractKey": null
};
})();

(node as any).hash = "8c20a1fd2221e4813c60d4f8c81dc016";

export default node;
