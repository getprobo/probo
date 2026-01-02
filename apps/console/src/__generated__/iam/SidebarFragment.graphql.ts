/**
 * @generated SignedSource<<c1333ad749a9c6aaf8f6f671bc578c04>>
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
  readonly canListRightsRequests: boolean;
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

const node: ReaderFragment = {
  "argumentDefinitions": [],
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
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:meeting:list\")"
    },
    {
      "alias": "canListTasks",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:task:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:task:list\")"
    },
    {
      "alias": "canListMeasures",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:measures:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:measures:list\")"
    },
    {
      "alias": "canListRisks",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:risk:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:risk:list\")"
    },
    {
      "alias": "canListFrameworks",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:frameworks:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:frameworks:list\")"
    },
    {
      "alias": "canListPeople",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:people:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:people:list\")"
    },
    {
      "alias": "canListVendors",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:vendor:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:vendor:list\")"
    },
    {
      "alias": "canListDocuments",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:document:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:document:list\")"
    },
    {
      "alias": "canListAssets",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:asset:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:asset:list\")"
    },
    {
      "alias": "canListData",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:datum:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:datum:list\")"
    },
    {
      "alias": "canListAudits",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:audit:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:audit:list\")"
    },
    {
      "alias": "canListNonconformities",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:nonconformity:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:nonconformity:list\")"
    },
    {
      "alias": "canListObligations",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:obligation:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:obligation:list\")"
    },
    {
      "alias": "canListContinualImprovements",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:continual-improvement:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:continual-improvement:list\")"
    },
    {
      "alias": "canListProcessingActivities",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:processing-activity:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:processing-activity:list\")"
    },
    {
      "alias": "canListRightsRequests",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:rights-request:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:rights-request:list\")"
    },
    {
      "alias": "canListSnapshots",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:snapshot:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:snapshot:list\")"
    },
    {
      "alias": "canGetTrustCenter",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:trust-center:get"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:trust-center:get\")"
    },
    {
      "alias": "canUpdateOrganization",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:organization:update"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"iam:organization:update\")"
    },
    {
      "alias": "canListStatesOfApplicability",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "core:state-of-applicability:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:state-of-applicability:list\")"
    }
  ],
  "type": "Organization",
  "abstractKey": null
};

(node as any).hash = "3a7071a1d9ed4ffc115f5bc22e283a50";

export default node;
