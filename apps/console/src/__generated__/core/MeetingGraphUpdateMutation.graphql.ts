/**
 * @generated SignedSource<<c417591e528a7c10a3e5d47fd288cc29>>
 * @lightSyntaxTransform
 * @nogrep
 */

/* tslint:disable */
/* eslint-disable */
// @ts-nocheck

import { ConcreteRequest } from 'relay-runtime';
export type UpdateMeetingInput = {
  attendeeIds?: ReadonlyArray<string> | null | undefined;
  date?: string | null | undefined;
  meetingId: string;
  minutes?: string | null | undefined;
  name?: string | null | undefined;
};
export type MeetingGraphUpdateMutation$variables = {
  input: UpdateMeetingInput;
};
export type MeetingGraphUpdateMutation$data = {
  readonly updateMeeting: {
    readonly meeting: {
      readonly attendees: ReadonlyArray<{
        readonly fullName: string;
        readonly id: string;
      }>;
      readonly date: string;
      readonly id: string;
      readonly minutes: string | null | undefined;
      readonly name: string;
    };
  };
};
export type MeetingGraphUpdateMutation = {
  response: MeetingGraphUpdateMutation$data;
  variables: MeetingGraphUpdateMutation$variables;
};

const node: ConcreteRequest = (function(){
var v0 = [
  {
    "defaultValue": null,
    "kind": "LocalArgument",
    "name": "input"
  }
],
v1 = {
  "alias": null,
  "args": null,
  "kind": "ScalarField",
  "name": "id",
  "storageKey": null
},
v2 = [
  {
    "alias": null,
    "args": [
      {
        "kind": "Variable",
        "name": "input",
        "variableName": "input"
      }
    ],
    "concreteType": "UpdateMeetingPayload",
    "kind": "LinkedField",
    "name": "updateMeeting",
    "plural": false,
    "selections": [
      {
        "alias": null,
        "args": null,
        "concreteType": "Meeting",
        "kind": "LinkedField",
        "name": "meeting",
        "plural": false,
        "selections": [
          (v1/*: any*/),
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
            "concreteType": "Profile",
            "kind": "LinkedField",
            "name": "attendees",
            "plural": true,
            "selections": [
              (v1/*: any*/),
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
  }
];
return {
  "fragment": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Fragment",
    "metadata": null,
    "name": "MeetingGraphUpdateMutation",
    "selections": (v2/*: any*/),
    "type": "Mutation",
    "abstractKey": null
  },
  "kind": "Request",
  "operation": {
    "argumentDefinitions": (v0/*: any*/),
    "kind": "Operation",
    "name": "MeetingGraphUpdateMutation",
    "selections": (v2/*: any*/)
  },
  "params": {
    "cacheID": "e40a6b6cf200c6c90ca61a21c15452c1",
    "id": null,
    "metadata": {},
    "name": "MeetingGraphUpdateMutation",
    "operationKind": "mutation",
    "text": "mutation MeetingGraphUpdateMutation(\n  $input: UpdateMeetingInput!\n) {\n  updateMeeting(input: $input) {\n    meeting {\n      id\n      name\n      date\n      minutes\n      attendees {\n        id\n        fullName\n      }\n    }\n  }\n}\n"
  }
};
})();

(node as any).hash = "b3d9ca4361201376b0d5ab78ec1ef316";

export default node;
