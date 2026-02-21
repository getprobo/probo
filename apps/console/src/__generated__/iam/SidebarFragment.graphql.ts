/**
 * @generated SignedSource<<de9ca0854fde114ecd6fc68643f999d3>>
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
  readonly canListMembers: boolean;
  readonly canListNonconformities: boolean;
  readonly canListObligations: boolean;
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
          "value": "core:measure:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:measure:list\")"
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
          "value": "core:framework:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"core:framework:list\")"
    },
    {
      "alias": "canListMembers",
      "args": [
        {
          "kind": "Literal",
          "name": "action",
          "value": "iam:membership:list"
        }
      ],
      "kind": "ScalarField",
      "name": "permission",
      "storageKey": "permission(action:\"iam:membership:list\")"
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

(node as any).hash = "73a632deb40d322a670cd72fd855723f";

export default node;
