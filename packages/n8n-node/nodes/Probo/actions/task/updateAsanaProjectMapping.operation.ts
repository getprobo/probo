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

import type { IExecuteFunctions, INodeExecutionData, INodeProperties } from 'n8n-workflow';
import { NodeOperationError } from 'n8n-workflow';
import { proboApiRequest } from '../../GenericFunctions';

export const description: INodeProperties[] = [
	{
		displayName: 'Organization ID',
		name: 'organizationId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateAsanaProjectMapping'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Project ID',
		name: 'projectId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateAsanaProjectMapping'],
			},
		},
		default: '',
		description: 'The Asana project GID',
		required: true,
	},
	{
		displayName: 'Clear Mapping',
		name: 'clear',
		type: 'boolean',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateAsanaProjectMapping'],
			},
		},
		default: false,
		description: 'Whether to remove the saved mapping for the project',
	},
	{
		displayName: 'States',
		name: 'states',
		type: 'json',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateAsanaProjectMapping'],
				clear: [false],
			},
		},
		default: '[]',
		description: 'State mapping as a JSON array of {state, sectionId, inbound}. Leave empty to keep the current states. When several states share a section, exactly one must be incoming.',
	},
	{
		displayName: 'Priority Field',
		name: 'priorityField',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateAsanaProjectMapping'],
				clear: [false],
			},
		},
		options: [
			{
				name: 'Keep',
				value: 'KEEP',
				description: 'Leave the priority field unchanged',
			},
			{
				name: 'Set',
				value: 'SET',
				description: 'Sync priority from an Asana enum custom field',
			},
			{
				name: 'Clear',
				value: 'CLEAR',
				description: 'Stop syncing priority',
			},
		],
		default: 'KEEP',
		description: 'Whether to change the enum custom field used as priority',
	},
	{
		displayName: 'Priority Field ID',
		name: 'priorityFieldId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateAsanaProjectMapping'],
				clear: [false],
				priorityField: ['SET'],
			},
		},
		default: '',
		description: 'The Asana enum custom field GID',
		required: true,
	},
	{
		displayName: 'Priorities',
		name: 'priorities',
		type: 'json',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateAsanaProjectMapping'],
				clear: [false],
			},
		},
		default: '[]',
		description: 'Priority mapping as a JSON array of {priority, optionId, inbound}. Leave empty to keep the current priorities.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const projectId = this.getNodeParameter('projectId', itemIndex) as string;
	const clear = this.getNodeParameter('clear', itemIndex) as boolean;

	const input: {
		organizationId: string;
		projectId: string;
		clear?: boolean;
		states?: unknown;
		priorityFieldId?: string | null;
		priorities?: unknown;
	} = { organizationId, projectId };

	if (clear) {
		input.clear = true;
	} else {
		const states = jsonRows(this.getNodeParameter('states', itemIndex));
		const priorities = jsonRows(this.getNodeParameter('priorities', itemIndex));
		const priorityField = this.getNodeParameter('priorityField', itemIndex) as string;

		if (states == null && priorities == null && priorityField === 'KEEP') {
			throw new NodeOperationError(
				this.getNode(),
				'Clear the mapping or provide states, priorities, or a priority field.',
				{ itemIndex },
			);
		}

		if (states != null) {
			input.states = states;
		}
		if (priorities != null) {
			input.priorities = priorities;
		}
		if (priorityField === 'CLEAR') {
			input.priorityFieldId = null;
		} else if (priorityField === 'SET') {
			input.priorityFieldId = this.getNodeParameter('priorityFieldId', itemIndex) as string;
		}
	}

	const query = `
		mutation UpdateAsanaProjectMapping($input: UpdateAsanaProjectMappingInput!) {
			updateAsanaProjectMapping(input: $input) {
				setup {
					project { id name }
					mapping {
						states { state sectionId inbound }
						priorityFieldId
						priorities { priority optionId inbound }
					}
				}
			}
		}
	`;

	const responseData = await proboApiRequest.call(this, query, { input });

	return {
		json: responseData,
		pairedItem: { item: itemIndex },
	};
}

function jsonRows(value: unknown): unknown[] | null {
	let parsed = value;
	if (typeof value === 'string') {
		if (value.trim() === '') {
			return null;
		}

		parsed = JSON.parse(value) as unknown;
	}

	if (!Array.isArray(parsed) || parsed.length === 0) {
		return null;
	}

	return parsed;
}
