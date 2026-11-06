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
				operation: ['updateAsanaSettings'],
			},
		},
		default: '',
		description: 'The ID of the organization',
		required: true,
	},
	{
		displayName: 'Default Project',
		name: 'defaultProject',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateAsanaSettings'],
			},
		},
		options: [
			{
				name: 'Keep',
				value: 'KEEP',
				description: 'Leave the default Asana project unchanged',
			},
			{
				name: 'Set',
				value: 'SET',
				description: 'Choose an Asana project as the organization default',
			},
			{
				name: 'Clear',
				value: 'CLEAR',
				description: 'Clear the default Asana project and turn task import off',
			},
		],
		default: 'KEEP',
		description: 'Whether to change the default Asana project. Clearing it also turns task import off.',
	},
	{
		displayName: 'Project ID',
		name: 'projectId',
		type: 'string',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateAsanaSettings'],
				defaultProject: ['SET'],
			},
		},
		default: '',
		description: 'The Asana project GID to use as the organization default',
		required: true,
	},
	{
		displayName: 'Task Import',
		name: 'taskImport',
		type: 'options',
		displayOptions: {
			show: {
				resource: ['task'],
				operation: ['updateAsanaSettings'],
			},
		},
		options: [
			{
				name: 'Keep',
				value: 'KEEP',
				description: 'Leave task import unchanged',
			},
			{
				name: 'Active',
				value: 'ACTIVE',
				description: 'Create a Probo task for each new task in the default Asana project and keep it in sync',
			},
			{
				name: 'Inactive',
				value: 'INACTIVE',
				description: 'Leave new tasks in the default Asana project in Asana',
			},
		],
		default: 'KEEP',
		description: 'Whether new tasks in the default Asana project are imported. Active requires a default project.',
	},
];

export async function execute(
	this: IExecuteFunctions,
	itemIndex: number,
): Promise<INodeExecutionData> {
	const organizationId = this.getNodeParameter('organizationId', itemIndex) as string;
	const defaultProject = this.getNodeParameter('defaultProject', itemIndex) as string;
	const taskImport = this.getNodeParameter('taskImport', itemIndex) as string;

	if (defaultProject === 'KEEP' && taskImport === 'KEEP') {
		throw new NodeOperationError(this.getNode(), 'Choose an Asana project or a task import status.');
	}

	const input: {
		organizationId: string;
		defaultProjectId?: string | null;
		taskImport?: string;
	} = { organizationId };

	if (defaultProject === 'CLEAR') {
		input.defaultProjectId = null;
	} else if (defaultProject === 'SET') {
		input.defaultProjectId = this.getNodeParameter('projectId', itemIndex) as string;
	}

	if (taskImport !== 'KEEP') {
		input.taskImport = taskImport;
	}

	const query = `
		mutation UpdateAsanaSettings($input: UpdateAsanaSettingsInput!) {
			updateAsanaSettings(input: $input) {
				organization {
					id
					asanaSettings {
						defaultProject {
							id
							name
						}
						taskImport
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
