/**
 * @generated SignedSource<<316ff962a0b9065de84719ae14ffa68b>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
export type ReportState = "COMPLETED" | "IN_PROGRESS" | "NOT_STARTED" | "OUTDATED" | "REJECTED";
export type TrustCenterVisibility = "NONE" | "PRIVATE" | "PUBLIC";
import { FragmentRefs } from "relay-runtime";
export type CompliancePageAuditListItem_reportFragment$data = {
  readonly framework: {
    readonly name: string;
  };
  readonly frameworkType: string | null | undefined;
  readonly id: string;
  readonly name: string | null | undefined;
  readonly state: ReportState;
  readonly trustCenterVisibility: TrustCenterVisibility;
  readonly validUntil: string | null | undefined;
  readonly " $fragmentType": "CompliancePageAuditListItem_reportFragment";
};
export type CompliancePageAuditListItem_reportFragment$key = {
  readonly " $data"?: CompliancePageAuditListItem_reportFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"CompliancePageAuditListItem_reportFragment">;
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
  "name": "CompliancePageAuditListItem_reportFragment",
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
  "type": "Report",
  "abstractKey": null
};
})();

(node as any).hash = "5864f297cdeb350379a74abe43268a0f";

export default node;
