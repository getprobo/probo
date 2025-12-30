/**
 * @generated SignedSource<<fb3a05b463e9f6065118d1b4441ce398>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type CreateMeetingInput = {
  attendeeIds?: ReadonlyArray<string> | null | undefined;
  date: any;
  minutes?: string | null | undefined;
  name: string;
  organizationId: string;
};
export type CreateMeetingDialogCreateMutation$variables = {
  connections: ReadonlyArray<string>;
  input: CreateMeetingInput;
};
export type CreateMeetingDialogCreateMutation$data = {
  readonly createMeeting: {
    readonly meetingEdge: {
      readonly node: {
        readonly attendees: ReadonlyArray<{
          readonly fullName: string;
          readonly id: string;
        }>;
        readonly date: any;
        readonly id: string;
        readonly minutes: string | null | undefined;
        readonly name: string;
      };
    };
  };
};
export type CreateMeetingDialogCreateMutation = {
  response: CreateMeetingDialogCreateMutation$data;
  variables: CreateMeetingDialogCreateMutation$variables;
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
  "name": "id",
  "storageKey": null
},
v4 = {
  "alias": null,
  "args": null,
  "concreteType": "MeetingEdge",
  "kind": "LinkedField",
  "name": "meetingEdge",
  "plural": false,
  "selections": [
    {
      "alias": null,
      "args": null,
      "concreteType": "Meeting",
      "kind": "LinkedField",
      "name": "node",
      "plural": false,
      "selections": [
        (v3/*: any*/),
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "name",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "date",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "kind": "ScalarField",
          "name": "minutes",
          "storageKey": null
        },
        {
          "alias": null,
          "args": null,
          "concreteType": "People",
          "kind": "LinkedField",
          "name": "attendees",
          "plural": true,
          "selections": [
            (v3/*: any*/),
            {
              "alias": null,
              "args": null,
              "kind": "ScalarField",
              "name": "fullName",
              "storageKey": null
            }
          ],
          "storageKey": null
        }
      ],
      "storageKey": null
    }
  ],
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
    "name": "CreateMeetingDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMeetingPayload",
        "kind": "LinkedField",
        "name": "createMeeting",
        "plural": false,
        "selections": [
          (v4/*: any*/)
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
    "name": "CreateMeetingDialogCreateMutation",
    "selections": [
      {
        "alias": null,
        "args": (v2/*: any*/),
        "concreteType": "CreateMeetingPayload",
        "kind": "LinkedField",
        "name": "createMeeting",
        "plural": false,
        "selections": [
          (v4/*: any*/),
          {
            "alias": null,
            "args": null,
            "filters": null,
            "handle": "prependEdge",
            "key": "",
            "kind": "LinkedHandle",
            "name": "meetingEdge",
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
    "cacheID": "cf37e45c7e8cffd8b44cdc85488209e8",
    "id": null,
    "metadata": {},
    "name": "CreateMeetingDialogCreateMutation",
    "operationKind": "mutation",
    "text": "mutation CreateMeetingDialogCreateMutation(\n  $input: CreateMeetingInput!\n) {\n  createMeeting(input: $input) {\n    meetingEdge {\n      node {\n        id\n        name\n        date\n        minutes\n        attendees {\n          id\n          fullName\n        }\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "5ba1bb7a0344e0ea19b383af6d75534f";

export default node;
