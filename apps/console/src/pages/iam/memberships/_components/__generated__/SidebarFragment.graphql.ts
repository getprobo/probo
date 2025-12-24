/**
 * @generated SignedSource<<d7cb6d45a4e2a8960d1984e0e1f74b0a>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ReaderFragment } from 'relay-runtime';
import { FragmentRefs } from "relay-runtime";
export type SidebarFragment$data = {
  readonly canGetTrustCenter: boolean;
  readonly canListAssets: boolean;
  readonly canListAudits: boolean;
  readonly canListContinualImprovements: boolean;
  readonly canListData: boolean;
  readonly canListDocuments: boolean;
  readonly canListFrameworks: boolean;
  readonly canListMeasures: boolean;
  readonly canListMeetings: boolean;
  readonly canListNonconformities: boolean;
  readonly canListObligations: boolean;
  readonly canListPeople: boolean;
  readonly canListProcessingActivities: boolean;
  readonly canListRisks: boolean;
  readonly canListSnapshots: boolean;
  readonly canListStatesOfApplicability: boolean;
  readonly canListTasks: boolean;
  readonly canListVendors: boolean;
  readonly canUpdateOrganization: boolean;
  readonly " $fragmentType": "SidebarFragment";
};
export type SidebarFragment$key = {
  readonly " $data"?: SidebarFragment$data;
  readonly " $fragmentSpreads": FragmentRefs<"SidebarFragment">;
};

const node: ReaderFragment = (function(){
var v0 = {
  "kind": "Variable",
  "name": "id",
  "variableName": "organizationId"
};
return {
  "argumentDefinitions": [
    {
      "defaultValue": null,
      "kind": "LocalArgument",
      "name": "organizationId"
    }
  ],
  "kind": "Fragment",
  "metadata": null,
  "name": "SidebarFragment",
  "selections": [
    {
      "alias": "canListMeetings",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:meeting:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListTasks",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:task:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListMeasures",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:measures:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListRisks",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:risk:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListFrameworks",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:frameworks:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListPeople",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:people:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListVendors",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:vendor:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListDocuments",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:document:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListAssets",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:asset:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListData",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:datum:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListAudits",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:audit:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListNonconformities",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:nonconformity:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListObligations",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:obligation:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListContinualImprovements",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:continual-improvement:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListProcessingActivities",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:processing-activity:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListStatesOfApplicability",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:state-of-applicability:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canListSnapshots",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:snapshot:list"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canGetTrustCenter",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:get"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    },
    {
      "alias": "canUpdateOrganization",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:organization:update"
        },
        (v0/*: any*/)
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": null
    }
  ],
  "type": "Identity",
  "abstractKey": null
};
})();

(node as any).hash = "a116eb6c3f1c2808f90316b34ef43979";

export default node;
