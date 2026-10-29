// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { graphql, useLazyLoadQuery } from "react-relay";

import type { AsanaMappingEditorQuery } from "#/__generated__/core/AsanaMappingEditorQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import { AsanaMappingForm } from "./AsanaMappingForm";

const setupQuery = graphql`
  query AsanaMappingEditorQuery($organizationId: ID!, $projectId: String!) {
    node(id: $organizationId) {
      __typename
      ... on Organization {
        asanaProjectSetup(projectId: $projectId) {
          sections {
            id
            name
          }
          enumFields {
            id
            name
            options {
              id
              name
            }
          }
          mapping {
            states {
              state
              sectionId
              inbound
            }
            priorityFieldId
            priorities {
              priority
              optionId
              inbound
            }
          }
        }
      }
    }
  }
`;

export interface AsanaMappingEditorProps {
  projectId: string;
  disabled?: boolean;
}

export function AsanaMappingEditor({
  projectId,
  disabled,
}: AsanaMappingEditorProps) {
  const organizationId = useOrganizationId();
  const data = useLazyLoadQuery<AsanaMappingEditorQuery>(
    setupQuery,
    { organizationId, projectId },
    { fetchPolicy: "network-only" },
  );
  const organization = data.node?.__typename === "Organization" ? data.node : null;
  const setup = organization?.asanaProjectSetup;
  if (setup == null) {
    return null;
  }

  return (
    <AsanaMappingForm
      key={projectId}
      organizationId={organizationId}
      projectId={projectId}
      disabled={disabled}
      sections={setup.sections.map(section => ({ id: section.id, name: section.name }))}
      fields={setup.enumFields.map(field => ({
        id: field.id,
        name: field.name,
        options: field.options.map(option => ({ id: option.id, name: option.name })),
      }))}
      savedStates={setup.mapping?.states.map(row => ({
        state: row.state,
        sectionId: row.sectionId,
        inbound: row.inbound,
      })) ?? null}
      savedFieldId={setup.mapping?.priorityFieldId ?? null}
      savedPriorities={setup.mapping?.priorities.map(row => ({
        priority: row.priority,
        optionId: row.optionId,
        inbound: row.inbound,
      })) ?? null}
    />
  );
}
